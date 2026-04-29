import { ConfigProvider } from 'antd';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import type { PropsWithChildren } from 'react';
import { MemoryRouter, Route, Routes, useLocation } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { OrgProvider } from '../../context/org-context';
import { WorkbenchProvider } from '../../workbench/provider';
import { useWorkbench } from '../../workbench/use-workbench';
import TaskDetailPage from './detail';
import { listOperationLogs } from '../../services/logs';
import { listResults } from '../../services/results';
import { getTaskDetail } from '../../services/tasks';

vi.mock('../../context/org-context', () => ({
  OrgProvider: ({ children }: PropsWithChildren) => <>{children}</>,
  useOrgContext: () => ({
    activeOrgId: 29,
    activeOrgName: '一县一端',
    isLoading: false
  })
}));

vi.mock('../../services/tasks', () => ({
  listTasks: vi.fn(),
  getTaskDetail: vi.fn(),
  createTask: vi.fn()
}));

vi.mock('../../services/results', () => ({
  listResults: vi.fn(),
  getResultDetail: vi.fn(),
  batchOfflineResults: vi.fn(),
  getArticleRectify: vi.fn(),
  rectifyArticle: vi.fn()
}));

vi.mock('../../services/logs', () => ({
  listOperationLogs: vi.fn(),
  listArticleOperationLogs: vi.fn(),
  listArticleFieldChanges: vi.fn()
}));

const mockedGetTaskDetail = vi.mocked(getTaskDetail);
const mockedListResults = vi.mocked(listResults);
const mockedListOperationLogs = vi.mocked(listOperationLogs);

function LocationProbe() {
  const location = useLocation();

  return <pre data-testid="location-probe">{`${location.pathname}${location.search}`}</pre>;
}

function renderPage(initialEntries: string[] = ['/tasks/77']) {
  return render(
    <ConfigProvider>
      <MemoryRouter initialEntries={initialEntries}>
        <OrgProvider>
          <WorkbenchProvider>
            <Routes>
              <Route path="/tasks/:taskId" element={<TaskDetailPage />} />
              <Route path="/articles/:articleId" element={<div>文稿详情探针</div>} />
            </Routes>
            <LocationProbe />
          </WorkbenchProvider>
        </OrgProvider>
      </MemoryRouter>
    </ConfigProvider>,
  );
}

function WorkbenchCycleControls() {
  const { activateTab } = useWorkbench();

  return (
    <>
      <button type="button" onClick={() => activateTab('/tasks')}>
        切换到任务列表
      </button>
      <button type="button" onClick={() => activateTab('task:77')}>
        切回任务详情
      </button>
    </>
  );
}

function renderWorkbenchPage(initialEntries: string[] = ['/tasks/77']) {
  return render(
    <ConfigProvider>
      <MemoryRouter initialEntries={initialEntries}>
        <OrgProvider>
          <WorkbenchProvider>
            <WorkbenchCycleControls />
            <Routes>
              <Route path="/tasks/:taskId" element={<TaskDetailPage />} />
              <Route path="/articles/:articleId" element={<div>文稿详情探针</div>} />
              <Route path="/tasks" element={<div>任务列表探针</div>} />
            </Routes>
            <LocationProbe />
          </WorkbenchProvider>
        </OrgProvider>
      </MemoryRouter>
    </ConfigProvider>,
  );
}

describe('TaskDetailPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockedGetTaskDetail.mockResolvedValue({
      id: 77,
      orgid: 29,
      task_no: 'inspect-20260420-01',
      status: 'running',
      total_scanned: 42,
      hit_articles: 4,
      hit_count: 8,
      creator_name: 'operator',
      created_at: '2026-04-20 12:00:00',
      rule_snapshot: '{"keywords":["spam","scam"]}',
      request_snapshot: '{"include_body":true,"title_like":"Alpha"}'
    } as never);

    mockedListResults.mockResolvedValue({
      page: 1,
      pageSize: 20,
      total: 1,
      items: [
        {
          id: 11,
          orgid: 29,
          task_id: 77,
          article_id: 501,
          article_title: 'Spam alert',
          risk_level: 'high',
          suggest_action: 'offline',
          disposition_status: 'pending',
          hit_count: 3,
          preview_field_name: 'title',
          preview_keyword_text: 'spam',
          preview_matched_text: 'spam',
          preview_snippet: 'This spam alert keeps repeating',
          extra_hit_count: 2
        }
      ]
    } as never);

    mockedListOperationLogs.mockResolvedValue({
      page: 1,
      pageSize: 20,
      total: 1,
      items: [
        {
          id: 9,
          orgid: 29,
          article_id: 501,
          task_id: 77,
          operation_type: 'offline',
          before_state: 'online',
          after_state: 'offline',
          summary: 'Task reviewed by auditor',
          operator_name: 'auditor',
          created_at: '2026-04-20 16:00:00'
        }
      ]
    } as never);
  });

  it('renders the full task detail workspace with tabs and linked hit results', async () => {
    const user = userEvent.setup();

    renderPage();

    expect((await screen.findAllByText('inspect-20260420-01')).length).toBeGreaterThan(0);
    expect(screen.getByRole('link', { name: '返回任务列表' })).toHaveAttribute('href', '/tasks');
    expect(screen.queryByText('集中查看任务配置、命中结果与关联日志。')).not.toBeInTheDocument();
    expect(screen.queryByText('查看当前批次的规则快照、执行摘要与命中概况。')).not.toBeInTheDocument();
    expect(screen.queryByText('帮助快速判断这批任务是否还需要继续跟进。')).not.toBeInTheDocument();
    expect(screen.queryByText('继续进入稿件和结果工作台。')).not.toBeInTheDocument();
    expect(screen.queryByText('查看命中结果、规则快照、请求参数与关联日志。')).not.toBeInTheDocument();
    expect(screen.getByRole('tab', { name: '命中结果' })).toBeInTheDocument();
    expect(screen.getByRole('tab', { name: '规则快照' })).toBeInTheDocument();
    expect(screen.getByRole('tab', { name: '请求快照' })).toBeInTheDocument();
    expect(screen.getByRole('tab', { name: '关联日志' })).toBeInTheDocument();
    expect(screen.getByText('标题')).toBeInTheDocument();
    expect(screen.getAllByText('spam').length).toBeGreaterThan(0);
    expect(screen.getByText((_, element) => element?.textContent === 'This spam alert keeps repeating')).toBeInTheDocument();
    expect(screen.getByText('另有 2 条命中')).toBeInTheDocument();

    await user.click(screen.getByRole('tab', { name: '请求快照' }));
    expect(screen.getByText(/title_like/i)).toBeInTheDocument();
    expect(screen.getByText(/alpha/i)).toBeInTheDocument();
    expect(screen.queryByText(/include_body/i)).not.toBeInTheDocument();

    await user.click(screen.getByRole('tab', { name: '关联日志' }));
    expect(screen.getAllByText(/task reviewed by auditor/i).length).toBeGreaterThan(0);

    await user.click(screen.getByRole('link', { name: 'Spam alert' }));
    await waitFor(() => {
      expect(screen.getByTestId('location-probe')).toHaveTextContent('/articles/501?return_to=%2Ftasks%2F77');
    });
  });

  it('restores the last active local tab after a workbench deactivate/reactivate cycle', async () => {
    const user = userEvent.setup();

    renderWorkbenchPage();

    expect((await screen.findAllByText('inspect-20260420-01')).length).toBeGreaterThan(0);

    await user.click(screen.getByRole('tab', { name: '请求快照' }));
    expect(screen.getByText(/title_like/i)).toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: '切换到任务列表' }));
    expect(await screen.findByText('任务列表探针')).toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: '切回任务详情' }));
    expect((await screen.findAllByText('inspect-20260420-01')).length).toBeGreaterThan(0);
    expect(screen.getByText(/title_like/i)).toBeInTheDocument();
  });
});
