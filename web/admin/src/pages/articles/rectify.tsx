import { ProForm, ProFormText } from '@ant-design/pro-components';
import { Button, Empty, Form, Space, Spin, Typography, message } from 'antd';
import { useEffect, useMemo, useState } from 'react';
import { useParams, useSearchParams } from 'react-router-dom';

import HtmlBodyEditor from '../../components/ui/html-body-editor';
import { PageHeader } from '../../components/ui/page-header';
import { SectionCard } from '../../components/ui/section-card';
import { SummaryCard } from '../../components/ui/summary-card';
import { useOrgContext } from '../../context/org-context';
import {
  getArticleDetail,
  rectifyArticle,
  republishArticle,
  type ArticleDetailRecord
} from '../../services/articles';

const { Text } = Typography;

type RectifyFormValues = {
  title: string;
  desc: string;
  body: string;
};

type SummaryMetric = {
  label: string;
  value: string | number;
  helper?: string;
};

export default function RectifyPage() {
  const { articleId } = useParams();
  const [searchParams] = useSearchParams();
  const [form] = Form.useForm<RectifyFormValues>();
  const [messageApi, contextHolder] = message.useMessage();
  const [loading, setLoading] = useState(true);
  const [record, setRecord] = useState<ArticleDetailRecord | null>(null);
  const { activeOrgId } = useOrgContext();

  const currentOrgId = activeOrgId ?? 29;
  const taskIdFromSearch = Number(searchParams.get('task_id') || 0) || undefined;
  const resultIdFromSearch = Number(searchParams.get('result_id') || 0) || undefined;
  const returnTarget = searchParams.get('return_to') || `/articles/${articleId ?? ''}`;

  useEffect(() => {
    if (!articleId) {
      setLoading(false);
      return;
    }

    setLoading(true);
    void getArticleDetail(Number(articleId), currentOrgId)
      .then((data) => {
        setRecord(data);
        form.setFieldsValue({
          title: data.title,
          desc: data.desc,
          body: data.body
        });
      })
      .catch(() => {
        setRecord(null);
      })
      .finally(() => setLoading(false));
  }, [articleId, currentOrgId, form]);

  const metrics = useMemo<SummaryMetric[]>(() => {
    if (!record) {
      return [];
    }

    return [
      {
        label: '文章编号',
        value: `#${record.id}`
      },
      {
        label: '原标题字数',
        value: record.title.length
      },
      {
        label: '原摘要字数',
        value: record.desc.length
      },
      {
        label: '原文字数',
        value: record.body.length
      }
    ];
  }, [record]);

  async function submitRectification(targetArticleState?: number) {
    if (!articleId || !record) {
      return;
    }

    const values = await form.validateFields();
    const taskId = taskIdFromSearch ?? record.latest_task_id;
    const resultId = resultIdFromSearch ?? record.latest_result_id;

    await rectifyArticle(Number(articleId), {
      orgid: currentOrgId,
      task_id: taskId,
      result_id: resultId,
      title: values.title,
      short_title: record.short_title,
      rich_title: record.rich_title,
      keyword: record.keyword,
      desc: values.desc,
      body: values.body
    });

    if (targetArticleState === 1) {
      await republishArticle(Number(articleId), {
        orgid: currentOrgId,
        task_id: taskId,
        result_id: resultId
      });
    }

    messageApi.success(targetArticleState === 1 ? '整改内容已保存并提交复核' : '整改内容已保存');
  }

  return (
    <>
      {contextHolder}
      <PageHeader
        title="内容整改"
        extra={(
          <Space wrap>
            <Button href={returnTarget}>返回上一页</Button>
            <Text type="secondary">整改稿件：{articleId ? `#${articleId}` : '未识别'}</Text>
          </Space>
        )}
      />

      {loading ? (
        <SectionCard title="整改载入中">
          <div style={{ padding: '32px 0', textAlign: 'center' }}>
            <Spin />
          </div>
        </SectionCard>
      ) : null}

      {!loading && record ? (
        <>
          <div className="summary-card-grid">
            {metrics.map((item) => (
              <SummaryCard key={item.label} label={item.label} value={item.value} helper={item.helper} />
            ))}
          </div>

          <div className="rectify-layout">
            <SectionCard title="整改内容">
              <div className="rectify-form">
                <ProForm<RectifyFormValues> form={form} submitter={false}>
                  <ProFormText
                    name="title"
                    label="整改标题"
                    rules={[{ required: true, message: '请填写整改标题' }]}
                    fieldProps={{ 'aria-label': '整改标题' }}
                  />
                  <ProFormText
                    name="desc"
                    label="整改摘要"
                    rules={[{ required: true, message: '请填写整改摘要' }]}
                    fieldProps={{ 'aria-label': '整改摘要' }}
                  />
                  <Form.Item
                    name="body"
                    label="整改正文"
                    rules={[{ required: true, message: '请填写整改正文' }]}
                  >
                    <HtmlBodyEditor label="整改正文" />
                  </Form.Item>

                  <Space wrap className="rectify-form__actions">
                    <Button type="primary" onClick={() => void submitRectification()}>
                      保存整改
                    </Button>
                    <Button onClick={() => void submitRectification(1)}>
                      保存并提交复核
                    </Button>
                  </Space>
                </ProForm>
              </div>
            </SectionCard>

            <div className="rectify-layout__side">
              <SectionCard title="原稿对照">
                <div className="rectify-reference">
                  <div className="rectify-reference__item">
                    <span className="rectify-reference__label">原标题</span>
                    <p className="rectify-reference__value">{record.title}</p>
                  </div>
                  <div className="rectify-reference__item">
                    <span className="rectify-reference__label">原摘要</span>
                    <p className="rectify-reference__value">{record.desc}</p>
                  </div>
                  <div className="rectify-reference__item">
                    <span className="rectify-reference__label">原正文</span>
                    {record.body ? (
                      <div className="rectify-reference__body" dangerouslySetInnerHTML={{ __html: record.body }} />
                    ) : (
                      <p className="rectify-reference__value">暂无正文</p>
                    )}
                  </div>
                </div>
              </SectionCard>
            </div>
          </div>
        </>
      ) : null}

      {!loading && !record ? (
        <SectionCard title="未获取到整改信息">
          <Empty description="当前文章暂无可用整改内容，请返回文稿详情或任务结果重新进入。" />
        </SectionCard>
      ) : null}
    </>
  );
}
