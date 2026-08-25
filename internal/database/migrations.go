package database

import "database/sql"

func migrate(db *sql.DB) error {
    // Drop existing tables (no production data)
    db.Exec("DROP TABLE IF EXISTS build")
    db.Exec("DROP TABLE IF EXISTS project")
    
    // Create project table
    projectQuery := `
    CREATE TABLE IF NOT EXISTS project (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        project_name TEXT NOT NULL DEFAULT '',
        url TEXT NOT NULL DEFAULT '',
        app_name TEXT NOT NULL DEFAULT '',
        logo_path TEXT NOT NULL DEFAULT '',
        primary_color TEXT NOT NULL DEFAULT '#6366f1',
        onboarding_config TEXT NOT NULL DEFAULT '{}',
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
    
    // Create build table
    buildQuery := `
    CREATE TABLE IF NOT EXISTS build (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        project_id INTEGER NOT NULL,
        version TEXT NOT NULL DEFAULT '1.0.0',
        status TEXT NOT NULL DEFAULT 'queued',
        build_provider TEXT NOT NULL DEFAULT '',
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
    
    _, err := db.Exec(buildQuery)
    return err
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
    INSERT INTO project (project_name, url, app_name, logo_path, primary_color, onboarding_config, current_step, status, version)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
    `
    
    result, err := db.Exec(query,
        project.ProjectName,
        project.URL,
        project.AppName,
        project.LogoPath,
        project.PrimaryColor,
        project.OnboardingConfig,
        project.CurrentStep,
        project.Status,
        project.Version,
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
        onboarding_config = ?,
        current_step = ?,
        status = ?,
        version = ?,
        updated_at = CURRENT_TIMESTAMP
    WHERE id = ?
    `
    
    _, err := db.Exec(query,
        project.ProjectName,
        project.URL,
        project.AppName,
        project.LogoPath,
        project.PrimaryColor,
        project.OnboardingConfig,
        project.CurrentStep,
        project.Status,
        project.Version,
        project.ID,
    )
    return err
}

// GetProject loads a project by ID
func GetProject(id int64) (*Project, error) {
    query := `
    SELECT id, project_name, url, app_name, logo_path, primary_color, onboarding_config, current_step, status, version, created_at, updated_at
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
        &project.OnboardingConfig,
        &project.CurrentStep,
        &project.Status,
        &project.Version,
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
    SELECT id, project_name, url, app_name, logo_path, primary_color, onboarding_config, current_step, status, version, created_at, updated_at
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
