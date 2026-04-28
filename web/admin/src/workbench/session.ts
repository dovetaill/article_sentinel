import type { PersistedWorkbenchSession } from './types';

const WORKBENCH_SESSION_PREFIX = 'admin-workbench';

function getSessionStorage() {
  if (typeof window === 'undefined') {
    return null;
  }

  return window.sessionStorage;
}

export function getWorkbenchSessionKey(orgId: number) {
  return `${WORKBENCH_SESSION_PREFIX}:${orgId}`;
}

export function readWorkbenchSession(orgId: number): PersistedWorkbenchSession | null {
  const storage = getSessionStorage();
  if (!storage) {
    return null;
  }

  try {
    const raw = storage.getItem(getWorkbenchSessionKey(orgId));
    if (!raw) {
      return null;
    }

    return JSON.parse(raw) as PersistedWorkbenchSession;
  } catch {
    return null;
  }
}

export function writeWorkbenchSession(session: PersistedWorkbenchSession) {
  const storage = getSessionStorage();
  if (!storage) {
    return;
  }

  storage.setItem(getWorkbenchSessionKey(session.orgId), JSON.stringify(session));
}

export function clearWorkbenchSession(orgId: number) {
  const storage = getSessionStorage();
  if (!storage) {
    return;
  }

  storage.removeItem(getWorkbenchSessionKey(orgId));
}
