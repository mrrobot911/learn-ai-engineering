import { useEffect, useState, type ReactNode } from 'react';
import { CheckCircle2, XCircle, AlertTriangle, Info, X } from 'lucide-react';
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

type ToastKind = 'success' | 'error' | 'warning' | 'info';
interface ToastItem { id: number; kind: ToastKind; message: string; }

let toastSeq = 0;
const TOAST_LISTENERS = new Set<(items: ToastItem[]) => void>();
let currentToasts: ToastItem[] = [];

export function toast(message: string, kind: ToastKind = 'info') {
  const id = ++toastSeq;
  currentToasts = [...currentToasts, { id, kind, message }];
  TOAST_LISTENERS.forEach((fn) => fn(currentToasts));
  setTimeout(() => {
    currentToasts = currentToasts.filter((t) => t.id !== id);
    TOAST_LISTENERS.forEach((fn) => fn(currentToasts));
  }, 3200);
}

export function ToastHost() {
  const [items, setItems] = useState<ToastItem[]>([]);
  useEffect(() => {
    const fn = (next: ToastItem[]) => setItems(next);
    TOAST_LISTENERS.add(fn);
    return () => { TOAST_LISTENERS.delete(fn); };
  }, []);

  const icons: Record<ToastKind, typeof CheckCircle2> = {
    success: CheckCircle2, error: XCircle, warning: AlertTriangle, info: Info,
  };
  const colors: Record<ToastKind, string> = {
    success: 'bg-brand-50 text-brand-800 border-brand-200',
    error: 'bg-error-50 text-error-700 border-error-100',
    warning: 'bg-warning-50 text-warning-700 border-warning-100',
    info: 'bg-ink-50 text-ink-800 border-ink-200',
  };

  return (
    <div className="fixed bottom-5 right-5 z-50 flex flex-col gap-2 w-[min(92vw,360px)]">
      {items.map((t) => {
        const Icon = icons[t.kind];
        return (
          <div
            key={t.id}
            role="status"
            className={`flex items-start gap-2.5 rounded-lg border px-3.5 py-3 shadow-card-hover animate-slide-up ${colors[t.kind]}`}
          >
            <Icon className="w-4 h-4 mt-0.5 shrink-0" />
            <p className="text-sm flex-1 leading-snug">{t.message}</p>
            <button
              className="opacity-50 hover:opacity-100 transition-opacity"
              onClick={() => {
                currentToasts = currentToasts.filter((x) => x.id !== t.id);
                TOAST_LISTENERS.forEach((fn) => fn(currentToasts));
              }}
              aria-label="Dismiss"
            >
              <X className="w-3.5 h-3.5" />
            </button>
          </div>
        );
      })}
    </div>
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
