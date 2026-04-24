import { ConfigProvider } from 'antd';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { OrgProvider } from '../../context/org-context';
import { listKeywords } from '../../services/keywords';
import { listOrgs } from '../../services/orgs';
import NewTaskPage from './new';
import { createTask } from '../../services/tasks';

const mockedNavigate = vi.fn();
const mockedMessageSuccess = vi.fn();
const mockedMessageError = vi.fn();

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom');
  return {
    ...actual,
    useNavigate: () => mockedNavigate
  };
});

vi.mock('antd', async () => {
  const actual = await vi.importActual<typeof import('antd')>('antd');
  return {
    ...actual,
    message: {
      ...actual.message,
      useMessage: () => [
        {
          success: mockedMessageSuccess,
          error: mockedMessageError
        },
        null
      ]
    }
  };
});

vi.mock('../../services/keywords', () => ({
  listKeywords: vi.fn()
}));

vi.mock('../../services/orgs', () => ({
  listOrgs: vi.fn()
}));

vi.mock('../../services/tasks', () => ({
  listTasks: vi.fn(),
  getTaskDetail: vi.fn(),
  createTask: vi.fn()
}));

const mockedListKeywords = vi.mocked(listKeywords);
const mockedListOrgs = vi.mocked(listOrgs);
const mockedCreateTask = vi.mocked(createTask);

describe('NewTaskPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockedListOrgs.mockResolvedValue([
      {
        id: 29,
        name: '一县一端',
        cateid: 0,
        enabled: true,
        sort: 1
      }
    ]);
    mockedListKeywords.mockResolvedValue({
      page: 1,
      pageSize: 20,
      total: 2,
      items: [
        { id: 7, name: 'spam', orgid: 29, category_id: 501, category_name: '政策红线', enabled: true },
        { id: 8, name: 'scam', orgid: 29, category_id: 502, category_name: '高频违规', enabled: true }
      ]
    } as never);
    mockedCreateTask.mockResolvedValue({ id: 88, task_no: 'inspect-20260420-88' } as never);
  });

  it('submits the simplified task form with the active org and selected rules', async () => {
    const user = userEvent.setup();

    render(
      <ConfigProvider>
        <OrgProvider>
          <NewTaskPage />
        </OrgProvider>
      </ConfigProvider>,
    );

    expect(screen.getByRole('heading', { name: '新建检测任务' })).toBeInTheDocument();
    expect(screen.queryByText('选择当前机构的规则集合并设定时间范围，任务提交后将在结果工作台继续处置。')).not.toBeInTheDocument();
    expect(screen.queryByText('执行提示')).not.toBeInTheDocument();
    expect(screen.queryByText('任务说明')).not.toBeInTheDocument();
    expect(screen.queryByText('范围建议')).not.toBeInTheDocument();
    expect(screen.queryByText('处理建议')).not.toBeInTheDocument();
    expect(screen.queryByText('填写巡检条件后即可发起一次新的异步扫描。')).not.toBeInTheDocument();
    expect(screen.queryByLabelText('文章编号')).not.toBeInTheDocument();
    expect(screen.queryByLabelText('标题检索')).not.toBeInTheDocument();
    expect(screen.queryByRole('switch', { name: '是否检索正文' })).not.toBeInTheDocument();
    expect(screen.getByLabelText('发布时间起')).toHaveAttribute('readonly');
    expect(screen.getByLabelText('发布时间止')).toHaveAttribute('readonly');
    expect(screen.getByDisplayValue('一县一端')).toBeDisabled();
    await waitFor(() => {
      expect(mockedListKeywords).toHaveBeenCalledWith(expect.objectContaining({
        orgid: 29,
        enabled: true
      }));
    });

    const ruleSelect = await screen.findByRole('combobox', { name: '规则选择' });
    await user.click(ruleSelect);
    await user.click(await screen.findByText('政策红线 / spam'));
    await waitFor(() => {
      expect(screen.getAllByText('政策红线 / spam').length).toBeGreaterThan(1);
    });
    await user.keyboard('{Escape}');
    await user.click(screen.getByRole('button', { name: '提交任务' }));

    await waitFor(() => {
      expect(mockedCreateTask).toHaveBeenCalledWith(expect.objectContaining({
        orgid: 29,
        keyword_ids: [7]
      }));
    });

    const payload = mockedCreateTask.mock.calls[0]?.[0];
    expect(payload).not.toHaveProperty('article_id');
    expect(payload).not.toHaveProperty('title_like');
    expect(payload).not.toHaveProperty('include_body');
  }, 10_000);

  it('prevents duplicate submits and redirects about 1 second after showing success', async () => {
    let resolveCreateTask: ((value: { id: number; task_no: string }) => void) | undefined;
    mockedCreateTask.mockImplementationOnce(() => new Promise((resolve) => {
      resolveCreateTask = resolve;
    }) as never);

    const user = userEvent.setup();

    render(
      <ConfigProvider>
        <OrgProvider>
          <NewTaskPage />
        </OrgProvider>
      </ConfigProvider>,
    );

    const ruleSelect = await screen.findByRole('combobox', { name: '规则选择' });
    await user.click(ruleSelect);
    await user.click(await screen.findByText('政策红线 / spam'));
    await waitFor(() => {
      expect(screen.getAllByText('政策红线 / spam').length).toBeGreaterThan(1);
    });
    await user.keyboard('{Escape}');

    const submitButton = screen.getByRole('button', { name: '提交任务' });
    await user.click(submitButton);
    await user.click(submitButton);

    await waitFor(() => {
      expect(mockedCreateTask).toHaveBeenCalledTimes(1);
    });
    expect(submitButton).toBeDisabled();

    resolveCreateTask?.({ id: 88, task_no: 'inspect-20260420-88' });

    await waitFor(() => {
      expect(mockedMessageSuccess).toHaveBeenCalledWith('检测任务已提交');
    });
    expect(mockedNavigate).not.toHaveBeenCalled();

    await waitFor(() => {
      expect(mockedNavigate).toHaveBeenCalledWith('/tasks');
    }, { timeout: 1_500 });
  }, 10_000);
});
