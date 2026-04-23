import { listResults, type ResultListParams, type ResultListResult, type ResultRecord } from './results';

export interface ArticleListItem {
  article_id: number;
  article_title: string;
  article_state?: number;
  risk_level: string;
  disposition_status: string;
  hit_count: number;
  latest_task_id: number;
  latest_operator_name?: string;
  latest_action_at?: string;
}

export interface ArticleListResult {
  page: number;
  pageSize: number;
  total: number;
  items: ArticleListItem[];
}

const riskRank: Record<string, number> = {
  low: 1,
  medium: 2,
  high: 3
};

function pickHigherRisk(current: string, next: string) {
  return (riskRank[next] ?? 0) > (riskRank[current] ?? 0) ? next : current;
}

function pickLatestRecord(current: ArticleListItem, next: ResultRecord) {
  const currentStamp = current.latest_action_at ?? '';
  const nextStamp = next.latest_action_at ?? '';

  if (nextStamp > currentStamp) {
    return {
      latest_task_id: next.task_id,
      latest_operator_name: next.latest_operator_name,
      latest_action_at: next.latest_action_at
    };
  }

  if (!currentStamp && next.task_id > current.latest_task_id) {
    return {
      latest_task_id: next.task_id,
      latest_operator_name: next.latest_operator_name,
      latest_action_at: next.latest_action_at
    };
  }

  return {
    latest_task_id: current.latest_task_id,
    latest_operator_name: current.latest_operator_name,
    latest_action_at: current.latest_action_at
  };
}

export function summarizeArticles(items: ResultRecord[]): ArticleListItem[] {
  const grouped = new Map<number, ArticleListItem>();

  items.forEach((item) => {
    const current = grouped.get(item.article_id);

    if (!current) {
      grouped.set(item.article_id, {
        article_id: item.article_id,
        article_title: item.article_title,
        article_state: item.article_state,
        risk_level: item.risk_level,
        disposition_status: item.disposition_status,
        hit_count: item.hit_count,
        latest_task_id: item.task_id,
        latest_operator_name: item.latest_operator_name,
        latest_action_at: item.latest_action_at
      });
      return;
    }

    const latest = pickLatestRecord(current, item);

    grouped.set(item.article_id, {
      ...current,
      article_state: item.article_state ?? current.article_state,
      risk_level: pickHigherRisk(current.risk_level, item.risk_level),
      disposition_status: current.disposition_status === 'pending' ? current.disposition_status : item.disposition_status,
      hit_count: current.hit_count + item.hit_count,
      ...latest
    });
  });

  return Array.from(grouped.values());
}

export async function listArticles(params: ResultListParams): Promise<ArticleListResult> {
  const data: ResultListResult = await listResults(params);

  return {
    page: data.page,
    pageSize: data.pageSize,
    total: summarizeArticles(data.items).length,
    items: summarizeArticles(data.items)
  };
}
