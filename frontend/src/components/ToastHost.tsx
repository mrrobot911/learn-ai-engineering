import { useEffect, useState } from 'react';
import { CheckCircle2, XCircle, AlertTriangle, Info, X } from 'lucide-react';
import { dismissToast, subscribeToToasts, type ToastItem, type ToastKind } from './toast';

const icons: Record<ToastKind, typeof CheckCircle2> = {
  success: CheckCircle2, error: XCircle, warning: AlertTriangle, info: Info,
};
const colors: Record<ToastKind, string> = {
  success: 'bg-brand-50 text-brand-800 border-brand-200',
  error: 'bg-error-50 text-error-700 border-error-100',
  warning: 'bg-warning-50 text-warning-700 border-warning-100',
  info: 'bg-ink-50 text-ink-800 border-ink-200',
};

export function ToastHost() {
  const [items, setItems] = useState<ToastItem[]>([]);
  useEffect(() => subscribeToToasts(setItems), []);

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
              onClick={() => dismissToast(t.id)}
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
