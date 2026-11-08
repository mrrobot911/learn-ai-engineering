import {
  useCallback,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react';
import type { User } from '@/api/types';
import { api, setOnUnauthorized } from '@/api';
import { toast } from '@/components/toast';
import { SessionContext, type SessionState } from './context';

export function SessionProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [authReady, setAuthReady] = useState(false);

  useEffect(() => {
    let active = true;
    void (async () => {
      try {
        const currentUser = await api.getSession();
        if (active) setUser(currentUser);
      } catch {
        // A missing or expired session is not an error state.
      } finally {
        if (active) setAuthReady(true);
      }
    })();
    return () => {
      active = false;
    };
  }, []);

  useEffect(() => {
    setOnUnauthorized(() => {
      setUser(null);
      toast('Your session expired. Please log in again.', 'error');
    });
    return () => {
      setOnUnauthorized(null);
    };
  }, []);

  const login = useCallback<SessionState['login']>(async (email, password) => {
    try {
      const loggedInUser = await api.login({ email, password });
      setUser(loggedInUser);
      return { ok: true };
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Login failed.';
      return { ok: false, error: message };
    }
  }, []);

  const register = useCallback<SessionState['register']>(
    async (email, password) => {
      try {
        const registeredUser = await api.register({ email, password });
        setUser(registeredUser);
        return { ok: true };
      } catch (error) {
        const message =
          error instanceof Error ? error.message : 'Registration failed.';
        return { ok: false, error: message };
      }
    },
    [],
  );

  const logout = useCallback<SessionState['logout']>(async () => {
    try {
      await api.logout();
    } finally {
      setUser(null);
    }
  }, []);

  const value = useMemo(
    () => ({ user, authReady, login, register, logout }),
    [user, authReady, login, register, logout],
  );

  return (
    <SessionContext.Provider value={value}>{children}</SessionContext.Provider>
  );
}
