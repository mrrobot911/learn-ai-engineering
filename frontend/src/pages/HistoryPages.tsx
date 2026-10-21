import { useEffect, useState } from 'react';
import {
  ArrowLeft, ArrowRight, CheckCircle2, Clock, History, Trophy, XCircle,
} from 'lucide-react';
import { api } from '@/api';
import type { QuizAnswerResult, QuizAttempt, QuizAttemptSummary } from '@/api';
import { useSession } from '@/session';
import { Link, Spinner, EmptyState, ErrorState } from '@/components/ui';

/** The UTC day of an ISO timestamp, so the date does not depend on the locale. */
function utcDay(iso: string): string {
  return new Date(iso).toISOString().slice(0, 10);
}

export function HistoryPage() {
  const { user } = useSession();
  const [attempts, setAttempts] = useState<QuizAttemptSummary[] | null>(null);
  const [loading, setLoading] = useState(true);
  const [failed, setFailed] = useState(false);
  const [retry, setRetry] = useState(0);
  const userId = user?.id;

  useEffect(() => {
    if (!userId) return;
    let cancelled = false;
    setLoading(true);
    setFailed(false);
    api.listAttempts()
      .then((list) => {
        if (cancelled) return;
        setAttempts(list);
        setLoading(false);
      })
      .catch(() => {
        if (cancelled) return;
        setFailed(true);
        setLoading(false);
      });
    return () => { cancelled = true; };
  }, [userId, retry]);

  const retryLoad = () => {
    setFailed(false);
    setLoading(true);
    setRetry((count) => count + 1);
  };

  if (!user) return <SignInPrompt />;
  if (loading) return <HistorySkeleton />;
  if (failed) return <ErrorState description="Your quiz history could not be loaded." onRetry={retryLoad} />;

  if (attempts && attempts.length === 0) {
    return (
      <EmptyState
        icon={Trophy}
        title="No quiz attempts yet"
        description="Take a quiz on any topic and your attempts will appear here, with the score and the answers of every question."
        action={<Link to={{ name: 'topics', params: {} }} className="btn-brand">Browse topics</Link>}
      />
    );
  }

  return (
    <div className="animate-fade-in">
      <div className="flex items-center gap-2 text-xs font-mono text-ink-400 mb-3">
        <History className="w-3 h-3" aria-hidden="true" /> {user.email}'s quiz history
      </div>
      <h1 className="text-2xl sm:text-3xl font-bold text-ink-900 tracking-tight mb-1">Quiz history</h1>
      <p className="text-ink-600 mb-6">Every attempt is saved with a snapshot of the questions and answers at the time.</p>

      <ul className="space-y-2.5">
        {(attempts ?? []).map((attempt) => (
          <AttemptRow key={attempt.id} attempt={attempt} />
        ))}
      </ul>
    </div>
  );
}

function AttemptRow({ attempt }: { attempt: QuizAttemptSummary }) {
  const passed = attempt.total > 0 && attempt.score / attempt.total >= 0.6;

  return (
    <li>
      <Link
        to={{ name: 'history-detail', params: { attemptId: attempt.id } }}
        className="card p-4 hover:shadow-card-hover hover:border-ink-300/80 transition-all group flex items-center gap-4"
      >
        <div className={`w-10 h-10 rounded-lg flex items-center justify-center shrink-0 ${
          passed ? 'bg-brand-50 text-brand-600' : 'bg-warning-50 text-warning-600'
        }`}>
          <Trophy className="w-5 h-5" aria-hidden="true" />
        </div>
        <div className="flex-1 min-w-0">
          <h3 className="text-sm font-medium text-ink-900 truncate group-hover:text-ink-950">{attempt.topicTitle}</h3>
          <div className="flex items-center gap-2 text-xs text-ink-400">
            <Clock className="w-3 h-3" aria-hidden="true" />
            <time dateTime={attempt.createdAt}>{utcDay(attempt.createdAt)}</time>
          </div>
        </div>
        <div className="flex items-center gap-3 shrink-0">
          <div className={`font-mono font-bold ${passed ? 'text-brand-700' : 'text-warning-700'}`}>
            {attempt.score}/{attempt.total}
          </div>
          <ArrowRight className="w-4 h-4 text-ink-300 group-hover:text-ink-600 group-hover:translate-x-0.5 transition-all" aria-hidden="true" />
        </div>
      </Link>
    </li>
  );
}

export function HistoryDetailPage({ attemptId }: { attemptId: string }) {
  const { user } = useSession();
  const [attempt, setAttempt] = useState<QuizAttempt | null>(null);
  const [loading, setLoading] = useState(true);
  const [failed, setFailed] = useState(false);
  const [retry, setRetry] = useState(0);
  const userId = user?.id;

  useEffect(() => {
    if (!userId) return;
    let cancelled = false;
    setLoading(true);
    setFailed(false);
    api.getAttempt(attemptId)
      .then((loaded) => {
        if (cancelled) return;
        setAttempt(loaded);
        setLoading(false);
      })
      .catch(() => {
        if (cancelled) return;
        setFailed(true);
        setLoading(false);
      });
    return () => { cancelled = true; };
  }, [userId, attemptId, retry]);

  const retryLoad = () => {
    setFailed(false);
    setLoading(true);
    setRetry((count) => count + 1);
  };

  if (!user) return <SignInPrompt />;
  if (loading) return <HistoryDetailSkeleton />;
  if (failed || !attempt) {
    return <ErrorState description="This attempt could not be loaded." onRetry={retryLoad} />;
  }

  return (
    <div className="max-w-2xl mx-auto animate-fade-in">
      <Link to={{ name: 'history', params: {} }} className="inline-flex items-center gap-1.5 text-sm text-ink-500 hover:text-ink-800 mb-5 transition-colors">
        <ArrowLeft className="w-4 h-4" aria-hidden="true" /> Back to history
      </Link>

      <div className="card p-6 mb-6">
        <div className="flex items-start justify-between gap-4">
          <div>
            <h1 className="text-xl font-bold text-ink-900 mb-1">{attempt.topicTitle}</h1>
            <p className="text-xs text-ink-400">
              <time dateTime={attempt.createdAt}>{utcDay(attempt.createdAt)}</time>
            </p>
          </div>
          <div className="font-mono font-bold text-2xl text-brand-700">
            {attempt.score}/{attempt.total}
          </div>
        </div>
      </div>

      <h2 className="text-sm font-semibold text-ink-500 uppercase tracking-wider mb-3">Answers</h2>
      <ul className="space-y-3">
        {attempt.answers.map((answer) => (
          <AnswerCard key={answer.questionId} answer={answer} />
        ))}
      </ul>

      <div className="mt-6">
        <Link to={{ name: 'history', params: {} }} className="btn-secondary">Back to history</Link>
      </div>
    </div>
  );
}

function AnswerCard({ answer }: { answer: QuizAnswerResult }) {
  return (
    <li className="card p-5">
      <div className="flex items-start gap-3 mb-3">
        <div className={`w-6 h-6 rounded-full flex items-center justify-center shrink-0 text-white text-xs font-mono ${
          answer.isCorrect ? 'bg-brand-600' : 'bg-error-500'
        }`}>
          {answer.isCorrect
            ? <CheckCircle2 className="w-3.5 h-3.5" aria-hidden="true" />
            : <XCircle className="w-3.5 h-3.5" aria-hidden="true" />}
        </div>
        <p className="text-sm font-medium text-ink-900 flex-1">{answer.questionText}</p>
      </div>
      <dl className="space-y-1.5 ml-9">
        <div className="flex items-start gap-2 text-sm">
          <dt className="text-ink-400 w-24 shrink-0 text-xs mt-0.5">Your answer</dt>
          <dd className={answer.isCorrect ? 'text-brand-700' : 'text-error-700'}>
            {answer.selectedOptionText || <span className="text-ink-400 italic">No answer selected</span>}
          </dd>
        </div>
        <div className="flex items-start gap-2 text-sm">
          <dt className="text-ink-400 w-24 shrink-0 text-xs mt-0.5">Correct answer</dt>
          <dd className="text-brand-700 font-medium">{answer.correctOptionText}</dd>
        </div>
      </dl>
    </li>
  );
}

function SignInPrompt() {
  return (
    <EmptyState
      icon={History}
      title="Log in to see your quiz history"
      description="Quiz attempts are only saved for registered users. Log in to see your past attempts."
      action={
        <div className="flex gap-2">
          <Link to={{ name: 'login', params: {} }} className="btn-primary">Log in</Link>
          <Link to={{ name: 'register', params: {} }} className="btn-secondary">Sign up</Link>
        </div>
      }
    />
  );
}

function HistorySkeleton() {
  return (
    <div>
      <div className="skeleton h-7 w-48 mb-2" />
      <div className="skeleton h-4 w-72 mb-6" />
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
