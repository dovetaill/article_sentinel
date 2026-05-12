import {
  PageContainer,
  ProForm,
  ProFormDateTimePicker,
  ProFormSelect,
  type ProFormInstance
} from '@ant-design/pro-components';
import { Button, Card, Typography, message } from 'antd';
import dayjs from 'dayjs';
import { useEffect, useMemo, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';

import { listKeywords, type KeywordRecord } from '@/services/keywords';
import { createTask } from '@/services/tasks';

const { Title, Paragraph } = Typography;
const timePayloadFormat = 'YYYY-MM-DDTHH:mm:ssZ';

type TaskCreateValues = {
  keyword_ids?: number[];
  publish_time_start?: dayjs.Dayjs;
  publish_time_end?: dayjs.Dayjs;
};

export default function TaskCreatePage() {
  const navigate = useNavigate();
  const formRef = useRef<ProFormInstance<TaskCreateValues>>();
  const redirectTimerRef = useRef<number | null>(null);
  const [messageApi, contextHolder] = message.useMessage();
  const [keywords, setKeywords] = useState<KeywordRecord[]>([]);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    let cancelled = false;

    void listKeywords({ page: 1, pageSize: 100, enabled: true })
      .then((result) => {
        if (!cancelled) {
          setKeywords((result.items ?? []).filter((item) => item.enabled));
        }
      })
      .catch(() => {
        if (!cancelled) {
          setKeywords([]);
        }
      });

    return () => {
      cancelled = true;
      if (redirectTimerRef.current !== null) {
        window.clearTimeout(redirectTimerRef.current);
      }
    };
  }, []);

  const keywordOptions = useMemo(
    () =>
      keywords.map((item) => ({
        value: item.id,
        label: `${item.category_name || '未分类规则'} / ${item.name}`
      })),
    [keywords]
  );

  return (
    <PageContainer title={false} pageHeaderRender={false}>
      {contextHolder}
      <div className="admin-domain-page">
        <div className="admin-domain-page__head">
          <div>
            <Title level={3} className="admin-domain-page__title">
              新建检测任务
            </Title>
            <Paragraph className="admin-domain-page__desc">
              选择当前机构已启用的规则集合，按发布时间范围发起一次新的异步扫描。
            </Paragraph>
          </div>
          <Button onClick={() => navigate('/inspection/tasks')}>返回任务列表</Button>
        </div>

        <Card className="admin-filter-card admin-surface-panel" variant="borderless">
          <ProForm<TaskCreateValues>
            formRef={formRef}
            submitter={{
              searchConfig: {
                submitText: '提交任务',
                resetText: '去规则管理'
              },
              resetButtonProps: {
                onClick: (event) => {
                  event.preventDefault();
                  navigate('/rules/keywords');
                }
              },
              submitButtonProps: {
                loading: submitting,
                disabled: submitting
              }
            }}
            onFinish={async (values) => {
              if (submitting) {
                return false;
              }

              const keywordIds = (values.keyword_ids ?? []).map((item) => Number(item)).filter(Boolean);
              if (keywordIds.length === 0) {
                messageApi.error('请先选择至少一条规则');
                return false;
              }

              setSubmitting(true);
              try {
                await createTask({
                  keyword_ids: keywordIds,
                  publish_time_start: values.publish_time_start?.format(timePayloadFormat),
                  publish_time_end: values.publish_time_end?.format(timePayloadFormat),
                  article_state: 9
                });
                messageApi.success('检测任务已提交');
                redirectTimerRef.current = window.setTimeout(() => {
                  navigate('/inspection/tasks');
                }, 1000);
                return true;
              } catch (error) {
                setSubmitting(false);
                messageApi.error(error instanceof Error ? error.message : '检测任务提交失败');
                return false;
              }
            }}
          >
            <ProFormSelect
              name="keyword_ids"
              label="规则选择"
              fieldProps={{
                mode: 'multiple',
                showSearch: true,
                optionFilterProp: 'label',
                options: keywordOptions,
                placeholder: keywords.length ? '先按分类选择要执行的规则' : '请先到规则管理新增规则',
                'aria-label': '规则选择'
              }}
            />
            <div className="admin-form-grid admin-form-grid--two-col">
              <ProFormDateTimePicker
                name="publish_time_start"
                label="发布时间起"
                fieldProps={{
                  inputReadOnly: true,
                  'aria-label': '发布时间起'
                }}
              />
              <ProFormDateTimePicker
                name="publish_time_end"
                label="发布时间止"
                fieldProps={{
                  inputReadOnly: true,
                  'aria-label': '发布时间止'
                }}
              />
            </div>
          </ProForm>
        </Card>
      </div>
    </PageContainer>
  );
}
