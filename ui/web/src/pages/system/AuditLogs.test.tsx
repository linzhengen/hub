import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { App } from 'antd';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { AuditLogs } from '@/pages/system/AuditLogs';
import { auditService } from '@/services/audit';
import type { AuditLog } from '@/services/audit';
import { userService } from '@/services/user';

vi.mock('@/services/audit', () => ({ auditService: { list: vi.fn() } }));
vi.mock('@/services/user', () => ({ userService: { listUsers: vi.fn() } }));

const atTheConsole: AuditLog = {
  id: 'log-1',
  actorUserId: 'user-1',
  action: 'AddRolesToGroup',
  resource: 'api.system.group.v1.GroupService',
  targetId: 'group-1',
  channel: 'CHANNEL_API',
  succeeded: true,
  arguments: '{"id":"group-1","roleIds":["role-admin"]}',
  createdAt: '2026-08-15T09:00:00Z',
};

const throughTheAssistant: AuditLog = {
  id: 'log-2',
  actorUserId: 'user-1',
  action: 'CreateAccessRequest',
  resource: 'api.system.access.v1.AccessRequestService',
  channel: 'CHANNEL_AI_CHAT',
  sessionId: 'session-9',
  succeeded: true,
  createdAt: '2026-08-15T09:05:00Z',
};

const refused: AuditLog = {
  id: 'log-3',
  actorUserId: 'user-2',
  action: 'AddGroupsToUser',
  resource: 'api.system.user.v1.UserService',
  channel: 'CHANNEL_API',
  succeeded: false,
  error: 'permission denied',
  createdAt: '2026-08-15T09:10:00Z',
};

const renderPage = () => {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <App>
        {/* PageBreadcrumb renders a Link, which needs a router in context. */}
        <MemoryRouter>
          <AuditLogs />
        </MemoryRouter>
      </App>
    </QueryClientProvider>,
  );
};

/** The last query the page asked for, which is what a filter change is judged by. */
const lastQuery = () => vi.mocked(auditService.list).mock.calls.at(-1)?.[0];

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(userService.listUsers).mockResolvedValue({
    users: [
      { id: 'user-1', username: 'hanako' },
      { id: 'user-2', username: 'taro' },
    ],
  } as Awaited<ReturnType<typeof userService.listUsers>>);
  vi.mocked(auditService.list).mockResolvedValue({
    auditLogs: [atTheConsole, throughTheAssistant, refused],
    total: '3',
  } as Awaited<ReturnType<typeof auditService.list>>);
});

describe('AuditLogs', () => {
  // The assistant is a channel, never an actor, so both rows name the same
  // person. "They did this at the console" and "they approved this in a
  // conversation" are different accounts of the same change, and the reader has
  // to be able to tell them apart without inferring it.
  it('distinguishes a change made through the assistant from one made directly', async () => {
    renderPage();

    expect(await screen.findByText('AddRolesToGroup')).toBeTruthy();
    expect(screen.getByText('AI chat')).toBeTruthy();
    expect(screen.getAllByText('Direct').length).toBe(2);
  });

  // Refused attempts are the events this log exists for: somebody trying to
  // grant themselves admin and being told no is exactly what an investigation
  // is looking for, so a failure is coloured rather than filtered out.
  it('shows attempts that were refused, with the reason', async () => {
    renderPage();

    expect(await screen.findByText('AddGroupsToUser')).toBeTruthy();
    expect(screen.getAllByText('Allowed').length).toBe(2);

    // Why it was refused is the other half of the answer, so the tag carries
    // the server's own reason rather than a generic failure.
    fireEvent.mouseEnter(screen.getByText('Refused'));
    expect(await screen.findByText('permission denied')).toBeTruthy();
  });

  it('narrows to one actor', async () => {
    renderPage();
    await screen.findByText('AddRolesToGroup');

    fireEvent.mouseDown(screen.getByLabelText('Actor'));
    fireEvent.click(await screen.findByTitle('taro'));

    await waitFor(() => expect(lastQuery()?.actorUserId).toBe('user-2'));
  });

  it('narrows to a period, in the zone the reader typed it in', async () => {
    renderPage();
    await screen.findByText('AddRolesToGroup');

    fireEvent.change(screen.getByLabelText('From'), { target: { value: '2026-08-15T00:00' } });
    fireEvent.change(screen.getByLabelText('Until'), { target: { value: '2026-08-16T00:00' } });

    await waitFor(() => {
      const query = lastQuery();
      expect(query?.since).toBe(new Date('2026-08-15T00:00').toISOString());
      expect(query?.until).toBe(new Date('2026-08-16T00:00').toISOString());
    });
  });

  it('narrows to the changes the assistant made', async () => {
    renderPage();
    await screen.findByText('AddRolesToGroup');

    fireEvent.mouseDown(screen.getByLabelText('Route'));
    fireEvent.click(await screen.findByTitle('AI chat'));

    await waitFor(() => expect(lastQuery()?.channel).toBe('CHANNEL_AI_CHAT'));
  });

  // "Somebody changed a group" is not something an investigation can conclude
  // anything from; the recorded request is what it is reading the log for.
  it('shows the recorded request behind the row', async () => {
    renderPage();
    await screen.findByText('AddRolesToGroup');

    fireEvent.click(screen.getAllByLabelText('Expand row')[0]);

    expect(await screen.findByText(/role-admin/)).toBeTruthy();
  });

  // A change stays in the log after the person who made it is deleted. An id a
  // reader can look up beats a word that tells them nothing.
  it('falls back to the actor id when the user is gone', async () => {
    vi.mocked(auditService.list).mockResolvedValue({
      auditLogs: [{ ...atTheConsole, actorUserId: 'user-deleted' }],
      total: '1',
    } as Awaited<ReturnType<typeof auditService.list>>);

    renderPage();

    expect(await screen.findByText('user-deleted')).toBeTruthy();
  });

  it('does not offer a newer page from the first one', async () => {
    renderPage();
    await screen.findByText('AddRolesToGroup');

    expect(screen.getByRole('button', { name: 'Newer' }).hasAttribute('disabled')).toBe(true);
    // Everything recorded is on this page, so there is nothing older either.
    expect(screen.getByRole('button', { name: 'Older' }).hasAttribute('disabled')).toBe(true);
  });
});
