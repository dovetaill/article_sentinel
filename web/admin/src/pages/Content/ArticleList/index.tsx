import { PageContainer, ProTable, type ActionType, type ProColumns } from '@ant-design/pro-components';
import { Button, Card, Input, Space, Statistic, Typography } from 'antd';
import { useEffect, useMemo, useRef, useState } from 'react';
import { useLocation, useNavigate, useSearchParams } from 'react-router-dom';

import StatusTag from '@/components/StatusTag';
import { listArticles, type ArticleListItem } from '@/services/articles';

const { Title, Paragraph, Text } = Typography;

type ArticleSearchValues = {
  title: string;
  articleId: string;
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

function normalizeArticleId(value: string) {
  const articleId = Number(value.trim());
  return Number.isInteger(articleId) && articleId > 0 ? articleId : undefined;
}

function normalizePage(value: string | null) {
  const page = Number(value || 0);
  return Number.isInteger(page) && page > 0 ? page : 1;
}

export default function ArticleListPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const actionRef = useRef<ActionType>();
  const [pageRows, setPageRows] = useState<ArticleListItem[]>([]);
  const [searchParams, setSearchParams] = useSearchParams();
  const [draftFilters, setDraftFilters] = useState<ArticleSearchValues>(() => ({
    title: searchParams.get('title') ?? '',
    articleId: searchParams.get('article_id') ?? ''
  }));

  const submittedTitle = searchParams.get('title')?.trim() ?? '';
  const submittedArticleIdText = searchParams.get('article_id') ?? '';
  const submittedArticleId = normalizeArticleId(submittedArticleIdText);
  const currentPage = normalizePage(searchParams.get('page'));
  const currentHref = `${location.pathname}${location.search}`;
  const requestParams = useMemo(
    () => ({
      title: submittedTitle,
      articleId: submittedArticleIdText,
      page: currentPage
    }),
    [currentPage, submittedArticleIdText, submittedTitle]
  );

  useEffect(() => {
    setDraftFilters({
      title: searchParams.get('title') ?? '',
      articleId: searchParams.get('article_id') ?? ''
    });
  }, [searchParams]);

  const summary = useMemo(() => {
    const published = pageRows.filter((item) => item.state === 9).length;
    const offline = pageRows.filter((item) => item.state === 8).length;
    const inspected = pageRows.filter((item) => item.latest_task_id).length;

    return {
      total: pageRows.length,
      published,
      offline,
      inspected
    };
  }, [pageRows]);

  const columns: ProColumns<ArticleListItem>[] = [
    {
      title: '文章标题',
      dataIndex: 'title',
      render: (_, record) => (
        <Button
          type="link"
          onClick={() => navigate(`/content/articles/${record.id}?return_to=${encodeURIComponent(currentHref)}`)}
        >
          {record.title}
        </Button>
      )
    },
    {
      title: '文章编号',
      dataIndex: 'id',
      width: 100,
      render: (_, record) => <Text>{record.id}</Text>
    },
    {
      title: '当前状态',
      dataIndex: 'state',
      width: 120,
      render: (_, record) => renderArticleState(record.state)
    },
    {
      title: '发布时间',
      dataIndex: 'publish_at_time',
      width: 180,
      render: (_, record) => record.publish_at_time || '-'
    },
    {
      title: '最近巡检',
      dataIndex: 'latest_risk_level',
      render: (_, record) =>
        record.latest_task_id ? (
          <Space size={8} wrap>
            <Text>任务 #{record.latest_task_id}</Text>
            {record.latest_risk_level ? <StatusTag kind="risk" value={record.latest_risk_level} /> : null}
          </Space>
        ) : (
          '-'
        )
    },
    {
      title: '操作',
      key: 'actions',
      width: 120,
      render: (_, record) => (
        <Button
          type="link"
          onClick={() => navigate(`/content/articles/${record.id}?return_to=${encodeURIComponent(currentHref)}`)}
        >
          查看详情
        </Button>
      )
    }
  ];

  return (
    <PageContainer title={false}>
      <div className="admin-domain-page">
        <div className="admin-domain-page__head">
          <div>
            <Title level={3} className="admin-domain-page__title">
              文稿中心
            </Title>
            <Paragraph className="admin-domain-page__desc">
              查看真实文稿元数据，并结合最近一次巡检结果快速进入明细与处置链路。
            </Paragraph>
          </div>
        </div>

        <Space size={16} wrap className="admin-summary-strip">
          <Card className="admin-summary-card admin-surface-panel" variant="borderless">
            <Statistic title="本页文稿数" value={summary.total} />
          </Card>
          <Card className="admin-summary-card admin-surface-panel" variant="borderless">
            <Statistic title="已发布" value={summary.published} />
          </Card>
          <Card className="admin-summary-card admin-surface-panel" variant="borderless">
            <Statistic title="已下线" value={summary.offline} />
          </Card>
          <Card className="admin-summary-card admin-surface-panel" variant="borderless">
            <Statistic title="含巡检记录" value={summary.inspected} />
          </Card>
        </Space>

        <Card className="admin-filter-card admin-surface-panel" variant="borderless">
          <div className="admin-filter-bar">
            <div className="admin-filter-bar__controls">
              <Input
                aria-label="标题模糊查询"
                className="admin-filter-bar__control"
                placeholder="按标题模糊查找"
                value={draftFilters.title}
                onChange={(event) => {
                  setDraftFilters((current) => ({ ...current, title: event.target.value }));
                }}
              />
              <Input
                aria-label="按文稿ID查询"
                className="admin-filter-bar__control"
                inputMode="numeric"
                placeholder="按文稿ID查询"
                value={draftFilters.articleId}
                onChange={(event) => {
                  setDraftFilters((current) => ({ ...current, articleId: event.target.value }));
                }}
              />
            </div>

            <Space wrap>
              <Button
                onClick={() => {
                  if (searchParams.toString().length === 0) {
                    actionRef.current?.reload?.();
                    return;
                  }

                  setSearchParams(new URLSearchParams());
                }}
              >
                重置
              </Button>
              <Button
                type="primary"
                onClick={() => {
                  const nextSearchParams = new URLSearchParams();
                  const nextTitle = draftFilters.title.trim();
                  const nextArticleId = normalizeArticleId(draftFilters.articleId);

                  if (nextTitle) {
                    nextSearchParams.set('title', nextTitle);
                  }

                  if (nextArticleId) {
                    nextSearchParams.set('article_id', String(nextArticleId));
                  }

                  if (nextSearchParams.toString() === searchParams.toString()) {
                    actionRef.current?.reload?.();
                    return;
                  }

                  setSearchParams(nextSearchParams);
                }}
              >
                查询文稿
              </Button>
            </Space>
          </div>

          <div className="admin-table-shell admin-surface-panel">
            <ProTable<ArticleListItem>
              rowKey="id"
              actionRef={actionRef}
              columns={columns}
              params={requestParams}
              search={false}
              options={false}
              cardBordered={false}
              headerTitle={false}
              toolBarRender={false}
              pagination={{
                current: currentPage,
                pageSize: 20,
                showSizeChanger: false,
                onChange: (nextPage) => {
                  if (nextPage === currentPage) {
                    return;
                  }

                  const nextSearchParams = new URLSearchParams(searchParams);
                  if (nextPage > 1) {
                    nextSearchParams.set('page', String(nextPage));
                  } else {
                    nextSearchParams.delete('page');
                  }

                  setSearchParams(nextSearchParams);
                }
              }}
              request={async (params) => {
                try {
                  const result = await listArticles({
                    page: Number(params.current ?? currentPage) || currentPage,
                    pageSize: params.pageSize ?? 20,
                    article_id: submittedArticleId,
                    title: submittedTitle || undefined
                  });
                  setPageRows(result.items ?? []);

                  return {
                    data: result.items ?? [],
                    success: true,
                    total: result.total
                  };
                } catch (error) {
                  setPageRows([]);
                  throw error;
                }
              }}
            />
          </div>
        </Card>
      </div>
    </PageContainer>
  );
}
