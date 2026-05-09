import { apiRequest } from './request';

export interface AuthSession {
  id: number;
  orgid: number;
  orgname: string;
  platform: string;
  priv: string;
  roleid: string;
  nickname: string;
  avatar?: string;
  departmentid?: number;
  is_open_edu?: boolean;
}

export function getSession() {
  return apiRequest<AuthSession>('/api/v1/auth/session');
}

export async function logout() {
  await apiRequest<{ ok: boolean }>('/api/v1/auth/logout', {
    method: 'POST'
  });
}
