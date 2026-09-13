package build

import (
"strings"
"testing"
)


func TestHardening_WorkflowYAML_StaticEnv(t *testing.T) {
params := WorkflowParams{SourceType: "upload", ProjectType: "static", BuildSystem: "none", NeedsBuild: false, SourceDir: ".", WebOutputDir: "www"}
yaml := GenerateWorkflowYAML(params, false)
if !strings.Contains(yaml, "NEEDS_BUILD: 'false'") {
t.Fatalf("expected NEEDS_BUILD false in static YAML")
}
if !strings.Contains(yaml, "PROJECT_TYPE: 'static'") {
t.Fatalf("expected PROJECT_TYPE static")
}
}

func TestHardening_WorkflowYAML_NodeEnv(t *testing.T) {
params := WorkflowParams{SourceType: "upload", ProjectType: "node", BuildSystem: "vite", NeedsBuild: true, SourceDir: "source", WebOutputDir: "www"}
yaml := GenerateWorkflowYAML(params, false)
if !strings.Contains(yaml, "NEEDS_BUILD: 'true'") {
t.Fatalf("expected NEEDS_BUILD true")
}
if !strings.Contains(yaml, "SOURCE_DIR: 'source'") {
t.Fatalf("expected SOURCE_DIR source")
}
if !strings.Contains(yaml, "BUILD_SYSTEM: 'vite'") {
t.Fatalf("expected BUILD_SYSTEM vite")
}
}

func TestHardening_WorkflowYAML_APKVsAAB(t *testing.T) {
p := WorkflowParams{SourceType: "upload", ProjectType: "static", SourceDir: ".", WebOutputDir: "www"}
apk := GenerateWorkflowYAML(p, false)
aab := GenerateWorkflowYAML(p, true)
if !strings.Contains(apk, "assembleRelease") {
t.Fatalf("APK YAML missing assembleRelease")
}
if !strings.Contains(aab, "bundleRelease") {
t.Fatalf("AAB YAML missing bundleRelease")
}
if !strings.Contains(apk, "Build Android APK") {
t.Fatalf("APK workflow name missing")
}
if !strings.Contains(aab, "Build Android AAB") {
t.Fatalf("AAB workflow name missing")
}
}

func TestHardening_WorkflowYAML_ArtifactName(t *testing.T) {
p := WorkflowParams{SourceType: "upload", ProjectType: "static", SourceDir: ".", WebOutputDir: "www"}
apk := GenerateWorkflowYAML(p, false)
aab := GenerateWorkflowYAML(p, true)
if !strings.Contains(apk, "local-apk-builder-debug") {
t.Fatalf("APK artifact name missing")
}
if !strings.Contains(aab, "local-apk-builder-aab") {
t.Fatalf("AAB artifact name missing")
}
}

func TestHardening_WorkflowYAML_EnvValuesSubstituted(t *testing.T) {
p := WorkflowParams{SourceType: "upload", ProjectType: "node", BuildSystem: "vite", NeedsBuild: true, SourceDir: "source", WebOutputDir: "www"}
yaml := GenerateWorkflowYAML(p, false)
if strings.Contains(yaml, "%s") {
t.Fatalf("template %%s marker left unsubstituted")
}
if !strings.Contains(yaml, "SOURCE_TYPE: 'upload'") {
t.Fatalf("SOURCE_TYPE not substituted")
}
}

func TestHardening_WorkflowYAML_HasNpmBuildStep(t *testing.T) {
p := WorkflowParams{SourceType: "upload", ProjectType: "node", BuildSystem: "vite", NeedsBuild: true, SourceDir: "source", WebOutputDir: "www"}
yaml := GenerateWorkflowYAML(p, false)
if !strings.Contains(yaml, "npm install") {
t.Fatalf("npm install missing")
}
if !strings.Contains(yaml, "npm run build") {
t.Fatalf("npm run build missing")
}
if !strings.Contains(yaml, "NEEDS_BUILD == 'true'") {
t.Fatalf("conditional guard missing")
}
}

func TestHardening_isAABtoExt(t *testing.T) {
if isAABtoExt(true) != "aab" {
t.Fatalf("expected aab, got %s", isAABtoExt(true))
}
if isAABtoExt(false) != "apk" {
t.Fatalf("expected apk, got %s", isAABtoExt(false))
}
}

func TestHardening_isAABtoGradle(t *testing.T) {
if isAABtoGradle(true) != "./gradlew bundleRelease" {
t.Fatalf("expected bundleRelease, got %s", isAABtoGradle(true))
}
if isAABtoGradle(false) != "./gradlew assembleRelease" {
t.Fatalf("expected assembleRelease, got %s", isAABtoGradle(false))
}
}
