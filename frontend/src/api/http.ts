import type {
  ApiClient,
  Credentials,
  Lesson,
  QuizAnswerInput,
  QuizAnswerResult,
  QuizAttempt,
  QuizAttemptSummary,
  QuizQuestion,
  QuizResult,
  QuizSubmission,
  Topic,
  TopicSummary,
  User,
} from './types';
import { ApiError } from './errors';
import {
  camelToSnake,
  mapLesson,
  mapTopic,
  mapTopicSummary,
  snakeToCamel,
} from './mappers';

type HttpMethod = 'GET' | 'POST' | 'PUT' | 'DELETE';

interface RequestOptions {
  method?: HttpMethod;
  body?: unknown;
  /** True for state-changing requests that must carry X-CSRF-Token when present. */
  csrf?: boolean;
}

let csrfToken: string | null = null;
let topicTitleCache = new Map<string, string>();
let quizQuestionsCache = new Map<string, QuizQuestion[]>();

/** Visible for tests: reset in-memory CSRF/caches. */
export function resetHttpState(): void {
  csrfToken = null;
  topicTitleCache = new Map();
  quizQuestionsCache = new Map();
}

export function getCsrfToken(): string | null {
  return csrfToken;
}

function baseUrl(): string {
  const url = import.meta.env.VITE_API_BASE_URL;
  if (typeof url === 'string' && url !== '') return url.replace(/\/$/, '');
  return '';
}

function captureCsrf(payload: unknown): void {
  if (!payload || typeof payload !== 'object') return;
  const token =
    (payload as { csrfToken?: unknown; csrf_token?: unknown }).csrfToken ??
    (payload as { csrf_token?: unknown }).csrf_token;
  if (typeof token === 'string' && token) csrfToken = token;
}

function cacheTopicTitle(id: string, slug: string, title: string): void {
  if (id) topicTitleCache.set(id, title);
  if (slug) topicTitleCache.set(slug, title);
}

function topicTitleFor(id?: string | null, slug?: string | null): string {
  if (slug && topicTitleCache.has(slug)) return topicTitleCache.get(slug)!;
  if (id && topicTitleCache.has(id)) return topicTitleCache.get(id)!;
  return slug ?? id ?? '';
}

async function doRequest<T>(
  path: string,
  options: RequestOptions = {},
): Promise<T> {
  const { method = 'GET', body, csrf = false } = options;
  const headers: Record<string, string> = {};
  if (body !== undefined) headers['Content-Type'] = 'application/json';
  if (csrf && csrfToken) headers['X-CSRF-Token'] = csrfToken;

  let response: Response;
  try {
    response = await fetch(`${baseUrl()}${path}`, {
      method,
      headers,
      body: body === undefined ? undefined : JSON.stringify(camelToSnake(body)),
      credentials: 'include',
    });
  } catch (error) {
    const message =
      error instanceof Error ? error.message : 'Network request failed.';
    throw new ApiError(0, 'network_error', message);
  }

  let payload: unknown = null;
  const text = await response.text();
  if (text) {
    try {
      payload = JSON.parse(text);
    } catch {
      payload = text;
    }
  }

  if (!response.ok) {
    const err = (
      payload && typeof payload === 'object'
        ? snakeToCamel<Record<string, unknown>>(payload)
        : {}
    ) as {
      code?: string;
      message?: string;
    };
    const code =
      typeof err.code === 'string' ? err.code : `http_${response.status}`;
    const message =
      typeof err.message === 'string'
        ? err.message
        : response.statusText || 'Request failed.';
    throw new ApiError(response.status, code, message);
  }

  captureCsrf(payload);
  return snakeToCamel<T>(payload);
}

async function request<T>(
  path: string,
  options: RequestOptions = {},
): Promise<T> {
  try {
    return await doRequest<T>(path, options);
  } catch (error) {
    if (
      options.csrf &&
      error instanceof ApiError &&
      error.status === 403 &&
      error.code === 'forbidden'
    ) {
      await doRequest<RawAuthResponse>('/auth/session');
      return doRequest<T>(path, options);
    }
    throw error;
  }
}

interface RawAuthResponse {
  user: { id: string; email: string };
  csrfToken: string;
}

function mapUser(raw: RawAuthResponse['user']): User {
  return {
    id: raw.id,
    email: raw.email,
  };
}

function mapQuizResult(
  raw: {
    score: number;
    total: number;
    answers: Array<{
      questionId: string;
      isCorrect: boolean;
      correctOptionText: string;
      explanation: string;
    }>;
  },
  questions: QuizQuestion[],
  submitted: QuizAnswerInput[],
): QuizResult {
  const qById = new Map(questions.map((q) => [q.id, q]));
  const submittedById = new Map(
    submitted.map((a) => [a.questionId, a.optionId]),
  );
  return {
    score: raw.score,
    total: raw.total,
    answers: raw.answers.map((a): QuizAnswerResult => {
      const q = qById.get(a.questionId);
      const selectedId = submittedById.get(a.questionId) ?? null;
      const selected = q?.options.find((o) => o.id === selectedId);
      const correct = q?.options.find((o) => o.text === a.correctOptionText);
      return {
        questionId: a.questionId,
        questionText: q?.text ?? '',
        selectedOptionId: selected?.id ?? null,
        selectedOptionText: selected?.text ?? '',
        correctOptionId: correct?.id ?? '',
        correctOptionText: a.correctOptionText,
        isCorrect: a.isCorrect,
        explanation: a.explanation,
      };
    }),
  };
}

function mapAttemptSummary(raw: {
  id: string;
  topicId: string;
  topicSlug: string;
  score: number;
  total: number;
  createdAt: string;
}): QuizAttemptSummary {
  return {
    id: raw.id,
    topicId: raw.topicId,
    topicTitle: topicTitleFor(raw.topicId, raw.topicSlug),
    score: raw.score,
    total: raw.total,
    createdAt: raw.createdAt,
  };
}

function mapAttempt(raw: {
  id: string;
  topicId: string;
  topicSlug: string;
  score: number;
  total: number;
  createdAt: string;
  answers: Array<{
    questionId: string | null;
    questionText: string;
    selectedOptionText: string;
    correctOptionText: string;
    isCorrect: boolean;
  }>;
}): QuizAttempt {
  return {
    ...mapAttemptSummary(raw),
    answers: raw.answers.map((a): QuizAnswerResult => ({
      questionId: a.questionId ?? '',
      questionText: a.questionText,
      selectedOptionId: a.selectedOptionText ? '' : null,
      selectedOptionText: a.selectedOptionText ?? '',
      correctOptionId: '',
      correctOptionText: a.correctOptionText,
      isCorrect: a.isCorrect,
      explanation: '',
    })),
  };
}

export function createHttpClient(): ApiClient {
  return {
    async listTopics(): Promise<TopicSummary[]> {
      const raw = await request<
        Array<{
          id: string;
          slug: string;
          title: string;
          description: string;
          position: number;
        }>
      >('/topics');
      const topics = raw.map(mapTopicSummary);
      for (const t of topics) cacheTopicTitle(t.id, t.slug, t.title);
      return topics;
    },

    async getTopic(slug: string): Promise<Topic> {
      const raw = await request<{
        id: string;
        slug: string;
        title: string;
        description: string;
        position: number;
        lessons: Array<{
          id: string;
          topicId: string;
          title: string;
          body: string;
          position: number;
        }>;
      }>(`/topics/${encodeURIComponent(slug)}`);
      cacheTopicTitle(raw.id, raw.slug, raw.title);
      return mapTopic(raw);
    },

    async getLesson(lessonId: string): Promise<Lesson> {
      const raw = await request<{
        id: string;
        topicId: string;
        topicSlug?: string;
        title: string;
        body: string;
        position: number;
      }>(`/lessons/${encodeURIComponent(lessonId)}`);
      let topicSlug = raw.topicSlug;
      if (!topicSlug) {
        const topics = await request<
          Array<{
            id: string;
            slug: string;
            title: string;
            description: string;
            position: number;
          }>
        >('/topics');
        const topic = topics.find((t) => t.id === raw.topicId);
        if (!topic)
          throw new ApiError(404, 'not_found', 'Topic not found for lesson.');
        cacheTopicTitle(topic.id, topic.slug, topic.title);
        topicSlug = topic.slug;
      }
      return mapLesson({ ...raw, topicSlug });
    },

    async startQuiz(slug: string): Promise<QuizQuestion[]> {
      const raw = await request<QuizQuestion[]>(
        `/topics/${encodeURIComponent(slug)}/quiz`,
        {
          method: 'POST',
        },
      );
      quizQuestionsCache.set(slug, raw);
      return raw;
    },

    async submitQuiz(submission: QuizSubmission): Promise<QuizResult> {
      const { topicSlug, answers } = submission;
      const raw = await request<{
        score: number;
        total: number;
        answers: Array<{
          questionId: string;
          isCorrect: boolean;
          correctOptionText: string;
          explanation: string;
        }>;
      }>(`/topics/${encodeURIComponent(topicSlug)}/quiz/submissions`, {
        method: 'POST',
        body: { answers },
        csrf: true,
      });
      const questions = quizQuestionsCache.get(topicSlug) ?? [];
      return mapQuizResult(raw, questions, answers);
    },

    async register(credentials: Credentials): Promise<User> {
      const raw = await request<RawAuthResponse>('/auth/register', {
        method: 'POST',
        body: credentials,
      });
      return mapUser(raw.user);
    },

    async login(credentials: Credentials): Promise<User> {
      const raw = await request<RawAuthResponse>('/auth/login', {
        method: 'POST',
        body: credentials,
      });
      return mapUser(raw.user);
    },

    async logout(): Promise<void> {
      await request<unknown>('/auth/logout', { method: 'POST', csrf: true });
      csrfToken = null;
    },

    async getSession(): Promise<User> {
      const raw = await request<RawAuthResponse>('/auth/session');
      return mapUser(raw.user);
    },

    async getProgress(): Promise<string[]> {
      return request<string[]>('/me/progress');
    },

    async markLessonComplete(lessonId: string): Promise<void> {
      await request<unknown>(
        `/me/progress/lessons/${encodeURIComponent(lessonId)}`,
        {
          method: 'PUT',
          csrf: true,
        },
      );
    },

    async undoLessonComplete(lessonId: string): Promise<void> {
      await request<unknown>(
        `/me/progress/lessons/${encodeURIComponent(lessonId)}`,
        {
          method: 'DELETE',
          csrf: true,
        },
      );
    },

    async listAttempts(): Promise<QuizAttemptSummary[]> {
      const raw = await request<
        Array<{
          id: string;
          topicId: string;
          topicSlug: string;
          score: number;
          total: number;
          createdAt: string;
        }>
      >('/me/attempts');
      return raw.map(mapAttemptSummary);
    },

    async getAttempt(attemptId: string): Promise<QuizAttempt> {
      const raw = await request<{
        id: string;
        topicId: string;
        topicSlug: string;
        score: number;
        total: number;
        createdAt: string;
        answers: Array<{
          questionId: string | null;
          questionText: string;
          selectedOptionText: string;
          correctOptionText: string;
          isCorrect: boolean;
        }>;
      }>(`/me/attempts/${encodeURIComponent(attemptId)}`);
      return mapAttempt(raw);
    },
  };
}

export const httpClient = createHttpClient();
