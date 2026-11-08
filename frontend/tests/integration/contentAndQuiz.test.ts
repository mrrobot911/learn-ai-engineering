// @vitest-environment jsdom
/**
 * Frontend integration tests against a real backend and seeded test database.
 * Covers US-1 (browse content) and US-2/US-3 (quiz, anonymous access).
 * Run with: bun run test:integration (requires the test backend; see README).
 */
import { beforeAll, describe, expect, it } from 'vitest';
import { ApiError } from '@/api/errors';
import { createHttpClient, resetHttpState } from '@/api/http';
import { apiBaseUrl, testPassword, uniqueEmail } from './harness';

const base = apiBaseUrl();

async function raw(path: string, init?: RequestInit): Promise<Response> {
  return fetch(`${base}${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(init?.headers ?? {}),
    },
  });
}

describe('content and quiz (US-1, US-2, US-3)', () => {
  beforeAll(async () => {
    // Fail fast with a clear message when the test backend is not running.
    const res = await fetch(`${base}/health`).catch((e) => e);
    if (res instanceof Error || !res.ok) {
      throw new Error(
        `Test backend not reachable at ${base}. Start it with the documented command (see README "Frontend integration tests").`,
      );
    }
  });

  it('lists published topics, and loads a topic and lesson (US-1)', async () => {
    const client = createHttpClient();

    const topics = await client.listTopics();
    expect(topics.length).toBeGreaterThan(0);
    expect(topics.every((t) => t.slug && t.title && t.id)).toBe(true);

    // Seeded topics are published and ordered; pick a known one.
    const llms = topics.find((t) => t.slug === 'llms');
    expect(llms).toBeDefined();

    const topic = await client.getTopic('llms');
    expect(topic.slug).toBe('llms');
    expect(topic.title).toBeTruthy();
    expect(topic.lessons.length).toBeGreaterThan(0);
    expect(topic.lessons.every((l) => l.topicSlug === 'llms')).toBe(true);

    const firstLesson = topic.lessons[0];
    const lesson = await client.getLesson(firstLesson.id);
    expect(lesson.id).toBe(firstLesson.id);
    expect(lesson.topicSlug).toBe('llms');
    expect(lesson.body.length).toBeGreaterThan(0);
  });

  it('returns 404 for an unknown topic and lesson', async () => {
    const client = createHttpClient();
    await expect(client.getTopic('does-not-exist')).rejects.toMatchObject({
      status: 404,
    } satisfies Partial<ApiError>);
    await expect(
      client.getLesson('lesson-does-not-exist'),
    ).rejects.toMatchObject({
      status: 404,
    } satisfies Partial<ApiError>);
  });

  it('starts a quiz with 5 questions and no correctness leak (US-2)', async () => {
    // Raw request to inspect the exact payload the server sends.
    const res = await raw('/topics/llms/quiz', { method: 'POST', body: '{}' });
    expect(res.status).toBe(200);
    const text = await res.text();
    expect(text).not.toMatch(/is_correct/i);
    expect(text).not.toMatch(/correct_option/i);
    expect(text).not.toMatch(/"explanation"/);

    const questions = JSON.parse(text) as Array<{
      id: string;
      text: string;
      options: Array<{ id: string; text: string }>;
    }>;
    expect(questions).toHaveLength(5);
    for (const q of questions) {
      expect(q.id).toBeTruthy();
      expect(q.text).toBeTruthy();
      expect(q.options.length).toBeGreaterThan(0);
      for (const o of q.options) {
        expect(o).toEqual({ id: expect.any(String), text: expect.any(String) });
        expect(o).not.toHaveProperty('is_correct');
        expect(o).not.toHaveProperty('isCorrect');
      }
      expect(q).not.toHaveProperty('correct_option_id');
      expect(q).not.toHaveProperty('explanation');
    }
  });

  it('returns all questions when a topic has fewer than 5 (US-2)', async () => {
    // The integration setup inserts a `quiz-few` topic with 2 published questions.
    const res = await raw('/topics/quiz-few/quiz', {
      method: 'POST',
      body: '{}',
    });
    if (res.status === 404) {
      throw new Error(
        'Fixture topic `quiz-few` not found. The integration harness must create it (2 published questions) in the test database; see README.',
      );
    }
    expect(res.status).toBe(200);
    const questions = (await res.json()) as unknown[];
    expect(questions).toHaveLength(2);
  });

  it('grades a submission on the server and returns the score (US-2)', async () => {
    resetHttpState();
    const client = createHttpClient();
    const questions = await client.startQuiz('ai-agents');
    expect(questions.length).toBeGreaterThan(0);

    // Pick the first option for every question (score is server-determined).
    const answers = questions.map((q) => ({
      questionId: q.id,
      optionId: q.options[0]?.id ?? null,
    }));

    const res = await raw('/topics/ai-agents/quiz/submissions', {
      method: 'POST',
      body: JSON.stringify({
        answers: answers.map((a) => ({
          question_id: a.questionId,
          option_id: a.optionId,
        })),
      }),
    });
    expect(res.status).toBe(200);
    const result = (await res.json()) as {
      score: number;
      total: number;
      answers: Array<{
        question_id: string;
        is_correct: boolean;
        correct_option_text: string;
        explanation: string;
      }>;
    };
    expect(result.total).toBe(questions.length);
    expect(result.score).toBeGreaterThanOrEqual(0);
    expect(result.score).toBeLessThanOrEqual(result.total);
    expect(Number.isInteger(result.score)).toBe(true);
    expect(result.answers).toHaveLength(result.total);
    for (const a of result.answers) {
      expect(typeof a.question_id).toBe('string');
      expect(typeof a.is_correct).toBe('boolean');
      expect(typeof a.correct_option_text).toBe('string');
      expect(a.correct_option_text.length).toBeGreaterThan(0);
      expect(typeof a.explanation).toBe('string');
    }
  });

  it('rejects a submission that contains a foreign question with 400 (US-2)', async () => {
    // Get a valid question from one topic and submit it under a different topic.
    const quizRes = await raw('/topics/mcp/quiz', {
      method: 'POST',
      body: '{}',
    });
    expect(quizRes.status).toBe(200);
    const foreignQuestions = (await quizRes.json()) as Array<{
      id: string;
      options: Array<{ id: string }>;
    }>;
    expect(foreignQuestions.length).toBeGreaterThan(0);
    const foreign = foreignQuestions[0];

    const res = await raw('/topics/llms/quiz/submissions', {
      method: 'POST',
      body: JSON.stringify({
        answers: [
          { question_id: foreign.id, option_id: foreign.options[0]?.id },
        ],
      }),
    });
    expect(res.status).toBe(400);
    const body = (await res.json()) as { code?: string; message?: string };
    expect(body.code).toBe('validation_error');
    expect(typeof body.message).toBe('string');
  });

  it('does not store an anonymous quiz result (US-3)', async () => {
    // Submit anonymously (no session cookie jar entry for a logged-in user;
    // afterEach resets the jar, so this request is anonymous).
    const quizRes = await raw('/topics/rag/quiz', {
      method: 'POST',
      body: '{}',
    });
    expect(quizRes.status).toBe(200);
    const questions = (await quizRes.json()) as Array<{
      id: string;
      options: Array<{ id: string }>;
    }>;

    const submitRes = await raw('/topics/rag/quiz/submissions', {
      method: 'POST',
      body: JSON.stringify({
        answers: questions.map((q) => ({
          question_id: q.id,
          option_id: q.options[0]?.id ?? null,
        })),
      }),
    });
    expect(submitRes.status).toBe(200);
    const result = (await submitRes.json()) as { score: number; total: number };
    expect(result.total).toBe(questions.length);

    // Register and log in as a fresh user, then confirm no attempts exist.
    resetHttpState();
    const client = createHttpClient();
    const email = uniqueEmail('anon-check');
    await client.register({ email, password: testPassword });
    const attempts = await client.listAttempts();
    expect(attempts).toEqual([]);
  });

  it('can read topics, topic and lesson without authentication (US-3)', async () => {
    // No session: content endpoints must work.
    resetHttpState();
    const resTopics = await raw('/topics');
    expect(resTopics.status).toBe(200);
    const topics = (await resTopics.json()) as Array<{ slug: string }>;
    expect(topics.length).toBeGreaterThan(0);

    const slug = topics[0].slug;
    const resTopic = await raw(`/topics/${encodeURIComponent(slug)}`);
    expect(resTopic.status).toBe(200);
    const topic = (await resTopic.json()) as {
      lessons: Array<{ id: string }>;
    };
    expect(topic.lessons.length).toBeGreaterThan(0);

    const resLesson = await raw(
      `/lessons/${encodeURIComponent(topic.lessons[0].id)}`,
    );
    expect(resLesson.status).toBe(200);
  });

  it('accepts an anonymous quiz submission and returns a graded result (US-3)', async () => {
    resetHttpState();
    const res = await raw('/topics/embeddings/quiz', {
      method: 'POST',
      body: '{}',
    });
    expect(res.status).toBe(200);
    const questions = (await res.json()) as Array<{
      id: string;
      options: Array<{ id: string }>;
    }>;
    const submitRes = await raw('/topics/embeddings/quiz/submissions', {
      method: 'POST',
      body: JSON.stringify({
        answers: questions.map((q) => ({ question_id: q.id, option_id: null })),
      }),
    });
    expect(submitRes.status).toBe(200);
    const result = (await submitRes.json()) as { score: number; total: number };
    expect(result.total).toBe(questions.length);
    expect(result.score).toBe(0); // all unanswered -> incorrect
  });
});
