import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { App } from 'antd';
import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { Organizations } from '@/pages/system/Organizations';
import { organizationService } from '@/services/organization';
import type { Organization } from '@/services/organization';

vi.mock('@/services/organization', () => ({
  organizationService: {
    list: vi.fn(),
    listMine: vi.fn(),
    get: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    remove: vi.fn(),
  },
}));

vi.mock('@/hooks/useNotify', () => ({
  useNotify: () => ({ success: vi.fn(), error: vi.fn() }),
}));

const platform: Organization = {
  id: '00000000-0000-0000-0000-000000000001',
  name: 'platform',
  slug: 'platform',
  kind: 'ORGANIZATION_KIND_PLATFORM',
  description: 'The operator of this installation.',
  status: 'ORGANIZATION_STATUS_ACTIVE',
  createdAt: '2026-09-01T09:00:00Z',
};

const acme: Organization = {
  id: '11111111-1111-1111-1111-111111111111',
  name: 'Acme Inc.',
  slug: 'acme',
  kind: 'ORGANIZATION_KIND_BUSINESS',
  description: 'a customer',
  status: 'ORGANIZATION_STATUS_ACTIVE',
  createdAt: '2026-09-02T09:00:00Z',
};

const renderPage = () => {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <App>
        {/* PageBreadcrumb renders a Link, which needs a router in context. */}
        <MemoryRouter>
          <Organizations />
        </MemoryRouter>
      </App>
    </QueryClientProvider>,
  );
};

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(organizationService.list).mockResolvedValue({
    organizations: [platform, acme],
    total: '2',
  });
});

describe('Organizations', () => {
  it('lists the tenants with their slug and kind', async () => {
    renderPage();

    expect(await screen.findByText('Acme Inc.')).toBeTruthy();
    expect(screen.getByText('acme')).toBeTruthy();
    expect(screen.getByText('Business')).toBeTruthy();
  });

  // The platform organization's grants reach every other organization, so it has
  // to be distinguishable at a glance from a tenant.
  it('marks the platform organization apart from a tenant', async () => {
    renderPage();

    expect(await screen.findByText('Platform')).toBeTruthy();
    expect(screen.getByText('Business')).toBeTruthy();
  });

  // Deleting it would take every group that predates organizations with it, so
  // the page offers no way to try.
  it('offers no delete for the platform organization', async () => {
    renderPage();

    await screen.findByText('Acme Inc.');
    expect(screen.getByText('Cannot be deleted')).toBeTruthy();
    expect(screen.getAllByRole('button', { name: 'Delete' })).toHaveLength(1);
  });

  it('reports a failure to load rather than showing an empty list', async () => {
    vi.mocked(organizationService.list).mockRejectedValue(new Error('boom'));

    renderPage();

    await waitFor(() => expect(screen.getByText('boom')).toBeTruthy());
  });
});
