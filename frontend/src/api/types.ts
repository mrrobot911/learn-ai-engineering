// The shapes a screen sees. US-2: the correct option and the explanation never
// reach the client before a submission, so they are absent from the quiz types.

export interface Lesson {
  id: string;
  topicId: string;
  title: string;
  summary: string;
  body: string;
  position: number;
  durationMin: number;
}

export interface TopicSummary {
  id: string;
  slug: string;
  title: string;
  description: string;
  icon: string;
  color: string;
  position: number;
}

export interface Topic extends TopicSummary {
  lessons: Lesson[];
}

export interface QuizOption {
  id: string;
  text: string;
}

export interface QuizQuestion {
  id: string;
  text: string;
  options: QuizOption[];
}

export interface QuizAnswerInput {
  questionId: string;
  optionId: string | null;
}

export interface QuizSubmission {
  topicId: string;
  answers: QuizAnswerInput[];
}

export interface QuizAnswerResult {
  questionId: string;
  questionText: string;
  selectedOptionId: string | null;
  selectedOptionText: string;
  correctOptionId: string;
  correctOptionText: string;
  isCorrect: boolean;
  explanation: string;
}

export interface QuizResult {
  score: number;
  total: number;
  answers: QuizAnswerResult[];
}

export interface QuizAttemptSummary {
  id: string;
  topicId: string;
  topicTitle: string;
  score: number;
  total: number;
  createdAt: string;
}

export interface QuizAttempt extends QuizAttemptSummary {
  answers: QuizAnswerResult[];
}

export interface Credentials {
  email: string;
  password: string;
}

export interface User {
  id: string;
  email: string;
  isMaintainer: boolean;
  createdAt: string;
}

/** The only surface the frontend has on the backend (product-spec.md section 7). */
export interface ApiClient {
  listTopics(): Promise<TopicSummary[]>;
  getTopic(topicId: string): Promise<Topic>;
  getLesson(lessonId: string): Promise<Lesson>;
  startQuiz(topicId: string): Promise<QuizQuestion[]>;
  submitQuiz(submission: QuizSubmission): Promise<QuizResult>;
  register(credentials: Credentials): Promise<User>;
  login(credentials: Credentials): Promise<User>;
  logout(): Promise<void>;
  getProgress(): Promise<string[]>;
  markLessonComplete(lessonId: string): Promise<void>;
  undoLessonComplete(lessonId: string): Promise<void>;
  listAttempts(): Promise<QuizAttemptSummary[]>;
  getAttempt(attemptId: string): Promise<QuizAttempt>;
}
