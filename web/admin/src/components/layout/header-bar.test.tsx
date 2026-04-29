import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { OrgProvider } from '../../context/org-context';
import { SessionProvider } from '../../context/session-context';
import { HeaderBar } from './header-bar';

const { mockedGetSession, mockedLogout } = vi.hoisted(() => ({
  mockedGetSession: vi.fn(),
  mockedLogout: vi.fn(),
}));

vi.mock('../../services/auth', () => ({
  getSession: mockedGetSession,
  logout: mockedLogout,
}));

function renderHeader() {
  return render(
    <SessionProvider>
      <OrgProvider>
        <HeaderBar
          pageTitle="检测任务"
          sectionLabel="检测任务"
          sidebarCollapsed={false}
          onToggleSidebar={() => undefined}
        />
      </OrgProvider>
    </SessionProvider>,
  );
}

describe('HeaderBar', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockedGetSession.mockResolvedValue({
      id: 90525,
      orgid: 29,
      orgname: '一县一端',
      platform: 'chuangqi',
      priv: 'super',
      roleid: '1',
      nickname: '用户A',
      avatar: 'https://example.com/a.png',
    });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('shows only the page title so top-level pages do not repeat section copy', async () => {
    renderHeader();

    expect(await screen.findByRole('heading', { name: '检测任务' })).toBeInTheDocument();
    expect(screen.queryByText('检测任务', { selector: '.admin-header__eyebrow' })).not.toBeInTheDocument();
  });

  it('shows the current org as read-only and displays the session nickname', async () => {
    const user = userEvent.setup();

    renderHeader();

    const orgButton = await screen.findByRole('button', { name: /一县一端/i });
    expect(await screen.findByText('用户A')).toBeInTheDocument();

    await user.click(orgButton);

    expect(screen.queryByRole('menu')).not.toBeInTheDocument();
  });
});
