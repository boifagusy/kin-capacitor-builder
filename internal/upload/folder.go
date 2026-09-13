package upload

import (
    "fmt"
    "io"
    "mime/multipart"
    "os"
    "path/filepath"
)

// SaveFolderUpload writes uploaded files to destDir with path normalization
func SaveFolderUpload(files []*multipart.FileHeader, destDir string) (*ExtractResult, error) {
    result := &ExtractResult{}
    var totalBytes int64
    
    for _, fh := range files {
        result.FileCount++
        if result.FileCount > MaxFileCount {
            return nil, ErrTooManyFiles
        }
        
        // Normalize the filename
        normalized, err := NormalizePath(fh.Filename)
        if err != nil {
            return nil, err
        }
        
        // Check file size
        if fh.Size > MaxSingleFile {
            return nil, fmt.Errorf("%w: %s (%d bytes)", ErrFileTooLarge, normalized, fh.Size)
        }
        
        totalBytes += fh.Size
        if totalBytes > MaxExtractedTotal {
            return nil, ErrTotalTooLarge
        }
        
        // Open source
        src, err := fh.Open()
        if err != nil {
            return nil, err
        }
        
        // Create destination
        destPath := filepath.Join(destDir, filepath.FromSlash(normalized))
        if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
            src.Close()
            return nil, err
        }
        
        dst, err := os.Create(destPath)
        if err != nil {
            src.Close()
            return nil, err
        }
        
        _, err = io.Copy(dst, src)
        src.Close()
        dst.Close()
        if err != nil {
            return nil, err
        }
    }
    
    result.TotalBytes = totalBytes
    return result, nil
}
