import { apiRequest } from '../lib/request';

export interface KeywordRecord {
  id: number;
  orgid: number;
  name: string;
  category_id: number;
  category_name: string;
  match_type: string;
  risk_level: string;
  suggest_action: string;
  enabled: boolean;
  remark?: string;
  scopes: string[];
}

export interface KeywordListResult {
  page: number;
  pageSize: number;
  total: number;
  items: KeywordRecord[];
}

export interface KeywordListParams {
  page?: number;
  pageSize?: number;
  categoryId?: number;
  keyword?: string;
  enabled?: boolean;
}

export interface KeywordMutationInput {
  name: string;
  category_id: number;
  match_type: string;
  risk_level: string;
  suggest_action: string;
  enabled: boolean;
  remark?: string;
  scopes: string[];
}

export async function listKeywords(params: KeywordListParams): Promise<KeywordListResult> {
  const query = new URLSearchParams();
  if (params.page) query.set('page', String(params.page));
  if (params.pageSize) query.set('page_size', String(params.pageSize));
  if (params.categoryId) query.set('category_id', String(params.categoryId));
  if (params.keyword) query.set('keyword', params.keyword);
  if (typeof params.enabled === 'boolean') query.set('enabled', String(params.enabled));

  const data = await apiRequest<{ page: number; page_size?: number; total: number; items: KeywordRecord[] }>(
    `/api/v1/article-inspect/keywords?${query.toString()}`,
  );

  return {
    page: data.page,
    pageSize: data.page_size ?? params.pageSize ?? 20,
    total: data.total,
    items: data.items ?? []
  };
}

export function createKeyword(input: KeywordMutationInput): Promise<KeywordRecord> {
  return apiRequest<KeywordRecord>('/api/v1/article-inspect/keywords', {
    method: 'POST',
    body: JSON.stringify(input)
  });
}

export function updateKeyword(id: number, input: KeywordMutationInput): Promise<KeywordRecord> {
  return apiRequest<KeywordRecord>(`/api/v1/article-inspect/keywords/${id}`, {
    method: 'PUT',
    body: JSON.stringify(input)
  });
}

export function patchKeywordStatus(id: number, enabled: boolean): Promise<KeywordRecord> {
  return apiRequest<KeywordRecord>(`/api/v1/article-inspect/keywords/${id}/status`, {
    method: 'PATCH',
    body: JSON.stringify({ enabled })
  });
}

export function deleteKeyword(id: number): Promise<{ id: number }> {
  return apiRequest<{ id: number }>(`/api/v1/article-inspect/keywords/${id}`, {
    method: 'DELETE'
  });
}
