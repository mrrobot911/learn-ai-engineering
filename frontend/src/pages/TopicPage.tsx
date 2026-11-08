import { useEffect, useMemo, useState } from 'react';
import {
  ArrowLeft,
  ArrowRight,
  BookOpen,
  CheckCircle2,
  Clock,
  ListChecks,
  Trophy,
} from 'lucide-react';
import { api } from '@/api';
import type { Lesson, Topic } from '@/api';
import { useSession } from '@/session';
import { useProgress } from '@/hooks/useProgress';
import {
  Link,
  Spinner,
  EmptyState,
  ErrorState,
  Badge,
  ProgressBar,
} from '@/components/ui';
import { TopicIcon } from '@/components/TopicIcon';

const colorMap: Record<string, { bg: string; text: string; ring: string }> = {
  brand: { bg: 'bg-brand-50', text: 'text-brand-600', ring: 'ring-brand-100' },
  accent: {
    bg: 'bg-accent-50',
    text: 'text-accent-600',
    ring: 'ring-accent-100',
  },
  warning: {
    bg: 'bg-warning-50',
    text: 'text-warning-600',
    ring: 'ring-warning-100',
  },
};

function isNotFound(error: unknown): boolean {
  return (
    typeof error === 'object' &&
    error !== null &&
    (error as { status?: number }).status === 404
  );
}

export function TopicPage({ slug }: { slug: string }) {
  const { user } = useSession();
  const { ids, failed: progressFailed } = useProgress(user?.id);
  const [topic, setTopic] = useState<Topic | null>(null);
  const [notFound, setNotFound] = useState(false);
  const [loading, setLoading] = useState(true);
  const [errored, setErrored] = useState(false);
  const [reloadKey, setReloadKey] = useState(0);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setErrored(false);
    setNotFound(false);
    setTopic(null);
    (async () => {
      try {
        const loaded = await api.getTopic(slug);
        if (!cancelled) {
          setTopic(loaded);
          setLoading(false);
        }
      } catch (error) {
        if (cancelled) return;
        if (isNotFound(error)) setNotFound(true);
        else setErrored(true);
        setLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [slug, reloadKey]);

  const completedIds = useMemo(() => new Set(ids ?? []), [ids]);

  if (loading) return <TopicSkeleton />;

  if (errored) {
    return <ErrorState onRetry={() => setReloadKey((key) => key + 1)} />;
  }

  if (notFound || !topic) {
    return (
      <EmptyState
        icon={BookOpen}
        title="Topic not found"
        description="This topic doesn't exist or isn't published."
        action={
          <Link to={{ name: 'topics', params: {} }} className="btn-secondary">
            Back to topics
          </Link>
        }
      />
    );
  }

  const c = colorMap[topic.color] ?? colorMap.brand;
  const completedCount = topic.lessons.filter((l) =>
    completedIds.has(l.id),
  ).length;

  return (
    <div className="animate-fade-in">
      <Link
        to={{ name: 'topics', params: {} }}
        className="inline-flex items-center gap-1.5 text-sm text-ink-500 hover:text-ink-800 mb-5 transition-colors"
      >
        <ArrowLeft className="w-4 h-4" /> All topics
      </Link>

      <div className="flex items-start gap-4 mb-6">
        <div
          className={`w-12 h-12 rounded-xl ${c.bg} ${c.text} flex items-center justify-center ring-4 ${c.ring} shrink-0`}
        >
          <TopicIcon name={topic.icon} className="w-6 h-6" />
        </div>
        <div className="flex-1 min-w-0">
          <h1 className="text-2xl sm:text-3xl font-bold text-ink-900 mb-1.5">
            {topic.title}
          </h1>
          <p className="text-ink-600 leading-relaxed max-w-2xl">
            {topic.description}
          </p>
        </div>
      </div>

      {/* Quiz CTA banner */}
      <div className="card p-5 mb-6 bg-gradient-to-br from-ink-900 to-ink-800 text-white border-ink-900">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div>
            <div className="flex items-center gap-2 text-xs font-mono text-ink-300 mb-1.5">
              <Trophy className="w-3.5 h-3.5" /> QUIZ
            </div>
            <h2 className="text-lg font-semibold text-white mb-1">
              Test your knowledge
            </h2>
            <p className="text-sm text-ink-300">
              Take a short multiple-choice quiz on this topic.
            </p>
          </div>
          <Link
            to={{ name: 'quiz-start', params: { slug: topic.slug } }}
            className="btn bg-white text-ink-900 hover:bg-ink-100 shrink-0"
          >
            Start quiz <ArrowRight className="w-4 h-4" />
          </Link>
        </div>
      </div>

      {/* Progress for logged-in users */}
      {user && ids && (
        <div className="card p-4 mb-6">
          <div className="flex items-center justify-between text-sm mb-2">
            <span className="font-medium text-ink-700">Your progress</span>
            <span className="font-mono text-ink-600">
              {completedCount} of {topic.lessons.length}
            </span>
          </div>
          <ProgressBar value={completedCount} total={topic.lessons.length} />
        </div>
      )}
      {user && progressFailed && (
        <p role="status" className="text-sm text-ink-500 mb-6">
          Your progress is unavailable right now. Reload the page to try again.
        </p>
      )}

      {/* Lesson list */}
      <div className="flex items-center justify-between mb-3">
        <h2 className="text-lg font-semibold text-ink-900 flex items-center gap-2">
          <ListChecks className="w-5 h-5 text-ink-400" /> Lessons
        </h2>
        <Badge tone="neutral">{topic.lessons.length} lessons</Badge>
      </div>

      {topic.lessons.length === 0 ? (
        <EmptyState
          icon={BookOpen}
          title="No lessons yet"
          description="Lessons for this topic haven't been published."
        />
      ) : (
        <div className="space-y-2.5">
          {topic.lessons.map((lesson, i) => (
            <LessonRow
              key={lesson.id}
              lesson={lesson}
              index={i}
              completed={completedIds.has(lesson.id)}
            />
          ))}
        </div>
      )}
    </div>
  );
}

function LessonRow({
  lesson,
  index,
  completed,
}: {
  lesson: Lesson;
  index: number;
  completed: boolean;
}) {
  return (
    <Link
      to={{ name: 'lesson', params: { id: lesson.id } }}
      className="card p-4 hover:shadow-card-hover hover:border-ink-300/80 transition-all group flex items-center gap-4"
    >
      <div
        className={`w-9 h-9 rounded-lg flex items-center justify-center text-sm font-mono font-bold shrink-0 ${
          completed ? 'bg-brand-100 text-brand-700' : 'bg-ink-100 text-ink-600'
        }`}
      >
        {completed ? (
          <CheckCircle2 className="w-4.5 h-4.5" aria-hidden="true" />
        ) : (
          index + 1
        )}
      </div>
      <div className="flex-1 min-w-0">
        <h3 className="text-sm font-medium text-ink-900 group-hover:text-ink-950 truncate">
          {lesson.title}
        </h3>
        <p className="text-xs text-ink-500 line-clamp-1 mt-0.5">
          {lesson.summary}
        </p>
      </div>
      <div className="flex items-center gap-3 shrink-0">
        {completed && (
          <Badge tone="success">
            <span aria-label="Completed">Completed</span>
          </Badge>
        )}
        <span className="hidden sm:inline-flex items-center gap-1 text-xs text-ink-400">
          <Clock className="w-3 h-3" /> {lesson.durationMin}m
        </span>
        <ArrowRight className="w-4 h-4 text-ink-300 group-hover:text-ink-600 group-hover:translate-x-0.5 transition-all" />
      </div>
    </Link>
  );
}

function TopicSkeleton() {
  return (
    <div>
      <div className="skeleton h-4 w-24 mb-5" />
      <div className="flex items-start gap-4 mb-6">
        <div className="skeleton w-12 h-12 rounded-xl" />
        <div className="flex-1 space-y-2">
          <div className="skeleton h-7 w-1/2" />
          <div className="skeleton h-4 w-full" />
          <div className="skeleton h-4 w-3/4" />
        </div>
      </div>
      <div className="skeleton h-28 w-full rounded-xl mb-6" />
      <div className="space-y-2.5">
        {Array.from({ length: 3 }).map((_, i) => (
          <div key={i} className="skeleton h-16 rounded-xl" />
        ))}
      </div>
      <div className="flex items-center justify-center gap-2 text-ink-400 text-sm py-4">
        <Spinner className="w-4 h-4" /> Loading topic…
      </div>
    </div>
  );
}
