import { ProTable } from '@ant-design/pro-components';
import { Button, Input, Popconfirm, Select, message } from 'antd';
import { useEffect, useMemo, useRef, useState } from 'react';
import { useSearchParams } from 'react-router-dom';

import { SectionCard } from '../../components/ui/section-card';
import { StatusBadge } from '../../components/ui/status-badge';
import { SummaryCard } from '../../components/ui/summary-card';
import { ToolbarStrip } from '../../components/ui/toolbar-strip';
import { useOrgContext } from '../../context/org-context';
import { deleteTask, listTasks, type TaskRecord } from '../../services/tasks';
import { useWorkbenchNavigation } from '../../workbench/navigation';

type Filters = {
  task_no?: string;
  status?: string;
};

type ActionRef = {
  reload?: () => void;
};

function normalizePage(value: string | null) {
  const page = Number(value || 0);

  return Number.isInteger(page) && page > 0 ? page : 1;
}

export default function TasksPage() {
  const actionRef = useRef<ActionRef>({});
  const [messageApi, contextHolder] = message.useMessage();
  const [pageRows, setPageRows] = useState<TaskRecord[]>([]);
  const [searchParams, setSearchParams] = useSearchParams();
  const [draftFilters, setDraftFilters] = useState(() => ({
    taskNo: searchParams.get('task_no') ?? '',
    status: searchParams.get('status') || undefined
  }));
  const { activeOrgId } = useOrgContext();
  const { buildHref, onLinkClick } = useWorkbenchNavigation();

  const currentOrgId = activeOrgId ?? 29;
  const submittedFilters: Filters = {
    task_no: searchParams.get('task_no')?.trim() || undefined,
    status: searchParams.get('status') || undefined
  };
  const currentPage = normalizePage(searchParams.get('page'));
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

  useEffect(() => {
    setDraftFilters({
      taskNo: searchParams.get('task_no') ?? '',
      status: searchParams.get('status') || undefined
    });
  }, [searchParams]);

  return (
    <>
      {contextHolder}
      <div className="summary-card-grid">
        <SummaryCard label="本页任务数" value={summary.total} />
        <SummaryCard label="执行中" value={summary.running} />
        <SummaryCard label="已完成" value={summary.success} />
        <SummaryCard label="命中总量" value={summary.hits} />
      </div>

      <SectionCard
        title="任务列表"
        extra={(
          <Button
            key="new-task"
            type="primary"
            href="/tasks/new"
            onClick={(event) => onLinkClick(event, '/tasks/new')}
          >
            新建任务
          </Button>
        )}
      >
        <ToolbarStrip>
          <div className="toolbar-strip__group">
            <div className="toolbar-strip__controls">
              <Input
                aria-label="任务编号"
                className="toolbar-strip__control"
                placeholder="任务编号 / inspect-20260420-01"
                value={draftFilters.taskNo}
                onChange={(event) => setDraftFilters((current) => ({ ...current, taskNo: event.target.value }))}
              />
              <Select
                allowClear
                aria-label="执行状态"
                className="toolbar-strip__control toolbar-strip__control--select"
                placeholder="执行状态"
                value={draftFilters.status}
                options={[
                  { label: '执行中', value: 'running' },
                  { label: '已完成', value: 'success' },
                  { label: '执行失败', value: 'failed' }
                ]}
                onChange={(value) => setDraftFilters((current) => ({ ...current, status: value }))}
              />
            </div>
          </div>

          <div className="toolbar-strip__actions">
            <Button
              onClick={() => {
                setSearchParams(new URLSearchParams());
              }}
            >
              重置
            </Button>
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
          </div>
        </ToolbarStrip>

        <ProTable<TaskRecord>
          rowKey="id"
          actionRef={actionRef as never}
          cardBordered={false}
          headerTitle={false}
          size="small"
          search={false}
          params={{ ...submittedFilters, orgid: currentOrgId, page: currentPage }}
          options={false}
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
            const result = await listTasks({
              orgid: currentOrgId,
              page: params.current ?? currentPage,
              pageSize: params.pageSize ?? 20,
              task_no: submittedFilters.task_no,
              status: submittedFilters.status
            });
            setPageRows(result.items);

            return {
              data: result.items,
              success: true,
              total: result.total
            };
          }}
          columns={[
            { title: '任务编号', dataIndex: 'task_no' },
            {
              title: '执行状态',
              dataIndex: 'status',
              render: (_, record) => <StatusBadge kind="task" value={record.status} />
            },
            {
              title: '扫描统计',
              key: 'stats',
              render: (_, record) => `已扫描 ${record.total_scanned ?? 0} 篇 / 命中 ${record.hit_count ?? 0} 次`
            },
            { title: '发起人', dataIndex: 'creator_name' },
            { title: '创建时间', dataIndex: 'created_at' },
            {
              title: '操作',
              valueType: 'option',
              render: (_, record) => {
                const resultsHref = buildHref(`/tasks/${record.id}/results`);
                const actions = [
                  <Button
                    key="results"
                    type="link"
                    href={resultsHref}
                    onClick={(event) => onLinkClick(event, resultsHref)}
                  >
                    运行结果
                  </Button>
                ];

                if (record.status === 'pending' || record.status === 'failed') {
                  actions.push(
                    <Popconfirm
                      key="delete"
                      title="删除任务"
                      description="仅待执行或执行失败的任务允许删除。"
                      okText="确认删除"
                      cancelText="取消"
                      onConfirm={async () => {
                        await deleteTask(record.id, currentOrgId);
                        messageApi.success('任务已删除');
                        actionRef.current.reload?.();
                      }}
                    >
                      <Button type="link" danger>
                        删除任务
                      </Button>
                    </Popconfirm>,
                  );
                } else {
                  actions.push(
                    <span key="protected" className="status-badge status-badge--neutral">
                      已执行不可删
                    </span>,
                  );
                }

                return actions;
              }
            }
          ]}
        />
      </SectionCard>
    </>
  );
}
