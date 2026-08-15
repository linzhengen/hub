import { describe, expect, it } from 'vitest';
import { expiresSoon, formatTerm, groupIdsOf, groupExpiryOf, roleIdsOf, roleExpiryOf } from '@/lib/grants';

describe('formatTerm', () => {
  // The same two phrasings as the access request screen. A second wording for
  // the same fact only makes a reader wonder whether it means something else.
  it('reads a grant with no term as permanent', () => {
    expect(formatTerm(undefined)).toBe('permanently');
  });

  it('reads a grant with a term as ending', () => {
    expect(formatTerm('2026-08-22T09:00:00Z')).toMatch(/^until /);
  });
});

describe('expiresSoon', () => {
  const now = new Date('2026-08-15T00:00:00Z').getTime();
  const at = (iso: string) => expiresSoon(iso, now);

  it('is false for a grant that does not end', () => {
    expect(expiresSoon(undefined, now)).toBe(false);
  });

  it('is true within the week', () => {
    expect(at('2026-08-18T00:00:00Z')).toBe(true);
  });

  it('is false further out', () => {
    expect(at('2026-09-18T00:00:00Z')).toBe(false);
  });

  // Already gone is not "soon": there is nothing left to act before.
  it('is false once it has passed', () => {
    expect(at('2026-08-14T00:00:00Z')).toBe(false);
  });
});

describe('stripping grants back to ids', () => {
  it('takes the ids out of memberships and grants', () => {
    expect(groupIdsOf([{ groupId: 'g1' }, { groupId: 'g2', expiresAt: 'x' }])).toEqual(['g1', 'g2']);
    expect(roleIdsOf([{ roleId: 'r1' }])).toEqual(['r1']);
  });

  it('drops entries with no id rather than emitting an empty one', () => {
    expect(groupIdsOf([{ expiresAt: 'x' }])).toEqual([]);
    expect(groupIdsOf(undefined)).toEqual([]);
  });
});

describe('finding the term attached to one id', () => {
  it('returns it when there is one, and nothing when there is not', () => {
    const memberships = [{ groupId: 'g1', expiresAt: '2026-08-22T09:00:00Z' }, { groupId: 'g2' }];
    expect(groupExpiryOf(memberships, 'g1')).toBe('2026-08-22T09:00:00Z');
    expect(groupExpiryOf(memberships, 'g2')).toBeUndefined();
    expect(groupExpiryOf(memberships, 'missing')).toBeUndefined();

    expect(roleExpiryOf([{ roleId: 'r1', expiresAt: 'x' }], 'r1')).toBe('x');
  });
});
