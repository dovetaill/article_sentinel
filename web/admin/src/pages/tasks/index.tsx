import { ProTable } from '@ant-design/pro-components';
import { Button, Input, Select } from 'antd';
import { useMemo, useState } from 'react';

import { PageHeader } from '../../components/ui/page-header';
import { SectionCard } from '../../components/ui/section-card';
import { StatusBadge } from '../../components/ui/status-badge';
import { SummaryCard } from '../../components/ui/summary-card';
import { ToolbarStrip } from '../../components/ui/toolbar-strip';
import { listTasks, type TaskRecord } from '../../services/tasks';

type Filters = {
  task_no?: string;
  status?: string;
};

export default function TasksPage() {
  const [pageRows, setPageRows] = useState<TaskRecord[]>([]);
  const [draftFilters, setDraftFilters] = useState({ taskNo: '', status: undefined as string | undefined });
  const [submittedFilters, setSubmittedFilters] = useState<Filters>({});

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

  return (
    <>
      <PageHeader
        title="检测任务"
        description="统一发起巡检任务，查看执行状态、扫描规模与命中情况。"
        extra={(
          <Button key="new-task" type="primary" href="/tasks/new">
            新建任务
          </Button>
        )}
      />

      <div className="summary-card-grid">
        <SummaryCard label="本页任务数" value={summary.total} helper="当前分页已加载任务数量" />
        <SummaryCard label="执行中" value={summary.running} helper="仍在巡检中的任务" />
        <SummaryCard label="已完成" value={summary.success} helper="已完成的任务批次" />
        <SummaryCard label="命中总量" value={summary.hits} helper="当前分页累计命中次数" />
      </div>

      <SectionCard title="任务列表" description="按任务编号与执行状态浏览当前批次。">
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
            <span className="toolbar-strip__meta">更快筛出进行中的批次或定位单个任务编号。</span>
          </div>

          <div className="toolbar-strip__actions">
            <Button
              onClick={() => {
                setDraftFilters({ taskNo: '', status: undefined });
                setSubmittedFilters({});
              }}
            >
              重置
            </Button>
            <Button
              type="primary"
              onClick={() => {
                setSubmittedFilters({
                  task_no: draftFilters.taskNo || undefined,
                  status: draftFilters.status
                });
              }}
            >
              查询任务
            </Button>
          </div>
        </ToolbarStrip>

        <ProTable<TaskRecord>
          rowKey="id"
          cardBordered={false}
          headerTitle={false}
          size="small"
          search={false}
          params={submittedFilters}
          options={false}
          toolBarRender={false}
          request={async (params) => {
            const result = await listTasks({
              orgid: 100,
              page: params.current,
              pageSize: params.pageSize,
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
              render: (_, record) => [
                <Button key="detail" type="link" href={`/tasks/${record.id}`}>
                  查看详情
                </Button>
              ]
            }
          ]}
        />
      </SectionCard>
    </>
  );
}
