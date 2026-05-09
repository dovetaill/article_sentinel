import { defineConfig } from '@umijs/max';

import routes from './routes';

export default defineConfig({
  npmClient: 'npm',
  antd: {},
  access: {},
  layout: {},
  model: {},
  initialState: {},
  routes,
  hash: true
});
