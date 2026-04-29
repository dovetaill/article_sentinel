import { createContext, useContext, useMemo, type PropsWithChildren } from 'react';

import { useSessionContext } from './session-context';

type OrgContextValue = {
  activeOrgId: number | null;
  activeOrgName: string;
  isLoading: boolean;
};

const OrgContext = createContext<OrgContextValue | null>(null);

export function OrgProvider({ children }: PropsWithChildren) {
  const { session, isLoading } = useSessionContext();

  const value = useMemo<OrgContextValue>(() => {
    return {
      activeOrgId: session?.orgid ?? null,
      activeOrgName: session?.orgname ?? '',
      isLoading
    };
  }, [isLoading, session]);

  return <OrgContext.Provider value={value}>{children}</OrgContext.Provider>;
}

export function useOrgContext() {
  const context = useContext(OrgContext);
  if (!context) {
    throw new Error('useOrgContext must be used within OrgProvider');
  }
  return context;
}
