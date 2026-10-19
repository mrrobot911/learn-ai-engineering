// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { api } from '@/api';
import type { TopicSummary } from '@/api';
import { TopicsPage } from './TopicsPage';

vi.mock('@/api', () => ({
  api: {
    listTopics: vi.fn(),
    getTopic: vi.fn(),
  },
}));

const listTopics = vi.mocked(api.listTopics);

const topicA: TopicSummary = {
  id: 'top-a',
  slug: 'topic-a',
  title: 'Topic A',
  description: 'The first topic',
  icon: 'BrainCircuit',
  color: 'brand',
  position: 1,
};
const topicB: TopicSummary = {
  id: 'top-b',
  slug: 'topic-b',
  title: 'Topic B',
  description: 'The second topic',
  icon: 'Bot',
  color: 'accent',
  position: 2,
};

function deferred<T>() {
  let resolve!: (value: T) => void;
  let reject!: (reason?: unknown) => void;
  const promise = new Promise<T>((res, rej) => { resolve = res; reject = rej; });
  return { promise, resolve, reject };
}

beforeEach(() => {
  vi.resetAllMocks();
});

afterEach(cleanup);

describe('TopicsPage', () => {
  it('renders topics in listTopics() response order with title, description and link', async () => {
    listTopics.mockResolvedValue([topicA, topicB]);

    render(<TopicsPage />);

    expect(await screen.findByRole('heading', { name: 'Topic A', level: 2 })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'Topic B', level: 2 })).toBeInTheDocument();
    expect(screen.getByText('The first topic')).toBeInTheDocument();
    expect(screen.getByText('The second topic')).toBeInTheDocument();

    const links = screen.getAllByRole('link');
    expect(links).toHaveLength(2);
    expect(links[0]).toHaveAttribute('href', '#/topic/topic-a');
    expect(links[1]).toHaveAttribute('href', '#/topic/topic-b');
  });

  it('shows the loading state while listTopics() is pending', async () => {
    listTopics.mockReturnValue(deferred<TopicSummary[]>().promise);

    render(<TopicsPage />);

    expect(await screen.findByText('Loading topics…')).toBeInTheDocument();
    expect(screen.queryByRole('heading', { name: 'Topic A' })).not.toBeInTheDocument();
  });

  it('shows the empty state when listTopics() resolves with []', async () => {
    listTopics.mockResolvedValue([]);

    render(<TopicsPage />);

    expect(await screen.findByText('No topics yet')).toBeInTheDocument();
    expect(screen.queryAllByRole('link')).toHaveLength(0);
  });

  it('shows an error state whose retry button calls listTopics() again and then renders the topics', async () => {
    listTopics
      .mockRejectedValueOnce(new Error('network down'))
      .mockResolvedValueOnce([topicA]);

    const user = userEvent.setup();
    render(<TopicsPage />);

    expect(await screen.findByText('Something went wrong')).toBeInTheDocument();
    expect(listTopics).toHaveBeenCalledTimes(1);

    await user.click(screen.getByRole('button', { name: 'Try again' }));

    expect(await screen.findByRole('heading', { name: 'Topic A', level: 2 })).toBeInTheDocument();
    expect(listTopics).toHaveBeenCalledTimes(2);
    expect(screen.queryByText('Something went wrong')).not.toBeInTheDocument();
  });
});
