import { PageContainer, ProDescriptions } from '@ant-design/pro-components';
import { Button, Card, Empty, List, Space, Spin, Statistic, Tabs, Tag, Typography } from 'antd';
import { useEffect, useMemo, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';

import HitPreview from '@/components/HitPreview';
import SnapshotViewer from '@/components/SnapshotViewer';
import StatusTag from '@/components/StatusTag';
import { listOperationLogs, type OperationLogRecord } from '@/services/logs';
import { listResults, type ResultRecord } from '@/services/results';
import { getTaskDetail, type TaskRecord } from '@/services/tasks';

const { Title, Paragraph, Text } = Typography;

type TaskTabKey = 'results' | 'rule-snapshot' | 'request-snapshot' | 'logs';

function taskStatusLabel(status: string | undefined) {
  if (status === 'running') return '执行中';
  if (status === 'success') return '已完成';
  if (status === 'failed') return '执行失败';
  if (status === 'pending') return '待执行';
  return status || '-';
}

export default function TaskDetailPage() {
  const navigate = useNavigate();
  const { taskId } = useParams();
  const numericTaskId = Number(taskId || 0) || 0;
  const [loading, setLoading] = useState(true);
  const [task, setTask] = useState<TaskRecord | null>(null);
  const [results, setResults] = useState<ResultRecord[]>([]);
  const [logs, setLogs] = useState<OperationLogRecord[]>([]);
  const [activeTab, setActiveTab] = useState<TaskTabKey>('results');

  useEffect(() => {
    if (!numericTaskId) {
      setLoading(false);
      setTask(null);
      setResults([]);
      setLogs([]);
      return;
    }

    let cancelled = false;
    setLoading(true);

    void Promise.all([
      getTaskDetail(numericTaskId),
      listResults({ task_id: numericTaskId, page: 1, pageSize: 20 }),
      listOperationLogs({ task_id: numericTaskId, page: 1, pageSize: 20 })
    ])
      .then(([taskDetail, resultList, logList]) => {
        if (cancelled) {
          return;
        }

        setTask(taskDetail);
        setResults(resultList.items ?? []);
        setLogs(logList.items ?? []);
      })
      .catch(() => {
        if (cancelled) {
          return;
        }

        setTask(null);
        setResults([]);
        setLogs([]);
      })
      .finally(() => {
        if (!cancelled) {
          setLoading(false);
        }
      });

    return () => {
      cancelled = true;
    };
  }, [numericTaskId]);

  const summary = useMemo(() => {
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
                  <Button
                    type="link"
                    className="detail-list__title"
                    onClick={() => navigate(`/content/articles/${item.article_id}`)}
                  >
                    {item.article_title}
                  </Button>
                  <Space size={8} wrap>
                    <StatusTag kind="risk" value={item.risk_level} />
                    <Text type="secondary">命中 {item.hit_count} 次</Text>
                    <Tag bordered={false} color={item.disposition_status === 'pending' ? 'warning' : 'success'}>
                      {item.disposition_status === 'pending' ? '待处置' : '已处置'}
                    </Tag>
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
        ) : (
          <Empty description="当前任务暂无命中结果。" />
        )
      },
      {
        key: 'rule-snapshot',
        label: '规则快照',
        children: <SnapshotViewer value={task?.rule_snapshot} emptyText="暂无规则快照。" />
      },
      {
        key: 'request-snapshot',
        label: '请求快照',
        children: <SnapshotViewer value={task?.request_snapshot} emptyText="暂无请求快照。" />
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
        ) : (
          <Empty description="当前任务暂无关联日志。" />
        )
      }
    ],
    [logs, navigate, results, task?.request_snapshot, task?.rule_snapshot]
  );

  return (
    <PageContainer title={false}>
      <div className="admin-domain-page">
        <div className="admin-domain-page__head">
          <div>
            <Title level={3} className="admin-domain-page__title">
              任务详情
            </Title>
            <Paragraph className="admin-domain-page__desc">
              查看当前批次的规则快照、执行摘要、命中结果与关联日志。
            </Paragraph>
          </div>
          <Space wrap>
            <Button onClick={() => navigate('/inspection/tasks')}>返回任务列表</Button>
            <Button
              type="primary"
              onClick={() => navigate(`/inspection/results${numericTaskId ? `?task_id=${numericTaskId}` : ''}`)}
            >
              查看任务结果
            </Button>
          </Space>
        </div>

        {loading ? (
          <Card className="admin-filter-card" variant="borderless">
            <div style={{ padding: '32px 0', textAlign: 'center' }}>
              <Spin />
            </div>
          </Card>
        ) : null}

        {!loading && !task ? (
          <Card className="admin-filter-card" variant="borderless">
            <Empty description="未查询到该任务的详细记录。" />
          </Card>
        ) : null}

        {!loading && task ? (
          <>
            <Space size={16} wrap>
              {summary.map((item) => (
                <Card key={item.label} variant="borderless">
                  <Statistic title={item.label} value={item.value} />
                </Card>
              ))}
            </Space>

            <div className="admin-detail-layout">
              <Card className="admin-filter-card admin-detail-layout__main" variant="borderless">
                <Space direction="vertical" size={16} style={{ width: '100%' }}>
                  <Space wrap>
                    <StatusTag kind="task" value={task.status} />
                    <Tag bordered={false}>{task.hit_articles ?? 0} 篇命中文稿</Tag>
                  </Space>
                  <ProDescriptions<TaskRecord>
                    dataSource={task}
                    column={1}
                    columns={[
                      { title: '任务编号', dataIndex: 'task_no' },
                      { title: '创建人', dataIndex: 'creator_name' },
                      { title: '创建时间', dataIndex: 'created_at' },
                      {
                        title: '执行摘要',
                        render: () => `已扫描 ${task.total_scanned ?? 0} 篇文章，累计命中 ${task.hit_count ?? 0} 次。`
                      }
                    ]}
                  />
                </Space>
              </Card>

              <Card className="admin-filter-card admin-detail-layout__side" variant="borderless">
                <Space direction="vertical" size={10} style={{ width: '100%' }}>
                  <Text>执行状态：{taskStatusLabel(task.status)}</Text>
                  <Text>命中文稿：{task.hit_articles ?? 0} 篇</Text>
                  <Text>命中次数：{task.hit_count ?? 0} 次</Text>
                  <Text>最近关联日志：{logs[0]?.summary || '暂无记录'}</Text>
                </Space>
              </Card>
            </div>

            <Card className="admin-filter-card" variant="borderless">
              <Tabs activeKey={activeTab} onChange={(key) => setActiveTab(key as TaskTabKey)} items={tabItems} />
            </Card>
          </>
        ) : null}
      </div>
    </PageContainer>
  );
}
