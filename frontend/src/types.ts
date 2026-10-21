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

export interface Route {
  name: 'topics' | 'topic' | 'lesson' | 'quiz-start' | 'quiz' | 'quiz-result' | 'history' | 'history-detail' | 'login' | 'register';
  params: Record<string, string>;
}
