// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { api } from '@/api';
import type { Lesson, Topic, TopicSummary } from '@/api';
import { TopicPage } from './TopicPage';

vi.mock('@/api', () => ({
  api: {
    listTopics: vi.fn(),
    getTopic: vi.fn(),
  },
}));

const listTopics = vi.mocked(api.listTopics);
const getTopic = vi.mocked(api.getTopic);

const summaryA: TopicSummary = {
  id: 'top-a',
  slug: 'topic-a',
  title: 'Topic A',
  description: 'The first topic',
  icon: 'BrainCircuit',
  color: 'brand',
  position: 1,
};
const summaryB: TopicSummary = {
  id: 'top-b',
  slug: 'topic-b',
  title: 'Topic B',
  description: 'The second topic',
  icon: 'Bot',
  color: 'accent',
  position: 2,
};

const lessonOne: Lesson = {
  id: 'les-1',
  topicId: 'top-a',
  title: 'Lesson One',
  summary: 'The first lesson',
  body: 'Body one',
  position: 1,
  durationMin: 5,
};
const lessonTwo: Lesson = {
  id: 'les-2',
  topicId: 'top-a',
  title: 'Lesson Two',
  summary: 'The second lesson',
  body: 'Body two',
  position: 2,
  durationMin: 7,
};
const topicA: Topic = { ...summaryA, lessons: [lessonOne, lessonTwo] };

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

describe('TopicPage', () => {
  it('resolves the slug through the client and lists lessons in response order with links', async () => {
    listTopics.mockResolvedValue([summaryA, summaryB]);
    getTopic.mockResolvedValue(topicA);

    render(<TopicPage slug="topic-a" />);

    expect(await screen.findByRole('heading', { name: 'Topic A', level: 1 })).toBeInTheDocument();
    expect(screen.getByText('The first topic')).toBeInTheDocument();
    expect(listTopics).toHaveBeenCalled();
    expect(getTopic).toHaveBeenCalledWith('top-a');

    const lessonHeadings = screen.getAllByRole('heading', { level: 3 });
    expect(lessonHeadings.map((h) => h.textContent)).toEqual(['Lesson One', 'Lesson Two']);

    const lessonLinks = screen
      .getAllByRole('link')
      .filter((a) => a.getAttribute('href')?.startsWith('#/lesson/'));
    expect(lessonLinks).toHaveLength(2);
    expect(lessonLinks[0]).toHaveAttribute('href', '#/lesson/les-1');
    expect(lessonLinks[1]).toHaveAttribute('href', '#/lesson/les-2');
  });

  it('shows the loading state while the client call is pending', async () => {
    listTopics.mockReturnValue(deferred<TopicSummary[]>().promise);

    render(<TopicPage slug="topic-a" />);

    expect(await screen.findByText('Loading topic…')).toBeInTheDocument();
    expect(screen.queryByRole('heading', { name: 'Topic A' })).not.toBeInTheDocument();
  });

  it('shows an error state whose retry button repeats the client call', async () => {
    listTopics.mockResolvedValue([summaryA]);
    getTopic
      .mockRejectedValueOnce(new Error('network down'))
      .mockResolvedValueOnce(topicA);

    const user = userEvent.setup();
    render(<TopicPage slug="topic-a" />);

    expect(await screen.findByText('Something went wrong')).toBeInTheDocument();
    expect(getTopic).toHaveBeenCalledTimes(1);

    await user.click(screen.getByRole('button', { name: 'Try again' }));

    expect(await screen.findByRole('heading', { name: 'Topic A', level: 1 })).toBeInTheDocument();
    expect(listTopics).toHaveBeenCalledTimes(2);
    expect(getTopic).toHaveBeenCalledTimes(2);
    expect(screen.queryByText('Something went wrong')).not.toBeInTheDocument();
  });

  it('shows a not-found state with a way back when listTopics() has no summary for the slug', async () => {
    listTopics.mockResolvedValue([summaryB]);
    getTopic.mockResolvedValue(topicA);

    render(<TopicPage slug="topic-a" />);

    expect(await screen.findByText('Topic not found')).toBeInTheDocument();
    const back = screen.getByRole('link', { name: 'Back to topics' });
    expect(back).toHaveAttribute('href', '#/topics');
    expect(getTopic).not.toHaveBeenCalled();
    expect(screen.queryByText('Lessons')).not.toBeInTheDocument();
    expect(screen.queryAllByRole('link').filter(
      (a) => a.getAttribute('href')?.startsWith('#/lesson/'),
    )).toHaveLength(0);
  });

  it('shows an empty lessons state when the loaded topic has zero lessons', async () => {
    listTopics.mockResolvedValue([summaryA]);
    getTopic.mockResolvedValue({ ...topicA, lessons: [] });

    render(<TopicPage slug="topic-a" />);

    expect(await screen.findByText('No lessons yet')).toBeInTheDocument();
    const lessonsHeading = screen.getByRole('heading', { name: 'Lessons' });
    expect(within(lessonsHeading.parentElement as HTMLElement).getByText('No lessons yet'))
      .toBeInTheDocument();
    expect(screen.queryAllByRole('link').filter(
      (a) => a.getAttribute('href')?.startsWith('#/lesson/'),
    )).toHaveLength(0);
  });
});
