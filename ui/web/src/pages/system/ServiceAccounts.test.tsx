import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { App } from 'antd';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { ServiceAccounts } from '@/pages/system/ServiceAccounts';
import { serviceAccountService } from '@/services/serviceAccount';
import type { ServiceAccount } from '@/services/serviceAccount';
import { userService } from '@/services/user';

vi.mock('@/services/serviceAccount', () => ({
  serviceAccountService: {
    list: vi.fn(),
    create: vi.fn(),
    rotateSecret: vi.fn(),
    remove: vi.fn(),
  },
}));
vi.mock('@/services/user', () => ({ userService: { listUsers: vi.fn() } }));

vi.mock('@/hooks/useNotify', () => ({
  useNotify: () => ({ success: vi.fn(), error: vi.fn() }),
}));

const account: ServiceAccount = {
  id: 'sa-1',
  userId: 'user-sa-1',
  name: 'ci-deploy',
  description: 'deploys main',
  clientId: 'hub-sa-ci-deploy',
  createdByUserId: 'creator-1',
  createdAt: '2026-08-15T09:00:00Z',
};

const renderPage = () => {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <App>
        {/* PageBreadcrumb renders a Link, which needs a router in context. */}
        <MemoryRouter>
          <ServiceAccounts />
        </MemoryRouter>
      </App>
    </QueryClientProvider>,
  );
};

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(userService.listUsers).mockResolvedValue({
    users: [{ id: 'creator-1', username: 'hanako' }],
  } as Awaited<ReturnType<typeof userService.listUsers>>);
  vi.mocked(serviceAccountService.list).mockResolvedValue({ serviceAccounts: [account] });
});

describe('ServiceAccounts', () => {
  it('shows what each machine is and who registered it', async () => {
    renderPage();

    expect(await screen.findByText('ci-deploy')).toBeTruthy();
    // The description is the answer to "can we turn this off", so it is on the
    // list rather than behind a click.
    expect(screen.getByText('deploys main')).toBeTruthy();
    expect(screen.getByText('hub-sa-ci-deploy')).toBeTruthy();
    expect(screen.getByText('hanako')).toBeTruthy();
  });

  // The secret exists in the browser exactly once. Nothing reads it back - not
  // the list, not a detail view, not this dialog after it closes - because hub
  // does not store it and Keycloak cannot always show it again.
  it('shows a new secret once and loses it when the dialog closes', async () => {
    vi.mocked(serviceAccountService.create).mockResolvedValue({
      serviceAccount: account,
      credentials: { clientId: 'hub-sa-ci-deploy', secret: 's3cr3t-value' },
    });

    renderPage();
    await screen.findByText('ci-deploy');

    fireEvent.click(screen.getByRole('button', { name: 'Register a machine' }));
    const name = await screen.findByLabelText('Name');
    fireEvent.change(name, { target: { value: 'ci-deploy' } });
    fireEvent.click(screen.getByRole('button', { name: 'Register' }));

    expect(await screen.findByText('s3cr3t-value')).toBeTruthy();
    // The dialog has to say it, not merely be true: a reader who closes it
    // expecting to come back has lost the credential.
    expect(screen.getByText('This secret is shown once')).toBeTruthy();

    fireEvent.click(screen.getByRole('button', { name: 'I have copied the secret' }));

    await waitFor(() => expect(screen.queryByText('s3cr3t-value')).toBeNull());
  });

  it('shows a rotated secret the same way', async () => {
    vi.mocked(serviceAccountService.rotateSecret).mockResolvedValue({
      credentials: { clientId: 'hub-sa-ci-deploy', secret: 'rotated-value' },
    });

    renderPage();
    await screen.findByText('ci-deploy');

    fireEvent.click(screen.getByRole('button', { name: 'Rotate secret' }));
    // The confirmation says the old secret stops working, because rotating
    // breaks whatever is still using it.
    expect(await screen.findByText(/current secret stops working immediately/)).toBeTruthy();
    fireEvent.click(screen.getByRole('button', { name: 'Rotate' }));

    expect(await screen.findByText('rotated-value')).toBeTruthy();
    expect(screen.getByText('This secret is shown once')).toBeTruthy();
  });

  it('warns that removing a machine stops it authenticating', async () => {
    renderPage();
    await screen.findByText('ci-deploy');

    fireEvent.click(screen.getByRole('button', { name: 'Remove' }));

    // Deleting takes the Keycloak client with it, so this is not a
    // hub-only tidy-up and must not read like one.
    expect(await screen.findByText(/can no longer authenticate/)).toBeTruthy();
  });

  it('refuses a name the server would refuse', async () => {
    renderPage();
    await screen.findByText('ci-deploy');

    fireEvent.click(screen.getByRole('button', { name: 'Register a machine' }));
    const name = await screen.findByLabelText('Name');
    fireEvent.change(name, { target: { value: 'CI Deploy' } });
    fireEvent.click(screen.getByRole('button', { name: 'Register' }));

    // The name becomes a Keycloak client id, so it is checked before the round
    // trip rather than after it.
    expect(
      await screen.findByText('Lowercase letters, digits and hyphens; 3-64 characters'),
    ).toBeTruthy();
    expect(serviceAccountService.create).not.toHaveBeenCalled();
  });
});
