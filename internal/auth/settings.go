package auth

import (
    "net/http"

    "local-apk-builder/internal/database"
)

// SettingsPageHandler shows GitHub settings
func (h *Handler) SettingsPageHandler(w http.ResponseWriter, r *http.Request) {
    cfg, _ := database.GetGitHubOAuthConfig()
    data := map[string]interface{}{
        "ClientID":     "",
        "ClientSecret": "",
        "CallbackURL":  "http://127.0.0.1:8080/auth/github/callback",
    }
    if cfg != nil {
        data["ClientID"] = cfg.ClientID
        data["ClientSecret"] = cfg.ClientSecret
        data["CallbackURL"] = cfg.CallbackURL
    }

    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    h.templates.ExecuteTemplate(w, "github_settings", data)
}

// SaveSettingsHandler saves GitHub OAuth config
func (h *Handler) SaveSettingsHandler(w http.ResponseWriter, r *http.Request) {
    r.ParseForm()
    clientID := r.FormValue("client_id")
    clientSecret := r.FormValue("client_secret")
    callbackURL := r.FormValue("callback_url")

    cfg := &database.GitHubOAuthConfig{
        ClientID:     clientID,
        ClientSecret: clientSecret,
        CallbackURL:  callbackURL,
    }
    if err := database.SaveGitHubOAuthConfig(cfg); err != nil {
        http.Error(w, "Failed to save", http.StatusInternalServerError)
        return
    }

    http.Redirect(w, r, "/dashboard", http.StatusFound)
}
