import { ConfigProvider } from 'antd';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';

import ResultDetailDrawer from './detail';
import { getResultDetail } from '../../services/results';

vi.mock('../../services/results', () => ({
  listResults: vi.fn(),
  getResultDetail: vi.fn(),
  batchOfflineResults: vi.fn(),
  getArticleRectify: vi.fn(),
  rectifyArticle: vi.fn()
}));

const mockedGetResultDetail = vi.mocked(getResultDetail);

describe('ResultDetailDrawer', () => {
  it('shows hits and operation history inside the drawer', async () => {
    const user = userEvent.setup();

    mockedGetResultDetail.mockResolvedValue({
      id: 11,
      orgid: 100,
      task_id: 77,
      article_id: 501,
      article_title: 'Spam alert',
      risk_level: 'high',
      disposition_status: 'pending',
      suggest_action: 'offline',
      hit_count: 3,
      article_state: 9,
      article_body: '<p>spam body</p>',
      hits: [
        {
          id: 1,
          field_name: 'title',
          keyword_text: 'spam',
          snippet: 'spam alert keeps repeating',
          matched_text: 'spam',
          risk_level: 'high'
        }
      ],
      operation_logs: [
        {
          id: 9,
          operation_type: 'offline',
          summary: 'Article sent offline',
          operator_name: 'auditor',
          created_at: '2026-04-20 16:00:00'
        }
      ],
      field_changes: []
    } as never);

    render(
      <ConfigProvider>
        <ResultDetailDrawer open resultId={11} orgid={100} onClose={() => undefined} />
      </ConfigProvider>,
    );

    const snippets = await screen.findAllByText((_, element) => element?.textContent === 'spam alert keeps repeating');
    expect(snippets[0]).toBeInTheDocument();
    expect(screen.getAllByText('高风险').length).toBeGreaterThan(0);
    await user.click(screen.getByRole('tab', { name: '操作记录' }));
    expect(screen.getByText(/article sent offline/i)).toBeInTheDocument();
    expect(screen.getByText(/auditor/i)).toBeInTheDocument();
  });
});
