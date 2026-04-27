import { apiRequest } from '../lib/request';

export interface OrgRecord {
  id: number;
  name: string;
  cate_id: number;
  enabled: boolean;
  sort: number;
}

export async function listOrgs(): Promise<OrgRecord[]> {
  const data = await apiRequest<{ items?: OrgRecord[] }>('/api/v1/article-inspect/orgs');
  return data.items ?? [];
}
