package generator

import (
    "context"
    "fmt"
    "os"
    "path/filepath"

    "local-apk-builder/internal/database"
)

// UploadSourceAdapter handles ZIP/folder uploads
type UploadSourceAdapter struct {
    Project   *database.Project
    Workspace *database.Workspace
}

func (a *UploadSourceAdapter) SourceType() string {
    return "upload"
}

func (a *UploadSourceAdapter) Validate(ctx context.Context) error {
    if _, err := os.Stat(a.Workspace.Source); err != nil {
        return fmt.Errorf("source directory missing: %w", err)
    }
    return nil
}

func (a *UploadSourceAdapter) Prepare(ctx context.Context) (*PreparedSource, error) {
    detector := &ProjectDetector{SourceDir: a.Workspace.Source}
    profile, err := detector.Detect()
    if err != nil {
        return nil, err
    }

    if profile.Type == "laravel" {
        return nil, fmt.Errorf("%w: Laravel backend detected. Mobile app requires reachable API", ErrLaravelDetected)
    }
    if profile.Type == "unknown" {
        return nil, fmt.Errorf("%w: cannot determine project type", ErrUnknownProject)
    }

    if profile.Type == "static" {
        if err := a.Workspace.CleanPrepared(); err != nil {
            return nil, err
        }
        if err := copyDir(detector.SourceDir, a.Workspace.Prepared); err != nil {
            return nil, fmt.Errorf("failed to prepare static source: %w", err)
        }

        // Inject runtime hooks into prepared/index.html
        indexPath := filepath.Join(a.Workspace.Prepared, "index.html")
        data, err := os.ReadFile(indexPath)
        if err != nil {
            return nil, fmt.Errorf("%w: prepared/index.html not found: %v", ErrMissingIndex, err)
        }
        injected := InjectRuntimeHooks(string(data))
        if err := os.WriteFile(indexPath, []byte(injected), 0644); err != nil {
            return nil, fmt.Errorf("failed to write injected index.html: %w", err)
        }
    }

    return &PreparedSource{
        SourceType:   "upload",
        WorkspaceDir: a.Workspace.Root,
        Profile:      *profile,
        WebOutputDir: "prepared",
        IndexHTML:    filepath.Join("prepared", "index.html"),
        Metadata:     map[string]string{},
    }, nil
}

func (a *UploadSourceAdapter) Cleanup() error {
    return nil
}

func copyDir(src, dst string) error {
    return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        rel, err := filepath.Rel(src, path)
        if err != nil {
            return err
        }
        target := filepath.Join(dst, rel)
        if info.IsDir() {
            return os.MkdirAll(target, 0755)
        }
        data, err := os.ReadFile(path)
        if err != nil {
            return err
        }
        return os.WriteFile(target, data, 0644)
    })
}
