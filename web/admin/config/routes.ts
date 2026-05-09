import { workspaceRouteItems } from '../src/components/PageTabs/route-meta';

const legacyRedirectPaths = [
  '/keywords',
  '/tasks',
  '/tasks/new',
  '/tasks/:taskId',
  '/tasks/:taskId/results',
  '/results',
  '/articles',
  '/articles/:articleId',
  '/articles/:articleId/rectify',
  '/logs'
];

export default [
  { path: '/user/login', layout: false, component: './User/LoginRedirect' },
  { path: '/', redirect: '/inspection/tasks' },
  { path: '/rules', redirect: '/rules/keywords' },
  ...legacyRedirectPaths.map((path) => ({
    path,
    layout: false,
    component: './LegacyRedirect'
  })),
  {
    path: '/',
    layout: false,
    component: '@/layouts/BasicLayout',
    routes: workspaceRouteItems
  },
  { path: '/*', component: '404', layout: false }
];
