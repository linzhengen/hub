import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import NotificationDropdown from '@/components/header/NotificationDropdown';
import { pendingForApprover } from '@/hooks/useAccessRequests';
import { accessRequestService } from '@/services/access';
import type { AccessRequest } from '@/services/access';

vi.mock('@/services/access', () => ({
  accessRequestService: { list: vi.fn() },
}));

const mockUseMe = vi.hoisted(() => vi.fn());
vi.mock('@/hooks/useMe', () => ({ useMe: mockUseMe }));

const ME = 'me-1';
const COLLEAGUE = 'colleague-1';

const request = (over: Partial<AccessRequest> = {}): AccessRequest => ({
  id: 'req-1',
  requesterUserId: COLLEAGUE,
  subjectUserId: COLLEAGUE,
  groupId: 'group-1',
  reason: 'on call this week',
  status: 'REQUEST_STATUS_PENDING',
  origin: 'REQUEST_ORIGIN_CONSOLE',
  createdAt: '2026-08-15T09:00:00Z',
  ...over,
});

const renderBell = () => {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter>
        <NotificationDropdown />
      </MemoryRouter>
    </QueryClientProvider>,
  );
};

beforeEach(() => {
  vi.clearAllMocks();
  mockUseMe.mockReturnValue({ data: { user: { id: ME, username: 'me' } } });
});

// The filter is shared with the page's queue, so it is worth pinning on its
// own: a second copy would let the badge and the list disagree about what is
// waiting on you.
describe('pendingForApprover', () => {
  it('keeps a colleague’s pending request', () => {
    expect(pendingForApprover([request()], ME)).toHaveLength(1);
  });

  it('drops your own request, because you may not decide it', () => {
    expect(pendingForApprover([request({ requesterUserId: ME })], ME)).toHaveLength(0);
  });

  it('drops anything already decided', () => {
    const decided: AccessRequest[] = [
      request({ status: 'REQUEST_STATUS_APPROVED' }),
      request({ status: 'REQUEST_STATUS_REJECTED' }),
      request({ status: 'REQUEST_STATUS_CANCELLED' }),
    ];
    expect(pendingForApprover(decided, ME)).toHaveLength(0);
  });
});

describe('NotificationDropdown', () => {
  it('marks the bell when something is waiting on you', async () => {
    vi.mocked(accessRequestService.list).mockResolvedValue({
      accessRequests: [request()],
    });

    renderBell();

    expect(
      await screen.findByRole('button', { name: 'Notifications, 1 waiting on you' }),
    ).toBeTruthy();
  });

  // A dot that is always lit tells the reader nothing, and they stop looking at
  // it - which costs more than having no dot at all.
  it('does not mark the bell when nothing is', async () => {
    vi.mocked(accessRequestService.list).mockResolvedValue({ accessRequests: [] });

    renderBell();

    await waitFor(() => expect(accessRequestService.list).toHaveBeenCalled());
    expect(screen.getByRole('button', { name: 'Notifications' })).toBeTruthy();
  });

  it('does not count your own request as waiting on you', async () => {
    vi.mocked(accessRequestService.list).mockResolvedValue({
      accessRequests: [request({ requesterUserId: ME, subjectUserId: ME })],
    });

    renderBell();

    await waitFor(() => expect(accessRequestService.list).toHaveBeenCalled());
    // The server refuses a self-approval, so promising one in the header would
    // send the reader to a refusal.
    expect(screen.getByRole('button', { name: 'Notifications' })).toBeTruthy();
  });
});
