import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { HeaderBar } from './header-bar';

vi.mock('./org-switcher', () => ({
  OrgSwitcher: () => <button type="button">一县一端</button>
}));

vi.mock('./user-menu', () => ({
  UserMenu: () => <button type="button">当前用户</button>
}));

describe('HeaderBar', () => {
  it('shows only the page title so top-level pages do not repeat section copy', () => {
    render(
      <HeaderBar
        pageTitle="检测任务"
        sectionLabel="检测任务"
        sidebarCollapsed={false}
        onToggleSidebar={() => undefined}
      />,
    );

    expect(screen.getByRole('heading', { name: '检测任务' })).toBeInTheDocument();
    expect(screen.queryByText('检测任务', { selector: '.admin-header__eyebrow' })).not.toBeInTheDocument();
  });
});
