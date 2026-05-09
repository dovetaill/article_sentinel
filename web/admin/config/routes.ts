import { workspaceRouteItems } from '../src/components/PageTabs/route-meta';

export default [
  { path: '/user/login', layout: false, component: './User/LoginRedirect' },
  { path: '/', redirect: '/inspection/tasks' },
  {
    path: '/',
    layout: false,
    component: '@/layouts/BasicLayout',
    routes: workspaceRouteItems
  },
  { path: '/*', component: '404', layout: false }
];
