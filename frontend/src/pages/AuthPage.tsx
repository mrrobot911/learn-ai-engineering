import { useEffect, useState, type ReactNode } from 'react';
import { ArrowRight, BookOpen, Mail, Lock, User as UserIcon, AlertCircle, Check, Info } from 'lucide-react';
import { useStore } from '@/store';
import { Link, Spinner, toast } from '@/components/ui';

type FieldErrors = { name?: string; email?: string; password?: string; confirm?: string };

function emailValid(email: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
}

export function AuthPage({ mode }: { mode: 'login' | 'register' }) {
  const { login, register, user } = useStore();
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirm, setConfirm] = useState('');
  const [errors, setErrors] = useState<FieldErrors>({});
  const [formError, setFormError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const isRegister = mode === 'register';

  // If already logged in, bounce to topics.
  useEffect(() => {
    if (user) window.location.hash = '#/topics';
  }, [user]);

  function validate(): boolean {
    const next: FieldErrors = {};
    if (isRegister) {
      if (!name.trim()) next.name = 'Name is required.';
      else if (name.trim().length < 2) next.name = 'Name must be at least 2 characters.';
    }
    if (!email.trim()) next.email = 'Email is required.';
    else if (!emailValid(email)) next.email = 'Enter a valid email address.';
    if (!password) next.password = 'Password is required.';
    else if (password.length < 8) next.password = 'Password must be at least 8 characters.';
    if (isRegister && password && confirm !== password) next.confirm = 'Passwords do not match.';
    setErrors(next);
    return Object.keys(next).length === 0;
  }

  function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setFormError(null);
    if (!validate()) return;
    setSubmitting(true);
    // Simulate latency.
    setTimeout(() => {
      if (isRegister) {
        const res = register(name, email, password);
        setSubmitting(false);
        if (!res.ok) { setFormError(res.error || 'Registration failed.'); return; }
        toast(`Welcome, ${name.trim()}. Your account is ready.`, 'success');
        window.location.hash = '#/topics';
      } else {
        const res = login(email, password);
        setSubmitting(false);
        if (!res.ok) { setFormError(res.error || 'Login failed.'); return; }
        toast('Logged in successfully.', 'success');
        window.location.hash = '#/topics';
      }
    }, 600);
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
          {isRegister
            ? 'Save your lesson progress and quiz history.'
            : 'Log in to pick up where you left off.'}
        </p>

        {formError && (
          <div className="mb-4 p-3 rounded-lg bg-error-50 border border-error-100 text-error-700 text-sm flex items-start gap-2">
            <AlertCircle className="w-4 h-4 mt-0.5 shrink-0" />
            <span>{formError}</span>
          </div>
        )}

        <form onSubmit={onSubmit} noValidate className="space-y-4">
          {isRegister && (
            <Field
              label="Name"
              icon={<UserIcon className="w-4 h-4" />}
              error={errors.name}
            >
              <input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="Ada Lovelace"
                className={`input pl-10 ${errors.name ? 'input-error' : ''}`}
                autoComplete="name"
                disabled={submitting}
              />
            </Field>
          )}

          <Field
            label="Email"
            icon={<Mail className="w-4 h-4" />}
            error={errors.email}
          >
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="you@example.com"
              className={`input pl-10 ${errors.email ? 'input-error' : ''}`}
              autoComplete="email"
              disabled={submitting}
            />
          </Field>

          <Field
            label="Password"
            icon={<Lock className="w-4 h-4" />}
            error={errors.password}
            hint={isRegister ? 'At least 8 characters.' : undefined}
          >
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="••••••••"
              className={`input pl-10 ${errors.password ? 'input-error' : ''}`}
              autoComplete={isRegister ? 'new-password' : 'current-password'}
              disabled={submitting}
            />
          </Field>

          {isRegister && (
            <Field
              label="Confirm password"
              icon={<Lock className="w-4 h-4" />}
              error={errors.confirm}
            >
              <input
                type="password"
                value={confirm}
                onChange={(e) => setConfirm(e.target.value)}
                placeholder="••••••••"
                className={`input pl-10 ${errors.confirm ? 'input-error' : ''}`}
                autoComplete="new-password"
                disabled={submitting}
              />
            </Field>
          )}

          <button type="submit" className="btn-brand w-full" disabled={submitting}>
            {submitting
              ? <><Spinner className="w-4 h-4" /> {isRegister ? 'Creating account…' : 'Logging in…'}</>
              : <>{isRegister ? 'Create account' : 'Log in'} <ArrowRight className="w-4 h-4" /></>}
          </button>
        </form>

        <p className="text-sm text-ink-500 text-center mt-5">
          {isRegister ? (
            <>Already have an account? <Link to={{ name: 'login', params: {} }} className="text-brand-600 font-medium hover:text-brand-700">Log in</Link></>
          ) : (
            <>New here? <Link to={{ name: 'register', params: {} }} className="text-brand-600 font-medium hover:text-brand-700">Create an account</Link></>
          )}
        </p>
      </div>

      {isRegister && (
        <div className="mt-4 p-3 rounded-lg bg-accent-50 border border-accent-100 text-accent-700 text-xs flex items-start gap-2">
          <Info className="w-3.5 h-3.5 mt-0.5 shrink-0" />
          <span>Authentication is simulated for this prototype — no real account is created and no data leaves your browser.</span>
        </div>
      )}

      {!isRegister && (
        <div className="mt-4 card p-4">
          <div className="flex items-start gap-2.5">
            <div className="w-7 h-7 rounded-lg bg-brand-50 text-brand-700 flex items-center justify-center shrink-0">
              <Check className="w-4 h-4" />
            </div>
            <div className="flex-1">
              <p className="text-xs font-medium text-ink-900">Try the demo account</p>
              <p className="text-xs text-ink-500 mb-2">Pre-loaded with quiz history and lesson progress.</p>
              <button
                type="button"
                className="btn-secondary text-xs py-1.5 px-3"
                onClick={() => {
                  setEmail('demo@forge.dev');
                  setPassword('password123');
                  setFormError(null);
                  toast('Demo credentials filled in. Click Log in.', 'info');
                }}
              >
                Fill demo credentials
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

function Field({
  label, icon, error, hint, children,
}: {
  label: string; icon: ReactNode; error?: string; hint?: string; children: ReactNode;
}) {
  return (
    <div>
      <label className="block text-sm font-medium text-ink-700 mb-1.5">{label}</label>
      <div className="relative">
        <span className="absolute left-3 top-1/2 -translate-y-1/2 text-ink-400">{icon}</span>
        {children}
      </div>
      {error ? (
        <p className="mt-1.5 text-xs text-error-600 flex items-center gap-1">
          <AlertCircle className="w-3 h-3" /> {error}
        </p>
      ) : hint ? (
        <p className="mt-1.5 text-xs text-ink-400">{hint}</p>
      ) : null}
    </div>
  );
}
