import { useEffect, useState, type ReactNode } from 'react';
import { ArrowRight, BookOpen, Mail, Lock, AlertCircle } from 'lucide-react';
import { useSession } from '@/session';
import { Link, Spinner } from '@/components/ui';
import { toast } from '@/components/toast';

type FieldErrors = { email?: string; password?: string };

function emailValid(email: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
}

export function AuthPage({ mode }: { mode: 'login' | 'register' }) {
  const { login, register, user } = useSession();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [errors, setErrors] = useState<FieldErrors>({});
  const [formError, setFormError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const isRegister = mode === 'register';

  useEffect(() => {
    if (user) window.location.hash = '#/topics';
  }, [user]);

  function validate(): boolean {
    const next: FieldErrors = {};
    if (!email.trim()) {
      next.email = 'Email is required.';
    } else if (!emailValid(email)) {
      next.email = 'Enter a valid email address.';
    }
    if (!password) {
      next.password = 'Password is required.';
    } else if (password.length < 8 || password.length > 128) {
      next.password = 'Password must be between 8 and 128 characters.';
    }
    setErrors(next);
    return Object.keys(next).length === 0;
  }

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setFormError(null);
    if (!validate()) {
      setPassword('');
      return;
    }
    setSubmitting(true);
    try {
      if (isRegister) {
        const res = await register(email, password);
        setSubmitting(false);
        if (!res.ok) {
          setFormError(res.error || 'Registration failed.');
          setPassword('');
          return;
        }
        setPassword('');
        toast('Welcome. Your account is ready.', 'success');
        window.location.hash = '#/topics';
      } else {
        const res = await login(email, password);
        setSubmitting(false);
        if (!res.ok) {
          setFormError(res.error || 'Login failed.');
          setPassword('');
          return;
        }
        setPassword('');
        toast('Logged in successfully.', 'success');
        window.location.hash = '#/topics';
      }
    } catch (err) {
      setSubmitting(false);
      setFormError(err instanceof Error ? err.message : 'Something went wrong.');
      setPassword('');
    }
  }

  return (
    <div className="max-w-md mx-auto animate-fade-in">
      <div className="card p-6 sm:p-8">
        <div className="flex items-center gap-2 text-xs font-mono text-ink-400 mb-4 justify-center">
          <BookOpen className="w-3.5 h-3.5" /> Forge
        </div>
        <h1 className="text-xl font-bold text-ink-900 text-center mb-1">
          {isRegister ? 'Create your account' : 'Welcome back'}
        </h1>
        <p className="text-sm text-ink-500 text-center mb-6">
          {isRegister ? 'Save your lesson progress and quiz history.' : 'Log in to pick up where you left off.'}
        </p>

        {formError && (
          <div role="alert" className="mb-4 p-3 rounded-lg bg-error-50 border border-error-100 text-error-700 text-sm flex items-start gap-2">
            <AlertCircle className="w-4 h-4 mt-0.5 shrink-0" />
            <span>{formError}</span>
          </div>
        )}

        <form onSubmit={onSubmit} noValidate className="space-y-4">
          <Field label="Email" id="email" icon={<Mail className="w-4 h-4" />} error={errors.email}>
            <input
              id="email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="you@example.com"
              className={`input pl-10 ${errors.email ? 'input-error' : ''}`}
              autoComplete="email"
              disabled={submitting}
              aria-invalid={errors.email ? 'true' : 'false'}
            />
          </Field>

          <Field label="Password" id="password" icon={<Lock className="w-4 h-4" />} error={errors.password}>
            <input
              id="password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="••••••••"
              className={`input pl-10 ${errors.password ? 'input-error' : ''}`}
              autoComplete={isRegister ? 'new-password' : 'current-password'}
              disabled={submitting}
              aria-invalid={errors.password ? 'true' : 'false'}
            />
          </Field>

          <button type="submit" className="btn-brand w-full" disabled={submitting}>
            {submitting ? (
              <>
                <Spinner className="w-4 h-4" /> {isRegister ? 'Creating account…' : 'Logging in…'}
              </>
            ) : (
              <>
                {isRegister ? 'Create account' : 'Log in'} <ArrowRight className="w-4 h-4" />
              </>
            )}
          </button>
        </form>

        <p className="text-sm text-ink-500 text-center mt-5">
          {isRegister ? (
            <>
              Already have an account?{' '}
              <Link to={{ name: 'login', params: {} }} className="text-brand-600 font-medium hover:text-brand-700">
                Log in
              </Link>
            </>
          ) : (
            <>
              New here?{' '}
              <Link to={{ name: 'register', params: {} }} className="text-brand-600 font-medium hover:text-brand-700">
                Create an account
              </Link>
            </>
          )}
        </p>
      </div>
    </div>
  );
}

function Field({
  label,
  id,
  icon,
  error,
  children,
}: {
  label: string;
  id: string;
  icon: ReactNode;
  error?: string;
  children: ReactNode;
}) {
  return (
    <div>
      <label htmlFor={id} className="block text-sm font-medium text-ink-700 mb-1.5">
        {label}
      </label>
      <div className="relative">
        <span className="absolute left-3 top-1/2 -translate-y-1/2 text-ink-400">{icon}</span>
        {children}
      </div>
      {error && (
        <p className="mt-1.5 text-xs text-error-600 flex items-center gap-1">
          <AlertCircle className="w-3 h-3" /> {error}
        </p>
      )}
    </div>
  );
}
