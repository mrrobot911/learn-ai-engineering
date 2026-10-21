import { useCallback, useMemo, useState, type ReactNode } from 'react';
import type { User } from '@/api/types';
import { api } from '@/api';
import { SessionContext, type SessionState } from './context';

export function SessionProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [authReady] = useState(true);

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

  const register = useCallback<SessionState['register']>(async (email, password) => {
    try {
      const registeredUser = await api.register({ email, password });
      setUser(registeredUser);
      return { ok: true };
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Registration failed.';
      return { ok: false, error: message };
    }
  }, []);

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

  return <SessionContext.Provider value={value}>{children}</SessionContext.Provider>;
}
