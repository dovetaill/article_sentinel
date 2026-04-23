import { Button, Empty, Spin, Table, Typography } from 'antd';
import { useEffect, useState } from 'react';

import { SectionCard } from '../../components/ui/section-card';
import { StatusBadge } from '../../components/ui/status-badge';
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

export default function ArticlesPage() {
  const [loading, setLoading] = useState(true);
  const [items, setItems] = useState<ArticleListItem[]>([]);

  useEffect(() => {
    setLoading(true);
    void listArticles({ orgid: 100, page: 1, pageSize: 20 })
      .then((result) => setItems(result.items))
      .finally(() => setLoading(false));
  }, []);

  return (
    <>
      <SectionCard title="巡检稿件" description="基于现有巡检结果聚合出的稿件工作台视图。">
        {loading ? (
          <div style={{ padding: '32px 0', textAlign: 'center' }}>
            <Spin />
          </div>
        ) : null}

        {!loading && items.length === 0 ? <Empty description="暂无巡检稿件。" /> : null}

        {!loading && items.length > 0 ? (
          <Table<ArticleListItem>
            rowKey="article_id"
            pagination={false}
            columns={[
              {
                title: '文稿标题',
                dataIndex: 'article_title',
                render: (_, record) => (
                  <a href={`/articles/${record.article_id}`}>{record.article_title}</a>
                )
              },
              {
                title: '文稿编号',
                dataIndex: 'article_id',
                render: (value: number) => <Text>#{value}</Text>
              },
              {
                title: '当前状态',
                dataIndex: 'article_state',
                render: (value?: number) => renderArticleState(value)
              },
              {
                title: '最近风险',
                dataIndex: 'risk_level',
                render: (_, record) => <StatusBadge kind="risk" value={record.risk_level} />
              },
              { title: '命中次数', dataIndex: 'hit_count' },
              {
                title: '最近任务',
                dataIndex: 'latest_task_id',
                render: (value: number) => `#${value}`
              },
              {
                title: '操作',
                key: 'actions',
                render: (_, record) => (
                  <Button type="link" href={`/articles/${record.article_id}`}>
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
