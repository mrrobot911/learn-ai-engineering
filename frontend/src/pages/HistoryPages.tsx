import { useEffect, useMemo, useState } from 'react';
import {
  ArrowLeft, ArrowRight, Trophy, History, CheckCircle2, XCircle, Clock,
} from 'lucide-react';
import type { QuizAttempt } from '@/types';
import { useStore } from '@/useStore';
import { getTopicById } from '@/data/content';
import { Link, Spinner, EmptyState, Badge } from '@/components/ui';
import { TopicIcon } from '@/components/TopicIcon';

export function HistoryPage() {
  const { user, quizHistory } = useStore();
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const t = setTimeout(() => setLoading(false), 350);
    return () => clearTimeout(t);
  }, []);

  const stats = useMemo(() => {
    if (quizHistory.length === 0) return { total: 0, avgPct: 0, bestPct: 0 };
    const attempts = quizHistory as QuizAttempt[];
    const pcts = attempts.map((a: QuizAttempt) => a.score / a.total);
    const avg = (pcts as number[]).reduce((s: number, x: number) => s + x, 0) / pcts.length;
    const best = Math.max(...(pcts as number[]));
    return {
      total: quizHistory.length,
      avgPct: Math.round(avg * 100),
      bestPct: Math.round(best * 100),
    };
  }, [quizHistory]);

  if (loading) return <HistorySkeleton />;

  if (!user) {
    return (
      <EmptyState
        icon={History}
        title="Log in to see your history"
        description="Quiz attempts are only saved for registered users. Log in or create an account to start tracking your progress."
        action={
          <div className="flex gap-2">
            <Link to={{ name: 'login', params: {} }} className="btn-primary">Log in</Link>
            <Link to={{ name: 'register', params: {} }} className="btn-secondary">Sign up</Link>
          </div>
        }
      />
    );
  }

  if (quizHistory.length === 0) {
    return (
      <EmptyState
        icon={Trophy}
        title="No quiz attempts yet"
        description="Take a quiz on any topic and your results will appear here, with scores and per-question breakdowns."
        action={<Link to={{ name: 'topics', params: {} }} className="btn-brand">Browse topics</Link>}
      />
    );
  }

  return (
    <div className="animate-fade-in">
      <div className="flex items-center gap-2 text-xs font-mono text-ink-400 mb-3">
        <History className="w-3 h-3" /> {user.email}'s quiz history
      </div>
      <h1 className="text-2xl sm:text-3xl font-bold text-ink-900 tracking-tight mb-1">Quiz history</h1>
      <p className="text-ink-600 mb-6">Every attempt is saved with a snapshot of the questions and answers at the time.</p>

      <div className="grid grid-cols-3 gap-3 mb-6">
        <StatBox label="Attempts" value={stats.total} />
        <StatBox label="Average" value={`${stats.avgPct}%`} />
        <StatBox label="Best score" value={`${stats.bestPct}%`} />
      </div>

      <div className="space-y-2.5">
        {(quizHistory as QuizAttempt[]).map((att: QuizAttempt) => (
          <AttemptRow key={att.id} attempt={att} />
        ))}
      </div>
    </div>
  );
}

function StatBox({ label, value }: { label: string; value: string | number }) {
  return (
    <div className="card p-4 text-center">
      <div className="text-xl sm:text-2xl font-bold font-mono text-ink-900">{value}</div>
      <div className="text-[11px] text-ink-500 uppercase tracking-wider mt-0.5">{label}</div>
    </div>
  );
}

function AttemptRow({ attempt }: { attempt: QuizAttempt }) {
  const topic = getTopicById(attempt.topicId);
  const pct = Math.round((attempt.score / attempt.total) * 100);
  const passed = pct >= 60;
  const date = new Date(attempt.createdAt);
  const dateStr = date.toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' });
  const timeStr = date.toLocaleTimeString(undefined, { hour: 'numeric', minute: '2-digit' });

  return (
    <Link
      to={{ name: 'history-detail', params: { attemptId: attempt.id } }}
      className="card p-4 hover:shadow-card-hover hover:border-ink-300/80 transition-all group flex items-center gap-4"
    >
      <div className={`w-10 h-10 rounded-lg flex items-center justify-center shrink-0 ${
        passed ? 'bg-brand-50 text-brand-600' : 'bg-warning-50 text-warning-600'
      }`}>
        <Trophy className="w-5 h-5" />
      </div>
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2 mb-0.5">
          <h3 className="text-sm font-medium text-ink-900 truncate group-hover:text-ink-950">{attempt.topicTitle}</h3>
          {topic && <TopicIcon name={topic.icon} className="w-3.5 h-3.5 text-ink-300 shrink-0" />}
        </div>
        <div className="flex items-center gap-2 text-xs text-ink-400">
          <Clock className="w-3 h-3" />
          <span>{dateStr} · {timeStr}</span>
        </div>
      </div>
      <div className="flex items-center gap-3 shrink-0">
        <div className="text-right">
          <div className={`font-mono font-bold ${passed ? 'text-brand-700' : 'text-warning-700'}`}>
            {attempt.score}/{attempt.total}
          </div>
          <div className="text-[11px] text-ink-400">{pct}%</div>
        </div>
        <ArrowRight className="w-4 h-4 text-ink-300 group-hover:text-ink-600 group-hover:translate-x-0.5 transition-all" />
      </div>
    </Link>
  );
}

export function HistoryDetailPage({ attemptId }: { attemptId: string }) {
  const { user, getAttempt } = useStore();
  const [loading, setLoading] = useState(true);
  const [attempt, setAttempt] = useState<QuizAttempt | undefined>(undefined);

  useEffect(() => {
    const t = setTimeout(() => {
      setAttempt(getAttempt(attemptId));
      setLoading(false);
    }, 300);
    return () => clearTimeout(t);
  }, [getAttempt, attemptId]);

  if (loading) return <HistoryDetailSkeleton />;

  if (!user || !attempt) {
    return (
      <EmptyState
        icon={History}
        title="Attempt not found"
        description="That quiz attempt doesn't exist or isn't available for this account."
        action={<Link to={{ name: 'history', params: {} }} className="btn-secondary">Back to history</Link>}
      />
    );
  }

  const pct = Math.round((attempt.score / attempt.total) * 100);
  const date = new Date(attempt.createdAt);
  const dateStr = date.toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' });
  const timeStr = date.toLocaleTimeString(undefined, { hour: 'numeric', minute: '2-digit' });

  return (
    <div className="max-w-2xl mx-auto animate-fade-in">
      <Link to={{ name: 'history', params: {} }} className="inline-flex items-center gap-1.5 text-sm text-ink-500 hover:text-ink-800 mb-5 transition-colors">
        <ArrowLeft className="w-4 h-4" /> Back to history
      </Link>

      <div className="card p-6 mb-6">
        <div className="flex items-start justify-between gap-4 mb-3">
          <div>
            <h1 className="text-xl font-bold text-ink-900 mb-1">{attempt.topicTitle}</h1>
            <p className="text-xs text-ink-400">{dateStr} at {timeStr}</p>
          </div>
          <div className="text-right">
            <div className={`text-2xl font-bold font-mono ${pct >= 60 ? 'text-brand-700' : 'text-warning-700'}`}>
              {attempt.score}/{attempt.total}
            </div>
            <Badge tone={pct >= 60 ? 'brand' : 'warning'}>{pct}%</Badge>
          </div>
        </div>
        <div className="flex gap-2 text-xs text-ink-500">
          <span className="flex items-center gap-1"><CheckCircle2 className="w-3.5 h-3.5 text-brand-600" /> {attempt.score} correct</span>
          <span className="flex items-center gap-1"><XCircle className="w-3.5 h-3.5 text-error-500" /> {attempt.total - attempt.score} incorrect</span>
        </div>
      </div>

      <h2 className="text-sm font-semibold text-ink-500 uppercase tracking-wider mb-3">Answers</h2>
      <div className="space-y-3">
        {(attempt.answers as QuizAttempt['answers']).map((a: QuizAttempt['answers'][number], i: number) => (
          <div key={i} className="card p-5">
            <div className="flex items-start gap-3 mb-3">
              <div className={`w-6 h-6 rounded-full flex items-center justify-center shrink-0 text-white text-xs font-mono ${
                a.isCorrect ? 'bg-brand-600' : 'bg-error-500'
              }`}>
                {a.isCorrect ? <CheckCircle2 className="w-3.5 h-3.5" /> : <XCircle className="w-3.5 h-3.5" />}
              </div>
              <p className="text-sm font-medium text-ink-900 flex-1">{a.questionText}</p>
            </div>
            <div className="space-y-1.5 ml-9">
              <div className="flex items-start gap-2 text-sm">
                <span className="text-ink-400 w-16 shrink-0 text-xs mt-0.5">Your answer</span>
                <span className={a.isCorrect ? 'text-brand-700' : 'text-error-700'}>
                  {a.selectedOptionText || <span className="text-ink-400 italic">No answer selected</span>}
                </span>
              </div>
              {!a.isCorrect && (
                <div className="flex items-start gap-2 text-sm">
                  <span className="text-ink-400 w-16 shrink-0 text-xs mt-0.5">Correct</span>
                  <span className="text-brand-700 font-medium">{a.correctOptionText}</span>
                </div>
              )}
              <div className="flex items-start gap-2 text-sm mt-2 pt-2 border-t border-ink-100">
                <span className="text-ink-400 w-16 shrink-0 text-xs mt-0.5">Why</span>
                <span className="text-ink-600 leading-relaxed">{a.explanation}</span>
              </div>
            </div>
          </div>
        ))}
      </div>

      <div className="mt-6 flex gap-3">
        <Link to={{ name: 'history', params: {} }} className="btn-secondary">Back to history</Link>
      </div>
    </div>
  );
}

function HistorySkeleton() {
  return (
    <div>
      <div className="skeleton h-7 w-48 mb-2" />
      <div className="skeleton h-4 w-72 mb-6" />
      <div className="grid grid-cols-3 gap-3 mb-6">
        {Array.from({ length: 3 }).map((_, i) => <div key={i} className="skeleton h-20 rounded-xl" />)}
      </div>
      <div className="space-y-2.5">
        {Array.from({ length: 3 }).map((_, i) => <div key={i} className="skeleton h-16 rounded-xl" />)}
      </div>
      <div className="flex items-center justify-center gap-2 text-ink-400 text-sm py-4">
        <Spinner className="w-4 h-4" /> Loading history…
      </div>
    </div>
  );
}

function HistoryDetailSkeleton() {
  return (
    <div className="max-w-2xl mx-auto">
      <div className="skeleton h-4 w-32 mb-5" />
      <div className="skeleton h-28 rounded-xl mb-6" />
      <div className="space-y-3">
        {Array.from({ length: 3 }).map((_, i) => <div key={i} className="skeleton h-24 rounded-xl" />)}
      </div>
      <div className="flex items-center justify-center gap-2 text-ink-400 text-sm py-4">
        <Spinner className="w-4 h-4" /> Loading attempt…
      </div>
    </div>
  );
}
