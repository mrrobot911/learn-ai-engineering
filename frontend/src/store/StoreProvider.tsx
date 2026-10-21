import { useCallback, useEffect, useMemo, useState, type ReactNode } from 'react';
import type { ID, QuizAttempt } from '@/types';
import { seedQuizHistory } from '@/data/content';
import { useSession } from '@/session';
import { StoreContext, type StoreValue } from './context';

const LS_HISTORY = 'forge.history';

function readJSON<T>(key: string, fallback: T): T {
  try {
    const raw = localStorage.getItem(key);
    if (!raw) return fallback;
    return JSON.parse(raw) as T;
  } catch {
    return fallback;
  }
}

function writeJSON(key: string, value: unknown): void {
  try {
    localStorage.setItem(key, JSON.stringify(value));
  } catch {
    // ignore
  }
}

export function StoreProvider({ children }: { children: ReactNode }) {
  const { user } = useSession();
  const [quizHistory, setQuizHistory] = useState<QuizAttempt[]>([]);

  useEffect(() => {
    if (!user) {
      setQuizHistory([]);
      return;
    }
    const hist = readJSON<Record<string, QuizAttempt[]>>(LS_HISTORY, {});
    if (user.email === 'demo@forge.dev' && (!hist[user.id] || hist[user.id].length === 0)) {
      hist[user.id] = seedQuizHistory;
      writeJSON(LS_HISTORY, hist);
    }
    setQuizHistory(hist[user.id] || []);
  }, [user]);

  const persistHistory = useCallback((uid: ID, hist: QuizAttempt[]) => {
    const all = readJSON<Record<string, QuizAttempt[]>>(LS_HISTORY, {});
    all[uid] = hist;
    writeJSON(LS_HISTORY, all);
  }, []);

  const addQuizAttempt = useCallback(
    (attempt: QuizAttempt) => {
      if (!user) return;
      setQuizHistory((prev) => {
        const next = [attempt, ...prev];
        persistHistory(user.id, next);
        return next;
      });
    },
    [user, persistHistory],
  );

  const getAttempt = useCallback((id: ID) => quizHistory.find((a) => a.id === id), [quizHistory]);

  const value = useMemo<StoreValue>(
    () => ({
      user,
      quizHistory,
      addQuizAttempt,
      getAttempt,
    }),
    [user, quizHistory, addQuizAttempt, getAttempt],
  );

  return <StoreContext.Provider value={value}>{children}</StoreContext.Provider>;
}
