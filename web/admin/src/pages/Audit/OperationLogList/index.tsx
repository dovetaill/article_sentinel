import { PageContainer, ProTable, type ActionType, type ProColumns } from '@ant-design/pro-components';
import { Button, Card, Modal, Space, Statistic, Typography, message } from 'antd';
import { useEffect, useMemo, useRef, useState } from 'react';
import { useLocation, useNavigate, useSearchParams } from 'react-router-dom';

import SnapshotViewer from '@/components/SnapshotViewer';
import { listOperationLogs, type OperationLogRecord } from '@/services/logs';

const { Title, Paragraph, Text } = Typography;

type LogFilterValues = {
  articleId: string;
  taskId: string;
  operatorName: string;
};

function normalizePage(value: string | null) {
  const page = Number(value || 0);
  return Number.isInteger(page) && page > 0 ? page : 1;
}

function normalizeNumber(value: string) {
  const result = Number(value.trim());
  return Number.isInteger(result) && result > 0 ? result : undefined;
}

export default function OperationLogListPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const actionRef = useRef<ActionType>();
  const [searchParams, setSearchParams] = useSearchParams();
  const [messageApi, contextHolder] = message.useMessage();
  const [pageRows, setPageRows] = useState<OperationLogRecord[]>([]);
  const [activeSnapshot, setActiveSnapshot] = useState<OperationLogRecord | null>(null);
  const [draftFilters, setDraftFilters] = useState<LogFilterValues>(() => ({
    articleId: searchParams.get('article_id') ?? '',
    taskId: searchParams.get('task_id') ?? '',
    operatorName: searchParams.get('operator_name') ?? ''
  }));

  const currentPage = normalizePage(searchParams.get('page'));
  const submittedArticleId = normalizeNumber(searchParams.get('article_id') ?? '');
  const submittedTaskId = normalizeNumber(searchParams.get('task_id') ?? '');
  const submittedOperatorName = searchParams.get('operator_name')?.trim() ?? '';
  const currentHref = `${location.pathname}${location.search}`;
  const requestParams = useMemo(
    () => ({
      articleId: searchParams.get('article_id') ?? '',
      taskId: searchParams.get('task_id') ?? '',
      operatorName: searchParams.get('operator_name') ?? '',
      page: currentPage
    }),
    [currentPage, searchParams]
  );

  useEffect(() => {
    setDraftFilters({
      articleId: searchParams.get('article_id') ?? '',
      taskId: searchParams.get('task_id') ?? '',
      operatorName: searchParams.get('operator_name') ?? ''
    });
  }, [searchParams]);

  const summary = useMemo(() => {
    const articleCount = new Set(pageRows.map((item) => item.article_id).filter(Boolean)).size;
    const taskCount = new Set(pageRows.map((item) => item.task_id).filter(Boolean)).size;
    const snapshotCount = pageRows.filter((item) => item.request_snapshot).length;

    return {
      total: pageRows.length,
      articleCount,
      taskCount,
      snapshotCount
    };
  }, [pageRows]);

  const columns: ProColumns<OperationLogRecord>[] = [
    {
      title: '文章编号',
      dataIndex: 'article_id',
      width: 120,
      render: (_, record) =>
        record.article_id ? (
          <Button
            type="link"
            onClick={() => navigate(`/content/articles/${record.article_id}?return_to=${encodeURIComponent(currentHref)}`)}
          >
            #{record.article_id}
          </Button>
        ) : (
          '-'
        )
    },
    {
      title: '任务编号',
      dataIndex: 'task_id',
      width: 120,
      render: (_, record) =>
        record.task_id ? (
          <Button type="link" onClick={() => navigate(`/inspection/results?task_id=${record.task_id}`)}>
            #{record.task_id}
          </Button>
        ) : (
          '-'
        )
    },
    {
      title: '操作类型',
      dataIndex: 'operation_type',
      width: 120
    },
    {
      title: '状态流转',
      key: 'stateFlow',
      width: 160,
      render: (_, record) => `${record.before_state || '-'} → ${record.after_state || '-'}`
    },
    {
      title: '摘要说明',
      dataIndex: 'summary'
    },
    {
      title: '操作人',
      dataIndex: 'operator_name',
      width: 140,
      render: (_, record) => record.operator_name || '-'
    },
    {
      title: '操作时间',
      dataIndex: 'created_at',
      width: 180,
      render: (_, record) => record.created_at || '-'
    },
    {
      title: '操作',
      valueType: 'option',
      width: 120,
      render: (_, record) => [
        <Button key="snapshot" type="link" onClick={() => setActiveSnapshot(record)}>
          查看快照
        </Button>
      ]
    }
  ];

  return (
    <PageContainer title={false}>
      {contextHolder}
      <div className="admin-domain-page">
        <div className="admin-domain-page__head">
          <div>
            <Title level={3} className="admin-domain-page__title">
              操作日志
            </Title>
            <Paragraph className="admin-domain-page__desc">
              按文章、任务与操作人回看审计留痕，并在弹窗中查看请求快照。
            </Paragraph>
          </div>
        </div>

        <Space size={16} wrap>
          <Card variant="borderless">
            <Statistic title="本页日志数" value={summary.total} />
          </Card>
          <Card variant="borderless">
            <Statistic title="关联文章数" value={summary.articleCount} />
          </Card>
          <Card variant="borderless">
            <Statistic title="关联任务数" value={summary.taskCount} />
          </Card>
          <Card variant="borderless">
            <Statistic title="含快照记录" value={summary.snapshotCount} />
          </Card>
        </Space>

        <Card className="admin-filter-card" variant="borderless">
          <div className="admin-filter-bar">
            <div className="admin-filter-bar__controls">
              <input
                aria-label="文章编号"
                className="ant-input ant-input-outlined admin-filter-bar__control"
                placeholder="文章编号"
                value={draftFilters.articleId}
                onChange={(event) => {
                  setDraftFilters((current) => ({ ...current, articleId: event.target.value }));
                }}
              />
              <input
                aria-label="任务编号"
                className="ant-input ant-input-outlined admin-filter-bar__control"
                placeholder="任务编号"
                value={draftFilters.taskId}
                onChange={(event) => {
                  setDraftFilters((current) => ({ ...current, taskId: event.target.value }));
                }}
              />
              <input
                aria-label="操作人"
                className="ant-input ant-input-outlined admin-filter-bar__control"
                placeholder="操作人"
                value={draftFilters.operatorName}
                onChange={(event) => {
                  setDraftFilters((current) => ({ ...current, operatorName: event.target.value }));
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
                  const nextArticleId = normalizeNumber(draftFilters.articleId);
                  const nextTaskId = normalizeNumber(draftFilters.taskId);
                  const nextOperatorName = draftFilters.operatorName.trim();

                  if (nextArticleId) {
                    nextSearchParams.set('article_id', String(nextArticleId));
                  }

                  if (nextTaskId) {
                    nextSearchParams.set('task_id', String(nextTaskId));
                  }

                  if (nextOperatorName) {
                    nextSearchParams.set('operator_name', nextOperatorName);
                  }

                  if (nextSearchParams.toString() === searchParams.toString()) {
                    actionRef.current?.reload?.();
                    return;
                  }

                  setSearchParams(nextSearchParams);
                }}
              >
                查询日志
              </Button>
            </Space>
          </div>

          <ProTable<OperationLogRecord>
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
                const result = await listOperationLogs({
                  page: Number(params.current ?? currentPage) || currentPage,
                  pageSize: params.pageSize ?? 20,
                  article_id: submittedArticleId,
                  task_id: submittedTaskId,
                  operator_name: submittedOperatorName || undefined
                });
                setPageRows(result.items ?? []);

                return {
                  data: result.items ?? [],
                  success: true,
                  total: result.total
                };
              } catch (error) {
                setPageRows([]);
                messageApi.error(error instanceof Error ? error.message : '操作日志加载失败');
                return {
                  data: [],
                  success: true,
                  total: 0
                };
              }
            }}
          />
        </Card>
      </div>

      <Modal
        open={Boolean(activeSnapshot)}
        footer={null}
        title="请求快照"
        onCancel={() => setActiveSnapshot(null)}
      >
        <SnapshotViewer value={activeSnapshot?.request_snapshot} emptyText="暂无请求快照。" />
      </Modal>
    </PageContainer>
  );
}
