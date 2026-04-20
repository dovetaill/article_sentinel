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
