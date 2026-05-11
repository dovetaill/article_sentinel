import { PageContainer, ProForm, ProFormText } from '@ant-design/pro-components';
import { Button, Card, Empty, Form, Space, Spin, Typography, message } from 'antd';
import { useEffect, useMemo, useState } from 'react';
import { useLocation, useNavigate, useParams, useSearchParams } from 'react-router-dom';
import { useModel } from '@umijs/max';

import type { AppInitialState } from '@/app';
import HtmlArticleEditor from '@/components/HtmlArticleEditor';
import { resolveTabDescriptor } from '@/components/PageTabs/route-meta';
import {
  getArticleDetail,
  rectifyArticle,
  republishArticle,
  type ArticleDetailRecord,
  type ArticleRectifyInput
} from '@/services/articles';

const { Title, Paragraph, Text } = Typography;

type RectifyFormValues = {
  title: string;
  desc: string;
  body: string;
};

type SubmitMode = 'save' | 'review' | null;

type DraftPayload = RectifyFormValues & {
  updatedAt: number;
};

const DRAFT_STORAGE_PREFIX = 'article-sentinel:rectify-draft';

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

function buildDraftStorageKey(orgId: number, routeKey: string) {
  return `${DRAFT_STORAGE_PREFIX}:${orgId}:${routeKey}`;
}

function readDraft(orgId: number, routeKey: string) {
  if (typeof window === 'undefined' || !orgId || !routeKey) {
    return null;
  }

  try {
    const rawValue = window.sessionStorage.getItem(buildDraftStorageKey(orgId, routeKey));
    if (!rawValue) {
      return null;
    }

    const parsed = JSON.parse(rawValue) as Partial<DraftPayload>;
    if (typeof parsed.title !== 'string' || typeof parsed.desc !== 'string' || typeof parsed.body !== 'string') {
      return null;
    }

    return {
      title: parsed.title,
      desc: parsed.desc,
      body: parsed.body,
      updatedAt: typeof parsed.updatedAt === 'number' ? parsed.updatedAt : Date.now()
    } satisfies DraftPayload;
  } catch {
    return null;
  }
}

function writeDraft(orgId: number, routeKey: string, draft: RectifyFormValues) {
  if (typeof window === 'undefined' || !orgId || !routeKey) {
    return;
  }

  const payload: DraftPayload = {
    ...draft,
    updatedAt: Date.now()
  };

  try {
    window.sessionStorage.setItem(buildDraftStorageKey(orgId, routeKey), JSON.stringify(payload));
  } catch {
    // Ignore quota/storage errors; editing must remain usable.
  }
}

function clearDraft(orgId: number, routeKey: string) {
  if (typeof window === 'undefined' || !orgId || !routeKey) {
    return;
  }

  try {
    window.sessionStorage.removeItem(buildDraftStorageKey(orgId, routeKey));
  } catch {
    // Best-effort cleanup only.
  }
}

export default function ArticleRectifyPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const { articleId } = useParams();
  const [searchParams] = useSearchParams();
  const [form] = Form.useForm<RectifyFormValues>();
  const [messageApi, contextHolder] = message.useMessage();
  const [loading, setLoading] = useState(true);
  const [submitMode, setSubmitMode] = useState<SubmitMode>(null);
  const [record, setRecord] = useState<ArticleDetailRecord | null>(null);
  const [draftRestored, setDraftRestored] = useState(false);
  const { initialState } = useModel('@@initialState') as { initialState?: AppInitialState };

  const currentOrgId = initialState?.currentOrgId ?? 0;
  const numericArticleId = Number(articleId || 0) || 0;
  const currentHref = `${location.pathname}${location.search}`;
  const routeKey = useMemo(() => resolveTabDescriptor(currentHref).key, [currentHref]);
  const taskIdFromSearch = Number(searchParams.get('task_id') || 0) || undefined;
  const resultIdFromSearch = Number(searchParams.get('result_id') || 0) || undefined;
  const returnTarget = searchParams.get('return_to') || `/content/articles/${articleId ?? ''}`;

  useEffect(() => {
    if (!numericArticleId) {
      setLoading(false);
      setRecord(null);
      return;
    }

    let cancelled = false;
    setLoading(true);

    void getArticleDetail(numericArticleId)
      .then((articleDetail) => {
        if (cancelled) {
          return;
        }

        const draft = readDraft(currentOrgId, routeKey);
        setRecord(articleDetail);
        setDraftRestored(Boolean(draft));
        form.setFieldsValue({
          title: draft?.title ?? articleDetail.title ?? '',
          desc: draft?.desc ?? articleDetail.desc ?? '',
          body: draft?.body ?? articleDetail.body ?? ''
        });
      })
      .catch(() => {
        if (cancelled) {
          return;
        }

        setRecord(null);
        setDraftRestored(false);
      })
      .finally(() => {
        if (!cancelled) {
          setLoading(false);
        }
      });

    return () => {
      cancelled = true;
    };
  }, [currentOrgId, form, numericArticleId, routeKey]);

  const summaryItems = useMemo(() => {
    if (!record) {
      return [];
    }

    return [
      { label: '文章编号', value: `#${record.id}` },
      { label: '当前状态', value: renderArticleState(record.state) },
      { label: '来源任务', value: `#${taskIdFromSearch ?? record.latest_task_id ?? '-'}` },
      { label: '来源结果', value: `#${resultIdFromSearch ?? record.latest_result_id ?? '-'}` }
    ];
  }, [record, resultIdFromSearch, taskIdFromSearch]);

  async function submitRectification(mode: Exclude<SubmitMode, null>) {
    if (!record || !numericArticleId) {
      return;
    }

    setSubmitMode(mode);

    try {
      const values = await form.validateFields();
      const requestPayload: ArticleRectifyInput = {
        task_id: taskIdFromSearch ?? record.latest_task_id,
        result_id: resultIdFromSearch ?? record.latest_result_id,
        title: values.title,
        short_title: record.short_title,
        rich_title: record.rich_title,
        keyword: record.keyword,
        desc: values.desc,
        body: values.body
      };

      await rectifyArticle(numericArticleId, requestPayload);

      if (mode === 'review') {
        await republishArticle(numericArticleId, {
          task_id: requestPayload.task_id,
          result_id: requestPayload.result_id
        });
      }

      clearDraft(currentOrgId, routeKey);
      setDraftRestored(false);
      messageApi.success(mode === 'review' ? '整改内容已保存并提交复核' : '整改内容已保存');
    } catch (error) {
      messageApi.error(error instanceof Error ? error.message : '整改提交失败');
    } finally {
      setSubmitMode(null);
    }
  }

  return (
    <PageContainer title={false}>
      {contextHolder}
      <div className="admin-domain-page">
        <div className="admin-domain-page__head">
          <div>
            <Title level={3} className="admin-domain-page__title">
              内容整改
            </Title>
            <Paragraph className="admin-domain-page__desc">
              对照原稿完成标题、摘要与正文修订，并保留返回来源页的工作路径。
            </Paragraph>
          </div>
          <Space wrap>
            <Button onClick={() => navigate(returnTarget)}>返回上一页</Button>
            <Text type="secondary">整改稿件：{numericArticleId ? `#${numericArticleId}` : '未识别'}</Text>
          </Space>
        </div>

        {loading ? (
          <Card className="admin-filter-card admin-surface-panel" variant="borderless">
            <div style={{ padding: '32px 0', textAlign: 'center' }}>
              <Spin />
            </div>
          </Card>
        ) : null}

        {!loading && !record ? (
          <Card className="admin-filter-card admin-surface-panel" variant="borderless">
            <Empty description="当前文章暂无可用整改内容，请返回文稿详情或风险结果重新进入。" />
          </Card>
        ) : null}

        {!loading && record ? (
          <>
            <Space size={16} wrap className="admin-summary-strip">
              {summaryItems.map((item) => (
                <Card key={item.label} className="admin-summary-card admin-surface-panel" variant="borderless">
                  <div className="admin-stat-card">
                    <div className="admin-stat-card__label">{item.label}</div>
                    <div className="admin-stat-card__value">{item.value}</div>
                  </div>
                </Card>
              ))}
            </Space>

            <div className="rectify-layout">
              <Card className="admin-filter-card rectify-layout__main admin-surface-panel" variant="borderless">
                <Space direction="vertical" size={16} style={{ width: '100%' }}>
                  <div>
                    <Title level={4} style={{ margin: 0 }}>
                      整改内容
                    </Title>
                    {draftRestored ? (
                      <Text type="secondary">已恢复本次会话暂存内容，可继续编辑后再保存。</Text>
                    ) : null}
                  </div>

                  <ProForm<RectifyFormValues>
                    form={form}
                    submitter={false}
                    onValuesChange={(_, values) => {
                      writeDraft(currentOrgId, routeKey, {
                        title: values.title ?? '',
                        desc: values.desc ?? '',
                        body: values.body ?? ''
                      });
                    }}
                  >
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
                      <HtmlArticleEditor label="整改正文" placeholder="请填写整改正文" />
                    </Form.Item>

                    <Space wrap className="rectify-form__actions">
                      <Button
                        onClick={() => {
                          void submitRectification('save');
                        }}
                        loading={submitMode === 'save'}
                      >
                        保存整改
                      </Button>
                      <Button
                        type="primary"
                        onClick={() => {
                          void submitRectification('review');
                        }}
                        loading={submitMode === 'review'}
                      >
                        保存并提交复核
                      </Button>
                    </Space>
                  </ProForm>
                </Space>
              </Card>

              <Card className="admin-filter-card rectify-layout__side admin-surface-panel" variant="borderless">
                <Space direction="vertical" size={16} style={{ width: '100%' }}>
                  <Title level={4} style={{ margin: 0 }}>
                    原稿对照
                  </Title>

                  <div className="rectify-reference">
                    <div className="rectify-reference__item">
                      <span className="rectify-reference__label">原标题</span>
                      <p className="rectify-reference__value">{record.title}</p>
                    </div>
                    <div className="rectify-reference__item">
                      <span className="rectify-reference__label">原摘要</span>
                      <p className="rectify-reference__value">{record.desc || '暂无摘要'}</p>
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
                </Space>
              </Card>
            </div>
          </>
        ) : null}
      </div>
    </PageContainer>
  );
}
