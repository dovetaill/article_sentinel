import { Empty } from 'antd';

function sanitizeSnapshotValue(value: unknown): unknown {
  if (Array.isArray(value)) {
    return value.map((item) => sanitizeSnapshotValue(item));
  }

  if (value && typeof value === 'object') {
    return Object.fromEntries(
      Object.entries(value)
        .filter(([key]) => key !== 'include_body')
        .map(([key, item]) => [key, sanitizeSnapshotValue(item)])
    );
  }

  return value;
}

function stripLegacySnapshotLines(snapshot: string) {
  return snapshot
    .split('\n')
    .filter((line) => !/"?include_body"?\s*:/.test(line))
    .join('\n')
    .trim();
}

export function formatSnapshotValue(snapshot: string | null | undefined, fallback: string) {
  if (!snapshot) {
    return fallback;
  }

  try {
    const sanitized = sanitizeSnapshotValue(JSON.parse(snapshot));
    return JSON.stringify(sanitized, null, 2);
  } catch {
    return stripLegacySnapshotLines(snapshot) || fallback;
  }
}

export interface SnapshotViewerProps {
  value?: string | null;
  emptyText?: string;
}

export default function SnapshotViewer({
  value,
  emptyText = '暂无快照。'
}: SnapshotViewerProps) {
  const formatted = formatSnapshotValue(value, emptyText);

  if (formatted === emptyText) {
    return <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={emptyText} />;
  }

  return <pre className="detail-code-block snapshot-viewer admin-surface-inline">{formatted}</pre>;
}
