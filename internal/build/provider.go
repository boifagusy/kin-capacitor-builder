package build

import (
    "context"
    "local-apk-builder/internal/database"
)

// BuildConfig is a provider-agnostic snapshot of project configuration
type BuildConfig struct {
    ProjectID    int64             `json:"project_id"`
    BuildID      int64             `json:"build_id"`
    URL          string            `json:"url"`
    AppName      string            `json:"app_name"`
    LogoPath     string            `json:"logo_path"`
    PrimaryColor string            `json:"primary_color"`
    Features     []string          `json:"features"`
    Splash       map[string]interface{} `json:"splash"`
    Navigation   map[string]interface{} `json:"navigation"`
}

type BuildResult struct {
    ProviderName string
    Status       string
    ProviderRunID string
    ArtifactPath string
    ArtifactSize int64
    ArtifactHash string
}

type Provider interface {
    Build(ctx context.Context, build *database.Build, cfg BuildConfig) (*BuildResult, error)
    GetStatus(ctx context.Context, providerRunID string) (*BuildResult, error)
    DownloadArtifact(ctx context.Context, providerRunID string) ([]byte, error)
}
