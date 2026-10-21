import { beforeEach, describe, expect, it } from 'vitest';
import { getQuestionsForTopic, seedQuizHistory, topics } from '@/data/content';
import { mockClient, resetMockState } from './mock';
import type { QuizQuestion } from './types';

const TOPIC_ID = 'top-llms';
const DEMO = { email: 'demo@forge.dev', password: 'password123' };

// The fixture question and the client question for the same id.
const fixtures = new Map(getQuestionsForTopic(TOPIC_ID).map((q) => [q.id, q]));

async function startAndAnswerFirstOption(): Promise<QuizQuestion[]> {
  const questions = await mockClient.startQuiz(TOPIC_ID);
  await mockClient.submitQuiz({
    topicId: TOPIC_ID,
    answers: questions.map((question) => ({
      questionId: question.id,
      optionId: question.options[0].id,
    })),
  });
  return questions;
}

beforeEach(() => {
  resetMockState();
});

describe('content operations', () => {
  it('listTopics resolves with the fixture topics', async () => {
    expect((await mockClient.listTopics()).map((topic) => topic.id)).toEqual(
      topics.map((topic) => topic.id),
    );
  });

  it('getTopic resolves with the fixture topic and its lessons', async () => {
    const topic = await mockClient.getTopic(TOPIC_ID);
    expect(topic.id).toBe(TOPIC_ID);
    expect(topic.lessons).toEqual(topics[0].lessons);
  });

  it('getLesson resolves with the fixture lesson', async () => {
    const lesson = await mockClient.getLesson(topics[0].lessons[0].id);
    expect(lesson).toEqual(topics[0].lessons[0]);
  });

  it('getLesson rejects an unknown id with the 404 not-found signal', async () => {
    await expect(mockClient.getLesson('les-does-not-exist')).rejects.toMatchObject({ status: 404 });
  });
});

describe('quiz operations', () => {
  it('startQuiz resolves with 5 questions from the topic', async () => {
    const questions = await mockClient.startQuiz(TOPIC_ID);
    expect(questions).toHaveLength(5);
    for (const question of questions) {
      expect(fixtures.has(question.id)).toBe(true);
    }
  });

  it('startQuiz never sends the correct option or the explanation', async () => {
    const questions = await mockClient.startQuiz(TOPIC_ID);
    for (const question of questions) {
      expect(Object.keys(question).sort()).toEqual(['id', 'options', 'text']);
      for (const option of question.options) {
        expect(Object.keys(option).sort()).toEqual(['id', 'text']);
      }
    }
    expect(JSON.stringify(questions)).not.toMatch(
      /"(isCorrect|correct|correctOptionId|correctOptionText|explanation)":/,
    );
  });

  it('submitQuiz grades the submission from the fixture answers', async () => {
    const questions = await mockClient.startQuiz(TOPIC_ID);
    const result = await mockClient.submitQuiz({
      topicId: TOPIC_ID,
      answers: questions.map((question) => ({
        questionId: question.id,
        optionId: fixtures.get(question.id)!.options.find((option) => option.isCorrect)!.id,
      })),
    });
    expect(result).toMatchObject({ score: 5, total: 5 });
    for (const [index, answer] of result.answers.entries()) {
      expect(answer.isCorrect).toBe(true);
      expect(answer.questionText).toBe(questions[index].text);
      expect(answer.correctOptionText).not.toBe('');
      expect(answer.explanation).toBe(fixtures.get(answer.questionId)!.explanation);
    }
  });

  it('submitQuiz counts an unanswered question as incorrect', async () => {
    const questions = await mockClient.startQuiz(TOPIC_ID);
    const result = await mockClient.submitQuiz({
      topicId: TOPIC_ID,
      answers: [{ questionId: questions[0].id, optionId: null }],
    });
    expect(result).toMatchObject({ score: 0, total: 1 });
    expect(result.answers[0]).toMatchObject({
      selectedOptionId: null,
      selectedOptionText: '',
      isCorrect: false,
    });
    expect(result.answers[0].correctOptionText).not.toBe('');
  });

  it('submitQuiz rejects a question that does not belong to the topic', async () => {
    const other = getQuestionsForTopic('top-rag')[0];
    await expect(
      mockClient.submitQuiz({
        topicId: TOPIC_ID,
        answers: [{ questionId: other.id, optionId: other.options[0].id }],
      }),
    ).rejects.toThrow();
  });

  it('submitQuiz returns an anonymous result without storing it', async () => {
    await startAndAnswerFirstOption();
    await mockClient.login(DEMO);
    expect(await mockClient.listAttempts()).toHaveLength(seedQuizHistory.length);
  });
});

describe('account operations', () => {
  it('register returns the new account and signs in', async () => {
    const user = await mockClient.register({ email: 'New@Forge.dev', password: 'password123' });
    expect(user).toMatchObject({ email: 'new@forge.dev', isMaintainer: false });
    await expect(mockClient.listAttempts()).resolves.toEqual([]);
  });

  it('login signs in and exposes the seeded history of that account', async () => {
    const user = await mockClient.login(DEMO);
    expect(user).toMatchObject({ id: 'usr-demo', email: 'demo@forge.dev' });
    expect(await mockClient.listAttempts()).toHaveLength(seedQuizHistory.length);
  });

  it('logout ends the session', async () => {
    await mockClient.login(DEMO);
    await mockClient.logout();
    await expect(mockClient.getProgress()).rejects.toThrow();
  });
});

describe('progress operations', () => {
  it('getProgress resolves with the completed lessons', async () => {
    await mockClient.login(DEMO);
    expect(await mockClient.getProgress()).toEqual([]);
  });

  it('markLessonComplete is reported by getProgress', async () => {
    await mockClient.login(DEMO);
    await mockClient.markLessonComplete('les-llm-1');
    expect(await mockClient.getProgress()).toEqual(['les-llm-1']);
  });

  it('undoLessonComplete removes the lesson from getProgress', async () => {
    await mockClient.login(DEMO);
    await mockClient.markLessonComplete('les-llm-1');
    await mockClient.undoLessonComplete('les-llm-1');
    expect(await mockClient.getProgress()).toEqual([]);
  });
});

describe('attempt operations', () => {
  it('listAttempts resolves with the seeded attempts, newest first', async () => {
    await mockClient.login(DEMO);
    const [first] = await mockClient.listAttempts();
    expect(first.id).toBe('att-seed-3');
    expect(Object.keys(first).sort()).toEqual([
      'createdAt',
      'id',
      'score',
      'topicId',
      'topicTitle',
      'total',
    ]);
  });

  it('listAttempts gives every summary an ISO-8601 createdAt', async () => {
    await mockClient.login(DEMO);
    const list = await mockClient.listAttempts();
    expect(list).toHaveLength(seedQuizHistory.length);
    for (const summary of list) {
      expect(summary.createdAt).toMatch(/^\d{4}-\d{2}-\d{2}T/);
    }
  });

  it('getAttempt resolves with one seeded attempt and its answers', async () => {
    await mockClient.login(DEMO);
    const attempt = await mockClient.getAttempt('att-seed-1');
    expect(attempt).toMatchObject({ topicId: TOPIC_ID, score: 4, total: 5 });
    expect(attempt.answers).toHaveLength(5);
  });

  it('a submitted quiz appears in listAttempts', async () => {
    await mockClient.login(DEMO);
    await startAndAnswerFirstOption();
    const list = await mockClient.listAttempts();
    expect(list).toHaveLength(seedQuizHistory.length + 1);
    const stored = await mockClient.getAttempt(list[0].id);
    expect(stored.answers).toHaveLength(5);
  });

  it('a registered account only sees its own attempts', async () => {
    await mockClient.login(DEMO);
    await startAndAnswerFirstOption();
    await mockClient.logout();
    await mockClient.register({ email: 'fresh@forge.dev', password: 'password123' });
    expect(await mockClient.listAttempts()).toEqual([]);
    await expect(mockClient.getAttempt('att-seed-1')).rejects.toThrow();
  });
});
