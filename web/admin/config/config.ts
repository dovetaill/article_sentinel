import { defineConfig } from '@umijs/max';

import proxy from './proxy';
import routes from './routes';

export default defineConfig({
  npmClient: 'npm',
  antd: {
    style: 'less',
    styleProvider: {
      hashPriority: 'high',
      legacyTransformer: true
    }
  },
  access: {},
  layout: {},
  model: {},
  initialState: {},
  request: {},
  routes,
  hash: true,
  targets: {
    chrome: 88,
    edge: 88
  },
  proxy: proxy[process.env.NODE_ENV || 'dev'] ?? proxy.dev
});
