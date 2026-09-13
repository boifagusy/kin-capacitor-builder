package database

import (
    "fmt"
    "os"
    "path/filepath"
)

// Workspace manages per-project file isolation
type Workspace struct {
    Root     string
    Source   string
    Prepared string
    Build    string
}

// NewWorkspace creates a workspace for a project
func NewWorkspace(dataDir string, projectID int64) *Workspace {
    root := filepath.Join(dataDir, "projects", fmt.Sprintf("%d", projectID))
    return &Workspace{
        Root:     root,
        Source:   filepath.Join(root, "source"),
        Prepared: filepath.Join(root, "prepared"),
        Build:    filepath.Join(root, "build"),
    }
}

// Ensure creates all workspace directories
func (w *Workspace) Ensure() error {
    for _, dir := range []string{w.Source, w.Prepared, w.Build} {
        if err := os.MkdirAll(dir, 0755); err != nil {
            return fmt.Errorf("workspace: failed to create %s: %w", dir, err)
        }
    }
    return nil
}

// CleanSource removes all files in source directory (keeps directory)
func (w *Workspace) CleanSource() error {
    if err := os.RemoveAll(w.Source); err != nil {
        return fmt.Errorf("workspace: failed to clean source: %w", err)
    }
    return os.MkdirAll(w.Source, 0755)
}

// CleanPrepared removes all files in prepared directory
func (w *Workspace) CleanPrepared() error {
    if err := os.RemoveAll(w.Prepared); err != nil {
        return fmt.Errorf("workspace: failed to clean prepared: %w", err)
    }
    return os.MkdirAll(w.Prepared, 0755)
}

// Cleanup removes the entire workspace
func (w *Workspace) Cleanup() error {
    return os.RemoveAll(w.Root)
}

// SafeJoin joins a relative path to the workspace root and verifies containment
func (w *Workspace) SafeJoin(relPath string) (string, error) {
    // Normalize separators
    relPath = filepath.FromSlash(relPath)
    
    // Clean the path
    cleaned := filepath.Clean(relPath)
    
    // Reject absolute paths
    if filepath.IsAbs(cleaned) {
        return "", fmt.Errorf("absolute path not allowed: %s", relPath)
    }
    
    // Join with root
    full := filepath.Join(w.Root, cleaned)
    
    // Verify containment using filepath.Rel
    rel, err := filepath.Rel(w.Root, full)
    if err != nil {
        return "", err
    }
    if rel == ".." || (len(rel) >= 3 && rel[:3] == ".."+string(filepath.Separator)) {
        return "", fmt.Errorf("path escapes workspace: %s", relPath)
    }
    
    return full, nil
}
