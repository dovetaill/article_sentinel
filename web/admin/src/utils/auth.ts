export const DEFAULT_ADMIN_PATH = '/inspection/tasks';
export const LOGIN_ENTRY_PATH = '/auth/login';

export function normalizeAdminPath(pathname: string) {
  return pathname === '/' ? DEFAULT_ADMIN_PATH : pathname;
}

export function redirectToFixedLogin() {
  if (typeof window !== 'undefined' && typeof window.location?.assign === 'function') {
    window.location.assign(LOGIN_ENTRY_PATH);
  }
}
