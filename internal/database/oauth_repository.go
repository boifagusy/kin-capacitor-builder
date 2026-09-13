package database

import "database/sql"

type GitHubOAuthConfig struct {
    ID           int64  `json:"id"`
    ClientID     string `json:"client_id"`
    ClientSecret string `json:"client_secret"`
    CallbackURL  string `json:"callback_url"`
    UpdatedAt    string `json:"updated_at"`
}

// SaveGitHubOAuthConfig upserts OAuth config
func SaveGitHubOAuthConfig(cfg *GitHubOAuthConfig) error {
    query := `
    INSERT INTO github_oauth_config (client_id, client_secret, callback_url)
    VALUES (?, ?, ?)
    ON CONFLICT(id) DO UPDATE SET
        client_id = excluded.client_id,
        client_secret = excluded.client_secret,
        callback_url = excluded.callback_url,
        updated_at = CURRENT_TIMESTAMP
    `
    _, err := db.Exec(query, cfg.ClientID, cfg.ClientSecret, cfg.CallbackURL)
    return err
}

// GetGitHubOAuthConfig returns the OAuth config
func GetGitHubOAuthConfig() (*GitHubOAuthConfig, error) {
    query := `SELECT id, client_id, client_secret, callback_url, updated_at FROM github_oauth_config ORDER BY updated_at DESC LIMIT 1`
    cfg := &GitHubOAuthConfig{}
    err := db.QueryRow(query).Scan(&cfg.ID, &cfg.ClientID, &cfg.ClientSecret, &cfg.CallbackURL, &cfg.UpdatedAt)
    if err == sql.ErrNoRows {
        return nil, nil
    }
    if err != nil {
        return nil, err
    }
    return cfg, nil
}
