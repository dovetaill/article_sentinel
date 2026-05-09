import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import StatusTag from './index';

describe('StatusTag', () => {
  it('renders task status labels in Chinese', () => {
    render(<StatusTag kind="task" value="running" />);

    expect(screen.getByText('执行中')).toBeInTheDocument();
  });
});
