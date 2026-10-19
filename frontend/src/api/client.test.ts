import { describe, expect, it } from 'vitest';
import { API_MODE, api } from '@/api';
import { mockClient } from '@/api/mock';

describe('the exported client', () => {
  it('is selected by one value and no screen can name the implementation', () => {
    expect(API_MODE).toBe('mock');
    expect(api).toBe(mockClient);
  });

  it('exposes exactly the 13 operations and nothing else', () => {
    expect(Object.keys(api).sort()).toEqual([
      'getAttempt',
      'getLesson',
      'getProgress',
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
