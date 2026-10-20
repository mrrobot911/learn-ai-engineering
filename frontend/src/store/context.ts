import { createContext } from 'react';
import type { ID, QuizAttempt } from '@/types';

export interface StoreValue {
  completedLessons: ID[];
  toggleLessonComplete: (lessonId: ID) => void;
  isLessonComplete: (lessonId: ID) => boolean;

  quizHistory: QuizAttempt[];
  addQuizAttempt: (attempt: QuizAttempt) => void;
  getAttempt: (id: ID) => QuizAttempt | undefined;
}

export const StoreContext = createContext<StoreValue | null>(null);
