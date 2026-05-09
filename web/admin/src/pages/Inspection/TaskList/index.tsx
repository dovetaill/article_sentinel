import { PageContainer, ProTable, type ActionType, type ProColumns } from '@ant-design/pro-components';
import { Button, Card, Input, Popconfirm, Select, Space, Statistic, Tag, Typography, message } from 'antd';
import { useEffect, useMemo, useRef, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';

import StatusTag from '@/components/StatusTag';
import { deleteTask, listTasks, type TaskRecord } from '@/services/tasks';

const { Title, Paragraph } = Typography;

type TaskSearchValues = {
  taskNo: string;
  status?: string;
};

function normalizePage(value: string | null) {
  const page = Number(value || 0);
  return Number.isInteger(page) && page > 0 ? page : 1;
}

export default function TaskListPage() {
  const navigate = useNavigate();
  const actionRef = useRef<ActionType>();
  const [messageApi, contextHolder] = message.useMessage();
  const [pageRows, setPageRows] = useState<TaskRecord[]>([]);
  const [searchParams, setSearchParams] = useSearchParams();
  const [draftFilters, setDraftFilters] = useState<TaskSearchValues>(() => ({
    taskNo: searchParams.get('task_no') ?? '',
    status: searchParams.get('status') || undefined
  }));

  const currentPage = normalizePage(searchParams.get('page'));
  const submittedTaskNo = searchParams.get('task_no')?.trim() ?? '';
  const submittedStatus = searchParams.get('status') || undefined;

  useEffect(() => {
    setDraftFilters({
      taskNo: searchParams.get('task_no') ?? '',
      status: searchParams.get('status') || undefined
    });
  }, [searchParams]);

  const summary = useMemo(() => {
    const running = pageRows.filter((item) => item.status === 'running').length;
    const success = pageRows.filter((item) => item.status === 'success').length;
    const hitCount = pageRows.reduce((total, item) => total + (item.hit_count ?? 0), 0);

    return {
      total: pageRows.length,
      running,
      success,
      hits: hitCount
    };
  }, [pageRows]);

  const columns: ProColumns<TaskRecord>[] = [
    {
      title: '任务编号',
      dataIndex: 'task_no'
    },
    {
      title: '执行状态',
      dataIndex: 'status',
      width: 120,
      render: (_, record) => <StatusTag kind="task" value={record.status} />
    },
    {
      title: '扫描统计',
      key: 'stats',
      render: (_, record) => `已扫描 ${record.total_scanned ?? 0} 篇 / 命中 ${record.hit_count ?? 0} 次`
    },
    {
      title: '发起人',
      dataIndex: 'creator_name',
      width: 120
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      width: 180
    },
    {
      title: '操作',
      valueType: 'option',
      width: 220,
      render: (_, record) => {
        if (record.status === 'pending' || record.status === 'failed') {
          return [
            <Popconfirm
              key="delete"
              title="删除任务"
              description="仅待执行或执行失败的任务允许删除。"
              okText="确认删除"
              cancelText="取消"
              onConfirm={async () => {
                try {
                  await deleteTask(record.id);
                  messageApi.success('任务已删除');
                  actionRef.current?.reload?.();
                } catch (error) {
                  messageApi.error(error instanceof Error ? error.message : '任务删除失败');
                }
              }}
            >
              <Button type="link" danger>
                删除任务
              </Button>
            </Popconfirm>
          ];
        }

        return [
          <Tag key="protected" bordered={false}>
            已执行不可删
          </Tag>
        ];
      }
    }
  ];

  return (
    <PageContainer title={false}>
      {contextHolder}
      <div className="admin-domain-page">
        <div className="admin-domain-page__head">
          <div>
            <Title level={3} className="admin-domain-page__title">
              检测任务
            </Title>
            <Paragraph className="admin-domain-page__desc">
              统一发起巡检任务，查看执行状态、扫描规模与命中情况。
            </Paragraph>
          </div>
          <Button type="primary" onClick={() => navigate('/inspection/tasks/create')}>
            新建任务
          </Button>
        </div>

        <Space size={16} wrap>
          <Card variant="borderless">
            <Statistic title="本页任务数" value={summary.total} />
          </Card>
          <Card variant="borderless">
            <Statistic title="执行中" value={summary.running} />
          </Card>
          <Card variant="borderless">
            <Statistic title="已完成" value={summary.success} />
          </Card>
          <Card variant="borderless">
            <Statistic title="命中总量" value={summary.hits} />
          </Card>
        </Space>

        <Card className="admin-filter-card" variant="borderless">
          <div className="admin-filter-bar">
            <div className="admin-filter-bar__controls">
              <Input
                aria-label="任务编号"
                className="admin-filter-bar__control"
                placeholder="任务编号 / inspect-20260420-01"
                value={draftFilters.taskNo}
                onChange={(event) => {
                  setDraftFilters((current) => ({
                    ...current,
                    taskNo: event.target.value
                  }));
                }}
              />
              <Select
                allowClear
                aria-label="执行状态"
                className="admin-filter-bar__control admin-filter-bar__control--select"
                placeholder="执行状态"
                value={draftFilters.status}
                options={[
                  { label: '执行中', value: 'running' },
                  { label: '已完成', value: 'success' },
                  { label: '执行失败', value: 'failed' }
                ]}
                onChange={(value) => {
                  setDraftFilters((current) => ({
                    ...current,
                    status: value
                  }));
                }}
              />
            </div>
            <Space wrap>
              <Button onClick={() => setSearchParams(new URLSearchParams())}>重置</Button>
              <Button
                type="primary"
                onClick={() => {
                  const nextSearchParams = new URLSearchParams();
                  const nextTaskNo = draftFilters.taskNo.trim();

                  if (nextTaskNo) {
                    nextSearchParams.set('task_no', nextTaskNo);
                  }

                  if (draftFilters.status) {
                    nextSearchParams.set('status', draftFilters.status);
                  }

                  setSearchParams(nextSearchParams);
                }}
              >
                查询任务
              </Button>
            </Space>
          </div>

          <ProTable<TaskRecord>
            rowKey="id"
            actionRef={actionRef}
            columns={columns}
            search={false}
            options={false}
            cardBordered={false}
            headerTitle={false}
            toolBarRender={false}
            params={{
              task_no: submittedTaskNo,
              status: submittedStatus,
              page: currentPage
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
                const result = await listTasks({
                  page: Number(params.current ?? params.page ?? currentPage) || currentPage,
                  pageSize: params.pageSize ?? 20,
                  task_no: submittedTaskNo || undefined,
                  status: submittedStatus || undefined
                });
                setPageRows(result.items);

                return {
                  data: result.items,
                  success: true,
                  total: result.total
                };
              } catch (error) {
                setPageRows([]);
                messageApi.error(error instanceof Error ? error.message : '任务列表加载失败');
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
    </PageContainer>
  );
}
