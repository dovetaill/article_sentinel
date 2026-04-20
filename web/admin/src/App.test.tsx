import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it } from 'vitest';

import App from './App';
import { appRoutes } from './routes';

describe('App shell', () => {
  it('bootstraps the admin shell heading', () => {
    render(
      <MemoryRouter initialEntries={['/']}>
        <App />
      </MemoryRouter>,
    );

    expect(screen.getByRole('heading', { name: /article sentinel/i })).toBeInTheDocument();
    expect(screen.getByText(/inspection console/i)).toBeInTheDocument();
  });

  it('renders the primary navigation routes', () => {
    expect(appRoutes.map((route) => route.label)).toEqual(['Keywords', 'Tasks', 'Results', 'Logs']);
  });
});
