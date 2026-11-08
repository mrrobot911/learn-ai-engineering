import { describe, expect, it } from 'vitest';
import { API_MODE, api } from '@/api';
import { mockClient } from '@/api/mock';

describe('the exported client', () => {
  it('defaults to mock mode in tests', () => {
    expect(API_MODE).toBe('mock');
    expect(api).toBe(mockClient);
  });

  it('exposes exactly the 14 operations and nothing else', () => {
    expect(Object.keys(api).sort()).toEqual([
      'getAttempt',
      'getLesson',
      'getProgress',
      'getSession',
      'getTopic',
      'listAttempts',
      'listTopics',
      'login',
      'logout',
      'markLessonComplete',
      'register',
      'startQuiz',
      'submitQuiz',
      'undoLessonComplete',
    ]);
  });
});
