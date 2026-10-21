import { useCallback, useEffect, useState } from 'react';
import { api } from '@/api';

/**
 * The lessons the logged-in user has completed, read through the client
 * (product-spec.md section 4 US-5). `ids` is null until the list has arrived
 * and when the call failed; `failed` tells the two apart, so a screen can
 * show its content without the completion control.
 */
export function useProgress(userId: string | undefined) {
  const [ids, setIds] = useState<string[] | null>(null);
  const [failed, setFailed] = useState(false);

  useEffect(() => {
    if (!userId) {
      setIds(null);
      setFailed(false);
      return;
    }
    let cancelled = false;
    (async () => {
      try {
        const list = await api.getProgress();
        if (!cancelled) {
          setIds(list);
          setFailed(false);
        }
      } catch {
        if (!cancelled) {
          setIds(null);
          setFailed(true);
        }
      }
    })();
    return () => { cancelled = true; };
  }, [userId]);

  /** Marks the lesson complete or undoes it; the list follows the client. */
  const setComplete = useCallback(async (lessonId: string, complete: boolean) => {
    if (complete) await api.markLessonComplete(lessonId);
    else await api.undoLessonComplete(lessonId);
    setIds((prev) => {
      if (prev === null) return prev;
      return complete
        ? [...new Set([...prev, lessonId])]
        : prev.filter((id) => id !== lessonId);
    });
  }, []);

  return { ids, failed, setComplete };
}
