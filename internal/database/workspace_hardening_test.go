package database

import (
"os"
"path/filepath"
"testing"
)


func TestHardening_WorkspacePathsIsolated(t *testing.T) {
dataDir := t.TempDir()
w1 := NewWorkspace(dataDir, 1)
w2 := NewWorkspace(dataDir, 2)
if w1.Root == w2.Root {
t.Fatalf("expected different Roots, both = %s", w1.Root)
}
if w1.Source == w2.Source {
t.Fatalf("expected different Sources")
}
if w1.Prepared == w2.Prepared {
t.Fatalf("expected different Prepared dirs")
}
if w1.Build == w2.Build {
t.Fatalf("expected different Build dirs")
}
}

func TestHardening_WorkspaceSubdirsNested(t *testing.T) {
dataDir := t.TempDir()
w := NewWorkspace(dataDir, 42)
prefix := filepath.Join(dataDir, "projects", "42")
if w.Root != prefix {
t.Fatalf("expected Root=%s, got %s", prefix, w.Root)
}
if w.Source != filepath.Join(prefix, "source") {
t.Fatalf("Source not nested under Root")
}
if w.Prepared != filepath.Join(prefix, "prepared") {
t.Fatalf("Prepared not nested under Root")
}
if w.Build != filepath.Join(prefix, "build") {
t.Fatalf("Build not nested under Root")
}
}

func TestHardening_CleanPreparedWipesAll(t *testing.T) {
dataDir := t.TempDir()
w := NewWorkspace(dataDir, 7)
if err := w.Ensure(); err != nil {
t.Fatalf("Ensure: %v", err)
}
stale := filepath.Join(w.Prepared, "stale.txt")
if err := os.WriteFile(stale, []byte("junk"), 0644); err != nil {
t.Fatalf("seed: %v", err)
}
if err := w.CleanPrepared(); err != nil {
t.Fatalf("CleanPrepared: %v", err)
}
if _, err := os.Stat(stale); !os.IsNotExist(err) {
t.Fatalf("stale file still present")
}
}

func TestHardening_CleanPreparedPreservesDir(t *testing.T) {
dataDir := t.TempDir()
w := NewWorkspace(dataDir, 8)
if err := w.Ensure(); err != nil {
t.Fatalf("Ensure: %v", err)
}
if err := w.CleanPrepared(); err != nil {
t.Fatalf("CleanPrepared: %v", err)
}
info, err := os.Stat(w.Prepared)
if err != nil {
t.Fatalf("Prepared dir missing after clean: %v", err)
}
if !info.IsDir() {
t.Fatalf("Prepared is not a directory")
}
}

func TestHardening_SafeJoinRejectsEscape(t *testing.T) {
dataDir := t.TempDir()
w := NewWorkspace(dataDir, 9)
for _, bad := range []string{"../foo", "../../etc/passwd", "a/../../b"} {
if _, err := w.SafeJoin(bad); err == nil {
t.Fatalf("expected rejection for %q", bad)
}
}
}

func TestHardening_SafeJoinRejectsAbsolute(t *testing.T) {
dataDir := t.TempDir()
w := NewWorkspace(dataDir, 10)
for _, bad := range []string{"/etc/passwd", "/tmp/x", "/"} {
if _, err := w.SafeJoin(bad); err == nil {
t.Fatalf("expected rejection for %q", bad)
}
}
}

func TestHardening_SafeJoinAcceptsValid(t *testing.T) {
dataDir := t.TempDir()
w := NewWorkspace(dataDir, 11)
for _, good := range []string{"source/index.html", "prepared/app.js", "build/output.apk"} {
got, err := w.SafeJoin(good)
if err != nil {
t.Fatalf("SafeJoin(%q) failed: %v", good, err)
}
if got == "" {
t.Fatalf("SafeJoin(%q) returned empty", good)
}
}
}
