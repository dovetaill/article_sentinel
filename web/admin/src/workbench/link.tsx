import { forwardRef, type AnchorHTMLAttributes, type MouseEvent as ReactMouseEvent } from 'react';

import { buildWorkbenchHref, type BuildWorkbenchHrefOptions, useWorkbenchNavigation } from './navigation';

type WorkbenchLinkProps = Omit<AnchorHTMLAttributes<HTMLAnchorElement>, 'href'> & {
  to: string;
  options?: BuildWorkbenchHrefOptions;
};

export const WorkbenchLink = forwardRef<HTMLAnchorElement, WorkbenchLinkProps>(function WorkbenchLink(
  { to, options, onClick, ...rest },
  ref,
) {
  const { onLinkClick } = useWorkbenchNavigation();
  const href = buildWorkbenchHref(to, options);

  return (
    <a
      {...rest}
      ref={ref}
      href={href}
      onClick={(event: ReactMouseEvent<HTMLAnchorElement>) => {
        onClick?.(event);

        if (event.defaultPrevented) {
          return;
        }

        onLinkClick(event, href);
      }}
    />
  );
});
