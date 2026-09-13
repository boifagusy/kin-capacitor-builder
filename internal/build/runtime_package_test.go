package build

import (
	"strings"
	"testing"
)

func TestRuntimeSourcesForRepo_ExactKeys(t *testing.T) {
	files := RuntimeSourcesForRepo()
	want := []string{"go-src/main.go", "go-src/static_html.go", "go-src/go.mod", "build-go.sh"}
	if len(files) != len(want) {
		t.Fatalf("expected %d files, got %d", len(want), len(files))
	}
	for _, k := range want {
		if _, ok := files[k]; !ok {
			t.Fatalf("missing key: %s", k)
		}
	}
}

func TestRuntimeSourcesForRepo_PackageMain(t *testing.T) {
	files := RuntimeSourcesForRepo()
	mainSrc := files["go-src/main.go"]
	staticSrc := files["go-src/static_html.go"]
	if !strings.HasPrefix(mainSrc, "package main") {
		t.Fatalf("main.go does not start with package main")
	}
	if !strings.HasPrefix(staticSrc, "package main") {
		t.Fatalf("static_html.go does not start with package main")
	}
}

func TestRuntimeSourcesForRepo_NoRuntimePackage(t *testing.T) {
	files := RuntimeSourcesForRepo()
	mainSrc := files["go-src/main.go"]
	staticSrc := files["go-src/static_html.go"]
	if strings.Contains(mainSrc, "package runtime") {
		t.Fatalf("main.go still contains package runtime")
	}
	if strings.Contains(staticSrc, "package runtime") {
		t.Fatalf("static_html.go still contains package runtime")
	}
}

func TestRuntimeSourcesForRepo_GoModHasCorrectToolchain(t *testing.T) {
	files := RuntimeSourcesForRepo()
	gomod := files["go-src/go.mod"]
	if !strings.Contains(gomod, "go 1.26.0") {
		t.Fatalf("go.mod missing go 1.26.0")
	}
	if !strings.Contains(gomod, "modernc.org/sqlite") {
		t.Fatalf("go.mod missing modernc.org/sqlite")
	}
	if !strings.Contains(gomod, "golang.org/x/mobile/cmd/gobind") {
		t.Fatalf("go.mod missing gobind tool directive")
	}
}

func TestRuntimeSourcesForRepo_BuildScriptHasGomobile(t *testing.T) {
	files := RuntimeSourcesForRepo()
	script := files["build-go.sh"]
	if !strings.Contains(script, "gomobile bind") {
		t.Fatalf("build-go.sh missing gomobile bind")
	}
	if !strings.Contains(script, "-target=android") {
		t.Fatalf("build-go.sh missing android target")
	}
	if !strings.Contains(script, "-androidapi") {
		t.Fatalf("build-go.sh missing androidapi flag")
	}
}

func TestRuntimeSourcesForRepo_NoBuilderImports(t *testing.T) {
	files := RuntimeSourcesForRepo()
	for name, content := range files {
		if strings.Contains(content, "local-apk-builder/") {
			t.Fatalf("%s imports local-apk-builder/* — runtime must be self-contained", name)
		}
	}
}
