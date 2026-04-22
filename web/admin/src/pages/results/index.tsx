import { ProTable } from '@ant-design/pro-components';
import { Button, Modal, Typography, message } from 'antd';
import { useMemo, useRef, useState } from 'react';

import { PageHeader } from '../../components/ui/page-header';
import { SectionCard } from '../../components/ui/section-card';
import { StatusBadge } from '../../components/ui/status-badge';
import { SummaryCard } from '../../components/ui/summary-card';
import { ToolbarStrip } from '../../components/ui/toolbar-strip';
import ResultDetailDrawer from './detail';
import { batchOfflineResults, listResults, type ResultRecord } from '../../services/results';

type ActionRef = {
  reload?: () => void;
};

const { Text } = Typography;

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

export default function ResultsPage() {
  const actionRef = useRef<ActionRef>({});
  const [messageApi, contextHolder] = message.useMessage();
  const [pageRows, setPageRows] = useState<ResultRecord[]>([]);
  const [selectedResultIds, setSelectedResultIds] = useState<number[]>([]);
  const [confirmIds, setConfirmIds] = useState<number[]>([]);
  const [detailResultId, setDetailResultId] = useState<number>();
  const [detailOpen, setDetailOpen] = useState(false);

  const selectedCount = selectedResultIds.length;
  const confirmOpen = confirmIds.length > 0;
  const confirmTitle = useMemo(() => confirmIds.length > 1 ? '批量下线处置' : '下线处置', [confirmIds.length]);
  const summary = useMemo(() => {
    const highRisk = pageRows.filter((item) => item.risk_level === 'high').length;
    const pending = pageRows.filter((item) => item.disposition_status === 'pending').length;
    const hitCount = pageRows.reduce((total, item) => total + (item.hit_count ?? 0), 0);

    return {
      total: pageRows.length,
      highRisk,
      pending,
      hitCount
    };
  }, [pageRows]);

  return (
    <>
      {contextHolder}
      <PageHeader
        title="风险结果"
        description="集中查看命中文稿、风险等级与处置状态，支持值守人员按批次完成研判、下线与整改。"
      />

      <div className="summary-card-grid">
        <SummaryCard label="本页结果数" value={summary.total} helper="当前分页已加载记录数量" />
        <SummaryCard label="高风险" value={summary.highRisk} helper="需优先复核的记录数" />
        <SummaryCard label="待处置" value={summary.pending} helper="尚未完成处置的记录数" />
        <SummaryCard label="命中总量" value={summary.hitCount} helper="当前分页累计命中次数" />
      </div>

      <SectionCard title="结果列表">
        <ToolbarStrip>
          <Text>已选 {selectedCount} 项</Text>
          <div className="toolbar-strip__actions">
            <Button
              key="select-page"
              onClick={() => setSelectedResultIds(pageRows.map((item) => item.id))}
            >
              本页全选
            </Button>
            <Button
              key="batch-offline"
              type="primary"
              danger
              disabled={selectedCount === 0}
              onClick={() => setConfirmIds(selectedResultIds)}
            >
              批量下线处置
            </Button>
          </div>
        </ToolbarStrip>

        <ProTable<ResultRecord>
          rowKey="id"
          actionRef={actionRef as never}
          cardBordered={false}
          search={false}
          headerTitle={false}
          options={{ density: false, fullScreen: false }}
          toolBarRender={false}
          rowSelection={{
            selectedRowKeys: selectedResultIds,
            onChange: (keys) => setSelectedResultIds(keys.map((item) => Number(item)))
          }}
          request={async (params) => {
            const result = await listResults({
              orgid: 100,
              page: params.current,
              pageSize: params.pageSize
            });
            setPageRows(result.items);
            return {
              data: result.items,
              success: true,
              total: result.total
            };
          }}
          columns={[
            { title: '文章标题', dataIndex: 'article_title' },
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
            { title: '命中次数', dataIndex: 'hit_count' },
            {
              title: '命中片段',
              dataIndex: 'snippet',
              render: (_, record) => renderSnippet(record.snippet, record.matched_keyword)
            },
            {
              title: '操作',
              valueType: 'option',
              render: (_, record) => [
                <Button
                  key="detail"
                  type="link"
                  onClick={() => {
                    setDetailResultId(record.id);
                    setDetailOpen(true);
                  }}
                >
                  查看详情
                </Button>,
                <Button key="offline" type="link" danger onClick={() => setConfirmIds([record.id])}>
                  下线处置
                </Button>,
                <Button key="rectify" type="link" href={`/articles/${record.article_id}/rectify`}>
                  进入整改
                </Button>
              ]
            }
          ]}
        />
      </SectionCard>

      <Modal
        open={confirmOpen}
        title={confirmTitle}
        okText="确认处置"
        cancelText="取消"
        onCancel={() => setConfirmIds([])}
        onOk={async () => {
          await batchOfflineResults({
            orgid: 100,
            result_ids: confirmIds,
            reason: 'manual batch offline'
          });
          messageApi.success('处置请求已提交');
          setSelectedResultIds([]);
          setConfirmIds([]);
          actionRef.current.reload?.();
        }}
      >
        <p>确认对 {confirmIds.length} 篇文章执行下线处置？</p>
      </Modal>

      <ResultDetailDrawer
        open={detailOpen}
        resultId={detailResultId}
        orgid={100}
        onClose={() => setDetailOpen(false)}
      />
    </>
  );
}
