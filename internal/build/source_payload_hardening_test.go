package build

import (
"os"
"path/filepath"
"strings"
"testing"
)


func TestHardening_CollectSourceFiles_SkipsHeavyDirs(t *testing.T) {
src := t.TempDir()
os.MkdirAll(filepath.Join(src, "node_modules", "lib"), 0755)
os.WriteFile(filepath.Join(src, "node_modules", "lib", "x.js"), []byte("x"), 0644)
os.MkdirAll(filepath.Join(src, ".git"), 0755)
os.WriteFile(filepath.Join(src, ".git", "HEAD"), []byte("ref"), 0644)
os.WriteFile(filepath.Join(src, "index.html"), []byte("<html></html>"), 0644)
got, err := CollectSourceFiles(src)
if err != nil {
t.Fatalf("CollectSourceFiles: %v", err)
}
if _, ok := got["index.html"]; !ok {
t.Fatalf("index.html missing")
}
for k := range got {
if strings.HasPrefix(k, "node_modules/") || strings.HasPrefix(k, ".git/") {
t.Fatalf("heavy dir file leaked: %s", k)
}
}
}

func TestHardening_CollectSourceFiles_SkipsLargeBinary(t *testing.T) {
src := t.TempDir()
big := make([]byte, 2*1024*1024)
os.WriteFile(filepath.Join(src, "huge.bin"), big, 0644)
os.WriteFile(filepath.Join(src, "small.txt"), []byte("ok"), 0644)
got, err := CollectSourceFiles(src)
if err != nil {
t.Fatalf("CollectSourceFiles: %v", err)
}
if _, ok := got["huge.bin"]; ok {
t.Fatalf("huge.bin should be skipped")
}
if _, ok := got["small.txt"]; !ok {
t.Fatalf("small.txt missing")
}
}

func TestHardening_CollectSourceFiles_PreservesStructure(t *testing.T) {
src := t.TempDir()
os.MkdirAll(filepath.Join(src, "assets", "js"), 0755)
os.WriteFile(filepath.Join(src, "assets", "js", "index-abc123.js"), []byte("x"), 0644)
os.WriteFile(filepath.Join(src, "assets", "logo.png"), []byte("PNG"), 0644)
os.WriteFile(filepath.Join(src, "index.html"), []byte("<html></html>"), 0644)
got, err := CollectSourceFiles(src)
if err != nil {
t.Fatalf("CollectSourceFiles: %v", err)
}
want := []string{"assets/js/index-abc123.js", "assets/logo.png", "index.html"}
for _, k := range want {
if _, ok := got[k]; !ok {
t.Fatalf("missing key %q", k)
}
}
}
