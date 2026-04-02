package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct{ db *sql.DB }

type Space struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	Icon      string `json:"icon,omitempty"`
	CreatedAt string `json:"created_at"`
	PageCount int    `json:"page_count"`
}

type Page struct {
	ID           string `json:"id"`
	SpaceID      string `json:"space_id"`
	ParentID     string `json:"parent_id,omitempty"`
	Title        string `json:"title"`
	Slug         string `json:"slug"`
	Body         string `json:"body"`
	Status       string `json:"status"` // draft, published
	Author       string `json:"author,omitempty"`
	Position     int    `json:"position"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
	WordCount    int    `json:"word_count"`
	RevisionCount int   `json:"revision_count"`
	CommentCount int    `json:"comment_count"`
	ChildCount   int    `json:"child_count"`
}

type Revision struct {
	ID        string `json:"id"`
	PageID    string `json:"page_id"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	Author    string `json:"author,omitempty"`
	CreatedAt string `json:"created_at"`
}

type Comment struct {
	ID        string `json:"id"`
	PageID    string `json:"page_id"`
	Author    string `json:"author,omitempty"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
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
	for _, q := range []string{
		`CREATE TABLE IF NOT EXISTS spaces (
			id TEXT PRIMARY KEY, name TEXT NOT NULL, slug TEXT UNIQUE NOT NULL,
			icon TEXT DEFAULT '', created_at TEXT DEFAULT (datetime('now'))
		)`,
		`CREATE TABLE IF NOT EXISTS pages (
			id TEXT PRIMARY KEY, space_id TEXT NOT NULL REFERENCES spaces(id),
			parent_id TEXT DEFAULT '', title TEXT NOT NULL, slug TEXT DEFAULT '',
			body TEXT DEFAULT '', status TEXT DEFAULT 'published',
			author TEXT DEFAULT '', position INTEGER DEFAULT 0,
			created_at TEXT DEFAULT (datetime('now')), updated_at TEXT DEFAULT (datetime('now'))
		)`,
		`CREATE TABLE IF NOT EXISTS revisions (
			id TEXT PRIMARY KEY, page_id TEXT NOT NULL REFERENCES pages(id) ON DELETE CASCADE,
			title TEXT DEFAULT '', body TEXT DEFAULT '', author TEXT DEFAULT '',
			created_at TEXT DEFAULT (datetime('now'))
		)`,
		`CREATE TABLE IF NOT EXISTS comments (
			id TEXT PRIMARY KEY, page_id TEXT NOT NULL REFERENCES pages(id) ON DELETE CASCADE,
			author TEXT DEFAULT '', body TEXT NOT NULL,
			created_at TEXT DEFAULT (datetime('now'))
		)`,
		`CREATE INDEX IF NOT EXISTS idx_pages_space ON pages(space_id)`,
		`CREATE INDEX IF NOT EXISTS idx_pages_parent ON pages(parent_id)`,
		`CREATE INDEX IF NOT EXISTS idx_revisions_page ON revisions(page_id)`,
		`CREATE INDEX IF NOT EXISTS idx_comments_page ON comments(page_id)`,
	} {
		if _, err := db.Exec(q); err != nil {
			return nil, fmt.Errorf("migrate: %w", err)
		}
	}
	return &DB{db: db}, nil
}

func (d *DB) Close() error { return d.db.Close() }
func genID() string        { return fmt.Sprintf("%d", time.Now().UnixNano()) }
func now() string          { return time.Now().UTC().Format(time.RFC3339) }

// ── Spaces ──

func (d *DB) CreateSpace(s *Space) error {
	s.ID = genID()
	s.CreatedAt = now()
	if s.Slug == "" {
		s.Slug = strings.ToLower(strings.ReplaceAll(strings.TrimSpace(s.Name), " ", "-"))
	}
	_, err := d.db.Exec(`INSERT INTO spaces (id,name,slug,icon,created_at) VALUES (?,?,?,?,?)`,
		s.ID, s.Name, s.Slug, s.Icon, s.CreatedAt)
	return err
}

func (d *DB) GetSpace(id string) *Space {
	var s Space
	if err := d.db.QueryRow(`SELECT id,name,slug,icon,created_at FROM spaces WHERE id=?`, id).Scan(&s.ID, &s.Name, &s.Slug, &s.Icon, &s.CreatedAt); err != nil {
		return nil
	}
	d.db.QueryRow(`SELECT COUNT(*) FROM pages WHERE space_id=?`, id).Scan(&s.PageCount)
	return &s
}

func (d *DB) ListSpaces() []Space {
	rows, err := d.db.Query(`SELECT id,name,slug,icon,created_at FROM spaces ORDER BY name ASC`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []Space
	for rows.Next() {
		var s Space
		if err := rows.Scan(&s.ID, &s.Name, &s.Slug, &s.Icon, &s.CreatedAt); err != nil {
			continue
		}
		d.db.QueryRow(`SELECT COUNT(*) FROM pages WHERE space_id=?`, s.ID).Scan(&s.PageCount)
		out = append(out, s)
	}
	return out
}

func (d *DB) UpdateSpace(id string, s *Space) error {
	_, err := d.db.Exec(`UPDATE spaces SET name=?,slug=?,icon=? WHERE id=?`, s.Name, s.Slug, s.Icon, id)
	return err
}

func (d *DB) DeleteSpace(id string) error {
	d.db.Exec(`DELETE FROM comments WHERE page_id IN (SELECT id FROM pages WHERE space_id=?)`, id)
	d.db.Exec(`DELETE FROM revisions WHERE page_id IN (SELECT id FROM pages WHERE space_id=?)`, id)
	d.db.Exec(`DELETE FROM pages WHERE space_id=?`, id)
	_, err := d.db.Exec(`DELETE FROM spaces WHERE id=?`, id)
	return err
}

// ── Pages ──

func (d *DB) hydratePage(p *Page) {
	p.WordCount = len(strings.Fields(p.Body))
	d.db.QueryRow(`SELECT COUNT(*) FROM revisions WHERE page_id=?`, p.ID).Scan(&p.RevisionCount)
	d.db.QueryRow(`SELECT COUNT(*) FROM comments WHERE page_id=?`, p.ID).Scan(&p.CommentCount)
	d.db.QueryRow(`SELECT COUNT(*) FROM pages WHERE parent_id=?`, p.ID).Scan(&p.ChildCount)
}

func (d *DB) CreatePage(p *Page) error {
	p.ID = genID()
	p.CreatedAt = now()
	p.UpdatedAt = p.CreatedAt
	if p.Slug == "" {
		p.Slug = strings.ToLower(strings.ReplaceAll(strings.TrimSpace(p.Title), " ", "-"))
	}
	if p.Status == "" {
		p.Status = "published"
	}
	_, err := d.db.Exec(`INSERT INTO pages (id,space_id,parent_id,title,slug,body,status,author,position,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		p.ID, p.SpaceID, p.ParentID, p.Title, p.Slug, p.Body, p.Status, p.Author, p.Position, p.CreatedAt, p.UpdatedAt)
	return err
}

func (d *DB) scanPage(s interface{ Scan(...any) error }) *Page {
	var p Page
	if err := s.Scan(&p.ID, &p.SpaceID, &p.ParentID, &p.Title, &p.Slug, &p.Body, &p.Status, &p.Author, &p.Position, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil
	}
	d.hydratePage(&p)
	return &p
}

const pageCols = `id,space_id,parent_id,title,slug,body,status,author,position,created_at,updated_at`

func (d *DB) GetPage(id string) *Page {
	return d.scanPage(d.db.QueryRow(`SELECT `+pageCols+` FROM pages WHERE id=?`, id))
}

func (d *DB) ListPages(spaceID, parentID string) []Page {
	q := `SELECT ` + pageCols + ` FROM pages WHERE space_id=? AND parent_id=? ORDER BY position ASC, title ASC`
	rows, err := d.db.Query(q, spaceID, parentID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []Page
	for rows.Next() {
		if p := d.scanPage(rows); p != nil {
			out = append(out, *p)
		}
	}
	return out
}

func (d *DB) SearchPages(spaceID, query string) []Page {
	where := "(title LIKE ? OR body LIKE ?)"
	args := []any{"%" + query + "%", "%" + query + "%"}
	if spaceID != "" {
		where += " AND space_id=?"
		args = append(args, spaceID)
	}
	rows, err := d.db.Query(`SELECT `+pageCols+` FROM pages WHERE `+where+` ORDER BY updated_at DESC LIMIT 50`, args...)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []Page
	for rows.Next() {
		if p := d.scanPage(rows); p != nil {
			out = append(out, *p)
		}
	}
	return out
}

func (d *DB) UpdatePage(id string, p *Page, author string) error {
	// save revision of current state
	old := d.GetPage(id)
	if old != nil {
		d.db.Exec(`INSERT INTO revisions (id,page_id,title,body,author,created_at) VALUES (?,?,?,?,?,?)`,
			genID(), id, old.Title, old.Body, author, now())
	}
	p.UpdatedAt = now()
	_, err := d.db.Exec(`UPDATE pages SET title=?,slug=?,body=?,status=?,author=?,parent_id=?,position=?,updated_at=? WHERE id=?`,
		p.Title, p.Slug, p.Body, p.Status, p.Author, p.ParentID, p.Position, p.UpdatedAt, id)
	return err
}

func (d *DB) DeletePage(id string) error {
	// reparent children
	d.db.Exec(`UPDATE pages SET parent_id='' WHERE parent_id=?`, id)
	d.db.Exec(`DELETE FROM comments WHERE page_id=?`, id)
	d.db.Exec(`DELETE FROM revisions WHERE page_id=?`, id)
	_, err := d.db.Exec(`DELETE FROM pages WHERE id=?`, id)
	return err
}

// ── Revisions ──

func (d *DB) ListRevisions(pageID string) []Revision {
	rows, err := d.db.Query(`SELECT id,page_id,title,body,author,created_at FROM revisions WHERE page_id=? ORDER BY created_at DESC`, pageID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []Revision
	for rows.Next() {
		var r Revision
		if err := rows.Scan(&r.ID, &r.PageID, &r.Title, &r.Body, &r.Author, &r.CreatedAt); err != nil {
			continue
		}
		out = append(out, r)
	}
	return out
}

func (d *DB) GetRevision(id string) *Revision {
	var r Revision
	if err := d.db.QueryRow(`SELECT id,page_id,title,body,author,created_at FROM revisions WHERE id=?`, id).Scan(&r.ID, &r.PageID, &r.Title, &r.Body, &r.Author, &r.CreatedAt); err != nil {
		return nil
	}
	return &r
}

// ── Comments ──

func (d *DB) CreateComment(c *Comment) error {
	c.ID = genID()
	c.CreatedAt = now()
	_, err := d.db.Exec(`INSERT INTO comments (id,page_id,author,body,created_at) VALUES (?,?,?,?,?)`,
		c.ID, c.PageID, c.Author, c.Body, c.CreatedAt)
	return err
}

func (d *DB) ListComments(pageID string) []Comment {
	rows, err := d.db.Query(`SELECT id,page_id,author,body,created_at FROM comments WHERE page_id=? ORDER BY created_at ASC`, pageID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []Comment
	for rows.Next() {
		var c Comment
		if err := rows.Scan(&c.ID, &c.PageID, &c.Author, &c.Body, &c.CreatedAt); err != nil {
			continue
		}
		out = append(out, c)
	}
	return out
}

func (d *DB) DeleteComment(id string) error {
	_, err := d.db.Exec(`DELETE FROM comments WHERE id=?`, id)
	return err
}

// ── Tree ──

type TreeNode struct {
	Page     Page       `json:"page"`
	Children []TreeNode `json:"children,omitempty"`
}

func (d *DB) PageTree(spaceID string) []TreeNode {
	return d.buildTree(spaceID, "")
}

func (d *DB) buildTree(spaceID, parentID string) []TreeNode {
	pages := d.ListPages(spaceID, parentID)
	var nodes []TreeNode
	for _, p := range pages {
		node := TreeNode{Page: p}
		if p.ChildCount > 0 {
			node.Children = d.buildTree(spaceID, p.ID)
		}
		nodes = append(nodes, node)
	}
	return nodes
}

// ── Stats ──

type Stats struct {
	Spaces    int `json:"spaces"`
	Pages     int `json:"pages"`
	Published int `json:"published"`
	Drafts    int `json:"drafts"`
	Revisions int `json:"revisions"`
	Comments  int `json:"comments"`
	Words     int `json:"words"`
}

func (d *DB) Stats() Stats {
	var s Stats
	d.db.QueryRow(`SELECT COUNT(*) FROM spaces`).Scan(&s.Spaces)
	d.db.QueryRow(`SELECT COUNT(*) FROM pages`).Scan(&s.Pages)
	d.db.QueryRow(`SELECT COUNT(*) FROM pages WHERE status='published'`).Scan(&s.Published)
	d.db.QueryRow(`SELECT COUNT(*) FROM pages WHERE status='draft'`).Scan(&s.Drafts)
	d.db.QueryRow(`SELECT COUNT(*) FROM revisions`).Scan(&s.Revisions)
	d.db.QueryRow(`SELECT COUNT(*) FROM comments`).Scan(&s.Comments)
	rows, _ := d.db.Query(`SELECT body FROM pages`)
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var b string
			rows.Scan(&b)
			s.Words += len(strings.Fields(b))
		}
	}
	return s
}
