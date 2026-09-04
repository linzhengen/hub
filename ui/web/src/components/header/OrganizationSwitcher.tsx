import React, { useSyncExternalStore } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { Select } from 'antd';
import { getActiveOrgId, setActiveOrgId, subscribeActiveOrgId } from '@/lib/active-org';
import { organizationService } from '@/services/organization';

/**
 * Picks which organization the user is acting in.
 *
 * "All organizations" is a real choice rather than a placeholder: it sends no
 * header, which is the question hub answered before organizations existed -
 * "may I, anywhere I hold access?". A user who belongs to one organization sees
 * the same thing either way, so the switcher only earns its place once they
 * belong to more than one, and it hides itself until then.
 *
 * The selection is read through useSyncExternalStore rather than mirrored into
 * component state. It lives outside React - in localStorage, read by the fetch
 * wrapper on every request - and copying it into state would mean two sources
 * of truth that have to be kept in step by an effect.
 */
const OrganizationSwitcher: React.FC = () => {
  const queryClient = useQueryClient();
  const activeOrgId = useSyncExternalStore(subscribeActiveOrgId, getActiveOrgId, () => '');

  const mine = useQuery({
    queryKey: ['organizations', 'mine'],
    queryFn: async () => {
      const response = await organizationService.listMine();
      // A stored id the user no longer has access to would narrow every request
      // to an organization they hold nothing in, and the screen would be empty
      // with no way to tell why. It is dropped here, where the list arrives,
      // rather than in an effect that would re-render to correct itself.
      const organizations = response.organizations ?? [];
      const stored = getActiveOrgId();
      if (stored && !organizations.some((organization) => organization.id === stored)) {
        setActiveOrgId('');
      }
      return response;
    },
  });

  const organizations = mine.data?.organizations ?? [];
  if (organizations.length < 2) return null;

  const onChange = (value: string) => {
    setActiveOrgId(value);
    // Every list is narrowed by the header, so anything already fetched was
    // fetched for a different question. Showing it under the new label would be
    // wrong rather than merely stale.
    void queryClient.invalidateQueries();
  };

  return (
    <Select
      aria-label="Organization"
      size="middle"
      className="min-w-44"
      value={activeOrgId}
      onChange={onChange}
      options={[
        { value: '', label: 'All organizations' },
        ...organizations.map((organization) => ({
          value: organization.id ?? '',
          label: organization.name ?? organization.slug ?? '',
        })),
      ]}
    />
  );
};

export default OrganizationSwitcher;
