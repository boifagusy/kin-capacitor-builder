package auth

import (
    "encoding/json"
    "net/http"

    "local-apk-builder/internal/database"
)

// StatusHandler returns GitHub connection status
func (h *Handler) StatusHandler(w http.ResponseWriter, r *http.Request) {
    conn, err := database.GetGitHubConnection()
    if err != nil {
        http.Error(w, "failed to get connection", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    if conn == nil {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "connected": false,
        })
        return
    }
    json.NewEncoder(w).Encode(map[string]interface{}{
        "connected": true,
        "username":  conn.GitHubUsername,
    })
}
