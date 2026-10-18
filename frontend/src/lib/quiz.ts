import type { Question } from '@/types';

// Deterministic shuffle so a given attempt is stable across renders.
export function shuffle<T>(arr: T[], seed: number): T[] {
  const a = [...arr];
  let s = seed;
  for (let i = a.length - 1; i > 0; i--) {
    s = (s * 1103515245 + 12345) & 0x7fffffff;
    const j = s % (i + 1);
    [a[i], a[j]] = [a[j], a[i]];
  }
  return a;
}

export const QUIZ_SIZE = 5;

export function pickQuizQuestions(all: Question[], seed: number): Question[] {
  const shuffled = shuffle(all, seed);
  return shuffled.slice(0, Math.min(QUIZ_SIZE, shuffled.length));
}

export function makeSeed(): number {
  return Math.floor(Math.random() * 1e9) + 1;
}

export function gradeAnswer(question: Question, selectedOptionId: string | null): {
  isCorrect: boolean;
  correctOptionId: string;
  correctOptionText: string;
  selectedOptionText: string;
} {
  const correct = question.options.find((o) => o.isCorrect);
  const selected = selectedOptionId ? question.options.find((o) => o.id === selectedOptionId) : null;
  return {
    isCorrect: !!selected && !!correct && selected.id === correct.id,
    correctOptionId: correct?.id ?? '',
    correctOptionText: correct?.text ?? '',
    selectedOptionText: selected?.text ?? '',
  };
}
