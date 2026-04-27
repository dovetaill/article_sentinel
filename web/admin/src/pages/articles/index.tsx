import { Button, Empty, Input, Space, Spin, Table, Typography } from 'antd';
import { useEffect, useRef, useState } from 'react';

import { SectionCard } from '../../components/ui/section-card';
import { StatusBadge } from '../../components/ui/status-badge';
import { ToolbarStrip } from '../../components/ui/toolbar-strip';
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

function normalizeArticleID(value: string) {
  const articleID = Number(value.trim());
  return Number.isInteger(articleID) && articleID > 0 ? articleID : undefined;
}

export default function ArticlesPage() {
  const [loading, setLoading] = useState(true);
  const [items, setItems] = useState<ArticleListItem[]>([]);
  const [draftTitle, setDraftTitle] = useState('');
  const [draftArticleID, setDraftArticleID] = useState('');
  const [submittedTitle, setSubmittedTitle] = useState('');
  const [submittedArticleID, setSubmittedArticleID] = useState<number | undefined>(undefined);
  const [reloadKey, setReloadKey] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [total, setTotal] = useState(0);
  const loadingRef = useRef(true);
  const requestIDRef = useRef(0);
  const { activeOrgId } = useOrgContext();

  const currentOrgId = activeOrgId ?? 29;

  function beginReload() {
    if (loadingRef.current) {
      return false;
    }
    loadingRef.current = true;
    setLoading(true);
    return true;
  }

  useEffect(() => {
    const requestID = requestIDRef.current + 1;
    requestIDRef.current = requestID;
    loadingRef.current = true;
    setLoading(true);
    void listArticles({
      orgid: currentOrgId,
      page,
      pageSize,
      article_id: submittedArticleID,
      title: submittedTitle || undefined
    })
      .then((result) => {
        if (requestID !== requestIDRef.current) {
          return;
        }
        setItems(result.items);
        setTotal(result.total);
      })
      .catch(() => {
        if (requestID !== requestIDRef.current) {
          return;
        }
        setItems([]);
        setTotal(0);
      })
      .finally(() => {
        if (requestID !== requestIDRef.current) {
          return;
        }
        loadingRef.current = false;
        setLoading(false);
      });
  }, [currentOrgId, page, pageSize, submittedArticleID, submittedTitle, reloadKey]);

  return (
    <>
      <SectionCard>
        <ToolbarStrip>
          <div className="toolbar-strip__group">
            <div className="toolbar-strip__controls">
              <Input
                aria-label="标题模糊查询"
                className="toolbar-strip__control"
                placeholder="按标题模糊查找"
                value={draftTitle}
                onChange={(event) => setDraftTitle(event.target.value)}
              />
              <Input
                aria-label="按文稿ID查询"
                className="toolbar-strip__control"
                inputMode="numeric"
                placeholder="按文稿ID查询"
                value={draftArticleID}
                onChange={(event) => setDraftArticleID(event.target.value)}
              />
            </div>
          </div>

          <div className="toolbar-strip__actions">
            <Button
              disabled={loading}
              onClick={() => {
                if (!beginReload()) {
                  return;
                }
                setDraftTitle('');
                setDraftArticleID('');
                setSubmittedTitle('');
                setSubmittedArticleID(undefined);
                setPage(1);
                setReloadKey((value) => value + 1);
              }}
            >
              重置
            </Button>
            <Button
              type="primary"
              disabled={loading}
              loading={loading}
              onClick={() => {
                if (!beginReload()) {
                  return;
                }
                setSubmittedTitle(draftTitle.trim());
                setSubmittedArticleID(normalizeArticleID(draftArticleID));
                setPage(1);
                setReloadKey((value) => value + 1);
              }}
            >
              查询文稿
            </Button>
          </div>
        </ToolbarStrip>

        {loading ? (
          <div style={{ padding: '32px 0', textAlign: 'center' }}>
            <Spin />
          </div>
        ) : null}

        {!loading && items.length === 0 ? <Empty description="暂无文章。" /> : null}

        {!loading && items.length > 0 ? (
          <Table<ArticleListItem>
            rowKey="id"
            pagination={{
              current: page,
              pageSize,
              total,
              disabled: loading,
              showSizeChanger: false,
              onChange: (nextPage, nextPageSize) => {
                if (!beginReload()) {
                  return;
                }
                setPage(nextPage);
                setPageSize(nextPageSize);
              }
            }}
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
                render: (value: number) => <Text>{value}</Text>
              },
              {
                title: '当前状态',
                dataIndex: 'state',
                render: (value?: number) => renderArticleState(value)
              },
              {
                title: '发布时间',
                dataIndex: 'publish_at_time',
                render: (value?: string) => value || '-'
              },
              {
                title: '最近巡检',
                dataIndex: 'latest_risk_level',
                render: (_, record) => record.latest_task_id
                  ? (
                    <Space size={8} wrap>
                      <Text>任务 #{record.latest_task_id}</Text>
                      {record.latest_risk_level ? <StatusBadge kind="risk" value={record.latest_risk_level} /> : null}
                    </Space>
                  )
                  : '-'
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
