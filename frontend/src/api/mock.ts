import {
  getLessonById,
  getQuestionsForTopic,
  getTopicBySlug,
  seedQuizHistory,
  topics,
} from '@/data/content';
import type { Question as FixtureQuestion, Topic as FixtureTopic } from '@/types';
import type {
  ApiClient,
  QuizAnswerResult,
  QuizAttempt,
  QuizAttemptSummary,
  QuizQuestion,
  QuizResult,
  User,
} from './types';

// product-spec.md US-2 fixes the quiz size; fewer questions are returned when a
// topic has fewer than that.
const QUIZ_SIZE = 5;

interface StoredAttempt extends QuizAttempt {
  userId: string;
}

function seedUsers(): User[] {
  return [
    {
      id: 'usr-demo',
      email: 'demo@forge.dev',
      isMaintainer: true,
      createdAt: new Date(Date.now() - 1000 * 60 * 60 * 24 * 30).toISOString(),
    },
  ];
}

function seedAttempts(): StoredAttempt[] {
  return seedQuizHistory.map((attempt) => ({
    ...attempt,
    userId: 'usr-demo',
    createdAt: new Date(attempt.createdAt).toISOString(),
  }));
}

// The mock keeps the session, the progress and the attempts in memory: it stands
// in for the backend, not for the browser storage the prototype used.
let users = seedUsers();
let attempts = seedAttempts();
let progress = new Map<string, Set<string>>();
let sessionUserId: string | null = null;
let nextAttemptNumber = 1;

/** Restores the fixture state so each test starts from the same point. */
export function resetMockState(): void {
  users = seedUsers();
  attempts = seedAttempts();
  progress = new Map();
  sessionUserId = null;
  nextAttemptNumber = 1;
}

function requireSession(): string {
  if (!sessionUserId) throw new Error('You must be logged in to do that.');
  return sessionUserId;
}

// The not-found signal: an error whose status is 404. The later HTTP client
// maps a real 404 onto the same shape (T27), so screens match on status only.
function notFoundError(message: string): Error & { status: number } {
  const error = new Error(message) as Error & { status: number };
  error.status = 404;
  return error;
}

function requireTopicBySlug(slug: string): FixtureTopic {
  const topic = getTopicBySlug(slug);
  if (!topic) throw notFoundError(`No topic with slug "${slug}".`);
  return topic;
}

function pickRandom<T>(items: T[], count: number): T[] {
  const pool = [...items];
  for (let i = pool.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [pool[i], pool[j]] = [pool[j], pool[i]];
  }
  return pool.slice(0, Math.min(count, pool.length));
}

// US-2: the question reaches the client without the correct option or the explanation.
function toQuizQuestion(question: FixtureQuestion): QuizQuestion {
  return {
    id: question.id,
    text: question.text,
    options: question.options.map(({ id, text }) => ({ id, text })),
  };
}

// US-2: the server grades; an unanswered question or an option from another
// question counts as incorrect.
function grade(question: FixtureQuestion, optionId: string | null): QuizAnswerResult {
  const correct = question.options.find((option) => option.isCorrect);
  const selected = optionId
    ? question.options.find((option) => option.id === optionId)
    : undefined;
  return {
    questionId: question.id,
    questionText: question.text,
    selectedOptionId: selected?.id ?? null,
    selectedOptionText: selected?.text ?? '',
    correctOptionId: correct?.id ?? '',
    correctOptionText: correct?.text ?? '',
    isCorrect: !!selected && selected.id === correct?.id,
    explanation: question.explanation,
  };
}

function toSummary(attempt: StoredAttempt): QuizAttemptSummary {
  const { id, topicId, topicTitle, score, total, createdAt } = attempt;
  return { id, topicId, topicTitle, score, total, createdAt };
}

function progressOf(userId: string): Set<string> {
  const done = progress.get(userId) ?? new Set<string>();
  progress.set(userId, done);
  return done;
}

export const mockClient: ApiClient = {
  async listTopics() {
    return topics;
  },

  async getTopic(slug) {
    return requireTopicBySlug(slug);
  },

  async getLesson(lessonId) {
    const found = getLessonById(lessonId);
    if (!found) throw notFoundError(`No lesson with id "${lessonId}".`);
    return found.lesson;
  },

  async startQuiz(slug) {
    const topic = requireTopicBySlug(slug);
    const questions = getQuestionsForTopic(topic.id);
    if (questions.length === 0) throw new Error(`Topic "${slug}" has no published questions.`);
    return pickRandom(questions, QUIZ_SIZE).map(toQuizQuestion);
  },

  async submitQuiz({ topicSlug, answers }) {
    const topic = requireTopicBySlug(topicSlug);
    const byId = new Map(getQuestionsForTopic(topic.id).map((q) => [q.id, q]));
    const graded = answers.map((answer) => {
      const question = byId.get(answer.questionId);
      if (!question) throw new Error(`Question "${answer.questionId}" does not belong to topic "${topicSlug}".`);
      return grade(question, answer.optionId);
    });
    const result: QuizResult = {
      score: graded.filter((answer) => answer.isCorrect).length,
      total: graded.length,
      answers: graded,
    };
    // US-3: an anonymous result is returned but never stored.
    if (sessionUserId) {
      attempts.unshift({
        id: `att-mock-${nextAttemptNumber++}`,
        userId: sessionUserId,
        topicId: topic.id,
        topicTitle: topic.title,
        createdAt: new Date().toISOString(),
        ...result,
      });
    }
    return result;
  },

  async register({ email }) {
    const address = email.trim().toLowerCase();
    if (users.some((user) => user.email === address)) {
      throw new Error('An account with that email already exists.');
    }
    const user: User = {
      id: `usr-${address.replace(/[^a-z0-9]/g, '-')}`,
      email: address,
      isMaintainer: false,
      createdAt: new Date().toISOString(),
    };
    users.push(user);
    sessionUserId = user.id;
    return user;
  },

  // The mock stores no password at all: it only checks that the account exists.
  // Password checks belong to the backend the real client (#27-#30) will call.
  async login({ email }) {
    const found = users.find((user) => user.email === email.trim().toLowerCase());
    if (!found) throw new Error('No account found with that email.');
    sessionUserId = found.id;
    return found;
  },

  async logout() {
    sessionUserId = null;
  },

  async getSession() {
    if (!sessionUserId) {
      const error = new Error('No valid session.') as Error & { status: number };
      error.status = 401;
      throw error;
    }
    const user = users.find((u) => u.id === sessionUserId);
    if (!user) {
      const error = new Error('No valid session.') as Error & { status: number };
      error.status = 401;
      throw error;
    }
    return user;
  },

  async getProgress() {
    return [...(progress.get(requireSession()) ?? [])];
  },

  async markLessonComplete(lessonId) {
    progressOf(requireSession()).add(lessonId);
  },

  async undoLessonComplete(lessonId) {
    progressOf(requireSession()).delete(lessonId);
  },

  async listAttempts() {
    const userId = requireSession();
    return attempts
      .filter((attempt) => attempt.userId === userId)
      .sort((a, b) => b.createdAt.localeCompare(a.createdAt))
      .map(toSummary);
  },

  async getAttempt(attemptId) {
    const userId = requireSession();
    const found = attempts.find((attempt) => attempt.userId === userId && attempt.id === attemptId);
    if (!found) throw new Error(`No attempt with id "${attemptId}".`);
    return found;
  },
};
