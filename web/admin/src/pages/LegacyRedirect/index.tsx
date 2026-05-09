import { useEffect } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';

function appendSearch(pathname: string, search: string) {
  return search ? `${pathname}${search}` : pathname;
}

export function resolveLegacyPath(href: string) {
  const url = new URL(href, 'http://legacy.local');
  const pathname = url.pathname;
  const search = url.search;

  const staticMap: Record<string, string> = {
    '/tasks': '/inspection/tasks',
    '/tasks/new': '/inspection/tasks/create',
    '/results': '/inspection/results',
    '/articles': '/content/articles',
    '/logs': '/audit/logs',
    '/keywords': '/rules/keywords'
  };

  const staticTarget = staticMap[pathname];
  if (staticTarget) {
    return appendSearch(staticTarget, search);
  }

  const taskResultMatch = pathname.match(/^\/tasks\/(\d+)\/results$/);
  if (taskResultMatch) {
    const nextSearchParams = new URLSearchParams(search);
    nextSearchParams.set('task_id', taskResultMatch[1]);
    return `/inspection/results?${nextSearchParams.toString()}`;
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
