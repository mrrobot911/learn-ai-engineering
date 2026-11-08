/**
 * Harness for frontend integration tests against a real backend and seeded test DB.
 * Keeps a cookie jar for HttpOnly session cookies and provides unique identities.
 */

type CookieJar = Map<string, string>;

let jar: CookieJar = new Map();

function parseSetCookie(
  header: string,
): { name: string; value: string } | null {
  const eq = header.indexOf('=');
  if (eq <= 0) return null;
  const name = header.slice(0, eq).trim();
  const rest = header.slice(eq + 1);
  const semi = rest.indexOf(';');
  const value = (semi === -1 ? rest : rest.slice(0, semi)).trim();
  return { name, value };
}

function setCookieHeaders(headers: Headers): string[] {
  const anyHeaders = headers as Headers & { getSetCookie?: () => string[] };
  if (typeof anyHeaders.getSetCookie === 'function') {
    return anyHeaders.getSetCookie();
  }
  const single = headers.get('set-cookie');
  return single ? [single] : [];
}

function cookieHeader(): string | null {
  if (jar.size === 0) return null;
  return Array.from(jar.entries())
    .map(([k, v]) => `${k}=${v}`)
    .join('; ');
}

const originalFetch = globalThis.fetch.bind(globalThis);

export function resetCookieJar(): void {
  jar = new Map();
}

export function installCookieJar(): void {
  globalThis.fetch = async (
    input: RequestInfo | URL,
    init?: RequestInit,
  ): Promise<Response> => {
    const headers = new Headers(
      init?.headers ?? (input instanceof Request ? input.headers : undefined),
    );
    const cookie = cookieHeader();
    if (cookie) headers.set('Cookie', cookie);

    const nextInit: RequestInit = { ...init, headers };
    // The real browser sends cookies; credentials is advisory in Node fetch.
    const response = await originalFetch(input, nextInit);

    for (const header of setCookieHeaders(response.headers)) {
      const parsed = parseSetCookie(header);
      if (!parsed) continue;
      if (parsed.value === '' || /;\s*max-age=0/i.test(header)) {
        jar.delete(parsed.name);
      } else {
        jar.set(parsed.name, parsed.value);
      }
    }
    return response;
  };
}

export function restoreFetch(): void {
  globalThis.fetch = originalFetch;
}

const runId = `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
let emailCounter = 0;

/** Unique email per call, stable within a test run to avoid cross-run conflicts. */
export function uniqueEmail(prefix = 'user'): string {
  emailCounter += 1;
  return `${prefix}-${runId}-${emailCounter}@example.test`;
}

export const testPassword = 'password123';

export function apiBaseUrl(): string {
  const env = (
    globalThis as { process?: { env?: Record<string, string | undefined> } }
  ).process?.env;
  return env?.TEST_API_BASE_URL ?? 'http://127.0.0.1:18080';
}
