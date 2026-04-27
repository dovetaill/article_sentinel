import { createContext, useContext, useEffect, useMemo, useState, type PropsWithChildren } from 'react';

import { listOrgs, type OrgRecord } from '../services/orgs';

type OrgContextValue = {
  activeOrgId: number | null;
  activeOrgName: string;
  orgs: OrgRecord[];
  isLoading: boolean;
  setActiveOrgId: (orgId: number) => void;
};

const OrgContext = createContext<OrgContextValue | null>(null);

export function OrgProvider({ children }: PropsWithChildren) {
  const [orgs, setOrgs] = useState<OrgRecord[]>([]);
  const [activeOrgId, setActiveOrgIdState] = useState<number | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;

    void listOrgs()
      .then((items) => {
        if (cancelled) return;
        setOrgs(items);
        setActiveOrgIdState((current) => {
          if (current && items.some((item) => item.id === current)) {
            return current;
          }

          // 当前本地种子数据会优先使用 org 29；若不存在则回退到第一个组织。
          const defaultOrg = items.find((item) => item.id === 29) ?? items[0];
          return defaultOrg?.id ?? null;
        });
      })
      .catch(() => {
        if (!cancelled) {
          setOrgs([]);
        }
      })
      .finally(() => {
        if (!cancelled) {
          setIsLoading(false);
        }
      });

    return () => {
      cancelled = true;
    };
  }, []);

  const value = useMemo<OrgContextValue>(() => {
    const activeOrg = orgs.find((item) => item.id === activeOrgId);

    return {
      activeOrgId,
      activeOrgName: activeOrg?.name ?? '',
      orgs,
      isLoading,
      setActiveOrgId: (orgId: number) => setActiveOrgIdState(orgId)
    };
  }, [activeOrgId, isLoading, orgs]);

  return <OrgContext.Provider value={value}>{children}</OrgContext.Provider>;
}

export function useOrgContext() {
  const context = useContext(OrgContext);
  if (!context) {
    throw new Error('useOrgContext must be used within OrgProvider');
  }
  return context;
}
