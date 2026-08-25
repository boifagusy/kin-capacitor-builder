package database

import "time"

// Project represents an app project
type Project struct {
    ID               int64     `json:"id"`
    ProjectName      string    `json:"project_name"`
    URL              string    `json:"url"`
    AppName          string    `json:"app_name"`
    LogoPath         string    `json:"logo_path"`
    PrimaryColor     string    `json:"primary_color"`
    OnboardingConfig string    `json:"onboarding_config"`
    CurrentStep      int       `json:"current_step"`
    Status           string    `json:"status"`
    Version          string    `json:"version"`
    CreatedAt        time.Time `json:"created_at"`
    UpdatedAt        time.Time `json:"updated_at"`
}

// Build represents a build record
type Build struct {
    ID            int64      `json:"id"`
    ProjectID     int64      `json:"project_id"`
    Version       string     `json:"version"`
    Status        string     `json:"status"`
    BuildProvider string     `json:"build_provider"`
    ArtifactPath  string     `json:"artifact_path"`
    ArtifactSize  int64      `json:"artifact_size"`
    ArtifactHash  string     `json:"artifact_hash"`
    ErrorMessage  string     `json:"error_message"`
    StartedAt     *time.Time `json:"started_at,omitempty"`
    CompletedAt   *time.Time `json:"completed_at,omitempty"`
    CreatedAt     time.Time  `json:"created_at"`
}

// NewProject creates a default project
func NewProject() *Project {
    return &Project{
        ProjectName:      "",
        URL:              "",
        AppName:          "",
        LogoPath:         "",
        PrimaryColor:     "#6366f1",
        OnboardingConfig: `{"has_splash_screen":true,"splash_color":"#FFFFFF","orientation":"portrait"}`,
        CurrentStep:      1,
        Status:           "draft",
        Version:          "1.0.0",
        CreatedAt:        time.Now(),
        UpdatedAt:        time.Now(),
    }
}

// NewBuild creates a new build record
func NewBuild(projectID int64) *Build {
    return &Build{
        ProjectID:     projectID,
        Version:       "1.0.0",
        Status:        "queued",
        BuildProvider: "",
        CreatedAt:     time.Now(),
    }
}
