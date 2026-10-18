import { useEffect, useState, useCallback } from 'react';
import type { Route } from '@/types';

function parseHash(hash: string): Route {
  const clean = hash.replace(/^#\/?/, '').trim();
  if (!clean) return { name: 'topics', params: {} };

  const parts = clean.split('/').filter(Boolean);
  const [head, ...rest] = parts;

  switch (head) {
    case 'topics':
      return { name: 'topics', params: {} };
    case 'topic':
      if (rest[0]) return { name: 'topic', params: { slug: rest[0] } };
      return { name: 'topics', params: {} };
    case 'lesson':
      if (rest[0]) return { name: 'lesson', params: { id: rest[0] } };
      return { name: 'topics', params: {} };
    case 'quiz': {
      const topicSlug = rest[0];
      const action = rest[1];
      if (!topicSlug) return { name: 'topics', params: {} };
      if (action === 'result' && rest[2]) {
        return { name: 'quiz-result', params: { slug: topicSlug, attemptId: rest[2] } };
      }
      if (action === 'play') {
        return { name: 'quiz', params: { slug: topicSlug } };
      }
      return { name: 'quiz-start', params: { slug: topicSlug } };
    }
    case 'history':
      if (rest[0]) return { name: 'history-detail', params: { attemptId: rest[0] } };
      return { name: 'history', params: {} };
    case 'login':
      return { name: 'login', params: {} };
    case 'register':
      return { name: 'register', params: {} };
    default:
      return { name: 'topics', params: {} };
  }
}

export function routeToHash(route: Route): string {
  switch (route.name) {
    case 'topics': return '#/topics';
    case 'topic': return `#/topic/${route.params.slug}`;
    case 'lesson': return `#/lesson/${route.params.id}`;
    case 'quiz-start': return `#/quiz/${route.params.slug}`;
    case 'quiz': return `#/quiz/${route.params.slug}/play`;
    case 'quiz-result': return `#/quiz/${route.params.slug}/result/${route.params.attemptId}`;
    case 'history': return '#/history';
    case 'history-detail': return `#/history/${route.params.attemptId}`;
    case 'login': return '#/login';
    case 'register': return '#/register';
  }
}

export function useRouter() {
  const [route, setRoute] = useState<Route>(() => parseHash(window.location.hash));

  useEffect(() => {
    const onHash = () => {
      setRoute(parseHash(window.location.hash));
      window.scrollTo({ top: 0, behavior: 'instant' as ScrollBehavior });
    };
    window.addEventListener('hashchange', onHash);
    return () => window.removeEventListener('hashchange', onHash);
  }, []);

  const navigate = useCallback((to: Route) => {
    const hash = routeToHash(to);
    if (hash === window.location.hash) {
      window.scrollTo({ top: 0, behavior: 'smooth' });
    } else {
      window.location.hash = hash;
    }
  }, []);

  return { route, navigate };
}
