import { apiRequest } from '../lib/request';

export interface OperationLogRecord {
  id: number;
  orgid: number;
  article_id?: number;
  task_id?: number;
  operation_type: string;
  before_state?: string;
  after_state?: string;
  summary: string;
  operator_name?: string;
  request_snapshot?: string;
  source_ip?: string;
  created_at?: string;
}

export interface OperationLogListResult {
  page: number;
  pageSize: number;
  total: number;
  items: OperationLogRecord[];
}

export interface OperationLogListParams {
  orgid: number;
  page?: number;
  pageSize?: number;
  article_id?: number;
  task_id?: number;
  operator_name?: string;
}

export interface FieldChangeLogRecord {
  id: number;
  orgid: number;
  article_id?: number;
  task_id?: number;
  result_id?: number;
  field_name: string;
  before_value?: string;
  after_value?: string;
  diff_summary?: string;
  operator_name?: string;
  created_at?: string;
}

export interface FieldChangeLogListResult {
  page: number;
  pageSize: number;
  total: number;
  items: FieldChangeLogRecord[];
}

export async function listOperationLogs(params: OperationLogListParams): Promise<OperationLogListResult> {
  const query = new URLSearchParams();
  query.set('orgid', String(params.orgid));
  if (params.page) query.set('page', String(params.page));
  if (params.pageSize) query.set('page_size', String(params.pageSize));
  if (params.article_id) query.set('article_id', String(params.article_id));
  if (params.task_id) query.set('task_id', String(params.task_id));
  if (params.operator_name) query.set('operator_name', params.operator_name);

  const data = await apiRequest<{ page: number; page_size?: number; total: number; items: OperationLogRecord[] }>(
    `/api/v1/article-inspect/logs/operations?${query.toString()}`,
  );

  return {
    page: data.page,
    pageSize: data.page_size ?? params.pageSize ?? 20,
    total: data.total,
    items: data.items ?? []
  };
}

export function listTaskOperationLogs(taskId: number, orgid = 100, page = 1, pageSize = 20): Promise<OperationLogListResult> {
  return listOperationLogs({
    orgid,
    task_id: taskId,
    page,
    pageSize
  });
}

export async function listArticleOperationLogs(articleId: number, orgid = 100, page = 1, pageSize = 20): Promise<OperationLogListResult> {
  const query = new URLSearchParams({
    orgid: String(orgid),
    page: String(page),
    page_size: String(pageSize)
  });

  const data = await apiRequest<{ page: number; page_size?: number; total: number; items: OperationLogRecord[] }>(
    `/api/v1/article-inspect/articles/${articleId}/operation-logs?${query.toString()}`,
  );

  return {
    page: data.page,
    pageSize: data.page_size ?? pageSize,
    total: data.total,
    items: data.items ?? []
  };
}

export async function listArticleFieldChanges(articleId: number, orgid = 100, page = 1, pageSize = 20): Promise<FieldChangeLogListResult> {
  const query = new URLSearchParams({
    orgid: String(orgid),
    page: String(page),
    page_size: String(pageSize)
  });

  const data = await apiRequest<{ page: number; page_size?: number; total: number; items: FieldChangeLogRecord[] }>(
    `/api/v1/article-inspect/articles/${articleId}/change-logs?${query.toString()}`,
  );

  return {
    page: data.page,
    pageSize: data.page_size ?? pageSize,
    total: data.total,
    items: data.items ?? []
  };
}
