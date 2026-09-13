package generator

import (
"testing"
)


func TestHardening_MissingIndexHTML(t *testing.T) {
dir := createTestProject(t, map[string]string{"style.css": "body{}"})
d := &ProjectDetector{SourceDir: dir}
profile, err := d.Detect()
if err != nil {
t.Fatalf("Detect failed: %v", err)
}
if profile.Type != "unknown" {
t.Fatalf("expected unknown, got %s", profile.Type)
}
}

func TestHardening_OnlyHiddenFiles(t *testing.T) {
dir := createTestProject(t, map[string]string{".gitignore": "node_modules/"})
d := &ProjectDetector{SourceDir: dir}
profile, err := d.Detect()
if err != nil {
t.Fatalf("Detect failed: %v", err)
}
if profile.Type != "unknown" {
t.Fatalf("expected unknown, got %s", profile.Type)
}
}

func TestHardening_InvalidPackageJSON(t *testing.T) {
dir := createTestProject(t, map[string]string{"package.json": "{ not valid json"})
d := &ProjectDetector{SourceDir: dir}
_, err := d.Detect()
if err == nil {
t.Fatal("expected error for invalid package.json")
}
}

func TestHardening_NodeMissingBuildScript(t *testing.T) {
dir := createTestProject(t, map[string]string{"package.json": "{\"scripts\": {\"dev\": \"vite\"}}"})
d := &ProjectDetector{SourceDir: dir}
_, err := d.Detect()
if err == nil {
t.Fatal("expected error for missing build script")
}
}

func TestHardening_NodeWithoutVite(t *testing.T) {
dir := createTestProject(t, map[string]string{"package.json": "{\"scripts\": {\"build\": \"webpack\"}}"})
d := &ProjectDetector{SourceDir: dir}
profile, err := d.Detect()
if err != nil {
t.Fatalf("Detect failed: %v", err)
}
if profile.Type != "node" {
t.Fatalf("expected node, got %s", profile.Type)
}
if profile.Builder != "none" {
t.Fatalf("expected none builder, got %s", profile.Builder)
}
}

func TestHardening_ReactDevDep(t *testing.T) {
dir := createTestProject(t, map[string]string{"package.json": "{\"devDependencies\": {\"react\": \"^18\"}, \"scripts\": {\"build\": \"vite\"}}", "vite.config.js": ""})
d := &ProjectDetector{SourceDir: dir}
profile, err := d.Detect()
if err != nil {
t.Fatalf("Detect failed: %v", err)
}
if profile.Framework != "react" {
t.Fatalf("expected react, got %s", profile.Framework)
}
}

func TestHardening_VueDevDep(t *testing.T) {
dir := createTestProject(t, map[string]string{"package.json": "{\"devDependencies\": {\"vue\": \"^3\"}, \"scripts\": {\"build\": \"vite\"}}", "vite.config.js": ""})
d := &ProjectDetector{SourceDir: dir}
profile, err := d.Detect()
if err != nil {
t.Fatalf("Detect failed: %v", err)
}
if profile.Framework != "vue" {
t.Fatalf("expected vue, got %s", profile.Framework)
}
}

func TestHardening_WrapperSingleFile(t *testing.T) {
dir := createTestProject(t, map[string]string{"only.txt": "hi"})
d := &ProjectDetector{SourceDir: dir}
profile, err := d.Detect()
if err != nil {
t.Fatalf("Detect failed: %v", err)
}
if profile.Type != "unknown" {
t.Fatalf("expected unknown, got %s", profile.Type)
}
}

func TestHardening_WrapperTwoDirs(t *testing.T) {
dir := createTestProject(t, map[string]string{"a/index.html": "<html>A</html>", "b/index.html": "<html>B</html>"})
d := &ProjectDetector{SourceDir: dir}
profile, err := d.Detect()
if err != nil {
t.Fatalf("Detect failed: %v", err)
}
if profile.Type != "unknown" {
t.Fatalf("expected unknown for ambiguous, got %s", profile.Type)
}
}

func TestHardening_WrapperWithHidden(t *testing.T) {
dir := createTestProject(t, map[string]string{"app/index.html": "<html>App</html>", ".gitignore": "node_modules/"})
d := &ProjectDetector{SourceDir: dir}
profile, err := d.Detect()
if err != nil {
t.Fatalf("Detect failed: %v", err)
}
if profile.Type != "static" {
t.Fatalf("expected static via wrapper, got %s", profile.Type)
}
}
