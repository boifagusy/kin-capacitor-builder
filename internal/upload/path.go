package upload

import (
    "fmt"
    "path/filepath"
    "strings"
)

// NormalizePath converts user-provided relative paths to safe workspace-relative paths
func NormalizePath(rawPath string) (string, error) {
    // Replace Windows backslashes with forward slashes
    p := strings.ReplaceAll(rawPath, "\\", "/")
    
    // Reject Windows drive prefixes (case-insensitive)
    if len(p) >= 2 && p[1] == ':' {
        return "", fmt.Errorf("%w: drive prefix not allowed", ErrAbsolutePath)
    }
    
    // Reject leading slash (absolute path)
    if strings.HasPrefix(p, "/") {
        return "", fmt.Errorf("%w: leading slash not allowed", ErrAbsolutePath)
    }
    
    // Split into parts and validate
    parts := strings.Split(p, "/")
    clean := make([]string, 0, len(parts))
    for _, part := range parts {
        switch part {
        case "", ".":
            continue
        case "..":
            return "", fmt.Errorf("%w: '..' not allowed", ErrPathTraversal)
        default:
            clean = append(clean, part)
        }
    }
    
    result := filepath.ToSlash(filepath.Join(clean...))
    if result == "" || result == "." {
        return "", fmt.Errorf("%w: empty path", ErrInvalidPath)
    }
    
    return result, nil
}
