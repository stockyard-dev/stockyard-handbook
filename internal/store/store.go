package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct { db *sql.DB }

type Page struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Slug         string   `json:"slug"`
	Content      string   `json:"content"`
	Author       string   `json:"author"`
	CreatedAt    string   `json:"created_at"`
}

func Open(dataDir string) (*DB, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}
	dsn := filepath.Join(dataDir, "handbook.db") + "?_journal_mode=WAL&_busy_timeout=5000"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS pages (
			id TEXT PRIMARY KEY,\n\t\t\ttitle TEXT DEFAULT '',\n\t\t\tslug TEXT DEFAULT '',\n\t\t\tcontent TEXT DEFAULT '',\n\t\t\tauthor TEXT DEFAULT '',
			created_at TEXT DEFAULT (datetime('now'))
		)`)
	if err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return &DB{db: db}, nil
}

func (d *DB) Close() error { return d.db.Close() }

func genID() string { return fmt.Sprintf("%d", time.Now().UnixNano()) }

func (d *DB) Create(e *Page) error {
	e.ID = genID()
	e.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	_, err := d.db.Exec(`INSERT INTO pages (id, title, slug, content, author, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		e.ID, e.Title, e.Slug, e.Content, e.Author, e.CreatedAt)
	return err
}

func (d *DB) Get(id string) *Page {
	row := d.db.QueryRow(`SELECT id, title, slug, content, author, created_at FROM pages WHERE id=?`, id)
	var e Page
	if err := row.Scan(&e.ID, &e.Title, &e.Slug, &e.Content, &e.Author, &e.CreatedAt); err != nil {
		return nil
	}
	return &e
}

func (d *DB) List() []Page {
	rows, err := d.db.Query(`SELECT id, title, slug, content, author, created_at FROM pages ORDER BY created_at DESC`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var result []Page
	for rows.Next() {
		var e Page
		if err := rows.Scan(&e.ID, &e.Title, &e.Slug, &e.Content, &e.Author, &e.CreatedAt); err != nil {
			continue
		}
		result = append(result, e)
	}
	return result
}

func (d *DB) Delete(id string) error {
	_, err := d.db.Exec(`DELETE FROM pages WHERE id=?`, id)
	return err
}

func (d *DB) Count() int {
	var n int
	d.db.QueryRow(`SELECT COUNT(*) FROM pages`).Scan(&n)
	return n
}
