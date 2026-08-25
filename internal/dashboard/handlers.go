package dashboard

import (
    "encoding/json"
    "html/template"
    "log"
    "net/http"
    "strconv"
    "strings"
    
    "local-apk-builder/internal/database"
    "local-apk-builder/internal/project"
)

type Handler struct {
    templates *template.Template
    service   *project.Service
}

func NewHandler(templates *template.Template) *Handler {
    return &Handler{
        templates: templates,
        service:   project.NewService(),
    }
}

func (h *Handler) DashboardHandler(w http.ResponseWriter, r *http.Request) {
    projects, err := h.service.List()
    if err != nil {
        log.Printf("Error listing projects: %v", err)
        http.Error(w, "Failed to load projects", http.StatusInternalServerError)
        return
    }
    
    data := map[string]interface{}{
        "Projects": projects,
        "Count":    len(projects),
    }
    
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    if err := h.templates.ExecuteTemplate(w, "dashboard", data); err != nil {
        log.Printf("Template error: %v", err)
        http.Error(w, "Failed to render dashboard", http.StatusInternalServerError)
    }
}

func (h *Handler) ProjectDetailHandler(w http.ResponseWriter, r *http.Request) {
    path := strings.TrimPrefix(r.URL.Path, "/projects/")
    id, err := strconv.ParseInt(path, 10, 64)
    if err != nil {
        http.NotFound(w, r)
        return
    }
    
    p, err := h.service.Get(id)
    if err != nil {
        log.Printf("Error loading project: %v", err)
        http.Error(w, "Failed to load project", http.StatusInternalServerError)
        return
    }
    
    if p == nil {
        http.NotFound(w, r)
        return
    }
    
    data := map[string]interface{}{
        "Project": p,
    }
    
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    if err := h.templates.ExecuteTemplate(w, "project_detail", data); err != nil {
        log.Printf("Template error: %v", err)
        http.Error(w, "Failed to render project", http.StatusInternalServerError)
    }
}

func (h *Handler) DeleteProjectHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    path := strings.TrimPrefix(r.URL.Path, "/projects/")
    path = strings.TrimSuffix(path, "/delete")
    id, err := strconv.ParseInt(path, 10, 64)
    if err != nil {
        http.NotFound(w, r)
        return
    }
    
    if err := h.service.Delete(id); err != nil {
        log.Printf("Error deleting project: %v", err)
        http.Error(w, "Failed to delete project", http.StatusInternalServerError)
        return
    }
    
    http.Redirect(w, r, "/dashboard", http.StatusFound)
}

func (h *Handler) APIListProjects(w http.ResponseWriter, r *http.Request) {
    projects, err := h.service.List()
    if err != nil {
        http.Error(w, "Failed to list projects", http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(projects)
}

func (h *Handler) APIGetProject(w http.ResponseWriter, r *http.Request) {
    path := strings.TrimPrefix(r.URL.Path, "/api/projects/")
    id, err := strconv.ParseInt(path, 10, 64)
    if err != nil {
        http.NotFound(w, r)
        return
    }
    
    p, err := h.service.Get(id)
    if err != nil {
        http.Error(w, "Failed to load project", http.StatusInternalServerError)
        return
    }
    
    if p == nil {
        http.NotFound(w, r)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(p)
}

var _ = database.Build{}
