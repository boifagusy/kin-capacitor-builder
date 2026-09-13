package build

import (
    "context"
    "fmt"
    "log"
    "os"
    "os/exec"
    "time"

    "local-apk-builder/internal/database"
)

// Status constants
const (
    StatusQueued    = "queued"
    StatusSubmitted = "submitted"
    StatusBuilding  = "building"
    StatusSuccess   = "success"
    StatusFailed    = "failed"
)

// Manager orchestrates builds
type Manager struct {
    provider Provider
}

// NewManager creates a build manager
func NewManager(provider Provider) *Manager {
    return &Manager{provider: provider}
}

// CreateBuild creates a queued build record and returns it
func (m *Manager) CreateBuild(projectID int64, version string) (*database.Build, error) {
    project, err := database.GetProject(projectID)
    if err != nil {
        return nil, err
    }
    if project == nil {
        return nil, fmt.Errorf("project not found")
    }
    if version == "" {
        version = project.Version
    }
    build := database.NewBuild(projectID)
    build.Version = version
    build.Status = StatusQueued
    if err := database.InsertBuild(build); err != nil {
        return nil, err
    }
    return build, nil
}

// EnqueueBuild creates a queued build with specified type
func (m *Manager) EnqueueBuild(projectID int64, version string, buildType string) (*database.Build, error) {
    project, err := database.GetProject(projectID)
    if err != nil {
        return nil, err
    }
    if project == nil {
        return nil, fmt.Errorf("project not found")
    }
    if version == "" {
        version = project.Version
    }
    build := database.NewBuild(projectID)
    build.Version = version
    build.Status = StatusQueued
    if buildType != "" {
        build.BuildType = buildType
    }
    log.Printf("DEBUG EnqueueBuild: buildType=%s, build.BuildType=%s", buildType, build.BuildType)
    if err := database.InsertBuild(build); err != nil {
        return nil, err
    }
    log.Printf("DEBUG EnqueueBuild: inserted build ID=%d with BuildType=%s", build.ID, build.BuildType)
    return build, nil
}

// ProcessBuild submits the build to provider and updates status
func (m *Manager) ProcessBuild(buildID int64) error {
    build, err := database.GetBuild(buildID)
    if err != nil || build == nil {
        return fmt.Errorf("build not found")
    }

    // Mark as submitted
    build.Status = StatusSubmitted
    build.StartedAt = timePtr(time.Now())
    database.UpdateBuild(build)

    // Load project and create config
    project, err := database.GetProject(build.ProjectID)
    if err != nil || project == nil {
        log.Printf("Build %d failed: project not found (ID: %d)", build.ID, build.ProjectID)
        build.Status = StatusFailed
        build.ErrorMessage = "project not found"
        build.CompletedAt = timePtr(time.Now())
        database.UpdateBuild(build)
        return fmt.Errorf("project not found")
    }
    cfg := BuildConfigFromProject(build, project)

    // Submit to provider
    result, err := m.provider.Build(context.Background(), build, cfg)
    if err != nil {
        log.Printf("Build %d failed: %v", build.ID, err)
        build.Status = StatusFailed
        build.ErrorMessage = err.Error()
        build.CompletedAt = timePtr(time.Now())
        database.UpdateBuild(build)
        return err
    }

    // Mark as building
    build.Status = StatusBuilding
    build.BuildProvider = result.ProviderName
    if result.ProviderRunID != "" {
        build.ProviderRunID = result.ProviderRunID
    }
    database.UpdateBuild(build)

    // Start background poller
    m.StartPoller(build.ID)

    // If provider gives immediate success
    if result.Status == StatusSuccess {
        build.Status = StatusSuccess
        build.ArtifactPath = result.ArtifactPath
        build.ArtifactSize = result.ArtifactSize
        build.ArtifactHash = result.ArtifactHash
        build.CompletedAt = timePtr(time.Now())
        database.UpdateBuild(build)
    }

    return nil
}

// GetBuild returns a build by ID
func (m *Manager) GetBuild(buildID int64) (*database.Build, error) {
    return database.GetBuild(buildID)
}

// ListProjectBuilds returns all builds for a project
func (m *Manager) ListProjectBuilds(projectID int64) ([]*database.Build, error) {
    return database.ListBuildsForProject(projectID)
}

// UpdateBuildStatus updates the status of a build
func (m *Manager) UpdateBuildStatus(buildID int64, status, artifactPath, errorMsg string) error {
    build, err := database.GetBuild(buildID)
    if err != nil || build == nil {
        return fmt.Errorf("build not found")
    }
    build.Status = status
    if artifactPath != "" {
        build.ArtifactPath = artifactPath
    }
    if errorMsg != "" {
        build.ErrorMessage = errorMsg
    }
    if status == StatusSuccess || status == StatusFailed {
        build.CompletedAt = timePtr(time.Now())
    }
    return database.UpdateBuild(build)
}

func timePtr(t time.Time) *time.Time {
    return &t
}



// SignAPKLocally signs an APK with v2 signature using local apksigner
func (m *Manager) SignAPKLocally(apkPath string) (string, int64, string, error) {
	if os.Getenv("KIN_DISABLE_LOCAL_SIGNING") == "1" {
		log.Println("SignAPKLocally: skipped (KIN_DISABLE_LOCAL_SIGNING=1)")
		return apkPath, 0, "", nil
	}
    // Find apksigner
    apksigner := "/data/data/com.termux/files/usr/lib/android-sdk/build-tools/34.0.0/apksigner"
    if _, err := os.Stat(apksigner); err != nil {
        return apkPath, 0, "", nil // Skip if apksigner not available
    }
    
    // Ensure debug keystore exists
    keystore := "debug.keystore"
    if _, err := os.Stat(keystore); err != nil {
        cmd := exec.Command("keytool", "-genkeypair", "-v",
            "-keystore", keystore,
            "-alias", "debug",
            "-keyalg", "RSA",
            "-keysize", "2048",
            "-validity", "10000",
            "-storepass", "android",
            "-keypass", "android",
            "-dname", "CN=Debug, OU=Debug, O=Debug, L=Debug, S=Debug, C=US")
        cmd.Run()
    }
    
    // Sign with apksigner
    signedPath := apkPath + ".signed"
    cmd := exec.Command(apksigner, "sign",
        "--ks", keystore,
        "--ks-key-alias", "debug",
        "--ks-pass", "pass:android",
        "--key-pass", "pass:android",
        "--out", signedPath,
        apkPath)
    output, err := cmd.CombinedOutput()
    if err != nil {
        return apkPath, 0, "", nil // Return original if signing fails
    }
    _ = output
    
    // Replace original with signed
    os.Rename(signedPath, apkPath)
    
    // Get file info
    info, err := os.Stat(apkPath)
    if err != nil {
        return apkPath, 0, "", nil
    }
    
    return apkPath, info.Size(), "", nil
}
