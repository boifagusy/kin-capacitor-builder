package build

// nativeRuntimeGoMod is the go.mod for the Go native runtime module.
// Matches the proven cap-go-test configuration.
const nativeRuntimeGoMod = `module capgotest

go 1.26.0

tool golang.org/x/mobile/cmd/gobind

require (
	modernc.org/sqlite v1.57.0
	golang.org/x/mobile v0.0.0-20260908204917-8b95e45f8d3e // indirect
)
`

// nativeRuntimeCoreGo is the Go source code for the native runtime.
// Exposed to Java via gomobile bind as capgotest.AppCore.
const nativeRuntimeCoreGo = `package capgotest

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"

	_ "modernc.org/sqlite"
)

type AppCore struct {
	mu sync.Mutex
	db *sql.DB
}

type Note struct {
	ID        int64  ` + "`json:\"id\"`" + `
	Text      string ` + "`json:\"text\"`" + `
	CreatedAt string ` + "`json:\"created_at\"`" + `
}

func NewAppCore(dbPath string) (*AppCore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(` + "`CREATE TABLE IF NOT EXISTS notes (" + `
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		text TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)` + "`" + `)
	if err != nil {
		return nil, err
	}
	return &AppCore{db: db}, nil
}

func (a *AppCore) Health() string {
	if a == nil || a.db == nil {
		return "error"
	}
	if err := a.db.Ping(); err != nil {
		return "error: " + err.Error()
	}
	return "ok"
}

func (a *AppCore) SaveNote(text string) (int64, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.db == nil {
		return 0, fmt.Errorf("database not open")
	}
	res, err := a.db.Exec("INSERT INTO notes (text) VALUES (?)", text)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (a *AppCore) GetNotes() (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.db == nil {
		return "[]", fmt.Errorf("database not open")
	}
	rows, err := a.db.Query("SELECT id, text, created_at FROM notes ORDER BY id DESC")
	if err != nil {
		return "[]", err
	}
	defer rows.Close()
	notes := []Note{}
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.ID, &n.Text, &n.CreatedAt); err != nil {
			continue
		}
		notes = append(notes, n)
	}
	data, err := json.Marshal(notes)
	if err != nil {
		return "[]", err
	}
	return string(data), nil
}

func (a *AppCore) Stop() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.db != nil {
		_ = a.db.Close()
		a.db = nil
	}
}
`

// nativeRuntimeBuildScript runs gomobile bind for multi-ABI.
// Target: arm64-v8a + armeabi-v7a (any real Android phone).
const nativeRuntimeBuildScript = `#!/usr/bin/env bash
set -euo pipefail

echo "=== Building Go native AAR ==="

GOBIN="$(go env GOPATH)/bin"
export PATH="$GOBIN:$PATH"

if ! command -v gomobile >/dev/null 2>&1; then
  echo "Installing gomobile and gobind..."
  go install golang.org/x/mobile/cmd/gomobile@latest
  go install golang.org/x/mobile/cmd/gobind@latest
  export PATH="$GOBIN:$PATH"
fi

mkdir -p android/app/libs

cd go-core

echo "=== Resolving Go dependencies ==="
go mod tidy

echo "=== Running gomobile bind (arm64 + armv7) ==="
gomobile bind \
    -target=android/arm64,android/arm \
    -androidapi 23 \
    -o ../android/app/libs/runtime.aar \
    .

cd ..

echo "=== AAR Result ==="
ls -la android/app/libs/runtime.aar
`

// NativeRuntimeSourcesForRepo returns the files required by a project
// with RuntimeType == "go-native". They are pushed into the ephemeral
// GitHub build repository.
func NativeRuntimeSourcesForRepo() map[string]string {
	return map[string]string{
		"go-core/go.mod":  nativeRuntimeGoMod,
		"go-core/core.go": nativeRuntimeCoreGo,
		"build-go.sh":     nativeRuntimeBuildScript,
	}
}
