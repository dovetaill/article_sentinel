import { ProForm, ProFormText, ProFormTextArea } from '@ant-design/pro-components';
import { Button, Empty, Form, Space, Spin, Typography, message } from 'antd';
import { useEffect, useMemo, useState } from 'react';
import { useParams, useSearchParams } from 'react-router-dom';

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

const { Paragraph, Text, Title } = Typography;

type RectifyFormValues = {
  title: string;
  desc: string;
  body: string;
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

  const metrics = useMemo(() => {
    if (!record) {
      return [];
    }

    return [
      {
        label: '文章编号',
        value: `#${record.id}`,
        helper: '当前进入整改流程的稿件编号'
      },
      {
        label: '原标题字数',
        value: record.title.length,
        helper: '用于对照整改前标题长度'
      },
      {
        label: '原摘要字数',
        value: record.desc.length,
        helper: '便于保持摘要信息完整'
      },
      {
        label: '原文字数',
        value: record.body.length,
        helper: '按原始内容核验调整范围'
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
        <SectionCard title="整改载入中" description="正在读取稿件原文与当前整改版本。">
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
            <SectionCard title="整改内容" description="在保留事实准确性的前提下完成风险修订。">
              <div className="rectify-form">
                <Paragraph className="rectify-form__intro">
                  请在保持稿件主旨准确的前提下，对存在风险的标题、摘要与正文进行审慎修订，避免再次触发同类规则。
                </Paragraph>

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
                  <ProFormTextArea
                    name="body"
                    label="整改正文"
                    rules={[{ required: true, message: '请填写整改正文' }]}
                    fieldProps={{ 'aria-label': '整改正文', rows: 12 }}
                  />

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
              <SectionCard title="原稿对照" description="对照原始标题、摘要与正文，确保修改范围可控。">
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

              <SectionCard title="办理提示" description="按当前处置流程完成暂存或提交复核。">
                <div className="rectify-notes">
                  <div>
                    <Title level={5}>处置原则</Title>
                    <Paragraph>
                      修改内容应与原稿事实一致，重点消除敏感表述、误导措辞与不当扩散风险，不得新增未经核验的信息。
                    </Paragraph>
                  </div>
                  <div>
                    <Title level={5}>提交流程</Title>
                    <Paragraph>
                      “保存整改”适用于暂存调整结果；如需进入后续审核，请选择“保存并提交复核”。
                    </Paragraph>
                  </div>
                </div>
              </SectionCard>
            </div>
          </div>
        </>
      ) : null}

      {!loading && !record ? (
        <SectionCard title="未获取到整改信息" description="当前稿件未返回可编辑内容。">
          <Empty description="当前文章暂无可用整改内容，请返回文稿详情或任务结果重新进入。" />
        </SectionCard>
      ) : null}
    </>
  );
}
