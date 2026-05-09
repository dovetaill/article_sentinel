import { describe, expect, it } from 'vitest';

import { LOGIN_ENTRY_PATH, normalizeAdminPath } from './auth';

describe('normalizeAdminPath', () => {
  it('maps the root path to /inspection/tasks', () => {
    expect(normalizeAdminPath('/')).toBe('/inspection/tasks');
  });

  it('keeps known business paths unchanged', () => {
    expect(normalizeAdminPath('/content/articles')).toBe('/content/articles');
  });
});

describe('LOGIN_ENTRY_PATH', () => {
  it('points to the fixed backend login entry', () => {
    expect(LOGIN_ENTRY_PATH).toBe('/auth/login');
  });
});
