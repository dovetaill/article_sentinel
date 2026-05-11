import userEvent from '@testing-library/user-event';
import { render, screen, waitFor, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import BasicLayout from './BasicLayout';

const { mockedLogout, mockedNavigate, mockedProLayout } = vi.hoisted(() => ({
  mockedLogout: vi.fn(),
  mockedNavigate: vi.fn(),
  mockedProLayout: vi.fn()
}));

vi.mock('@ant-design/pro-layout', () => ({
  default: (props: Record<string, unknown> & { children?: React.ReactNode }) => {
    mockedProLayout(props);

    return (
      <div data-testid="pro-layout" data-layout={String(props.layout)} data-nav-theme={String(props.navTheme)}>
        <aside data-testid="pro-layout-sidebar">侧边栏</aside>
        <div>{props.children}</div>
      </div>
    );
  }
}));

vi.mock('@umijs/max', () => ({
  Outlet: () => <div data-testid="layout-outlet">页面内容</div>,
  useLocation: () => ({
    pathname: '/inspection/tasks',
    search: ''
  }),
  useModel: () => ({
    initialState: {
      currentUser: {
        orgid: 29,
        orgname: '示例机构',
        nickname: '测试用户'
      },
      currentOrgId: 29,
      currentOrgName: '示例机构'
    }
  }),
  useNavigate: () => mockedNavigate
}));

vi.mock('@/services/auth', () => ({
  logout: mockedLogout
}));

describe('BasicLayout', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockedLogout.mockResolvedValue(undefined);
    window.localStorage.clear();
  });

  it('renders the admin header, breadcrumb, user menu, and page tabs on light surfaces', async () => {
    const user = userEvent.setup();
    const { container } = render(<BasicLayout />);

    expect(container.querySelector('.admin-header.admin-light-surface')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '收缩侧边栏' })).toBeInTheDocument();
    expect(screen.getByText('首页')).toBeInTheDocument();
    expect(screen.getByText('巡检业务')).toBeInTheDocument();
    const pageTabs = container.querySelector('.admin-page-tabs.admin-light-surface');
    expect(pageTabs).toBeInTheDocument();
    expect(within(pageTabs as HTMLElement).getByText('检测任务')).toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: '用户菜单' }));
    expect(screen.getByText('个人中心')).toBeInTheDocument();
    expect(screen.getByText('退出登录')).toBeInTheDocument();

    await user.click(screen.getByText('退出登录'));

    await waitFor(() => {
      expect(mockedLogout).toHaveBeenCalledTimes(1);
      expect(mockedNavigate).toHaveBeenCalledWith('/user/login', { replace: true });
    });
  });
});
