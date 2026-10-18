import { useEffect, useMemo, useState } from 'react';
import {
  ArrowLeft, ArrowRight, Trophy, ListChecks, Sparkles, AlertTriangle, Check,
} from 'lucide-react';
import type { Question, QuizAttempt, QuizResultAnswer, Topic } from '@/types';
import { getTopicBySlug, getQuestionsForTopic } from '@/data/content';
import { useStore } from '@/store';
import { useStore as useStoreForHistory } from '@/store';
import { Link, Spinner, EmptyState, ProgressBar, toast } from '@/components/ui';
import { TopicIcon } from '@/components/TopicIcon';
import { pickQuizQuestions, makeSeed, gradeAnswer, QUIZ_SIZE } from '@/lib/quiz';

export function QuizStartPage({ slug }: { slug: string }) {
  const { user, quizHistory } = useStore();
  const [loading, setLoading] = useState(true);
  const [topic, setTopic] = useState<Topic | undefined>(undefined);
  const [questions, setQuestions] = useState<Question[]>([]);

  useEffect(() => {
    setLoading(true);
    const t = setTimeout(() => {
      const tp = getTopicBySlug(slug);
      setTopic(tp);
      setQuestions(tp ? getQuestionsForTopic(tp.id) : []);
      setLoading(false);
    }, 300);
    return () => clearTimeout(t);
  }, [slug]);

  const attempts = useMemo(() => topic ? quizHistory.filter((a) => a.topicId === topic.id) : [], [topic, quizHistory]);

  if (loading) return <QuizStartSkeleton />;
  if (!topic) {
    return (
      <EmptyState
        icon={Trophy}
        title="Quiz not found"
        description="This topic doesn't exist, so there's no quiz to start."
        action={<Link to={{ name: 'topics', params: {} }} className="btn-secondary">Back to topics</Link>}
      />
    );
  }

  const quizSize = Math.min(QUIZ_SIZE, questions.length);
  const lastAttempt = attempts[0];

  return (
    <div className="max-w-2xl mx-auto animate-fade-in">
      <Link to={{ name: 'topic', params: { slug: topic.slug } }} className="inline-flex items-center gap-1.5 text-sm text-ink-500 hover:text-ink-800 mb-5 transition-colors">
        <ArrowLeft className="w-4 h-4" /> Back to {topic.title}
      </Link>

      <div className="card p-6 sm:p-8 text-center">
        <div className="w-14 h-14 rounded-2xl bg-ink-900 text-white flex items-center justify-center mx-auto mb-4">
          <Trophy className="w-7 h-7" />
        </div>
        <h1 className="text-2xl font-bold text-ink-900 mb-2">{topic.title} Quiz</h1>
        <p className="text-ink-600 mb-6 max-w-md mx-auto">
          Answer {quizSize} multiple-choice questions drawn at random from this topic's question bank.
          Each question has exactly one correct answer.
        </p>

        <div className="grid grid-cols-3 gap-3 mb-6">
          <Stat label="Questions" value={quizSize} />
          <Stat label="In bank" value={questions.length} />
          <Stat label="Attempts" value={attempts.length} />
        </div>

        {lastAttempt && (
          <div className="mb-6 p-3.5 rounded-lg bg-ink-50 border border-ink-200/70 text-left">
            <div className="flex items-center justify-between text-sm">
              <span className="text-ink-600 flex items-center gap-1.5">
                <Sparkles className="w-3.5 h-3.5 text-accent-500" /> Last attempt
              </span>
              <span className="font-mono text-ink-900 font-semibold">
                {lastAttempt.score}/{lastAttempt.total}
                <span className="text-ink-400 ml-1">({Math.round((lastAttempt.score / lastAttempt.total) * 100)}%)</span>
              </span>
            </div>
          </div>
        )}

        {!user && (
          <div className="mb-5 p-3 rounded-lg bg-accent-50 border border-accent-100 text-accent-700 text-xs flex items-start gap-2 text-left">
            <AlertTriangle className="w-3.5 h-3.5 mt-0.5 shrink-0" />
            <span>You can take this quiz without an account, but your result won't be saved. <Link to={{ name: 'login', params: {} }} className="underline font-medium">Log in</Link> to keep your history.</span>
          </div>
        )}

        <Link
          to={{ name: 'quiz', params: { slug: topic.slug } }}
          className="btn-brand w-full sm:w-auto"
        >
          Start quiz <ArrowRight className="w-4 h-4" />
        </Link>
      </div>

      {questions.length < QUIZ_SIZE && (
        <p className="text-xs text-ink-400 text-center mt-4">
          This topic has {questions.length} published question{questions.length === 1 ? '' : 's'} — all of them will appear.
        </p>
      )}
    </div>
  );
}

function Stat({ label, value }: { label: string; value: number }) {
  return (
    <div className="rounded-lg bg-ink-50 py-3">
      <div className="text-xl font-bold text-ink-900 font-mono">{value}</div>
      <div className="text-[11px] text-ink-500 uppercase tracking-wider">{label}</div>
    </div>
  );
}

export function QuizPlayPage({ slug }: { slug: string }) {
  const { addQuizAttempt, user } = useStoreForHistory();
  const [loading, setLoading] = useState(true);
  const [topic, setTopic] = useState<Topic | undefined>(undefined);
  const [quizQuestions, setQuizQuestions] = useState<Question[]>([]);
  const [current, setCurrent] = useState(0);
  const [answers, setAnswers] = useState<Record<string, string | null>>({});
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    setLoading(true);
    const t = setTimeout(() => {
      const tp = getTopicBySlug(slug);
      setTopic(tp);
      if (tp) {
        const pool = getQuestionsForTopic(tp.id);
        setQuizQuestions(pickQuizQuestions(pool, makeSeed()));
      }
      setLoading(false);
    }, 300);
    return () => clearTimeout(t);
  }, [slug]);

  if (loading) return <QuizPlaySkeleton />;
  if (!topic || quizQuestions.length === 0) {
    return (
      <EmptyState
        icon={ListChecks}
        title="No questions available"
        description="This topic doesn't have any published quiz questions yet."
        action={<Link to={{ name: 'topic', params: { slug } }} className="btn-secondary">Back to topic</Link>}
      />
    );
  }

  const q = quizQuestions[current];
  const selected = answers[q.id] ?? null;
  const answeredCount = Object.values(answers).filter((v) => v !== null && v !== undefined).length;
  const isLast = current === quizQuestions.length - 1;

  function selectOption(optionId: string) {
    setAnswers((prev) => ({ ...prev, [q.id]: optionId }));
  }

  function go(delta: number) {
    setCurrent((c) => Math.min(quizQuestions.length - 1, Math.max(0, c + delta)));
  }

  function handleSubmit() {
    setSubmitting(true);
    // Simulate grading latency.
    setTimeout(() => {
      const resultAnswers: QuizResultAnswer[] = quizQuestions.map((qq) => {
        const sel = answers[qq.id] ?? null;
        const g = gradeAnswer(qq, sel);
        return {
          questionId: qq.id,
          questionText: qq.text,
          selectedOptionId: sel,
          selectedOptionText: g.selectedOptionText,
          correctOptionId: g.correctOptionId,
          correctOptionText: g.correctOptionText,
          isCorrect: g.isCorrect,
          explanation: qq.explanation,
        };
      });
      const score = resultAnswers.filter((a) => a.isCorrect).length;
      const attempt: QuizAttempt = {
        id: `att-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
        topicId: topic!.id,
        topicTitle: topic!.title,
        score,
        total: quizQuestions.length,
        createdAt: Date.now(),
        answers: resultAnswers,
      };
      if (user) {
        addQuizAttempt(attempt);
        toast(`Quiz submitted — scored ${score}/${quizQuestions.length}. Saved to your history.`, 'success');
      } else {
        toast(`Quiz submitted — scored ${score}/${quizQuestions.length}.`, 'info');
      }
      // Stash the transient attempt for the result page (anonymous case).
      try { sessionStorage.setItem('forge.lastAttempt', JSON.stringify(attempt)); } catch { /* ignore */ }
      window.location.hash = `#/quiz/${topic!.slug}/result/${attempt.id}`;
    }, 700);
  }

  return (
    <div className="max-w-2xl mx-auto animate-fade-in">
      <div className="flex items-center justify-between mb-4">
        <Link to={{ name: 'quiz-start', params: { slug: topic.slug } }} className="text-sm text-ink-500 hover:text-ink-800 flex items-center gap-1.5">
          <ArrowLeft className="w-4 h-4" /> Exit
        </Link>
        <span className="text-sm font-mono text-ink-600">
          Question <b className="text-ink-900">{current + 1}</b> of {quizQuestions.length}
        </span>
      </div>

      <ProgressBar value={answeredCount} total={quizQuestions.length} className="mb-6" />

      <div key={q.id} className="card p-6 sm:p-7 animate-slide-up">
        <p className="text-xs font-mono text-ink-400 uppercase tracking-wider mb-3">Question {current + 1}</p>
        <h2 className="text-lg sm:text-xl font-semibold text-ink-900 leading-snug mb-5">{q.text}</h2>

        <div className="space-y-2.5">
          {q.options.map((opt, i) => {
            const isSelected = selected === opt.id;
            return (
              <button
                key={opt.id}
                onClick={() => selectOption(opt.id)}
                className={`w-full text-left rounded-lg border px-4 py-3.5 transition-all flex items-center gap-3 ${
                  isSelected
                    ? 'border-brand-500 bg-brand-50/60 ring-2 ring-brand-500/20'
                    : 'border-ink-200 bg-white hover:border-ink-300 hover:bg-ink-50/50'
                }`}
              >
                <div className={`w-6 h-6 rounded-full flex items-center justify-center text-xs font-mono shrink-0 border ${
                  isSelected ? 'bg-brand-600 text-white border-brand-600' : 'bg-white text-ink-500 border-ink-300'
                }`}>
                  {isSelected ? <Check className="w-3.5 h-3.5" /> : String.fromCharCode(65 + i)}
                </div>
                <span className={`text-sm ${isSelected ? 'text-ink-900 font-medium' : 'text-ink-700'}`}>{opt.text}</span>
              </button>
            );
          })}
        </div>
      </div>

      {/* Question dots */}
      <div className="flex items-center justify-center gap-1.5 mt-6">
        {quizQuestions.map((qq, i) => {
          const isAnswered = answers[qq.id] !== undefined && answers[qq.id] !== null;
          const isCurrent = i === current;
          return (
            <button
              key={qq.id}
              onClick={() => setCurrent(i)}
              className={`w-2.5 h-2.5 rounded-full transition-all ${
                isCurrent ? 'bg-ink-900 w-5' : isAnswered ? 'bg-brand-500' : 'bg-ink-200 hover:bg-ink-300'
              }`}
              aria-label={`Go to question ${i + 1}`}
            />
          );
        })}
      </div>

      {/* Nav buttons */}
      <div className="flex items-center justify-between mt-6">
        <button
          className="btn-secondary"
          onClick={() => go(-1)}
          disabled={current === 0}
        >
          <ArrowLeft className="w-4 h-4" /> Previous
        </button>

        {isLast ? (
          <button className="btn-brand" onClick={handleSubmit} disabled={submitting}>
            {submitting ? <><Spinner className="w-4 h-4" /> Grading…</> : <>Submit quiz <Check className="w-4 h-4" /></>}
          </button>
        ) : (
          <button className="btn-primary" onClick={() => go(1)}>
            Next <ArrowRight className="w-4 h-4" />
          </button>
        )}
      </div>
    </div>
  );
}

export function QuizResultPage({ slug, attemptId }: { slug: string; attemptId: string }) {
  const { user, getAttempt } = useStore();
  const [loading, setLoading] = useState(true);
  const [attempt, setAttempt] = useState<QuizAttempt | undefined>(undefined);
  const [topic, setTopic] = useState<Topic | undefined>(undefined);

  useEffect(() => {
    setLoading(true);
    const t = setTimeout(() => {
      const tp = getTopicBySlug(slug);
      setTopic(tp);
      let att = user ? getAttempt(attemptId) : undefined;
      if (!att) {
        // fall back to transient anonymous attempt from sessionStorage
        try {
          const raw = sessionStorage.getItem('forge.lastAttempt');
          if (raw) {
            const parsed = JSON.parse(raw) as QuizAttempt;
            if (parsed.id === attemptId) att = parsed;
          }
        } catch { /* ignore */ }
      }
      setAttempt(att);
      setLoading(false);
    }, 350);
    return () => clearTimeout(t);
  }, [attemptId, slug, user, getAttempt]);

  if (loading) return <QuizResultSkeleton />;
  if (!attempt || !topic) {
    return (
      <EmptyState
        icon={Trophy}
        title="Result not found"
        description="We couldn't find this quiz attempt. It may have been an anonymous attempt in a previous session."
        action={<Link to={{ name: 'quiz-start', params: { slug } }} className="btn-secondary">Back to quiz</Link>}
      />
    );
  }

  const pct = Math.round((attempt.score / attempt.total) * 100);
  const passed = pct >= 60;
  const gradeLabel = pct >= 80 ? 'Excellent' : pct >= 60 ? 'Good' : pct >= 40 ? 'Keep going' : 'Needs work';
  const gradeColor = pct >= 80 ? 'text-brand-600' : pct >= 60 ? 'text-accent-600' : 'text-warning-600';

  return (
    <div className="max-w-2xl mx-auto animate-fade-in">
      <div className="card p-6 sm:p-8 mb-6 text-center">
        <div className={`w-14 h-14 rounded-2xl mx-auto mb-4 flex items-center justify-center ${
          passed ? 'bg-brand-50 text-brand-600' : 'bg-warning-50 text-warning-600'
        }`}>
          <Trophy className="w-7 h-7" />
        </div>
        <p className="text-xs font-mono text-ink-400 uppercase tracking-wider mb-1">{topic.title} · Quiz result</p>
        <div className="flex items-baseline justify-center gap-1.5 mb-2">
          <span className={`text-4xl font-bold font-mono ${gradeColor}`}>{attempt.score}</span>
          <span className="text-xl text-ink-400 font-mono">/ {attempt.total}</span>
        </div>
        <p className={`text-sm font-medium ${gradeColor}`}>{gradeLabel} · {pct}%</p>

        {!user && (
          <p className="text-xs text-ink-500 mt-3 max-w-sm mx-auto">
            This was an anonymous attempt — the result isn't saved. <Link to={{ name: 'register', params: {} }} className="underline">Sign up</Link> to keep your history.
          </p>
        )}
      </div>

      <h2 className="text-sm font-semibold text-ink-500 uppercase tracking-wider mb-3">Review answers</h2>
      <div className="space-y-3">
        {attempt.answers.map((a, i) => (
          <div key={i} className="card p-5">
            <div className="flex items-start gap-3 mb-3">
              <div className={`w-6 h-6 rounded-full flex items-center justify-center shrink-0 text-white text-xs font-mono ${
                a.isCorrect ? 'bg-brand-600' : 'bg-error-500'
              }`}>
                {a.isCorrect ? <Check className="w-3.5 h-3.5" /> : '✕'}
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

      <div className="flex flex-col sm:flex-row gap-3 mt-6">
        <Link to={{ name: 'quiz-start', params: { slug: topic.slug } }} className="btn-brand flex-1">
          Take another attempt <ArrowRight className="w-4 h-4" />
        </Link>
        <Link to={{ name: 'topic', params: { slug: topic.slug } }} className="btn-secondary flex-1">
          Back to topic
        </Link>
        {user && (
          <Link to={{ name: 'history', params: {} }} className="btn-secondary flex-1">
            View history
          </Link>
        )}
      </div>
    </div>
  );
}

function QuizStartSkeleton() {
  return (
    <div className="max-w-2xl mx-auto">
      <div className="skeleton h-4 w-32 mb-5" />
      <div className="card p-8 text-center">
        <div className="skeleton w-14 h-14 rounded-2xl mx-auto mb-4" />
        <div className="skeleton h-7 w-1/2 mx-auto mb-2" />
        <div className="skeleton h-4 w-3/4 mx-auto mb-6" />
        <div className="grid grid-cols-3 gap-3 mb-6">
          {Array.from({ length: 3 }).map((_, i) => <div key={i} className="skeleton h-16 rounded-lg" />)}
        </div>
        <div className="skeleton h-10 w-40 mx-auto rounded-lg" />
      </div>
      <div className="flex items-center justify-center gap-2 text-ink-400 text-sm py-4">
        <Spinner className="w-4 h-4" /> Loading quiz…
      </div>
    </div>
  );
}

function QuizPlaySkeleton() {
  return (
    <div className="max-w-2xl mx-auto">
      <div className="skeleton h-2 w-full rounded-full mb-6" />
      <div className="card p-7">
        <div className="skeleton h-3 w-24 mb-3" />
        <div className="skeleton h-6 w-full mb-2" />
        <div className="skeleton h-6 w-4/5 mb-5" />
        <div className="space-y-2.5">
          {Array.from({ length: 4 }).map((_, i) => <div key={i} className="skeleton h-12 rounded-lg" />)}
        </div>
      </div>
      <div className="flex items-center justify-center gap-2 text-ink-400 text-sm py-5">
        <Spinner className="w-4 h-4" /> Loading question…
      </div>
    </div>
  );
}

function QuizResultSkeleton() {
  return (
    <div className="max-w-2xl mx-auto">
      <div className="card p-8 mb-6">
        <div className="skeleton w-14 h-14 rounded-2xl mx-auto mb-4" />
        <div className="skeleton h-10 w-32 mx-auto mb-2" />
        <div className="skeleton h-4 w-24 mx-auto" />
      </div>
      <div className="flex items-center justify-center gap-2 text-ink-400 text-sm py-4">
        <Spinner className="w-4 h-4" /> Loading result…
      </div>
    </div>
  );
}
