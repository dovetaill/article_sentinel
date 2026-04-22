const riskLabelMap: Record<string, string> = {
  low: '低风险',
  medium: '中风险',
  high: '高风险'
};

const taskLabelMap: Record<string, string> = {
  pending: '待执行',
  running: '执行中',
  success: '已完成',
  failed: '执行失败',
  partial_success: '部分完成'
};

const badgeToneMap: Record<string, string> = {
  low: 'success',
  medium: 'warning',
  high: 'danger',
  pending: 'neutral',
  running: 'info',
  success: 'success',
  failed: 'danger',
  partial_success: 'warning'
};

export interface StatusBadgeProps {
  kind: 'risk' | 'task';
  value: string;
}

export function StatusBadge({ kind, value }: StatusBadgeProps) {
  const labelMap = kind === 'risk' ? riskLabelMap : taskLabelMap;
  const label = labelMap[value] ?? value;
  const tone = badgeToneMap[value] ?? 'neutral';

  return <span className={`status-badge status-badge--${tone}`}>{label}</span>;
}
