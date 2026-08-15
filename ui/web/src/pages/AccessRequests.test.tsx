import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { App } from 'antd';
import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { AccessRequests } from '@/pages/AccessRequests';
import { accessRequestService } from '@/services/access';
import type { AccessRequest } from '@/services/access';
import { groupService } from '@/services/group';
import { userService } from '@/services/user';

vi.mock('@/services/access', () => ({
  accessRequestService: {
    list: vi.fn(),
    create: vi.fn(),
    cancel: vi.fn(),
    decide: vi.fn(),
  },
}));
vi.mock('@/services/group', () => ({ groupService: { listGroups: vi.fn() } }));
vi.mock('@/services/user', () => ({ userService: { listUsers: vi.fn() } }));

const mockUseMe = vi.hoisted(() => vi.fn());
vi.mock('@/hooks/useMe', () => ({ useMe: mockUseMe }));

vi.mock('@/hooks/useNotify', () => ({
  useNotify: () => ({ success: vi.fn(), error: vi.fn() }),
}));

const ME = 'me-1';
const COLLEAGUE = 'colleague-1';
// The requester and the subject differ by default, because that is the
// interesting case: a manager asking on behalf of a report, or the assistant
// asking on behalf of whoever it is answering.
const MANAGER = 'manager-1';

const request = (over: Partial<AccessRequest> = {}): AccessRequest => ({
  id: 'req-1',
  requesterUserId: MANAGER,
  subjectUserId: COLLEAGUE,
  groupId: 'group-1',
  reason: 'on call this week',
  status: 'REQUEST_STATUS_PENDING',
  origin: 'REQUEST_ORIGIN_CONSOLE',
  createdAt: '2026-08-15T09:00:00Z',
  ...over,
});

const renderPage = () => {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <App>
        {/* PageBreadcrumb renders a Link, which needs a router in context. */}
        <MemoryRouter>
          <AccessRequests />
        </MemoryRouter>
      </App>
    </QueryClientProvider>,
  );
};

beforeEach(() => {
  vi.clearAllMocks();
  mockUseMe.mockReturnValue({ data: { user: { id: ME, username: 'me' } } });
  vi.mocked(groupService.listGroups).mockResolvedValue({
    groups: [{ id: 'group-1', name: 'production-readonly' }],
  } as Awaited<ReturnType<typeof groupService.listGroups>>);
  vi.mocked(userService.listUsers).mockResolvedValue({
    users: [
      { id: ME, username: 'me' },
      { id: COLLEAGUE, username: 'taro' },
      { id: MANAGER, username: 'hanako' },
    ],
  } as Awaited<ReturnType<typeof userService.listUsers>>);
});

describe('AccessRequests', () => {
  it('shows what would actually be granted, not a summary of it', async () => {
    vi.mocked(accessRequestService.list).mockResolvedValue({
      accessRequests: [request({ requestedUntil: '2026-08-22T09:00:00Z' })],
    });

    renderPage();

    // The subject, the group and the term are all on screen: "grant some
    // access" is not something anybody can meaningfully agree to.
    expect(await screen.findByText('taro')).toBeTruthy();
    expect(screen.getByText('production-readonly')).toBeTruthy();
    expect(screen.getByText(/^until /)).toBeTruthy();
    expect(screen.getByText('on call this week')).toBeTruthy();
    // Who asked is shown as well as who it is for: the two are different
    // people here, and an approver needs both.
    expect(screen.getByText('hanako')).toBeTruthy();
  });

  it('says when a request came from the assistant', async () => {
    vi.mocked(accessRequestService.list).mockResolvedValue({
      accessRequests: [request({ origin: 'REQUEST_ORIGIN_AI_CHAT' })],
    });

    renderPage();

    // An approver has to be able to tell a colleague's ask from one a model was
    // talked into composing. This label is the only thing that tells them.
    expect(await screen.findByText('AI chat')).toBeTruthy();
  });

  it('does not offer a console request as if it came from the assistant', async () => {
    vi.mocked(accessRequestService.list).mockResolvedValue({
      accessRequests: [request()],
    });

    renderPage();

    await screen.findByText('taro');
    expect(screen.queryByText('AI chat')).toBeNull();
  });

  it('keeps your own request out of the queue you decide from', async () => {
    vi.mocked(accessRequestService.list).mockResolvedValue({
      accessRequests: [request({ id: 'mine', requesterUserId: ME, subjectUserId: ME })],
    });

    renderPage();

    // The server refuses a self-approval; showing a Review button that leads to
    // a refusal would only invite the click.
    await waitFor(() => expect(accessRequestService.list).toHaveBeenCalled());
    expect(screen.queryByRole('button', { name: 'Review' })).toBeNull();
    expect(await screen.findByText('Nothing waiting on you')).toBeTruthy();
  });

  it('offers a colleague’s pending request for review', async () => {
    vi.mocked(accessRequestService.list).mockResolvedValue({
      accessRequests: [request()],
    });

    renderPage();

    expect(await screen.findByRole('button', { name: 'Review' })).toBeTruthy();
  });

  it('reads a request with no term as permanent', async () => {
    vi.mocked(accessRequestService.list).mockResolvedValue({
      accessRequests: [request({ requestedUntil: undefined })],
    });

    renderPage();

    // Asking for good is a bigger ask than asking for a week, and has to read
    // as one.
    expect(await screen.findByText('permanently')).toBeTruthy();
  });
});
