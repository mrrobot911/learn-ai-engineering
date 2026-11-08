// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, render, screen, waitFor } from '@testing-library/react';
import { api, setOnUnauthorized } from '@/api';
import { toast } from '@/components/toast';
import { SessionProvider } from './SessionProvider';
import { useSession } from './useSession';

vi.mock('@/api', () => ({
  api: {
    getSession: vi.fn(),
    login: vi.fn(),
    register: vi.fn(),
    logout: vi.fn(),
  },
  setOnUnauthorized: vi.fn(),
}));

vi.mock('@/components/toast', () => ({
  toast: vi.fn(),
}));

const getSession = vi.mocked(api.getSession);
const setOnUnauthorizedMock = vi.mocked(setOnUnauthorized);
const toastMock = vi.mocked(toast);

function deferred<T>() {
  let resolve!: (value: T) => void;
  let reject!: (reason?: unknown) => void;
  const promise = new Promise<T>((res, rej) => {
    resolve = res;
    reject = rej;
  });
  return { promise, resolve, reject };
}

function Probe() {
  const { user, authReady } = useSession();
  return (
    <div>
      <span data-testid="ready">{String(authReady)}</span>
      <span data-testid="email">{user?.email ?? 'anon'}</span>
    </div>
  );
}

beforeEach(() => {
  vi.resetAllMocks();
});

afterEach(() => {
  cleanup();
});

describe('SessionProvider', () => {
  it('keeps authReady false until the session bootstrap settles (pending)', async () => {
    const d = deferred<{ id: string; email: string }>();
    getSession.mockReturnValue(d.promise);

    render(
      <SessionProvider>
        <Probe />
      </SessionProvider>,
    );

    expect(api.getSession).toHaveBeenCalledTimes(1);
    expect(screen.getByTestId('ready')).toHaveTextContent('false');
    expect(screen.getByTestId('email')).toHaveTextContent('anon');

    d.resolve({ id: 'u1', email: 'ada@example.com' });
    await waitFor(() =>
      expect(screen.getByTestId('ready')).toHaveTextContent('true'),
    );
  });

  it('sets the user on a 200 bootstrap response', async () => {
    getSession.mockResolvedValue({ id: 'u1', email: 'ada@example.com' });

    render(
      <SessionProvider>
        <Probe />
      </SessionProvider>,
    );

    await waitFor(() =>
      expect(screen.getByTestId('ready')).toHaveTextContent('true'),
    );
    expect(screen.getByTestId('email')).toHaveTextContent('ada@example.com');
    expect(toastMock).not.toHaveBeenCalled();
  });

  it('leaves the user anonymous without an error banner on a 401 bootstrap', async () => {
    getSession.mockRejectedValue(
      Object.assign(new Error('No valid session.'), { status: 401 }),
    );

    render(
      <SessionProvider>
        <Probe />
      </SessionProvider>,
    );

    await waitFor(() =>
      expect(screen.getByTestId('ready')).toHaveTextContent('true'),
    );
    expect(screen.getByTestId('email')).toHaveTextContent('anon');
    expect(toastMock).not.toHaveBeenCalled();
  });

  it('clears the session and shows a clear message when a protected call returns 401 (expiry)', async () => {
    getSession.mockResolvedValue({ id: 'u1', email: 'ada@example.com' });

    render(
      <SessionProvider>
        <Probe />
      </SessionProvider>,
    );

    await waitFor(() =>
      expect(screen.getByTestId('email')).toHaveTextContent('ada@example.com'),
    );

    const calls = setOnUnauthorizedMock.mock.calls;
    const handler = calls[calls.length - 1]?.[0] as (() => void) | undefined;
    expect(handler).toEqual(expect.any(Function));
    handler?.();

    await waitFor(() =>
      expect(screen.getByTestId('email')).toHaveTextContent('anon'),
    );
    expect(toastMock).toHaveBeenCalledWith(
      expect.stringMatching(/session expired/i),
      'error',
    );
  });
});
