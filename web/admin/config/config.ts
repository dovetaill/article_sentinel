import { defineConfig } from '@umijs/max';

import proxy from './proxy';
import routes from './routes';

const proxyEnv = (process.env.UMI_ENV || process.env.NODE_ENV || 'dev') as keyof typeof proxy;

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
  esbuildMinifyIIFE: true,
  targets: {
    chrome: 88,
    edge: 88
  },
  proxy: proxy[proxyEnv] ?? proxy.dev
});
