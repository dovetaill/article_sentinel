import { ProTable } from '@ant-design/pro-components';
import { Button, Drawer, Space, Tag, Typography } from 'antd';
import { useState } from 'react';

import { getTaskDetail, listTasks, type TaskRecord } from '../../services/tasks';

const { Paragraph, Text } = Typography;

const statusColors: Record<string, string> = {
  pending: 'default',
  running: 'processing',
  success: 'green',
  failed: 'red',
  partial_success: 'gold'
};

export default function TasksPage() {
  const [detailOpen, setDetailOpen] = useState(false);
  const [detail, setDetail] = useState<TaskRecord | null>(null);

  return (
    <>
      <ProTable<TaskRecord>
        rowKey="id"
        cardBordered
        headerTitle="Inspection Tasks"
        search={false}
        toolBarRender={() => [
          <Button key="new-task" type="primary" href="/tasks/new">
            New Task
          </Button>
        ]}
        request={async (params) => {
          const result = await listTasks({
            orgid: 100,
            page: params.current,
            pageSize: params.pageSize
          });

          return {
            data: result.items,
            success: true,
            total: result.total
          };
        }}
        columns={[
          { title: 'Task No', dataIndex: 'task_no' },
          {
            title: 'Status',
            dataIndex: 'status',
            render: (_, record) => <Tag color={statusColors[record.status] ?? 'default'}>{record.status}</Tag>
          },
          {
            title: 'Stats',
            key: 'stats',
            render: (_, record) => `${record.total_scanned ?? 0} scanned / ${record.hit_count ?? 0} hits`
          },
          { title: 'Creator', dataIndex: 'creator_name' },
          { title: 'Created At', dataIndex: 'created_at' },
          {
            title: 'Action',
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
                Detail
              </Button>
            ]
          }
        ]}
      />

      <Drawer
        title="Task Detail"
        open={detailOpen}
        width={520}
        onClose={() => setDetailOpen(false)}
      >
        {detail ? (
          <Space direction="vertical" size={12} style={{ width: '100%' }}>
            <Text strong>{detail.task_no}</Text>
            <Tag color={statusColors[detail.status] ?? 'default'}>{detail.status}</Tag>
            <Paragraph>{detail.rule_snapshot || 'No snapshot available yet.'}</Paragraph>
            <Paragraph type="secondary">Creator: {detail.creator_name || 'unknown'}</Paragraph>
          </Space>
        ) : null}
      </Drawer>
    </>
  );
}
