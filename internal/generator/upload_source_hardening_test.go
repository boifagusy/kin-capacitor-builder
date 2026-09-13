package generator

import (
"context"
"os"
"path/filepath"
"strings"
"testing"

"local-apk-builder/internal/database"
)


func makeAdapter(t *testing.T, files map[string]string) *UploadSourceAdapter {
t.Helper()
root := t.TempDir()
ws := &database.Workspace{
Root:     root,
Source:   filepath.Join(root, "source"),
Prepared: filepath.Join(root, "prepared"),
Build:    filepath.Join(root, "build"),
}
os.MkdirAll(ws.Source, 0755)
os.MkdirAll(ws.Prepared, 0755)
for rel, content := range files {
full := filepath.Join(ws.Source, filepath.FromSlash(rel))
os.MkdirAll(filepath.Dir(full), 0755)
if err := os.WriteFile(full, []byte(content), 0644); err != nil {
t.Fatalf("seed %s: %v", rel, err)
}
}
return &UploadSourceAdapter{Project: &database.Project{SourceType: "upload"}, Workspace: ws}
}

func TestHardening_PrepareRejectsLaravel(t *testing.T) {
a := makeAdapter(t, map[string]string{
"artisan":       "#!/usr/bin/env php",
"composer.json": "{}",
})
_, err := a.Prepare(context.Background())
if err == nil {
t.Fatal("expected error for laravel")
}
if !strings.Contains(err.Error(), "LARAVEL") {
t.Fatalf("expected LARAVEL error, got %v", err)
}
}

func TestHardening_PrepareRejectsUnknown(t *testing.T) {
a := makeAdapter(t, map[string]string{
"README.md": "# nothing",
})
_, err := a.Prepare(context.Background())
if err == nil {
t.Fatal("expected error for unknown project")
}
if !strings.Contains(err.Error(), "UNKNOWN") {
t.Fatalf("expected UNKNOWN error, got %v", err)
}
}

func TestHardening_PrepareCleansPrepared(t *testing.T) {
a := makeAdapter(t, map[string]string{"index.html": "<html></html>"})
stale := filepath.Join(a.Workspace.Prepared, "stale.txt")
os.WriteFile(stale, []byte("junk"), 0644)
if _, err := a.Prepare(context.Background()); err != nil {
t.Fatalf("Prepare: %v", err)
}
if _, err := os.Stat(stale); !os.IsNotExist(err) {
t.Fatalf("stale file survived Prepare")
}
}

func TestHardening_PrepareInjectsHooks(t *testing.T) {
a := makeAdapter(t, map[string]string{"index.html": "<html><head></head><body></body></html>"})
if _, err := a.Prepare(context.Background()); err != nil {
t.Fatalf("Prepare: %v", err)
}
data, err := os.ReadFile(filepath.Join(a.Workspace.Prepared, "index.html"))
if err != nil {
t.Fatalf("read prepared/index.html: %v", err)
}
if !strings.Contains(string(data), "kin-safe-area") {
t.Fatalf("injected safe-area marker missing")
}
if !strings.Contains(string(data), "app_bridge.js") {
t.Fatalf("app_bridge.js hook missing")
}
}
