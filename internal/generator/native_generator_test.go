package generator

import (
	"strings"
	"testing"

	"local-apk-builder/internal/database"
)

func newTestGenerator() *Generator {
	return &Generator{
		project: &database.Project{
			ID:          1,
			AppName:     "Test App",
			RuntimeType: "go-native",
		},
	}
}

func TestNativeMainActivity_ProducesBridgeActivity(t *testing.T) {
	g := newTestGenerator()
	code := g.GenerateNativeMainActivityJava()
	if !strings.Contains(code, "extends BridgeActivity") {
		t.Fatal("expected BridgeActivity subclass")
	}
	if !strings.Contains(code, "registerPlugin(AppBridgePlugin.class)") {
		t.Fatal("expected plugin registration")
	}
}

func TestNativeAppBridgePlugin_ExposesThreeMethods(t *testing.T) {
	g := newTestGenerator()
	code := g.GenerateNativeAppBridgePlugin()
	for _, m := range []string{"health", "saveNote", "getNotes"} {
		if !strings.Contains(code, m) {
			t.Fatalf("missing method: %s", m)
		}
	}
	if !strings.Contains(code, "@CapacitorPlugin(name = \"AppBridge\")") {
		t.Fatal("expected @CapacitorPlugin annotation")
	}
}

func TestNativeAppBridgeJS_ExposesWindowAppBridge(t *testing.T) {
	g := newTestGenerator()
	code := g.GenerateNativeAppBridgeJS()
	if !strings.Contains(code, "window.AppBridge") {
		t.Fatal("expected window.AppBridge")
	}
	for _, m := range []string{"health", "saveNote", "getNotes"} {
		if !strings.Contains(code, m) {
			t.Fatalf("missing JS method: %s", m)
		}
	}
	if strings.Contains(code, "http://127.0.0.1") {
		t.Fatal("native runtime must not use localhost")
	}
}
