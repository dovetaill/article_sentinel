import { ProTable } from '@ant-design/pro-components';
import { Button, Input, Modal } from 'antd';
import { useState } from 'react';
import { useLocation } from 'react-router-dom';

import { SectionCard } from '../../components/ui/section-card';
import { ToolbarStrip } from '../../components/ui/toolbar-strip';
import { useOrgContext } from '../../context/org-context';
import { formatInspectionSnapshot } from '../../lib/inspection-snapshot';
import { listOperationLogs, type OperationLogRecord } from '../../services/logs';
import { WorkbenchLink } from '../../workbench/link';
import { useWorkbenchNavigation } from '../../workbench/navigation';

type Filters = {
  article_id?: number;
  task_id?: number;
  operator_name?: string;
};

export default function LogsPage() {
  const [draftFilters, setDraftFilters] = useState({ articleId: '', taskId: '', operator: '' });
  const [submittedFilters, setSubmittedFilters] = useState<Filters>({});
  const [activeSnapshot, setActiveSnapshot] = useState<OperationLogRecord | null>(null);
  const location = useLocation();
  const { activeOrgId } = useOrgContext();
  const { buildHref } = useWorkbenchNavigation();

  const currentOrgId = activeOrgId ?? 29;
  const currentHref = `${location.pathname}${location.search}`;

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
              render: (_, record) => record.article_id ? (
                <WorkbenchLink to={`/articles/${record.article_id}`} options={{ returnTo: currentHref }}>
                  #{record.article_id}
                </WorkbenchLink>
              ) : '-'
            },
            {
              title: '任务编号',
              dataIndex: 'task_id',
              render: (_, record) => record.task_id ? (
                <WorkbenchLink to={buildHref(`/tasks/${record.task_id}/results`)}>
                  #{record.task_id}
                </WorkbenchLink>
              ) : '-'
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
        <pre className="detail-code-block">
          {formatInspectionSnapshot(activeSnapshot?.request_snapshot, '暂无请求快照。')}
        </pre>
      </Modal>
    </>
  );
}
