import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import OrganizationSwitcher from '@/components/header/OrganizationSwitcher';
import { getActiveOrgId, setActiveOrgId } from '@/lib/active-org';
import { organizationService } from '@/services/organization';
import type { Organization } from '@/services/organization';

vi.mock('@/services/organization', () => ({
  organizationService: { listMine: vi.fn() },
}));

const acme: Organization = { id: 'org-acme', name: 'Acme Inc.', slug: 'acme' };
const globex: Organization = { id: 'org-globex', name: 'Globex', slug: 'globex' };

const renderSwitcher = () => {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <OrganizationSwitcher />
    </QueryClientProvider>,
  );
};

beforeEach(() => {
  vi.clearAllMocks();
  setActiveOrgId('');
});

describe('OrganizationSwitcher', () => {
  // A switcher with one option is a control that cannot change anything, and
  // the narrowed and unnarrowed views are identical for such a user.
  it('stays hidden while the user belongs to a single organization', async () => {
    vi.mocked(organizationService.listMine).mockResolvedValue({ organizations: [acme] });

    const { container } = renderSwitcher();

    await waitFor(() => expect(organizationService.listMine).toHaveBeenCalled());
    expect(container.querySelector('.ant-select')).toBeNull();
  });

  it('offers each organization plus the unnarrowed view', async () => {
    vi.mocked(organizationService.listMine).mockResolvedValue({
      organizations: [acme, globex],
    });

    renderSwitcher();

    expect(await screen.findByText('All organizations')).toBeTruthy();
  });

  // A stored organization the user has been removed from would narrow every
  // request to somewhere they hold nothing, and the screens would simply be
  // empty with nothing saying why.
  it('drops a stored organization the user no longer belongs to', async () => {
    setActiveOrgId('org-gone');
    vi.mocked(organizationService.listMine).mockResolvedValue({
      organizations: [acme, globex],
    });

    renderSwitcher();

    await waitFor(() => expect(getActiveOrgId()).toBe(''));
  });

  it('keeps a stored organization the user still belongs to', async () => {
    setActiveOrgId('org-globex');
    vi.mocked(organizationService.listMine).mockResolvedValue({
      organizations: [acme, globex],
    });

    renderSwitcher();

    await waitFor(() => expect(organizationService.listMine).toHaveBeenCalled());
    expect(getActiveOrgId()).toBe('org-globex');
  });
});
