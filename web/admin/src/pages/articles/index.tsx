import { Button, Empty, Spin, Table, Typography } from 'antd';
import { useEffect, useState } from 'react';

import { SectionCard } from '../../components/ui/section-card';
import { StatusBadge } from '../../components/ui/status-badge';
import { useOrgContext } from '../../context/org-context';
import { listArticles, type ArticleListItem } from '../../services/articles';

const { Text } = Typography;

function renderArticleState(value?: number) {
  switch (value) {
    case 9:
      return '已发布';
    case 8:
      return '已下线';
    case 1:
      return '待审';
    default:
      return '-';
  }
}

function renderDispositionStatus(value?: string) {
  if (value === 'pending') {
    return <span className="status-badge status-badge--warning">待处置</span>;
  }

  if (value === 'processed') {
    return <span className="status-badge status-badge--success">已处置</span>;
  }

  return '-';
}

export default function ArticlesPage() {
  const [loading, setLoading] = useState(true);
  const [items, setItems] = useState<ArticleListItem[]>([]);
  const { activeOrgId } = useOrgContext();

  const currentOrgId = activeOrgId ?? 29;

  useEffect(() => {
    setLoading(true);
    void listArticles({ orgid: currentOrgId, page: 1, pageSize: 20 })
      .then((result) => setItems(result.items))
      .catch(() => setItems([]))
      .finally(() => setLoading(false));
  }, [currentOrgId]);

  return (
    <>
      <SectionCard title="文章中心" description="查看真实稿件元数据，并结合最近巡检结果快速进入处置。">
        {loading ? (
          <div style={{ padding: '32px 0', textAlign: 'center' }}>
            <Spin />
          </div>
        ) : null}

        {!loading && items.length === 0 ? <Empty description="暂无文章。" /> : null}

        {!loading && items.length > 0 ? (
          <Table<ArticleListItem>
            rowKey="id"
            pagination={false}
            columns={[
              {
                title: '文章标题',
                dataIndex: 'title',
                render: (_, record) => (
                  <a href={`/articles/${record.id}`}>{record.title}</a>
                )
              },
              {
                title: '文章编号',
                dataIndex: 'id',
                render: (value: number) => <Text>#{value}</Text>
              },
              {
                title: '当前状态',
                dataIndex: 'state',
                render: (value?: number) => renderArticleState(value)
              },
              {
                title: '最新风险',
                dataIndex: 'latest_risk_level',
                render: (value?: string) => value ? <StatusBadge kind="risk" value={value} /> : '-'
              },
              {
                title: '最近任务',
                dataIndex: 'latest_task_id',
                render: (value?: number) => value ? `#${value}` : '-'
              },
              {
                title: '处置状态',
                dataIndex: 'latest_disposition_status',
                render: (value?: string) => renderDispositionStatus(value)
              },
              {
                title: '操作',
                key: 'actions',
                render: (_, record) => (
                  <Button type="link" href={`/articles/${record.id}`}>
                    查看详情
                  </Button>
                )
              }
            ]}
            dataSource={items}
          />
        ) : null}
      </SectionCard>
    </>
  );
}
