package dashboard

import (
    "encoding/json"
    "net/http"
    "strconv"
    "strings"
    
    "local-apk-builder/internal/database"
)

// SaveNavigationHandler handles POST /api/projects/{id}/navigation
func (h *Handler) SaveNavigationHandler(w http.ResponseWriter, r *http.Request) {
    path := strings.TrimPrefix(r.URL.Path, "/api/projects/")
    path = strings.TrimSuffix(path, "/navigation")
    id, err := strconv.ParseInt(path, 10, 64)
    if err != nil {
        http.NotFound(w, r)
        return
    }
    
    var payload struct {
        NavStyle string `json:"nav_style"`
        Items    []struct {
            Icon      string `json:"icon"`
            URL       string `json:"url"`
            Animation string `json:"animation"`
        } `json:"items"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    
    project, err := h.service.Get(id)
    if err != nil || project == nil {
        http.NotFound(w, r)
        return
    }
    
    type NavItemWithAnimation struct {
        Icon      string `json:"icon"`
        URL       string `json:"url"`
        Animation string `json:"animation"`
    }

    navItems := make([]NavItemWithAnimation, len(payload.Items))
    for i, item := range payload.Items {
        navItems[i] = NavItemWithAnimation{
            Icon:      item.Icon,
            URL:       item.URL,
            Animation: item.Animation,
        }
    }

    navConfig := map[string]interface{}{
        "style": payload.NavStyle,
        "items": navItems,
    }
    
    navJSON, _ := json.Marshal(navConfig)
    project.OnboardingConfig = string(navJSON)
    
    if err := database.UpdateProject(project); err != nil {
        http.Error(w, "Failed to save", http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"status": "saved"})
}
