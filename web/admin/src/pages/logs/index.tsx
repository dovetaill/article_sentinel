import { ProTable } from '@ant-design/pro-components';
import { Button, Input, Modal } from 'antd';
import { useEffect, useState } from 'react';
import { useLocation, useSearchParams } from 'react-router-dom';

import { SectionCard } from '../../components/ui/section-card';
import { ToolbarStrip } from '../../components/ui/toolbar-strip';
import { formatInspectionSnapshot } from '../../lib/inspection-snapshot';
import { listOperationLogs, type OperationLogRecord } from '../../services/logs';
import { WorkbenchLink } from '../../workbench/link';
import { useWorkbenchNavigation } from '../../workbench/navigation';

type Filters = {
  article_id?: number;
  task_id?: number;
  operator_name?: string;
};

function normalizePage(value: string | null) {
  const page = Number(value || 0);

  return Number.isInteger(page) && page > 0 ? page : 1;
}

export default function LogsPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const [draftFilters, setDraftFilters] = useState(() => ({
    articleId: searchParams.get('article_id') ?? '',
    taskId: searchParams.get('task_id') ?? '',
    operator: searchParams.get('operator_name') ?? ''
  }));
  const [activeSnapshot, setActiveSnapshot] = useState<OperationLogRecord | null>(null);
  const location = useLocation();
  const { buildHref } = useWorkbenchNavigation();

  const currentHref = `${location.pathname}${location.search}`;
  const submittedFilters: Filters = {
    article_id: searchParams.get('article_id') ? Number(searchParams.get('article_id')) : undefined,
    task_id: searchParams.get('task_id') ? Number(searchParams.get('task_id')) : undefined,
    operator_name: searchParams.get('operator_name') || undefined
  };
  const currentPage = normalizePage(searchParams.get('page'));

  useEffect(() => {
    setDraftFilters({
      articleId: searchParams.get('article_id') ?? '',
      taskId: searchParams.get('task_id') ?? '',
      operator: searchParams.get('operator_name') ?? ''
    });
  }, [searchParams]);

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
                setSearchParams(new URLSearchParams());
              }}
            >
              重置
            </Button>
            <Button
              type="primary"
              onClick={() => {
                const nextSearchParams = new URLSearchParams();
                const nextOperator = draftFilters.operator.trim();

                if (draftFilters.articleId.trim()) {
                  nextSearchParams.set('article_id', draftFilters.articleId.trim());
                }

                if (draftFilters.taskId.trim()) {
                  nextSearchParams.set('task_id', draftFilters.taskId.trim());
                }

                if (nextOperator) {
                  nextSearchParams.set('operator_name', nextOperator);
                }

                setSearchParams(nextSearchParams);
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
          params={{ ...submittedFilters, page: currentPage }}
          headerTitle={false}
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
            const result = await listOperationLogs({
              page: params.current ?? currentPage,
              pageSize: params.pageSize ?? 20,
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
