import { StoreProvider } from '@/store';
import { SessionProvider, useSession } from '@/session';
import { useRouter } from '@/router';
import { AppShell } from '@/components/AppShell';
import { ToastHost } from '@/components/ToastHost';
import { Spinner } from '@/components/ui';
import { TopicsPage } from '@/pages/TopicsPage';
import { TopicPage } from '@/pages/TopicPage';
import { LessonPage } from '@/pages/LessonPage';
import { QuizStartPage, QuizPlayPage, QuizResultPage } from '@/pages/QuizPages';
import { HistoryPage, HistoryDetailPage } from '@/pages/HistoryPages';
import { AuthPage } from '@/pages/AuthPage';

function Routes() {
  const { route } = useRouter();
  const { authReady } = useSession();

  if (!authReady) {
    return (
      <div className="flex items-center justify-center py-32 text-ink-400">
        <Spinner className="w-5 h-5" /> <span className="ml-2 text-sm">Loading…</span>
      </div>
    );
  }

  switch (route.name) {
    case 'topics':
      return <TopicsPage />;
    case 'topic':
      return <TopicPage slug={route.params.slug} />;
    case 'lesson':
      return <LessonPage id={route.params.id} />;
    case 'quiz-start':
      return <QuizStartPage slug={route.params.slug} />;
    case 'quiz':
      return <QuizPlayPage slug={route.params.slug} />;
    case 'quiz-result':
      return <QuizResultPage slug={route.params.slug} attemptId={route.params.attemptId} />;
    case 'history':
      return <HistoryPage />;
    case 'history-detail':
      return <HistoryDetailPage attemptId={route.params.attemptId} />;
    case 'login':
      return <AuthPage mode="login" />;
    case 'register':
      return <AuthPage mode="register" />;
    default:
      return <TopicsPage />;
  }
}

export default function App() {
  return (
    <SessionProvider>
      <StoreProvider>
        <AppShell>
          <Routes />
        </AppShell>
        <ToastHost />
      </StoreProvider>
    </SessionProvider>
  );
}
