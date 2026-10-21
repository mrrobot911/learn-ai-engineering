// @vitest-environment jsdom
// The lesson completion across the screens and the client: the mock stands in
// for the backend, so the flow runs from the button to the stored progress.
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { useEffect } from 'react';
import { cleanup, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { mockClient, resetMockState } from '@/api/mock';
import { SessionProvider, useSession } from '@/session';
import { LessonPage } from '@/pages/LessonPage';
import { TopicPage } from '@/pages/TopicPage';

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

describe('lesson completion', () => {
  it('keeps the completion in the client and shows it on the topic page', async () => {
    const user = userEvent.setup();
    render(
      <SessionProvider>
        <LoginOnMount />
        <LessonPage id="les-llm-1" />
      </SessionProvider>,
    );

    const button = await screen.findByRole('button', { name: /complete/i });
    expect(button).toHaveAttribute('aria-pressed', 'false');

    await user.click(button);
    await waitFor(() => expect(button).toHaveAttribute('aria-pressed', 'true'));

    expect(await mockClient.getProgress()).toEqual(['les-llm-1']);

    cleanup();
    render(
      <SessionProvider>
        <LoginOnMount />
        <TopicPage slug="llms" />
      </SessionProvider>,
    );

    expect(await screen.findByText('1 of 3')).toBeInTheDocument();
    expect(screen.getAllByText('Completed')).toHaveLength(1);
  });

  it('rejects the progress calls without a session', async () => {
    await expect(mockClient.getProgress()).rejects.toThrow();
    await expect(mockClient.markLessonComplete('les-llm-1')).rejects.toThrow();
    await expect(mockClient.undoLessonComplete('les-llm-1')).rejects.toThrow();
  });
});
