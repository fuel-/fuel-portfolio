// Package store persists contact-form inquiries to SQLite so a bug or
// missing email config can never silently eat a lead.
package store

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type Inquiry struct {
	ID        int64
	Name      string
	Email     string
	Company   string
	Kind      string // "hiring" | "contract" | "other"
	Message   string
	CreatedAt time.Time
}

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	const schema = `CREATE TABLE IF NOT EXISTS inquiries (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		name       TEXT NOT NULL,
		email      TEXT NOT NULL,
		company    TEXT NOT NULL DEFAULT '',
		kind       TEXT NOT NULL,
		message    TEXT NOT NULL,
		created_at TEXT NOT NULL
	)`
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("create schema: %w", err)
	}
	return &Store{db: db}, nil
}

func (s *Store) SaveInquiry(q Inquiry) (int64, error) {
	res, err := s.db.Exec(
		`INSERT INTO inquiries (name, email, company, kind, message, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		q.Name, q.Email, q.Company, q.Kind, q.Message,
		time.Now().UTC().Format(time.RFC3339),
	)
	if err != nil {
		return 0, fmt.Errorf("insert inquiry: %w", err)
	}
	return res.LastInsertId()
}

func (s *Store) ListInquiries() ([]Inquiry, error) {
	rows, err := s.db.Query(
		`SELECT id, name, email, company, kind, message, created_at
		 FROM inquiries ORDER BY id DESC`)
	if err != nil {
		return nil, fmt.Errorf("query inquiries: %w", err)
	}
	defer rows.Close()

	var out []Inquiry
	for rows.Next() {
		var q Inquiry
		var created string
		if err := rows.Scan(&q.ID, &q.Name, &q.Email, &q.Company, &q.Kind, &q.Message, &created); err != nil {
			return nil, fmt.Errorf("scan inquiry: %w", err)
		}
		q.CreatedAt, _ = time.Parse(time.RFC3339, created)
		out = append(out, q)
	}
	return out, rows.Err()
}

func (s *Store) Close() error { return s.db.Close() }
