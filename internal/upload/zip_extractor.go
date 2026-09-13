package upload

import (
    "archive/zip"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "strings"
)

const (
    MaxZIPFileSize     = 50 * 1024 * 1024    // 50MB
    MaxSingleFile      = 50 * 1024 * 1024    // 50MB per file
    MaxExtractedTotal  = 200 * 1024 * 1024   // 200MB total
    MaxFileCount       = 5000
    MaxCompressionRatio = 100.0
)

type ExtractResult struct {
    FileCount  int
    TotalBytes int64
}

type ZipExtractor struct{}

func NewZipExtractor() *ZipExtractor {
    return &ZipExtractor{}
}

// Extract safely extracts a ZIP file to destDir
func (e *ZipExtractor) Extract(zipPath, destDir string) (*ExtractResult, error) {
    reader, err := zip.OpenReader(zipPath)
    if err != nil {
        return nil, fmt.Errorf("%w: %v", ErrInvalidZIP, err)
    }
    defer reader.Close()
    
    result := &ExtractResult{}
    var totalBytes int64
    
    for _, entry := range reader.File {
        // Check file count
        result.FileCount++
        if result.FileCount > MaxFileCount {
            return nil, ErrTooManyFiles
        }
        
        // Normalize path
        normalized, err := NormalizePath(entry.Name)
        if err != nil {
            return nil, err
        }
        
        // Check for symlink
        if entry.Mode()&os.ModeSymlink != 0 {
            return nil, fmt.Errorf("%w: %s", ErrSymlink, entry.Name)
        }
        
        // Check for nested archives (case-insensitive)
        lower := strings.ToLower(normalized)
        for _, ext := range []string{".zip", ".tar", ".gz", ".tgz", ".7z", ".rar"} {
            if strings.HasSuffix(lower, ext) {
                return nil, fmt.Errorf("%w: %s", ErrNestedArchive, entry.Name)
            }
        }
        
        // Check per-file size
        if entry.UncompressedSize64 > MaxSingleFile {
            return nil, fmt.Errorf("%w: %s (%d bytes)", ErrFileTooLarge, entry.Name, entry.UncompressedSize64)
        }
        
        // Check compression ratio (skip if compressed size is 0)
        if entry.CompressedSize64 > 0 {
            ratio := float64(entry.UncompressedSize64) / float64(entry.CompressedSize64)
            if ratio > MaxCompressionRatio {
                return nil, fmt.Errorf("%w: %s (ratio %.1f)", ErrCompressionBomb, entry.Name, ratio)
            }
        }
        
        // Check total extracted size
        totalBytes += int64(entry.UncompressedSize64)
        if totalBytes > MaxExtractedTotal {
            return nil, ErrTotalTooLarge
        }
        
        // Create destination path
        destPath := filepath.Join(destDir, filepath.FromSlash(normalized))
        
        // Ensure parent directory exists
        if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
            return nil, err
        }
        
        if entry.FileInfo().IsDir() {
            os.MkdirAll(destPath, 0755)
            continue
        }
        
        // Extract file
        src, err := entry.Open()
        if err != nil {
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
