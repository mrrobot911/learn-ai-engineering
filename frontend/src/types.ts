export type ID = string;

export interface AnswerOption {
  id: ID;
  text: string;
  isCorrect: boolean;
}

export interface Question {
  id: ID;
  topicId: ID;
  text: string;
  explanation: string;
  options: AnswerOption[];
  status?: 'published';
}

export interface Lesson {
  id: ID;
  topicId: ID;
  title: string;
  summary: string;
  body: string;
  position: number;
  durationMin: number;
  status: 'published';
}

export interface Topic {
  id: ID;
  slug: string;
  title: string;
  description: string;
  icon: string;
  color: string;
  position: number;
  status: 'published';
  lessons: Lesson[];
}

export interface QuizResultAnswer {
  questionId: ID;
  questionText: string;
  selectedOptionId: ID | null;
  selectedOptionText: string;
  correctOptionId: ID;
  correctOptionText: string;
  isCorrect: boolean;
  explanation: string;
}

export interface QuizAttempt {
  id: ID;
  topicId: ID;
  topicTitle: string;
  score: number;
  total: number;
  createdAt: number;
  answers: QuizResultAnswer[];
}

export interface Route {
  name: 'topics' | 'topic' | 'lesson' | 'quiz-start' | 'quiz' | 'quiz-result' | 'history' | 'history-detail' | 'login' | 'register';
  params: Record<string, string>;
}
