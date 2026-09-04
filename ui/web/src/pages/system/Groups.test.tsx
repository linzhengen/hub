import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { App } from 'antd';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { Groups } from '@/pages/system/Groups';
import { groupService } from '@/services/group.ts';
import { organizationService } from '@/services/organization';
import { roleService } from '@/services/role.ts';
import { userService } from '@/services/user.ts';
import { setActiveOrgId } from '@/lib/active-org';

vi.mock('@/services/group.ts', () => ({
  groupService: {
    listGroups: vi.fn(),
    createGroup: vi.fn(),
    updateGroup: vi.fn(),
    deleteGroup: vi.fn(),
    addRoles: vi.fn(),
    removeRoles: vi.fn(),
    addUsers: vi.fn(),
    removeUsers: vi.fn(),
  },
}));
vi.mock('@/services/role.ts', () => ({ roleService: { listRoles: vi.fn() } }));
vi.mock('@/services/user.ts', () => ({ userService: { listUsers: vi.fn() } }));
vi.mock('@/services/organization', () => ({
  organizationService: { list: vi.fn(), listMine: vi.fn() },
}));
vi.mock('@/hooks/useNotify', () => ({
  useNotify: () => ({ success: vi.fn(), error: vi.fn() }),
}));

const ACME = '11111111-1111-1111-1111-111111111111';
const GLOBEX = '22222222-2222-2222-2222-222222222222';

const renderPage = () => {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <App>
        <MemoryRouter>
          <Groups />
        </MemoryRouter>
      </App>
    </QueryClientProvider>,
  );
};

const openCreateForm = async () => {
  // The organization column proves the organization list has arrived, which is
  // what the form preselects from.
  await screen.findByText('Acme Inc.');
  fireEvent.click(await screen.findByRole('button', { name: /add group/i }));
  await screen.findByLabelText('Name');
};

beforeEach(() => {
  vi.clearAllMocks();
  setActiveOrgId('');
  vi.mocked(groupService.listGroups).mockResolvedValue({
    groups: [
      { id: 'group-1', name: 'support', description: '', status: 'STATUS_ACTIVE', orgId: ACME },
    ],
    total: '1',
  });
  vi.mocked(roleService.listRoles).mockResolvedValue({ roles: [], total: '0' });
  vi.mocked(userService.listUsers).mockResolvedValue({ users: [], total: '0' });
  vi.mocked(organizationService.list).mockResolvedValue({
    organizations: [
      { id: ACME, name: 'Acme Inc.', slug: 'acme' },
      { id: GLOBEX, name: 'Globex', slug: 'globex' },
    ],
    total: '2',
  });
  vi.mocked(groupService.createGroup).mockResolvedValue({});
});

describe('Groups', () => {
  // A group belongs to exactly one organization and the server refuses one
  // without it, so the form has to carry it. Leaving it out turned every
  // creation into "org_id: value is empty, which is not a valid UUID".
  it('sends the organization when creating a group', async () => {
    renderPage();
    await openCreateForm();

    fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'platform-team' } });
    fireEvent.mouseDown(screen.getByLabelText('Organization'));
    fireEvent.click(await screen.findByTitle('Globex'));
    fireEvent.click(screen.getByRole('button', { name: 'Create' }));

    // react-query hands the mutationFn a second argument of its own, so the
    // payload is checked rather than the whole call.
    await waitFor(() => expect(groupService.createGroup).toHaveBeenCalled());
    expect(vi.mocked(groupService.createGroup).mock.calls[0][0]).toMatchObject({
      name: 'platform-team',
      orgId: GLOBEX,
    });
  });

  it('refuses to submit without an organization', async () => {
    renderPage();
    await openCreateForm();

    fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'platform-team' } });
    fireEvent.click(screen.getByRole('button', { name: 'Create' }));

    expect(await screen.findByText('Please select an organization!')).toBeTruthy();
    expect(groupService.createGroup).not.toHaveBeenCalled();
  });

  // The switcher says which organization the user is working in; repeating that
  // choice by hand on every group is the kind of step people get wrong.
  it('preselects the organization the header switcher is set to', async () => {
    setActiveOrgId(GLOBEX);

    renderPage();
    await openCreateForm();

    expect(await screen.findByTitle('Globex')).toBeTruthy();
  });

  it('shows which organization each group belongs to', async () => {
    renderPage();

    expect(await screen.findByText('Acme Inc.')).toBeTruthy();
  });
});
