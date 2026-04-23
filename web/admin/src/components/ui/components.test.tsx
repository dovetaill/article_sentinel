import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { PageHeader } from './page-header';
import { SectionCard } from './section-card';
import { StatusBadge } from './status-badge';
import { SummaryCard } from './summary-card';

describe('shared admin presentation components', () => {
  it('renders compact page headers without the old eyebrow copy', () => {
    render(<PageHeader title="关键词规则" description="统一维护巡检词库与处置建议。" />);

    expect(screen.getByRole('heading', { name: '关键词规则' })).toBeInTheDocument();
    expect(screen.getByText('统一维护巡检词库与处置建议。')).toBeInTheDocument();
    expect(screen.queryByText('栏目说明')).not.toBeInTheDocument();
  });

  it('renders summary cards with helper text', () => {
    render(<SummaryCard label="任务总数" value="128" helper="含历史任务累计统计" />);

    expect(screen.getByText('任务总数')).toBeInTheDocument();
    expect(screen.getByText('128')).toBeInTheDocument();
    expect(screen.getByText('含历史任务累计统计')).toBeInTheDocument();
  });

  it('renders section cards with optional supporting description', () => {
    render(
      <SectionCard title="结果列表" description="按筛选条件查看当前命中记录。">
        <div>content</div>
      </SectionCard>,
    );

    expect(screen.getByRole('heading', { name: '结果列表' })).toBeInTheDocument();
    expect(screen.getByText('按筛选条件查看当前命中记录。')).toBeInTheDocument();
  });

  it('renders formal badge labels for workflow and risk states', () => {
    render(
      <>
        <StatusBadge kind="risk" value="high" />
        <StatusBadge kind="task" value="running" />
      </>,
    );

    expect(screen.getByText('高风险')).toBeInTheDocument();
    expect(screen.getByText('执行中')).toBeInTheDocument();
  });
});
