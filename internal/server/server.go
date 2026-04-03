package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/stockyard-dev/stockyard-handbook/internal/store"
)

type Server struct {
	db     *store.DB
	mux    *http.ServeMux
	limits Limits
}

func New(db *store.DB, limits Limits) *Server {
	s := &Server{db: db, mux: http.NewServeMux(), limits: limits}

	s.mux.HandleFunc("GET /api/spaces", s.listSpaces)
	s.mux.HandleFunc("POST /api/spaces", s.createSpace)
	s.mux.HandleFunc("GET /api/spaces/{id}", s.getSpace)
	s.mux.HandleFunc("PUT /api/spaces/{id}", s.updateSpace)
	s.mux.HandleFunc("DELETE /api/spaces/{id}", s.deleteSpace)
	s.mux.HandleFunc("GET /api/spaces/{id}/tree", s.pageTree)

	s.mux.HandleFunc("GET /api/pages", s.listPages)
	s.mux.HandleFunc("POST /api/pages", s.createPage)
	s.mux.HandleFunc("GET /api/pages/{id}", s.getPage)
	s.mux.HandleFunc("PUT /api/pages/{id}", s.updatePage)
	s.mux.HandleFunc("DELETE /api/pages/{id}", s.deletePage)

	s.mux.HandleFunc("GET /api/pages/{id}/revisions", s.listRevisions)
	s.mux.HandleFunc("GET /api/revisions/{id}", s.getRevision)

	s.mux.HandleFunc("GET /api/pages/{id}/comments", s.listComments)
	s.mux.HandleFunc("POST /api/pages/{id}/comments", s.createComment)
	s.mux.HandleFunc("DELETE /api/comments/{id}", s.deleteComment)

	s.mux.HandleFunc("GET /api/search", s.search)
	s.mux.HandleFunc("GET /api/stats", s.stats)
	s.mux.HandleFunc("GET /api/health", s.health)

	s.mux.HandleFunc("GET /ui", s.dashboard)
	s.mux.HandleFunc("GET /ui/", s.dashboard)
	s.mux.HandleFunc("GET /", s.root)
s.mux.HandleFunc("GET /api/tier",func(w http.ResponseWriter,r *http.Request){wj(w,200,map[string]any{"tier":s.limits.Tier,"upgrade_url":"https://stockyard.dev/handbook/"})})
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.mux.ServeHTTP(w, r) }
func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json"); w.WriteHeader(code); json.NewEncoder(w).Encode(v)
}
func writeErr(w http.ResponseWriter, code int, msg string) { writeJSON(w, code, map[string]string{"error": msg}) }
func (s *Server) root(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" { http.NotFound(w, r); return }
	http.Redirect(w, r, "/ui", http.StatusFound)
}

func (s *Server) listSpaces(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, map[string]any{"spaces": orEmpty(s.db.ListSpaces())}) }
func (s *Server) createSpace(w http.ResponseWriter, r *http.Request) {
	var sp store.Space
	if err := json.NewDecoder(r.Body).Decode(&sp); err != nil { writeErr(w, 400, "invalid json"); return }
	if sp.Name == "" { writeErr(w, 400, "name required"); return }
	if err := s.db.CreateSpace(&sp); err != nil { writeErr(w, 500, err.Error()); return }
	writeJSON(w, 201, sp)
}
func (s *Server) getSpace(w http.ResponseWriter, r *http.Request) {
	sp := s.db.GetSpace(r.PathValue("id")); if sp == nil { writeErr(w, 404, "not found"); return }; writeJSON(w, 200, sp)
}
func (s *Server) updateSpace(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id"); ex := s.db.GetSpace(id); if ex == nil { writeErr(w, 404, "not found"); return }
	var sp store.Space; json.NewDecoder(r.Body).Decode(&sp)
	if sp.Name == "" { sp.Name = ex.Name }; if sp.Slug == "" { sp.Slug = ex.Slug }
	s.db.UpdateSpace(id, &sp); writeJSON(w, 200, s.db.GetSpace(id))
}
func (s *Server) deleteSpace(w http.ResponseWriter, r *http.Request) {
	s.db.DeleteSpace(r.PathValue("id")); writeJSON(w, 200, map[string]string{"deleted": "ok"})
}
func (s *Server) pageTree(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{"tree": orEmpty(s.db.PageTree(r.PathValue("id")))})
}

func (s *Server) listPages(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	writeJSON(w, 200, map[string]any{"pages": orEmpty(s.db.ListPages(q.Get("space_id"), q.Get("parent_id")))})
}
func (s *Server) createPage(w http.ResponseWriter, r *http.Request) {
	var p store.Page; if err := json.NewDecoder(r.Body).Decode(&p); err != nil { writeErr(w, 400, "invalid json"); return }
	if p.Title == "" { writeErr(w, 400, "title required"); return }
	if p.SpaceID == "" { writeErr(w, 400, "space_id required"); return }
	if err := s.db.CreatePage(&p); err != nil { writeErr(w, 500, err.Error()); return }
	writeJSON(w, 201, s.db.GetPage(p.ID))
}
func (s *Server) getPage(w http.ResponseWriter, r *http.Request) {
	p := s.db.GetPage(r.PathValue("id")); if p == nil { writeErr(w, 404, "not found"); return }; writeJSON(w, 200, p)
}
func (s *Server) updatePage(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id"); ex := s.db.GetPage(id); if ex == nil { writeErr(w, 404, "not found"); return }
	var p store.Page; json.NewDecoder(r.Body).Decode(&p)
	if p.Title == "" { p.Title = ex.Title }; if p.Body == "" { p.Body = ex.Body }
	if p.Status == "" { p.Status = ex.Status }; if p.Slug == "" { p.Slug = ex.Slug }
	if p.SpaceID == "" { p.SpaceID = ex.SpaceID }
	author := p.Author; if author == "" { author = ex.Author }
	s.db.UpdatePage(id, &p, author); writeJSON(w, 200, s.db.GetPage(id))
}
func (s *Server) deletePage(w http.ResponseWriter, r *http.Request) {
	s.db.DeletePage(r.PathValue("id")); writeJSON(w, 200, map[string]string{"deleted": "ok"})
}

func (s *Server) listRevisions(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{"revisions": orEmpty(s.db.ListRevisions(r.PathValue("id")))})
}
func (s *Server) getRevision(w http.ResponseWriter, r *http.Request) {
	rev := s.db.GetRevision(r.PathValue("id")); if rev == nil { writeErr(w, 404, "not found"); return }; writeJSON(w, 200, rev)
}

func (s *Server) listComments(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{"comments": orEmpty(s.db.ListComments(r.PathValue("id")))})
}
func (s *Server) createComment(w http.ResponseWriter, r *http.Request) {
	pid := r.PathValue("id"); if s.db.GetPage(pid) == nil { writeErr(w, 404, "page not found"); return }
	var c store.Comment; json.NewDecoder(r.Body).Decode(&c)
	if c.Body == "" { writeErr(w, 400, "body required"); return }
	c.PageID = pid; s.db.CreateComment(&c); writeJSON(w, 201, c)
}
func (s *Server) deleteComment(w http.ResponseWriter, r *http.Request) {
	s.db.DeleteComment(r.PathValue("id")); writeJSON(w, 200, map[string]string{"deleted": "ok"})
}

func (s *Server) search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q"); sid := r.URL.Query().Get("space_id")
	if q == "" { writeJSON(w, 200, map[string]any{"pages": []any{}}); return }
	writeJSON(w, 200, map[string]any{"pages": orEmpty(s.db.SearchPages(sid, q))})
}
func (s *Server) stats(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, s.db.Stats()) }
func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	st := s.db.Stats(); writeJSON(w, 200, map[string]any{"status": "ok", "service": "handbook", "pages": st.Pages, "spaces": st.Spaces})
}
func orEmpty[T any](s []T) []T { if s == nil { return []T{} }; return s }
func init() { log.SetFlags(log.LstdFlags | log.Lshortfile) }
