import { apiRequest } from '../lib/request';

export interface CategoryRecord {
  id: number;
  orgid: number;
  name: string;
  code: string;
  enabled: boolean;
  sort: number;
  creator_id?: number;
  creator_name?: string;
  updater_id?: number;
  updater_name?: string;
  create_at?: string;
  update_at?: string;
}

export interface CategoryListResult {
  page: number;
  pageSize: number;
  total: number;
  items: CategoryRecord[];
}

export interface CategoryListParams {
  orgid: number;
  page?: number;
  pageSize?: number;
  name?: string;
  enabled?: boolean;
}

export interface CategoryMutationInput {
  orgid: number;
  name: string;
  code: string;
  enabled: boolean;
  sort?: number;
}

export async function listCategories(params: CategoryListParams): Promise<CategoryListResult> {
  const query = new URLSearchParams();
  query.set('orgid', String(params.orgid));
  if (params.page) query.set('page', String(params.page));
  if (params.pageSize) query.set('page_size', String(params.pageSize));
  if (params.name) query.set('name', params.name);
  if (typeof params.enabled === 'boolean') query.set('enabled', String(params.enabled));

  const data = await apiRequest<{ page: number; page_size?: number; total: number; items: CategoryRecord[] }>(
    `/api/v1/article-inspect/categories?${query.toString()}`,
  );

  return {
    page: data.page,
    pageSize: data.page_size ?? params.pageSize ?? 20,
    total: data.total,
    items: data.items ?? []
  };
}

export async function listEnabledCategories(orgid: number): Promise<CategoryRecord[]> {
  const result = await listCategories({
    orgid,
    page: 1,
    pageSize: 200,
    enabled: true
  });

  return result.items.filter((item) => item.enabled);
}

export function createCategory(input: CategoryMutationInput): Promise<CategoryRecord> {
  return apiRequest<CategoryRecord>('/api/v1/article-inspect/categories', {
    method: 'POST',
    body: JSON.stringify(input)
  });
}

export function updateCategory(id: number, input: CategoryMutationInput): Promise<CategoryRecord> {
  return apiRequest<CategoryRecord>(`/api/v1/article-inspect/categories/${id}`, {
    method: 'PUT',
    body: JSON.stringify(input)
  });
}

export function patchCategoryStatus(id: number, orgid: number, enabled: boolean): Promise<CategoryRecord> {
  return apiRequest<CategoryRecord>(`/api/v1/article-inspect/categories/${id}/status`, {
    method: 'PATCH',
    body: JSON.stringify({ orgid, enabled })
  });
}

export function deleteCategory(id: number, orgid: number): Promise<{ id: number }> {
  return apiRequest<{ id: number }>(`/api/v1/article-inspect/categories/${id}?orgid=${orgid}`, {
    method: 'DELETE'
  });
}
