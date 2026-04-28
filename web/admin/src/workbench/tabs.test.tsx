import { ConfigProvider } from 'antd';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { useEffect } from 'react';
import { MemoryRouter } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { OrgProvider } from '../context/org-context';
import { listOrgs } from '../services/orgs';
import { WorkbenchProvider } from './provider';
import { WorkbenchTabs } from './tabs';
import { useWorkbench } from './use-workbench';

vi.mock('../services/orgs', () => ({
  listOrgs: vi.fn()
}));

const mockedListOrgs = vi.mocked(listOrgs);

function SeedWorkbenchTabs() {
  const { openTab } = useWorkbench();

  useEffect(() => {
    openTab('/articles');
    openTab('/articles');
    openTab('/tasks/new');
  }, [openTab]);

  return <WorkbenchTabs />;
}

describe('WorkbenchTabs', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockedListOrgs.mockResolvedValue([
      { id: 29, name: '一县一端', cate_id: 0, enabled: true, sort: 1 }
    ]);
  });

  it('renders one tab per open descriptor and keeps list tabs deduplicated', async () => {
    render(
      <ConfigProvider>
        <MemoryRouter initialEntries={['/tasks']}>
          <OrgProvider>
            <WorkbenchProvider>
              <SeedWorkbenchTabs />
            </WorkbenchProvider>
          </OrgProvider>
        </MemoryRouter>
      </ConfigProvider>,
    );

    expect(await screen.findByRole('tab', { name: '检测任务' })).toBeInTheDocument();
    expect(screen.getAllByRole('tab', { name: '文稿中心' })).toHaveLength(1);
    expect(screen.getByRole('tab', { name: '新建任务' })).toBeInTheDocument();
  });

  it('activates an existing tab and exposes bulk close actions from the right-side menu', async () => {
    const user = userEvent.setup();

    render(
      <ConfigProvider>
        <MemoryRouter initialEntries={['/tasks']}>
          <OrgProvider>
            <WorkbenchProvider>
              <SeedWorkbenchTabs />
            </WorkbenchProvider>
          </OrgProvider>
        </MemoryRouter>
      </ConfigProvider>,
    );

    await user.click(await screen.findByRole('tab', { name: '文稿中心' }));

    await waitFor(() => {
      expect(screen.getByRole('tab', { name: '文稿中心' })).toHaveAttribute('aria-selected', 'true');
    });

    await user.click(screen.getByRole('button', { name: '标签操作' }));

    expect(await screen.findByRole('menuitem', { name: '关闭当前' })).toBeInTheDocument();
    expect(screen.getByRole('menuitem', { name: '关闭其他' })).toBeInTheDocument();
    expect(screen.getByRole('menuitem', { name: '关闭左侧' })).toBeInTheDocument();
    expect(screen.getByRole('menuitem', { name: '关闭右侧' })).toBeInTheDocument();
    expect(screen.getByRole('menuitem', { name: '关闭全部' })).toBeInTheDocument();
  });
});
