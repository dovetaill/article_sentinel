import { PageContainer } from '@ant-design/pro-components';
import { Button, Card, Empty, List, Space, Spin, Tabs, Tag, Typography } from 'antd';
import { useEffect, useMemo, useState } from 'react';
import { useNavigate, useParams, useSearchParams } from 'react-router-dom';

import StatusTag from '@/components/StatusTag';
import { getArticleDetail, type ArticleDetailRecord } from '@/services/articles';
import {
  listArticleFieldChanges,
  listArticleOperationLogs,
  type FieldChangeLogRecord,
  type OperationLogRecord
} from '@/services/logs';
import { getResultDetail, type ResultDetailRecord } from '@/services/results';

const { Title, Paragraph, Text } = Typography;

type ArticleTabKey = 'hits' | 'logs' | 'changes';

type SummaryMetric = {
  label: string;
  value: string | number;
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
  const navigate = useNavigate();
  const { articleId } = useParams();
  const [searchParams] = useSearchParams();
  const [loading, setLoading] = useState(true);
  const [detail, setDetail] = useState<ArticleDetailRecord | null>(null);
  const [inspectDetail, setInspectDetail] = useState<ResultDetailRecord | null>(null);
  const [operationLogs, setOperationLogs] = useState<OperationLogRecord[]>([]);
  const [fieldChanges, setFieldChanges] = useState<FieldChangeLogRecord[]>([]);
  const [activeTabKey, setActiveTabKey] = useState<ArticleTabKey>('hits');

  const numericArticleId = Number(articleId || 0) || 0;
  const returnTarget = searchParams.get('return_to') || '/content/articles';

  useEffect(() => {
    if (!numericArticleId) {
      setLoading(false);
      return;
    }

    let cancelled = false;
    setLoading(true);

    void getArticleDetail(numericArticleId)
      .then(async (articleDetail) => {
        const [resultDetail, logsResult, changesResult] = await Promise.allSettled([
          articleDetail.latest_result_id ? getResultDetail(articleDetail.latest_result_id) : Promise.resolve(null),
          listArticleOperationLogs(numericArticleId),
          listArticleFieldChanges(numericArticleId)
        ]);

        if (cancelled) {
          return;
        }

        setDetail(articleDetail);
        setInspectDetail(resultDetail.status === 'fulfilled' ? resultDetail.value : null);
        setOperationLogs(logsResult.status === 'fulfilled' ? logsResult.value.items ?? [] : []);
        setFieldChanges(changesResult.status === 'fulfilled' ? changesResult.value.items ?? [] : []);
      })
      .catch(() => {
        if (cancelled) {
          return;
        }

        setDetail(null);
        setInspectDetail(null);
        setOperationLogs([]);
        setFieldChanges([]);
      })
      .finally(() => {
        if (!cancelled) {
          setLoading(false);
        }
      });

    return () => {
      cancelled = true;
    };
  }, [numericArticleId]);

  const rectifyHref = `/content/articles/${articleId ?? ''}/rectify?return_to=${encodeURIComponent(returnTarget)}${detail?.latest_task_id ? `&task_id=${detail.latest_task_id}` : ''}${detail?.latest_result_id ? `&result_id=${detail.latest_result_id}` : ''}`;

  const metrics = useMemo<SummaryMetric[]>(() => {
    if (!detail) {
      return [];
    }

    return [
      { label: '文章编号', value: detail.id },
      { label: '当前状态', value: renderArticleState(detail.state) },
      { label: '最新任务', value: detail.latest_task_id ? `#${detail.latest_task_id}` : '-' },
      { label: '处置状态', value: renderDispositionStatus(detail.latest_disposition_status) }
    ];
  }, [detail]);

  const tabItems = useMemo(
    () => [
      {
        key: 'hits',
        label: '命中记录',
        children: inspectDetail?.hits?.length ? (
          <List
            dataSource={inspectDetail.hits}
            renderItem={(item) => (
              <List.Item>
                <Space direction="vertical" size={4} style={{ width: '100%' }}>
                  <Space size={8} wrap>
                    <StatusTag kind="risk" value={item.risk_level} />
                    <Text strong>{item.field_name}</Text>
                    <Text type="secondary">{item.keyword_text}</Text>
                  </Space>
                  <Paragraph>{item.snippet}</Paragraph>
                </Space>
              </List.Item>
            )}
          />
        ) : (
          <Empty description="暂无最新命中记录。" />
        )
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
                  <Text type="secondary">
                    {item.operator_name || '未知'} · {item.created_at || '-'}
                  </Text>
                </Space>
              </List.Item>
            )}
          />
        ) : (
          <Empty description="暂无操作记录。" />
        )
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
                  <Text type="secondary">
                    {item.before_value || '-'} → {item.after_value || '-'}
                  </Text>
                </Space>
              </List.Item>
            )}
          />
        ) : (
          <Empty description="暂无字段变更。" />
        )
      }
    ],
    [fieldChanges, inspectDetail, operationLogs]
  );

  return (
    <PageContainer title={false} pageHeaderRender={false}>
      <div className="admin-domain-page">
        <div className="admin-domain-page__head">
          <div>
            <Title level={3} className="admin-domain-page__title">
              文稿详情
            </Title>
            <Paragraph className="admin-domain-page__desc">
              查看真实文稿原文，并附带最近一次巡检留痕作为研判与整改参考。
            </Paragraph>
          </div>
          <Space wrap>
            <Button onClick={() => navigate(returnTarget)}>返回上一页</Button>
            <Button type="primary" onClick={() => navigate(rectifyHref)}>
              进入整改
            </Button>
          </Space>
        </div>

        {loading ? (
          <Card className="admin-filter-card admin-surface-panel" variant="borderless">
            <div style={{ padding: '32px 0', textAlign: 'center' }}>
              <Spin />
            </div>
          </Card>
        ) : null}

        {!loading && !detail ? (
          <Card className="admin-filter-card admin-surface-panel" variant="borderless">
            <Empty description="未查询到该文章的中心数据。" />
          </Card>
        ) : null}

        {!loading && detail ? (
          <>
            <Space size={16} wrap className="admin-summary-strip">
              {metrics.map((item) => (
                <Card key={item.label} className="admin-summary-card admin-surface-panel" variant="borderless">
                  <div className="admin-stat-card">
                    <div className="admin-stat-card__label">{item.label}</div>
                    <div className="admin-stat-card__value">{item.value}</div>
                  </div>
                </Card>
              ))}
            </Space>

            <div className="admin-detail-layout">
              <Card className="admin-filter-card admin-detail-layout__main admin-surface-panel" variant="borderless">
                <Space direction="vertical" size={16} style={{ width: '100%' }}>
                  <Space wrap className="admin-status-strip">
                    <Tag bordered={false}>{renderArticleState(detail.state)}</Tag>
                    {detail.latest_risk_level ? <StatusTag kind="risk" value={detail.latest_risk_level} /> : null}
                    {detail.latest_disposition_status ? (
                      <Tag bordered={false} color={detail.latest_disposition_status === 'pending' ? 'warning' : 'success'}>
                        {renderDispositionStatus(detail.latest_disposition_status)}
                      </Tag>
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
              </Card>

              <Card className="admin-filter-card admin-detail-layout__side admin-surface-panel" variant="borderless">
                <Space direction="vertical" size={10} style={{ width: '100%' }}>
                  <Text>最新任务：{detail.latest_task_id ? `#${detail.latest_task_id}` : '-'}</Text>
                  <Text>最新风险：{detail.latest_risk_level || '-'}</Text>
                  <Text>建议动作：{renderSuggestAction(detail.latest_suggest_action)}</Text>
                  <Text>处置状态：{renderDispositionStatus(detail.latest_disposition_status)}</Text>
                  <Text>最近处理人：{detail.latest_operator_name || operationLogs[0]?.operator_name || '-'}</Text>
                  <Text>最近处理时间：{detail.latest_action_at || operationLogs[0]?.created_at || '-'}</Text>
                </Space>
              </Card>
            </div>

            <Card className="admin-filter-card admin-detail-tabs admin-surface-panel" variant="borderless">
              <Tabs activeKey={activeTabKey} onChange={(key) => setActiveTabKey(key as ArticleTabKey)} items={tabItems} />
            </Card>
          </>
        ) : null}
      </div>
    </PageContainer>
  );
}
