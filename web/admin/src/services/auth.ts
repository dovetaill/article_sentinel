import { apiRequest } from '../lib/request';

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

export function getSession(): Promise<AuthSession> {
  return apiRequest<AuthSession>('/api/v1/auth/session');
}

export async function logout(): Promise<void> {
  await apiRequest<{ ok: boolean }>('/api/v1/auth/logout', {
    method: 'POST'
  });
}
