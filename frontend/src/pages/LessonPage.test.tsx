// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { api } from '@/api';
import type { Lesson, Topic, User } from '@/api';
import { useSession } from '@/session';
import { LessonPage } from './LessonPage';

vi.mock('@/api', () => ({
  api: {
    getLesson: vi.fn(),
    getTopic: vi.fn(),
    getProgress: vi.fn(),
    markLessonComplete: vi.fn(),
    undoLessonComplete: vi.fn(),
  },
}));

vi.mock('@/session', () => ({
  useSession: vi.fn(),
}));

const getLesson = vi.mocked(api.getLesson);
const getTopic = vi.mocked(api.getTopic);
const getProgress = vi.mocked(api.getProgress);
const markLessonComplete = vi.mocked(api.markLessonComplete);
const undoLessonComplete = vi.mocked(api.undoLessonComplete);
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

const topic: Topic = {
  id: 'top-graph',
  slug: 'graph-rag',
  title: 'Graph RAG',
  description: 'Connecting the dots',
  icon: 'Network',
  color: 'brand',
  position: 1,
  lessons: [],
};

const lesson: Lesson = {
  id: 'les-1',
  topicId: 'top-graph',
  title: 'Why tokens matter',
  summary: 'Tokens are the unit of cost and context',
  body: 'Tokens are how the model sees text.',
  position: 1,
  durationMin: 5,
};

const topicWithLesson: Topic = { ...topic, lessons: [lesson] };

const markdownBody = [
  'Tokens are how the model sees text.',
  '',
  '## Key takeaway',
  '',
  '- Cost scales with tokens',
  '- Context windows are measured in tokens',
  '',
  '```ts',
  'const tokens = text.split(/\\s+/);',
  '```',
  '',
  '<b>hello</b>',
].join('\n');

/** The client's not-found signal: an error whose status is 404. */
function notFoundError(): Error & { status: number } {
  return Object.assign(new Error('No lesson with id "les-1".'), { status: 404 });
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

describe('LessonPage', () => {
  it('loads the lesson and its topic through the client and shows them', async () => {
    getLesson.mockResolvedValue(lesson);
    getTopic.mockResolvedValue(topicWithLesson);

    render(<LessonPage id="les-1" />);

    expect(await screen.findByRole('heading', { name: 'Why tokens matter', level: 1 })).toBeInTheDocument();
    expect(getLesson).toHaveBeenCalledWith('les-1');
    expect(getTopic).toHaveBeenCalledWith('top-graph');
    expect(screen.getByRole('link', { name: 'Graph RAG' })).toHaveAttribute('href', '#/topic/graph-rag');
    expect(screen.getByText(lesson.summary)).toBeInTheDocument();
    expect(screen.getByText('Tokens are how the model sees text.')).toBeInTheDocument();
    expect(screen.getAllByRole('heading', { level: 1 })).toHaveLength(1);
  });

  it('renders the Markdown body as elements and keeps raw HTML as text', async () => {
    getLesson.mockResolvedValue({ ...lesson, body: markdownBody });
    getTopic.mockResolvedValue(topicWithLesson);

    const { container } = render(<LessonPage id="les-1" />);
    await screen.findByRole('heading', { name: 'Why tokens matter', level: 1 });

    const heading = screen.getByRole('heading', { name: 'Key takeaway', level: 2 });
    expect(heading).toBeInTheDocument();
    expect(heading.textContent).not.toContain('##');

    const items = screen.getAllByRole('listitem');
    expect(items).toHaveLength(2);
    expect(items[0]).toHaveTextContent('Cost scales with tokens');
    expect(items[1]).toHaveTextContent('Context windows are measured in tokens');

    const code = screen.getByText('const tokens = text.split(/\\s+/);').closest('pre');
    expect(code).not.toBeNull();
    expect(code?.querySelector('code')).not.toBeNull();

    expect(screen.getByText('<b>hello</b>')).toBeInTheDocument();
    expect(container.querySelector('b')).toBeNull();
  });

  it('shows the loading state while the client calls are pending, with no fake latency', async () => {
    const lessonCall = deferred<Lesson>();
    getLesson.mockReturnValue(lessonCall.promise);
    getTopic.mockResolvedValue(topicWithLesson);

    render(<LessonPage id="les-1" />);

    expect(await screen.findByText('Loading lesson…')).toBeInTheDocument();
    expect(screen.queryByRole('heading', { name: 'Why tokens matter' })).not.toBeInTheDocument();
    expect(screen.queryByText(lesson.summary)).not.toBeInTheDocument();
    expect(getTopic).not.toHaveBeenCalled();

    lessonCall.resolve(lesson);
    expect(await screen.findByRole('heading', { name: 'Why tokens matter', level: 1 })).toBeInTheDocument();
    expect(screen.queryByText('Loading lesson…')).not.toBeInTheDocument();
  });

  it('shows an error state whose retry button repeats both client calls', async () => {
    getLesson
      .mockRejectedValueOnce(new Error('network down'))
      .mockResolvedValueOnce(lesson);
    getTopic.mockResolvedValue(topicWithLesson);

    const user = userEvent.setup();
    render(<LessonPage id="les-1" />);

    expect(await screen.findByText('Something went wrong')).toBeInTheDocument();
    expect(screen.queryByRole('heading', { name: 'Why tokens matter' })).not.toBeInTheDocument();
    expect(screen.queryByText(lesson.summary)).not.toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: 'Try again' }));

    expect(await screen.findByRole('heading', { name: 'Why tokens matter', level: 1 })).toBeInTheDocument();
    expect(getLesson).toHaveBeenCalledTimes(2);
    expect(getTopic).toHaveBeenCalledTimes(1);
    expect(getTopic).toHaveBeenCalledWith('top-graph');
    expect(screen.queryByText('Something went wrong')).not.toBeInTheDocument();
  });

  it('shows the error state when getTopic rejects without the not-found signal', async () => {
    getLesson.mockResolvedValue(lesson);
    getTopic.mockRejectedValue(new Error('network down'));

    render(<LessonPage id="les-1" />);

    expect(await screen.findByText('Something went wrong')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Try again' })).toBeInTheDocument();
    expect(screen.queryByRole('heading', { name: 'Why tokens matter' })).not.toBeInTheDocument();
  });

  it('shows a not-found state with a way back when getLesson rejects with the 404 signal', async () => {
    getLesson.mockRejectedValue(notFoundError());

    render(<LessonPage id="les-1" />);

    expect(await screen.findByText('Lesson not found')).toBeInTheDocument();
    const back = screen.getByRole('link', { name: 'Back to topics' });
    expect(back).toHaveAttribute('href', '#/topics');
    expect(getTopic).not.toHaveBeenCalled();
    expect(screen.queryByRole('button', { name: 'Try again' })).not.toBeInTheDocument();
    expect(screen.queryByText(lesson.summary)).not.toBeInTheDocument();
    expect(screen.queryByText(lesson.title)).not.toBeInTheDocument();
  });

  it('shows one completion control for a logged-in visitor and marks the lesson complete', async () => {
    setSession(loggedInUser);
    getLesson.mockResolvedValue(lesson);
    getTopic.mockResolvedValue(topicWithLesson);
    getProgress.mockResolvedValue([]);

    const user = userEvent.setup();
    render(<LessonPage id="les-1" />);

    const button = await screen.findByRole('button', { name: /complete/i });
    expect(screen.getAllByRole('button', { name: /complete/i })).toHaveLength(1);
    expect(button).toHaveAttribute('aria-pressed', 'false');
    expect(getProgress).toHaveBeenCalledTimes(1);

    await user.click(button);

    await waitFor(() => expect(markLessonComplete).toHaveBeenCalledWith('les-1'));
    await waitFor(() => expect(button).toHaveAttribute('aria-pressed', 'true'));
    expect(undoLessonComplete).not.toHaveBeenCalled();
  });

  it('shows the completed state and undoes it for a lesson the progress lists', async () => {
    setSession(loggedInUser);
    getLesson.mockResolvedValue(lesson);
    getTopic.mockResolvedValue(topicWithLesson);
    getProgress.mockResolvedValue(['les-1']);

    const user = userEvent.setup();
    render(<LessonPage id="les-1" />);

    const button = await screen.findByRole('button', { name: /complete/i });
    expect(button).toHaveAttribute('aria-pressed', 'true');

    await user.click(button);

    await waitFor(() => expect(undoLessonComplete).toHaveBeenCalledWith('les-1'));
    await waitFor(() => expect(button).toHaveAttribute('aria-pressed', 'false'));
    expect(markLessonComplete).not.toHaveBeenCalled();
  });

  it('disables the control while the toggle is pending and ignores a second activation', async () => {
    setSession(loggedInUser);
    getLesson.mockResolvedValue(lesson);
    getTopic.mockResolvedValue(topicWithLesson);
    getProgress.mockResolvedValue([]);
    const pending = deferred<void>();
    markLessonComplete.mockReturnValue(pending.promise);

    const user = userEvent.setup();
    render(<LessonPage id="les-1" />);

    const button = await screen.findByRole('button', { name: /complete/i });
    await user.click(button);

    await waitFor(() => expect(button).toBeDisabled());
    await user.click(button);
    expect(markLessonComplete).toHaveBeenCalledTimes(1);

    pending.resolve();
    await waitFor(() => expect(button).not.toBeDisabled());
    expect(button).toHaveAttribute('aria-pressed', 'true');
  });

  it('shows an error and keeps the state when the toggle rejects', async () => {
    setSession(loggedInUser);
    getLesson.mockResolvedValue(lesson);
    getTopic.mockResolvedValue(topicWithLesson);
    getProgress.mockResolvedValue([]);
    markLessonComplete.mockRejectedValue(new Error('network down'));

    const user = userEvent.setup();
    render(<LessonPage id="les-1" />);

    const button = await screen.findByRole('button', { name: /complete/i });
    await user.click(button);

    expect(await screen.findByRole('alert')).toBeInTheDocument();
    expect(button).toHaveAttribute('aria-pressed', 'false');
    expect(button).not.toBeDisabled();
  });

  it('shows no completion control and never reads the progress for an anonymous visitor', async () => {
    getLesson.mockResolvedValue(lesson);
    getTopic.mockResolvedValue(topicWithLesson);

    render(<LessonPage id="les-1" />);

    expect(await screen.findByRole('heading', { name: 'Why tokens matter', level: 1 })).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /complete/i })).not.toBeInTheDocument();
    expect(getProgress).not.toHaveBeenCalled();
  });

  it('shows the lesson without a completion control when the progress rejects', async () => {
    setSession(loggedInUser);
    getLesson.mockResolvedValue(lesson);
    getTopic.mockResolvedValue(topicWithLesson);
    getProgress.mockRejectedValue(new Error('network down'));

    render(<LessonPage id="les-1" />);

    expect(await screen.findByRole('heading', { name: 'Why tokens matter', level: 1 })).toBeInTheDocument();
    expect(screen.getByText(lesson.summary)).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /complete/i })).not.toBeInTheDocument();
  });
});
