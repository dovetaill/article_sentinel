import { ProTable } from '@ant-design/pro-components';
import { Button, Input, Modal, Space } from 'antd';
import { useState } from 'react';

import { listOperationLogs, type OperationLogRecord } from '../../services/logs';

type Filters = {
  article_id?: number;
  task_id?: number;
  operator_name?: string;
};

export default function LogsPage() {
  const [draftFilters, setDraftFilters] = useState({ articleId: '', taskId: '', operator: '' });
  const [submittedFilters, setSubmittedFilters] = useState<Filters>({});
  const [activeSnapshot, setActiveSnapshot] = useState<OperationLogRecord | null>(null);

  return (
    <>
      <Space style={{ marginBottom: 16 }} wrap>
        <Input
          aria-label="Article ID"
          placeholder="Article ID"
          value={draftFilters.articleId}
          onChange={(event) => setDraftFilters((current) => ({ ...current, articleId: event.target.value }))}
        />
        <Input
          aria-label="Task ID"
          placeholder="Task ID"
          value={draftFilters.taskId}
          onChange={(event) => setDraftFilters((current) => ({ ...current, taskId: event.target.value }))}
        />
        <Input
          aria-label="Operator"
          placeholder="Operator"
          value={draftFilters.operator}
          onChange={(event) => setDraftFilters((current) => ({ ...current, operator: event.target.value }))}
        />
        <Button
          type="primary"
          onClick={() => {
            setSubmittedFilters({
              article_id: draftFilters.articleId ? Number(draftFilters.articleId) : undefined,
              task_id: draftFilters.taskId ? Number(draftFilters.taskId) : undefined,
              operator_name: draftFilters.operator || undefined
            });
          }}
        >
          Search Logs
        </Button>
      </Space>

      <ProTable<OperationLogRecord>
        rowKey="id"
        cardBordered
        search={false}
        params={submittedFilters}
        headerTitle="Operation Logs"
        request={async (params) => {
          const result = await listOperationLogs({
            orgid: 100,
            page: params.current,
            pageSize: params.pageSize,
            article_id: submittedFilters.article_id,
            task_id: submittedFilters.task_id,
            operator_name: submittedFilters.operator_name
          });
          return {
            data: result.items,
            success: true,
            total: result.total
          };
        }}
        columns={[
          { title: 'Article ID', dataIndex: 'article_id' },
          { title: 'Task ID', dataIndex: 'task_id' },
          { title: 'Type', dataIndex: 'operation_type' },
          {
            title: 'State Flow',
            key: 'stateFlow',
            render: (_, record) => `${record.before_state || '-'} → ${record.after_state || '-'}`
          },
          { title: 'Summary', dataIndex: 'summary' },
          { title: 'Operator', dataIndex: 'operator_name' },
          { title: 'Created At', dataIndex: 'created_at' },
          {
            title: 'Action',
            valueType: 'option',
            render: (_, record) => [
              <Button key="snapshot" type="link" onClick={() => setActiveSnapshot(record)}>
                View Snapshot
              </Button>
            ]
          }
        ]}
      />

      <Modal open={Boolean(activeSnapshot)} footer={null} title="Request Snapshot" onCancel={() => setActiveSnapshot(null)}>
        <pre>{activeSnapshot?.request_snapshot || 'No snapshot available.'}</pre>
      </Modal>
    </>
  );
}
