import type { GroupMembership } from '@/services/user';
import type { RoleGrant } from '@/services/group';

/**
 * How a term reads.
 *
 * Wording matched to the access request screen on purpose: "until Friday" and
 * "permanently" mean the same thing wherever they appear, and a second phrasing
 * would only make a reader wonder whether it meant something else.
 */
export const formatTerm = (expiresAt?: string) =>
  expiresAt ? `until ${new Date(expiresAt).toLocaleString()}` : 'permanently';

/** expiresSoon is what the console highlights: within a week, but not already gone. */
const SOON_MS = 7 * 24 * 60 * 60 * 1000;

export const expiresSoon = (expiresAt?: string, now: number = Date.now()): boolean => {
  if (!expiresAt) return false;
  const at = new Date(expiresAt).getTime();
  return at > now && at - now <= SOON_MS;
};

/**
 * groupIdsOf and roleIdsOf strip a set of grants back to ids.
 *
 * The operations that replace a whole set - editing a user, editing a group -
 * take ids and nothing else, so they need the list without the terms. A grant
 * with a term is made by the add operations, not by re-stating the set.
 */
export const groupIdsOf = (memberships?: GroupMembership[]): string[] =>
  (memberships ?? []).map((m) => m.groupId ?? '').filter(Boolean);

export const roleIdsOf = (grants?: RoleGrant[]): string[] =>
  (grants ?? []).map((g) => g.roleId ?? '').filter(Boolean);

/** expiryOf finds the term attached to one id, when there is one. */
export const groupExpiryOf = (
  memberships: GroupMembership[] | undefined,
  groupId: string,
): string | undefined => (memberships ?? []).find((m) => m.groupId === groupId)?.expiresAt;

export const roleExpiryOf = (
  grants: RoleGrant[] | undefined,
  roleId: string,
): string | undefined => (grants ?? []).find((g) => g.roleId === roleId)?.expiresAt;
