package generator

import (
    "context"
    "fmt"
    "local-apk-builder/internal/database"
)

// URLSourceAdapter handles URL-based projects (existing behavior)
type URLSourceAdapter struct {
    Project *database.Project
}

func (a *URLSourceAdapter) SourceType() string {
    return "url"
}

func (a *URLSourceAdapter) Validate(ctx context.Context) error {
    if a.Project == nil || a.Project.URL == "" {
        return fmt.Errorf("URL is required")
    }
    return nil
}

func (a *URLSourceAdapter) Prepare(ctx context.Context) (*PreparedSource, error) {
    return &PreparedSource{
        SourceType:   "url",
        WorkspaceDir: "",
        Profile: ProjectProfile{
            Type:      "static",
            Framework: "none",
            Builder:   "none",
            EntryDir:  "",
            OutputDir: "",
        },
        WebOutputDir: "",
        IndexHTML:    "",
        Metadata:     map[string]string{},
    }, nil
}

func (a *URLSourceAdapter) Cleanup() error {
    return nil
}
