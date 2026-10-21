// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { api } from '@/api';
import type { QuizAttempt, QuizAttemptSummary, User } from '@/api';
import { useSession } from '@/session';
import { HistoryDetailPage, HistoryPage } from './HistoryPages';

vi.mock('@/api', () => ({
  api: {
    listAttempts: vi.fn(),
    getAttempt: vi.fn(),
  },
}));

vi.mock('@/session', () => ({
  useSession: vi.fn(),
}));

const listAttempts = vi.mocked(api.listAttempts);
const getAttempt = vi.mocked(api.getAttempt);
const useSessionMock = vi.mocked(useSession);

const loggedInUser: User = {
  id: 'u1',
  email: 'user@example.com',
  isMaintainer: false,
  createdAt: '2026-01-01T00:00:00.000Z',
};

/** The visitor of the page: anonymous by default, logged in on demand. */
function setSession(user: User | null) {
  useSessionMock.mockReturnValue({
    user,
    authReady: true,
    login: vi.fn(),
    register: vi.fn(),
    logout: vi.fn(),
  });
}

const summaryOne: QuizAttemptSummary = {
  id: 'att-1',
  topicId: 'top-rag',
  topicTitle: 'Retrieval-Augmented Generation',
  score: 3,
  total: 5,
  createdAt: '2026-10-15T09:00:00Z',
};
const summaryTwo: QuizAttemptSummary = {
  id: 'att-2',
  topicId: 'top-llms',
  topicTitle: 'Large Language Models',
  score: 4,
  total: 5,
  createdAt: '2026-10-14T09:00:00Z',
};
const summaryThree: QuizAttemptSummary = {
  id: 'att-3',
  topicId: 'top-agents',
  topicTitle: 'AI Agents',
  score: 5,
  total: 5,
  createdAt: '2026-10-12T09:00:00Z',
};

const attempt: QuizAttempt = {
  id: 'att-1',
  topicId: 'top-rag',
  topicTitle: 'Retrieval-Augmented Generation',
  score: 1,
  total: 3,
  createdAt: '2026-10-15T09:00:00Z',
  answers: [
    {
      questionId: 'q-1',
      questionText: 'What is RAG?',
      selectedOptionId: 'o-2',
      selectedOptionText: 'Retrieval-augmented generation',
      correctOptionId: 'o-2',
      correctOptionText: 'Retrieval-augmented generation',
      isCorrect: true,
      explanation: 'It grounds the answer in the retrieved context.',
    },
    {
      questionId: 'q-2',
      questionText: 'What breaks a naive chunking?',
      selectedOptionId: 'o-1',
      selectedOptionText: 'Fine-tuning every document',
      correctOptionId: 'o-2',
      correctOptionText: 'Chunks too small lose the surrounding context',
      isCorrect: false,
      explanation: 'Small chunks lose the context around the passage.',
    },
    {
      questionId: 'q-3',
      questionText: 'What is a token?',
      selectedOptionId: null,
      selectedOptionText: '',
      correctOptionId: 'o-4',
      correctOptionText: 'A unit of text the model reads',
      isCorrect: false,
      explanation: 'A token is a unit of text.',
    },
  ],
};

/** The links to an attempt, in the order the page shows them. */
function attemptLinks() {
  return screen
    .getAllByRole('link')
    .filter((link) => link.getAttribute('href')?.startsWith('#/history/'));
}

function deferred<T>() {
  let resolve!: (value: T) => void;
  let reject!: (reason?: unknown) => void;
  const promise = new Promise<T>((res, rej) => { resolve = res; reject = rej; });
  return { promise, resolve, reject };
}

beforeEach(() => {
  vi.resetAllMocks();
  setSession(null);
});

afterEach(cleanup);

describe('HistoryPage', () => {
  it('lists the attempts in the order the client resolved them', async () => {
    setSession(loggedInUser);
    listAttempts.mockResolvedValue([summaryOne, summaryTwo, summaryThree]);

    render(<HistoryPage />);

    expect(await screen.findByRole('heading', { name: 'Quiz history', level: 1 })).toBeInTheDocument();
    expect(listAttempts).toHaveBeenCalledTimes(1);
    expect(screen.getAllByRole('listitem')).toHaveLength(3);
    expect(attemptLinks().map((link) => link.getAttribute('href'))).toEqual([
      '#/history/att-1',
      '#/history/att-2',
      '#/history/att-3',
    ]);
  });

  it('shows the topic, the score and the UTC day of an attempt in its row', async () => {
    setSession(loggedInUser);
    listAttempts.mockResolvedValue([summaryOne]);

    render(<HistoryPage />);

    const row = await screen.findByRole('listitem');
    expect(within(row).getByText('Retrieval-Augmented Generation')).toBeInTheDocument();
    expect(within(row).getByText('3/5')).toBeInTheDocument();
    expect(within(row).getByText('2026-10-15')).toBeInTheDocument();
    expect(within(row).getByRole('link')).toHaveAttribute('href', '#/history/att-1');
  });

  it('shows the empty state when the client resolves no attempts', async () => {
    setSession(loggedInUser);
    listAttempts.mockResolvedValue([]);

    render(<HistoryPage />);

    expect(await screen.findByRole('heading', { name: 'No quiz attempts yet' })).toBeInTheDocument();
    expect(attemptLinks()).toHaveLength(0);
  });

  it('asks an anonymous visitor to log in and never reads the attempts', async () => {
    render(<HistoryPage />);

    expect(await screen.findByRole('link', { name: 'Log in' })).toHaveAttribute('href', '#/login');
    expect(screen.queryByRole('listitem')).not.toBeInTheDocument();
    expect(listAttempts).not.toHaveBeenCalled();
  });

  it('shows the loading state while the client call is pending', async () => {
    setSession(loggedInUser);
    listAttempts.mockReturnValue(deferred<QuizAttemptSummary[]>().promise);

    render(<HistoryPage />);

    expect(await screen.findByText('Loading history…')).toBeInTheDocument();
    expect(screen.queryByRole('listitem')).not.toBeInTheDocument();
  });

  it('shows the error state whose retry button repeats the client call', async () => {
    setSession(loggedInUser);
    listAttempts
      .mockRejectedValueOnce(new Error('network down'))
      .mockResolvedValueOnce([summaryOne]);

    const user = userEvent.setup();
    render(<HistoryPage />);

    expect(await screen.findByRole('heading', { name: 'Something went wrong' })).toBeInTheDocument();
    expect(screen.queryByRole('listitem')).not.toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: 'Try again' }));

    expect(await screen.findByRole('listitem')).toBeInTheDocument();
    expect(listAttempts).toHaveBeenCalledTimes(2);
    expect(screen.queryByRole('heading', { name: 'Something went wrong' })).not.toBeInTheDocument();
  });
});

describe('HistoryDetailPage', () => {
  it('shows the stored question, selected answer and correct answer of every answer', async () => {
    setSession(loggedInUser);
    getAttempt.mockResolvedValue(attempt);

    render(<HistoryDetailPage attemptId="att-1" />);

    expect(await screen.findByRole('heading', { name: 'Retrieval-Augmented Generation', level: 1 })).toBeInTheDocument();
    expect(getAttempt).toHaveBeenCalledTimes(1);
    expect(getAttempt).toHaveBeenCalledWith('att-1');

    const answers = screen.getAllByRole('listitem');
    expect(answers).toHaveLength(3);

    expect(within(answers[0]).getByText('What is RAG?')).toBeInTheDocument();
    expect(within(answers[0]).getAllByText('Retrieval-augmented generation')).toHaveLength(2);

    expect(within(answers[1]).getByText('What breaks a naive chunking?')).toBeInTheDocument();
    expect(within(answers[1]).getByText('Fine-tuning every document')).toBeInTheDocument();
    expect(within(answers[1]).getByText('Chunks too small lose the surrounding context')).toBeInTheDocument();

    expect(within(answers[2]).getByText('What is a token?')).toBeInTheDocument();
    expect(within(answers[2]).getByText('No answer selected')).toBeInTheDocument();
    expect(within(answers[2]).getByText('A unit of text the model reads')).toBeInTheDocument();
  });

  it('shows the wording stored with the attempt, not the current wording of the question', async () => {
    setSession(loggedInUser);
    getAttempt.mockResolvedValue({
      ...attempt,
      answers: [
        {
          ...attempt.answers[0],
          questionText: 'What is the main training objective of a large language model?',
          selectedOptionText: 'Predict the next word in a sequence',
          correctOptionText: 'Predict the next word in a sequence',
        },
      ],
    });

    render(<HistoryDetailPage attemptId="att-seed-1" />);

    expect(await screen.findByText('What is the main training objective of a large language model?')).toBeInTheDocument();
    expect(screen.getAllByText('Predict the next word in a sequence')).toHaveLength(2);
    expect(screen.queryByText('What is the core training objective of a large language model during pre-training?')).not.toBeInTheDocument();
  });

  it('asks an anonymous visitor to log in and never reads the attempt', async () => {
    render(<HistoryDetailPage attemptId="att-1" />);

    expect(await screen.findByRole('link', { name: 'Log in' })).toHaveAttribute('href', '#/login');
    expect(screen.queryByText('What is RAG?')).not.toBeInTheDocument();
    expect(getAttempt).not.toHaveBeenCalled();
  });

  it('shows the loading state while the client call is pending', async () => {
    setSession(loggedInUser);
    getAttempt.mockReturnValue(deferred<QuizAttempt>().promise);

    render(<HistoryDetailPage attemptId="att-1" />);

    expect(await screen.findByText('Loading attempt…')).toBeInTheDocument();
    expect(screen.queryByText('What is RAG?')).not.toBeInTheDocument();
  });

  it('shows the error state whose retry button repeats the client call', async () => {
    setSession(loggedInUser);
    getAttempt
      .mockRejectedValueOnce(new Error('network down'))
      .mockResolvedValueOnce(attempt);

    const user = userEvent.setup();
    render(<HistoryDetailPage attemptId="att-1" />);

    expect(await screen.findByRole('heading', { name: 'Something went wrong' })).toBeInTheDocument();
    expect(screen.queryByText('What is RAG?')).not.toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: 'Try again' }));

    expect(await screen.findByText('What is RAG?')).toBeInTheDocument();
    expect(getAttempt).toHaveBeenCalledTimes(2);
    expect(getAttempt).toHaveBeenLastCalledWith('att-1');
    expect(screen.queryByRole('heading', { name: 'Something went wrong' })).not.toBeInTheDocument();
  });
});
