import type { ReactNode } from 'react';
import { CheckCircle2, AlertTriangle } from 'lucide-react';
import type { Route } from '@/types';
import { routeToHash } from '@/router';

export function Spinner({ className = 'w-4 h-4' }: { className?: string }) {
  return (
    <svg className={`animate-spin ${className}`} viewBox="0 0 24 24" fill="none" aria-hidden="true">
      <circle className="opacity-20" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="3" />
      <path className="opacity-90" d="M22 12a10 10 0 0 1-10 10" stroke="currentColor" strokeWidth="3" strokeLinecap="round" />
    </svg>
  );
}

export function Link({
  to, children, className, onClick,
}: {
  to: Route; children: ReactNode; className?: string; onClick?: () => void;
}) {
  return (
    <a
      href={routeToHash(to)}
      onClick={onClick}
      className={className}
    >
      {children}
    </a>
  );
}

export function ProgressBar({ value, total, className = '' }: { value: number; total: number; className?: string }) {
  const pct = total > 0 ? Math.round((value / total) * 100) : 0;
  return (
    <div className={`h-2 bg-ink-100 rounded-full overflow-hidden ${className}`}>
      <div
        className="h-full bg-brand-500 rounded-full transition-all duration-500 ease-out"
        style={{ width: `${pct}%` }}
      />
    </div>
  );
}

export function Badge({ children, tone = 'neutral', className = '' }: {
  children: ReactNode;
  tone?: 'neutral' | 'brand' | 'accent' | 'warning' | 'success' | 'error';
  className?: string;
}) {
  const tones: Record<string, string> = {
    neutral: 'bg-ink-100 text-ink-600',
    brand: 'bg-brand-50 text-brand-700',
    accent: 'bg-accent-50 text-accent-700',
    warning: 'bg-warning-50 text-warning-700',
    success: 'bg-brand-50 text-brand-700',
    error: 'bg-error-50 text-error-700',
  };
  return <span className={`badge ${tones[tone]} ${className}`}>{children}</span>;
}

export function EmptyState({
  icon: Icon, title, description, action,
}: {
  icon: typeof CheckCircle2;
  title: string;
  description?: string;
  action?: ReactNode;
}) {
  return (
    <div className="card p-10 text-center flex flex-col items-center animate-fade-in">
      <div className="w-12 h-12 rounded-full bg-ink-100 flex items-center justify-center mb-4">
        <Icon className="w-6 h-6 text-ink-400" />
      </div>
      <h3 className="text-base text-ink-900 mb-1">{title}</h3>
      {description && <p className="text-sm text-ink-500 max-w-sm mb-4">{description}</p>}
      {action}
    </div>
  );
}

export function ErrorState({
  title = 'Something went wrong', description, onRetry,
}: {
  title?: string; description?: string; onRetry?: () => void;
}) {
  return (
    <div className="card p-10 text-center flex flex-col items-center animate-fade-in border-error-100">
      <div className="w-12 h-12 rounded-full bg-error-50 flex items-center justify-center mb-4">
        <AlertTriangle className="w-6 h-6 text-error-500" />
      </div>
      <h3 className="text-base text-ink-900 mb-1">{title}</h3>
      {description && <p className="text-sm text-ink-500 max-w-sm mb-4">{description}</p>}
      {onRetry && <button className="btn-secondary" onClick={onRetry}>Try again</button>}
    </div>
  );
}
