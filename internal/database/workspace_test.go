package database

import (
    "os"
    "path/filepath"
    "testing"
)

func TestWorkspaceEnsure(t *testing.T) {
    tmp := t.TempDir()
    ws := NewWorkspace(tmp, 1)
    
    if err := ws.Ensure(); err != nil {
        t.Fatalf("Ensure failed: %v", err)
    }
    
    for _, dir := range []string{ws.Source, ws.Prepared, ws.Build} {
        if _, err := os.Stat(dir); err != nil {
            t.Errorf("Directory not created: %s", dir)
        }
    }
}

func TestWorkspaceSafeJoin(t *testing.T) {
    tmp := t.TempDir()
    ws := NewWorkspace(tmp, 1)
    ws.Ensure()
    
    // Valid path
    path, err := ws.SafeJoin("css/app.css")
    if err != nil {
        t.Errorf("SafeJoin failed for valid path: %v", err)
    }
    expected := filepath.Join(ws.Root, "css", "app.css")
    if path != expected {
        t.Errorf("Expected %s, got %s", expected, path)
    }
    
    // Traversal
    _, err = ws.SafeJoin("../evil.txt")
    if err == nil {
        t.Error("SafeJoin should reject traversal")
    }
    
    // Absolute
    _, err = ws.SafeJoin("/etc/passwd")
    if err == nil {
        t.Error("SafeJoin should reject absolute path")
    }
}

func TestWorkspaceCleanSource(t *testing.T) {
    tmp := t.TempDir()
    ws := NewWorkspace(tmp, 1)
    ws.Ensure()
    
    // Create a file in source
    os.WriteFile(filepath.Join(ws.Source, "test.txt"), []byte("test"), 0644)
    
    // Clean
    ws.CleanSource()
    
    // Verify empty
    entries, _ := os.ReadDir(ws.Source)
    if len(entries) != 0 {
        t.Errorf("Source not cleaned: %d entries", len(entries))
    }
}
