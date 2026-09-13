package build

import (
    "context"
    "fmt"
    "local-apk-builder/internal/database"
)

type LocalProvider struct{}

func NewLocalProvider() *LocalProvider { return &LocalProvider{} }

func (p *LocalProvider) Build(ctx context.Context, build *database.Build, cfg BuildConfig) (*BuildResult, error) {
    return nil, fmt.Errorf("local build not available — connect GitHub to build APK/AAB")
}

func (p *LocalProvider) GetStatus(ctx context.Context, providerRunID string) (*BuildResult, error) {
    return nil, fmt.Errorf("local build not available")
}

func (p *LocalProvider) DownloadArtifact(ctx context.Context, providerRunID string) ([]byte, error) {
    return nil, fmt.Errorf("not implemented")
}
