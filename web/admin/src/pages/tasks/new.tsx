import { ProForm } from '@ant-design/pro-components';
import { Button, Form, Input, Select, Space, Switch, Typography, message } from 'antd';
import { useEffect, useState } from 'react';

import { PageHeader } from '../../components/ui/page-header';
import { SectionCard } from '../../components/ui/section-card';
import { useOrgContext } from '../../context/org-context';
import { listKeywords, type KeywordRecord } from '../../services/keywords';
import { createTask } from '../../services/tasks';

const { Paragraph, Title } = Typography;

type TaskDraft = {
  keyword_ids: number[];
  publish_time_start: string;
  publish_time_end: string;
  include_body: boolean;
};

export default function NewTaskPage() {
  const [messageApi, contextHolder] = message.useMessage();
  const [keywords, setKeywords] = useState<KeywordRecord[]>([]);
  const [draft, setDraft] = useState<TaskDraft>({
    keyword_ids: [],
    publish_time_start: '',
    publish_time_end: '',
    include_body: false
  });
  const { activeOrgId, activeOrgName } = useOrgContext();

  const currentOrgId = activeOrgId ?? 29;
  const currentOrgName = activeOrgName || '一县一端';

  useEffect(() => {
    void listKeywords({ orgid: currentOrgId, page: 1, pageSize: 100, enabled: true })
      .then((result) => {
        setKeywords(result.items.filter((item) => item.enabled));
      })
      .catch(() => {
        setKeywords([]);
      });
  }, [currentOrgId]);

  const keywordOptions = keywords.map((item) => ({
    value: item.id,
    label: `${item.category_name || '未分类规则'} / ${item.name}`
  }));

  async function submitTask() {
    if (draft.keyword_ids.length === 0) {
      messageApi.error('请先选择至少一条规则');
      return;
    }

    await createTask({
      orgid: currentOrgId,
      keyword_ids: draft.keyword_ids,
      publish_time_start: draft.publish_time_start || undefined,
      publish_time_end: draft.publish_time_end || undefined,
      include_body: draft.include_body,
      article_state: 9
    });
    messageApi.success('检测任务已提交');
  }

  return (
    <>
      {contextHolder}
      <PageHeader
        title="新建检测任务"
        description="选择当前机构的规则集合并设定时间范围，任务提交后将在结果工作台继续处置。"
        extra={<Button href="/tasks">返回任务列表</Button>}
      />

      <div className="task-form-layout">
        <SectionCard title="任务配置" description="填写巡检条件后即可发起一次新的异步扫描。">
          <ProForm submitter={false}>
            <Form.Item label="所属机构">
              <Input aria-label="所属机构" disabled value={currentOrgName} />
            </Form.Item>
            <Form.Item label="规则选择">
              <Select
                aria-label="规则选择"
                mode="multiple"
                showSearch
                optionFilterProp="label"
                value={draft.keyword_ids}
                options={keywordOptions}
                placeholder={keywords.length ? '先按分类选择要执行的规则' : '请先到规则管理新增规则'}
                onChange={(value) => setDraft((current) => ({
                  ...current,
                  keyword_ids: value.map((item) => Number(item))
                }))}
              />
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
            <Space>
              <Button type="primary" onClick={() => void submitTask()}>
                提交任务
              </Button>
              <Button href="/rules/keywords">去规则管理</Button>
            </Space>
          </ProForm>
        </SectionCard>

        <SectionCard title="执行提示" description="结合当前班次安排合理设置扫描范围。">
          <div className="task-form-layout__tips">
            <div>
              <Title level={5}>任务说明</Title>
              <Paragraph>任务提交后将按异步方式执行，不阻塞当前操作界面，适合批量巡检场景。</Paragraph>
            </div>
            <div>
              <Title level={5}>范围建议</Title>
              <Paragraph>优先从高风险规则组合开始巡检，再按发布时间逐步扩大范围，减少首轮噪音。</Paragraph>
            </div>
            <div>
              <Title level={5}>处理建议</Title>
              <Paragraph>任务提交后可直接在“任务结果”工作台查看命中文稿、批量处置与跟踪日志。</Paragraph>
            </div>
          </div>
        </SectionCard>
      </div>
    </>
  );
}
