import { describe, expect, it } from 'vitest';
import { allQuestions, seedQuizHistory } from './content';

// US-6: an attempt keeps the wording of the moment of the attempt, so a later
// edit of the question does not change a past attempt.
describe('the seeded quiz history', () => {
  it('stores the wording of the attempt, not the current wording of the question', () => {
    const attempt = seedQuizHistory.find((seeded) => seeded.id === 'att-seed-1');
    const answer = attempt?.answers.find((stored) => stored.questionId === 'q-llm-1');
    const question = allQuestions.find((current) => current.id === 'q-llm-1');
    const correct = question?.options.find((option) => option.isCorrect);

    expect(answer?.questionText).toBeDefined();
    expect(question?.text).toBeDefined();
    expect(correct?.text).toBeDefined();

    expect(answer?.questionText).not.toBe(question?.text);
    expect(answer?.correctOptionText).not.toBe(correct?.text);
  });
});
