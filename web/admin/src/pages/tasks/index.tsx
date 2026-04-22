import { ProTable } from '@ant-design/pro-components';
import { Button, Drawer, Space, Typography } from 'antd';
import { useMemo, useState } from 'react';

import { PageHeader } from '../../components/ui/page-header';
import { SectionCard } from '../../components/ui/section-card';
import { StatusBadge } from '../../components/ui/status-badge';
import { SummaryCard } from '../../components/ui/summary-card';
import { getTaskDetail, listTasks, type TaskRecord } from '../../services/tasks';

const { Paragraph, Text } = Typography;

export default function TasksPage() {
  const [detailOpen, setDetailOpen] = useState(false);
  const [detail, setDetail] = useState<TaskRecord | null>(null);
  const [pageRows, setPageRows] = useState<TaskRecord[]>([]);

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
        description="统一发起巡检任务，掌握执行状态、扫描规模与命中情况，便于值守人员连续跟进。"
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

      <SectionCard title="任务列表">
        <ProTable<TaskRecord>
          rowKey="id"
          cardBordered={false}
          headerTitle={false}
          search={false}
          options={{ density: false, fullScreen: false }}
          toolBarRender={false}
          request={async (params) => {
            const result = await listTasks({
              orgid: 100,
              page: params.current,
              pageSize: params.pageSize
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
                <Button
                  key="detail"
                  type="link"
                  onClick={async () => {
                    const task = await getTaskDetail(record.id, record.orgid);
                    setDetail(task);
                    setDetailOpen(true);
                  }}
                >
                  查看详情
                </Button>
              ]
            }
          ]}
        />
      </SectionCard>

      <Drawer
        title="任务详情"
        open={detailOpen}
        width={520}
        onClose={() => setDetailOpen(false)}
      >
        {detail ? (
          <Space direction="vertical" size={12} style={{ width: '100%' }}>
            <Text strong>{detail.task_no}</Text>
            <StatusBadge kind="task" value={detail.status} />
            <Paragraph>{detail.rule_snapshot || '当前暂无规则快照。'}</Paragraph>
            <Paragraph type="secondary">发起人：{detail.creator_name || '未知'}</Paragraph>
          </Space>
        ) : null}
      </Drawer>
    </>
  );
}
