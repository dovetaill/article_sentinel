import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type PropsWithChildren
} from 'react';

import { redirectToFixedLogin } from '../lib/request';
import { getSession, logout, type AuthSession } from '../services/auth';

type SessionContextValue = {
  session: AuthSession | null;
  isLoading: boolean;
  logout: () => Promise<void>;
  redirectToLogin: () => void;
};

const SessionContext = createContext<SessionContextValue | null>(null);

export function SessionProvider({ children }: PropsWithChildren) {
  const [session, setSession] = useState<AuthSession | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;

    void getSession()
      .then((currentSession) => {
        if (!cancelled) {
          setSession(currentSession);
        }
      })
      .catch(() => {
        if (!cancelled) {
          setSession(null);
          redirectToFixedLogin();
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

  const handleLogout = useCallback(async () => {
    try {
      await logout();
    } finally {
      setSession(null);
      redirectToFixedLogin();
    }
  }, []);

  const value = useMemo<SessionContextValue>(() => ({
    session,
    isLoading,
    logout: handleLogout,
    redirectToLogin: redirectToFixedLogin
  }), [handleLogout, isLoading, session]);

  return <SessionContext.Provider value={value}>{children}</SessionContext.Provider>;
}

export function useSessionContext() {
  const context = useContext(SessionContext);
  if (!context) {
    throw new Error('useSessionContext must be used within SessionProvider');
  }
  return context;
}
