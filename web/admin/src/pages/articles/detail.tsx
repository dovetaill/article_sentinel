import { Button, Descriptions, Empty, List, Space, Spin, Tabs, Typography, message } from 'antd';
import { useEffect, useMemo, useState } from 'react';
import { useParams } from 'react-router-dom';

import { PageHeader } from '../../components/ui/page-header';
import { SectionCard } from '../../components/ui/section-card';
import { StatusBadge } from '../../components/ui/status-badge';
import { SummaryCard } from '../../components/ui/summary-card';
import { useOrgContext } from '../../context/org-context';
import {
  getArticleDetail,
  offlineArticle,
  republishArticle,
  type ArticleDetailRecord
} from '../../services/articles';
import {
  listArticleFieldChanges,
  listArticleOperationLogs,
  type FieldChangeLogRecord,
  type OperationLogRecord
} from '../../services/logs';
import { getResultDetail, type ResultDetailRecord } from '../../services/results';

const { Paragraph, Text } = Typography;

type LifecycleAction = 'offline' | 'republish' | null;

function renderArticleState(value?: number) {
  switch (value) {
    case 9:
      return '已发布';
    case 8:
      return '已下线';
    case 1:
      return '待审';
    default:
      return '-';
  }
}

function renderDispositionStatus(value?: string) {
  if (value === 'pending') {
    return '待处置';
  }

  if (value === 'processed') {
    return '已处置';
  }

  return '-';
}

function renderSuggestAction(value?: string) {
  if (value === 'offline') {
    return '下线处置';
  }

  if (value === 'process') {
    return '人工处理';
  }

  return value || '-';
}

export default function ArticleDetailPage() {
  const { articleId } = useParams();
  const [messageApi, contextHolder] = message.useMessage();
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState<LifecycleAction>(null);
  const [detail, setDetail] = useState<ArticleDetailRecord | null>(null);
  const [inspectDetail, setInspectDetail] = useState<ResultDetailRecord | null>(null);
  const [operationLogs, setOperationLogs] = useState<OperationLogRecord[]>([]);
  const [fieldChanges, setFieldChanges] = useState<FieldChangeLogRecord[]>([]);
  const [refreshSeed, setRefreshSeed] = useState(0);
  const { activeOrgId } = useOrgContext();

  const currentOrgId = activeOrgId ?? 29;
  const numericArticleId = Number(articleId || 0);

  useEffect(() => {
    if (!numericArticleId) {
      setLoading(false);
      return;
    }

    setLoading(true);

    void getArticleDetail(numericArticleId, currentOrgId)
      .then(async (articleDetail) => {
        const [resultDetail, logsResult, changesResult] = await Promise.all([
          articleDetail.latest_result_id
            ? getResultDetail(articleDetail.latest_result_id, currentOrgId)
            : Promise.resolve(null),
          listArticleOperationLogs(numericArticleId, currentOrgId),
          listArticleFieldChanges(numericArticleId, currentOrgId)
        ]);

        setDetail(articleDetail);
        setInspectDetail(resultDetail);
        setOperationLogs(logsResult.items);
        setFieldChanges(changesResult.items);
      })
      .catch(() => {
        setDetail(null);
        setInspectDetail(null);
        setOperationLogs([]);
        setFieldChanges([]);
      })
      .finally(() => setLoading(false));
  }, [currentOrgId, numericArticleId, refreshSeed]);

  const metrics = useMemo(() => {
    if (!detail) {
      return [];
    }

    return [
      { label: '文章编号', value: `#${detail.id}`, helper: '真实文章中心编号' },
      { label: '当前状态', value: renderArticleState(detail.state), helper: '当前稿件生命周期状态' },
      { label: '最新任务', value: detail.latest_task_id ? `#${detail.latest_task_id}` : '-', helper: '最近一次巡检任务' },
      {
        label: '处置状态',
        value: renderDispositionStatus(detail.latest_disposition_status),
        helper: '最近一次巡检结果的处置进度'
      }
    ];
  }, [detail]);

  async function handleLifecycle(kind: Exclude<LifecycleAction, null>) {
    if (!detail) {
      return;
    }

    setActionLoading(kind);

    try {
      if (kind === 'offline') {
        await offlineArticle(detail.id, {
          orgid: currentOrgId,
          task_id: detail.latest_task_id,
          result_id: detail.latest_result_id
        });
        messageApi.success('下线处置已提交');
      } else {
        await republishArticle(detail.id, {
          orgid: currentOrgId,
          task_id: detail.latest_task_id,
          result_id: detail.latest_result_id
        });
        messageApi.success('重新发布已提交');
      }

      setRefreshSeed((current) => current + 1);
    } finally {
      setActionLoading(null);
    }
  }

  const tabItems = useMemo(() => [
    {
      key: 'hits',
      label: '命中记录',
      children: inspectDetail?.hits.length ? (
        <List
          dataSource={inspectDetail.hits}
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
      ) : <Empty description="暂无最新命中记录。" />
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
  ], [fieldChanges, inspectDetail, operationLogs]);

  return (
    <>
      {contextHolder}
      <PageHeader
        title="文章详情"
        description="查看文章原文、最新巡检摘要和历史处置记录。"
        extra={(
          <Space wrap>
            <Button href="/articles">返回列表</Button>
            <Button type="primary" href={`/articles/${articleId ?? ''}/rectify`}>
              进入整改
            </Button>
            {detail?.state === 8 ? (
              <Button loading={actionLoading === 'republish'} onClick={() => void handleLifecycle('republish')}>
                重新发布
              </Button>
            ) : (
              <Button danger loading={actionLoading === 'offline'} onClick={() => void handleLifecycle('offline')}>
                下线处置
              </Button>
            )}
          </Space>
        )}
      />

      {loading ? (
        <SectionCard title="文章详情" description="正在加载当前文章中心数据。">
          <div style={{ padding: '32px 0', textAlign: 'center' }}>
            <Spin />
          </div>
        </SectionCard>
      ) : null}

      {!loading && !detail ? (
        <SectionCard title="文章详情" description="未返回可展示的文章数据。">
          <Empty description="未查询到该文章的中心数据。" />
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
            <SectionCard title={detail.title} description="查看文章基本字段、摘要和正文快照。">
              <Space direction="vertical" size={16} style={{ width: '100%' }}>
                <Space wrap>
                  <span className="status-badge status-badge--neutral">{renderArticleState(detail.state)}</span>
                  {detail.latest_risk_level ? <StatusBadge kind="risk" value={detail.latest_risk_level} /> : null}
                  {detail.latest_disposition_status ? (
                    <span className={`status-badge ${detail.latest_disposition_status === 'pending' ? 'status-badge--warning' : 'status-badge--success'}`}>
                      {renderDispositionStatus(detail.latest_disposition_status)}
                    </span>
                  ) : null}
                </Space>
                <Descriptions column={1} size="small">
                  <Descriptions.Item label="文章编号">#{detail.id}</Descriptions.Item>
                  <Descriptions.Item label="摘要">{detail.desc || '暂无摘要'}</Descriptions.Item>
                  <Descriptions.Item label="关键词">{detail.keyword || '-'}</Descriptions.Item>
                  <Descriptions.Item label="短标题">{detail.short_title || '-'}</Descriptions.Item>
                  <Descriptions.Item label="富标题">{detail.rich_title || '-'}</Descriptions.Item>
                  <Descriptions.Item label="正文快照">
                    {detail.body ? (
                      <div dangerouslySetInnerHTML={{ __html: detail.body }} />
                    ) : '暂无正文快照'}
                  </Descriptions.Item>
                </Descriptions>
              </Space>
            </SectionCard>

            <div className="rectify-layout__side">
              <SectionCard title="最近巡检摘要" description="结合最近一次巡检补充信息做出处置判断。">
                <Space direction="vertical" size={10} style={{ width: '100%' }}>
                  <Text>最新任务：{detail.latest_task_id ? `#${detail.latest_task_id}` : '-'}</Text>
                  <Text>最新风险：{detail.latest_risk_level || '-'}</Text>
                  <Text>建议动作：{renderSuggestAction(detail.latest_suggest_action)}</Text>
                  <Text>处置状态：{renderDispositionStatus(detail.latest_disposition_status)}</Text>
                  <Text>最近处理人：{detail.latest_operator_name || operationLogs[0]?.operator_name || '-'}</Text>
                  <Text>最近处理时间：{detail.latest_action_at || operationLogs[0]?.created_at || '-'}</Text>
                </Space>
              </SectionCard>
            </div>
          </div>

          <SectionCard title="详情记录" description="查看最近命中记录、操作留痕和字段变更。">
            <Tabs defaultActiveKey="hits" items={tabItems} />
          </SectionCard>
        </>
      ) : null}
    </>
  );
}
