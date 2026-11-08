import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { ApiError } from './errors';
import { createHttpClient, getCsrfToken, resetHttpState } from './http';

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}



describe('http client', () => {
  const fetchMock = vi.fn<typeof fetch>();

  beforeEach(() => {
    resetHttpState();
    fetchMock.mockReset();
    vi.stubGlobal('fetch', fetchMock);
    vi.stubEnv('VITE_API_BASE_URL', 'https://api.test');
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.unstubAllEnvs();
  });

  it('maps path/method and sends cookies for topic and lesson reads', async () => {
    const client = createHttpClient();
    fetchMock.mockResolvedValueOnce(
      jsonResponse([{ id: 't1', slug: 'llms', title: 'LLMs', description: 'd', position: 1, status: 'published' }]),
    );
    await client.listTopics();
    expect(fetchMock).toHaveBeenCalledWith(
      'https://api.test/topics',
      expect.objectContaining({ method: 'GET', credentials: 'include' }),
    );

    fetchMock.mockResolvedValueOnce(
      jsonResponse({
        id: 't1', slug: 'llms', title: 'LLMs', description: 'd', position: 1, status: 'published',
        lessons: [{ id: 'l1', topic_id: 't1', title: 'L', body: 'Body text here.', position: 1, status: 'published' }],
      }),
    );
    const topic = await client.getTopic('llms');
    expect(fetchMock).toHaveBeenLastCalledWith(
      'https://api.test/topics/llms',
      expect.objectContaining({ method: 'GET', credentials: 'include' }),
    );
    expect(topic).toMatchObject({ slug: 'llms', icon: 'BrainCircuit', color: 'brand' });
    expect(topic.lessons[0]).toMatchObject({ topicId: 't1', summary: 'Body text here.', body: 'Body text here.' });

    fetchMock.mockResolvedValueOnce(
      jsonResponse({ id: 'l1', topic_id: 't1', title: 'L', body: '', position: 1, status: 'published' }),
    );
    fetchMock.mockResolvedValueOnce(
      jsonResponse([{ id: 't1', slug: 'llms', title: 'LLMs', description: 'd', position: 1 }]),
    );
    const lesson = await client.getLesson('l1');
    expect(fetchMock).toHaveBeenNthCalledWith(
      3,
      'https://api.test/lessons/l1',
      expect.objectContaining({ method: 'GET', credentials: 'include' }),
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      4,
      'https://api.test/topics',
      expect.objectContaining({ method: 'GET', credentials: 'include' }),
    );
    expect(lesson).toMatchObject({ topicId: 't1', topicSlug: 'llms', summary: '', durationMin: 0 });
  });

  it('maps quiz operations to slug paths and converts submission body to snake_case', async () => {
    const client = createHttpClient();
    fetchMock.mockResolvedValueOnce(
      jsonResponse([{ id: 'q1', text: 'Q?', options: [{ id: 'o1', text: 'A' }] }]),
    );
    const questions = await client.startQuiz('llms');
    expect(fetchMock).toHaveBeenCalledWith(
      'https://api.test/topics/llms/quiz',
      expect.objectContaining({ method: 'POST', credentials: 'include' }),
    );
    expect(questions[0]).toEqual({ id: 'q1', text: 'Q?', options: [{ id: 'o1', text: 'A' }] });

    fetchMock.mockResolvedValueOnce(
      jsonResponse({
        score: 1,
        total: 1,
        answers: [{ question_id: 'q1', is_correct: true, correct_option_text: 'A', explanation: 'Because.' }],
      }),
    );
    const result = await client.submitQuiz({
      topicSlug: 'llms',
      answers: [{ questionId: 'q1', optionId: 'o1' }],
    });
    const submitCall = fetchMock.mock.calls[1][1] as RequestInit;
    expect(fetchMock.mock.calls[1][0]).toBe('https://api.test/topics/llms/quiz/submissions');
    expect(submitCall.method).toBe('POST');
    expect(JSON.parse(String(submitCall.body))).toEqual({
      answers: [{ question_id: 'q1', option_id: 'o1' }],
    });
    expect(result.answers[0]).toMatchObject({
      questionId: 'q1',
      questionText: 'Q?',
      selectedOptionId: 'o1',
      selectedOptionText: 'A',
      correctOptionId: 'o1',
      correctOptionText: 'A',
      isCorrect: true,
      explanation: 'Because.',
    });
  });

  it('attaches the CSRF header to submitQuiz when a session token is present', async () => {
    const client = createHttpClient();
    fetchMock.mockResolvedValueOnce(
      jsonResponse({ user: { id: 'u1', email: 'a@b.c' }, csrf_token: 'csrf-1' }),
    );
    await client.login({ email: 'a@b.c', password: 'secret' });

    fetchMock.mockResolvedValueOnce(
      jsonResponse([{ id: 'q1', text: 'Q?', options: [{ id: 'o1', text: 'A' }] }]),
    );
    await client.startQuiz('llms');

    fetchMock.mockResolvedValueOnce(
      jsonResponse({
        score: 0,
        total: 1,
        answers: [{ question_id: 'q1', is_correct: false, correct_option_text: 'A', explanation: 'Because.' }],
      }),
    );
    await client.submitQuiz({
      topicSlug: 'llms',
      answers: [{ questionId: 'q1', optionId: null }],
    });

    const submitCall = fetchMock.mock.calls[2][1] as RequestInit;
    expect(submitCall.headers).toMatchObject({ 'X-CSRF-Token': 'csrf-1' });
  });

  it('sends no CSRF header for an anonymous submitQuiz and still returns the graded result', async () => {
    const client = createHttpClient();
    fetchMock.mockResolvedValueOnce(
      jsonResponse([{ id: 'q1', text: 'Q?', options: [{ id: 'o1', text: 'A' }] }]),
    );
    await client.startQuiz('llms');

    fetchMock.mockResolvedValueOnce(
      jsonResponse({
        score: 1,
        total: 1,
        answers: [{ question_id: 'q1', is_correct: true, correct_option_text: 'A', explanation: 'Because.' }],
      }),
    );
    const result = await client.submitQuiz({
      topicSlug: 'llms',
      answers: [{ questionId: 'q1', optionId: 'o1' }],
    });

    const submitCall = fetchMock.mock.calls[1][1] as RequestInit;
    expect((submitCall.headers as Record<string, string>)['X-CSRF-Token']).toBeUndefined();
    expect(result).toMatchObject({ score: 1, total: 1 });
  });

  it('keeps CSRF token in memory from auth/session and sends X-CSRF-Token on state-changing requests', async () => {
    const client = createHttpClient();
    fetchMock.mockResolvedValueOnce(
      jsonResponse({ user: { id: 'u1', email: 'a@b.c' }, csrf_token: 'csrf-1' }),
    );
    await client.login({ email: 'a@b.c', password: 'secret' });
    expect(getCsrfToken()).toBe('csrf-1');
    expect(fetchMock.mock.calls[0][0]).toBe('https://api.test/auth/login');

    fetchMock.mockResolvedValueOnce(jsonResponse(['l1']));
    await client.getProgress();
    const progressHeaders = (fetchMock.mock.calls[1][1] as RequestInit).headers as Record<string, string>;
    expect(progressHeaders['X-CSRF-Token']).toBeUndefined();

    fetchMock.mockResolvedValueOnce(jsonResponse(['l1', 'l2']));
    await client.markLessonComplete('l2');
    const putHeaders = (fetchMock.mock.calls[2][1] as RequestInit).headers as Record<string, string>;
    expect(putHeaders['X-CSRF-Token']).toBe('csrf-1');
    expect(fetchMock.mock.calls[2][0]).toBe('https://api.test/me/progress/lessons/l2');
    expect((fetchMock.mock.calls[2][1] as RequestInit).method).toBe('PUT');

    fetchMock.mockResolvedValueOnce(jsonResponse({ user: { id: 'u1', email: 'a@b.c' }, csrf_token: 'csrf-2' }));
    await client.getSession();
    expect(getCsrfToken()).toBe('csrf-2');
    expect(fetchMock.mock.calls[3][0]).toBe('https://api.test/auth/session');

    fetchMock.mockResolvedValueOnce(jsonResponse(['l1']));
    await client.undoLessonComplete('l2');
    const delHeaders = (fetchMock.mock.calls[4][1] as RequestInit).headers as Record<string, string>;
    expect(delHeaders['X-CSRF-Token']).toBe('csrf-2');
    expect((fetchMock.mock.calls[4][1] as RequestInit).method).toBe('DELETE');

    fetchMock.mockResolvedValueOnce(jsonResponse({ ok: true }));
    await client.logout();
    expect((fetchMock.mock.calls[5][1] as RequestInit).headers).toMatchObject({ 'X-CSRF-Token': 'csrf-2' });
    expect(getCsrfToken()).toBeNull();
  });

  it('maps HTTP failures to ApiError with status and contract code', async () => {
    const client = createHttpClient();
    const cases: Array<[number, string, string]> = [
      [400, 'validation_error', 'The submission contains a question that does not belong to the topic.'],
      [401, 'authentication_error', 'No session'],
      [403, 'forbidden', 'CSRF failed'],
      [404, 'not_found', 'Missing'],
      [409, 'conflict', 'Duplicate'],
      [500, 'internal_error', 'Boom'],
    ];
    for (const [status, code, message] of cases) {
      fetchMock.mockResolvedValueOnce(jsonResponse({ code, message }, status));
      await expect(client.getSession()).rejects.toSatisfy((err: unknown) => {
        return err instanceof ApiError && err.status === status && err.code === code && err.message === message;
      });
    }
  });

  it('maps network failures to a typed error', async () => {
    const client = createHttpClient();
    fetchMock.mockRejectedValueOnce(new TypeError('fetch failed'));
    await expect(client.listTopics()).rejects.toMatchObject({
      status: 0,
      code: 'network_error',
    });
  });

  it('maps progress and attempt endpoints with snake_case conversion', async () => {
    const client = createHttpClient();
    fetchMock.mockResolvedValueOnce(
      jsonResponse([{ id: 't1', slug: 'llms', title: 'LLMs', description: 'd', position: 1, status: 'published' }]),
    );
    await client.listTopics();

    fetchMock.mockResolvedValueOnce(
      jsonResponse([{ id: 'a1', topic_id: 't1', topic_slug: 'llms', score: 3, total: 5, created_at: '2026-01-01T00:00:00Z' }]),
    );
    const attempts = await client.listAttempts();
    expect(fetchMock.mock.calls[1][0]).toBe('https://api.test/me/attempts');
    expect(attempts[0]).toMatchObject({ id: 'a1', topicId: 't1', topicTitle: 'LLMs', createdAt: '2026-01-01T00:00:00Z' });

    fetchMock.mockResolvedValueOnce(
      jsonResponse({
        id: 'a1', topic_id: 't1', topic_slug: 'llms', score: 3, total: 5, created_at: '2026-01-01T00:00:00Z',
        answers: [{ question_id: null, question_text: 'Q', selected_option_text: '', correct_option_text: 'A', is_correct: false }],
      }),
    );
    const attempt = await client.getAttempt('a1');
    expect(fetchMock.mock.calls[2][0]).toBe('https://api.test/me/attempts/a1');
    expect(attempt.answers[0]).toMatchObject({
      questionId: '',
      questionText: 'Q',
      selectedOptionText: '',
      correctOptionText: 'A',
      isCorrect: false,
    });
  });
});
