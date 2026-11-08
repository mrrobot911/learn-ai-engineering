// @vitest-environment jsdom
/**
 * Frontend integration tests against a real backend and seeded test database.
 * Covers US-4 (accounts), US-5 (progress), US-6 (quiz history). Issue #86.
 * Run with: bun run test:integration (requires the test backend; see README).
 */
import { beforeAll, describe, expect, it } from 'vitest';
import { ApiError } from '@/api/errors';
import { createHttpClient, resetHttpState } from '@/api/http';
import {
  apiBaseUrl,
  resetCookieJar,
  testPassword,
  uniqueEmail,
} from './harness';

const base = apiBaseUrl();

async function getSeededLessonId(): Promise<string> {
  const client = createHttpClient();
  const topic = await client.getTopic('llms');
  expect(topic.lessons.length).toBeGreaterThan(0);
  return topic.lessons[0].id;
}

describe('accounts, progress and history (US-4, US-5, US-6)', () => {
  beforeAll(async () => {
    const res = await fetch(`${base}/health`).catch((e) => e);
    if (res instanceof Error || !res.ok) {
      throw new Error(
        `Test backend not reachable at ${base}. Start it with the documented command (see README "Frontend integration tests").`,
      );
    }
  });

  it('registers, returns current session, logs in and logs out (US-4)', async () => {
    const client = createHttpClient();
    const email = uniqueEmail('acct');

    const registered = await client.register({ email, password: testPassword });
    expect(registered.email).toBe(email);
    expect(registered.id).toBeTruthy();

    const session = await client.getSession();
    expect(session).toMatchObject({ id: registered.id, email });

    await client.logout();
    await expect(client.getSession()).rejects.toMatchObject({
      status: 401,
      code: 'authentication_error',
    } satisfies Partial<ApiError>);

    const loggedIn = await client.login({ email, password: testPassword });
    expect(loggedIn).toMatchObject({ id: registered.id, email });
    expect(await client.getSession()).toMatchObject({ email });

    await client.logout();
  });

  it('rejects a duplicate email on register (US-4)', async () => {
    const client = createHttpClient();
    const email = uniqueEmail('dup');
    await client.register({ email, password: testPassword });
    await client.logout();
    resetHttpState();

    await expect(
      client.register({ email, password: testPassword }),
    ).rejects.toMatchObject({
      status: 409,
    } satisfies Partial<ApiError>);
  });

  it('refreshes CSRF and retries a state-changing request (US-4)', async () => {
    const client = createHttpClient();
    const email = uniqueEmail('csrf');
    await client.register({ email, password: testPassword });
    const lessonId = await getSeededLessonId();

    // Simulate a reload: in-memory CSRF token is lost, session cookie remains.
    resetHttpState();
    await expect(client.markLessonComplete(lessonId)).resolves.toBeUndefined();

    const progress = await client.getProgress();
    expect(progress).toContain(lessonId);

    resetHttpState();
    await client.logout();
    await expect(client.getSession()).rejects.toMatchObject({
      status: 401,
    } satisfies Partial<ApiError>);
  });

  it('progress mark and unmark survive a reload and are isolated between users (US-5)', async () => {
    const lessonId = await getSeededLessonId();
    const clientA = createHttpClient();
    const emailA = uniqueEmail('prog-a');
    await clientA.register({ email: emailA, password: testPassword });

    await clientA.markLessonComplete(lessonId);
    expect(await clientA.getProgress()).toContain(lessonId);

    // Reload: keep cookie jar (session) but drop in-memory client state.
    resetHttpState();
    expect(await clientA.getProgress()).toContain(lessonId);

    await clientA.undoLessonComplete(lessonId);
    resetHttpState();
    expect(await clientA.getProgress()).not.toContain(lessonId);

    // Mark again to prove persistence for user A across user switch.
    await clientA.markLessonComplete(lessonId);
    expect(await clientA.getProgress()).toContain(lessonId);

    // Switch to user B: clear session/cookies and client state.
    await clientA.logout();
    resetCookieJar();
    resetHttpState();

    const clientB = createHttpClient();
    const emailB = uniqueEmail('prog-b');
    await clientB.register({ email: emailB, password: testPassword });
    const progressB = await clientB.getProgress();
    expect(progressB).not.toContain(lessonId);

    await clientB.markLessonComplete(lessonId);
    expect(await clientB.getProgress()).toContain(lessonId);

    // Back to user A: A still has their own progress.
    await clientB.logout();
    resetCookieJar();
    resetHttpState();
    const clientA2 = createHttpClient();
    await clientA2.login({ email: emailA, password: testPassword });
    expect(await clientA2.getProgress()).toContain(lessonId);
  });

  it('stores a logged-in quiz attempt with snapshot; another user gets 404 (US-6)', async () => {
    // User A submits a quiz while logged in.
    resetHttpState();
    const clientA = createHttpClient();
    const emailA = uniqueEmail('hist-a');
    await clientA.register({ email: emailA, password: testPassword });

    const questions = await clientA.startQuiz('ai-agents');
    expect(questions.length).toBeGreaterThan(0);
    const answers = questions.map((q) => ({
      questionId: q.id,
      optionId: q.options[0]?.id ?? null,
    }));
    const result = await clientA.submitQuiz({
      topicSlug: 'ai-agents',
      answers,
    });
    expect(result.total).toBe(questions.length);

    const list = await clientA.listAttempts();
    expect(list.length).toBeGreaterThan(0);
    const mine = list.find(
      (a) =>
        a.topicId &&
        a.score === result.score &&
        a.total === result.total &&
        typeof a.createdAt === 'string',
    );
    expect(mine).toBeDefined();
    expect(mine?.topicTitle).toBeTruthy();

    const detail = await clientA.getAttempt(mine!.id);
    expect(detail).toMatchObject({
      id: mine!.id,
      score: result.score,
      total: result.total,
    });
    expect(detail.answers).toHaveLength(result.total);
    for (const a of detail.answers) {
      expect(typeof a.questionText).toBe('string');
      expect(a.questionText.length).toBeGreaterThan(0);
      expect(typeof a.correctOptionText).toBe('string');
      expect(a.correctOptionText.length).toBeGreaterThan(0);
      expect(typeof a.selectedOptionText).toBe('string');
      expect(typeof a.isCorrect).toBe('boolean');
    }

    const attemptId = detail.id;

    // User B must not access user A's attempt.
    await clientA.logout();
    resetCookieJar();
    resetHttpState();
    const clientB = createHttpClient();
    await clientB.register({
      email: uniqueEmail('hist-b'),
      password: testPassword,
    });
    await expect(clientB.getAttempt(attemptId)).rejects.toMatchObject({
      status: 404,
    } satisfies Partial<ApiError>);
  });
});
