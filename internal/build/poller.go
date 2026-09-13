package build

import (
    "encoding/hex"
    "crypto/sha256"
    "context"
    "fmt"
    "log"
    "os"
    "time"

    "local-apk-builder/internal/database"
)

// StartPoller starts a background poller for a build
func (m *Manager) StartPoller(buildID int64) {
    go func() {
        ticker := time.NewTicker(10 * time.Second)
        defer ticker.Stop()

        for {
            select {
            case <-ticker.C:
                build, err := database.GetBuild(buildID)
                if err != nil || build == nil {
                    return
                }

                // Only poll if submitted or building
                if build.Status != StatusSubmitted && build.Status != StatusBuilding {
                    return
                }

                // If no provider run ID yet, try to find it
                if build.ProviderRunID == "" {
                    // Can't poll without run ID; wait for next tick
                    continue
                }

                result, err := m.provider.GetStatus(context.Background(), build.ProviderRunID)
                if err != nil {
                    build.Status = StatusFailed
                    build.ErrorMessage = err.Error()
                    build.CompletedAt = timePtr(time.Now())
                    database.UpdateBuild(build)
                    return
                }

                if result.Status == StatusSuccess {
                    // Download artifact with retry
                    var artifactData []byte
                    var err error
                    var apkPath string
                    var hash string

                    for retry := 0; retry < 5; retry++ {
                        artifactData, err = m.provider.DownloadArtifact(context.Background(), build.ProviderRunID)
                        if err == nil {
                            break
                        }
                        log.Printf("DEBUG: Artifact download retry %d failed: %v", retry+1, err)
                        time.Sleep(time.Duration(10*(retry+1)) * time.Second)
                    }
                    if err != nil {
                        build.Status = StatusFailed
                        build.ErrorMessage = "artifact download failed: " + err.Error()
                        build.CompletedAt = timePtr(time.Now())
                        database.UpdateBuild(build)
                        return
                    }

                    // Extract APK if provider is GitHub
                    if gh, ok := m.provider.(*GitHubActionsProvider); ok {
                        apkData, size, extractedHash, err := gh.ExtractAPK(artifactData)
                        hash = extractedHash
                        if err != nil {
                            build.Status = StatusFailed
                            build.ErrorMessage = "APK extraction failed: " + err.Error()
                            build.CompletedAt = timePtr(time.Now())
                            database.UpdateBuild(build)
                            return
                        }
                        // Save APK to disk
                        apkDir := "builds"
                        os.MkdirAll(apkDir, 0755)
                        ext := ".apk"
                        if build.BuildType == "aab" {
                            ext = ".aab"
                        }
                        apkPath = fmt.Sprintf("%s/build-%d%s", apkDir, build.ID, ext)
                        if err := os.WriteFile(apkPath, []byte(apkData), 0644); err != nil {
                            build.Status = StatusFailed
                            build.ErrorMessage = "failed to save APK: " + err.Error()
                            build.CompletedAt = timePtr(time.Now())
                            database.UpdateBuild(build)
                            return
                        }
                        build.ArtifactPath = apkPath
                        build.ArtifactSize = size
                        build.ArtifactHash = hash
                    } else {
                        apkDir := "builds"
                        os.MkdirAll(apkDir, 0755)
                        ext := ".apk"
                        if build.BuildType == "aab" {
                            ext = ".aab"
                        }
                        apkPath = fmt.Sprintf("%s/build-%d%s", apkDir, build.ID, ext)
                        os.WriteFile(apkPath, artifactData, 0644)
                        build.ArtifactPath = apkPath
                        build.ArtifactSize = int64(len(artifactData))
                    }

                    // Verify artifact integrity
                    savedData, err := os.ReadFile(apkPath)
                    if err == nil {
                        savedHash := sha256.Sum256(savedData)
                        if hex.EncodeToString(savedHash[:]) != hash {
                            log.Printf("Build %d: hash mismatch! expected %s, got %s", build.ID, hash, hex.EncodeToString(savedHash[:]))
                            build.Status = StatusFailed
                            build.ErrorMessage = "artifact hash verification failed"
                            build.CompletedAt = timePtr(time.Now())
                            database.UpdateBuild(build)
                            return
                        }
                        log.Printf("Build %d: artifact hash verified", build.ID)
                    }

                    build.Status = StatusSuccess
                    build.CompletedAt = timePtr(time.Now())
                    database.UpdateBuild(build)
                    log.Printf("Build %d succeeded", build.ID)
                    return
                }

                if result.Status == StatusFailed {
                    build.Status = StatusFailed
                    build.ErrorMessage = "GitHub workflow failed"
                    build.CompletedAt = timePtr(time.Now())
                    database.UpdateBuild(build)
                    return
                }

                // Update status
                build.Status = result.Status
                database.UpdateBuild(build)
            }
        }
    }()
}
