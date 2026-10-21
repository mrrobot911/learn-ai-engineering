// @vitest-environment jsdom
// The quiz history across the screens and the client: the mock stands in for
// the backend, so the flow runs from the list of attempts to a stored attempt.
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { useEffect } from 'react';
import { cleanup, render, screen, within } from '@testing-library/react';
import { mockClient, resetMockState } from '@/api/mock';
import { SessionProvider, useSession } from '@/session';
import { HistoryDetailPage, HistoryPage } from '@/pages/HistoryPages';
import { allQuestions, seedQuizHistory } from '@/data/content';

/** The seeded account of the mock, logged in when the page mounts. */
function LoginOnMount() {
  const { login } = useSession();
  useEffect(() => { void login('demo@forge.dev', 'password123'); }, [login]);
  return null;
}

beforeEach(() => {
  resetMockState();
});

afterEach(cleanup);

describe('quiz history', () => {
  it('lists the seeded attempts newest first and opens the stored attempt', async () => {
    render(
      <SessionProvider>
        <LoginOnMount />
        <HistoryPage />
      </SessionProvider>,
    );

    const rows = await screen.findAllByRole('listitem');
    expect(rows).toHaveLength(seedQuizHistory.length);

    const links = screen
      .getAllByRole('link')
      .filter((link) => link.getAttribute('href')?.startsWith('#/history/'));
    expect(links.map((link) => link.getAttribute('href'))).toEqual([
      '#/history/att-seed-3',
      '#/history/att-seed-2',
      '#/history/att-seed-1',
    ]);
    expect(within(rows[0]).getByText('2026-10-15')).toBeInTheDocument();
    expect(within(rows[0]).getByText('3/5')).toBeInTheDocument();

    cleanup();
    render(
      <SessionProvider>
        <LoginOnMount />
        <HistoryDetailPage attemptId="att-seed-1" />
      </SessionProvider>,
    );

    const attempt = seedQuizHistory.find((seeded) => seeded.id === 'att-seed-1');
    const stored = attempt?.answers.find((answer) => answer.questionId === 'q-llm-1');
    const question = allQuestions.find((current) => current.id === 'q-llm-1');
    const currentCorrect = question?.options.find((option) => option.isCorrect);

    expect(await screen.findByText(stored?.questionText ?? '')).toBeInTheDocument();
    expect(screen.getAllByText(stored?.correctOptionText ?? '')).toHaveLength(2);
    expect(screen.queryByText(question?.text ?? '')).not.toBeInTheDocument();
    expect(screen.queryByText(currentCorrect?.text ?? '')).not.toBeInTheDocument();
  });

  it('shows only the attempts of the signed-in account', async () => {
    await mockClient.register({ email: 'other@forge.dev', password: 'password123' });

    expect(await mockClient.listAttempts()).toEqual([]);
    await expect(mockClient.getAttempt('att-seed-1')).rejects.toThrow();
  });
});
