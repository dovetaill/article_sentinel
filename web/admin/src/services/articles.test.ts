import { beforeEach, describe, expect, it, vi } from 'vitest';

import { listArticles } from './articles';

const { mockedApiRequest, mockedListResults } = vi.hoisted(() => ({
  mockedApiRequest: vi.fn(),
  mockedListResults: vi.fn()
}));

vi.mock('../lib/request', () => ({
  apiRequest: mockedApiRequest
}));

vi.mock('./results', () => ({
  listResults: mockedListResults
}));

describe('listArticles', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockedApiRequest.mockResolvedValue({
      page: 1,
      page_size: 20,
      total: 1,
      items: [
        {
          id: 501,
          orgid: 29,
          title: '县域融媒今日要闻',
          state: 9,
          latest_risk_level: 'high',
          latest_task_id: 208
        }
      ]
    });
    mockedListResults.mockResolvedValue({ page: 0, pageSize: 0, total: 0, items: [] });
  });

  it('requests the real article-center endpoint instead of aggregating result rows', async () => {
    const result = await listArticles({ orgid: 29, page: 1, pageSize: 20 } as never);

    expect(mockedApiRequest).toHaveBeenCalledWith('/api/v1/article-inspect/articles?orgid=29&page=1&page_size=20');
    expect(mockedListResults).not.toHaveBeenCalled();
    expect(result.items[0]).toMatchObject({
      id: 501,
      title: '县域融媒今日要闻',
      latest_risk_level: 'high',
      latest_task_id: 208
    });
  });
});
