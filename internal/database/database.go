package database

import (
    "database/sql"
    "fmt"
    "os"
    "path/filepath"
    "sync"
    
    _ "modernc.org/sqlite"
)

var (
    db     *sql.DB
    once   sync.Once
    dbPath string
)

func Init(dataDir string) error {
    var initErr error
    
    once.Do(func() {
        if err := os.MkdirAll(dataDir, 0755); err != nil {
            initErr = fmt.Errorf("failed to create data directory: %w", err)
            return
        }
        
        dbPath = filepath.Join(dataDir, "builder.db")
        
        db, initErr = sql.Open("sqlite", dbPath)
        if initErr != nil {
            initErr = fmt.Errorf("failed to open database: %w", initErr)
            return
        }
        
        db.SetMaxOpenConns(1)
        db.SetMaxIdleConns(1)
        
        if initErr = db.Ping(); initErr != nil {
            initErr = fmt.Errorf("failed to ping database: %w", initErr)
            return
        }
        
        if initErr = migrate(db); initErr != nil {
            initErr = fmt.Errorf("failed to migrate database: %w", initErr)
            return
        }
    })
    
    return initErr
}

func GetDB() *sql.DB {
    return db
}

func Close() error {
    if db != nil {
        return db.Close()
    }
    return nil
}
