// Package seed loads versioned learning content into the database.
package seed

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mrrobot911/learn-ai-engineering/internal/password"
)

// Report summarizes what the loader inserted or skipped.
type Report struct {
	TopicsInserted    int
	TopicsSkipped     int
	LessonsInserted   int
	LessonsSkipped    int
	QuestionsInserted int
	QuestionsSkipped  int
	OptionsInserted   int
	OptionsSkipped    int
}

// Created reports whether any records were inserted.
func (r Report) Created() int {
	return r.TopicsInserted + r.LessonsInserted + r.QuestionsInserted + r.OptionsInserted
}

func (r Report) String() string {
	return fmt.Sprintf(
		"topics: %d inserted, %d skipped; lessons: %d inserted, %d skipped; questions: %d inserted, %d skipped; options: %d inserted, %d skipped",
		r.TopicsInserted, r.TopicsSkipped,
		r.LessonsInserted, r.LessonsSkipped,
		r.QuestionsInserted, r.QuestionsSkipped,
		r.OptionsInserted, r.OptionsSkipped,
	)
}

type contentFile struct {
	Topics []topicSeed `json:"topics"`
}

type topicSeed struct {
	ID          string         `json:"id"`
	Slug        string         `json:"slug"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Position    int            `json:"position"`
	Status      string         `json:"status"`
	Lessons     []lessonSeed   `json:"lessons"`
	Questions   []questionSeed `json:"questions"`
}

type lessonSeed struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Body     string `json:"body"`
	Position int    `json:"position"`
	Status   string `json:"status"`
}

type questionSeed struct {
	ID          string       `json:"id"`
	Text        string       `json:"text"`
	Explanation string       `json:"explanation"`
	Status      string       `json:"status"`
	Options     []optionSeed `json:"options"`
}

type optionSeed struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	Position  int    `json:"position"`
	IsCorrect bool   `json:"is_correct"`
}

// Load reads seed JSON files from fsys and inserts any missing records into the
// database. It is repeatable: loading twice creates no duplicates and reports
// nothing created on the second run. When maintainerEmail and maintainerPassword
// are both non-empty, it also ensures a maintainer account exists for that email.
func Load(ctx context.Context, pool *pgxpool.Pool, fsys fs.FS, maintainerEmail, maintainerPassword string) (Report, error) {
	var report Report
	matches, err := fs.Glob(fsys, "*.json")
	if err != nil {
		return report, fmt.Errorf("glob seed files: %w", err)
	}
	for _, name := range matches {
		body, err := fs.ReadFile(fsys, name)
		if err != nil {
			return report, fmt.Errorf("read seed file %s: %w", name, err)
		}
		var cf contentFile
		if err := json.Unmarshal(body, &cf); err != nil {
			return report, fmt.Errorf("parse seed file %s: %w", name, err)
		}
		r, err := loadTopics(ctx, pool, cf.Topics)
		if err != nil {
			return report, fmt.Errorf("load seed file %s: %w", name, err)
		}
		report.TopicsInserted += r.TopicsInserted
		report.TopicsSkipped += r.TopicsSkipped
		report.LessonsInserted += r.LessonsInserted
		report.LessonsSkipped += r.LessonsSkipped
		report.QuestionsInserted += r.QuestionsInserted
		report.QuestionsSkipped += r.QuestionsSkipped
		report.OptionsInserted += r.OptionsInserted
		report.OptionsSkipped += r.OptionsSkipped
	}
	if maintainerEmail != "" || maintainerPassword != "" {
		if maintainerEmail == "" || maintainerPassword == "" {
			return report, fmt.Errorf("both maintainer email and password are required to seed a maintainer account")
		}
		if err := loadMaintainer(ctx, pool, maintainerEmail, maintainerPassword); err != nil {
			return report, err
		}
	}
	return report, nil
}

func loadMaintainer(ctx context.Context, pool *pgxpool.Pool, email, pass string) error {
	hash, err := password.Hash(pass)
	if err != nil {
		return fmt.Errorf("hash maintainer password: %w", err)
	}
	id, err := newToken()
	if err != nil {
		return fmt.Errorf("generate maintainer user id: %w", err)
	}
	tag, err := pool.Exec(ctx,
		`INSERT INTO users (id, email, password_hash, is_maintainer, created_at)
		 VALUES ($1, $2, $3, TRUE, $4)
		 ON CONFLICT (email) DO NOTHING`,
		id, email, hash, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("seed maintainer user: %w", err)
	}
	if tag.RowsAffected() == 1 {
		return nil
	}

	// The email is already registered. Seeding must never promote an existing
	// account, because anyone could have registered this email before the seed
	// ran. Accept it only if it already is the maintainer.
	var isMaintainer bool
	if err := pool.QueryRow(ctx,
		`SELECT is_maintainer FROM users WHERE email = $1`, email).Scan(&isMaintainer); err != nil {
		return fmt.Errorf("check existing maintainer user: %w", err)
	}
	if !isMaintainer {
		return fmt.Errorf("account %q already exists and is not a maintainer; refusing to promote it", email)
	}
	return nil
}

func newToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func loadTopics(ctx context.Context, pool *pgxpool.Pool, topics []topicSeed) (Report, error) {
	var report Report
	for _, t := range topics {
		tag, err := pool.Exec(ctx,
			`INSERT INTO topics (id, slug, title, description, position, status)
			 VALUES ($1, $2, $3, $4, $5, $6)
			 ON CONFLICT (id) DO NOTHING`,
			t.ID, t.Slug, t.Title, t.Description, t.Position, t.Status)
		if err != nil {
			return report, fmt.Errorf("insert topic %s: %w", t.ID, err)
		}
		if tag.RowsAffected() == 1 {
			report.TopicsInserted++
		} else {
			report.TopicsSkipped++
		}

		for _, l := range t.Lessons {
			tag, err := pool.Exec(ctx,
				`INSERT INTO lessons (id, topic_id, title, body, position, status)
				 VALUES ($1, $2, $3, $4, $5, $6)
				 ON CONFLICT (id) DO NOTHING`,
				l.ID, t.ID, l.Title, l.Body, l.Position, l.Status)
			if err != nil {
				return report, fmt.Errorf("insert lesson %s: %w", l.ID, err)
			}
			if tag.RowsAffected() == 1 {
				report.LessonsInserted++
			} else {
				report.LessonsSkipped++
			}
		}

		for _, q := range t.Questions {
			tag, err := pool.Exec(ctx,
				`INSERT INTO questions (id, topic_id, text, explanation, status)
				 VALUES ($1, $2, $3, $4, $5)
				 ON CONFLICT (id) DO NOTHING`,
				q.ID, t.ID, q.Text, q.Explanation, q.Status)
			if err != nil {
				return report, fmt.Errorf("insert question %s: %w", q.ID, err)
			}
			if tag.RowsAffected() == 1 {
				report.QuestionsInserted++
			} else {
				report.QuestionsSkipped++
			}

			for _, o := range q.Options {
				tag, err := pool.Exec(ctx,
					`INSERT INTO answer_options (id, question_id, text, position, is_correct, status)
					 VALUES ($1, $2, $3, $4, $5, $6)
					 ON CONFLICT (id) DO NOTHING`,
					o.ID, q.ID, o.Text, o.Position, o.IsCorrect, q.Status)
				if err != nil {
					return report, fmt.Errorf("insert option %s: %w", o.ID, err)
				}
				if tag.RowsAffected() == 1 {
					report.OptionsInserted++
				} else {
					report.OptionsSkipped++
				}
			}
		}
	}
	return report, nil
}
