import type { PropsWithChildren } from 'react';

export function ToolbarStrip({ children }: PropsWithChildren) {
  return <div className="toolbar-strip">{children}</div>;
}
