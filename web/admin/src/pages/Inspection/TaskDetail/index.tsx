import { PageContainer, ProDescriptions, ProTable, type ActionType, type ProColumns } from '@ant-design/pro-components';
import { Button, Card, Empty, List, Modal, Space, Spin, Statistic, Tabs, Tag, Typography, message } from 'antd';
import { useEffect, useMemo, useRef, useState } from 'react';
import { useLocation, useNavigate, useParams, useSearchParams } from 'react-router-dom';

import HitPreview from '@/components/HitPreview';
import SnapshotViewer from '@/components/SnapshotViewer';
import StatusTag from '@/components/StatusTag';
import { listOperationLogs, type OperationLogRecord } from '@/services/logs';
import { batchOfflineResults, listResults, type ResultRecord } from '@/services/results';
import { getTaskDetail, type TaskRecord } from '@/services/tasks';

const { Title, Paragraph, Text } = Typography;

type TaskTabKey = 'results' | 'rule-snapshot' | 'request-snapshot' | 'logs';

function normalizePage(value: string | null) {
  const page = Number(value || 0);
  return Number.isInteger(page) && page > 0 ? page : 1;
}

function normalizeTaskTab(value: string | null): TaskTabKey {
  if (value === 'rule-snapshot' || value === 'request-snapshot' || value === 'logs') {
    return value;
  }

  return 'results';
}

function buildTaskSearchParams(searchParams: URLSearchParams, tab: TaskTabKey, page?: number) {
  const nextSearchParams = new URLSearchParams();
  nextSearchParams.set('tab', tab);

  if (tab === 'results' && page && page > 1) {
    nextSearchParams.set('page', String(page));
  }

  return nextSearchParams;
}

function taskStatusLabel(status: string | undefined) {
  if (status === 'running') return '执行中';
  if (status === 'success') return '已完成';
  if (status === 'failed') return '执行失败';
  if (status === 'pending') return '待执行';
  return status || '-';
}

export default function TaskDetailPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const { taskId } = useParams();
  const [searchParams, setSearchParams] = useSearchParams();
  const numericTaskId = Number(taskId || 0) || 0;
  const activeTab = normalizeTaskTab(searchParams.get('tab'));
  const currentPage = normalizePage(searchParams.get('page'));
  const actionRef = useRef<ActionType>();
  const [messageApi, contextHolder] = message.useMessage();
  const [loading, setLoading] = useState(true);
  const [task, setTask] = useState<TaskRecord | null>(null);
  const [logs, setLogs] = useState<OperationLogRecord[]>([]);
  const [pageRows, setPageRows] = useState<ResultRecord[]>([]);
  const [selectedResultIds, setSelectedResultIds] = useState<number[]>([]);
  const [confirmIds, setConfirmIds] = useState<number[]>([]);

  useEffect(() => {
    if (!numericTaskId) {
      setLoading(false);
      setTask(null);
      setLogs([]);
      return;
    }

    let cancelled = false;
    setLoading(true);

    void Promise.all([
      getTaskDetail(numericTaskId),
      listOperationLogs({ task_id: numericTaskId, page: 1, pageSize: 20 })
    ])
      .then(([taskDetail, logList]) => {
        if (cancelled) {
          return;
        }

        setTask(taskDetail);
        setLogs(logList.items ?? []);
      })
      .catch(() => {
        if (cancelled) {
          return;
        }

        setTask(null);
        setLogs([]);
      })
      .finally(() => {
        if (!cancelled) {
          setLoading(false);
        }
      });

    return () => {
      cancelled = true;
    };
  }, [numericTaskId]);

  const currentResultHref = useMemo(() => {
    const nextSearchParams = buildTaskSearchParams(searchParams, 'results', currentPage);
    return `${location.pathname}?${nextSearchParams.toString()}`;
  }, [currentPage, location.pathname, searchParams]);

  const selectedCount = selectedResultIds.length;
  const confirmOpen = confirmIds.length > 0;
  const confirmTitle = confirmIds.length > 1 ? '批量下线处置' : '下线处置';

  const summary = useMemo(() => {
    if (!task) {
      return [];
    }

    return [
      { label: '任务编号', value: task.task_no },
      { label: '执行状态', value: taskStatusLabel(task.status) },
      { label: '已扫描数量', value: task.total_scanned ?? 0 },
      { label: '命中结果', value: `${task.hit_articles ?? 0} / ${task.hit_count ?? 0}` }
    ];
  }, [task]);

  const resultColumns: ProColumns<ResultRecord>[] = [
    {
      title: '文章标题',
      dataIndex: 'article_title',
      render: (_, record) => (
        <Button
          type="link"
          onClick={() => navigate(`/content/articles/${record.article_id}?return_to=${encodeURIComponent(currentResultHref)}`)}
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
          onClick={() => navigate(`/content/articles/${record.article_id}?return_to=${encodeURIComponent(currentResultHref)}`)}
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
              `/content/articles/${record.article_id}/rectify?return_to=${encodeURIComponent(currentResultHref)}&task_id=${record.task_id}&result_id=${record.id}`
            )
          }
        >
          进入整改
        </Button>
      ]
    }
  ];

  const tabItems = [
    {
      key: 'results',
      label: '命中结果',
      children: (
        <div className="admin-task-results">
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
              columns={resultColumns}
              search={false}
              options={false}
              cardBordered={false}
              headerTitle={false}
              toolBarRender={false}
              params={{
                taskId: numericTaskId,
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

                  setSearchParams(buildTaskSearchParams(searchParams, 'results', nextPage));
                }
              }}
              request={async (params) => {
                try {
                  const result = await listResults({
                    page: Number(params.current ?? params.page ?? currentPage) || currentPage,
                    pageSize: params.pageSize ?? 20,
                    task_id: numericTaskId
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
        </div>
      )
    },
    {
      key: 'rule-snapshot',
      label: '规则快照',
      children: <SnapshotViewer value={task?.rule_snapshot} emptyText="暂无规则快照。" />
    },
    {
      key: 'request-snapshot',
      label: '请求快照',
      children: <SnapshotViewer value={task?.request_snapshot} emptyText="暂无请求快照。" />
    },
    {
      key: 'logs',
      label: '关联日志',
      children: logs.length ? (
        <List
          className="detail-list"
          dataSource={logs}
          renderItem={(item) => (
            <List.Item>
              <Space direction="vertical" size={4} style={{ width: '100%' }}>
                <Text strong>{item.summary}</Text>
                <Text type="secondary">
                  {item.operator_name || '未知操作人'} · {item.created_at || '-'}
                </Text>
                <Text type="secondary">
                  {item.before_state || '-'} → {item.after_state || '-'}
                </Text>
              </Space>
            </List.Item>
          )}
        />
      ) : (
        <Empty description="当前任务暂无关联日志。" />
      )
    }
  ];

  return (
    <PageContainer title={false}>
      {contextHolder}
      <div className="admin-domain-page">
        <div className="admin-domain-page__head">
          <div>
            <Title level={3} className="admin-domain-page__title">
              任务详情
            </Title>
            <Paragraph className="admin-domain-page__desc">
              查看当前批次的规则快照、执行摘要、命中结果与关联日志。
            </Paragraph>
          </div>
          <Space wrap>
            <Button onClick={() => navigate('/inspection/tasks')}>返回任务列表</Button>
            <Button type="primary" onClick={() => setSearchParams(buildTaskSearchParams(searchParams, 'results', currentPage))}>
              查看任务结果
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

        {!loading && !task ? (
          <Card className="admin-filter-card admin-surface-panel" variant="borderless">
            <Empty description="未查询到该任务的详细记录。" />
          </Card>
        ) : null}

        {!loading && task ? (
          <>
            <Space size={16} wrap className="admin-summary-strip">
              {summary.map((item) => (
                <Card key={item.label} className="admin-summary-card admin-surface-panel" variant="borderless">
                  <Statistic title={item.label} value={item.value} />
                </Card>
              ))}
            </Space>

            <div className="admin-detail-layout">
              <Card className="admin-filter-card admin-detail-layout__main admin-surface-panel" variant="borderless">
                <Space direction="vertical" size={16} style={{ width: '100%' }}>
                  <Space wrap className="admin-status-strip">
                    <StatusTag kind="task" value={task.status} />
                    <Tag bordered={false}>{task.hit_articles ?? 0} 篇命中文稿</Tag>
                  </Space>
                  <ProDescriptions<TaskRecord>
                    dataSource={task}
                    column={1}
                    columns={[
                      { title: '任务编号', dataIndex: 'task_no' },
                      { title: '创建人', dataIndex: 'creator_name' },
                      { title: '创建时间', dataIndex: 'created_at' },
                      {
                        title: '执行摘要',
                        render: () => `已扫描 ${task.total_scanned ?? 0} 篇文章，累计命中 ${task.hit_count ?? 0} 次。`
                      }
                    ]}
                  />
                </Space>
              </Card>

              <Card className="admin-filter-card admin-detail-layout__side admin-surface-panel" variant="borderless">
                <Space direction="vertical" size={10} style={{ width: '100%' }}>
                  <Text>执行状态：{taskStatusLabel(task.status)}</Text>
                  <Text>命中文稿：{task.hit_articles ?? 0} 篇</Text>
                  <Text>命中次数：{task.hit_count ?? 0} 次</Text>
                  <Text>最近关联日志：{logs[0]?.summary || '暂无记录'}</Text>
                </Space>
              </Card>
            </div>

            <Card className="admin-filter-card admin-detail-tabs admin-surface-panel" variant="borderless">
              <Tabs
                activeKey={activeTab}
                onChange={(key) => setSearchParams(buildTaskSearchParams(searchParams, key as TaskTabKey, currentPage))}
                items={tabItems}
              />
            </Card>
          </>
        ) : null}
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
              task_id: numericTaskId,
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
