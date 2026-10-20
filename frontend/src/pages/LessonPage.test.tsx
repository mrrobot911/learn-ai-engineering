// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { api } from '@/api';
import type { Lesson, Topic } from '@/api';
import { LessonPage } from './LessonPage';

vi.mock('@/api', () => ({
  api: {
    getLesson: vi.fn(),
    getTopic: vi.fn(),
  },
}));

const getLesson = vi.mocked(api.getLesson);
const getTopic = vi.mocked(api.getTopic);

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
});
