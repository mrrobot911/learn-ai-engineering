import type { Lesson, Topic, TopicSummary } from './types';

const SNAKE_RE = /_([a-z])/g;
const CAMEL_BOUNDARY_RE = /([a-z0-9])([A-Z])/g;

/** Maps snake_case JSON keys from the contract to camelCase screen types. */
export function snakeToCamel<T>(value: unknown): T {
  if (Array.isArray(value)) return value.map((item) => snakeToCamel(item)) as T;
  if (value && typeof value === 'object') {
    const out: Record<string, unknown> = {};
    for (const [key, val] of Object.entries(value as Record<string, unknown>)) {
      out[key.replace(SNAKE_RE, (_, c: string) => c.toUpperCase())] = snakeToCamel(val);
    }
    return out as T;
  }
  return value as T;
}

/** Maps camelCase screen values to snake_case for contract request bodies. */
export function camelToSnake(value: unknown): unknown {
  if (Array.isArray(value)) return value.map((item) => camelToSnake(item));
  if (value && typeof value === 'object') {
    const out: Record<string, unknown> = {};
    for (const [key, val] of Object.entries(value as Record<string, unknown>)) {
      out[key.replace(CAMEL_BOUNDARY_RE, '$1_$2').toLowerCase()] = camelToSnake(val);
    }
    return out;
  }
  return value;
}

// Presentation fields the contract does not carry. The mapping is deterministic
// from the slug/body so the same content always renders the same way.

const TOPIC_PRESENTATION: Record<string, { icon: string; color: string }> = {
  llms: { icon: 'BrainCircuit', color: 'brand' },
  agents: { icon: 'Bot', color: 'accent' },
  mcp: { icon: 'Plug', color: 'warning' },
  rag: { icon: 'Search', color: 'brand' },
  embeddings: { icon: 'Boxes', color: 'accent' },
  evaluation: { icon: 'ClipboardCheck', color: 'warning' },
  deployment: { icon: 'Rocket', color: 'brand' },
};

const DEFAULT_PRESENTATION = { icon: 'BookOpen', color: 'brand' };

/** Words per minute used to estimate reading time. Documented rate: 200 wpm. */
export const READING_WPM = 200;

export function presentationForSlug(slug: string): { icon: string; color: string } {
  return TOPIC_PRESENTATION[slug] ?? DEFAULT_PRESENTATION;
}

/** First non-empty paragraph of a Markdown body, trimmed; empty body falls back. */
export function summaryFromBody(body: string): string {
  if (!body) return '';
  const paragraphs = body.split(/\n\s*\n/);
  for (const p of paragraphs) {
    if (/^\s*#{1,6}\s/.test(p)) continue;
    const text = p
      .replace(/```[\s\S]*?```/g, '')
      .replace(/`[^`]*`/g, '')
      .replace(/[*_~>[\]()!-]/g, ' ')
      .replace(/\s+/g, ' ')
      .trim();
    if (text) return text.length > 160 ? `${text.slice(0, 157).trimEnd()}...` : text;
  }
  return '';
}

/** Reading time in minutes from body word count at READING_WPM; empty body -> 0. */
export function durationMinFromBody(body: string): number {
  if (!body || !body.trim()) return 0;
  const words = body.trim().split(/\s+/).filter(Boolean).length;
  return Math.max(1, Math.ceil(words / READING_WPM));
}

interface RawTopicSummary {
  id: string;
  slug: string;
  title: string;
  description: string;
  position: number;
}

interface RawLesson {
  id: string;
  topicId: string;
  title: string;
  body: string;
  position: number;
}

interface RawTopicDetail extends RawTopicSummary {
  lessons: RawLesson[];
}

export function mapTopicSummary(raw: RawTopicSummary): TopicSummary {
  const presentation = presentationForSlug(raw.slug);
  return {
    id: raw.id,
    slug: raw.slug,
    title: raw.title,
    description: raw.description,
    icon: presentation.icon,
    color: presentation.color,
    position: raw.position,
  };
}

export function mapLesson(raw: RawLesson): Lesson {
  const body = raw.body ?? '';
  return {
    id: raw.id,
    topicId: raw.topicId,
    title: raw.title,
    summary: summaryFromBody(body),
    body,
    position: raw.position,
    durationMin: durationMinFromBody(body),
  };
}

export function mapTopic(raw: RawTopicDetail): Topic {
  return {
    ...mapTopicSummary(raw),
    lessons: (raw.lessons ?? []).map(mapLesson),
  };
}
