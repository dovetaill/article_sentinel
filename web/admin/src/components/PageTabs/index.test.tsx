import userEvent from '@testing-library/user-event';
import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import PageTabs from './index';

describe('PageTabs', () => {
  it('renders a scrollable tab rail with close and more actions', async () => {
    const user = userEvent.setup();
    const onActivate = vi.fn();
    const onClose = vi.fn();
    const onCloseOthers = vi.fn();
    const onCloseAll = vi.fn();

    const { container } = render(
      <PageTabs
        state={{
          orgId: 29,
          activeKey: '/inspection/tasks',
          tabs: [
            {
              key: '/inspection/tasks',
              pathname: '/inspection/tasks',
              search: '',
              title: '检测任务',
              closable: true,
              menuKey: '/inspection/tasks'
            },
            {
              key: '/inspection/results',
              pathname: '/inspection/results',
              search: '',
              title: '风险结果',
              closable: true,
              menuKey: '/inspection/results'
            }
          ]
        }}
        onActivate={onActivate}
        onClose={onClose}
        onCloseOthers={onCloseOthers}
        onCloseAll={onCloseAll}
        onRefresh={() => {}}
      />
    );

    expect(container.querySelector('.admin-page-tabs__scroll')).toBeInTheDocument();
    expect(container.querySelector('.admin-page-tabs__item--active')).toHaveTextContent('检测任务');

    await user.click(screen.getByRole('button', { name: '风险结果' }));
    expect(onActivate).toHaveBeenCalledWith('/inspection/results');

    await user.click(screen.getByLabelText('关闭 风险结果'));
    expect(onClose).toHaveBeenCalledWith('/inspection/results');

    await user.click(screen.getByRole('button', { name: '更多标签操作' }));
    expect(screen.getByText('关闭当前')).toBeInTheDocument();
    expect(screen.getByText('关闭其他')).toBeInTheDocument();
    expect(screen.getByText('关闭全部')).toBeInTheDocument();
  });
});
