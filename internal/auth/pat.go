package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"local-apk-builder/internal/database"
)

// PATValidateHandler accepts a POST with form field "pat_token",
// validates it against GitHub, and on success saves it with the username.
func (h *Handler) PATValidateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	token := r.FormValue("pat_token")
	if token == "" {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"valid": false, "error": "Empty token",
		})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"valid": false, "error": "request build failed",
		})
		return
	}
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "local-apk-builder")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]interface{}{
			"valid": false, "error": "network error reaching GitHub",
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"valid": false, "error": "Invalid or expired token",
		})
		return
	}
	if resp.StatusCode != http.StatusOK {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"valid": false, "error": fmt.Sprintf("GitHub returned 0", resp.StatusCode),
		})
		return
	}

	var user struct {
		Login string `json:"login"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil || user.Login == "" {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"valid": false, "error": "could not read username",
		})
		return
	}

	// Save with username
	if err := database.UpdateAccessTokenAndUsername(token, user.Login); err != nil {
		// If no existing row, insert fresh
		conn := &database.GitHubConnection{
			GitHubUsername: user.Login,
			AccessToken:    token,
		}
		if err2 := database.SaveGitHubConnection(conn); err2 != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"valid": false, "error": "failed to save: " + err2.Error(),
			})
			return
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"valid": true, "username": user.Login,
	})
}

func writeJSON(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}
