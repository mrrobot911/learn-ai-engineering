import { createContext } from 'react';
import type { ID, QuizAttempt } from '@/types';
import type { User } from '@/api/types';

export interface StoreValue {
  user: User | null;

  quizHistory: QuizAttempt[];
  addQuizAttempt: (attempt: QuizAttempt) => void;
  getAttempt: (id: ID) => QuizAttempt | undefined;
}

export const StoreContext = createContext<StoreValue | null>(null);
