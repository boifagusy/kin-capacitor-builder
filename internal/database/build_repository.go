package database

import "database/sql"

func InsertBuild(build *Build) error {
    query := `
    INSERT INTO build (build_type, project_id, version, status, build_provider, provider_run_id, provider_artifact_name, artifact_path, artifact_size, artifact_hash, error_message, started_at, completed_at)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `
    result, err := db.Exec(query,
        build.BuildType, build.ProjectID, build.Version, build.Status, build.BuildProvider,
        build.ProviderRunID, build.ProviderArtifactName,
        build.ArtifactPath, build.ArtifactSize, build.ArtifactHash,
        build.ErrorMessage, build.StartedAt, build.CompletedAt,
    )
    if err != nil {
        return err
    }
    build.ID, err = result.LastInsertId()
    return err
}

func UpdateBuild(build *Build) error {
    query := `
    UPDATE build SET
        build_type = ?,
        repo_name = ?,
        status = ?,
        build_provider = ?,
        provider_run_id = ?,
        provider_artifact_name = ?,
        artifact_path = ?,
        artifact_size = ?,
        artifact_hash = ?,
        error_message = ?,
        started_at = ?,
        completed_at = ?
    WHERE id = ?
    `
    _, err := db.Exec(query,
        build.BuildType, build.RepoName, build.Status, build.BuildProvider, build.ProviderRunID,
        build.ProviderArtifactName, build.ArtifactPath, build.ArtifactSize,
        build.ArtifactHash, build.ErrorMessage, build.StartedAt, build.CompletedAt,
        build.ID,
    )
    return err
}

func GetBuild(id int64) (*Build, error) {
    query := `SELECT id, build_type, repo_name, project_id, version, status, build_provider, provider_run_id, provider_artifact_name, artifact_path, artifact_size, artifact_hash, error_message, started_at, completed_at, created_at FROM build WHERE id = ?`
    build := &Build{}
    err := db.QueryRow(query, id).Scan(
        &build.ID, &build.BuildType, &build.RepoName, &build.ProjectID, &build.Version, &build.Status,
        &build.BuildProvider, &build.ProviderRunID, &build.ProviderArtifactName,
        &build.ArtifactPath, &build.ArtifactSize, &build.ArtifactHash,
        &build.ErrorMessage, &build.StartedAt, &build.CompletedAt, &build.CreatedAt,
    )
    if err == sql.ErrNoRows { return nil, nil }
    if err != nil { return nil, err }
    return build, nil
}

func ListBuildsForProject(projectID int64) ([]*Build, error) {
    query := `SELECT id, build_type, repo_name, project_id, version, status, build_provider, provider_run_id, provider_artifact_name, artifact_path, artifact_size, artifact_hash, error_message, started_at, completed_at, created_at FROM build WHERE project_id = ? ORDER BY created_at DESC`
    rows, err := db.Query(query, projectID)
    if err != nil { return nil, err }
    defer rows.Close()

    var builds []*Build
    for rows.Next() {
        build := &Build{}
        err := rows.Scan(
            &build.ID, &build.BuildType, &build.RepoName, &build.ProjectID, &build.Version, &build.Status,
            &build.BuildProvider, &build.ProviderRunID, &build.ProviderArtifactName,
            &build.ArtifactPath, &build.ArtifactSize, &build.ArtifactHash,
            &build.ErrorMessage, &build.StartedAt, &build.CompletedAt, &build.CreatedAt,
        )
        if err != nil { return nil, err }
        builds = append(builds, build)
    }
    return builds, nil
}
