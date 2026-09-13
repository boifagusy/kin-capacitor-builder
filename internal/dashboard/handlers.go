package dashboard

import (
    "encoding/json"
    "html/template"
	"time"
    "log"
    "net/http"
    "strconv"
    "strings"

    "local-apk-builder/internal/database"
    "local-apk-builder/internal/generator"
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
    // Auto-redirect to GitHub auth if token is expired
    if conn, _ := database.GetGitHubConnection(); conn != nil && conn.AccessToken != "" {
        if parsedTime, err := time.Parse("2006-01-02 15:04:05", conn.UpdatedAt); err == nil {
            if time.Since(parsedTime).Minutes() > 120 {
                http.Redirect(w, r, "/auth/github", http.StatusFound)
                return
            }
        }
    }
    projects, err := h.service.List()
    if err != nil {
        log.Printf("Error listing projects: %v", err)
        http.Error(w, "Failed to load projects", http.StatusInternalServerError)
        return
    }

    conn, _ := database.GetGitHubConnection()
    githubUsername := ""
    tokenAgeMinutes := -1
    tokenExpired := true
    if conn != nil {
        githubUsername = conn.GitHubUsername
        if parsedTime, err := time.Parse("2006-01-02 15:04:05", conn.UpdatedAt); err == nil {
            tokenAgeMinutes = int(time.Since(parsedTime).Minutes())
        } else {
            log.Printf("DEBUG Dashboard: time parse error: %v for value: %q", err, conn.UpdatedAt)
        }
        tokenExpired = tokenAgeMinutes > 60 // Warn after 1 hour
    log.Printf("DEBUG Dashboard: username=%s, age=%d min, expired=%v", githubUsername, tokenAgeMinutes, tokenExpired)
    }

    data := map[string]interface{}{
        "Projects":        projects,
        "Count":           len(projects),
        "GitHubUsername":  githubUsername,
        "TokenAgeMinutes":  tokenAgeMinutes,
        "TokenExpired":     tokenExpired,
        "TokenExpiringSoon": tokenAgeMinutes > 30 && tokenAgeMinutes <= 60,
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

    // Get latest build (any status) and latest successful build
    builds, _ := database.ListBuildsForProject(p.ID)
    var latestBuild *database.Build
    var latestSuccessBuild *database.Build
    for i := range builds {
        if latestBuild == nil {
            latestBuild = builds[i]
        }
        if builds[i].Status == "success" && latestSuccessBuild == nil {
            latestSuccessBuild = builds[i]
        }
        if latestBuild != nil && latestSuccessBuild != nil {
            break
        }
    }

    // Get all builds for history
    allBuilds, _ := database.ListBuildsForProject(p.ID)

    // Get token update time
    tokenUpdate := ""
    if conn, _ := database.GetGitHubConnection(); conn != nil {
        tokenUpdate = conn.UpdatedAt
    }

    data := map[string]interface{}{
        "Project":           p,
        "LatestBuild":       latestBuild,
        "LatestSuccessBuild": latestSuccessBuild,
        "Builds":            allBuilds,
        "LastTokenUpdate":   tokenUpdate,
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
    // Handle POST for project creation
    if r.Method == http.MethodPost {
        h.CreateProjectHandler(w, r)
        return
    }
    // ... GET handling below
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
var _ = generator.GenerateForProject

// LoginPageHandler shows login screen
func (h *Handler) LoginPageHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    h.templates.ExecuteTemplate(w, "login", nil)
}

// PreviewPageHandler shows preview screen
func (h *Handler) PreviewPageHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    h.templates.ExecuteTemplate(w, "preview", nil)
}
