import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import HitPreview from './index';

describe('HitPreview', () => {
  it('renders the hit preview inside a light inline surface', () => {
    const { container } = render(
      <HitPreview
        fieldName="body"
        keywordText="北京市"
        matchedText="北京市"
        snippet="正文中提到北京市相关风险线索。"
        extraHitCount={2}
      />
    );

    expect(container.querySelector('.hit-preview.admin-surface-inline')).toBeInTheDocument();
    expect(screen.getByText('正文')).toBeInTheDocument();
    expect(container.querySelector('.hit-preview__keyword-tag')).toHaveTextContent('北京市');
    expect(screen.getByText('另有 2 条命中')).toBeInTheDocument();
    expect(container.querySelector('.hit-preview__snippet mark')).toHaveTextContent('北京市');
  });
});
