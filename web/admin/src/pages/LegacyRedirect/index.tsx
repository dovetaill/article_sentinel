import { useEffect } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';

function appendSearch(pathname: string, search: string) {
  return search ? `${pathname}${search}` : pathname;
}

function buildTaskResultPath(taskId: string, search: string) {
  const searchParams = new URLSearchParams(search);
  const nextSearchParams = new URLSearchParams();

  nextSearchParams.set('tab', 'results');

  for (const [key, value] of searchParams.entries()) {
    if (key === 'task_id' || key === 'tab') {
      continue;
    }

    nextSearchParams.append(key, value);
  }

  return `/inspection/tasks/${taskId}?${nextSearchParams.toString()}`;
}

export function resolveLegacyPath(href: string) {
  const url = new URL(href, 'http://legacy.local');
  const pathname = url.pathname;
  const search = url.search;

  const staticMap: Record<string, string> = {
    '/tasks': '/inspection/tasks',
    '/tasks/new': '/inspection/tasks/create',
    '/articles': '/content/articles',
    '/logs': '/audit/logs',
    '/keywords': '/rules/keywords'
  };

  const staticTarget = staticMap[pathname];
  if (staticTarget) {
    return appendSearch(staticTarget, search);
  }

  if (pathname === '/results' || pathname === '/inspection/results') {
    const taskId = url.searchParams.get('task_id');
    if (taskId) {
      return buildTaskResultPath(taskId, search);
    }

    return '/inspection/tasks';
  }

  const taskResultMatch = pathname.match(/^\/tasks\/(\d+)\/results$/);
  if (taskResultMatch) {
    return buildTaskResultPath(taskResultMatch[1], search);
  }

  const taskDetailMatch = pathname.match(/^\/tasks\/(\d+)$/);
  if (taskDetailMatch) {
    return appendSearch(`/inspection/tasks/${taskDetailMatch[1]}`, search);
  }

  const articleRectifyMatch = pathname.match(/^\/articles\/(\d+)\/rectify$/);
  if (articleRectifyMatch) {
    return appendSearch(`/content/articles/${articleRectifyMatch[1]}/rectify`, search);
  }

  const articleDetailMatch = pathname.match(/^\/articles\/(\d+)$/);
  if (articleDetailMatch) {
    return appendSearch(`/content/articles/${articleDetailMatch[1]}`, search);
  }

  return '/inspection/tasks';
}

export default function LegacyRedirect() {
  const navigate = useNavigate();
  const location = useLocation();

  useEffect(() => {
    const nextPath = resolveLegacyPath(`${location.pathname}${location.search}`);
    navigate(nextPath, { replace: true });
  }, [location.pathname, location.search, navigate]);

  return null;
}
