import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes, useLocation } from 'react-router-dom';
import { describe, expect, it } from 'vitest';

import LegacyRedirect, { resolveLegacyPath } from './index';

function LocationProbe() {
  const location = useLocation();

  return <pre data-testid="location-probe">{`${location.pathname}${location.search}`}</pre>;
}

describe('LegacyRedirect', () => {
  it('maps /tasks to /inspection/tasks', () => {
    expect(resolveLegacyPath('/tasks')).toBe('/inspection/tasks');
  });

  it('maps legacy task result routes into task detail results', () => {
    expect(resolveLegacyPath('/tasks/77/results')).toBe('/inspection/tasks/77?tab=results');
  });

  it('maps result workspace links into task detail results when task context exists', () => {
    expect(resolveLegacyPath('/results?task_id=77&page=2')).toBe('/inspection/tasks/77?tab=results&page=2');
  });

  it('falls back to the task list when a result workspace link has no task context', () => {
    expect(resolveLegacyPath('/results')).toBe('/inspection/tasks');
  });

  it('preserves query filters when redirecting legacy log links', () => {
    expect(resolveLegacyPath('/logs?article_id=501&task_id=77&operator_name=alice&page=2')).toBe(
      '/audit/logs?article_id=501&task_id=77&operator_name=alice&page=2'
    );
  });

  it('maps legacy rectify routes before article detail fallback', () => {
    expect(resolveLegacyPath('/articles/501/rectify?task_id=77&result_id=91')).toBe(
      '/content/articles/501/rectify?task_id=77&result_id=91'
    );
  });

  it('navigates from a legacy article route to the new content route', async () => {
    render(
      <MemoryRouter initialEntries={['/articles/501']}>
        <Routes>
          <Route path="/articles/:articleId" element={<LegacyRedirect />} />
          <Route path="/content/articles/:articleId" element={<div>文稿详情探针</div>} />
        </Routes>
        <LocationProbe />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByTestId('location-probe')).toHaveTextContent('/content/articles/501');
    });
  });
});
