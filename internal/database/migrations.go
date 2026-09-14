package database

import (
	"database/sql"
	"strings"
)

func migrate(db *sql.DB) error {
    // Drop existing tables (no production data)
    
    // Create project table
    projectQuery := `
    CREATE TABLE IF NOT EXISTS project (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        project_name TEXT NOT NULL DEFAULT '',
        url TEXT NOT NULL DEFAULT '',
        app_name TEXT NOT NULL DEFAULT '',
        logo_path TEXT NOT NULL DEFAULT '',
        primary_color TEXT NOT NULL DEFAULT '#6366f1',
        features_config TEXT NOT NULL DEFAULT '[]',
        splash_config TEXT NOT NULL DEFAULT '{}',
        onboarding_config TEXT NOT NULL DEFAULT '{}',
        bridge_config TEXT NOT NULL DEFAULT '{}',
        google_services TEXT NOT NULL DEFAULT '',
        source_type TEXT NOT NULL DEFAULT 'url',
        source_path TEXT NOT NULL DEFAULT '',
        project_type TEXT NOT NULL DEFAULT 'static',
        original_zip_name TEXT NOT NULL DEFAULT '',
        runtime_type TEXT NOT NULL DEFAULT 'none',
        current_step INTEGER NOT NULL DEFAULT 1,
        status TEXT NOT NULL DEFAULT 'draft',
        version TEXT NOT NULL DEFAULT '1.0.0',
        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
    );
    CREATE INDEX IF NOT EXISTS idx_project_id ON project(id);
    `
    
    if _, err := db.Exec(projectQuery); err != nil {
        return err
    }

    // Create github oauth config table
    oauthConfigQuery := `
    CREATE TABLE IF NOT EXISTS github_oauth_config (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        client_id TEXT NOT NULL DEFAULT '',
        client_secret TEXT NOT NULL DEFAULT '',
        callback_url TEXT NOT NULL DEFAULT 'http://127.0.0.1:8080/auth/github/callback',
        updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
    );
    `
    if _, err := db.Exec(oauthConfigQuery); err != nil {
        return err
    }

    // Create github connection table
    githubQuery := `
    CREATE TABLE IF NOT EXISTS github_connection (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        github_username TEXT NOT NULL UNIQUE,
        access_token TEXT NOT NULL,
        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
    );
    `
    if _, err := db.Exec(githubQuery); err != nil {
        return err
    }
    
    // Create build table
    buildQuery := `
    CREATE TABLE IF NOT EXISTS build (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        project_id INTEGER NOT NULL,
        version TEXT NOT NULL DEFAULT '1.0.0',
        status TEXT NOT NULL DEFAULT 'queued',
        build_provider TEXT NOT NULL DEFAULT '',
        provider_run_id TEXT NOT NULL DEFAULT '',
        provider_artifact_name TEXT NOT NULL DEFAULT '',
        artifact_path TEXT NOT NULL DEFAULT '',
        artifact_size INTEGER NOT NULL DEFAULT 0,
        artifact_hash TEXT NOT NULL DEFAULT '',
        error_message TEXT NOT NULL DEFAULT '',
        started_at TIMESTAMP,
        completed_at TIMESTAMP,
        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (project_id) REFERENCES project(id) ON DELETE CASCADE
    );
    CREATE INDEX IF NOT EXISTS idx_build_project_id ON build(project_id);
    `
    
    if _, err := db.Exec(buildQuery); err != nil {
        return err
    }
    return migrateV5(db)
}

// SaveProject saves or updates a project
func SaveProject(project *Project) error {
    if project.ID == 0 {
        return InsertProject(project)
    }
    return UpdateProject(project)
}

// InsertProject creates a new project
func InsertProject(project *Project) error {
    query := `
    INSERT INTO project (project_name, url, app_name, logo_path, primary_color, features_config, splash_config, onboarding_config, bridge_config, google_services, current_step, status, version, runtime_type)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `
    
    result, err := db.Exec(query,
        project.ProjectName,
        project.URL,
        project.AppName,
        project.LogoPath,
        project.PrimaryColor,
        project.FeaturesConfig,
        project.SplashConfig,
        project.OnboardingConfig,
        project.BridgeConfig,
        project.GoogleServices,
        project.CurrentStep,
        project.Status,
        project.Version,
        project.RuntimeType,
    )
    if err != nil {
        return err
    }
    
    project.ID, err = result.LastInsertId()
    return err
}

// UpdateProject updates an existing project
func UpdateProject(project *Project) error {
    query := `
    UPDATE project SET 
        project_name = ?,
        url = ?,
        app_name = ?,
        logo_path = ?,
        primary_color = ?,
        features_config = ?,
        splash_config = ?,
        onboarding_config = ?,
        bridge_config = ?,
        google_services = ?,
        current_step = ?,
        status = ?,
        version = ?,
        source_type = ?,
        source_path = ?,
        project_type = ?,
        original_zip_name = ?,
        runtime_type = ?,
        updated_at = CURRENT_TIMESTAMP
    WHERE id = ?
    `
    
    _, err := db.Exec(query,
        project.ProjectName,
        project.URL,
        project.AppName,
        project.LogoPath,
        project.PrimaryColor,
        project.FeaturesConfig,
        project.SplashConfig,
        project.OnboardingConfig,
        project.BridgeConfig,
        project.GoogleServices,
        project.CurrentStep,
        project.Status,
        project.Version,
        project.SourceType,
        project.SourcePath,
        project.ProjectType,
        project.OriginalZipName,
        project.RuntimeType,
        project.ID,
    )
    return err
}

// GetProject loads a project by ID
func GetProject(id int64) (*Project, error) {
    query := `
    SELECT id, project_name, url, app_name, logo_path, primary_color, features_config, splash_config, onboarding_config, current_step, status, version, source_type, source_path, project_type, original_zip_name, runtime_type, created_at, updated_at
    FROM project WHERE id = ?
    `
    
    project := &Project{}
    err := db.QueryRow(query, id).Scan(
        &project.ID,
        &project.ProjectName,
        &project.URL,
        &project.AppName,
        &project.LogoPath,
        &project.PrimaryColor,
        &project.FeaturesConfig,
        &project.SplashConfig,
        &project.OnboardingConfig,
        &project.CurrentStep,
        &project.Status,
        &project.Version,
        &project.SourceType,
        &project.SourcePath,
        &project.ProjectType,
        &project.OriginalZipName,
        &project.RuntimeType,
        &project.CreatedAt,
        &project.UpdatedAt,
    )
    
    if err == sql.ErrNoRows {
        return nil, nil
    }
    if err != nil {
        return nil, err
    }
    
    return project, nil
}

// ListProjects returns all projects
func ListProjects() ([]*Project, error) {
    query := `
    SELECT id, project_name, url, app_name, logo_path, primary_color, features_config, splash_config, onboarding_config, current_step, status, version, created_at, updated_at
    FROM project ORDER BY updated_at DESC
    `
    
    rows, err := db.Query(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var projects []*Project
    for rows.Next() {
        project := &Project{}
        err := rows.Scan(
            &project.ID,
            &project.ProjectName,
            &project.URL,
            &project.AppName,
            &project.LogoPath,
            &project.PrimaryColor,
            &project.FeaturesConfig,
            &project.SplashConfig,
            &project.OnboardingConfig,
            &project.CurrentStep,
            &project.Status,
            &project.Version,
            &project.CreatedAt,
            &project.UpdatedAt,
        )
        if err != nil {
            return nil, err
        }
        projects = append(projects, project)
    }
    
    return projects, nil
}

// DeleteProject deletes a project by ID
func DeleteProject(id int64) error {
    _, err := db.Exec("DELETE FROM project WHERE id = ?", id)
    return err
}

// LoadProject loads the first project (backward compatibility)
func LoadProject() (*Project, error) {
    projects, err := ListProjects()
    if err != nil {
        return nil, err
    }
    
    if len(projects) == 0 {
        project := NewProject()
        err = InsertProject(project)
        if err != nil {
            return nil, err
        }
        return project, nil
    }
    
    return projects[0], nil
}


// Migration v5: Ensure all project columns exist (idempotent)
func migrateV5(db *sql.DB) error {
    alterations := []string{
        "ALTER TABLE project ADD COLUMN bridge_config TEXT NOT NULL DEFAULT '{}'",
        "ALTER TABLE project ADD COLUMN google_services TEXT NOT NULL DEFAULT ''",
        "ALTER TABLE project ADD COLUMN source_type TEXT NOT NULL DEFAULT 'url'",
        "ALTER TABLE project ADD COLUMN source_path TEXT NOT NULL DEFAULT ''",
        "ALTER TABLE project ADD COLUMN project_type TEXT NOT NULL DEFAULT 'static'",
        "ALTER TABLE project ADD COLUMN original_zip_name TEXT NOT NULL DEFAULT ''",
        "ALTER TABLE project ADD COLUMN runtime_type TEXT NOT NULL DEFAULT 'none'",
    }
    for _, stmt := range alterations {
        if _, err := db.Exec(stmt); err != nil {
            if !strings.Contains(err.Error(), "duplicate column") {
                return err
            }
        }
    }
    return nil
}
