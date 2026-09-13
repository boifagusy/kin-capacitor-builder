package dashboard

import (
    "encoding/json"
    "net/http"
    "strconv"
    "strings"

    "local-apk-builder/internal/database"
)

// UploadGoogleServicesHandler handles POST /api/projects/{id}/google-services
func (h *Handler) UploadGoogleServicesHandler(w http.ResponseWriter, r *http.Request) {
    path := strings.TrimPrefix(r.URL.Path, "/api/projects/")
    path = strings.TrimSuffix(path, "/google-services")
    projectID, err := strconv.ParseInt(path, 10, 64)
    if err != nil {
        http.NotFound(w, r)
        return
    }

    project, err := h.service.Get(projectID)
    if err != nil || project == nil {
        http.NotFound(w, r)
        return
    }

    var payload struct {
        GoogleServicesJSON string `json:"google_services_json"`
    }
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    project.GoogleServices = payload.GoogleServicesJSON
    if err := database.UpdateProject(project); err != nil {
        http.Error(w, "Failed to save", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"status": "saved"})
}
