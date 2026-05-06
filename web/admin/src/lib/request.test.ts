import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { apiRequest } from './request';

const LOGIN_ENTRY_PATH = '/auth/login';
const originalLocation = window.location;

describe('apiRequest', () => {
  let assignSpy: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    assignSpy = vi.fn();
    Object.defineProperty(window, 'location', {
      configurable: true,
      value: {
        ...originalLocation,
        assign: assignSpy,
      },
    });
  });

  afterEach(() => {
    Object.defineProperty(window, 'location', {
      configurable: true,
      value: originalLocation,
    });
    vi.restoreAllMocks();
  });

  it('unwraps the PureMux success envelope', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: true,
        json: async () => ({
          code: 0,
          message: 'ok',
          data: { id: 7, name: 'keyword' },
        }),
      }),
    );

    await expect(apiRequest<{ id: number; name: string }>('/api/v1/article-inspect/keywords/7')).resolves.toEqual({
      id: 7,
      name: 'keyword',
    });

    expect(fetch).toHaveBeenCalledWith(
      '/api/v1/article-inspect/keywords/7',
      expect.objectContaining({
        credentials: 'same-origin',
      }),
    );
  });

  it('throws a readable error when the API returns a failure envelope', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: false,
        json: async () => ({
          code: 400,
          message: 'invalid keyword input',
        }),
      }),
    );

    await expect(apiRequest('/api/v1/article-inspect/keywords')).rejects.toThrow('invalid keyword input');
  });

  it('redirects to the auth login entry when the API returns 401', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: false,
        status: 401,
        json: async () => ({
          code: 401,
          message: 'unauthorized',
        }),
      }),
    );

    await expect(apiRequest('/api/v1/auth/session')).rejects.toThrow('unauthorized');

    expect(assignSpy).toHaveBeenCalledWith(LOGIN_ENTRY_PATH);
  });
});
