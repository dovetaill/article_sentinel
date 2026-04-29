import { Button, Empty, List, Space, Spin, Tabs, Typography } from 'antd';
import { useEffect, useMemo, useState } from 'react';
import { useLocation, useParams, useSearchParams } from 'react-router-dom';

import { PageHeader } from '../../components/ui/page-header';
import { SectionCard } from '../../components/ui/section-card';
import { StatusBadge } from '../../components/ui/status-badge';
import { SummaryCard } from '../../components/ui/summary-card';
import { useOrgContext } from '../../context/org-context';
import { getArticleDetail, type ArticleDetailRecord } from '../../services/articles';
import {
  listArticleFieldChanges,
  listArticleOperationLogs,
  type FieldChangeLogRecord,
  type OperationLogRecord
} from '../../services/logs';
import { getResultDetail, type ResultDetailRecord } from '../../services/results';
import { useWorkbenchNavigation } from '../../workbench/navigation';
import { readPageSession, writePageSession } from '../../workbench/page-session';
import { resolveWorkbenchRoute } from '../../workbench/registry';

const { Paragraph, Text } = Typography;

type SummaryMetric = {
  label: string;
  value: string | number;
  helper?: string;
};

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
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const [loading, setLoading] = useState(true);
  const [detail, setDetail] = useState<ArticleDetailRecord | null>(null);
  const [inspectDetail, setInspectDetail] = useState<ResultDetailRecord | null>(null);
  const [operationLogs, setOperationLogs] = useState<OperationLogRecord[]>([]);
  const [fieldChanges, setFieldChanges] = useState<FieldChangeLogRecord[]>([]);
  const { activeOrgId } = useOrgContext();
  const { buildHref, goBack, onLinkClick } = useWorkbenchNavigation();

  const currentOrgId = activeOrgId ?? 29;
  const tabKey = useMemo(
    () => resolveWorkbenchRoute(`${location.pathname}${location.search}`).key,
    [location.pathname, location.search],
  );
  const [activeTabKey, setActiveTabKey] = useState(() => (
    readPageSession<{ activeTab?: string }>(currentOrgId, tabKey)?.activeTab ?? 'hits'
  ));
  const numericArticleId = Number(articleId || 0);
  const returnTarget = searchParams.get('return_to') || '/articles';
  const rectifyHref = buildHref(`/articles/${articleId ?? ''}/rectify`, {
    returnTo: returnTarget,
    taskId: detail?.latest_task_id,
    resultId: detail?.latest_result_id
  });

  useEffect(() => {
    if (!numericArticleId) {
      setLoading(false);
      return;
    }

    setLoading(true);

    void getArticleDetail(numericArticleId)
      .then(async (articleDetail) => {
        const [resultDetail, logsResult, changesResult] = await Promise.all([
          articleDetail.latest_result_id
            ? getResultDetail(articleDetail.latest_result_id)
            : Promise.resolve(null),
          listArticleOperationLogs(numericArticleId),
          listArticleFieldChanges(numericArticleId)
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
  }, [currentOrgId, numericArticleId]);

  useEffect(() => {
    setActiveTabKey(readPageSession<{ activeTab?: string }>(currentOrgId, tabKey)?.activeTab ?? 'hits');
  }, [currentOrgId, tabKey]);

  useEffect(() => {
    writePageSession(currentOrgId, tabKey, { activeTab: activeTabKey });
  }, [activeTabKey, currentOrgId, tabKey]);

  const metrics = useMemo<SummaryMetric[]>(() => {
    if (!detail) {
      return [];
    }

    return [
      { label: '文章编号', value: `#${detail.id}` },
      { label: '当前状态', value: renderArticleState(detail.state) },
      { label: '最新任务', value: detail.latest_task_id ? `#${detail.latest_task_id}` : '-' },
      {
        label: '处置状态',
        value: renderDispositionStatus(detail.latest_disposition_status)
      }
    ];
  }, [detail]);

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
      <PageHeader
        title="文稿详情"
        extra={(
          <Space wrap>
            <Button href={returnTarget} onClick={(event) => {
              event.preventDefault();
              goBack({ returnTo: returnTarget, fallbackTo: '/articles' });
            }}
            >
              返回上一页
            </Button>
            <Button
              type="primary"
              href={rectifyHref}
              onClick={(event) => onLinkClick(event, rectifyHref)}
            >
              进入整改
            </Button>
          </Space>
        )}
      />

      {loading ? (
        <SectionCard title="文章详情">
          <div style={{ padding: '32px 0', textAlign: 'center' }}>
            <Spin />
          </div>
        </SectionCard>
      ) : null}

      {!loading && !detail ? (
        <SectionCard title="文章详情">
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
            <SectionCard title={detail.title}>
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
                <div className="article-detail__meta">
                  <div className="article-detail__meta-row">
                    <span className="article-detail__meta-label">文章编号</span>
                    <span className="article-detail__meta-value">#{detail.id}</span>
                  </div>
                  <div className="article-detail__meta-row">
                    <span className="article-detail__meta-label">摘要</span>
                    <span className="article-detail__meta-value">{detail.desc || '暂无摘要'}</span>
                  </div>
                  <div className="article-detail__meta-row">
                    <span className="article-detail__meta-label">关键词</span>
                    <span className="article-detail__meta-value">{detail.keyword || '-'}</span>
                  </div>
                  <div className="article-detail__meta-row">
                    <span className="article-detail__meta-label">短标题</span>
                    <span className="article-detail__meta-value">{detail.short_title || '-'}</span>
                  </div>
                </div>

                {detail.thumbnail ? (
                  <div className="article-detail__section">
                    <span className="article-detail__section-title">封面图</span>
                    <div className="article-detail__cover">
                      <img src={detail.thumbnail} alt="文稿封面" />
                    </div>
                  </div>
                ) : null}

                {detail.rich_title ? (
                  <div className="article-detail__section">
                    <span className="article-detail__section-title">富标题</span>
                    <div className="article-detail__html" dangerouslySetInnerHTML={{ __html: detail.rich_title }} />
                  </div>
                ) : null}

                <div className="article-detail__section">
                  <span className="article-detail__section-title">正文</span>
                  {detail.body ? (
                    <div className="article-detail__html" dangerouslySetInnerHTML={{ __html: detail.body }} />
                  ) : (
                    <span className="article-detail__meta-value">暂无正文快照</span>
                  )}
                </div>
              </Space>
            </SectionCard>

            <div className="rectify-layout__side">
              <SectionCard title="最近巡检摘要">
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

          <SectionCard title="详情记录">
            <Tabs activeKey={activeTabKey} onChange={setActiveTabKey} items={tabItems} />
          </SectionCard>
        </>
      ) : null}
    </>
  );
}
