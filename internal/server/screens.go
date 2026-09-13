package server

import (
    "io"
    "html/template"
    "encoding/json"
    "net/http"
    "strconv"

    "local-apk-builder/internal/database"
	"local-apk-builder/internal/generator"
)

// Screen3Handler - Splash Screen
func (h *Handler) Screen3Handler(w http.ResponseWriter, r *http.Request) {
    projectID := getProjectID(r)
    project, err := database.GetProject(projectID)
    if err != nil || project == nil {
        http.Redirect(w, r, "/dashboard", http.StatusFound)
        return
    }

    var sc struct {
        BackgroundColor string `json:"background_color"`
        ShowLogo        bool   `json:"show_logo"`
        LoadingText     string `json:"loading_text"`
        Duration        int    `json:"duration"`
    }
    json.Unmarshal([]byte(project.SplashConfig), &sc)
    if sc.BackgroundColor == "" {
        sc.BackgroundColor = project.PrimaryColor
    }
    if sc.LoadingText == "" {
        sc.LoadingText = "Loading..."
    }
    if sc.Duration == 0 {
        sc.Duration = 1500
    }

    data := map[string]interface{}{
        "Project":     project,
        "Step":        3,
        "ProjectID":   project.ID,
        "SplashBg":    sc.BackgroundColor,
        "ShowLogo":    sc.ShowLogo,
        "LoadingText": sc.LoadingText,
        "Duration":    sc.Duration,
        "LogoPath":    project.LogoPath,
    }
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    h.templates.ExecuteTemplate(w, "root", data)
}

func (h *Handler) Screen3PostHandler(w http.ResponseWriter, r *http.Request) {
    r.ParseMultipartForm(10 << 20)
    projectID := getProjectID(r)
    project, err := database.GetProject(projectID)
    if err != nil || project == nil {
        http.Error(w, "Project not found", http.StatusNotFound)
        return
    }
    splashConfig := map[string]interface{}{
        "background_color": r.FormValue("splash_bg"),
        "show_logo":        r.FormValue("show_logo") == "on",
        "loading_text":     r.FormValue("loading_text"),
        "duration":         r.FormValue("duration"),
    }
    splashJSON, _ := json.Marshal(splashConfig)
    project.SplashConfig = string(splashJSON)
    project.CurrentStep = 3
    database.UpdateProject(project)
    w.Header().Set("HX-Redirect", "/wizard/4?project_id="+strconv.FormatInt(projectID, 10))
    w.WriteHeader(http.StatusOK)
}

// Screen4Handler - Features (Plugins)
func (h *Handler) Screen4Handler(w http.ResponseWriter, r *http.Request) {
    projectID := getProjectID(r)
    project, err := database.GetProject(projectID)
    if err != nil || project == nil {
        http.Redirect(w, r, "/dashboard", http.StatusFound)
        return
    }
    data := map[string]interface{}{"Project": project, "Step": 4, "ProjectID": project.ID}
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    h.templates.ExecuteTemplate(w, "root", data)
}

func (h *Handler) Screen4PostHandler(w http.ResponseWriter, r *http.Request) {
    r.ParseMultipartForm(10 << 20)
    projectID := getProjectID(r)
    project, err := database.GetProject(projectID)
    if err != nil || project == nil {
        http.Error(w, "Project not found", http.StatusNotFound)
        return
    }
    features := r.FormValue("features")
    project.FeaturesConfig = features

    bridgeEnabled := r.FormValue("bridge_enabled") == "on"
    if bridgeEnabled {
        bridgeConfig := map[string]interface{}{
            "enabled":           true,
            "url":               r.FormValue("bridge_url"),
            "auth_mode":         r.FormValue("bridge_auth_mode"),
            "notification_mode": r.FormValue("bridge_notification_mode"),
            "poll_interval":     r.FormValue("bridge_poll_interval"),
        }
        jsonBytes, _ := json.Marshal(bridgeConfig)
        project.BridgeConfig = string(jsonBytes)
    } else {
        project.BridgeConfig = ""
    }

    // Handle google-services.json upload
    file, _, err := r.FormFile("google_services_json")
    if err == nil {
        defer file.Close()
        data, _ := io.ReadAll(file)
        if len(data) > 0 {
            project.GoogleServices = string(data)
        }
    }

    project.CurrentStep = 4
    database.UpdateProject(project)
    w.Header().Set("HX-Redirect", "/wizard/5?project_id="+strconv.FormatInt(projectID, 10))
    w.WriteHeader(http.StatusOK)
}

// Screen5Handler - Navigation
func (h *Handler) Screen5Handler(w http.ResponseWriter, r *http.Request) {
    projectID := getProjectID(r)
    project, err := database.GetProject(projectID)
    if err != nil || project == nil {
        http.Redirect(w, r, "/dashboard", http.StatusFound)
        return
    }
    navConfig := project.OnboardingConfig
    if navConfig == "" {
        navConfig = `{"style":"bottom","template":"bottom-tabs","items":[]}`
    }
    data := map[string]interface{}{
        "Project":   project,
        "Step":      5,
        "ProjectID": project.ID,
        "NavConfig": template.JS(navConfig),
    }
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    h.templates.ExecuteTemplate(w, "root", data)
}

func (h *Handler) Screen5PostHandler(w http.ResponseWriter, r *http.Request) {
    projectID := getProjectID(r)
    project, err := database.GetProject(projectID)
    if err != nil || project == nil {
        http.Error(w, "Project not found", http.StatusNotFound)
        return
    }
    project.CurrentStep = 5
    database.UpdateProject(project)
    w.Header().Set("HX-Redirect", "/wizard/6?project_id="+strconv.FormatInt(projectID, 10))
    w.WriteHeader(http.StatusOK)
}

// Screen6Handler - Review & Preview
func (h *Handler) Screen6Handler(w http.ResponseWriter, r *http.Request) {
    projectID := getProjectID(r)
    project, err := database.GetProject(projectID)
    if err != nil || project == nil {
        http.Redirect(w, r, "/dashboard", http.StatusFound)
        return
    }
    data := map[string]interface{}{"Project": project, "Step": 6, "ProjectID": project.ID}
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    h.templates.ExecuteTemplate(w, "root", data)
}

// PreviewHandler serves the generated splash/redirect page for native browser preview
func (h *Handler) PreviewHandler(w http.ResponseWriter, r *http.Request) {
    projectID := getProjectID(r)
    project, err := database.GetProject(projectID)
    if err != nil || project == nil {
        http.NotFound(w, r)
        return
    }
    gen := generator.NewGenerator(project)
    webIndex := gen.GenerateWebIndex()
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    w.Write([]byte(webIndex))
}
