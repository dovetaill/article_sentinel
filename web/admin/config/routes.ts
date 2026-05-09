export default [
  { path: '/user/login', layout: false, component: './User/LoginRedirect' },
  { path: '/', redirect: '/inspection/tasks' },
  {
    path: '/',
    component: '@/layouts/BasicLayout',
    routes: [
      { path: '/inspection/tasks', name: '检测任务', component: './Inspection/TaskList' }
    ]
  },
  { path: '/*', component: '404', layout: false }
];
