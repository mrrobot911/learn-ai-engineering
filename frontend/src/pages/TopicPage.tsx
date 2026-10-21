import { useEffect, useMemo, useState } from 'react';
import {
  ArrowLeft, ArrowRight, BookOpen, CheckCircle2, Clock, ListChecks, Trophy,
} from 'lucide-react';
import { api } from '@/api';
import type { Lesson, Topic } from '@/api';
import { useSession } from '@/session';
import { useProgress } from '@/hooks/useProgress';
import { Link, Spinner, EmptyState, ErrorState, Badge, ProgressBar } from '@/components/ui';
import { TopicIcon } from '@/components/TopicIcon';

const colorMap: Record<string, { bg: string; text: string; ring: string }> = {
  brand: { bg: 'bg-brand-50', text: 'text-brand-600', ring: 'ring-brand-100' },
  accent: { bg: 'bg-accent-50', text: 'text-accent-600', ring: 'ring-accent-100' },
  warning: { bg: 'bg-warning-50', text: 'text-warning-600', ring: 'ring-warning-100' },
};

export function TopicPage({ slug }: { slug: string }) {
  const { user } = useSession();
  const { ids, failed: progressFailed } = useProgress(user?.id);
  const [topic, setTopic] = useState<Topic | null>(null);
  const [notFound, setNotFound] = useState(false);
  const [loading, setLoading] = useState(true);
  const [errored, setErrored] = useState(false);
  const [reloadKey, setReloadKey] = useState(0);

  // The slug is resolved only through the API client: listTopics() finds the
  // summary, getTopic() loads the lessons for its id.
  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setErrored(false);
    setNotFound(false);
    setTopic(null);
    (async () => {
      try {
        const summaries = await api.listTopics();
        const summary = summaries.find((t) => t.slug === slug);
        if (!summary) {
          if (!cancelled) {
            setNotFound(true);
            setLoading(false);
          }
          return;
        }
        const loaded = await api.getTopic(summary.id);
        if (!cancelled) {
          setTopic(loaded);
          setLoading(false);
        }
      } catch {
        if (!cancelled) {
          setErrored(true);
          setLoading(false);
        }
      }
    })();
    return () => { cancelled = true; };
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
        action={<Link to={{ name: 'topics', params: {} }} className="btn-secondary">Back to topics</Link>}
      />
    );
  }

  const c = colorMap[topic.color] ?? colorMap.brand;
  const completedCount = topic.lessons.filter((l) => completedIds.has(l.id)).length;

  return (
    <div className="animate-fade-in">
      <Link to={{ name: 'topics', params: {} }} className="inline-flex items-center gap-1.5 text-sm text-ink-500 hover:text-ink-800 mb-5 transition-colors">
        <ArrowLeft className="w-4 h-4" /> All topics
      </Link>

      <div className="flex items-start gap-4 mb-6">
        <div className={`w-12 h-12 rounded-xl ${c.bg} ${c.text} flex items-center justify-center ring-4 ${c.ring} shrink-0`}>
          <TopicIcon name={topic.icon} className="w-6 h-6" />
        </div>
        <div className="flex-1 min-w-0">
          <h1 className="text-2xl sm:text-3xl font-bold text-ink-900 mb-1.5">{topic.title}</h1>
          <p className="text-ink-600 leading-relaxed max-w-2xl">{topic.description}</p>
        </div>
      </div>

      {/* Quiz CTA banner */}
      <div className="card p-5 mb-6 bg-gradient-to-br from-ink-900 to-ink-800 text-white border-ink-900">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div>
            <div className="flex items-center gap-2 text-xs font-mono text-ink-300 mb-1.5">
              <Trophy className="w-3.5 h-3.5" /> QUIZ
            </div>
            <h2 className="text-lg font-semibold text-white mb-1">Test your knowledge</h2>
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
          <div className="flex items-center justify-between mb-2 text-sm">
            <span className="text-ink-700 font-medium flex items-center gap-2">
              <CheckCircle2 className="w-4 h-4 text-brand-600" aria-hidden="true" /> Your progress
            </span>
            <span className="text-ink-500 text-xs">
              <span className="font-mono font-semibold text-ink-700">{completedCount} of {topic.lessons.length}</span> lessons done
            </span>
          </div>
          <ProgressBar value={completedCount} total={topic.lessons.length} />
        </div>
      )}

      {user && progressFailed && (
        <p role="status" className="card p-4 mb-6 text-sm text-ink-500">
          Your progress is unavailable right now. Reload the page to try again.
        </p>
      )}

      <h2 className="text-sm font-semibold text-ink-500 uppercase tracking-wider mb-3 flex items-center gap-2">
        <ListChecks className="w-4 h-4" /> Lessons
      </h2>

      {topic.lessons.length === 0 ? (
        <EmptyState
          icon={BookOpen}
          title="No lessons yet"
          description="Lessons for this topic will appear here once they are published."
        />
      ) : (
        <div className="space-y-2.5">
          {topic.lessons.map((lesson, idx) => (
            <LessonRow
              key={lesson.id}
              lesson={lesson}
              index={idx + 1}
              completed={completedIds.has(lesson.id)}
            />
          ))}
        </div>
      )}
    </div>
  );
}

function LessonRow({ lesson, index, completed }: { lesson: Lesson; index: number; completed: boolean }) {
  return (
    <Link
      to={{ name: 'lesson', params: { id: lesson.id } }}
      className="card p-4 hover:shadow-card-hover hover:border-ink-300/80 transition-all duration-200 group flex items-center gap-4"
    >
      <div className={`w-9 h-9 rounded-lg flex items-center justify-center text-sm font-mono shrink-0 ${
        completed ? 'bg-brand-50 text-brand-700' : 'bg-ink-100 text-ink-500'
      }`}>
        {completed
          ? <CheckCircle2 className="w-5 h-5" aria-hidden="true" />
          : <span>{index}</span>}
      </div>
      <div className="flex-1 min-w-0">
        <h3 className="text-sm sm:text-base font-medium text-ink-900 group-hover:text-ink-950 truncate">{lesson.title}</h3>
        <p className="text-xs sm:text-sm text-ink-500 truncate">{lesson.summary}</p>
      </div>
      <div className="flex items-center gap-3 shrink-0">
        <span className="hidden sm:flex items-center gap-1 text-xs text-ink-400">
          <Clock className="w-3 h-3" /> {lesson.durationMin}m
        </span>
        {completed && (
          <Badge tone="brand" aria-label="Completed">
            <CheckCircle2 className="w-3 h-3" aria-hidden="true" /> Completed
          </Badge>
        )}
        <ArrowRight className="w-4 h-4 text-ink-300 group-hover:text-ink-600 group-hover:translate-x-0.5 transition-all" />
      </div>
    </Link>
  );
}

function TopicSkeleton() {
  return (
    <div>
      <div className="skeleton h-4 w-28 mb-5" />
      <div className="flex items-start gap-4 mb-6">
        <div className="skeleton w-12 h-12 rounded-xl" />
        <div className="flex-1 space-y-2">
          <div className="skeleton h-7 w-2/5" />
          <div className="skeleton h-4 w-3/4" />
        </div>
      </div>
      <div className="skeleton h-24 w-full rounded-xl mb-6" />
      <div className="skeleton h-4 w-32 mb-3" />
      <div className="space-y-2.5">
        {Array.from({ length: 3 }).map((_, i) => (
          <div key={i} className="card p-4 flex items-center gap-4">
            <div className="skeleton w-9 h-9 rounded-lg" />
            <div className="flex-1 space-y-2">
              <div className="skeleton h-4 w-1/2" />
              <div className="skeleton h-3 w-3/4" />
            </div>
          </div>
        ))}
      </div>
      <div className="flex items-center justify-center gap-2 text-ink-400 text-sm py-4">
        <Spinner className="w-4 h-4" /> Loading topic…
      </div>
    </div>
  );
}
