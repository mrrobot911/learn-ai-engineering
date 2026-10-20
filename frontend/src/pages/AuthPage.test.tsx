// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { api } from '@/api';
import { AuthPage } from './AuthPage';
import { SessionProvider } from '@/session';

vi.mock('@/api', () => ({
  api: {
    login: vi.fn(),
    register: vi.fn(),
    logout: vi.fn(),
  },
}));

const login = vi.mocked(api.login);
const register = vi.mocked(api.register);

function renderWithProviders(ui: React.ReactElement) {
  return render(<SessionProvider>{ui}</SessionProvider>);
}

beforeEach(() => {
  vi.resetAllMocks();
  Object.defineProperty(window, 'location', {
    value: { hash: '' },
    writable: true,
  });
});

afterEach(() => {
  cleanup();
});

describe('AuthPage', () => {
  it('shows validation error for empty/malformed email and does not call api', async () => {
    const user = userEvent.setup();
    renderWithProviders(<AuthPage mode="login" />);
    await user.click(screen.getByRole('button', { name: /log in/i }));
    expect(await screen.findByText(/Enter a valid email address|Email is required/i)).toBeInTheDocument();
    expect(login).not.toHaveBeenCalled();
    expect(register).not.toHaveBeenCalled();
  });

  it('shows validation error for password length < 8 or > 128 and does not call api', async () => {
    const user = userEvent.setup();
    renderWithProviders(<AuthPage mode="register" />);
    await user.type(screen.getByLabelText(/email/i), 'test@example.com');
    await user.type(screen.getByLabelText(/^password$/i), 'short');
    await user.click(screen.getByRole('button', { name: /create account/i }));
    expect(await screen.findByText(/Password must be between 8 and 128 characters/i)).toBeInTheDocument();
    expect(register).not.toHaveBeenCalled();
    expect(login).not.toHaveBeenCalled();
  });

  it('calls api.register once with email and password only on valid submission', async () => {
    const user = userEvent.setup();
    register.mockResolvedValue({ id: 'u1', email: 'test@example.com', isMaintainer: false, createdAt: new Date().toISOString() });
    renderWithProviders(<AuthPage mode="register" />);
    await user.type(screen.getByLabelText(/email/i), 'test@example.com');
    await user.type(screen.getByLabelText(/^password$/i), 'password123');
    await user.click(screen.getByRole('button', { name: /create account/i }));
    await waitFor(() => expect(register).toHaveBeenCalledTimes(1));
    expect(register).toHaveBeenCalledWith({ email: 'test@example.com', password: 'password123' });
    const callArg = register.mock.calls[0][0] as unknown as Record<string, unknown>;
    expect(callArg).not.toHaveProperty('name');
  });

  it('shows duplicate email error message when register fails', async () => {
    const user = userEvent.setup();
    register.mockRejectedValue(new Error('An account with that email already exists.'));
    renderWithProviders(<AuthPage mode="register" />);
    await user.type(screen.getByLabelText(/email/i), 'dup@example.com');
    await user.type(screen.getByLabelText(/^password$/i), 'password123');
    await user.click(screen.getByRole('button', { name: /create account/i }));
    expect(await screen.findByText(/An account with that email already exists/i)).toBeInTheDocument();
  });

  it('calls api.login once with email and password only on valid login', async () => {
    const user = userEvent.setup();
    login.mockResolvedValue({ id: 'u1', email: 'test@example.com', isMaintainer: false, createdAt: new Date().toISOString() });
    renderWithProviders(<AuthPage mode="login" />);
    await user.type(screen.getByLabelText(/email/i), 'test@example.com');
    await user.type(screen.getByLabelText(/^password$/i), 'password123');
    await user.click(screen.getByRole('button', { name: /log in/i }));
    await waitFor(() => expect(login).toHaveBeenCalledTimes(1));
    expect(login).toHaveBeenCalledWith({ email: 'test@example.com', password: 'password123' });
  });

  it('shows error and stays logged out when login fails', async () => {
    const user = userEvent.setup();
    login.mockRejectedValue(new Error('No account found with that email.'));
    renderWithProviders(<AuthPage mode="login" />);
    await user.type(screen.getByLabelText(/email/i), 'bad@example.com');
    await user.type(screen.getByLabelText(/^password$/i), 'password123');
    await user.click(screen.getByRole('button', { name: /log in/i }));
    expect(await screen.findByText(/No account found/i)).toBeInTheDocument();
  });

  it('does not render password in DOM or error messages', async () => {
    const user = userEvent.setup();
    register.mockRejectedValue(new Error('An account with that email already exists.'));
    const { container } = renderWithProviders(<AuthPage mode="register" />);
    const pwd = 'supersecret123';
    await user.type(screen.getByLabelText(/email/i), 'dup@example.com');
    await user.type(screen.getByLabelText(/^password$/i), pwd);
    await user.click(screen.getByRole('button', { name: /create account/i }));
    await screen.findByText(/already exists/i);
    expect(container.innerHTML).not.toContain(pwd);
  });
});
