import { describe, expect, it } from 'vitest';

import { summarizeArticles } from './articles';

describe('summarizeArticles', () => {
  it('collapses repeated result rows into one article summary', () => {
    const items = summarizeArticles([
      {
        id: 11,
        orgid: 100,
        task_id: 201,
        article_id: 501,
        article_title: 'Spam alert',
        risk_level: 'medium',
        suggest_action: 'offline',
        disposition_status: 'pending',
        hit_count: 2,
        latest_operator_name: '值班员甲',
        latest_action_at: '2026-04-20 10:00:00'
      },
      {
        id: 12,
        orgid: 100,
        task_id: 208,
        article_id: 501,
        article_title: 'Spam alert',
        risk_level: 'high',
        suggest_action: 'offline',
        disposition_status: 'processed',
        hit_count: 3,
        latest_operator_name: '值班员乙',
        latest_action_at: '2026-04-20 12:00:00'
      }
    ]);

    expect(items).toHaveLength(1);
    expect(items[0]).toMatchObject({
      article_id: 501,
      article_title: 'Spam alert',
      risk_level: 'high',
      hit_count: 5,
      latest_task_id: 208,
      latest_operator_name: '值班员乙'
    });
  });
});
