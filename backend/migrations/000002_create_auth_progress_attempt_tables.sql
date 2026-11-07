-- Users, sessions, lesson progress, and quiz attempt tables.
-- Fields follow product-spec.md section 9.

CREATE TABLE IF NOT EXISTS users (
    id            TEXT PRIMARY KEY,
    email         TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    is_maintainer BOOLEAN NOT NULL DEFAULT FALSE,
    created_at    TIMESTAMPTZ NOT NULL,
    UNIQUE (email)
);

CREATE TABLE IF NOT EXISTS sessions (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    csrf_token TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS sessions_user_id_idx ON sessions (user_id);
CREATE INDEX IF NOT EXISTS sessions_expires_at_idx ON sessions (expires_at);

CREATE TABLE IF NOT EXISTS lesson_progress (
    user_id      TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    lesson_id    TEXT NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
    completed_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (user_id, lesson_id)
);

CREATE TABLE IF NOT EXISTS quiz_attempts (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    topic_id   TEXT NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
    score      INTEGER NOT NULL,
    total      INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS quiz_attempt_answers (
    attempt_id           TEXT NOT NULL REFERENCES quiz_attempts(id) ON DELETE CASCADE,
    position             INTEGER NOT NULL,
    question_id          TEXT REFERENCES questions(id) ON DELETE SET NULL,
    question_text        TEXT NOT NULL,
    selected_option_text TEXT NOT NULL DEFAULT '',
    correct_option_text  TEXT NOT NULL,
    is_correct           BOOLEAN NOT NULL,
    PRIMARY KEY (attempt_id, position)
);
