package dashboard

import (
    "fmt"
    "net/http"
    "os"
    "path/filepath"
    "strconv"
    "strings"

    "local-apk-builder/internal/database"
    "local-apk-builder/internal/upload"
)

// UploadSourceHandler handles ZIP/folder upload
func (h *Handler) UploadSourceHandler(w http.ResponseWriter, r *http.Request) {
    // Parse project ID from URL
    path := strings.TrimPrefix(r.URL.Path, "/projects/")
    path = strings.TrimSuffix(path, "/source/upload")
    projectID, err := strconv.ParseInt(path, 10, 64)
    if err != nil {
        http.Error(w, "Invalid project ID", http.StatusBadRequest)
        return
    }

    project, err := database.GetProject(projectID)
    if err != nil || project == nil {
        http.Error(w, "Project not found", http.StatusNotFound)
        return
    }

    // Get config for data directory
    // Use the same data dir as the database
    dataDir := os.Getenv("KIN_DATA_DIR")
	if dataDir == "" {
		dataDir = filepath.Join(os.Getenv("HOME"), ".local-apk-builder")
	}
    ws := database.NewWorkspace(dataDir, projectID)
    if err := ws.Ensure(); err != nil {
        http.Error(w, "Failed to create workspace", http.StatusInternalServerError)
        return
    }

    // Clean source before new upload
    ws.CleanSource()

    // Parse multipart form (50MB limit)
    if err := r.ParseMultipartForm(50 << 20); err != nil {
        http.Error(w, "Upload too large or invalid", http.StatusBadRequest)
        return
    }

    // Check for ZIP file
    file, header, err := r.FormFile("file")
    if err == nil {
        defer file.Close()

        // Save ZIP to temp
        tmpZIP := filepath.Join(ws.Root, "upload.zip")
        dst, err := os.Create(tmpZIP)
        if err != nil {
            http.Error(w, "Failed to save upload", http.StatusInternalServerError)
            return
        }
        defer dst.Close()

        if _, err := dst.ReadFrom(file); err != nil {
            http.Error(w, "Failed to read upload", http.StatusInternalServerError)
            return
        }

        // Extract
        ext := upload.NewZipExtractor()
        if _, err := ext.Extract(tmpZIP, ws.Source); err != nil {
            ws.CleanSource()
            http.Error(w, "ZIP extraction failed: "+err.Error(), http.StatusBadRequest)
            return
        }

        os.Remove(tmpZIP)

        // Update project
        project.SourceType = "upload"
        project.SourcePath = ws.Source
        project.OriginalZipName = header.Filename
        if err := database.UpdateProject(project); err != nil {
            http.Error(w, "Failed to update project", http.StatusInternalServerError)
            return
        }
    } else {
        // Check for folder upload
        files := r.MultipartForm.File["files"]
        if len(files) == 0 {
            http.Error(w, "No file uploaded", http.StatusBadRequest)
            return
        }

        if _, err := upload.SaveFolderUpload(files, ws.Source); err != nil {
            ws.CleanSource()
            http.Error(w, "Folder upload failed: "+err.Error(), http.StatusBadRequest)
            return
        }

        // Update project
        project.SourceType = "upload"
        project.SourcePath = ws.Source
        project.OriginalZipName = "folder-upload"
        if err := database.UpdateProject(project); err != nil {
            http.Error(w, "Failed to update project", http.StatusInternalServerError)
            return
        }
    }

    // Return success
    w.Header().Set("Content-Type", "application/json")
    fmt.Fprintf(w, `{"project_id":%d,"source_type":"upload","status":"uploaded"}`, projectID)
}
