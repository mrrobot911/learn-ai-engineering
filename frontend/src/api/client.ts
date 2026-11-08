import type { ApiClient } from './types';
import { mockClient } from './mock';
import { httpClient } from './http';

export type ApiMode = 'mock' | 'http';

function resolveMode(): ApiMode {
  const raw = import.meta.env.VITE_API_MODE;
  if (import.meta.env.PROD) return 'http';
  if (raw === 'http') return 'http';
  return 'mock';
}

export const API_MODE: ApiMode = resolveMode();

export const api: ApiClient = API_MODE === 'http' ? httpClient : mockClient;
