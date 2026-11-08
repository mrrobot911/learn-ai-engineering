import { describe, expect, it } from 'vitest';
import {
  READING_WPM,
  camelToSnake,
  durationMinFromBody,
  mapLesson,
  mapTopic,
  mapTopicSummary,
  presentationForSlug,
  snakeToCamel,
  summaryFromBody,
} from './mappers';

describe('snake/camel conversion', () => {
  it('maps snake_case keys to camelCase recursively', () => {
    expect(
      snakeToCamel({
        topic_id: 't1',
        created_at: '2026-01-01T00:00:00Z',
        answers: [{ question_id: 'q1', is_correct: true }],
      }),
    ).toEqual({
      topicId: 't1',
      createdAt: '2026-01-01T00:00:00Z',
      answers: [{ questionId: 'q1', isCorrect: true }],
    });
  });

  it('maps camelCase keys to snake_case recursively', () => {
    expect(
      camelToSnake({
        topicSlug: 'llms',
        answers: [{ questionId: 'q1', optionId: null }],
      }),
    ).toEqual({
      topic_slug: 'llms',
      answers: [{ question_id: 'q1', option_id: null }],
    });
  });

  it('preserves arrays and primitives', () => {
    expect(snakeToCamel(['a', 1, null])).toEqual(['a', 1, null]);
    expect(camelToSnake('x')).toBe('x');
  });
});

describe('presentation-field mapper', () => {
  it('maps icon and color from the topic slug', () => {
    expect(presentationForSlug('llms')).toEqual({ icon: 'BrainCircuit', color: 'brand' });
    expect(presentationForSlug('rag')).toEqual({ icon: 'Search', color: 'brand' });
    expect(presentationForSlug('agents')).toEqual({ icon: 'Bot', color: 'accent' });
  });

  it('falls back for an unknown slug', () => {
    expect(presentationForSlug('unknown-topic')).toEqual({ icon: 'BookOpen', color: 'brand' });
  });

  it('derives summary from the first body paragraph', () => {
    const body = '# Title\n\nFirst paragraph with useful text.\n\nSecond paragraph.';
    expect(summaryFromBody(body)).toBe('First paragraph with useful text.');
  });

  it('falls back without error for an empty body', () => {
    expect(summaryFromBody('')).toBe('');
    expect(summaryFromBody('   \n\n  ')).toBe('');
    expect(durationMinFromBody('')).toBe(0);
  });

  it('computes durationMin from body word count at the documented rate', () => {
    const words = Array.from({ length: READING_WPM * 2 + 1 }, () => 'word').join(' ');
    expect(durationMinFromBody(words)).toBe(3);
    expect(durationMinFromBody('short body here')).toBe(1);
  });

  it('maps contract topic/lesson fields into screen types', () => {
    const topic = mapTopic({
      id: 't1',
      slug: 'llms',
      title: 'LLMs',
      description: 'desc',
      position: 1,
      lessons: [
        { id: 'l1', topicId: 't1', title: 'Lesson', body: 'Hello world content.', position: 1 },
      ],
    });
    expect(topic).toMatchObject({
      id: 't1',
      slug: 'llms',
      icon: 'BrainCircuit',
      color: 'brand',
      lessons: [{ id: 'l1', topicId: 't1', summary: 'Hello world content.', durationMin: 1 }],
    });

    const summary = mapTopicSummary({ id: 't2', slug: 'mcp', title: 'MCP', description: 'd', position: 2 });
    expect(summary).toMatchObject({ icon: 'Plug', color: 'warning' });

    const lesson = mapLesson({ id: 'l2', topicId: 't1', title: 'Empty', body: '', position: 2 });
    expect(lesson).toMatchObject({ summary: '', durationMin: 0, body: '' });
  });
});
