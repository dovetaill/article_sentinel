import { ProTable } from '@ant-design/pro-components';
import { Button, Modal, Space, Tag, Typography, message } from 'antd';
import { useMemo, useRef, useState } from 'react';

import ResultDetailDrawer from './detail';
import { batchOfflineResults, listResults, type ResultRecord } from '../../services/results';

type ActionRef = {
  reload?: () => void;
};

const { Text } = Typography;

const riskColors: Record<string, string> = {
  low: 'green',
  medium: 'gold',
  high: 'red'
};

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
  const confirmTitle = useMemo(() => confirmIds.length > 1 ? 'Confirm Batch Offline' : 'Confirm Offline', [confirmIds.length]);

  return (
    <>
      {contextHolder}
      <ProTable<ResultRecord>
        rowKey="id"
        actionRef={actionRef as never}
        cardBordered
        search={false}
        headerTitle="Hit Results"
        rowSelection={{
          selectedRowKeys: selectedResultIds,
          onChange: (keys) => setSelectedResultIds(keys.map((item) => Number(item)))
        }}
        toolBarRender={() => [
          <Text key="selected">{selectedCount} selected</Text>,
          <Button
            key="select-page"
            onClick={() => setSelectedResultIds(pageRows.map((item) => item.id))}
          >
            Select Current Page
          </Button>,
          <Button
            key="batch-offline"
            type="primary"
            danger
            disabled={selectedCount === 0}
            onClick={() => setConfirmIds(selectedResultIds)}
          >
            Batch Offline
          </Button>
        ]}
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
          { title: 'Article', dataIndex: 'article_title' },
          {
            title: 'Risk',
            dataIndex: 'risk_level',
            render: (_, record) => <Tag color={riskColors[record.risk_level] ?? 'default'}>{record.risk_level}</Tag>
          },
          { title: 'Disposition', dataIndex: 'disposition_status' },
          { title: 'Hits', dataIndex: 'hit_count' },
          {
            title: 'Snippet',
            dataIndex: 'snippet',
            render: (_, record) => renderSnippet(record.snippet, record.matched_keyword)
          },
          {
            title: 'Action',
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
                View Detail
              </Button>,
              <Button key="offline" type="link" danger onClick={() => setConfirmIds([record.id])}>
                Offline
              </Button>,
              <Button key="rectify" type="link" href={`/articles/${record.article_id}/rectify`}>
                Rectify
              </Button>
            ]
          }
        ]}
      />

      <Modal
        open={confirmOpen}
        title={confirmTitle}
        okText="Confirm Offline"
        cancelText="Cancel"
        onCancel={() => setConfirmIds([])}
        onOk={async () => {
          await batchOfflineResults({
            orgid: 100,
            result_ids: confirmIds,
            reason: 'manual batch offline'
          });
          messageApi.success('Offline action submitted');
          setSelectedResultIds([]);
          setConfirmIds([]);
          actionRef.current.reload?.();
        }}
      >
        <p>Offline {confirmIds.length} selected article(s)?</p>
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
