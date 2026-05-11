import type { ReactNode } from 'react';

import { App as AntdApp, ConfigProvider } from 'antd';
import type { RequestConfig, RunTimeLayoutConfig } from '@umijs/max';
import { history } from '@umijs/max';

import defaultSettings from '../config/defaultSettings';
import { getSession, type AuthSession } from './services/auth';
import { ApiRequestError } from './services/request';
import { adminAntdTheme } from './styles/admin-theme';
import { redirectToFixedLogin } from './utils/auth';

export interface AppInitialState {
  settings: typeof defaultSettings;
  currentUser: AuthSession | null;
  currentOrgId: number | null;
  currentOrgName: string | null;
  fetchSession: () => Promise<AuthSession | null>;
}

async function fetchSession() {
  try {
    return await getSession();
  } catch (error) {
    if (error instanceof ApiRequestError && error.status === 401) {
      return null;
    }

    return null;
  }
}

export async function getInitialState() {
  const currentUser = await fetchSession();

  return {
    settings: defaultSettings,
    currentUser,
    currentOrgId: currentUser?.orgid ?? null,
    currentOrgName: currentUser?.orgname ?? null,
    fetchSession
  } satisfies AppInitialState;
}

export function rootContainer(container: ReactNode) {
  return (
    <ConfigProvider theme={adminAntdTheme}>
      <AntdApp>{container}</AntdApp>
    </ConfigProvider>
  );
}

export const layout: RunTimeLayoutConfig = ({ initialState }) => {
  return {
    ...initialState?.settings,
    title: initialState?.settings?.title ?? '文章哨兵管理台',
    onPageChange: () => {
      if (!initialState?.currentUser && history.location.pathname !== '/user/login') {
        history.push('/user/login');
      }
    }
  };
};

export const request: RequestConfig = {
  timeout: 10000,
  withCredentials: true,
  headers: {
    Accept: 'application/json',
    'Content-Type': 'application/json'
  },
  errorConfig: {
    errorHandler(error) {
      const status = (error as { response?: { status?: number } })?.response?.status;

      if (status === 401) {
        redirectToFixedLogin();
      }

      throw error;
    }
  }
};
