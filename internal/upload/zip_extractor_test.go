package upload

import (
    "archive/zip"
    "os"
    "path/filepath"
    "testing"
)

func createTestZIP(t *testing.T, entries map[string]string) string {
    t.Helper()
    tmp := t.TempDir()
    zipPath := filepath.Join(tmp, "test.zip")
    
    f, err := os.Create(zipPath)
    if err != nil {
        t.Fatal(err)
    }
    defer f.Close()
    
    zw := zip.NewWriter(f)
    for name, content := range entries {
        w, err := zw.Create(name)
        if err != nil {
            t.Fatal(err)
        }
        w.Write([]byte(content))
    }
    zw.Close()
    
    return zipPath
}

func TestZipExtractValid(t *testing.T) {
    zipPath := createTestZIP(t, map[string]string{
        "index.html":          "<html>Test</html>",
        "css/app.css":         "body{}",
        "js/app.js":           "console.log('hi')",
        "images/logo.png":     "PNGDATA",
        "fonts/font.woff2":    "WOFFDATA",
    })
    
    destDir := t.TempDir()
    ext := NewZipExtractor()
    result, err := ext.Extract(zipPath, destDir)
    if err != nil {
        t.Fatalf("Extract failed: %v", err)
    }
    
    if result.FileCount != 5 {
        t.Errorf("Expected 5 files, got %d", result.FileCount)
    }
    
    // Verify index.html exists
    if _, err := os.Stat(filepath.Join(destDir, "index.html")); err != nil {
        t.Error("index.html not extracted")
    }
}

func TestZipExtractPathTraversal(t *testing.T) {
    zipPath := createTestZIP(t, map[string]string{
        "../evil.txt": "evil",
    })
    
    ext := NewZipExtractor()
    _, err := ext.Extract(zipPath, t.TempDir())
    if err == nil {
        t.Error("Expected traversal error")
    }
}

func TestZipExtractAbsolutePath(t *testing.T) {
    zipPath := createTestZIP(t, map[string]string{
        "/etc/passwd": "root:x:0:0",
    })
    
    ext := NewZipExtractor()
    _, err := ext.Extract(zipPath, t.TempDir())
    if err == nil {
        t.Error("Expected absolute path error")
    }
}

func TestZipExtractNestedArchive(t *testing.T) {
    zipPath := createTestZIP(t, map[string]string{
        "inner.zip": "ZIPDATA",
    })
    
    ext := NewZipExtractor()
    _, err := ext.Extract(zipPath, t.TempDir())
    if err == nil {
        t.Error("Expected nested archive error")
    }
}
