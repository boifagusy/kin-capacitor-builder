package build

import (
	"strings"
	"testing"
)

func TestNativeRuntimeSourcesForRepo_ExactKeys(t *testing.T) {
	files := NativeRuntimeSourcesForRepo()
	want := []string{"go-core/go.mod", "go-core/core.go", "build-go.sh"}
	if len(files) != len(want) {
		t.Fatalf("expected %d files, got %d", len(want), len(files))
	}
	for _, k := range want {
		if _, ok := files[k]; !ok {
			t.Fatalf("missing key: %s", k)
		}
	}
}

func TestNativeRuntimeSources_GoModHasGobindTool(t *testing.T) {
	files := NativeRuntimeSourcesForRepo()
	gomod := files["go-core/go.mod"]
	if !strings.Contains(gomod, "golang.org/x/mobile/cmd/gobind") {
		t.Fatal("go.mod missing gobind tool directive")
	}
	if !strings.Contains(gomod, "modernc.org/sqlite") {
		t.Fatal("go.mod missing modernc.org/sqlite")
	}
}

func TestNativeRuntimeSources_CoreHasAppCore(t *testing.T) {
	files := NativeRuntimeSourcesForRepo()
	core := files["go-core/core.go"]
	for _, want := range []string{
		"type AppCore struct",
		"func NewAppCore(",
		"func (a *AppCore) Health(",
		"func (a *AppCore) SaveNote(",
		"func (a *AppCore) GetNotes(",
		"CREATE TABLE IF NOT EXISTS notes",
	} {
		if !strings.Contains(core, want) {
			t.Fatalf("core.go missing: %s", want)
		}
	}
}

func TestNativeRuntimeSources_BuildScriptHasMultiABI(t *testing.T) {
	files := NativeRuntimeSourcesForRepo()
	script := files["build-go.sh"]
	if !strings.Contains(script, "gomobile bind") {
		t.Fatal("build-go.sh missing gomobile bind")
	}
	if !strings.Contains(script, "arm64") {
		t.Fatal("build-go.sh missing arm64 target")
	}
	if !strings.Contains(script, "android/arm") {
		t.Fatal("build-go.sh missing armeabi-v7a target")
	}
	if !strings.Contains(script, "go mod tidy") {
		t.Fatal("build-go.sh missing go mod tidy")
	}
}

func TestNativeRuntimeSources_NoLocalhost(t *testing.T) {
	files := NativeRuntimeSourcesForRepo()
	for name, content := range files {
		if strings.Contains(content, "127.0.0.1") {
			t.Fatalf("%s contains 127.0.0.1 — native runtime must not use localhost", name)
		}
	}
}
