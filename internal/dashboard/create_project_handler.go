package dashboard

import (
    "encoding/json"
    "fmt"
    "net/http"

    "local-apk-builder/internal/database"
)

// CreateProjectHandler handles POST /api/projects
func (h *Handler) CreateProjectHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var input struct {
        ProjectName string `json:"project_name"`
        SourceType  string `json:"source_type"`
    }
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    p := database.NewProject()
    p.ProjectName = input.ProjectName
    p.AppName = input.ProjectName
    if input.SourceType == "upload" {
        p.SourceType = "upload"
    }

    if err := database.SaveProject(p); err != nil {
        http.Error(w, "Failed to create project: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    fmt.Fprintf(w, `{"id":%d,"project_name":"%s"}`, p.ID, p.ProjectName)
}
