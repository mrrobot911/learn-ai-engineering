import { afterEach } from 'vitest';
import { installCookieJar, resetCookieJar } from './harness';
import { resetHttpState } from '../../src/api/http';

installCookieJar();

// Each test gets a fresh cookie jar and cleared in-memory HTTP client state
// (CSRF token, caches). The backend/database state is managed by the suite
// harness command; tests use unique emails so they do not conflict.
afterEach(() => {
  resetCookieJar();
  resetHttpState();
});
