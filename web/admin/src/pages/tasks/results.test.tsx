import { ConfigProvider } from 'antd';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { useEffect } from 'react';
import { MemoryRouter, Route, Routes, useLocation } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { OrgProvider } from '../../context/org-context';
import { listOperationLogs } from '../../services/logs';
import { listOrgs } from '../../services/orgs';
import {
  batchIgnoreResults,
  batchOfflineResults,
  batchProcessResults,
  listResults
} from '../../services/results';
import { getTaskDetail } from '../../services/tasks';
import { WorkbenchProvider } from '../../workbench/provider';
import { useWorkbench } from '../../workbench/use-workbench';
import TaskResultsPage from './results';

vi.mock('../../services/orgs', () => ({
  listOrgs: vi.fn()
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
  batchIgnoreResults: vi.fn(),
  batchProcessResults: vi.fn(),
  getArticleRectify: vi.fn(),
  rectifyArticle: vi.fn()
}));

vi.mock('../../services/logs', () => ({
  listOperationLogs: vi.fn(),
  listArticleOperationLogs: vi.fn(),
  listArticleFieldChanges: vi.fn()
}));

const mockedListOrgs = vi.mocked(listOrgs);
const mockedGetTaskDetail = vi.mocked(getTaskDetail);
const mockedListResults = vi.mocked(listResults);
const mockedBatchOfflineResults = vi.mocked(batchOfflineResults);
const mockedBatchIgnoreResults = vi.mocked(batchIgnoreResults);
const mockedBatchProcessResults = vi.mocked(batchProcessResults);
const mockedListOperationLogs = vi.mocked(listOperationLogs);
let mockScrollY = 0;

function LocationProbe() {
  const location = useLocation();

  return <pre data-testid="location-probe">{`${location.pathname}${location.search}`}</pre>;
}

function renderPage(initialEntries: string[] = ['/tasks/77/results']) {
  return render(
    <ConfigProvider>
      <MemoryRouter initialEntries={initialEntries}>
        <OrgProvider>
          <WorkbenchProvider>
            <Routes>
              <Route path="/tasks/:taskId/results" element={<TaskResultsPage />} />
              <Route path="/articles/:articleId" element={<div>文稿详情探针</div>} />
              <Route path="/articles/:articleId/rectify" element={<div>整改页探针</div>} />
            </Routes>
            <LocationProbe />
          </WorkbenchProvider>
        </OrgProvider>
      </MemoryRouter>
    </ConfigProvider>,
  );
}

function TaskListProbe() {
  useEffect(() => {
    window.scrollTo(0, 0);
  }, []);

  return <div>任务列表探针</div>;
}

function WorkbenchCycleControls() {
  const { activateTab } = useWorkbench();

  return (
    <>
      <button type="button" onClick={() => activateTab('/tasks')}>
        切换到任务列表
      </button>
      <button type="button" onClick={() => activateTab('task:77:results')}>
        切回任务结果
      </button>
    </>
  );
}

function renderWorkbenchPage(initialEntries: string[] = ['/tasks/77/results']) {
  return render(
    <ConfigProvider>
      <MemoryRouter initialEntries={initialEntries}>
        <OrgProvider>
          <WorkbenchProvider>
            <WorkbenchCycleControls />
            <Routes>
              <Route path="/tasks/:taskId/results" element={<TaskResultsPage />} />
              <Route path="/articles/:articleId" element={<div>文稿详情探针</div>} />
              <Route path="/articles/:articleId/rectify" element={<div>整改页探针</div>} />
              <Route path="/tasks" element={<TaskListProbe />} />
            </Routes>
            <LocationProbe />
          </WorkbenchProvider>
        </OrgProvider>
      </MemoryRouter>
    </ConfigProvider>,
  );
}

describe('TaskResultsPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockScrollY = 0;
    Object.defineProperty(window, 'scrollY', {
      configurable: true,
      get: () => mockScrollY
    });
    window.scrollTo = vi.fn((xOrOptions?: number | ScrollToOptions, y?: number) => {
      if (xOrOptions === undefined) {
        mockScrollY = 0;
        return;
      }

      if (typeof xOrOptions === 'object') {
        mockScrollY = Number(xOrOptions.top ?? 0);
        return;
      }

      mockScrollY = Number(y ?? 0);
    });
    mockedListOrgs.mockResolvedValue([
      {
        id: 29,
        name: '一县一端',
        cate_id: 0,
        enabled: true,
        sort: 1
      }
    ]);
    mockedGetTaskDetail.mockResolvedValue({
      id: 77,
      orgid: 29,
      task_no: 'inspect-20260420-01',
      status: 'running',
      total_scanned: 42,
      hit_articles: 4,
      hit_count: 8,
      creator_name: 'operator',
      created_at: '2026-04-20 12:00:00'
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
    mockedBatchOfflineResults.mockResolvedValue({
      action_id: 11,
      target_count: 1,
      success_count: 1,
      fail_count: 0,
      skip_count: 0,
      status: 'success',
      action_type: 'offline'
    } as never);
    mockedBatchIgnoreResults.mockResolvedValue({
      action_id: 12,
      target_count: 1,
      success_count: 1,
      fail_count: 0,
      skip_count: 0,
      status: 'success',
      action_type: 'batch_ignore'
    } as never);
    mockedBatchProcessResults.mockResolvedValue({
      action_id: 13,
      target_count: 1,
      success_count: 1,
      fail_count: 0,
      skip_count: 0,
      status: 'success',
      action_type: 'batch_process'
    } as never);
  });

  it('renders the dedicated task-results workspace with result actions, batch actions, and log section', async () => {
    const user = userEvent.setup();

    renderPage();

    expect(await screen.findByText('inspect-20260420-01')).toBeInTheDocument();
    expect(screen.queryByText('围绕单个任务查看命中结果、执行摘要和处置日志。')).not.toBeInTheDocument();
    expect(screen.queryByText('查看当前任务的执行概况与责任信息。')).not.toBeInTheDocument();
    expect(screen.queryByText('在当前任务上下文中完成单条或批量处置。')).not.toBeInTheDocument();
    expect(screen.queryByText('跟踪该任务下的处置动作和状态变化。')).not.toBeInTheDocument();
    expect(screen.getByText('文稿ID')).toBeInTheDocument();
    expect(screen.getByText('Spam alert')).toBeInTheDocument();
    expect(screen.getByText('501')).toBeInTheDocument();
    expect(screen.queryByText('#501')).not.toBeInTheDocument();
    expect(screen.getByText('标题')).toBeInTheDocument();
    expect(screen.getAllByText('spam').length).toBeGreaterThan(0);
    expect(screen.getByText((_, element) => element?.textContent === 'This spam alert keeps repeating')).toBeInTheDocument();
    expect(screen.getByText('另有 2 条命中')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '下线处置' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '批量忽略' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '批量标记已处理' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '批量下线处置' })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: '关联日志' })).toBeInTheDocument();
    expect(screen.getByText('Task reviewed by auditor')).toBeInTheDocument();

    await user.click(screen.getByRole('link', { name: '进入整改' }));

    await waitFor(() => {
      expect(screen.getByTestId('location-probe')).toHaveTextContent('/articles/501/rectify?return_to=%2Ftasks%2F77%2Fresults&task_id=77&result_id=11');
    });
  });

  it('opens article detail from the task-results workspace through the workbench route', async () => {
    const user = userEvent.setup();

    renderPage();

    expect(await screen.findByText('Spam alert')).toBeInTheDocument();

    await user.click(screen.getByRole('link', { name: '查看详情' }));

    await waitFor(() => {
      expect(screen.getByTestId('location-probe')).toHaveTextContent('/articles/501?return_to=%2Ftasks%2F77%2Fresults');
    });
  });

  it('restores result-page scroll position after a workbench deactivate/reactivate cycle', async () => {
    const user = userEvent.setup();

    renderWorkbenchPage();

    expect(await screen.findByText('Spam alert')).toBeInTheDocument();

    mockScrollY = 360;
    fireEvent.scroll(window);

    await user.click(screen.getByRole('button', { name: '切换到任务列表' }));
    expect(await screen.findByText('任务列表探针')).toBeInTheDocument();
    expect(mockScrollY).toBe(0);

    await user.click(screen.getByRole('button', { name: '切回任务结果' }));
    expect(await screen.findByText('Spam alert')).toBeInTheDocument();

    await waitFor(() => {
      expect(mockScrollY).toBe(360);
    });
  });
});
