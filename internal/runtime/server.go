package runtime

import (
"context"
"crypto/rand"
"database/sql"
"encoding/hex"
"encoding/json"
"fmt"
"io"
"net"
"net/http"
"strconv"
"strings"
"sync"
"time"

_ "modernc.org/sqlite"
)

// Server is the bindable runtime server.
type Server struct {
mu        sync.Mutex
port      int
authToken string
httpSrv   *http.Server
listener  net.Listener
db        *sql.DB
dbPath    string
}

// NewServer creates a new unstarted Server. dbPath is where the SQLite file lives.
func NewServer(dbPath string) *Server {
return &Server{dbPath: dbPath}
}

func generateToken() string {
b := make([]byte, 32)
rand.Read(b)
return hex.EncodeToString(b)
}

func (s *Server) openDB() error {
if s.db != nil {
return nil
}
db, err := sql.Open("sqlite", s.dbPath)
if err != nil {
return err
}
_, err = db.Exec(`CREATE TABLE IF NOT EXISTS notes (
id INTEGER PRIMARY KEY AUTOINCREMENT,
title TEXT NOT NULL,
body TEXT NOT NULL,
created_at DATETIME DEFAULT CURRENT_TIMESTAMP
)`)
if err != nil {
return err
}
s.db = db
return nil
}

type note struct {
ID        int64  `json:"id"`
Title     string `json:"title"`
Body      string `json:"body"`
CreatedAt string `json:"created_at"`
}

func (s *Server) authOK(r *http.Request) bool {
return r.Header.Get("Authorization") == "Bearer "+s.authToken
}

func (s *Server) buildMux() *http.ServeMux {
mux := http.NewServeMux()

mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
w.Header().Set("Content-Type", "text/html; charset=utf-8")
w.Write([]byte(testHTML))
})

mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
w.Write([]byte("ok"))
})

mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
dbStatus := "closed"
if s.db != nil {
if err := s.db.Ping(); err == nil {
dbStatus = "ok"
} else {
dbStatus = "error"
}
}
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(map[string]string{
"status": "ok",
"db":     dbStatus,
})
})

mux.HandleFunc("/api/notes", func(w http.ResponseWriter, r *http.Request) {
if !s.authOK(r) {
w.WriteHeader(401)
w.Write([]byte("unauthorized"))
return
}

switch r.Method {
case "GET":
rows, err := s.db.Query("SELECT id, title, body, created_at FROM notes ORDER BY id DESC")
if err != nil {
w.WriteHeader(500)
fmt.Fprintf(w, "query error: %v", err)
return
}
defer rows.Close()
notes := []note{}
for rows.Next() {
var n note
if err := rows.Scan(&n.ID, &n.Title, &n.Body, &n.CreatedAt); err != nil {
continue
}
notes = append(notes, n)
}
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(notes)

case "POST":
body, _ := io.ReadAll(r.Body)
var n note
if err := json.Unmarshal(body, &n); err != nil {
w.WriteHeader(400)
w.Write([]byte("bad json"))
return
}
res, err := s.db.Exec("INSERT INTO notes (title, body) VALUES (?, ?)", n.Title, n.Body)
if err != nil {
w.WriteHeader(500)
fmt.Fprintf(w, "insert error: %v", err)
return
}
n.ID, _ = res.LastInsertId()
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(201)
json.NewEncoder(w).Encode(n)

default:
w.WriteHeader(405)
}
})

mux.HandleFunc("/api/notes/", func(w http.ResponseWriter, r *http.Request) {
if !s.authOK(r) {
w.WriteHeader(401)
w.Write([]byte("unauthorized"))
return
}
if r.Method != "DELETE" {
w.WriteHeader(405)
return
}
idStr := strings.TrimPrefix(r.URL.Path, "/api/notes/")
id, err := strconv.ParseInt(idStr, 10, 64)
if err != nil {
w.WriteHeader(400)
w.Write([]byte("bad id"))
return
}
_, err = s.db.Exec("DELETE FROM notes WHERE id = ?", id)
if err != nil {
w.WriteHeader(500)
fmt.Fprintf(w, "delete error: %v", err)
return
}
w.WriteHeader(204)
})

return mux
}

// Start binds a listener on 127.0.0.1:0 and serves. Returns port (0 on failure).
func (s *Server) Start() int {
s.mu.Lock()
defer s.mu.Unlock()

if s.httpSrv != nil {
return s.port
}

if err := s.openDB(); err != nil {
return 0
}

s.authToken = generateToken()

l, err := net.Listen("tcp", "127.0.0.1:0")
if err != nil {
return 0
}

s.port = l.Addr().(*net.TCPAddr).Port
s.listener = l
s.httpSrv = &http.Server{Handler: s.buildMux()}
srv := s.httpSrv

go func() {
_ = srv.Serve(l)
}()

return s.port
}

// Port returns the bound port, or 0.
func (s *Server) Port() int {
s.mu.Lock()
defer s.mu.Unlock()
return s.port
}

// Token returns the bearer token, or empty string.
func (s *Server) Token() string {
s.mu.Lock()
defer s.mu.Unlock()
return s.authToken
}

// Stop gracefully shuts the server down.
func (s *Server) Stop() {
s.mu.Lock()
defer s.mu.Unlock()
if s.httpSrv != nil {
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel()
_ = s.httpSrv.Shutdown(ctx)
s.httpSrv = nil
s.listener = nil
s.port = 0
}
if s.db != nil {
_ = s.db.Close()
s.db = nil
}
}
