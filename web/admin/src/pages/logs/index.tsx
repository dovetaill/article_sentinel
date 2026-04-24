import { ProTable } from '@ant-design/pro-components';
import { Button, Input, Modal } from 'antd';
import { useState } from 'react';

import { SectionCard } from '../../components/ui/section-card';
import { ToolbarStrip } from '../../components/ui/toolbar-strip';
import { useOrgContext } from '../../context/org-context';
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
  const { activeOrgId } = useOrgContext();

  const currentOrgId = activeOrgId ?? 29;

  return (
    <>
      <SectionCard>
        <ToolbarStrip>
          <div className="toolbar-strip__group">
            <div className="toolbar-strip__controls">
              <Input
                aria-label="文章编号"
                className="toolbar-strip__control"
                placeholder="文章编号"
                value={draftFilters.articleId}
                onChange={(event) => setDraftFilters((current) => ({ ...current, articleId: event.target.value }))}
              />
              <Input
                aria-label="任务编号"
                className="toolbar-strip__control"
                placeholder="任务编号"
                value={draftFilters.taskId}
                onChange={(event) => setDraftFilters((current) => ({ ...current, taskId: event.target.value }))}
              />
              <Input
                aria-label="操作人"
                className="toolbar-strip__control"
                placeholder="操作人"
                value={draftFilters.operator}
                onChange={(event) => setDraftFilters((current) => ({ ...current, operator: event.target.value }))}
              />
            </div>
            <span className="toolbar-strip__meta">快速串联任务、文章与操作人三个维度的记录。</span>
          </div>

          <div className="toolbar-strip__actions">
            <Button
              onClick={() => {
                setDraftFilters({ articleId: '', taskId: '', operator: '' });
                setSubmittedFilters({});
              }}
            >
              重置
            </Button>
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
          </div>
        </ToolbarStrip>

        <ProTable<OperationLogRecord>
          rowKey="id"
          cardBordered={false}
          size="small"
          search={false}
          params={{ ...submittedFilters, orgid: currentOrgId }}
          headerTitle={false}
          options={false}
          toolBarRender={false}
          request={async (params) => {
            const result = await listOperationLogs({
              orgid: currentOrgId,
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
            {
              title: '文章编号',
              dataIndex: 'article_id',
              render: (_, record) => record.article_id ? <a href={`/articles/${record.article_id}`}>#{record.article_id}</a> : '-'
            },
            {
              title: '任务编号',
              dataIndex: 'task_id',
              render: (_, record) => record.task_id ? <a href={`/tasks/${record.task_id}/results`}>#{record.task_id}</a> : '-'
            },
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
        <pre className="detail-code-block">{activeSnapshot?.request_snapshot || '暂无请求快照。'}</pre>
      </Modal>
    </>
  );
}
