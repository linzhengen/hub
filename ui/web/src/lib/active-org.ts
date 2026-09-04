/**
 * The organization the user is currently acting in.
 *
 * Unlike the access token this is kept in localStorage: it is a preference
 * rather than a credential - the worst an attacker gains by reading or setting
 * it is a narrower or wider view that the server still authorizes from scratch
 * on every request. Surviving a reload matters, because a switcher that resets
 * to "everywhere" whenever the page reloads is one nobody trusts.
 *
 * An empty value means the user has not chosen, and no header is sent. That is
 * the question hub answered before organizations existed - "may I, anywhere I
 * hold access?" - so a user who never touches the switcher sees exactly what
 * they saw before.
 */

const STORAGE_KEY = 'hub.activeOrgId';

type Listener = (orgId: string) => void;

const listeners = new Set<Listener>();

export function getActiveOrgId(): string {
  try {
    return localStorage.getItem(STORAGE_KEY) ?? '';
  } catch {
    // A browser with site data blocked still has to be able to use hub; it
    // simply gets the unnarrowed view.
    return '';
  }
}

export function setActiveOrgId(orgId: string): void {
  try {
    if (orgId) {
      localStorage.setItem(STORAGE_KEY, orgId);
    } else {
      localStorage.removeItem(STORAGE_KEY);
    }
  } catch {
    // Ignored for the same reason as above.
  }
  for (const listener of listeners) listener(orgId);
}

/** Subscribes to changes, and returns the unsubscribe function. */
export function subscribeActiveOrgId(listener: Listener): () => void {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}
