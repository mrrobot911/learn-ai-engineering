export type ToastKind = 'success' | 'error' | 'warning' | 'info';
export interface ToastItem { id: number; kind: ToastKind; message: string; }

let toastSeq = 0;
let currentToasts: ToastItem[] = [];
const TOAST_LISTENERS = new Set<(items: ToastItem[]) => void>();

export function toast(message: string, kind: ToastKind = 'info') {
  const id = ++toastSeq;
  currentToasts = [...currentToasts, { id, kind, message }];
  TOAST_LISTENERS.forEach((fn) => fn(currentToasts));
  setTimeout(() => {
    currentToasts = currentToasts.filter((t) => t.id !== id);
    TOAST_LISTENERS.forEach((fn) => fn(currentToasts));
  }, 3200);
}

/** Subscribes to the toast queue; the returned function unsubscribes. */
export function subscribeToToasts(fn: (items: ToastItem[]) => void): () => void {
  TOAST_LISTENERS.add(fn);
  return () => { TOAST_LISTENERS.delete(fn); };
}

/** Removes one toast from the queue (used by the host's dismiss button). */
export function dismissToast(id: number) {
  currentToasts = currentToasts.filter((t) => t.id !== id);
  TOAST_LISTENERS.forEach((fn) => fn(currentToasts));
}
