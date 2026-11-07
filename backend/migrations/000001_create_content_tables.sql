-- Content tables for topics, lessons, questions, and answer options.
-- Fields follow product-spec.md section 9; every table has status.

CREATE TABLE IF NOT EXISTS topics (
    id          TEXT PRIMARY KEY,
    slug        TEXT NOT NULL,
    title       TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    position    INTEGER NOT NULL,
    status      TEXT NOT NULL CHECK (status IN ('draft', 'published')),
    UNIQUE (slug)
);

CREATE TABLE IF NOT EXISTS lessons (
    id       TEXT PRIMARY KEY,
    topic_id TEXT NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
    title    TEXT NOT NULL,
    body     TEXT NOT NULL DEFAULT '',
    position INTEGER NOT NULL,
    status   TEXT NOT NULL CHECK (status IN ('draft', 'published')),
    UNIQUE (topic_id, position)
);

CREATE TABLE IF NOT EXISTS questions (
    id          TEXT PRIMARY KEY,
    topic_id    TEXT NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
    text        TEXT NOT NULL,
    explanation TEXT NOT NULL DEFAULT '',
    status      TEXT NOT NULL CHECK (status IN ('draft', 'published'))
);

CREATE TABLE IF NOT EXISTS answer_options (
    id          TEXT PRIMARY KEY,
    question_id TEXT NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    text        TEXT NOT NULL,
    position    INTEGER NOT NULL,
    is_correct  BOOLEAN NOT NULL DEFAULT FALSE,
    status      TEXT NOT NULL CHECK (status IN ('draft', 'published')),
    UNIQUE (question_id, position)
);
