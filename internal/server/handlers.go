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

func (h *Handler) IndexHandler(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path != "/" {
        http.NotFound(w, r)
        return
    }
    
    if r.Method != http.MethodGet && r.Method != http.MethodHead {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    project, err := database.LoadProject()
    if err != nil {
        log.Printf("Error loading project: %v", err)
        http.Error(w, "Failed to load project", http.StatusInternalServerError)
        return
    }
    
    redirectURL := "/wizard/" + strconv.Itoa(project.CurrentStep)
    if r.Method == http.MethodHead {
        w.Header().Set("Location", redirectURL)
        w.WriteHeader(http.StatusFound)
        return
    }
    
    http.Redirect(w, r, redirectURL, http.StatusFound)
}

func (h *Handler) Screen1Handler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet && r.Method != http.MethodHead {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    project, err := database.LoadProject()
    if err != nil {
        log.Printf("Error loading project: %v", err)
        http.Error(w, "Failed to load project", http.StatusInternalServerError)
        return
    }
    
    data := map[string]interface{}{
        "Project": project,
        "Step":    1,
    }
    
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    if err := h.templates.ExecuteTemplate(w, "root", data); err != nil {
        log.Printf("Template execution error: %v", err)
        http.Error(w, "Failed to render template", http.StatusInternalServerError)
    }
}

func (h *Handler) Screen1PostHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    if err := r.ParseForm(); err != nil {
        http.Error(w, "Failed to parse form", http.StatusBadRequest)
        return
    }
    
    url := r.FormValue("url")
    
    if err := ValidateURL(url); err != nil {
        w.Header().Set("Content-Type", "text/html")
        w.WriteHeader(http.StatusBadRequest)
        h.templates.ExecuteTemplate(w, "url_error", map[string]string{"Error": err.Error()})
        return
    }
    
    normalizedURL := NormalizeURL(url)
    
    project, err := database.LoadProject()
    if err != nil {
        log.Printf("Error loading project: %v", err)
        http.Error(w, "Failed to load project", http.StatusInternalServerError)
        return
    }
    
    project.URL = normalizedURL
    project.CurrentStep = 2
    
    if err := database.SaveProject(project); err != nil {
        log.Printf("Error saving project: %v", err)
        http.Error(w, "Failed to save project", http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("HX-Redirect", "/wizard/2")
    w.WriteHeader(http.StatusOK)
}

func (h *Handler) ValidateURLHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
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

func (h *Handler) Screen2Handler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet && r.Method != http.MethodHead {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    project, err := database.LoadProject()
    if err != nil {
        log.Printf("Error loading project: %v", err)
        http.Error(w, "Failed to load project", http.StatusInternalServerError)
        return
    }
    
    if project.URL == "" {
        if r.Method == http.MethodHead {
            w.Header().Set("Location", "/wizard/1")
            w.WriteHeader(http.StatusFound)
            return
        }
        http.Redirect(w, r, "/wizard/1", http.StatusFound)
        return
    }
    
    data := map[string]interface{}{
        "Project": project,
        "Step":    2,
    }
    
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    if err := h.templates.ExecuteTemplate(w, "root", data); err != nil {
        log.Printf("Template execution error: %v", err)
        http.Error(w, "Failed to render template", http.StatusInternalServerError)
    }
}

func (h *Handler) Screen2PostHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    if err := r.ParseForm(); err != nil {
        http.Error(w, "Failed to parse form", http.StatusBadRequest)
        return
    }
    
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
    
    project, err := database.LoadProject()
    if err != nil {
        log.Printf("Error loading project: %v", err)
        http.Error(w, "Failed to load project", http.StatusInternalServerError)
        return
    }
    
    project.AppName = appName
    project.PrimaryColor = primaryColor
    project.LogoPath = logoPath
    
    if err := database.SaveProject(project); err != nil {
        log.Printf("Error saving project: %v", err)
        http.Error(w, "Failed to save project", http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "text/html")
    h.templates.ExecuteTemplate(w, "branding_success", map[string]interface{}{"Project": project})
}

func (h *Handler) ProjectStateHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet && r.Method != http.MethodHead {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    project, err := database.LoadProject()
    if err != nil {
        log.Printf("Error loading project: %v", err)
        http.Error(w, "Failed to load project", http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    if r.Method == http.MethodHead {
        return
    }
    json.NewEncoder(w).Encode(project)
}
