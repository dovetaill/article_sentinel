import { ProForm } from '@ant-design/pro-components';
import { Button, DatePicker, Form, Input, Select, Space, message } from 'antd';
import dayjs, { type Dayjs } from 'dayjs';
import { useEffect, useRef, useState } from 'react';

import { PageHeader } from '../../components/ui/page-header';
import { SectionCard } from '../../components/ui/section-card';
import { useOrgContext } from '../../context/org-context';
import { listKeywords, type KeywordRecord } from '../../services/keywords';
import { createTask } from '../../services/tasks';
import { useWorkbenchNavigation } from '../../workbench/navigation';

const timeDisplayFormat = 'YYYY-MM-DD HH:mm:ss';
const timePayloadFormat = 'YYYY-MM-DDTHH:mm:ssZ';

type TaskDraft = {
  keyword_ids: number[];
  publish_time_start: Dayjs | null;
  publish_time_end: Dayjs | null;
};

export default function NewTaskPage() {
  const [messageApi, contextHolder] = message.useMessage();
  const [keywords, setKeywords] = useState<KeywordRecord[]>([]);
  const [submitting, setSubmitting] = useState(false);
  const [draft, setDraft] = useState<TaskDraft>({
    keyword_ids: [],
    publish_time_start: null,
    publish_time_end: null
  });
  const { activeOrgId, activeOrgName } = useOrgContext();
  const { open, onLinkClick } = useWorkbenchNavigation();
  const redirectTimerRef = useRef<number | null>(null);

  const currentOrgName = activeOrgName || '一县一端';

  useEffect(() => {
    void listKeywords({ page: 1, pageSize: 100, enabled: true })
      .then((result) => {
        setKeywords(result.items.filter((item) => item.enabled));
      })
      .catch(() => {
        setKeywords([]);
      });
  }, [activeOrgId]);

  useEffect(() => () => {
    if (redirectTimerRef.current !== null) {
      window.clearTimeout(redirectTimerRef.current);
    }
  }, []);

  const keywordOptions = keywords.map((item) => ({
    value: item.id,
    label: `${item.category_name || '未分类规则'} / ${item.name}`
  }));

  async function submitTask() {
    if (submitting) {
      return;
    }
    if (draft.keyword_ids.length === 0) {
      messageApi.error('请先选择至少一条规则');
      return;
    }

    setSubmitting(true);

    try {
      await createTask({
        keyword_ids: draft.keyword_ids,
        publish_time_start: draft.publish_time_start?.format(timePayloadFormat),
        publish_time_end: draft.publish_time_end?.format(timePayloadFormat),
        article_state: 9
      });
      messageApi.success('检测任务已提交');
      redirectTimerRef.current = window.setTimeout(() => {
        open('/tasks');
      }, 1000);
    } catch (error) {
      setSubmitting(false);
      messageApi.error(error instanceof Error ? error.message : '检测任务提交失败');
    }
  }

  return (
    <>
      {contextHolder}
      <PageHeader
        title="新建检测任务"
        extra={(
          <Button href="/tasks" onClick={(event) => onLinkClick(event, '/tasks')}>
            返回任务列表
          </Button>
        )}
      />

      <SectionCard title="任务配置">
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
            <DatePicker
              aria-label="发布时间起"
              showTime={{ format: 'HH:mm:ss' }}
              format={timeDisplayFormat}
              inputReadOnly
              style={{ width: '100%' }}
              value={draft.publish_time_start}
              onChange={(value) => setDraft((current) => ({ ...current, publish_time_start: value ? dayjs(value) : null }))}
            />
          </Form.Item>
          <Form.Item label="发布时间止">
            <DatePicker
              aria-label="发布时间止"
              showTime={{ format: 'HH:mm:ss' }}
              format={timeDisplayFormat}
              inputReadOnly
              style={{ width: '100%' }}
              value={draft.publish_time_end}
              onChange={(value) => setDraft((current) => ({ ...current, publish_time_end: value ? dayjs(value) : null }))}
            />
          </Form.Item>
          <Space>
            <Button type="primary" loading={submitting} disabled={submitting} onClick={() => void submitTask()}>
              提交任务
            </Button>
            <Button href="/rules/keywords" onClick={(event) => onLinkClick(event, '/rules/keywords')}>
              去规则管理
            </Button>
          </Space>
        </ProForm>
      </SectionCard>
    </>
  );
}
