import { useEffect, useState } from 'react';
import { ArrowRight, Clock, FileText, CircleDot, Trophy } from 'lucide-react';
import type { Topic } from '@/types';
import { topics } from '@/data/content';
import { useStore } from '@/store';
import { Link, Spinner, ErrorState } from '@/components/ui';
import { TopicIcon } from '@/components/TopicIcon';

const colorMap: Record<string, { bg: string; text: string; ring: string }> = {
  brand: { bg: 'bg-brand-50', text: 'text-brand-600', ring: 'ring-brand-100' },
  accent: { bg: 'bg-accent-50', text: 'text-accent-600', ring: 'ring-accent-100' },
  warning: { bg: 'bg-warning-50', text: 'text-warning-600', ring: 'ring-warning-100' },
};

export function TopicsPage() {
  const { user, completedLessons, quizHistory } = useStore();
  const [loading, setLoading] = useState(true);
  const [errored, setErrored] = useState(false);
  const [data, setData] = useState<Topic[]>([]);

  useEffect(() => {
    setLoading(true);
    setErrored(false);
    // Simulate async fetch to demonstrate loading states.
    const t = setTimeout(() => {
      setData(topics);
      setLoading(false);
    }, 450);
    return () => clearTimeout(t);
  }, []);

  function retry() {
    setLoading(true);
    setErrored(false);
    setTimeout(() => { setData(topics); setLoading(false); }, 450);
  }

  const totalLessons = topics.reduce((s, t) => s + t.lessons.length, 0);

  return (
    <div className="animate-fade-in">
      <section className="mb-8 sm:mb-10">
        <div className="flex items-center gap-2 text-xs font-mono text-ink-400 mb-3">
          <CircleDot className="w-3 h-3 text-brand-500" />
          <span>{topics.length} topics · {totalLessons} lessons · quizzes included</span>
        </div>
        <h1 className="text-3xl sm:text-4xl font-bold text-ink-900 tracking-tight mb-3">
          Learn AI engineering, one topic at a time.
        </h1>
        <p className="text-ink-600 text-base sm:text-lg max-w-2xl leading-relaxed">
          Structured lessons on LLMs, agents, MCP, RAG, embeddings, evaluation, and deployment —
          each paired with a short quiz to check what you actually understood.
        </p>

        {user && (
          <div className="mt-5 flex flex-wrap items-center gap-3 text-sm">
            <div className="flex items-center gap-2 px-3 py-1.5 rounded-full bg-white border border-ink-200/70">
              <FileText className="w-3.5 h-3.5 text-brand-600" />
              <span className="text-ink-700"><b className="text-ink-900">{completedLessons.length}</b> lessons completed</span>
            </div>
            <div className="flex items-center gap-2 px-3 py-1.5 rounded-full bg-white border border-ink-200/70">
              <Trophy className="w-3.5 h-3.5 text-accent-600" />
              <span className="text-ink-700"><b className="text-ink-900">{quizHistory.length}</b> quiz attempts</span>
            </div>
          </div>
        )}
      </section>

      {loading ? (
        <TopicsSkeleton />
      ) : errored ? (
        <ErrorState onRetry={retry} />
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {data.map((t) => {
            const c = colorMap[t.color] ?? colorMap.brand;
            const completedInTopic = user
              ? t.lessons.filter((l) => completedLessons.includes(l.id)).length
              : 0;
            const attemptCount = quizHistory.filter((a) => a.topicId === t.id).length;
            return (
              <Link
                key={t.id}
                to={{ name: 'topic', params: { slug: t.slug } }}
                className="card p-5 hover:shadow-card-hover hover:border-ink-300/80 transition-all duration-200 group animate-fade-in"
              >
                <div className="flex items-start gap-4">
                  <div className={`w-11 h-11 rounded-xl ${c.bg} ${c.text} flex items-center justify-center ring-4 ${c.ring} shrink-0`}>
                    <TopicIcon name={t.icon} className="w-5 h-5" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center justify-between gap-2 mb-1">
                      <h3 className="text-base font-semibold text-ink-900 group-hover:text-ink-950 transition-colors truncate">
                        {t.title}
                      </h3>
                      <ArrowRight className="w-4 h-4 text-ink-300 group-hover:text-ink-600 group-hover:translate-x-0.5 transition-all shrink-0" />
                    </div>
                    <p className="text-sm text-ink-500 leading-relaxed mb-3 line-clamp-2">{t.description}</p>
                    <div className="flex items-center gap-3 text-xs text-ink-400">
                      <span className="flex items-center gap-1"><FileText className="w-3 h-3" /> {t.lessons.length} lessons</span>
                      {user && completedInTopic > 0 && (
                        <span className="flex items-center gap-1 text-brand-600">
                          <CircleDot className="w-3 h-3" /> {completedInTopic}/{t.lessons.length} done
                        </span>
                      )}
                      {attemptCount > 0 && (
                        <span className="flex items-center gap-1 text-accent-600">
                          <Trophy className="w-3 h-3" /> {attemptCount} attempt{attemptCount > 1 ? 's' : ''}
                        </span>
                      )}
                    </div>
                  </div>
                </div>
              </Link>
            );
          })}
        </div>
      )}
    </div>
  );
}

function TopicsSkeleton() {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
      {Array.from({ length: 6 }).map((_, i) => (
        <div key={i} className="card p-5">
          <div className="flex items-start gap-4">
            <div className="skeleton w-11 h-11 rounded-xl" />
            <div className="flex-1 space-y-2.5">
              <div className="skeleton h-4 w-2/3" />
              <div className="skeleton h-3 w-full" />
              <div className="skeleton h-3 w-4/5" />
              <div className="flex gap-3 pt-1">
                <div className="skeleton h-3 w-20" />
                <div className="skeleton h-3 w-16" />
              </div>
            </div>
          </div>
        </div>
      ))}
      <div className="col-span-full flex items-center justify-center gap-2 text-ink-400 text-sm py-2">
        <Spinner className="w-4 h-4" /> Loading topics…
      </div>
    </div>
  );
}
