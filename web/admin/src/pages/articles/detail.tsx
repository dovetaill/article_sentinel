import { Button, Descriptions, Empty, List, Space, Spin, Tabs, Typography } from 'antd';
import { useEffect, useMemo, useState } from 'react';
import { useParams } from 'react-router-dom';

import { PageHeader } from '../../components/ui/page-header';
import { SectionCard } from '../../components/ui/section-card';
import { StatusBadge } from '../../components/ui/status-badge';
import { SummaryCard } from '../../components/ui/summary-card';
import { listArticleFieldChanges, listArticleOperationLogs, type FieldChangeLogRecord, type OperationLogRecord } from '../../services/logs';
import { getResultDetail, listResults, type ResultDetailRecord } from '../../services/results';

const { Paragraph, Text } = Typography;

export default function ArticleDetailPage() {
  const { articleId } = useParams();
  const [loading, setLoading] = useState(true);
  const [detail, setDetail] = useState<ResultDetailRecord | null>(null);
  const [operationLogs, setOperationLogs] = useState<OperationLogRecord[]>([]);
  const [fieldChanges, setFieldChanges] = useState<FieldChangeLogRecord[]>([]);

  useEffect(() => {
    if (!articleId) {
      setLoading(false);
      return;
    }

    setLoading(true);

    void listResults({ orgid: 100, article_id: Number(articleId), page: 1, pageSize: 20 })
      .then(async (result) => {
        const primary = result.items[0];
        if (!primary) {
          setDetail(null);
          setOperationLogs([]);
          setFieldChanges([]);
          return;
        }

        const [detailResult, logsResult, changesResult] = await Promise.all([
          getResultDetail(primary.id, 100),
          listArticleOperationLogs(primary.article_id, 100),
          listArticleFieldChanges(primary.article_id, 100)
        ]);

        setDetail(detailResult);
        setOperationLogs(logsResult.items);
        setFieldChanges(changesResult.items);
      })
      .finally(() => setLoading(false));
  }, [articleId]);

  const metrics = useMemo(() => {
    if (!detail) {
      return [];
    }

    return [
      { label: '文稿编号', value: `#${detail.article_id}`, helper: '当前稿件唯一编号' },
      { label: '任务编号', value: `#${detail.task_id}`, helper: '关联的最近巡检任务' },
      { label: '命中次数', value: detail.hit_count, helper: '累计命中记录次数' },
      { label: '处置状态', value: detail.disposition_status === 'pending' ? '待处置' : '已处置', helper: '当前处置进度' }
    ];
  }, [detail]);

  const tabItems = useMemo(() => [
    {
      key: 'hits',
      label: '命中记录',
      children: detail?.hits.length ? (
        <List
          dataSource={detail.hits}
          renderItem={(item) => (
            <List.Item>
              <Space direction="vertical" size={4} style={{ width: '100%' }}>
                <Space size={8} wrap>
                  <StatusBadge kind="risk" value={item.risk_level} />
                  <Text strong>{item.field_name}</Text>
                  <Text type="secondary">{item.keyword_text}</Text>
                </Space>
                <Paragraph>{item.snippet}</Paragraph>
              </Space>
            </List.Item>
          )}
        />
      ) : <Empty description="暂无命中记录。" />
    },
    {
      key: 'logs',
      label: '操作记录',
      children: operationLogs.length ? (
        <List
          dataSource={operationLogs}
          renderItem={(item) => (
            <List.Item>
              <Space direction="vertical" size={2} style={{ width: '100%' }}>
                <Text strong>{item.summary}</Text>
                <Text type="secondary">{item.operator_name || '未知'} · {item.created_at || '-'}</Text>
              </Space>
            </List.Item>
          )}
        />
      ) : <Empty description="暂无操作记录。" />
    },
    {
      key: 'changes',
      label: '字段变更',
      children: fieldChanges.length ? (
        <List
          dataSource={fieldChanges}
          renderItem={(item) => (
            <List.Item>
              <Space direction="vertical" size={2} style={{ width: '100%' }}>
                <Text strong>{item.field_name}</Text>
                <Text type="secondary">{item.before_value || '-'} → {item.after_value || '-'}</Text>
              </Space>
            </List.Item>
          )}
        />
      ) : <Empty description="暂无字段变更。" />
    }
  ], [detail, fieldChanges, operationLogs]);

  return (
    <>
      <PageHeader
        title="文稿详情"
        extra={(
          <Space wrap>
            <Button href="/articles">返回列表</Button>
            <Button type="primary" href={`/articles/${articleId ?? ''}/rectify`}>
              进入整改
            </Button>
          </Space>
        )}
      />

      {loading ? (
        <SectionCard title="文稿详情" description="正在加载当前稿件的巡检详情。">
          <div style={{ padding: '32px 0', textAlign: 'center' }}>
            <Spin />
          </div>
        </SectionCard>
      ) : null}

      {!loading && !detail ? (
        <SectionCard title="文稿详情" description="未返回可展示的巡检详情。">
          <Empty description="未查询到该稿件的巡检记录。" />
        </SectionCard>
      ) : null}

      {!loading && detail ? (
        <>
          <div className="summary-card-grid">
            {metrics.map((item) => (
              <SummaryCard key={item.label} label={item.label} value={item.value} helper={item.helper} />
            ))}
          </div>

          <div className="rectify-layout">
            <SectionCard title={detail.article_title}>
              <Space direction="vertical" size={16} style={{ width: '100%' }}>
                <Space wrap>
                  <StatusBadge kind="risk" value={detail.risk_level} />
                  <span className={`status-badge ${detail.disposition_status === 'pending' ? 'status-badge--warning' : 'status-badge--success'}`}>
                    {detail.disposition_status === 'pending' ? '待处置' : '已处置'}
                  </span>
                </Space>
                <Descriptions column={1} size="small">
                  <Descriptions.Item label="文稿编号">#{detail.article_id}</Descriptions.Item>
                  <Descriptions.Item label="任务编号">#{detail.task_id}</Descriptions.Item>
                  <Descriptions.Item label="正文快照">
                    {detail.article_body ? (
                      <div dangerouslySetInnerHTML={{ __html: detail.article_body }} />
                    ) : '暂无正文快照'}
                  </Descriptions.Item>
                </Descriptions>
              </Space>
            </SectionCard>

            <div className="rectify-layout__side">
              <SectionCard title="处置建议" description="结合最近巡检结果快速完成处置判断。">
                <Space direction="vertical" size={10} style={{ width: '100%' }}>
                  <Text>建议处置：{detail.suggest_action === 'offline' ? '下线处置' : detail.suggest_action}</Text>
                  <Text>最近处理人：{operationLogs[0]?.operator_name || detail.latest_operator_name || '-'}</Text>
                  <Text>最近处置时间：{operationLogs[0]?.created_at || detail.latest_action_at || '-'}</Text>
                </Space>
              </SectionCard>
            </div>
          </div>

          <SectionCard title="详情记录" description="查看命中记录、操作留痕和字段变更。">
            <Tabs defaultActiveKey="hits" items={tabItems} />
          </SectionCard>
        </>
      ) : null}
    </>
  );
}
