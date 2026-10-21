// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { api } from '@/api';
import type { Lesson, Topic, TopicSummary, User } from '@/api';
import { useSession } from '@/session';
import { TopicPage } from './TopicPage';

vi.mock('@/api', () => ({
  api: {
    listTopics: vi.fn(),
    getTopic: vi.fn(),
    getProgress: vi.fn(),
  },
}));

vi.mock('@/session', () => ({
  useSession: vi.fn(),
}));

const listTopics = vi.mocked(api.listTopics);
const getTopic = vi.mocked(api.getTopic);
const getProgress = vi.mocked(api.getProgress);
const useSessionMock = vi.mocked(useSession);

const loggedInUser: User = {
  id: 'u1',
  email: 'user@example.com',
  isMaintainer: false,
  createdAt: '2024-01-01T00:00:00.000Z',
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
const lessonThree: Lesson = {
  id: 'les-3',
  topicId: 'top-a',
  title: 'Lesson Three',
  summary: 'The third lesson',
  body: 'Body three',
  position: 3,
  durationMin: 9,
};
const topicA: Topic = { ...summaryA, lessons: [lessonOne, lessonTwo] };
const topicAThree: Topic = { ...summaryA, lessons: [lessonOne, lessonTwo, lessonThree] };

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

  it('shows the completed count and one Completed badge per completed lesson', async () => {
    setSession(loggedInUser);
    listTopics.mockResolvedValue([summaryA]);
    getTopic.mockResolvedValue(topicAThree);
    getProgress.mockResolvedValue(['les-1', 'les-2']);

    render(<TopicPage slug="topic-a" />);

    expect(await screen.findByText('2 of 3')).toBeInTheDocument();
    expect(screen.getAllByText('Completed')).toHaveLength(2);
    expect(screen.getAllByLabelText('Completed')).toHaveLength(2);
    expect(getProgress).toHaveBeenCalledTimes(1);
  });

  it('shows the count without a badge when no lesson of the topic is completed', async () => {
    setSession(loggedInUser);
    listTopics.mockResolvedValue([summaryA]);
    getTopic.mockResolvedValue(topicAThree);
    getProgress.mockResolvedValue([]);

    render(<TopicPage slug="topic-a" />);

    expect(await screen.findByText('0 of 3')).toBeInTheDocument();
    expect(screen.queryByText('Completed')).not.toBeInTheDocument();
  });

  it('shows no count and no badge for an anonymous visitor and never reads the progress', async () => {
    listTopics.mockResolvedValue([summaryA]);
    getTopic.mockResolvedValue(topicAThree);

    render(<TopicPage slug="topic-a" />);

    expect(await screen.findByRole('heading', { name: 'Topic A', level: 1 })).toBeInTheDocument();
    expect(screen.queryByText('0 of 3')).not.toBeInTheDocument();
    expect(screen.queryByText('Completed')).not.toBeInTheDocument();
    expect(getProgress).not.toHaveBeenCalled();
  });

  it('shows no count and no badge when the progress rejects', async () => {
    setSession(loggedInUser);
    listTopics.mockResolvedValue([summaryA]);
    getTopic.mockResolvedValue(topicAThree);
    getProgress.mockRejectedValue(new Error('network down'));

    render(<TopicPage slug="topic-a" />);

    expect(await screen.findByRole('heading', { name: 'Topic A', level: 1 })).toBeInTheDocument();
    expect(screen.getByText('The first lesson')).toBeInTheDocument();
    expect(screen.queryByText('0 of 3')).not.toBeInTheDocument();
    expect(screen.queryByText('Completed')).not.toBeInTheDocument();
  });
});
