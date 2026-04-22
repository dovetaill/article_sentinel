import { Drawer, Empty, List, Space, Spin, Tabs, Typography } from 'antd';
import { useEffect, useMemo, useState } from 'react';

import { StatusBadge } from '../../components/ui/status-badge';
import { getResultDetail, type ResultDetailRecord, type ResultHitRecord } from '../../services/results';

const { Paragraph, Text, Title } = Typography;

function highlightSnippet(hit: ResultHitRecord) {
  if (!hit.matched_text) {
    return hit.snippet;
  }

  const marker = hit.matched_text;
  const parts = hit.snippet.split(new RegExp(`(${escapePattern(marker)})`, 'gi'));
  return parts.map((part, index) => (
    part.toLowerCase() === marker.toLowerCase()
      ? <mark key={`${hit.id}-${index}`}>{part}</mark>
      : <span key={`${hit.id}-${index}`}>{part}</span>
  ));
}

function escapePattern(value: string) {
  return value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}

export interface ResultDetailDrawerProps {
  open: boolean;
  resultId?: number;
  orgid?: number;
  onClose: () => void;
}

export default function ResultDetailDrawer({ open, resultId, orgid = 100, onClose }: ResultDetailDrawerProps) {
  const [loading, setLoading] = useState(false);
  const [detail, setDetail] = useState<ResultDetailRecord | null>(null);

  useEffect(() => {
    if (!open || !resultId) {
      return;
    }

    setLoading(true);
    void getResultDetail(resultId, orgid)
      .then((data) => setDetail(data))
      .finally(() => setLoading(false));
  }, [open, orgid, resultId]);

  const items = useMemo(() => [
    {
      key: 'hits',
      label: '命中详情',
      children: detail?.hits.length ? (
        <List
          dataSource={detail.hits}
          renderItem={(item) => (
            <List.Item>
              <Space direction="vertical" size={4} style={{ width: '100%' }}>
                <Space>
                  <StatusBadge kind="risk" value={item.risk_level} />
                  <Text strong>{item.field_name}</Text>
                  <Text type="secondary">{item.keyword_text}</Text>
                </Space>
                <Paragraph>{highlightSnippet(item)}</Paragraph>
              </Space>
            </List.Item>
          )}
        />
      ) : <Empty description="暂无命中记录" />
    },
    {
      key: 'body',
      label: '正文快照',
      children: detail?.article_body ? <div dangerouslySetInnerHTML={{ __html: detail.article_body }} /> : <Empty description="暂无正文快照" />
    },
    {
      key: 'operations',
      label: '操作记录',
      children: detail?.operation_logs.length ? (
        <List
          dataSource={detail.operation_logs}
          renderItem={(item) => (
            <List.Item>
              <Space direction="vertical" size={2} style={{ width: '100%' }}>
                <Text strong>{item.summary}</Text>
                <Text type="secondary">{item.operator_name || '未知'} · {item.created_at || '未知时间'}</Text>
              </Space>
            </List.Item>
          )}
        />
      ) : <Empty description="暂无操作记录" />
    },
    {
      key: 'changes',
      label: '字段变更',
      children: detail?.field_changes.length ? (
        <List
          dataSource={detail.field_changes}
          renderItem={(item) => (
            <List.Item>
              <Space direction="vertical" size={2} style={{ width: '100%' }}>
                <Text strong>{item.field_name}</Text>
                <Text type="secondary">{item.old_value} → {item.new_value}</Text>
              </Space>
            </List.Item>
          )}
        />
      ) : <Empty description="暂无字段变更" />
    }
  ], [detail]);

  return (
    <Drawer title="结果详情" open={open} width={760} onClose={onClose}>
      {loading ? <Spin /> : null}
      {detail ? (
        <Space direction="vertical" size={16} style={{ width: '100%' }}>
          <div>
            <StatusBadge kind="risk" value={detail.risk_level} />
            <Title level={4}>{detail.article_title}</Title>
            <Paragraph>
              文章编号 #{detail.article_id} · 任务编号 #{detail.task_id} · 当前状态 {detail.article_state ?? '-'}
            </Paragraph>
          </div>
          <Tabs defaultActiveKey="hits" items={items} />
        </Space>
      ) : null}
    </Drawer>
  );
}
