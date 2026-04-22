import type { PropsWithChildren, ReactNode } from 'react';

export interface SectionCardProps extends PropsWithChildren {
  title?: string;
  extra?: ReactNode;
}

export function SectionCard({ title, extra, children }: SectionCardProps) {
  return (
    <section className="section-card">
      {(title || extra) ? (
        <div className="section-card__header">
          {title ? <h3>{title}</h3> : <span />}
          {extra ? <div>{extra}</div> : null}
        </div>
      ) : null}
      <div className="section-card__body">{children}</div>
    </section>
  );
}
