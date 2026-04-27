import { ProTable } from '@ant-design/pro-components';
import { Button, Empty, Space, Spin, Typography, message } from 'antd';
import { useEffect, useMemo, useState } from 'react';
import { useLocation, useParams } from 'react-router-dom';

import { HitPreview } from '../../components/ui/hit-preview';
import { PageHeader } from '../../components/ui/page-header';
import { SectionCard } from '../../components/ui/section-card';
import { StatusBadge } from '../../components/ui/status-badge';
import { SummaryCard } from '../../components/ui/summary-card';
import { useOrgContext } from '../../context/org-context';
import { listOperationLogs, type OperationLogRecord } from '../../services/logs';
import {
  batchIgnoreResults,
  batchOfflineResults,
  batchProcessResults,
  listResults,
  type ResultRecord
} from '../../services/results';
import { getTaskDetail, type TaskRecord } from '../../services/tasks';

const { Text } = Typography;

function taskStatusLabel(status: string | undefined) {
  if (status === 'running') return '执行中';
  if (status === 'success') return '已完成';
  if (status === 'failed') return '执行失败';
  return status || '-';
}

export default function TaskResultsPage() {
  const { taskId } = useParams();
  const location = useLocation();
  const [messageApi, contextHolder] = message.useMessage();
  const [loading, setLoading] = useState(true);
  const [task, setTask] = useState<TaskRecord | null>(null);
  const [results, setResults] = useState<ResultRecord[]>([]);
  const [logs, setLogs] = useState<OperationLogRecord[]>([]);
  const [selectedResultIds, setSelectedResultIds] = useState<number[]>([]);
  const [refreshSeed, setRefreshSeed] = useState(0);
  const { activeOrgId } = useOrgContext();

  const currentOrgId = activeOrgId ?? 29;
  const numericTaskId = Number(taskId || 0);
  const returnTo = `${location.pathname}${location.search}`;

  useEffect(() => {
    if (!numericTaskId) {
      setLoading(false);
      return;
    }

    setLoading(true);

    void Promise.all([
      getTaskDetail(numericTaskId, currentOrgId),
      listResults({ orgid: currentOrgId, task_id: numericTaskId, page: 1, pageSize: 50 }),
      listOperationLogs({ orgid: currentOrgId, task_id: numericTaskId, page: 1, pageSize: 20 })
    ])
      .then(([taskDetail, resultList, logList]) => {
        setTask(taskDetail);
        setResults(resultList.items);
        setLogs(logList.items);
      })
      .catch(() => {
        setTask(null);
        setResults([]);
        setLogs([]);
      })
      .finally(() => setLoading(false));
  }, [currentOrgId, numericTaskId, refreshSeed]);

  const summary = useMemo(() => {
    const highRisk = results.filter((item) => item.risk_level === 'high').length;
    const pending = results.filter((item) => item.disposition_status === 'pending').length;

    return {
      total: results.length,
      highRisk,
      pending,
      logs: logs.length
    };
  }, [logs.length, results]);

  async function runBatchAction(kind: 'offline' | 'ignore' | 'process', ids: number[]) {
    if (!ids.length) {
      return;
    }

    const input = {
      orgid: currentOrgId,
      result_ids: ids,
      reason: `task-${kind}-action`
    };

    if (kind === 'offline') {
      await batchOfflineResults(input);
      messageApi.success('下线处置已提交');
    } else if (kind === 'ignore') {
      await batchIgnoreResults(input);
      messageApi.success('忽略处置已提交');
    } else {
      await batchProcessResults(input);
      messageApi.success('已标记为人工处理');
    }

    setSelectedResultIds([]);
    setRefreshSeed((current) => current + 1);
  }

  return (
    <>
      {contextHolder}
      <PageHeader
        title="任务结果"
        extra={(
          <Space wrap>
            <Button href="/tasks">返回任务列表</Button>
            <Button href={`/tasks/${numericTaskId}`}>任务概览</Button>
          </Space>
        )}
      />

      {loading ? (
        <SectionCard title="任务结果">
          <div style={{ padding: '32px 0', textAlign: 'center' }}>
            <Spin />
          </div>
        </SectionCard>
      ) : null}

      {!loading && !task ? (
        <SectionCard title="任务结果">
          <Empty description="未查询到该任务的结果工作台数据。" />
        </SectionCard>
      ) : null}

      {!loading && task ? (
        <>
          <div className="summary-card-grid">
            <SummaryCard label="命中结果" value={summary.total} />
            <SummaryCard label="高风险" value={summary.highRisk} />
            <SummaryCard label="待处置" value={summary.pending} />
            <SummaryCard label="关联日志" value={summary.logs} />
          </div>

          <SectionCard title="任务摘要">
            <Space direction="vertical" size={10} style={{ width: '100%' }}>
              <Space wrap>
                <StatusBadge kind="task" value={task.status} />
                <span className="status-badge status-badge--neutral">任务编号</span>
              </Space>
              <Text strong>{task.task_no}</Text>
              <Text>创建人：{task.creator_name || '-'}</Text>
              <Text>创建时间：{task.created_at || '-'}</Text>
              <Text>执行状态：{taskStatusLabel(task.status)}</Text>
              <Text>扫描摘要：已扫描 {task.total_scanned ?? 0} 篇，命中 {task.hit_articles ?? 0} 篇 / {task.hit_count ?? 0} 次。</Text>
            </Space>
          </SectionCard>

          <SectionCard title="结果列表">
            <Space size={8} wrap style={{ marginBottom: 12 }}>
              <Button onClick={() => setSelectedResultIds(results.map((item) => item.id))}>本页全选</Button>
              <Button disabled={selectedResultIds.length === 0} onClick={() => void runBatchAction('ignore', selectedResultIds)}>
                批量忽略
              </Button>
              <Button disabled={selectedResultIds.length === 0} onClick={() => void runBatchAction('process', selectedResultIds)}>
                批量标记已处理
              </Button>
              <Button
                type="primary"
                danger
                disabled={selectedResultIds.length === 0}
                onClick={() => void runBatchAction('offline', selectedResultIds)}
              >
                批量下线处置
              </Button>
            </Space>

            <ProTable<ResultRecord>
              rowKey="id"
              cardBordered={false}
              size="small"
              search={false}
              headerTitle={false}
              options={false}
              toolBarRender={false}
              dataSource={results}
              pagination={false}
              rowSelection={{
                selectedRowKeys: selectedResultIds,
                onChange: (keys) => setSelectedResultIds(keys.map((item) => Number(item)))
              }}
              columns={[
                {
                  title: '文章标题',
                  dataIndex: 'article_title'
                },
                {
                  title: '文稿ID',
                  dataIndex: 'article_id',
                  render: (_, record) => <Text>{record.article_id}</Text>
                },
                {
                  title: '风险等级',
                  dataIndex: 'risk_level',
                  render: (_, record) => <StatusBadge kind="risk" value={record.risk_level} />
                },
                {
                  title: '处置状态',
                  dataIndex: 'disposition_status',
                  render: (_, record) => (
                    <span className={`status-badge ${record.disposition_status === 'pending' ? 'status-badge--warning' : 'status-badge--success'}`}>
                      {record.disposition_status === 'pending' ? '待处置' : '已处置'}
                    </span>
                  )
                },
                {
                  title: '命中片段',
                  dataIndex: 'snippet',
                  render: (_, record) => (
                    <HitPreview
                      fieldName={record.preview_field_name}
                      keywordText={record.preview_keyword_text ?? record.matched_keyword}
                      matchedText={record.preview_matched_text ?? record.matched_keyword}
                      snippet={record.preview_snippet ?? record.snippet}
                      extraHitCount={record.extra_hit_count}
                    />
                  )
                },
                {
                  title: '操作',
                  valueType: 'option',
                  render: (_, record) => [
                    <Button
                      key="detail"
                      type="link"
                      href={`/articles/${record.article_id}?${new URLSearchParams({ return_to: returnTo }).toString()}`}
                    >
                      查看详情
                    </Button>,
                    <Button
                      key="rectify"
                      type="link"
                      href={`/articles/${record.article_id}/rectify?${new URLSearchParams({
                        return_to: returnTo,
                        task_id: String(numericTaskId),
                        result_id: String(record.id)
                      }).toString()}`}
                    >
                      进入整改
                    </Button>,
                    <Button key="offline" type="link" danger onClick={() => void runBatchAction('offline', [record.id])}>
                      下线处置
                    </Button>,
                    <Button key="ignore" type="link" onClick={() => void runBatchAction('ignore', [record.id])}>
                      忽略
                    </Button>,
                    <Button key="process" type="link" onClick={() => void runBatchAction('process', [record.id])}>
                      标记已处理
                    </Button>
                  ]
                }
              ]}
            />
          </SectionCard>

          <SectionCard title="关联日志">
            {logs.length ? (
              <Space direction="vertical" size={12} style={{ width: '100%' }}>
                {logs.map((item) => (
                  <div key={item.id} className="rectify-reference__item">
                    <Text strong>{item.summary}</Text>
                    <Text type="secondary">{item.operator_name || '未知操作人'} · {item.created_at || '-'}</Text>
                    <Text type="secondary">{item.before_state || '-'}{' -> '}{item.after_state || '-'}</Text>
                  </div>
                ))}
              </Space>
            ) : (
              <Empty description="当前任务暂无关联日志。" />
            )}
          </SectionCard>
        </>
      ) : null}
    </>
  );
}
