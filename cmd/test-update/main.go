package main

import (
    "fmt"
    "local-apk-builder/internal/database"
)

func main() {
    database.Init("/data/data/com.termux/files/home/.local-apk-builder")
    
    p, err := database.GetProject(3)
    if err != nil || p == nil {
        fmt.Printf("Error: %v, p=%v\n", err, p)
        return
    }
    
    fmt.Printf("Before: source_type='%s', path='%s'\n", p.SourceType, p.SourcePath)
    
    p.SourceType = "upload"
    p.SourcePath = "/test/path"
    p.OriginalZipName = "test.zip"
    
    err = database.UpdateProject(p)
    if err != nil {
        fmt.Printf("Update error: %v\n", err)
        return
    }
    
    p2, _ := database.GetProject(3)
    fmt.Printf("After: source_type='%s', path='%s'\n", p2.SourceType, p2.SourcePath)
}
