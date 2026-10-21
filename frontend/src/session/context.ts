import { createContext } from 'react';
import type { User } from '@/api/types';

export interface SessionState {
  user: User | null;
  authReady: boolean;
  login: (email: string, password: string) => Promise<{ ok: boolean; error?: string }>;
  register: (email: string, password: string) => Promise<{ ok: boolean; error?: string }>;
  logout: () => Promise<void>;
}

export const SessionContext = createContext<SessionState | null>(null);
