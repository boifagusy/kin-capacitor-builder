package database

import "time"

type Project struct {
    ID               int64     `json:"id"`
    ProjectName      string    `json:"project_name"`
    URL              string    `json:"url"`
    AppName          string    `json:"app_name"`
    LogoPath         string    `json:"logo_path"`
    PrimaryColor     string    `json:"primary_color"`
    FeaturesConfig   string    `json:"features_config"`
    SplashConfig     string    `json:"splash_config"`
    OnboardingConfig string    `json:"onboarding_config"`
    BridgeConfig     string    `json:"bridge_config"`
    GoogleServices   string    `json:"google_services"`
    SourceType       string    `json:"source_type"`
    SourcePath       string    `json:"source_path"`
    ProjectType      string    `json:"project_type"`
    OriginalZipName  string    `json:"original_zip_name"`
    RuntimeType      string    `json:"runtime_type"`
    CurrentStep      int       `json:"current_step"`
    Status           string    `json:"status"`
    Version          string    `json:"version"`
    CreatedAt        time.Time `json:"created_at"`
    UpdatedAt        time.Time `json:"updated_at"`
}

type Build struct {
    ID                   int64      `json:"id"`
    BuildType            string     `json:"build_type"`
    RepoName             string     `json:"repo_name"`
    ProjectID            int64      `json:"project_id"`
    Version              string     `json:"version"`
    Status               string     `json:"status"`
    BuildProvider        string     `json:"build_provider"`
    ProviderRunID        string     `json:"provider_run_id"`
    ProviderArtifactName string     `json:"provider_artifact_name"`
    ArtifactPath         string     `json:"artifact_path"`
    ArtifactSize         int64      `json:"artifact_size"`
    ArtifactHash         string     `json:"artifact_hash"`
    ErrorMessage         string     `json:"error_message"`
    StartedAt            *time.Time `json:"started_at,omitempty"`
    CompletedAt          *time.Time `json:"completed_at,omitempty"`
    CreatedAt            time.Time  `json:"created_at"`
}

func NewProject() *Project {
    return &Project{
        URL:              "",
        AppName:          "",
        LogoPath:         "",
        PrimaryColor:     "#6366f1",
        FeaturesConfig:   "[]",
        SplashConfig:     `{"background_color":"#6366f1","show_logo":true,"loading_text":"Loading...","duration":1500}`,
        OnboardingConfig: `{"style":"bottom","template":"bottom-tabs","items":[{"icon":"fa-house","url":""}]}`,
        CurrentStep:      1,
        Status:           "draft",
        Version:          "1.0.0",
        SourceType:       "url",
        SourcePath:       "",
        ProjectType:      "static",
        OriginalZipName:  "",
        CreatedAt:        time.Now(),
        UpdatedAt:        time.Now(),
    }
}

func NewBuild(projectID int64) *Build {
    return &Build{
        ProjectID: projectID,
        Version:   "1.0.0",
        Status:    "queued",
        BuildType: "apk",
        CreatedAt: time.Now(),
    }
}
