import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import SnapshotViewer from './index';

describe('SnapshotViewer', () => {
  it('shows a fallback when the snapshot is empty', () => {
    render(<SnapshotViewer value={null} emptyText="暂无请求快照。" />);

    expect(screen.getByText('暂无请求快照。')).toBeInTheDocument();
  });
});
