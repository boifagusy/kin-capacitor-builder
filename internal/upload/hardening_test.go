package upload

import (
"archive/zip"
"bytes"
"errors"
"fmt"
"os"
"path/filepath"
"testing"
)

func buildZIP(t *testing.T, entries map[string][]byte) []byte {
var buf bytes.Buffer
zw := zip.NewWriter(&buf)
for name, data := range entries {
w, _ := zw.Create(name)
w.Write(data)
}
zw.Close()
return buf.Bytes()
}
func saveZIP(t *testing.T, data []byte) string {
zipPath := filepath.Join(t.TempDir(), "test.zip")
os.WriteFile(zipPath, data, 0644)
return zipPath
}

func assertExtractError(t *testing.T, zipPath string, want error) {
ext := NewZipExtractor()
_, err := ext.Extract(zipPath, t.TempDir())
if err == nil {
t.Fatal("expected error, got nil")
}
if !errors.Is(err, want) {
t.Fatalf("expected %v, got %v", want, err)
}
}

func TestHardening_NestedArchiveExtensions(t *testing.T) {
for _, ext := range []string{".tar", ".gz", ".tgz", ".7z", ".rar"} {
zipPath := saveZIP(t, buildZIP(t, map[string][]byte{"inner" + ext: []byte("nested")}))
assertExtractError(t, zipPath, ErrNestedArchive)
}
}

func TestHardening_NestedArchiveCaseInsensitive(t *testing.T) {
for _, name := range []string{"INNER.ZIP", "Inner.Zip", "inner.zip"} {
zipPath := saveZIP(t, buildZIP(t, map[string][]byte{name: []byte("nested")}))
assertExtractError(t, zipPath, ErrNestedArchive)
}
}

func TestHardening_Symlink(t *testing.T) {
var buf bytes.Buffer
zw := zip.NewWriter(&buf)
hdr := &zip.FileHeader{Name: "link"}
hdr.SetMode(os.ModeSymlink | 0777)
f, _ := zw.CreateHeader(hdr)
f.Write([]byte("../../../etc/passwd"))
zw.Close()
zipPath := saveZIP(t, buf.Bytes())
assertExtractError(t, zipPath, ErrSymlink)
}

func TestHardening_CompressionBomb(t *testing.T) {
big := make([]byte, 1024*1024)
zipPath := saveZIP(t, buildZIP(t, map[string][]byte{"bomb.bin": big}))
assertExtractError(t, zipPath, ErrCompressionBomb)
}

func TestHardening_FileCountLimit(t *testing.T) {
entries := make(map[string][]byte, MaxFileCount+10)
for i := 0; i < MaxFileCount+10; i++ {
entries[fmt.Sprintf("file%d.txt", i)] = []byte("x")
}
zipPath := saveZIP(t, buildZIP(t, entries))
assertExtractError(t, zipPath, ErrTooManyFiles)
}

func TestHardening_MalformedZIP(t *testing.T) {
zipPath := filepath.Join(t.TempDir(), "bad.zip")
os.WriteFile(zipPath, []byte("not a real zip file"), 0644)
ext := NewZipExtractor()
_, err := ext.Extract(zipPath, t.TempDir())
if err == nil {
t.Fatal("expected error for malformed ZIP, got nil")
}
if !errors.Is(err, ErrInvalidZIP) {
t.Fatal("expected ErrInvalidZIP")
}
}

func TestHardening_EmptyZIP(t *testing.T) {
var buf bytes.Buffer
zw := zip.NewWriter(&buf)
zw.Close()
zipPath := saveZIP(t, buf.Bytes())
ext := NewZipExtractor()
result, err := ext.Extract(zipPath, t.TempDir())
if err != nil {
t.Fatal("expected nil error for empty zip")
}
if result.FileCount != 0 {
t.Fatal("expected 0 files")
}
}

func TestHardening_TotalSizeTooLarge(t *testing.T) {
entries := map[string][]byte{}
for i := 0; i < 5; i++ {
chunk := make([]byte, 45*1024*1024)
seed := uint32(i + 1)
for j := range chunk {
seed = seed*1103515245 + 12345
chunk[j] = byte(seed >> 16)
}
entries[fmt.Sprintf("big%d.bin", i)] = chunk
}
zipPath := saveZIP(t, buildZIP(t, entries))
assertExtractError(t, zipPath, ErrTotalTooLarge)
}
