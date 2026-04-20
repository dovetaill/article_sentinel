import { apiRequest } from '../lib/request';

export interface TaskRecord {
  id: number;
  orgid: number;
  task_no: string;
  status: string;
  total_scanned?: number;
  hit_articles?: number;
  hit_count?: number;
  creator_name?: string;
  created_at?: string;
  rule_snapshot?: string;
  request_snapshot?: string;
}

export interface TaskListResult {
  page: number;
  pageSize: number;
  total: number;
  items: TaskRecord[];
}

export interface TaskListParams {
  orgid: number;
  page?: number;
  pageSize?: number;
  status?: string;
  task_no?: string;
}

export interface CreateTaskInput {
  orgid: number;
  keyword_ids: number[];
  publish_time_start?: string;
  publish_time_end?: string;
  article_id?: number;
  title_like?: string;
  include_body: boolean;
  article_state?: number;
}

export async function listTasks(params: TaskListParams): Promise<TaskListResult> {
  const query = new URLSearchParams();
  query.set('orgid', String(params.orgid));
  if (params.page) query.set('page', String(params.page));
  if (params.pageSize) query.set('page_size', String(params.pageSize));
  if (params.status) query.set('status', params.status);
  if (params.task_no) query.set('task_no', params.task_no);

  const data = await apiRequest<{ page: number; page_size?: number; total: number; items: TaskRecord[] }>(
    `/api/v1/article-inspect/tasks?${query.toString()}`,
  );

  return {
    page: data.page,
    pageSize: data.page_size ?? params.pageSize ?? 20,
    total: data.total,
    items: data.items ?? []
  };
}

export function getTaskDetail(id: number, orgid = 100): Promise<TaskRecord> {
  return apiRequest<TaskRecord>(`/api/v1/article-inspect/tasks/${id}?orgid=${orgid}`);
}

export function createTask(input: CreateTaskInput): Promise<TaskRecord> {
  return apiRequest<TaskRecord>('/api/v1/article-inspect/tasks', {
    method: 'POST',
    body: JSON.stringify(input)
  });
}
