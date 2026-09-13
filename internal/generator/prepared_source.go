package generator

import "context"

// PreparedSource represents a prepared web source ready for Capacitor
type PreparedSource struct {
    SourceType    string                 // "url" | "upload"
    WorkspaceDir  string                 // absolute server-side path
    Profile       ProjectProfile         // detection result
    WebOutputDir  string                 // relative path within workspace
    IndexHTML     string                 // relative path to index.html
    Metadata      map[string]string
}

// SourceAdapter converts a source input into a PreparedSource
type SourceAdapter interface {
    SourceType() string
    Prepare(ctx context.Context) (*PreparedSource, error)
    Validate(ctx context.Context) error
    Cleanup() error
}
