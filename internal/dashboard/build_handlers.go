package dashboard

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"local-apk-builder/internal/build"
	"local-apk-builder/internal/config"
	"local-apk-builder/internal/database"
)

func getBuildManager() *build.Manager {
	// Read token fresh from DB every time
	conn, _ := database.GetGitHubConnection()
	if conn != nil && conn.AccessToken != "" {
		provider := build.NewGitHubActionsProvider(
			conn.AccessToken,
			conn.GitHubUsername+"/app-builder",
			"build-apk.yml",
			"local-apk-builder-debug",
		)
		return build.NewManager(provider)
	}

	// Fallback to env config
	ghCfg := config.LoadGitHubConfig()
	if ghCfg.Token != "" && ghCfg.Repo != "" {
		provider := build.NewGitHubActionsProvider(
			ghCfg.Token,
			ghCfg.Repo,
			ghCfg.WorkflowID,
			ghCfg.ArtifactName,
		)
		return build.NewManager(provider)
	}

	// Fallback to local
	return build.NewManager(build.NewLocalProvider())
}

// EnqueueBuildHandler handles POST /projects/{id}/build
func (h *Handler) EnqueueBuildHandler(w http.ResponseWriter, r *http.Request) {
	// Check token freshness before build
	conn, _ := database.GetGitHubConnection()
	if conn == nil || conn.AccessToken == "" {
		w.Header().Set("HX-Redirect", "/auth/github")
		http.Error(w, "GitHub authentication required", http.StatusUnauthorized)
		return
	}

	// Check token age (expires ~2-3 hours)
	if parsedTime, err := time.Parse("2006-01-02 15:04:05", conn.UpdatedAt); err == nil {
		if time.Since(parsedTime).Minutes() > 120 {
			w.Header().Set("HX-Redirect", "/auth/github")
			http.Error(w, "GitHub token expired. Redirecting to re-authenticate...", http.StatusUnauthorized)
			return
		}
	}
	path := strings.TrimPrefix(r.URL.Path, "/projects/")
	path = strings.TrimSuffix(path, "/build")
	projectID, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// WEB-SOURCE-005: read runtime_type from POST body and persist it
	_ = r.ParseForm()
	if rt := r.FormValue("runtime_type"); rt != "" {
		if rt != "none" && rt != "go-sqlite" && rt != "go-native" {
			http.Error(w, "invalid runtime_type", http.StatusBadRequest)
			return
		}
		if project, err := database.GetProject(projectID); err == nil && project != nil {
			if project.RuntimeType != rt {
				project.RuntimeType = rt
				_ = database.UpdateProject(project)
				log.Printf("DEBUG: RuntimeType updated to %q for project %d", rt, projectID)
			}
		}
	}

	version := r.URL.Query().Get("version")
	b, err := getBuildManager().EnqueueBuild(projectID, version, "apk")
	if err != nil {
		log.Printf("Build error: %v", err)
		http.Error(w, "Could not start build. Please try again.", http.StatusInternalServerError)
		return
	}

	go getBuildManager().ProcessBuild(b.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(b)
}

// EnqueueAABBuildHandler handles POST /projects/{id}/build-aab
func (h *Handler) EnqueueAABBuildHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/projects/")
	path = strings.TrimSuffix(path, "/build-aab")
	projectID, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// WEB-SOURCE-005: read runtime_type from POST body and persist it
	_ = r.ParseForm()
	if rt := r.FormValue("runtime_type"); rt != "" {
		if rt != "none" && rt != "go-sqlite" && rt != "go-native" {
			http.Error(w, "invalid runtime_type", http.StatusBadRequest)
			return
		}
		if project, err := database.GetProject(projectID); err == nil && project != nil {
			if project.RuntimeType != rt {
				project.RuntimeType = rt
				_ = database.UpdateProject(project)
				log.Printf("DEBUG: RuntimeType updated to %q for project %d", rt, projectID)
			}
		}
	}

	version := r.URL.Query().Get("version")
	b, err := getBuildManager().EnqueueBuild(projectID, version, "aab")
	if err != nil {
		log.Printf("Build error: %v", err)
		http.Error(w, "Could not start build. Please try again.", http.StatusInternalServerError)
		return
	}

	go getBuildManager().ProcessBuild(b.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(b)
}

// ListBuildsHandler handles GET /projects/{id}/builds
func (h *Handler) ListBuildsHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/projects/")
	path = strings.TrimSuffix(path, "/builds")
	projectID, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	builds, err := getBuildManager().ListProjectBuilds(projectID)
	if err != nil {
		log.Printf("Build error: %v", err)
		http.Error(w, "Could not start build. Please try again.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(builds)
}

// BuildStatusHTMLHandler handles GET /projects/{id}/build-status (returns HTML fragment)
func (h *Handler) BuildStatusHTMLHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/projects/")
	path = strings.TrimSuffix(path, "/build-status")
	projectID, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	builds, err := getBuildManager().ListProjectBuilds(projectID)
	if err != nil {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, "<span style='color:#ef4444;'>Error: %v</span>", err)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if len(builds) == 0 {
		fmt.Fprint(w, "<span style='color:#9ca3af;'>No builds yet</span>")
		return
	}

	latest := builds[0]
	if latest.Status == "success" && latest.ArtifactPath != "" {
		fmt.Fprintf(w, "<div style='display:flex;flex-direction:column;gap:0.5rem;width:100%%;'>")
		fmt.Fprintf(w, "<a href='/projects/%d/download' class='btn btn-secondary' style='width:100%%;'>📥 Download APK (%d bytes)</a>", latest.ID, latest.ArtifactSize)
		fmt.Fprintf(w, "<span style='color:#16a34a;'>✅ Build #%d succeeded</span>", latest.ID)
		fmt.Fprintf(w, "</div>")
		return
	}
	switch latest.Status {
	case "queued":
		fmt.Fprintf(w, "<span style='color:#6366f1;'>📤 Queued — preparing build...</span>")
	case "submitted":
		fmt.Fprintf(w, "<span style='color:#6366f1;'>📤 Submitted — waiting for GitHub...</span>")
	case "building":
		// Calculate realistic progress based on elapsed time (5 min total)
		elapsed := int64(0)
		if latest.StartedAt != nil {
			elapsed = int64(time.Since(*latest.StartedAt).Seconds())
		} else {
			elapsed = int64(time.Since(latest.CreatedAt).Seconds())
		}
		progress := int(elapsed / 300.0 * 100)
		if progress > 90 {
			progress = 90
		}
		remaining := (300 - elapsed) / 60
		if remaining < 0 {
			remaining = 0
		}
		fmt.Fprintf(w, "<div style='display:flex;align-items:center;gap:0.5rem;'><span style='color:#f59e0b;font-size:0.85rem;'>🔨 Building</span><div style='flex:1;background:#e5e7eb;border-radius:999px;height:6px;overflow:hidden;'><div style='background:#f59e0b;height:100%%;width:%d%%;border-radius:999px;transition:width 1s;'></div></div><span style='font-size:0.75rem;color:#6b7280;'>%d%%</span></div><span style='font-size:0.7rem;color:#6b7280;'>~%d min remaining</span>", progress, progress, remaining)
	case "success":
		if latest.ArtifactPath == "" {
			fmt.Fprintf(w, "<span style='color:#16a34a;'>✅ Build #%d succeeded</span>", latest.ID)
			if latest.ProviderRunID != "" {
				username := ""
				if conn, _ := database.GetGitHubConnection(); conn != nil {
					username = conn.GitHubUsername
				}
				fmt.Fprintf(w, "<a href='https://github.com/%s?tab=repositories' target='_blank' style='color:#6366f1;font-size:0.8rem;'>📥 Download from GitHub</a>", username)
			}
		} else {
			fmt.Fprintf(w, "<span style='color:#16a34a;'>✅ Build #%d succeeded — APK ready</span>", latest.ID)
		}

	case "failed":
		if latest.ProviderRunID != "" && latest.ErrorMessage != "" && len(latest.ErrorMessage) > 0 {
			fmt.Fprintf(w, "<span style='color:#f59e0b;'>⚠️ Build #%d needs manual download</span>", latest.ID)
			username := ""
			if conn, _ := database.GetGitHubConnection(); conn != nil {
				username = conn.GitHubUsername
			}
			fmt.Fprintf(w, "<a href='https://github.com/%s?tab=repositories' target='_blank' style='color:#6366f1;font-size:0.8rem;'>📥 Download from GitHub</a>", username)
		} else {
			fmt.Fprintf(w, "<span style='color:#ef4444;'>❌ Build #%d failed</span>", latest.ID)
		}
	default:
		fmt.Fprintf(w, "<span>Status: %s</span>", latest.Status)
	}
}
