package generator

import (
    "context"
    "os"
    "path/filepath"
    "testing"
    "local-apk-builder/internal/database"
)

func TestStaticSourceAdapter(t *testing.T) {
    // Create test workspace with static HTML
    tmp := t.TempDir()
    sourceDir := filepath.Join(tmp, "source")
    os.MkdirAll(sourceDir, 0755)
    os.WriteFile(filepath.Join(sourceDir, "index.html"), []byte("<html>Test</html>"), 0644)
    os.WriteFile(filepath.Join(sourceDir, "style.css"), []byte("body{}"), 0644)
    
    ws := &database.Workspace{
        Source:   sourceDir,
        Prepared: filepath.Join(tmp, "prepared"),
        Build:    filepath.Join(tmp, "build"),
    }
    ws.Ensure()
    
    project := &database.Project{}
    adapter := &UploadSourceAdapter{Project: project, Workspace: ws}
    
    prepared, err := adapter.Prepare(context.Background())
    if err != nil {
        t.Fatalf("Prepare failed: %v", err)
    }
    
    if prepared.SourceType != "upload" {
        t.Errorf("Expected upload, got %s", prepared.SourceType)
    }
    if prepared.Profile.Type != "static" {
        t.Errorf("Expected static, got %s", prepared.Profile.Type)
    }
    
    // Verify index.html copied
    preparedIndex := filepath.Join(ws.Prepared, "index.html")
    if _, err := os.Stat(preparedIndex); err != nil {
        t.Error("index.html not prepared")
    }
}

func TestURLSourceAdapter(t *testing.T) {
    project := &database.Project{URL: "https://example.com"}
    adapter := &URLSourceAdapter{Project: project}
    
    prepared, err := adapter.Prepare(context.Background())
    if err != nil {
        t.Fatalf("Prepare failed: %v", err)
    }
    
    if prepared.SourceType != "url" {
        t.Errorf("Expected url, got %s", prepared.SourceType)
    }
}
