package build

import (
    "os"
    "path/filepath"
    "strings"
)

// SourcePayload describes upload source for GitHub push
type SourcePayload struct {
    HasSource    bool   // true for upload projects
    SourceDir    string // server-side path to source workspace
    Profile      string // "static" | "node"
    BuildSystem  string // "vite" | "none"
    Framework    string // "react" | "vue" | "none"
    OutputDir    string // "dist" or "" for static
}

// CollectSourceFiles walks the source directory and returns relative path → content map
func CollectSourceFiles(sourceDir string) (map[string]string, error) {
    files := make(map[string]string)
    
    err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if info.IsDir() {
            return nil
        }
        
        rel, err := filepath.Rel(sourceDir, path)
        if err != nil {
            return err
        }
        
        // Skip node_modules and other heavy dirs
        rel = filepath.ToSlash(rel)
        if strings.HasPrefix(rel, "node_modules/") ||
           strings.HasPrefix(rel, ".git/") ||
           strings.HasPrefix(rel, "dist/") ||
           strings.HasPrefix(rel, "build/") {
            return nil
        }
        
        data, err := os.ReadFile(path)
        if err != nil {
            return err
        }
        
        // Skip binary files > 1MB
        if len(data) > 1024*1024 {
            return nil
        }
        
        files[rel] = string(data)
        return nil
    })
    
    return files, err
}
