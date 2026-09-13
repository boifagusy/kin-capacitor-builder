package dashboard

import (
    "fmt"
    "net/http"
    "os"
    "strconv"
    "strings"

    "local-apk-builder/internal/database"
)

// DownloadBuildHandler serves the APK artifact for a build
func (h *Handler) DownloadBuildHandler(w http.ResponseWriter, r *http.Request) {
    path := strings.TrimPrefix(r.URL.Path, "/projects/")
    path = strings.TrimSuffix(path, "/download")
    buildID, err := strconv.ParseInt(path, 10, 64)
    if err != nil {
        http.NotFound(w, r)
        return
    }

    build, err := database.GetBuild(buildID)
    if err != nil || build == nil {
        http.NotFound(w, r)
        return
    }

    if build.Status != "success" || build.ArtifactPath == "" {
        http.Error(w, "APK not ready", http.StatusBadRequest)
        return
    }

    // Check if file exists
    if _, err := os.Stat(build.ArtifactPath); os.IsNotExist(err) {
        http.Error(w, "APK file not found", http.StatusNotFound)
        return
    }

    if build.BuildType == "aab" {
        w.Header().Set("Content-Type", "application/octet-stream")
        w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"app-%d.aab\"", build.ID))
    } else {
        w.Header().Set("Content-Type", "application/vnd.android.package-archive")
        w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"app-%d.apk\"", build.ID))
    }
    http.ServeFile(w, r, build.ArtifactPath)
}
