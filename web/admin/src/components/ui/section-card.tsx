import type { PropsWithChildren, ReactNode } from 'react';

export interface SectionCardProps extends PropsWithChildren {
  title?: string;
  description?: string;
  extra?: ReactNode;
}

export function SectionCard({ title, description, extra, children }: SectionCardProps) {
  return (
    <section className="section-card">
      {(title || extra) ? (
        <div className="section-card__header">
          <div className="section-card__heading">
            {title ? <h3>{title}</h3> : <span />}
            {description ? <p className="section-card__description">{description}</p> : null}
          </div>
          {extra ? <div>{extra}</div> : null}
        </div>
      ) : null}
      <div className="section-card__body">{children}</div>
    </section>
  );
}
