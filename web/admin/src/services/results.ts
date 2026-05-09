import { apiRequest } from './request';

export interface ResultRecord {
  id: number;
  orgid: number;
  task_id: number;
  article_id: number;
  article_title: string;
  article_state?: number;
  risk_level: string;
  suggest_action: string;
  disposition_status: string;
  hit_count: number;
  latest_operator_name?: string;
  latest_action_at?: string;
  preview_field_name?: string;
  preview_keyword_text?: string;
  preview_matched_text?: string;
  preview_snippet?: string;
  extra_hit_count?: number;
  snippet?: string;
  matched_keyword?: string;
}

export interface ResultHitRecord {
  id: number;
  field_name: string;
  keyword_text: string;
  snippet: string;
  matched_text?: string;
  risk_level: string;
}

export interface ResultOperationRecord {
  id: number;
  operation_type: string;
  summary: string;
  operator_name?: string;
  created_at?: string;
}

export interface ResultFieldChangeRecord {
  id: number;
  field_name: string;
  old_value?: string;
  new_value?: string;
  before_value?: string;
  after_value?: string;
}

export interface ResultDetailRecord extends ResultRecord {
  article_state?: number;
  article_body?: string;
  hits: ResultHitRecord[];
  operation_logs: ResultOperationRecord[];
  field_changes: ResultFieldChangeRecord[];
}

export interface ResultListResult {
  page: number;
  pageSize: number;
  total: number;
  items: ResultRecord[];
}

export interface ResultListParams {
  page?: number;
  pageSize?: number;
  task_id?: number;
  article_id?: number;
  risk_level?: string;
  disposition_status?: string;
  keyword_text?: string;
  field_name?: string;
  title?: string;
}

export interface BatchOfflineInput {
  task_id?: number;
  result_ids?: number[];
  filter_snapshot?: Record<string, unknown>;
  reason?: string;
}

export interface BatchActionResult {
  action_id: number;
  target_count: number;
  success_count: number;
  fail_count: number;
  skip_count: number;
  status: string;
  action_type: string;
}

export interface RectifyArticleRecord {
  article_id: number;
  orgid: number;
  title: string;
  desc: string;
  body: string;
}

export interface RectifyArticleInput {
  title: string;
  desc: string;
  body: string;
  target_article_state?: number;
}

export async function listResults(params: ResultListParams): Promise<ResultListResult> {
  const query = new URLSearchParams();
  if (params.page) query.set('page', String(params.page));
  if (params.pageSize) query.set('page_size', String(params.pageSize));
  if (params.task_id) query.set('task_id', String(params.task_id));
  if (params.article_id) query.set('article_id', String(params.article_id));
  if (params.risk_level) query.set('risk_level', params.risk_level);
  if (params.disposition_status) query.set('disposition_status', params.disposition_status);
  if (params.keyword_text) query.set('keyword_text', params.keyword_text);
  if (params.field_name) query.set('field_name', params.field_name);
  if (params.title) query.set('title', params.title);

  const data = await apiRequest<{ page: number; page_size?: number; total: number; items: ResultRecord[] }>(
    `/api/v1/article-inspect/results?${query.toString()}`
  );

  return {
    page: data.page,
    pageSize: data.page_size ?? params.pageSize ?? 20,
    total: data.total,
    items: data.items ?? []
  };
}

export function getResultDetail(id: number): Promise<ResultDetailRecord> {
  return apiRequest<
    | ResultDetailRecord
    | {
        result: ResultRecord;
        hits?: ResultHitRecord[];
        operation_logs?: ResultOperationRecord[];
        field_change_logs?: ResultFieldChangeRecord[];
      }
  >(`/api/v1/article-inspect/results/${id}`).then((data) => {
    if ('result' in data) {
      return {
        ...data.result,
        hits: data.hits ?? [],
        operation_logs: data.operation_logs ?? [],
        field_changes: (data.field_change_logs ?? []).map((item) => ({
          ...item,
          old_value: item.old_value ?? item.before_value,
          new_value: item.new_value ?? item.after_value
        }))
      };
    }

    return {
      ...data,
      hits: data.hits ?? [],
      operation_logs: data.operation_logs ?? [],
      field_changes: (data.field_changes ?? []).map((item) => ({
        ...item,
        old_value: item.old_value ?? item.before_value,
        new_value: item.new_value ?? item.after_value
      }))
    };
  });
}

function postBatchAction(path: string, input: BatchOfflineInput): Promise<BatchActionResult> {
  return apiRequest<BatchActionResult>(path, {
    method: 'POST',
    body: JSON.stringify(input)
  });
}

export function batchOfflineResults(input: BatchOfflineInput): Promise<BatchActionResult> {
  return postBatchAction('/api/v1/article-inspect/actions/batch-offline', input);
}

export function batchIgnoreResults(input: BatchOfflineInput): Promise<BatchActionResult> {
  return postBatchAction('/api/v1/article-inspect/actions/batch-ignore', input);
}

export function batchProcessResults(input: BatchOfflineInput): Promise<BatchActionResult> {
  return postBatchAction('/api/v1/article-inspect/actions/batch-process', input);
}

export function getArticleRectify(articleId: number): Promise<RectifyArticleRecord> {
  return apiRequest<RectifyArticleRecord>(`/api/v1/article-inspect/articles/${articleId}/rectify`);
}

export function rectifyArticle(articleId: number, input: RectifyArticleInput): Promise<{ article_id: number; status: string }> {
  return apiRequest<{ article_id: number; status: string }>(`/api/v1/article-inspect/articles/${articleId}/rectify`, {
    method: 'PUT',
    body: JSON.stringify(input)
  });
}
