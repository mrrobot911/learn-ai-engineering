package migrate

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
)

// Apply reads numbered *.sql migration files from fsys, applies any that have
// not been recorded in schema_migrations, and records each applied migration.
// It is safe to call repeatedly: applying twice leaves the same schema.
func Apply(ctx context.Context, conn *pgx.Conn, fsys fs.FS) error {
	if _, err := conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)
	`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	applied, err := appliedVersions(ctx, conn)
	if err != nil {
		return err
	}

	files, err := migrationFiles(fsys)
	if err != nil {
		return err
	}

	for _, name := range files {
		version := migrationVersion(name)
		if applied[version] {
			continue
		}
		body, err := fs.ReadFile(fsys, name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		if err := applyOne(ctx, conn, version, string(body)); err != nil {
			return fmt.Errorf("apply migration %s: %w", name, err)
		}
	}
	return nil
}

func appliedVersions(ctx context.Context, conn *pgx.Conn) (map[string]bool, error) {
	rows, err := conn.Query(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("list schema_migrations: %w", err)
	}
	defer rows.Close()

	out := make(map[string]bool)
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, fmt.Errorf("scan schema_migrations: %w", err)
		}
		out[v] = true
	}
	return out, rows.Err()
}

func migrationFiles(fsys fs.FS) ([]string, error) {
	matches, err := fs.Glob(fsys, "*.sql")
	if err != nil {
		return nil, fmt.Errorf("glob migrations: %w", err)
	}
	sort.Strings(matches)
	return matches, nil
}

func migrationVersion(name string) string {
	base := filepath.Base(name)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

func applyOne(ctx context.Context, conn *pgx.Conn, version, body string) error {
	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	for _, stmt := range splitStatements(body) {
		if _, err := tx.Exec(ctx, stmt); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, version); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func splitStatements(sql string) []string {
	var (
		out      []string
		start    int
		inLine   bool
		inQuote  bool
		inDollar bool
		dollar   string
	)
	for i := 0; i < len(sql); i++ {
		ch := sql[i]
		if inLine {
			if ch == '\n' {
				inLine = false
			}
			continue
		}
		if inDollar {
			if ch == '$' && strings.HasPrefix(sql[i:], dollar) {
				i += len(dollar) - 1
				inDollar = false
				dollar = ""
			}
			continue
		}
		if inQuote {
			if ch == '\'' {
				if i+1 < len(sql) && sql[i+1] == '\'' {
					i++ // escaped quote ''
					continue
				}
				inQuote = false
			}
			continue
		}
		// normal state
		if ch == '-' && i+1 < len(sql) && sql[i+1] == '-' {
			inLine = true
			i++
			continue
		}
		if ch == '\'' {
			inQuote = true
			continue
		}
		if ch == '$' {
			if tag, ok := dollarTag(sql[i:]); ok {
				inDollar = true
				dollar = tag
				i += len(tag) - 1
				continue
			}
		}
		if ch == ';' {
			stmt := strings.TrimSpace(sql[start:i])
			if hasCode(stmt) {
				out = append(out, stmt)
			}
			start = i + 1
		}
	}
	if tail := strings.TrimSpace(sql[start:]); hasCode(tail) {
		out = append(out, tail)
	}
	return out
}

// dollarTag reports whether s starts with a PostgreSQL dollar-quote tag
// ($$, $tag$, etc.) and returns the tag including both $ delimiters.
func dollarTag(s string) (string, bool) {
	if len(s) == 0 || s[0] != '$' {
		return "", false
	}
	for i := 1; i < len(s); i++ {
		c := s[i]
		if c == '$' {
			return s[:i+1], true
		}
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' {
			continue
		}
		return "", false
	}
	return "", false
}

func hasCode(stmt string) bool {
	if stmt == "" {
		return false
	}
	for _, line := range strings.Split(stmt, "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "--") {
			return true
		}
	}
	return false
}
