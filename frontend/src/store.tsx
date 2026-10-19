import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from 'react';
import type { ID, QuizAttempt, SimUser } from '@/types';
import { seedQuizHistory } from '@/data/content';

const LS_USER = 'forge.currentUser';
const LS_USERS = 'forge.users';
const LS_PROGRESS = 'forge.progress'; // userId -> lessonId[]
const LS_HISTORY = 'forge.history'; // userId -> QuizAttempt[]

interface StoredAccount {
  id: ID;
  email: string;
  name: string;
  password: string; // simulated only — NOT secure
  isMaintainer: boolean;
  createdAt: number;
}

interface StoreValue {
  user: SimUser | null;
  authReady: boolean;
  login: (email: string, password: string) => { ok: boolean; error?: string };
  register: (name: string, email: string, password: string) => { ok: boolean; error?: string };
  logout: () => void;

  completedLessons: ID[];
  toggleLessonComplete: (lessonId: ID) => void;
  isLessonComplete: (lessonId: ID) => boolean;

  quizHistory: QuizAttempt[];
  addQuizAttempt: (attempt: QuizAttempt) => void;
  getAttempt: (id: ID) => QuizAttempt | undefined;
}

const StoreContext = createContext<StoreValue | null>(null);

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
    // ignore quota errors in the prototype
  }
}

export function StoreProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<SimUser | null>(null);
  const [authReady, setAuthReady] = useState(false);
  const [completedLessons, setCompletedLessons] = useState<ID[]>([]);
  const [quizHistory, setQuizHistory] = useState<QuizAttempt[]>([]);

  // Bootstrap: load persisted user + their data.
  useEffect(() => {
    // Seed the demo account once so the "Try demo" flow works.
    const accounts = readJSON<StoredAccount[]>(LS_USERS, []);
    if (!accounts.some((a) => a.email === 'demo@forge.dev')) {
      accounts.push({
        id: 'usr-demo',
        email: 'demo@forge.dev',
        name: 'Ada Demo',
        password: 'password123',
        isMaintainer: true,
        createdAt: Date.now() - 1000 * 60 * 60 * 24 * 30,
      });
      writeJSON(LS_USERS, accounts);
    }
    const stored = readJSON<SimUser | null>(LS_USER, null);
    if (stored) {
      setUser(stored);
      const prog = readJSON<Record<string, ID[]>>(LS_PROGRESS, {});
      setCompletedLessons(prog[stored.id] || []);
      const hist = readJSON<Record<string, QuizAttempt[]>>(LS_HISTORY, {});
      // Seed history for the demo account on first load.
      if (stored.email === 'demo@forge.dev' && (!hist[stored.id] || hist[stored.id].length === 0)) {
        hist[stored.id] = seedQuizHistory;
        writeJSON(LS_HISTORY, hist);
      }
      setQuizHistory(hist[stored.id] || []);
    }
    setAuthReady(true);
  }, []);

  const persistProgress = useCallback((uid: ID, lessons: ID[]) => {
    const all = readJSON<Record<string, ID[]>>(LS_PROGRESS, {});
    all[uid] = lessons;
    writeJSON(LS_PROGRESS, all);
  }, []);

  const persistHistory = useCallback((uid: ID, hist: QuizAttempt[]) => {
    const all = readJSON<Record<string, QuizAttempt[]>>(LS_HISTORY, {});
    all[uid] = hist;
    writeJSON(LS_HISTORY, all);
  }, []);

  const login = useCallback<StoreValue['login']>((email, password) => {
    const accounts = readJSON<StoredAccount[]>(LS_USERS, []);
    const found = accounts.find((a) => a.email.toLowerCase() === email.trim().toLowerCase());
    if (!found) return { ok: false, error: 'No account found with that email. Try registering first.' };
    if (found.password !== password) return { ok: false, error: 'Incorrect password. Please try again.' };
    const sim: SimUser = { id: found.id, email: found.email, name: found.name, isMaintainer: found.isMaintainer, createdAt: found.createdAt };
    setUser(sim);
    writeJSON(LS_USER, sim);
    const prog = readJSON<Record<string, ID[]>>(LS_PROGRESS, {});
    setCompletedLessons(prog[sim.id] || []);
    const hist = readJSON<Record<string, QuizAttempt[]>>(LS_HISTORY, {});
    if (sim.email === 'demo@forge.dev' && (!hist[sim.id] || hist[sim.id].length === 0)) {
      hist[sim.id] = seedQuizHistory;
      writeJSON(LS_HISTORY, hist);
    }
    setQuizHistory(hist[sim.id] || []);
    return { ok: true };
  }, []);

  const register = useCallback<StoreValue['register']>((name, email, password) => {
    const accounts = readJSON<StoredAccount[]>(LS_USERS, []);
    const exists = accounts.some((a) => a.email.toLowerCase() === email.trim().toLowerCase());
    if (exists) return { ok: false, error: 'An account with that email already exists.' };
    const account: StoredAccount = {
      id: `usr-${Math.random().toString(36).slice(2, 10)}`,
      email: email.trim(),
      name: name.trim(),
      password,
      isMaintainer: false,
      createdAt: Date.now(),
    };
    accounts.push(account);
    writeJSON(LS_USERS, accounts);
    const sim: SimUser = { id: account.id, email: account.email, name: account.name, isMaintainer: account.isMaintainer, createdAt: account.createdAt };
    setUser(sim);
    writeJSON(LS_USER, sim);
    setCompletedLessons([]);
    setQuizHistory([]);
    return { ok: true };
  }, []);

  const logout = useCallback(() => {
    setUser(null);
    setCompletedLessons([]);
    setQuizHistory([]);
    try { localStorage.removeItem(LS_USER); } catch { /* ignore */ }
  }, []);

  const toggleLessonComplete = useCallback((lessonId: ID) => {
    if (!user) return;
    setCompletedLessons((prev) => {
      const next = prev.includes(lessonId)
        ? prev.filter((id) => id !== lessonId)
        : [...prev, lessonId];
      persistProgress(user.id, next);
      return next;
    });
  }, [user, persistProgress]);

  const isLessonComplete = useCallback((lessonId: ID) => completedLessons.includes(lessonId), [completedLessons]);

  const addQuizAttempt = useCallback((attempt: QuizAttempt) => {
    if (!user) return; // anonymous attempts are not stored
    setQuizHistory((prev) => {
      const next = [attempt, ...prev];
      persistHistory(user.id, next);
      return next;
    });
  }, [user, persistHistory]);

  const getAttempt = useCallback((id: ID) => quizHistory.find((a) => a.id === id), [quizHistory]);

  const value = useMemo<StoreValue>(() => ({
    user, authReady, login, register, logout,
    completedLessons, toggleLessonComplete, isLessonComplete,
    quizHistory, addQuizAttempt, getAttempt,
  }), [user, authReady, login, register, logout, completedLessons, toggleLessonComplete, isLessonComplete, quizHistory, addQuizAttempt, getAttempt]);

  return <StoreContext.Provider value={value}>{children}</StoreContext.Provider>;
}

export function useStore(): StoreValue {
  const ctx = useContext(StoreContext);
  if (!ctx) throw new Error('useStore must be used within StoreProvider');
  return ctx;
}
