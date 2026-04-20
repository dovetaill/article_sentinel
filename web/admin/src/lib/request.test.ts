import { afterEach, describe, expect, it, vi } from 'vitest';

import { apiRequest } from './request';

describe('apiRequest', () => {
  afterEach(() => {
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
});
