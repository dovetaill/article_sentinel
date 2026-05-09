import { Tag } from 'antd';

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

const tagColorMap: Record<string, string> = {
  low: 'success',
  medium: 'warning',
  high: 'error',
  pending: 'default',
  running: 'processing',
  success: 'success',
  failed: 'error',
  partial_success: 'warning'
};

export interface StatusTagProps {
  kind: 'risk' | 'task';
  value: string;
}

export default function StatusTag({ kind, value }: StatusTagProps) {
  const labelMap = kind === 'risk' ? riskLabelMap : taskLabelMap;
  const label = labelMap[value] ?? value;
  const color = tagColorMap[value] ?? 'default';

  return (
    <Tag color={color} bordered={false}>
      {label}
    </Tag>
  );
}
