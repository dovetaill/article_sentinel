import { ProForm } from '@ant-design/pro-components';
import { Button, Card, Form, Input, Space, Switch, message } from 'antd';
import { useEffect, useState } from 'react';

import { listKeywords, type KeywordRecord } from '../../services/keywords';
import { createTask } from '../../services/tasks';

type TaskDraft = {
  orgid: string;
  keyword_ids: number[];
  publish_time_start: string;
  publish_time_end: string;
  include_body: boolean;
  article_id: string;
  title_like: string;
};

export default function NewTaskPage() {
  const [messageApi, contextHolder] = message.useMessage();
  const [keywords, setKeywords] = useState<KeywordRecord[]>([]);
  const [draft, setDraft] = useState<TaskDraft>({
    orgid: '',
    keyword_ids: [],
    publish_time_start: '',
    publish_time_end: '',
    include_body: false,
    article_id: '',
    title_like: ''
  });

  useEffect(() => {
    void listKeywords({ orgid: 100, page: 1, pageSize: 100 }).then((result) => setKeywords(result.items));
  }, []);

  return (
    <>
      {contextHolder}
      <Card title="Launch Inspection" variant="borderless">
        <ProForm
          submitter={false}
          onFinish={async () => {
            await createTask({
              orgid: Number(draft.orgid),
              keyword_ids: draft.keyword_ids,
              publish_time_start: draft.publish_time_start || undefined,
              publish_time_end: draft.publish_time_end || undefined,
              include_body: draft.include_body,
              article_id: draft.article_id ? Number(draft.article_id) : undefined,
              title_like: draft.title_like || undefined,
              article_state: 9
            });
            messageApi.success('Inspection launched');
            return true;
          }}
        >
          <Form.Item label="OrgID">
            <Input aria-label="OrgID" value={draft.orgid} onChange={(event) => setDraft((current) => ({ ...current, orgid: event.target.value }))} />
          </Form.Item>
          <Form.Item label="Keyword Set">
            <select
              aria-label="Keyword Set"
              multiple
              value={draft.keyword_ids.map(String)}
              onChange={(event) => {
                const values = Array.from(event.currentTarget.selectedOptions, (option) => Number(option.value));
                setDraft((current) => ({ ...current, keyword_ids: values }));
              }}
            >
              {keywords.map((item) => (
                <option key={item.id} value={item.id}>
                  {item.name}
                </option>
              ))}
            </select>
          </Form.Item>
          <Form.Item label="Publish Start">
            <Input aria-label="Publish Start" value={draft.publish_time_start} onChange={(event) => setDraft((current) => ({ ...current, publish_time_start: event.target.value }))} />
          </Form.Item>
          <Form.Item label="Publish End">
            <Input aria-label="Publish End" value={draft.publish_time_end} onChange={(event) => setDraft((current) => ({ ...current, publish_time_end: event.target.value }))} />
          </Form.Item>
          <Form.Item label="Include Body">
            <Switch aria-label="Include Body" checked={draft.include_body} onChange={(checked) => setDraft((current) => ({ ...current, include_body: checked }))} />
          </Form.Item>
          <Form.Item label="Article ID">
            <Input aria-label="Article ID" value={draft.article_id} onChange={(event) => setDraft((current) => ({ ...current, article_id: event.target.value }))} />
          </Form.Item>
          <Form.Item label="Title Like">
            <Input aria-label="Title Like" value={draft.title_like} onChange={(event) => setDraft((current) => ({ ...current, title_like: event.target.value }))} />
          </Form.Item>
          <Space>
            <Button type="primary" htmlType="submit">
              Launch Inspection
            </Button>
          </Space>
        </ProForm>
      </Card>
    </>
  );
}
