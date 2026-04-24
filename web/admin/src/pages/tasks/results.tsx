import { ProTable } from '@ant-design/pro-components';
import { Button, Empty, Space, Typography, message } from 'antd';
import { useEffect, useMemo, useState } from 'react';
import { useParams } from 'react-router-dom';

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

const { Paragraph, Text } = Typography;

function taskStatusLabel(status: string | undefined) {
  if (status === 'running') return '执行中';
  if (status === 'success') return '已完成';
  if (status === 'failed') return '执行失败';
  return status || '-';
}

function renderSnippet(snippet?: string, keyword?: string) {
  if (!snippet || !keyword) {
    return snippet || '-';
  }

  const parts = snippet.split(new RegExp(`(${escapePattern(keyword)})`, 'gi'));
  return parts.map((part, index) => (
    part.toLowerCase() === keyword.toLowerCase()
      ? <mark key={`${keyword}-${index}`}>{part}</mark>
      : <span key={`${keyword}-${index}`}>{part}</span>
  ));
}

function escapePattern(value: string) {
  return value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}

export default function TaskResultsPage() {
  const { taskId } = useParams();
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
        description="围绕单个任务查看命中结果、执行摘要和处置日志。"
        extra={(
          <Space wrap>
            <Button href="/tasks">返回任务列表</Button>
            <Button href={`/tasks/${numericTaskId}`}>任务概览</Button>
          </Space>
        )}
      />

      {loading ? (
        <SectionCard title="任务结果" description="正在加载任务结果工作台。">
          <Paragraph>正在汇总任务、命中结果与日志记录。</Paragraph>
        </SectionCard>
      ) : null}

      {!loading && !task ? (
        <SectionCard title="任务结果" description="未返回可展示的任务结果。">
          <Empty description="未查询到该任务的结果工作台数据。" />
        </SectionCard>
      ) : null}

      {!loading && task ? (
        <>
          <div className="summary-card-grid">
            <SummaryCard label="命中结果" value={summary.total} helper="当前任务命中结果总数" />
            <SummaryCard label="高风险" value={summary.highRisk} helper="需优先处置的命中记录" />
            <SummaryCard label="待处置" value={summary.pending} helper="尚未完成处置的记录" />
            <SummaryCard label="关联日志" value={summary.logs} helper="当前任务已记录的操作条目" />
          </div>

          <SectionCard title="任务摘要" description="查看当前任务的执行概况与责任信息。">
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

          <SectionCard title="结果列表" description="在当前任务上下文中完成单条或批量处置。">
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
                  render: (_, record) => renderSnippet(record.snippet, record.matched_keyword)
                },
                {
                  title: '操作',
                  valueType: 'option',
                  render: (_, record) => [
                    <Button key="detail" type="link" href={`/articles/${record.article_id}`}>
                      查看详情
                    </Button>,
                    <Button key="rectify" type="link" href={`/articles/${record.article_id}/rectify`}>
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

          <SectionCard title="关联日志" description="跟踪该任务下的处置动作和状态变化。">
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
