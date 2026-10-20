// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { api } from '@/api';
import type { QuizQuestion, QuizResult, TopicSummary } from '@/api';
import { QuizPlayPage, QuizResultPage, QuizStartPage } from './QuizPages';

vi.mock('@/api', () => ({
  api: {
    listTopics: vi.fn(),
    startQuiz: vi.fn(),
    submitQuiz: vi.fn(),
  },
}));

const listTopics = vi.mocked(api.listTopics);
const startQuiz = vi.mocked(api.startQuiz);
const submitQuiz = vi.mocked(api.submitQuiz);

const summary: TopicSummary = {
  id: 'top-a',
  slug: 'topic-a',
  title: 'Topic A',
  description: 'The first topic',
  icon: 'BrainCircuit',
  color: 'brand',
  position: 1,
};

const questions: QuizQuestion[] = [
  {
    id: 'q1',
    text: 'What does RAG stand for?',
    options: [
      { id: 'q1-a', text: 'Retrieval-augmented generation' },
      { id: 'q1-b', text: 'Routed application gateway' },
      { id: 'q1-c', text: 'Random access generation' },
    ],
  },
  {
    id: 'q2',
    text: 'What is a token?',
    options: [
      { id: 'q2-a', text: 'A unit of text a model reads' },
      { id: 'q2-b', text: 'A login credential' },
    ],
  },
];

const result: QuizResult = {
  score: 1,
  total: 2,
  answers: [
    {
      questionId: 'q1',
      questionText: 'What does RAG stand for?',
      selectedOptionId: 'q1-a',
      selectedOptionText: 'Retrieval-augmented generation',
      correctOptionId: 'q1-a',
      correctOptionText: 'Retrieval-augmented generation',
      isCorrect: true,
      explanation: 'RAG grounds the answer in retrieved documents.',
    },
    {
      questionId: 'q2',
      questionText: 'What is a token?',
      selectedOptionId: null,
      selectedOptionText: '',
      correctOptionId: 'q2-a',
      correctOptionText: 'A unit of text a model reads',
      isCorrect: false,
      explanation: 'A token is a unit of text, not a credential.',
    },
  ],
};

function deferred<T>() {
  let resolve!: (value: T) => void;
  let reject!: (reason?: unknown) => void;
  const promise = new Promise<T>((res, rej) => { resolve = res; reject = rej; });
  return { promise, resolve, reject };
}

function optionButton(name: RegExp): HTMLElement {
  return screen.getByRole('button', { name });
}

beforeEach(() => {
  vi.resetAllMocks();
});

afterEach(cleanup);

describe('QuizStartPage', () => {
  it('resolves the slug through the client and shows no question-bank size, attempts or signed-out banner', async () => {
    listTopics.mockResolvedValue([summary]);

    render(<QuizStartPage slug="topic-a" />);

    expect(await screen.findByRole('heading', { name: 'Topic A Quiz', level: 1 })).toBeInTheDocument();
    expect(listTopics).toHaveBeenCalledTimes(1);
    expect(screen.getByRole('link', { name: /Start quiz/ })).toHaveAttribute('href', '#/quiz/topic-a/play');

    // Trimmed from the import: bank size, attempts, last attempt, store banner.
    expect(screen.queryByText('In bank')).not.toBeInTheDocument();
    expect(screen.queryByText('Attempts')).not.toBeInTheDocument();
    expect(screen.queryByText('Last attempt')).not.toBeInTheDocument();
    expect(screen.queryByText(/without an account/)).not.toBeInTheDocument();
    expect(screen.queryByText(/This topic has \d+ published question/)).not.toBeInTheDocument();
    expect(screen.queryByText(/Answer \d+/)).not.toBeInTheDocument();
    // No way to change the quiz size.
    expect(screen.queryByRole('spinbutton')).not.toBeInTheDocument();
    expect(screen.queryByRole('combobox')).not.toBeInTheDocument();
  });

  it('shows the loading state while the client call is pending', async () => {
    listTopics.mockReturnValue(deferred<TopicSummary[]>().promise);

    render(<QuizStartPage slug="topic-a" />);

    expect(await screen.findByText('Loading quiz…')).toBeInTheDocument();
    expect(screen.queryByRole('heading', { name: 'Topic A Quiz' })).not.toBeInTheDocument();
  });

  it('shows an error state whose retry button repeats the client call', async () => {
    listTopics
      .mockRejectedValueOnce(new Error('network down'))
      .mockResolvedValueOnce([summary]);

    const user = userEvent.setup();
    render(<QuizStartPage slug="topic-a" />);

    expect(await screen.findByText('Something went wrong')).toBeInTheDocument();
    expect(listTopics).toHaveBeenCalledTimes(1);

    await user.click(screen.getByRole('button', { name: 'Try again' }));

    expect(await screen.findByRole('heading', { name: 'Topic A Quiz', level: 1 })).toBeInTheDocument();
    expect(listTopics).toHaveBeenCalledTimes(2);
    expect(screen.queryByText('Something went wrong')).not.toBeInTheDocument();
  });

  it('shows a not-found state when no summary matches the slug', async () => {
    listTopics.mockResolvedValue([]);

    render(<QuizStartPage slug="missing" />);

    expect(await screen.findByText('Quiz not found')).toBeInTheDocument();
    expect(listTopics).toHaveBeenCalledTimes(1);
  });
});

describe('QuizPlayPage', () => {
  it('renders exactly the startQuiz questions in client order with no correct-answer data before submission', async () => {
    listTopics.mockResolvedValue([summary]);
    startQuiz.mockResolvedValue(questions);

    render(<QuizPlayPage slug="topic-a" />);

    expect(await screen.findByRole('heading', { name: 'What does RAG stand for?' })).toBeInTheDocument();
    expect(startQuiz).toHaveBeenCalledWith('top-a');

    // Options of the current question, in the order the client returned them.
    const optionTexts = screen
      .getAllByRole('button')
      .filter((b) => b.hasAttribute('aria-pressed'))
      .map((b) => b.textContent);
    expect(optionTexts).toEqual([
      'ARetrieval-augmented generation',
      'BRouted application gateway',
      'CRandom access generation',
    ]);

    // Nothing marks a correct option and no explanation reaches the DOM.
    expect(screen.queryByText('RAG grounds the answer in retrieved documents.')).not.toBeInTheDocument();
    expect(screen.queryByText('Correct')).not.toBeInTheDocument();
    expect(screen.queryByText('Incorrect')).not.toBeInTheDocument();
    expect(screen.queryByRole('img', { name: 'Correct' })).not.toBeInTheDocument();
    expect(screen.queryByRole('img', { name: 'Incorrect' })).not.toBeInTheDocument();
  });

  it('selects, clears and submits unanswered questions as optionId: null, then shows the response result', async () => {
    listTopics.mockResolvedValue([summary]);
    startQuiz.mockResolvedValue(questions);
    submitQuiz.mockResolvedValue(result);

    const user = userEvent.setup();
    const play = render(<QuizPlayPage slug="topic-a" />);
    await screen.findByRole('heading', { name: 'What does RAG stand for?' });

    // Select, clear by activating again, select again.
    await user.click(optionButton(/Retrieval-augmented generation/));
    expect(optionButton(/Retrieval-augmented generation/)).toHaveAttribute('aria-pressed', 'true');
    await user.click(optionButton(/Retrieval-augmented generation/));
    expect(optionButton(/Retrieval-augmented generation/)).toHaveAttribute('aria-pressed', 'false');
    await user.click(optionButton(/Retrieval-augmented generation/));
    expect(optionButton(/Retrieval-augmented generation/)).toHaveAttribute('aria-pressed', 'true');

    // Question 2 stays unanswered; submit from the last question.
    await user.click(screen.getByRole('button', { name: /Next/ }));
    await screen.findByRole('heading', { name: 'What is a token?' });
    const hashBefore = window.location.hash;
    await user.click(screen.getByRole('button', { name: /Submit quiz/ }));

    // The result route is only entered once the submission response arrived.
    await waitFor(() => expect(window.location.hash).not.toBe(hashBefore));
    expect(submitQuiz).toHaveBeenCalledTimes(1);
    expect(submitQuiz).toHaveBeenCalledWith({
      topicId: 'top-a',
      answers: [
        { questionId: 'q1', optionId: 'q1-a' },
        { questionId: 'q2', optionId: null },
      ],
    });

    // Nothing is written to browser storage.
    expect(window.localStorage.length).toBe(0);
    expect(window.sessionStorage.length).toBe(0);
    expect(window.location.hash).toMatch(/^#\/quiz\/topic-a\/result\//);

    // The result view renders the response only.
    play.unmount();
    const attemptId = window.location.hash.split('/').pop() as string;
    render(<QuizResultPage slug="topic-a" attemptId={attemptId} />);

    expect(await screen.findByText('Topic A · Quiz result')).toBeInTheDocument();
    const total = screen.getByText('/ 2');
    expect(total.previousElementSibling?.textContent).toBe('1');
    expect(screen.getByRole('img', { name: 'Correct' })).toBeInTheDocument();
    expect(screen.getByRole('img', { name: 'Incorrect' })).toBeInTheDocument();
    expect(screen.getByText('RAG grounds the answer in retrieved documents.')).toBeInTheDocument();
    expect(screen.getByText('A unit of text a model reads')).toBeInTheDocument();
    expect(screen.getByText('No answer selected')).toBeInTheDocument();
    expect(screen.getByText('A token is a unit of text, not a credential.')).toBeInTheDocument();
    expect(screen.queryByText(/%/)).not.toBeInTheDocument();
  });

  it('keeps the answers and offers a retry when the submission rejects', async () => {
    listTopics.mockResolvedValue([summary]);
    startQuiz.mockResolvedValue(questions);
    submitQuiz
      .mockRejectedValueOnce(new Error('network down'))
      .mockResolvedValueOnce(result);

    const user = userEvent.setup();
    render(<QuizPlayPage slug="topic-a" />);
    await screen.findByRole('heading', { name: 'What does RAG stand for?' });
    await user.click(optionButton(/Retrieval-augmented generation/));
    await user.click(screen.getByRole('button', { name: /Next/ }));
    await screen.findByRole('heading', { name: 'What is a token?' });
    await user.click(screen.getByRole('button', { name: /Submit quiz/ }));

    const alert = await screen.findByRole('alert');
    expect(alert).toHaveTextContent("Couldn't submit the quiz. Your answers are still here.");
    expect(submitQuiz).toHaveBeenCalledTimes(1);
    // The unanswered selection survived the failure.
    await user.click(screen.getByRole('button', { name: /Go to question 1/ }));
    expect(optionButton(/Retrieval-augmented generation/)).toHaveAttribute('aria-pressed', 'true');

    await user.click(within(alert).getByRole('button', { name: 'Try again' }));

    await waitFor(() => expect(submitQuiz).toHaveBeenCalledTimes(2));
    expect(submitQuiz).toHaveBeenLastCalledWith({
      topicId: 'top-a',
      answers: [
        { questionId: 'q1', optionId: 'q1-a' },
        { questionId: 'q2', optionId: null },
      ],
    });
  });

  it('shows the loading state while startQuiz is pending', async () => {
    listTopics.mockResolvedValue([summary]);
    startQuiz.mockReturnValue(deferred<QuizQuestion[]>().promise);

    render(<QuizPlayPage slug="topic-a" />);

    expect(await screen.findByText('Loading question…')).toBeInTheDocument();
    expect(screen.queryByRole('heading', { name: 'What does RAG stand for?' })).not.toBeInTheDocument();
  });

  it('shows an error state whose retry button repeats the client calls', async () => {
    listTopics.mockResolvedValue([summary]);
    startQuiz
      .mockRejectedValueOnce(new Error('network down'))
      .mockResolvedValueOnce(questions);

    const user = userEvent.setup();
    render(<QuizPlayPage slug="topic-a" />);

    expect(await screen.findByText('Something went wrong')).toBeInTheDocument();
    expect(startQuiz).toHaveBeenCalledTimes(1);

    await user.click(screen.getByRole('button', { name: 'Try again' }));

    expect(await screen.findByRole('heading', { name: 'What does RAG stand for?' })).toBeInTheDocument();
    expect(startQuiz).toHaveBeenCalledTimes(2);
    expect(screen.queryByText('Something went wrong')).not.toBeInTheDocument();
  });

  it('selects and clears an option with Enter and Space from the keyboard', async () => {
    listTopics.mockResolvedValue([summary]);
    startQuiz.mockResolvedValue(questions);

    const user = userEvent.setup();
    render(<QuizPlayPage slug="topic-a" />);
    await screen.findByRole('heading', { name: 'What does RAG stand for?' });

    const option = optionButton(/Retrieval-augmented generation/);
    option.focus();
    expect(document.activeElement).toBe(option);

    await user.keyboard('{Enter}');
    expect(optionButton(/Retrieval-augmented generation/)).toHaveAttribute('aria-pressed', 'true');

    await user.keyboard(' ');
    expect(optionButton(/Retrieval-augmented generation/)).toHaveAttribute('aria-pressed', 'false');
    expect(document.activeElement).toBe(optionButton(/Retrieval-augmented generation/));
  });
});

describe('QuizResultPage', () => {
  it('shows the empty state when no submission is in memory, as after a reload', () => {
    render(<QuizResultPage slug="topic-a" attemptId="reloaded-attempt" />);

    expect(screen.getByText('Result not found')).toBeInTheDocument();
    expect(screen.queryByText('Topic A · Quiz result')).not.toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Back to quiz' })).toHaveAttribute('href', '#/quiz/topic-a');
  });
});
