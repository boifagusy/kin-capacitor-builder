package runtime

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEmbed_SourceServerGoNonEmpty(t *testing.T) {
	if SourceServerGo == "" {
		t.Fatal("SourceServerGo is empty")
	}
	if !strings.HasPrefix(SourceServerGo, "package runtime") {
		t.Fatalf("SourceServerGo does not begin with package runtime")
	}
}

func TestEmbed_SourceStaticGoNonEmpty(t *testing.T) {
	if SourceStaticGo == "" {
		t.Fatal("SourceStaticGo is empty")
	}
	if !strings.HasPrefix(SourceStaticGo, "package runtime") {
		t.Fatalf("SourceStaticGo does not begin with package runtime")
	}
}

// TestEmbed_AllRuntimeFilesEmbedded enumerates every non-test .go file in
// this package directory and asserts that each is either embedded or
// explicitly excluded. If a new runtime source file is added without
// being embedded, this test fails loudly.
func TestEmbed_AllRuntimeFilesEmbedded(t *testing.T) {
	embedded := map[string]bool{
		"server.go": true,
		"static.go": true,
	}
	excluded := map[string]bool{
		"embed.go":      true,
		"embed_test.go": true,
	}
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() {
			continue
		}
		if !strings.HasSuffix(name, ".go") {
			continue
		}
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		if excluded[name] {
			continue
		}
		if embedded[name] {
			continue
		}
		abs, _ := filepath.Abs(name)
		t.Fatalf("runtime source file not embedded: %s (abs: %s)", name, abs)
	}
}
