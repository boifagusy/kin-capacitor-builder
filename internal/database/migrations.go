package database

import "database/sql"

func migrate(db *sql.DB) error {
    query := `
    CREATE TABLE IF NOT EXISTS project (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        url TEXT NOT NULL DEFAULT '',
        app_name TEXT NOT NULL DEFAULT '',
        logo_path TEXT NOT NULL DEFAULT '',
        primary_color TEXT NOT NULL DEFAULT '#3B82F6',
        onboarding_config TEXT NOT NULL DEFAULT '{"has_splash_screen":true,"splash_color":"#FFFFFF","orientation":"portrait"}',
        current_step INTEGER NOT NULL DEFAULT 1,
        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
    );
    CREATE INDEX IF NOT EXISTS idx_project_id ON project(id);
    `
    
    _, err := db.Exec(query)
    return err
}

func SaveProject(project *Project) error {
    query := `
    UPDATE project SET 
        url = ?,
        app_name = ?,
        logo_path = ?,
        primary_color = ?,
        onboarding_config = ?,
        current_step = ?,
        updated_at = CURRENT_TIMESTAMP
    WHERE id = ?
    `
    
    result, err := db.Exec(query,
        project.URL,
        project.AppName,
        project.LogoPath,
        project.PrimaryColor,
        project.OnboardingConfig,
        project.CurrentStep,
        project.ID,
    )
    if err != nil {
        return err
    }
    
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return err
    }
    
    if rowsAffected == 0 {
        query = `
        INSERT INTO project (url, app_name, logo_path, primary_color, onboarding_config, current_step)
        VALUES (?, ?, ?, ?, ?, ?)
        `
        
        result, err = db.Exec(query,
            project.URL,
            project.AppName,
            project.LogoPath,
            project.PrimaryColor,
            project.OnboardingConfig,
            project.CurrentStep,
        )
        if err != nil {
            return err
        }
        
        project.ID, err = result.LastInsertId()
        if err != nil {
            return err
        }
    }
    
    return nil
}

func LoadProject() (*Project, error) {
    query := `
    SELECT id, url, app_name, logo_path, primary_color, onboarding_config, current_step, created_at, updated_at
    FROM project
    ORDER BY id ASC
    LIMIT 1
    `
    
    project := &Project{}
    err := db.QueryRow(query).Scan(
        &project.ID,
        &project.URL,
        &project.AppName,
        &project.LogoPath,
        &project.PrimaryColor,
        &project.OnboardingConfig,
        &project.CurrentStep,
        &project.CreatedAt,
        &project.UpdatedAt,
    )
    
    if err == sql.ErrNoRows {
        project = NewProject()
        err = SaveProject(project)
        if err != nil {
            return nil, err
        }
    } else if err != nil {
        return nil, err
    }
    
    return project, nil
}
