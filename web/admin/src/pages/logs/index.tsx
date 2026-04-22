import { ProTable } from '@ant-design/pro-components';
import { Button, Input, Modal, Space } from 'antd';
import { useState } from 'react';

import { PageHeader } from '../../components/ui/page-header';
import { SectionCard } from '../../components/ui/section-card';
import { ToolbarStrip } from '../../components/ui/toolbar-strip';
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
      <PageHeader
        title="操作日志"
        description="查询任务执行、结果处置与请求快照，保留关键留痕信息，便于后续复盘与审计。"
      />

      <SectionCard title="日志检索">
        <ToolbarStrip>
          <Space wrap>
            <Input
              aria-label="文章编号"
              placeholder="文章编号"
              value={draftFilters.articleId}
              onChange={(event) => setDraftFilters((current) => ({ ...current, articleId: event.target.value }))}
            />
            <Input
              aria-label="任务编号"
              placeholder="任务编号"
              value={draftFilters.taskId}
              onChange={(event) => setDraftFilters((current) => ({ ...current, taskId: event.target.value }))}
            />
            <Input
              aria-label="操作人"
              placeholder="操作人"
              value={draftFilters.operator}
              onChange={(event) => setDraftFilters((current) => ({ ...current, operator: event.target.value }))}
            />
          </Space>
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
            查询日志
          </Button>
        </ToolbarStrip>

        <ProTable<OperationLogRecord>
          rowKey="id"
          cardBordered={false}
          search={false}
          params={submittedFilters}
          headerTitle={false}
          options={{ density: false, fullScreen: false }}
          toolBarRender={false}
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
            { title: '文章编号', dataIndex: 'article_id' },
            { title: '任务编号', dataIndex: 'task_id' },
            { title: '操作类型', dataIndex: 'operation_type' },
            {
              title: '状态流转',
              key: 'stateFlow',
              render: (_, record) => `${record.before_state || '-'} → ${record.after_state || '-'}`
            },
            { title: '摘要说明', dataIndex: 'summary' },
            { title: '操作人', dataIndex: 'operator_name' },
            { title: '操作时间', dataIndex: 'created_at' },
            {
              title: '操作',
              valueType: 'option',
              render: (_, record) => [
                <Button key="snapshot" type="link" onClick={() => setActiveSnapshot(record)}>
                  查看快照
                </Button>
              ]
            }
          ]}
        />
      </SectionCard>

      <Modal open={Boolean(activeSnapshot)} footer={null} title="请求快照" onCancel={() => setActiveSnapshot(null)}>
        <pre>{activeSnapshot?.request_snapshot || '暂无请求快照。'}</pre>
      </Modal>
    </>
  );
}
