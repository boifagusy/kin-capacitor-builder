package build

import (
	"strings"

	"local-apk-builder/internal/runtime"
)

// runtimeGoModTemplate matches the proven gomobile toolchain configuration
// from the Gate 2A probe. Do not modify without re-verifying gomobile bind.
const runtimeGoModTemplate = `module runtime

go 1.26.0

tool golang.org/x/mobile/cmd/gobind

require (
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/ncruces/go-strftime v1.0.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	golang.org/x/mobile v0.0.0-20260908204917-8b95e45f8d3e // indirect
	golang.org/x/mod v0.41.0 // indirect
	golang.org/x/sync v0.23.0 // indirect
	golang.org/x/sys v0.48.0 // indirect
	golang.org/x/tools v0.50.0 // indirect
	modernc.org/libc v1.74.4 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.11.0 // indirect
	modernc.org/sqlite v1.57.0 // indirect
)
`

// runtimeBuildScriptTemplate matches the proven gomobile bind workflow from
// the Gate 2A probe. It is intentionally minimal and network-free.
const runtimeBuildScriptTemplate = `#!/usr/bin/env bash
set -euo pipefail

echo "=== Building Go runtime AAR via gomobile bind ==="

GOBIN="$(go env GOPATH)/bin"
export PATH="$GOBIN:$PATH"

if ! command -v gomobile >/dev/null 2>&1; then
  echo "Installing gomobile and gobind..."
  go install golang.org/x/mobile/cmd/gomobile@latest
  go install golang.org/x/mobile/cmd/gobind@latest
  export PATH="$GOBIN:$PATH"
fi

mkdir -p app/libs

cd go-src
gomobile bind -target=android -androidapi 23 -o ../app/libs/runtime.aar .
cd ..

echo "=== runtime.aar ==="
ls -la app/libs/
`

// RuntimeSourcesForRepo returns the four files that must be pushed into the
// ephemeral GitHub build repository when a project has RuntimeType "go-sqlite".
//
// The runtime package source files have their package declaration transformed
// from "package runtime" to "package main" using a single-replacement
// strings.Replace. Any other content is passed through unmodified.
func RuntimeSourcesForRepo() map[string]string {
	mainSrc := strings.Replace(runtime.SourceServerGo, "package runtime", "package main", 1)
	staticSrc := strings.Replace(runtime.SourceStaticGo, "package runtime", "package main", 1)

	return map[string]string{
		"go-src/main.go":        mainSrc,
		"go-src/static_html.go": staticSrc,
		"go-src/go.mod":         runtimeGoModTemplate,
		"build-go.sh":           runtimeBuildScriptTemplate,
	}
}
