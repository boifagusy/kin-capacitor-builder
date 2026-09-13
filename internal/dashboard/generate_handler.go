package dashboard

import (
    "encoding/json"
    "net/http"
    "strconv"
    "strings"
    
    "local-apk-builder/internal/generator"
)

// GenerateProjectHandler handles GET/POST /projects/{id}/generate
func (h *Handler) GenerateProjectHandler(w http.ResponseWriter, r *http.Request) {
    path := strings.TrimPrefix(r.URL.Path, "/projects/")
    path = strings.TrimSuffix(path, "/generate")
    id, err := strconv.ParseInt(path, 10, 64)
    if err != nil {
        http.NotFound(w, r)
        return
    }
    
    project, err := h.service.Get(id)
    if err != nil || project == nil {
        http.NotFound(w, r)
        return
    }
    
    generated, err := generator.GenerateForProject(project)
    if err != nil {
        http.Error(w, "Failed to generate", http.StatusInternalServerError)
        return
    }
    
    // Check if request wants JSON (API) or HTML (browser)
    if r.Header.Get("Accept") == "application/json" {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(generated)
        return
    }
    
    // Browser request - render HTML template
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    h.templates.ExecuteTemplate(w, "generate_result", generated)
}
