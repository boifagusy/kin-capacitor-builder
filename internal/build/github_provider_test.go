package build

import (
    "archive/zip"
    "bytes"
    "testing"
)

func TestExtractAPK(t *testing.T) {
    // Create a ZIP with a fake APK
    var buf bytes.Buffer
    zw := zip.NewWriter(&buf)

    apkContent := []byte("fake-apk-content")
    f, err := zw.Create("app-debug.apk")
    if err != nil {
        t.Fatal(err)
    }
    if _, err := f.Write(apkContent); err != nil {
        t.Fatal(err)
    }
    zw.Close()

    provider := NewGitHubActionsProvider("", "", "", "local-apk-builder-debug")
    path, size, hash, err := provider.ExtractAPK(buf.Bytes())
    if err != nil {
        t.Fatalf("ExtractAPK failed: %v", err)
    }
    if string(path) != string(apkContent) {
        t.Errorf("unexpected content: %s", path)
    }
    if size != int64(len(apkContent)) {
        t.Errorf("unexpected size: %d", size)
    }
    if hash == "" {
        t.Error("hash is empty")
    }
    t.Logf("path=%s size=%d hash=%s", path, size, hash)
}

func TestExtractAPKNoAPK(t *testing.T) {
    var buf bytes.Buffer
    zw := zip.NewWriter(&buf)
    f, _ := zw.Create("readme.txt")
    f.Write([]byte("hello"))
    zw.Close()

    provider := NewGitHubActionsProvider("", "", "", "local-apk-builder-debug")
    _, _, _, err := provider.ExtractAPK(buf.Bytes())
    if err == nil {
        t.Fatal("expected error for missing APK")
    }
}

func TestRunStatusToString(t *testing.T) {
    tests := []struct {
        run    map[string]interface{}
        expect string
    }{
        {map[string]interface{}{"status": "queued"}, StatusQueued},
        {map[string]interface{}{"status": "in_progress"}, StatusBuilding},
        {map[string]interface{}{"status": "completed", "conclusion": "success"}, StatusSuccess},
        {map[string]interface{}{"status": "completed", "conclusion": "failure"}, StatusFailed},
        {map[string]interface{}{"status": "completed", "conclusion": "cancelled"}, StatusFailed},
    }
    for _, tt := range tests {
        got := runStatusToString(tt.run)
        if got != tt.expect {
            t.Errorf("expected %s, got %s", tt.expect, got)
        }
    }
}
