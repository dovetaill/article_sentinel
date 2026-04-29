import { apiRequest } from '../lib/request';

export interface ArticleListParams {
  page?: number;
  pageSize?: number;
  state?: number;
  article_id?: number;
  title?: string;
}

export interface ArticleListItem {
  id: number;
  orgid: number;
  title: string;
  thumbnail?: string;
  state: number;
  publish_at_time?: string;
  latest_risk_level?: string;
  latest_task_id?: number;
  latest_result_id?: number;
  latest_suggest_action?: string;
  latest_disposition_status?: string;
  latest_operator_name?: string;
  latest_action_at?: string;
}

export interface ArticleListResult {
  page: number;
  pageSize: number;
  total: number;
  items: ArticleListItem[];
}

export interface ArticleDetailRecord extends ArticleListItem {
  short_title: string;
  rich_title: string;
  keyword: string;
  desc: string;
  body: string;
}

export interface ArticleLifecycleInput {
  task_id?: number;
  result_id?: number;
  action_id?: string;
  reason?: string;
}

export interface ArticleRectifyInput extends ArticleLifecycleInput {
  title: string;
  short_title?: string;
  rich_title?: string;
  keyword?: string;
  desc: string;
  body: string;
}

export interface ArticleLifecycleResult {
  article_id: number;
  status: string;
  before_state?: number;
  after_state?: number;
}

export async function listArticles(params: ArticleListParams): Promise<ArticleListResult> {
  const query = new URLSearchParams();
  if (params.page) query.set('page', String(params.page));
  if (params.pageSize) query.set('page_size', String(params.pageSize));
  if (params.state !== undefined) query.set('state', String(params.state));
  if (params.article_id) query.set('article_id', String(params.article_id));
  if (params.title) query.set('title', params.title);

  const data = await apiRequest<{ page: number; page_size?: number; total: number; items: ArticleListItem[] }>(
    `/api/v1/article-inspect/articles?${query.toString()}`,
  );

  return {
    page: data.page,
    pageSize: data.page_size ?? params.pageSize ?? 20,
    total: data.total,
    items: data.items ?? []
  };
}

export function getArticleDetail(articleId: number): Promise<ArticleDetailRecord> {
  return apiRequest<ArticleDetailRecord>(`/api/v1/article-inspect/articles/${articleId}`);
}

function postArticleAction(path: string, input: ArticleLifecycleInput): Promise<ArticleLifecycleResult> {
  return apiRequest<ArticleLifecycleResult>(path, {
    method: 'POST',
    body: JSON.stringify(input)
  });
}

export function offlineArticle(articleId: number, input: ArticleLifecycleInput): Promise<ArticleLifecycleResult> {
  return postArticleAction(`/api/v1/article-inspect/articles/${articleId}/offline`, input);
}

export function republishArticle(articleId: number, input: ArticleLifecycleInput): Promise<ArticleLifecycleResult> {
  return postArticleAction(`/api/v1/article-inspect/articles/${articleId}/republish`, input);
}

export function rectifyArticle(articleId: number, input: ArticleRectifyInput) {
  return apiRequest<Array<{
    field_name: string;
    before_value: string;
    after_value: string;
    diff_summary: string;
  }>>(`/api/v1/article-inspect/articles/${articleId}/rectify`, {
    method: 'PUT',
    body: JSON.stringify(input)
  });
}
