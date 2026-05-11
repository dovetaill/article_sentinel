import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import BasicLayout from './BasicLayout';

const { mockedNavigate, mockedProLayout } = vi.hoisted(() => ({
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
        orgname: '示例机构'
      },
      currentOrgId: 29,
      currentOrgName: '示例机构'
    }
  }),
  useNavigate: () => mockedNavigate
}));

describe('BasicLayout', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    window.localStorage.clear();
  });

  it('keeps a dark sidebar while rendering a light workspace shell', () => {
    const { container } = render(<BasicLayout />);

    expect(screen.getByTestId('pro-layout')).toHaveAttribute('data-layout', 'side');
    expect(screen.getByTestId('pro-layout')).toHaveAttribute('data-nav-theme', 'light');
    expect(screen.getByTestId('pro-layout')).not.toHaveAttribute('data-nav-theme', 'realDark');

    expect(container.querySelector('.admin-workspace-shell')).toHaveClass('admin-workspace-shell--light');
    expect(container.querySelector('.admin-workspace-body')).toHaveClass('admin-workspace-body--light');
    expect(container.querySelector('.admin-page-tabs')).toHaveClass('admin-page-tabs--light');
    expect(screen.getByTestId('layout-outlet')).toBeInTheDocument();
  });
});
