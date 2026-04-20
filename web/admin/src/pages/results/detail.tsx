import { Drawer, Empty, List, Space, Spin, Tabs, Tag, Typography } from 'antd';
import { useEffect, useMemo, useState } from 'react';

import { getResultDetail, type ResultDetailRecord, type ResultHitRecord } from '../../services/results';

const { Paragraph, Text, Title } = Typography;

const riskColors: Record<string, string> = {
  low: 'green',
  medium: 'gold',
  high: 'red'
};

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
      label: 'Hit Details',
      children: detail?.hits.length ? (
        <List
          dataSource={detail.hits}
          renderItem={(item) => (
            <List.Item>
              <Space direction="vertical" size={4} style={{ width: '100%' }}>
                <Space>
                  <Tag color={riskColors[item.risk_level] ?? 'default'}>{item.risk_level}</Tag>
                  <Text strong>{item.field_name}</Text>
                  <Text type="secondary">{item.keyword_text}</Text>
                </Space>
                <Paragraph>{highlightSnippet(item)}</Paragraph>
              </Space>
            </List.Item>
          )}
        />
      ) : <Empty description="No hits" />
    },
    {
      key: 'body',
      label: 'Body Snippet',
      children: detail?.article_body ? <div dangerouslySetInnerHTML={{ __html: detail.article_body }} /> : <Empty description="No body snapshot" />
    },
    {
      key: 'operations',
      label: 'Operation History',
      children: detail?.operation_logs.length ? (
        <List
          dataSource={detail.operation_logs}
          renderItem={(item) => (
            <List.Item>
              <Space direction="vertical" size={2} style={{ width: '100%' }}>
                <Text strong>{item.summary}</Text>
                <Text type="secondary">{item.operator_name || 'unknown'} · {item.created_at || 'unknown time'}</Text>
              </Space>
            </List.Item>
          )}
        />
      ) : <Empty description="No operations" />
    },
    {
      key: 'changes',
      label: 'Field Changes',
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
      ) : <Empty description="No field changes" />
    }
  ], [detail]);

  return (
    <Drawer title="Result Detail" open={open} width={760} onClose={onClose}>
      {loading ? <Spin /> : null}
      {detail ? (
        <Space direction="vertical" size={16} style={{ width: '100%' }}>
          <div>
            <Tag color={riskColors[detail.risk_level] ?? 'default'}>{detail.risk_level}</Tag>
            <Title level={4}>{detail.article_title}</Title>
            <Paragraph>
              Article #{detail.article_id} · Task #{detail.task_id} · State {detail.article_state ?? '-'}
            </Paragraph>
          </div>
          <Tabs defaultActiveKey="hits" items={items} />
        </Space>
      ) : null}
    </Drawer>
  );
}
