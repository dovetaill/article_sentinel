import { PageContainer, ProTable, type ActionType, type ProColumns } from '@ant-design/pro-components';
import { Button, Card, Modal, Space, Statistic, Tag, Typography, message } from 'antd';
import { useMemo, useRef, useState } from 'react';
import { useLocation, useNavigate, useSearchParams } from 'react-router-dom';

import HitPreview from '@/components/HitPreview';
import StatusTag from '@/components/StatusTag';
import { batchOfflineResults, listResults, type ResultRecord } from '@/services/results';

const { Title, Paragraph, Text } = Typography;

function normalizePage(value: string | null) {
  const page = Number(value || 0);
  return Number.isInteger(page) && page > 0 ? page : 1;
}

function normalizeTaskId(value: string | null) {
  const taskId = Number(value || 0);
  return Number.isInteger(taskId) && taskId > 0 ? taskId : undefined;
}

function resolveSharedTaskId(resultIds: number[], rows: ResultRecord[]) {
  const selectedIds = new Set(resultIds);
  const taskIds = Array.from(
    new Set(
      rows
        .filter((row) => selectedIds.has(row.id))
        .map((row) => row.task_id)
        .filter((taskId) => Number.isFinite(taskId) && taskId > 0)
    )
  );

  return taskIds.length === 1 ? taskIds[0] : undefined;
}

export default function ResultListPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const actionRef = useRef<ActionType>();
  const [searchParams, setSearchParams] = useSearchParams();
  const [messageApi, contextHolder] = message.useMessage();
  const [pageRows, setPageRows] = useState<ResultRecord[]>([]);
  const [selectedResultIds, setSelectedResultIds] = useState<number[]>([]);
  const [confirmIds, setConfirmIds] = useState<number[]>([]);
  const currentHref = `${location.pathname}${location.search}`;
  const currentPage = normalizePage(searchParams.get('page'));
  const submittedTaskId = normalizeTaskId(searchParams.get('task_id'));

  const selectedCount = selectedResultIds.length;
  const confirmOpen = confirmIds.length > 0;
  const confirmTitle = useMemo(() => (confirmIds.length > 1 ? '批量下线处置' : '下线处置'), [confirmIds.length]);
  const summary = useMemo(() => {
    const highRisk = pageRows.filter((item) => item.risk_level === 'high').length;
    const pending = pageRows.filter((item) => item.disposition_status === 'pending').length;
    const hitCount = pageRows.reduce((total, item) => total + (item.hit_count ?? 0), 0);

    return {
      total: pageRows.length,
      highRisk,
      pending,
      hitCount
    };
  }, [pageRows]);

  const columns: ProColumns<ResultRecord>[] = [
    {
      title: '文章标题',
      dataIndex: 'article_title',
      render: (_, record) => (
        <Button
          type="link"
          onClick={() => navigate(`/content/articles/${record.article_id}?return_to=${encodeURIComponent(currentHref)}`)}
        >
          {record.article_title}
        </Button>
      )
    },
    {
      title: '文稿ID',
      dataIndex: 'article_id',
      width: 100,
      render: (_, record) => <Text>{record.article_id}</Text>
    },
    {
      title: '风险等级',
      dataIndex: 'risk_level',
      width: 120,
      render: (_, record) => <StatusTag kind="risk" value={record.risk_level} />
    },
    {
      title: '处置状态',
      dataIndex: 'disposition_status',
      width: 120,
      render: (_, record) => (
        <Tag
          className="admin-result-status-tag"
          bordered={false}
          color={record.disposition_status === 'pending' ? 'warning' : 'success'}
        >
          {record.disposition_status === 'pending' ? '待处置' : '已处置'}
        </Tag>
      )
    },
    {
      title: '命中次数',
      dataIndex: 'hit_count',
      width: 100
    },
    {
      title: '命中片段',
      dataIndex: 'snippet',
      render: (_, record) => (
        <HitPreview
          fieldName={record.preview_field_name}
          keywordText={record.preview_keyword_text ?? record.matched_keyword}
          matchedText={record.preview_matched_text ?? record.matched_keyword}
          snippet={record.preview_snippet ?? record.snippet}
          extraHitCount={record.extra_hit_count}
        />
      )
    },
    {
      title: '操作',
      valueType: 'option',
      width: 220,
      render: (_, record) => [
        <Button
          key="detail"
          type="link"
          onClick={() => navigate(`/content/articles/${record.article_id}?return_to=${encodeURIComponent(currentHref)}`)}
        >
          查看详情
        </Button>,
        <Button key="offline" type="link" danger onClick={() => setConfirmIds([record.id])}>
          下线处置
        </Button>,
        <Button
          key="rectify"
          type="link"
          onClick={() =>
            navigate(
              `/content/articles/${record.article_id}/rectify?return_to=${encodeURIComponent(currentHref)}&task_id=${record.task_id}&result_id=${record.id}`
            )
          }
        >
          进入整改
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
              风险结果
            </Title>
            <Paragraph className="admin-domain-page__desc">
              按当前筛选条件查看命中文稿，集中执行研判、下线与整改入口跳转。
            </Paragraph>
          </div>
        </div>

        <Space size={16} wrap className="admin-summary-strip">
          <Card className="admin-summary-card admin-surface-panel" variant="borderless">
            <Statistic title="本页结果数" value={summary.total} />
          </Card>
          <Card className="admin-summary-card admin-surface-panel" variant="borderless">
            <Statistic title="高风险" value={summary.highRisk} />
          </Card>
          <Card className="admin-summary-card admin-surface-panel" variant="borderless">
            <Statistic title="待处置" value={summary.pending} />
          </Card>
          <Card className="admin-summary-card admin-surface-panel" variant="borderless">
            <Statistic title="命中总量" value={summary.hitCount} />
          </Card>
        </Space>

        <Card className="admin-filter-card admin-surface-panel" variant="borderless">
          <div className="admin-filter-bar">
            <Text>已选 {selectedCount} 项</Text>
            <Space wrap>
              <Button onClick={() => setSelectedResultIds(pageRows.map((item) => item.id))}>本页全选</Button>
              <Button type="primary" danger disabled={selectedCount === 0} onClick={() => setConfirmIds(selectedResultIds)}>
                批量下线处置
              </Button>
            </Space>
          </div>

          <div className="admin-table-shell admin-surface-panel">
            <ProTable<ResultRecord>
              rowKey="id"
              actionRef={actionRef}
              columns={columns}
              search={false}
              options={false}
              cardBordered={false}
              headerTitle={false}
              toolBarRender={false}
              params={{
                taskId: submittedTaskId,
                page: currentPage
              }}
              rowSelection={{
                selectedRowKeys: selectedResultIds,
                onChange: (keys) => setSelectedResultIds(keys.map((item) => Number(item)))
              }}
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
                  const result = await listResults({
                    page: Number(params.current ?? params.page ?? currentPage) || currentPage,
                    pageSize: params.pageSize ?? 20,
                    task_id: submittedTaskId
                  });
                  setPageRows(result.items ?? []);
                  return {
                    data: result.items ?? [],
                    success: true,
                    total: result.total
                  };
                } catch (error) {
                  setPageRows([]);
                  messageApi.error(error instanceof Error ? error.message : '结果列表加载失败');
                  return {
                    data: [],
                    success: true,
                    total: 0
                  };
                }
              }}
            />
          </div>
        </Card>
      </div>

      <Modal
        open={confirmOpen}
        rootClassName="admin-light-modal admin-result-confirm-modal"
        title={confirmTitle}
        okText="确认处置"
        cancelText="取消"
        onCancel={() => setConfirmIds([])}
        onOk={async () => {
          try {
            await batchOfflineResults({
              task_id: resolveSharedTaskId(confirmIds, pageRows),
              result_ids: confirmIds,
              reason: 'manual batch offline'
            });
            messageApi.success('处置请求已提交');
            setSelectedResultIds([]);
            setConfirmIds([]);
            actionRef.current?.reload?.();
          } catch (error) {
            messageApi.error(error instanceof Error ? error.message : '批量处置提交失败');
          }
        }}
      >
        <p>确认对 {confirmIds.length} 篇文章执行下线处置？</p>
      </Modal>
    </PageContainer>
  );
}
