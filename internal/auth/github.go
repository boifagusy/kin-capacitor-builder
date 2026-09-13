package auth

import (
    "encoding/json"
    "fmt"
    "io"
    "html/template"
    "net/http"
    "net/url"
    "strings"

    "local-apk-builder/internal/config"
    "local-apk-builder/internal/database"
)

type Handler struct {
    oauth     *config.OAuthConfig
    templates *template.Template
}

func NewHandler(oauth *config.OAuthConfig, templates *template.Template) *Handler {
    return &Handler{oauth: oauth, templates: templates}
}

// Login redirects to GitHub OAuth
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
    // Try DB config first
    dbCfg, _ := database.GetGitHubOAuthConfig()
    if dbCfg != nil && dbCfg.ClientID != "" {
        h.oauth.ClientID = dbCfg.ClientID
        h.oauth.ClientSecret = dbCfg.ClientSecret
        h.oauth.CallbackURL = dbCfg.CallbackURL
    }

    if h.oauth.ClientID == "" {
        http.Error(w, "GitHub OAuth not configured", http.StatusInternalServerError)
        return
    }
    redirectURL := fmt.Sprintf(
        "https://github.com/login/oauth/authorize?client_id=%s&scope=repo,workflow&redirect_uri=%s",
        h.oauth.ClientID,
        url.QueryEscape(h.oauth.CallbackURL),
    )
    http.Redirect(w, r, redirectURL, http.StatusFound)
}

// Callback handles GitHub OAuth callback
func (h *Handler) Callback(w http.ResponseWriter, r *http.Request) {
    code := r.URL.Query().Get("code")
    if code == "" {
        http.Error(w, "missing code", http.StatusBadRequest)
        return
    }

    token, err := h.exchangeCode(code)
    if err != nil {
        http.Error(w, fmt.Sprintf("token exchange failed: %v", err), http.StatusInternalServerError)
        return
    }

    username, err := h.getUsername(token)
    if err != nil {
        http.Error(w, "failed to get username", http.StatusInternalServerError)
        return
    }

    conn := &database.GitHubConnection{
        GitHubUsername: username,
        AccessToken:    token,
    }
    // Don't overwrite PAT with OAuth token
    existing, _ := database.GetGitHubConnection()
    if existing != nil && strings.HasPrefix(existing.AccessToken, "ghp_") {
        http.Redirect(w, r, "/dashboard", http.StatusFound)
        return
    }

    if err := database.SaveGitHubConnection(conn); err != nil {
        http.Error(w, "failed to save token", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "text/html")
    fmt.Fprintf(w, `<html><body><h2>Connected as @%s</h2><p><a href="/dashboard">Go to Dashboard</a></p></body></html>`, username)
}

func (h *Handler) exchangeCode(code string) (string, error) {
    data := url.Values{}
    data.Set("client_id", h.oauth.ClientID)
    data.Set("client_secret", h.oauth.ClientSecret)
    data.Set("code", code)
    data.Set("redirect_uri", h.oauth.CallbackURL)

    req, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", strings.NewReader(data.Encode()))
    if err != nil {
        return "", err
    }
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    req.Header.Set("Accept", "application/json")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    body, _ := io.ReadAll(resp.Body)
    var result struct {
        AccessToken string `json:"access_token"`
        Error       string `json:"error"`
        ErrorDesc   string `json:"error_description"`
    }
    if err := json.Unmarshal(body, &result); err != nil {
        return "", fmt.Errorf("failed to parse response: %s", string(body))
    }

    if result.AccessToken != "" {
        return result.AccessToken, nil
    }
    if result.Error != "" {
        return "", fmt.Errorf("%s: %s", result.Error, result.ErrorDesc)
    }
    return "", fmt.Errorf("no access token in response: %s", string(body))
}

func (h *Handler) getUsername(token string) (string, error) {
    req, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
    req.Header.Set("Authorization", "Bearer "+token)
    req.Header.Set("Accept", "application/vnd.github+json")
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    body, _ := io.ReadAll(resp.Body)
    var user struct {
        Login string `json:"login"`
    }
    if err := json.Unmarshal(body, &user); err != nil {
        return "", err
    }
    return user.Login, nil
}
