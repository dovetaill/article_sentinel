import type { RequestConfig, RunTimeLayoutConfig } from '@umijs/max';

import defaultSettings from '../config/defaultSettings';

export async function getInitialState() {
  return {
    settings: defaultSettings
  };
}

export const layout: RunTimeLayoutConfig = ({ initialState }) => {
  return {
    ...initialState?.settings,
    title: initialState?.settings?.title ?? '文章哨兵管理台'
  };
};

export const request: RequestConfig = {};
