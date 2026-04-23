import type { ReactNode } from 'react';

export interface PageHeaderProps {
  title: string;
  description: string;
  extra?: ReactNode;
}

export function PageHeader({ title, description, extra }: PageHeaderProps) {
  return (
    <div className="page-header">
      <div>
        <h2>{title}</h2>
        <p className="page-header__description">{description}</p>
      </div>
      {extra ? <div className="page-header__extra">{extra}</div> : null}
    </div>
  );
}
