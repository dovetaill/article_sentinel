function escapePattern(value: string) {
  return value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}

function highlightText(text: string, marker?: string) {
  if (!marker) {
    return text;
  }

  const parts = text.split(new RegExp(`(${escapePattern(marker)})`, 'gi'));
  return parts.map((part, index) => (
    part.toLowerCase() === marker.toLowerCase()
      ? <mark key={`${marker}-${index}`}>{part}</mark>
      : <span key={`${marker}-${index}`}>{part}</span>
  ));
}

function fieldLabel(fieldName?: string) {
  switch (fieldName) {
    case 'title':
      return '标题';
    case 'body':
      return '正文';
    case 'keyword':
      return '关键词';
    case 'rich_title':
      return '富标题';
    case 'short_title':
      return '副标题';
    case 'desc':
      return '摘要';
    default:
      return fieldName || '未知字段';
  }
}

export interface HitPreviewProps {
  fieldName?: string;
  keywordText?: string;
  matchedText?: string;
  snippet?: string;
  extraHitCount?: number;
}

export function HitPreview({ fieldName, keywordText, matchedText, snippet, extraHitCount = 0 }: HitPreviewProps) {
  const content = snippet?.trim() || '-';

  return (
    <div className="hit-preview">
      <div className="hit-preview__meta">
        <span className="status-badge status-badge--neutral">{fieldLabel(fieldName)}</span>
        {keywordText ? <span className="status-badge status-badge--warning">{keywordText}</span> : null}
        {extraHitCount > 0 ? <span className="hit-preview__extra">另有 {extraHitCount} 条命中</span> : null}
      </div>
      <div className="hit-preview__snippet" title={content}>
        {content === '-' ? content : highlightText(content, matchedText)}
      </div>
    </div>
  );
}
