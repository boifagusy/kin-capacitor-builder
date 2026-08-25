package server

import (
    "encoding/json"
    "html/template"
    "log"
    "net/http"
    "strconv"
    
    "local-apk-builder/internal/database"
)

type Handler struct {
    templates *template.Template
}

func NewHandler(templates *template.Template) *Handler {
    return &Handler{templates: templates}
}

func getProjectID(r *http.Request) int64 {
    idStr := r.URL.Query().Get("project_id")
    if idStr == "" {
        return 0
    }
    id, _ := strconv.ParseInt(idStr, 10, 64)
    return id
}

// IndexHandler redirects to dashboard
func (h *Handler) IndexHandler(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path != "/" {
        http.NotFound(w, r)
        return
    }
    http.Redirect(w, r, "/dashboard", http.StatusFound)
}

// Screen1Handler handles GET /wizard/1
// If project_id provided, loads that project. Otherwise shows empty form.
func (h *Handler) Screen1Handler(w http.ResponseWriter, r *http.Request) {
    projectID := getProjectID(r)
    
    var project *database.Project
    
    if projectID > 0 {
        var err error
        project, err = database.GetProject(projectID)
        if err != nil {
            http.Error(w, "Failed to load project", http.StatusInternalServerError)
            return
        }
        if project == nil {
            http.NotFound(w, r)
            return
        }
    } else {
        // New project - don't save to DB yet
        project = database.NewProject()
        project.ID = 0
    }
    
    data := map[string]interface{}{
        "Project":   project,
        "Step":      1,
        "ProjectID": project.ID,
    }
    
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    if err := h.templates.ExecuteTemplate(w, "root", data); err != nil {
        log.Printf("Template error: %v", err)
        http.Error(w, "Failed to render template", http.StatusInternalServerError)
    }
}

// Screen1PostHandler handles POST /wizard/1
func (h *Handler) Screen1PostHandler(w http.ResponseWriter, r *http.Request) {
    if err := r.ParseForm(); err != nil {
        http.Error(w, "Failed to parse form", http.StatusBadRequest)
        return
    }
    
    projectID := getProjectID(r)
    url := r.FormValue("url")
    
    if err := ValidateURL(url); err != nil {
        w.Header().Set("Content-Type", "text/html")
        w.WriteHeader(http.StatusBadRequest)
        h.templates.ExecuteTemplate(w, "url_error", map[string]string{"Error": err.Error()})
        return
    }
    
    normalizedURL := NormalizeURL(url)
    
    var project *database.Project
    var err error
    
    if projectID > 0 {
        project, err = database.GetProject(projectID)
        if err != nil {
            http.Error(w, "Failed to load project", http.StatusInternalServerError)
            return
        }
    } else {
        // Create new project on first form submission
        project = database.NewProject()
        err = database.InsertProject(project)
        if err != nil {
            http.Error(w, "Failed to create project", http.StatusInternalServerError)
            return
        }
        projectID = project.ID
    }
    
    project.URL = normalizedURL
    project.CurrentStep = 2
    
    if err := database.UpdateProject(project); err != nil {
        http.Error(w, "Failed to save project", http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("HX-Redirect", "/wizard/2?project_id="+strconv.FormatInt(projectID, 10))
    w.WriteHeader(http.StatusOK)
}

// ValidateURLHandler handles POST /api/validate-url
func (h *Handler) ValidateURLHandler(w http.ResponseWriter, r *http.Request) {
    if err := r.ParseForm(); err != nil {
        http.Error(w, "Failed to parse form", http.StatusBadRequest)
        return
    }
    
    url := r.FormValue("url")
    
    if err := ValidateURL(url); err != nil {
        w.Header().Set("Content-Type", "text/html")
        h.templates.ExecuteTemplate(w, "url_error", map[string]string{"Error": err.Error()})
        return
    }
    
    w.Header().Set("Content-Type", "text/html")
    h.templates.ExecuteTemplate(w, "url_success", nil)
}

// Screen2Handler handles GET /wizard/2
func (h *Handler) Screen2Handler(w http.ResponseWriter, r *http.Request) {
    projectID := getProjectID(r)
    
    var project *database.Project
    var err error
    
    if projectID > 0 {
        project, err = database.GetProject(projectID)
        if err != nil || project == nil {
            http.Redirect(w, r, "/dashboard", http.StatusFound)
            return
        }
    } else {
        http.Redirect(w, r, "/wizard/1", http.StatusFound)
        return
    }
    
    if project.URL == "" {
        http.Redirect(w, r, "/wizard/1?project_id="+strconv.FormatInt(project.ID, 10), http.StatusFound)
        return
    }
    
    data := map[string]interface{}{
        "Project":   project,
        "Step":      2,
        "ProjectID": project.ID,
    }
    
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    if err := h.templates.ExecuteTemplate(w, "root", data); err != nil {
        http.Error(w, "Failed to render template", http.StatusInternalServerError)
    }
}

// Screen2PostHandler handles POST /wizard/2
func (h *Handler) Screen2PostHandler(w http.ResponseWriter, r *http.Request) {
    if err := r.ParseForm(); err != nil {
        http.Error(w, "Failed to parse form", http.StatusBadRequest)
        return
    }
    
    projectID := getProjectID(r)
    appName := r.FormValue("app_name")
    primaryColor := r.FormValue("primary_color")
    logoPath := r.FormValue("logo_path")
    
    if err := ValidateAppName(appName); err != nil {
        w.Header().Set("Content-Type", "text/html")
        w.WriteHeader(http.StatusBadRequest)
        h.templates.ExecuteTemplate(w, "app_name_error", map[string]string{"Error": err.Error()})
        return
    }
    
    if err := ValidateColor(primaryColor); err != nil {
        w.Header().Set("Content-Type", "text/html")
        w.WriteHeader(http.StatusBadRequest)
        h.templates.ExecuteTemplate(w, "color_error", map[string]string{"Error": err.Error()})
        return
    }
    
    project, err := database.GetProject(projectID)
    if err != nil || project == nil {
        http.Error(w, "Project not found", http.StatusNotFound)
        return
    }
    
    project.AppName = appName
    project.ProjectName = appName
    project.PrimaryColor = primaryColor
    project.LogoPath = logoPath
    project.Status = "draft"
    
    if err := database.UpdateProject(project); err != nil {
        http.Error(w, "Failed to save project", http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("HX-Redirect", "/dashboard")
    w.WriteHeader(http.StatusOK)
}

// ProjectStateHandler handles GET /api/project-state
func (h *Handler) ProjectStateHandler(w http.ResponseWriter, r *http.Request) {
    projectID := getProjectID(r)
    
    var project *database.Project
    var err error
    
    if projectID > 0 {
        project, err = database.GetProject(projectID)
    } else {
        project, err = database.LoadProject()
    }
    
    if err != nil || project == nil {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]string{"error": "No project found"})
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(project)
}
