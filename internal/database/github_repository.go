package database

import "database/sql"

type GitHubConnection struct {
    ID             int64  `json:"id"`
    GitHubUsername string `json:"github_username"`
    AccessToken    string `json:"access_token"`
    CreatedAt      string `json:"created_at"`
    UpdatedAt      string `json:"updated_at"`
}

// SaveGitHubConnection upserts a github connection
func SaveGitHubConnection(conn *GitHubConnection) error {
    // Delete existing and insert fresh
    _, err := db.Exec("DELETE FROM github_connection WHERE github_username = ?", conn.GitHubUsername)
    if err != nil {
        return err
    }
    _, err = db.Exec("INSERT INTO github_connection (github_username, access_token) VALUES (?, ?)", conn.GitHubUsername, conn.AccessToken)
    return err
}

// GetGitHubConnection returns the most recent connection
func GetGitHubConnection() (*GitHubConnection, error) {
    query := `SELECT id, github_username, access_token, created_at, updated_at FROM github_connection ORDER BY updated_at DESC LIMIT 1`
    conn := &GitHubConnection{}
    err := db.QueryRow(query).Scan(&conn.ID, &conn.GitHubUsername, &conn.AccessToken, &conn.CreatedAt, &conn.UpdatedAt)
    if err == sql.ErrNoRows {
        return nil, nil
    }
    if err != nil {
        return nil, err
    }
    return conn, nil
}


// UpdateAccessToken updates just the access token
func UpdateAccessToken(token string) error {
    database := GetDB()
    // Update the most recent connection
    _, err := database.Exec("UPDATE github_connection SET access_token = ?, updated_at = CURRENT_TIMESTAMP WHERE id = (SELECT id FROM github_connection ORDER BY updated_at DESC LIMIT 1)", token)
    return err
}

// UpdateAccessTokenAndUsername updates both the token and username of the most recent connection.
func UpdateAccessTokenAndUsername(token, username string) error {
	db := GetDB()
	_, err := db.Exec(
		"UPDATE github_connection SET access_token = ?, github_username = ?, updated_at = CURRENT_TIMESTAMP WHERE id = (SELECT id FROM github_connection ORDER BY updated_at DESC LIMIT 1)",
		token, username,
	)
	return err
}
