import { ProForm } from '@ant-design/pro-components';
import { Button, Form, Input, Space, Switch, Typography, message } from 'antd';
import { useEffect, useState } from 'react';

import { PageHeader } from '../../components/ui/page-header';
import { SectionCard } from '../../components/ui/section-card';
import { listKeywords, type KeywordRecord } from '../../services/keywords';
import { createTask } from '../../services/tasks';

const { Paragraph, Title } = Typography;

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
      <PageHeader
        title="新建检测任务"
        description="设置巡检范围、关键词条件与筛选口径，发起后将按异步任务方式执行。"
      />

      <div className="task-form-layout">
        <SectionCard title="任务配置">
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
              messageApi.success('检测任务已提交');
              return true;
            }}
          >
            <Form.Item label="所属机构">
              <Input aria-label="所属机构" value={draft.orgid} onChange={(event) => setDraft((current) => ({ ...current, orgid: event.target.value }))} />
            </Form.Item>
            <Form.Item label="关键词范围">
              <select
                aria-label="关键词范围"
                multiple
                className="task-form-layout__select"
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
            <Form.Item label="发布时间起">
              <Input aria-label="发布时间起" value={draft.publish_time_start} onChange={(event) => setDraft((current) => ({ ...current, publish_time_start: event.target.value }))} />
            </Form.Item>
            <Form.Item label="发布时间止">
              <Input aria-label="发布时间止" value={draft.publish_time_end} onChange={(event) => setDraft((current) => ({ ...current, publish_time_end: event.target.value }))} />
            </Form.Item>
            <Form.Item label="是否检索正文">
              <Switch aria-label="是否检索正文" checked={draft.include_body} onChange={(checked) => setDraft((current) => ({ ...current, include_body: checked }))} />
            </Form.Item>
            <Form.Item label="文章编号">
              <Input aria-label="文章编号" value={draft.article_id} onChange={(event) => setDraft((current) => ({ ...current, article_id: event.target.value }))} />
            </Form.Item>
            <Form.Item label="标题检索">
              <Input aria-label="标题检索" value={draft.title_like} onChange={(event) => setDraft((current) => ({ ...current, title_like: event.target.value }))} />
            </Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">
                提交任务
              </Button>
            </Space>
          </ProForm>
        </SectionCard>

        <SectionCard title="执行提示">
          <div className="task-form-layout__tips">
            <div>
              <Title level={5}>任务说明</Title>
              <Paragraph>任务提交后将按异步方式执行，不阻塞当前操作界面，适合批量巡检场景。</Paragraph>
            </div>
            <div>
              <Title level={5}>范围建议</Title>
              <Paragraph>如需精准定位，请同时设置关键词范围、时间区间与文章编号，避免误扫无关内容。</Paragraph>
            </div>
            <div>
              <Title level={5}>值守建议</Title>
              <Paragraph>任务提交后请及时返回“检测任务”页查看执行进展，并在“风险结果”页完成处置闭环。</Paragraph>
            </div>
          </div>
        </SectionCard>
      </div>
    </>
  );
}
