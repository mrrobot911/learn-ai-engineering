import { useEffect, useState } from 'react';
import { ArrowRight, BookOpen } from 'lucide-react';
import { api } from '@/api';
import type { TopicSummary } from '@/api';
import { Link, Spinner, EmptyState, ErrorState } from '@/components/ui';
import { TopicIcon } from '@/components/TopicIcon';

const colorMap: Record<string, { bg: string; text: string; ring: string }> = {
  brand: { bg: 'bg-brand-50', text: 'text-brand-600', ring: 'ring-brand-100' },
  accent: { bg: 'bg-accent-50', text: 'text-accent-600', ring: 'ring-accent-100' },
  warning: { bg: 'bg-warning-50', text: 'text-warning-600', ring: 'ring-warning-100' },
};

export function TopicsPage() {
  const [topics, setTopics] = useState<TopicSummary[]>([]);
  const [loading, setLoading] = useState(true);
  const [errored, setErrored] = useState(false);
  const [reloadKey, setReloadKey] = useState(0);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setErrored(false);
    api.listTopics().then(
      (data) => {
        if (cancelled) return;
        setTopics(data);
        setLoading(false);
      },
      () => {
        if (cancelled) return;
        setErrored(true);
        setLoading(false);
      },
    );
    return () => { cancelled = true; };
  }, [reloadKey]);

  return (
    <div className="animate-fade-in">
      <section className="mb-8 sm:mb-10">
        <h1 className="text-3xl sm:text-4xl font-bold text-ink-900 tracking-tight mb-3">
          Learn AI engineering, one topic at a time.
        </h1>
        <p className="text-ink-600 text-base sm:text-lg max-w-2xl leading-relaxed">
          Structured lessons on LLMs, agents, MCP, RAG, embeddings, evaluation, and deployment —
          each paired with a short quiz to check what you actually understood.
        </p>
      </section>

      {loading ? (
        <TopicsSkeleton />
      ) : errored ? (
        <ErrorState onRetry={() => setReloadKey((key) => key + 1)} />
      ) : topics.length === 0 ? (
        <EmptyState
          icon={BookOpen}
          title="No topics yet"
          description="Published topics will appear here as soon as they are added."
        />
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {topics.map((t) => {
            const c = colorMap[t.color] ?? colorMap.brand;
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
                      <h2 className="text-base font-semibold text-ink-900 group-hover:text-ink-950 transition-colors truncate">
                        {t.title}
                      </h2>
                      <ArrowRight className="w-4 h-4 text-ink-300 group-hover:text-ink-600 group-hover:translate-x-0.5 transition-all shrink-0" />
                    </div>
                    <p className="text-sm text-ink-500 leading-relaxed line-clamp-2">{t.description}</p>
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
