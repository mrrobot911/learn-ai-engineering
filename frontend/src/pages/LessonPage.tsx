import { useEffect, useMemo, useState } from 'react';
import {
  ArrowLeft, ArrowRight, Clock, ChevronLeft, ChevronRight,
  BookOpen, ListChecks,
} from 'lucide-react';
import { api } from '@/api';
import type { Lesson, Topic } from '@/api';
import { Link, Spinner, EmptyState, ErrorState, Badge } from '@/components/ui';
import { Markdown } from '@/components/Markdown';

// The client's not-found signal is an error whose status is 404; message text
// is never matched, so a real HTTP 404 maps onto the same check later (T27).
function isNotFound(error: unknown): boolean {
  return typeof error === 'object' && error !== null
    && (error as { status?: number }).status === 404;
}

export function LessonPage({ id }: { id: string }) {
  const [record, setRecord] = useState<{ lesson: Lesson; topic: Topic } | undefined>(undefined);
  const [notFound, setNotFound] = useState(false);
  const [loading, setLoading] = useState(true);
  const [errored, setErrored] = useState(false);
  const [reloadKey, setReloadKey] = useState(0);

  // The lesson and its topic both arrive through the API client: getLesson()
  // first, then getTopic() for the lesson's topicId. No fake latency — the
  // loading state lasts exactly as long as the client promises are pending.
  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setErrored(false);
    setNotFound(false);
    setRecord(undefined);
    (async () => {
      try {
        const lesson = await api.getLesson(id);
        const topic = await api.getTopic(lesson.topicId);
        if (!cancelled) {
          setRecord({ lesson, topic });
          setLoading(false);
        }
      } catch (error) {
        if (cancelled) return;
        if (isNotFound(error)) setNotFound(true);
        else setErrored(true);
        setLoading(false);
      }
    })();
    return () => { cancelled = true; };
  }, [id, reloadKey]);

  const { prev, next } = useMemo(() => {
    if (!record) return { prev: undefined, next: undefined };
    const lessons = record.topic.lessons;
    const idx = lessons.findIndex((l) => l.id === record.lesson.id);
    return {
      prev: idx > 0 ? lessons[idx - 1] : undefined,
      next: idx < lessons.length - 1 ? lessons[idx + 1] : undefined,
    };
  }, [record]);

  if (loading) return <LessonSkeleton />;

  if (errored) {
    return <ErrorState onRetry={() => setReloadKey((key) => key + 1)} />;
  }

  if (notFound || !record) {
    return (
      <EmptyState
        icon={BookOpen}
        title="Lesson not found"
        description="This lesson doesn't exist or isn't published."
        action={<Link to={{ name: 'topics', params: {} }} className="btn-secondary">Back to topics</Link>}
      />
    );
  }

  const { lesson, topic } = record;
  const idxInTopic = topic.lessons.findIndex((l) => l.id === lesson.id) + 1;

  return (
    <div className="animate-fade-in max-w-3xl mx-auto">
      {/* Breadcrumb */}
      <div className="flex items-center gap-1.5 text-sm text-ink-500 mb-5 flex-wrap">
        <Link to={{ name: 'topics', params: {} }} className="hover:text-ink-800 transition-colors">Topics</Link>
        <span className="text-ink-300">/</span>
        <Link to={{ name: 'topic', params: { slug: topic.slug } }} className="hover:text-ink-800 transition-colors">{topic.title}</Link>
        <span className="text-ink-300">/</span>
        <span className="text-ink-700 truncate">Lesson {idxInTopic}</span>
      </div>

      <article>
        <div className="flex items-center gap-3 text-xs text-ink-400 mb-3">
          <Badge tone="neutral"><ListChecks className="w-3 h-3" /> Lesson {idxInTopic} of {topic.lessons.length}</Badge>
          <span className="flex items-center gap-1"><Clock className="w-3 h-3" /> {lesson.durationMin} min read</span>
        </div>

        <h1 className="text-2xl sm:text-3xl font-bold text-ink-900 tracking-tight mb-3">{lesson.title}</h1>
        <p className="text-ink-600 text-base sm:text-lg leading-relaxed mb-6">{lesson.summary}</p>

        <div className="h-px bg-ink-200/70 mb-6" />

        <Markdown source={lesson.body} />
      </article>

      {/* Prev / Next */}
      <nav className="mt-6 grid grid-cols-1 sm:grid-cols-2 gap-3">
        {prev ? (
          <Link
            to={{ name: 'lesson', params: { id: prev.id } }}
            className="card p-4 hover:shadow-card-hover transition-all group flex items-center gap-3"
          >
            <ChevronLeft className="w-5 h-5 text-ink-400 group-hover:text-ink-700 shrink-0" />
            <div className="flex-1 min-w-0">
              <div className="text-[11px] text-ink-400 uppercase tracking-wider">Previous</div>
              <div className="text-sm font-medium text-ink-900 truncate">{prev.title}</div>
            </div>
          </Link>
        ) : <div />}
        {next ? (
          <Link
            to={{ name: 'lesson', params: { id: next.id } }}
            className="card p-4 hover:shadow-card-hover transition-all group flex items-center gap-3 text-right"
          >
            <div className="flex-1 min-w-0">
              <div className="text-[11px] text-ink-400 uppercase tracking-wider">Next</div>
              <div className="text-sm font-medium text-ink-900 truncate">{next.title}</div>
            </div>
            <ChevronRight className="w-5 h-5 text-ink-400 group-hover:text-ink-600 group-hover:translate-x-0.5 transition-all shrink-0" />
          </Link>
        ) : (
          <Link
            to={{ name: 'quiz-start', params: { slug: topic.slug } }}
            className="card p-4 hover:shadow-card-hover transition-all group flex items-center gap-3 bg-brand-50/60 border-brand-200/70"
          >
            <div className="flex-1 min-w-0">
              <div className="text-[11px] text-brand-700 uppercase tracking-wider">Up next</div>
              <div className="text-sm font-medium text-brand-800 truncate">Take the {topic.title} quiz</div>
            </div>
            <ArrowRight className="w-5 h-5 text-brand-600 group-hover:translate-x-0.5 transition-all shrink-0" />
          </Link>
        )}
      </nav>

      <div className="mt-6">
        <Link to={{ name: 'topic', params: { slug: topic.slug } }} className="inline-flex items-center gap-1.5 text-sm text-ink-500 hover:text-ink-800 transition-colors">
          <ArrowLeft className="w-4 h-4" /> Back to {topic.title}
        </Link>
      </div>
    </div>
  );
}

function LessonSkeleton() {
  return (
    <div className="max-w-3xl mx-auto">
      <div className="skeleton h-4 w-48 mb-5" />
      <div className="flex gap-2 mb-3">
        <div className="skeleton h-5 w-32" />
        <div className="skeleton h-5 w-20" />
      </div>
      <div className="skeleton h-8 w-3/4 mb-3" />
      <div className="skeleton h-4 w-full mb-6" />
      <div className="space-y-3">
        {Array.from({ length: 6 }).map((_, i) => (
          <div key={i} className="skeleton h-4 w-full" style={{ width: `${85 + (i % 3) * 5}%` }} />
        ))}
      </div>
      <div className="flex items-center justify-center gap-2 text-ink-400 text-sm py-6">
        <Spinner className="w-4 h-4" /> Loading lesson…
      </div>
    </div>
  );
}
