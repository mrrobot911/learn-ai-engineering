// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { AppShell } from './AppShell';
import { SessionProvider } from '@/session';
import { api } from '@/api';

vi.mock('@/api', () => ({
  api: {
    login: vi.fn(),
    register: vi.fn(),
    logout: vi.fn(),
  },
}));

const login = vi.mocked(api.login);
const logout = vi.mocked(api.logout);

function Wrapper({ children }: { children: React.ReactNode }) {
  return <SessionProvider>{children}</SessionProvider>;
}

beforeEach(() => {
  vi.resetAllMocks();
  Object.defineProperty(window, 'location', {
    value: { hash: '#/topics' },
    writable: true,
  });
});

afterEach(() => {
  cleanup();
});

describe('AppShell', () => {
  it('shows logged-out state when not authenticated', () => {
    render(
      <Wrapper>
        <AppShell>hi</AppShell>
      </Wrapper>,
    );
    expect(screen.getAllByRole('link', { name: /log in/i })[0]).toBeInTheDocument();
    expect(screen.getAllByRole('link', { name: /sign up/i })[0]).toBeInTheDocument();
    expect(screen.queryByText(/log out/i)).not.toBeInTheDocument();
  });

  it('shows logged-in header with email and logout control, and logs out', async () => {
    const user = userEvent.setup();
    login.mockResolvedValue({ id: 'u1', email: 'user@example.com', isMaintainer: false, createdAt: new Date().toISOString() });
    logout.mockResolvedValue();
    // We need to set user in session - render a helper component
    const { useSession } = await import('@/session');
    function LoginHelper() {
      const s = useSession();
      return <button onClick={() => s.login('user@example.com', 'password123')}>do</button>;
    }
    render(
      <Wrapper>
        <LoginHelper />
        <AppShell>hi</AppShell>
      </Wrapper>,
    );
    await user.click(screen.getByText('do'));
    await waitFor(() => expect(screen.getByText('user@example.com')).toBeInTheDocument());
    expect(screen.queryByText(/log in/i)).not.toBeInTheDocument();
    expect(screen.queryByText('Ada Demo')).not.toBeInTheDocument();
    expect(screen.queryByText(/name/i)).not.toBeInTheDocument();
    const logoutBtn = screen.getAllByRole('button', { name: /log out/i })[0];
    await user.click(logoutBtn);
    await waitFor(() => expect(logout).toHaveBeenCalledTimes(1));
  });
});
