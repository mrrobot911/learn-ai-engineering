import { useEffect, useState } from 'react';
import { ArrowLeft, ArrowRight, Check, ListChecks, Trophy } from 'lucide-react';
import { api } from '@/api';
import type { QuizAnswerResult, QuizQuestion, QuizResult, TopicSummary } from '@/api';
import { Link, Spinner, EmptyState, ErrorState, ProgressBar } from '@/components/ui';
import { routeToHash } from '@/router';

// The play screen hands the submission response to the result screen in memory
// only: nothing is written to browser storage, so reloading the result route
// shows the empty state instead of a stale result.
let submittedResult: { id: string; slug: string; topicTitle: string; result: QuizResult } | null = null;

export function QuizStartPage({ slug }: { slug: string }) {
  const [topic, setTopic] = useState<TopicSummary | null>(null);
  const [notFound, setNotFound] = useState(false);
  const [loading, setLoading] = useState(true);
  const [errored, setErrored] = useState(false);
  const [reloadKey, setReloadKey] = useState(0);

  // The route slug resolves through the API client only: listTopics() supplies
  // the summary the start screen shows. No fake latency — the loading state
  // lasts exactly as long as the client promise is pending.
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
        if (!cancelled) {
          if (summary) setTopic(summary);
          else setNotFound(true);
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

  if (loading) return <QuizStartSkeleton />;

  if (errored) {
    return <ErrorState onRetry={() => setReloadKey((key) => key + 1)} />;
  }

  if (notFound || !topic) {
    return (
      <EmptyState
        icon={Trophy}
        title="Quiz not found"
        description="This topic doesn't exist, so there's no quiz to start."
        action={<Link to={{ name: 'topics', params: {} }} className="btn-secondary">Back to topics</Link>}
      />
    );
  }

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
          Answer up to five multiple-choice questions chosen at random from this topic's published questions.
          Each question has exactly one correct answer.
        </p>

        <Link
          to={{ name: 'quiz', params: { slug: topic.slug } }}
          className="btn-brand w-full sm:w-auto"
        >
          Start quiz <ArrowRight className="w-4 h-4" />
        </Link>
      </div>
    </div>
  );
}

export function QuizPlayPage({ slug }: { slug: string }) {
  const [topic, setTopic] = useState<TopicSummary | null>(null);
  const [quizQuestions, setQuizQuestions] = useState<QuizQuestion[]>([]);
  const [notFound, setNotFound] = useState(false);
  const [loading, setLoading] = useState(true);
  const [errored, setErrored] = useState(false);
  const [reloadKey, setReloadKey] = useState(0);
  const [current, setCurrent] = useState(0);
  const [answers, setAnswers] = useState<Record<string, string | null>>({});
  const [submitting, setSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState(false);

  // The questions come from startQuiz() and nothing else: the browser never
  // picks them and never sees a correct option before the submission response.
  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setErrored(false);
    setNotFound(false);
    setTopic(null);
    setQuizQuestions([]);
    setCurrent(0);
    setAnswers({});
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
        const questions = await api.startQuiz(summary.id);
        if (!cancelled) {
          setTopic(summary);
          setQuizQuestions(questions);
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

  if (loading) return <QuizPlaySkeleton />;

  if (errored) {
    return <ErrorState onRetry={() => setReloadKey((key) => key + 1)} />;
  }

  if (notFound || !topic) {
    return (
      <EmptyState
        icon={Trophy}
        title="Quiz not found"
        description="This topic doesn't exist, so there's no quiz to take."
        action={<Link to={{ name: 'topics', params: {} }} className="btn-secondary">Back to topics</Link>}
      />
    );
  }

  if (quizQuestions.length === 0) {
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
    // Activating the selected option again returns the question to unanswered.
    setAnswers((prev) => ({ ...prev, [q.id]: prev[q.id] === optionId ? null : optionId }));
  }

  function go(delta: number) {
    setCurrent((c) => Math.min(quizQuestions.length - 1, Math.max(0, c + delta)));
  }

  async function handleSubmit() {
    if (!topic) return;
    setSubmitting(true);
    setSubmitError(false);
    try {
      const result = await api.submitQuiz({
        topicId: topic.id,
        answers: quizQuestions.map((question) => ({
          questionId: question.id,
          optionId: answers[question.id] ?? null,
        })),
      });
      const id = `${Date.now().toString(36)}${Math.random().toString(36).slice(2, 8)}`;
      submittedResult = { id, slug, topicTitle: topic.title, result };
      window.location.hash = routeToHash({ name: 'quiz-result', params: { slug, attemptId: id } });
    } catch {
      setSubmitting(false);
      setSubmitError(true);
    }
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
                type="button"
                onClick={() => selectOption(opt.id)}
                aria-pressed={isSelected}
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
              type="button"
              onClick={() => setCurrent(i)}
              aria-label={`Go to question ${i + 1}`}
              aria-current={isCurrent ? 'step' : undefined}
              className={`w-2.5 h-2.5 rounded-full transition-all ${
                isCurrent ? 'bg-ink-900 w-5' : isAnswered ? 'bg-brand-500' : 'bg-ink-200 hover:bg-ink-300'
              }`}
            />
          );
        })}
      </div>

      {/* Nav buttons */}
      <div className="flex items-center justify-between mt-6">
        <button
          type="button"
          className="btn-secondary"
          onClick={() => go(-1)}
          disabled={current === 0}
        >
          <ArrowLeft className="w-4 h-4" /> Previous
        </button>

        {isLast ? (
          <button type="button" className="btn-brand" onClick={() => { void handleSubmit(); }} disabled={submitting}>
            {submitting ? <><Spinner className="w-4 h-4" /> Submitting…</> : <>Submit quiz <Check className="w-4 h-4" /></>}
          </button>
        ) : (
          <button type="button" className="btn-primary" onClick={() => go(1)}>
            Next <ArrowRight className="w-4 h-4" />
          </button>
        )}
      </div>

      {submitError && (
        <div role="alert" className="mt-4 flex flex-col sm:flex-row items-center justify-center gap-3 text-sm text-error-700">
          <span>Couldn't submit the quiz. Your answers are still here.</span>
          <button type="button" className="btn-secondary" onClick={() => { void handleSubmit(); }}>
            Try again
          </button>
        </div>
      )}
    </div>
  );
}

export function QuizResultPage({ slug, attemptId }: { slug: string; attemptId: string }) {
  const submitted = submittedResult;
  const attempt = submitted && submitted.slug === slug && submitted.id === attemptId
    ? submitted
    : null;

  if (!attempt) {
    return (
      <EmptyState
        icon={Trophy}
        title="Result not found"
        description="We couldn't find this quiz attempt. It may have been an anonymous attempt in a previous session."
        action={<Link to={{ name: 'quiz-start', params: { slug } }} className="btn-secondary">Back to quiz</Link>}
      />
    );
  }

  const { topicTitle, result } = attempt;

  return (
    <div className="max-w-2xl mx-auto animate-fade-in">
      <div className="card p-6 sm:p-8 mb-6 text-center">
        <div className="w-14 h-14 rounded-2xl bg-brand-50 text-brand-600 flex items-center justify-center mx-auto mb-4">
          <Trophy className="w-7 h-7" />
        </div>
        <p className="text-xs font-mono text-ink-400 uppercase tracking-wider mb-1">{topicTitle} · Quiz result</p>
        <div className="flex items-baseline justify-center gap-1.5 mb-2">
          <span className="text-4xl font-bold font-mono text-ink-900">{result.score}</span>
          <span className="text-xl text-ink-400 font-mono">/ {result.total}</span>
        </div>
      </div>

      <h2 className="text-sm font-semibold text-ink-500 uppercase tracking-wider mb-3">Review answers</h2>
      <div className="space-y-3">
        {result.answers.map((a: QuizAnswerResult) => (
          <div key={a.questionId} className="card p-5">
            <div className="flex items-start gap-3 mb-3">
              <div
                role="img"
                aria-label={a.isCorrect ? 'Correct' : 'Incorrect'}
                className={`w-6 h-6 rounded-full flex items-center justify-center shrink-0 text-white text-xs font-mono ${
                  a.isCorrect ? 'bg-brand-600' : 'bg-error-500'
                }`}
              >
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
              <div className="flex items-start gap-2 text-sm">
                <span className="text-ink-400 w-16 shrink-0 text-xs mt-0.5">Correct</span>
                <span className="text-brand-700 font-medium">{a.correctOptionText}</span>
              </div>
              <div className="flex items-start gap-2 text-sm mt-2 pt-2 border-t border-ink-100">
                <span className="text-ink-400 w-16 shrink-0 text-xs mt-0.5">Why</span>
                <span className="text-ink-600 leading-relaxed">{a.explanation}</span>
              </div>
            </div>
          </div>
        ))}
      </div>

      <div className="flex flex-col sm:flex-row gap-3 mt-6">
        <Link to={{ name: 'quiz-start', params: { slug } }} className="btn-brand flex-1">
          Take another attempt <ArrowRight className="w-4 h-4" />
        </Link>
        <Link to={{ name: 'topic', params: { slug } }} className="btn-secondary flex-1">
          Back to topic
        </Link>
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
