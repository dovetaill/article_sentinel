import { Button, Descriptions, Empty, List, Space, Spin, Tabs, Typography } from 'antd';
import { useEffect, useMemo, useState } from 'react';
import { useLocation, useParams } from 'react-router-dom';

import { HitPreview } from '../../components/ui/hit-preview';
import { PageHeader } from '../../components/ui/page-header';
import { SectionCard } from '../../components/ui/section-card';
import { StatusBadge } from '../../components/ui/status-badge';
import { SummaryCard } from '../../components/ui/summary-card';
import { useOrgContext } from '../../context/org-context';
import { formatInspectionSnapshot } from '../../lib/inspection-snapshot';
import { listOperationLogs, type OperationLogRecord } from '../../services/logs';
import { listResults, type ResultRecord } from '../../services/results';
import { getTaskDetail, type TaskRecord } from '../../services/tasks';

const { Text } = Typography;

type SummaryMetric = {
  label: string;
  value: string | number;
  helper?: string;
};

function taskStatusLabel(status: string | undefined) {
  if (status === 'running') return '执行中';
  if (status === 'success') return '已完成';
  if (status === 'failed') return '执行失败';
  return status || '-';
}

export default function TaskDetailPage() {
  const { taskId } = useParams();
  const location = useLocation();
  const [loading, setLoading] = useState(true);
  const [task, setTask] = useState<TaskRecord | null>(null);
  const [results, setResults] = useState<ResultRecord[]>([]);
  const [logs, setLogs] = useState<OperationLogRecord[]>([]);
  const { activeOrgId } = useOrgContext();

  const currentOrgId = activeOrgId ?? 29;
  const returnTo = `${location.pathname}${location.search}`;

  useEffect(() => {
    if (!taskId) {
      setLoading(false);
      return;
    }

    setLoading(true);

    void Promise.all([
      getTaskDetail(Number(taskId), currentOrgId),
      listResults({ orgid: currentOrgId, task_id: Number(taskId), page: 1, pageSize: 20 }),
      listOperationLogs({ orgid: currentOrgId, task_id: Number(taskId), page: 1, pageSize: 20 })
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
  }, [currentOrgId, taskId]);

  const metrics = useMemo<SummaryMetric[]>(() => {
    if (!task) {
      return [];
    }

    return [
      { label: '任务编号', value: task.task_no },
      { label: '执行状态', value: taskStatusLabel(task.status) },
      { label: '已扫描数量', value: task.total_scanned ?? 0 },
      { label: '命中结果', value: `${task.hit_articles ?? 0} / ${task.hit_count ?? 0}` }
    ];
  }, [task]);

  const tabItems = useMemo(
    () => [
      {
        key: 'results',
        label: '命中结果',
        children: results.length ? (
          <List
            className="detail-list"
            dataSource={results}
            renderItem={(item) => (
              <List.Item>
                <Space direction="vertical" size={4} style={{ width: '100%' }}>
                  <a
                    className="detail-list__title"
                    href={`/articles/${item.article_id}?${new URLSearchParams({ return_to: returnTo }).toString()}`}
                  >
                    {item.article_title}
                  </a>
                  <Space size={8} wrap>
                    <StatusBadge kind="risk" value={item.risk_level} />
                    <Text type="secondary">命中 {item.hit_count} 次</Text>
                    <Text type="secondary">{item.disposition_status === 'pending' ? '待处置' : '已处置'}</Text>
                  </Space>
                  <HitPreview
                    fieldName={item.preview_field_name}
                    keywordText={item.preview_keyword_text ?? item.matched_keyword}
                    matchedText={item.preview_matched_text ?? item.matched_keyword}
                    snippet={item.preview_snippet ?? item.snippet}
                    extraHitCount={item.extra_hit_count}
                  />
                </Space>
              </List.Item>
            )}
          />
        ) : <Empty description="当前任务暂无命中结果。" />
      },
      {
        key: 'rule-snapshot',
        label: '规则快照',
        children: <pre className="detail-code-block">{formatInspectionSnapshot(task?.rule_snapshot, '暂无规则快照。')}</pre>
      },
      {
        key: 'request-snapshot',
        label: '请求快照',
        children: <pre className="detail-code-block">{formatInspectionSnapshot(task?.request_snapshot, '暂无请求快照。')}</pre>
      },
      {
        key: 'logs',
        label: '关联日志',
        children: logs.length ? (
          <List
            className="detail-list"
            dataSource={logs}
            renderItem={(item) => (
              <List.Item>
                <Space direction="vertical" size={4} style={{ width: '100%' }}>
                  <Text strong>{item.summary}</Text>
                  <Text type="secondary">
                    {item.operator_name || '未知操作人'} · {item.created_at || '-'}
                  </Text>
                  <Text type="secondary">
                    {item.before_state || '-'} → {item.after_state || '-'}
                  </Text>
                </Space>
              </List.Item>
            )}
          />
        ) : <Empty description="当前任务暂无关联日志。" />
      }
    ],
    [logs, results, task?.request_snapshot, task?.rule_snapshot],
  );

  return (
    <>
      <PageHeader
        title="任务详情"
        extra={(
          <Space wrap>
            <Button href="/tasks">返回任务列表</Button>
            <Button type="primary" href={`/tasks/${task?.id ?? taskId}/results`}>
              查看任务结果
            </Button>
          </Space>
        )}
      />

      {loading ? (
        <SectionCard title="任务详情">
          <div style={{ padding: '32px 0', textAlign: 'center' }}>
            <Spin />
          </div>
        </SectionCard>
      ) : null}

      {!loading && !task ? (
        <SectionCard title="任务详情">
          <Empty description="未查询到该任务的详细记录。" />
        </SectionCard>
      ) : null}

      {!loading && task ? (
        <>
          <div className="summary-card-grid">
            {metrics.map((item) => (
              <SummaryCard key={item.label} label={item.label} value={item.value} helper={item.helper} />
            ))}
          </div>

          <div className="detail-workspace">
            <SectionCard title={task.task_no}>
              <Space direction="vertical" size={16} style={{ width: '100%' }}>
                <Space wrap>
                  <StatusBadge kind="task" value={task.status} />
                  <span className="status-badge status-badge--neutral">{task.hit_articles ?? 0} 篇命中文稿</span>
                </Space>

                <Descriptions column={1} size="small">
                  <Descriptions.Item label="任务编号">{task.task_no}</Descriptions.Item>
                  <Descriptions.Item label="创建人">{task.creator_name || '-'}</Descriptions.Item>
                  <Descriptions.Item label="创建时间">{task.created_at || '-'}</Descriptions.Item>
                  <Descriptions.Item label="执行摘要">
                    已扫描 {task.total_scanned ?? 0} 篇文章，累计命中 {task.hit_count ?? 0} 次。
                  </Descriptions.Item>
                </Descriptions>
              </Space>
            </SectionCard>

            <div className="detail-workspace__side">
              <SectionCard title="当前状态">
                <Space direction="vertical" size={10} style={{ width: '100%' }}>
                  <Text>执行状态：{taskStatusLabel(task.status)}</Text>
                  <Text>命中文稿：{task.hit_articles ?? 0} 篇</Text>
                  <Text>命中次数：{task.hit_count ?? 0} 次</Text>
                  <Text>最近关联日志：{logs[0]?.summary || '暂无记录'}</Text>
                </Space>
              </SectionCard>

              <SectionCard title="快捷入口">
                <Space direction="vertical" size={10} style={{ width: '100%' }}>
                  <Button href={`/tasks/${task.id}/results`}>前往任务结果</Button>
                  <Button href="/articles">前往文稿列表</Button>
                  {results[0] ? (
                    <Button
                      type="link"
                      href={`/articles/${results[0].article_id}?${new URLSearchParams({ return_to: returnTo }).toString()}`}
                    >
                      查看首条命中文稿
                    </Button>
                  ) : (
                    <Text type="secondary">当前暂无可跳转的命中文稿。</Text>
                  )}
                </Space>
              </SectionCard>
            </div>
          </div>

          <SectionCard title="任务记录">
            <Tabs defaultActiveKey="results" items={tabItems} />
          </SectionCard>
        </>
      ) : null}
    </>
  );
}
