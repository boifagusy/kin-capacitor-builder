package build

import (
    "path/filepath"
    "archive/zip"
    "bytes"
    "context"
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "io"
    "log"
    "strings"
    "time"

    "local-apk-builder/internal/build/github"
    "local-apk-builder/internal/database"
    "local-apk-builder/internal/generator"
)

const workflowYAML = `name: Build Android APK

on:
  workflow_dispatch:
    inputs:
      build_id:
        description: 'Build ID'
        required: true

permissions:
  contents: read

env:
  SOURCE_TYPE: 'url'
  PROJECT_TYPE: 'static'
  BUILD_SYSTEM: 'none'
  NEEDS_BUILD: 'false'
  SOURCE_DIR: '.'
  WEB_OUTPUT_DIR: 'www'

jobs:
  build:
    name: Build Debug APK
    runs-on: ubuntu-latest

    steps:
      - name: Checkout repository
        uses: actions/checkout@v4

      - name: Write build payload to file
        env:
          BUILD_ID: ${{ github.event.inputs.build_id }}
        run: |
          echo "Build ID: $BUILD_ID"
          echo "$BUILD_ID" > /tmp/build_id.txt

      - name: Setup Node.js 22
        uses: actions/setup-node@v4
        with:
          node-version: '22'

      - name: Build frontend (Node projects only)
        if: ${{ env.NEEDS_BUILD == 'true' }}
        run: |
          set -e
          echo "=== Frontend build ==="
          echo "Project type: $PROJECT_TYPE"
          echo "Build system: $BUILD_SYSTEM"
          echo "Working dir: $SOURCE_DIR"
          
          cd "$SOURCE_DIR"
          
          echo "Installing dependencies..."
          npm install --legacy-peer-deps
          
          echo "Running build..."
          npm run build
          
          echo "Checking build output..."
          if [ ! -d "dist" ]; then
            echo "ERROR: dist/ not found after build"
            ls -la
            exit 1
          fi
          
          echo "Copying dist/ to www/"
          mkdir -p ../www
          cp -r dist/* ../www/
          
          echo "Build output copied successfully"

      - name: Install dependencies and add Android platform
        run: |
          set -e
          echo "Step 1: Installing npm dependencies..."
          npm install --legacy-peer-deps
          echo "npm install succeeded"
          
          echo "Step 2: Checking for android directory..."
          if [ ! -d android ]; then
            echo "android/ directory not found. Running: npx cap add android"
            npx cap add android || { echo "ERROR: npx cap add android failed"; exit 1; }
            echo "Successfully created android/ directory"
          else
            echo "android/ directory already exists"
          fi
          
          echo "Step 3: Running npx cap sync android..."
          npx cap sync android || { echo "ERROR: npx cap sync android failed"; exit 1; }
          echo "Android platform setup complete"

          echo "Injecting native navigation files..."
          mkdir -p android/app/src/main/assets
          cp android-custom/app_config.json android/app/src/main/assets/app_config.json
          # MainActivity left as Capacitor default - no replacement

          # Decode app tray icon from base64
          if [ -f android-custom/icon_base64.txt ]; then
            mkdir -p assets
            base64 -d android-custom/icon_base64.txt > assets/icon.png
            echo "App tray icon decoded to assets/icon.png"
          fi

          echo "Generating launcher icons with @capacitor/assets..."
          npx capacitor-assets generate --android

          if [ -f android-custom/google-services.json ]; then
            cp android-custom/google-services.json android/app/google-services.json
          fi

          if [ -f android-custom/splash.png ]; then
            for d in drawable drawable-port-mdpi drawable-port-hdpi drawable-port-xhdpi drawable-port-xxhdpi drawable-port-xxxhdpi drawable-land-mdpi drawable-land-hdpi drawable-land-xhdpi drawable-land-xxhdpi drawable-land-xxxhdpi; do
              mkdir -p android/app/src/main/res/$d
              cp android-custom/splash.png android/app/src/main/res/$d/splash.png
            done
          fi


      - name: Setup Java 21
        uses: actions/setup-java@v4
        with:
          distribution: 'temurin'
          java-version: '21'

      - name: Setup Android SDK
        uses: android-actions/setup-android@v3

      - name: Build release APK
        working-directory: android
        run: |
          chmod +x gradlew
          ./gradlew assembleRelease --no-daemon

      - name: Verify APK exists
        working-directory: android
        run: |
          echo "Searching for APK..."
          find app/build/outputs/apk -name "*.apk" 2>/dev/null || echo "No APK directory"
          APK_PATH=""
          if [ -f "app/build/outputs/apk/release/app-release.apk" ]; then
            APK_PATH="app/build/outputs/apk/release/app-release.apk"
          elif [ -f "app/build/outputs/apk/release/app-release-unsigned.apk" ]; then
            APK_PATH="app/build/outputs/apk/release/app-release-unsigned.apk"
          elif [ -f "app/build/outputs/apk/debug/app-debug.apk" ]; then
            APK_PATH="app/build/outputs/apk/debug/app-debug.apk"
          fi
          if [ -n "$APK_PATH" ] && [ -f "$APK_PATH" ]; then
            echo "APK found: $APK_PATH"
            ls -la "$APK_PATH"
            sha256sum "$APK_PATH"
            cp "$APK_PATH" app-release.apk
          else
            echo "ERROR: No APK found"
            find . -name "*.apk" -type f 2>/dev/null
            exit 1
          fi

      - name: Sign APK
        working-directory: android
        run: |
          echo "Signing APK..."
          if [ ! -f debug.keystore ]; then
            keytool -genkeypair -v -keystore debug.keystore -alias debug -keyalg RSA -keysize 2048 -validity 10000 -storepass android -keypass android -dname "CN=Debug, OU=Debug, O=Debug, L=Debug, S=Debug, C=US"
          fi
          if [ -f app-release.apk ]; then
            ${ANDROID_HOME}/build-tools/34.0.0/apksigner sign --v1-signing-enabled true --v2-signing-enabled true --v3-signing-enabled true --ks debug.keystore --ks-key-alias debug --ks-pass pass:android --key-pass pass:android --out app-release-signed.apk app-release.apk && mv app-release-signed.apk app-release.apk
            echo "APK signed successfully"
          else
            echo "ERROR: app-release.apk not found"
            exit 1
          fi

      - name: Upload APK artifact
        uses: actions/upload-artifact@v4
        with:
          name: local-apk-builder-debug
          path: android/app-release.apk
          if-no-files-found: error
`
const workflowAABYAML = `name: Build Android AAB

on:
  workflow_dispatch:
    inputs:
      build_id:
        description: 'Build ID'
        required: true

permissions:
  contents: read

env:
  SOURCE_TYPE: 'url'
  PROJECT_TYPE: 'static'
  BUILD_SYSTEM: 'none'
  NEEDS_BUILD: 'false'
  SOURCE_DIR: '.'
  WEB_OUTPUT_DIR: 'www'

jobs:
  build:
    name: Build Release AAB
    runs-on: ubuntu-latest

    steps:
      - name: Checkout repository
        uses: actions/checkout@v4

      - name: Write build payload to file
        env:
          BUILD_ID: ${{ github.event.inputs.build_id }}
        run: |
          echo "Build ID: $BUILD_ID"
          echo "$BUILD_ID" > /tmp/build_id.txt

      - name: Setup Node.js 22
        uses: actions/setup-node@v4
        with:
          node-version: '22'

      - name: Build frontend (Node projects only)
        if: ${{ env.NEEDS_BUILD == 'true' }}
        run: |
          set -e
          echo "=== Frontend build ==="
          echo "Project type: $PROJECT_TYPE"
          echo "Build system: $BUILD_SYSTEM"
          echo "Working dir: $SOURCE_DIR"
          
          cd "$SOURCE_DIR"
          
          echo "Installing dependencies..."
          npm install --legacy-peer-deps
          
          echo "Running build..."
          npm run build
          
          echo "Checking build output..."
          if [ ! -d "dist" ]; then
            echo "ERROR: dist/ not found after build"
            ls -la
            exit 1
          fi
          
          echo "Copying dist/ to www/"
          mkdir -p ../www
          cp -r dist/* ../www/
          
          echo "Build output copied successfully"

      - name: Install dependencies and add Android platform
        run: |
          set -e
          echo "Step 1: Installing npm dependencies..."
          npm install --legacy-peer-deps
          echo "npm install succeeded"

          echo "Step 2: Checking for android directory..."
          if [ ! -d android ]; then
            echo "android/ directory not found. Running: npx cap add android"
            npx cap add android || { echo "ERROR: npx cap add android failed"; exit 1; }
            echo "Successfully created android/ directory"
          else
            echo "android/ directory already exists"
          fi

          echo "Step 3: Running npx cap sync android..."
          npx cap sync android || { echo "ERROR: npx cap sync android failed"; exit 1; }
          echo "Android platform setup complete"

          echo "Injecting native navigation files..."
          mkdir -p android/app/src/main/assets
          cp android-custom/app_config.json android/app/src/main/assets/app_config.json
          # MainActivity left as Capacitor default - no replacement

          # Decode app tray icon from base64
          if [ -f android-custom/icon_base64.txt ]; then
            mkdir -p assets
            base64 -d android-custom/icon_base64.txt > assets/icon.png
            echo "App tray icon decoded to assets/icon.png"
          fi

          echo "Generating launcher icons with @capacitor/assets..."
          npx capacitor-assets generate --android

          if [ -f android-custom/google-services.json ]; then
            cp android-custom/google-services.json android/app/google-services.json
          fi

          if [ -f android-custom/splash.png ]; then
            for d in drawable drawable-port-mdpi drawable-port-hdpi drawable-port-xhdpi drawable-port-xxhdpi drawable-port-xxxhdpi drawable-land-mdpi drawable-land-hdpi drawable-land-xhdpi drawable-land-xxhdpi drawable-land-xxxhdpi; do
              mkdir -p android/app/src/main/res/$d
              cp android-custom/splash.png android/app/src/main/res/$d/splash.png
            done
          fi


      - name: Setup Java 21
        uses: actions/setup-java@v4
        with:
          distribution: 'temurin'
          java-version: '21'

      - name: Setup Android SDK
        uses: android-actions/setup-android@v3

      - name: Build release AAB
        working-directory: android
        run: |
          chmod +x gradlew
          ./gradlew bundleRelease --no-daemon

      - name: Verify AAB exists
        working-directory: android
        run: |
          AAB_PATH="app/build/outputs/bundle/release/app-release.aab"
          if [ -f "$AAB_PATH" ]; then
            echo "AAB exists: $AAB_PATH"
            ls -la "$AAB_PATH"
            sha256sum "$AAB_PATH"
          else
            echo "AAB not found"
            exit 1
          fi

      - name: Upload AAB artifact
        uses: actions/upload-artifact@v4
        with:
          name: local-apk-builder-aab
          path: android/app/build/outputs/bundle/release/app-release.aab
          if-no-files-found: error
`



type GitHubActionsProvider struct {
    client          *github.Client
    workflowID      string
    artifactName    string
    aabArtifactName string
}

func NewGitHubActionsProvider(token, repo, workflowID, artifactName string) *GitHubActionsProvider {
    return &GitHubActionsProvider{
        client:          github.NewClient(token, repo),
        workflowID:      workflowID,
        artifactName:    artifactName,
        aabArtifactName: "local-apk-builder-aab",
    }
}

func (p *GitHubActionsProvider) Build(ctx context.Context, build *database.Build, cfg BuildConfig) (*BuildResult, error) {
    if build.ProviderRunID != "" {
        return &BuildResult{ProviderName: "github", Status: StatusSubmitted, ProviderRunID: build.ProviderRunID}, nil
    }

    // 1. Generate project files
    project, err := database.GetProject(build.ProjectID)
    if err != nil || project == nil {
        return nil, fmt.Errorf("project not found")
    }
    // Create generator with source adapter for upload projects
    var gen *generator.Generator
    if project.SourceType == "upload" && project.SourcePath != "" {
        // source_path already points to the workspace source directory
        // source_path = .../projects/{id}/source, so workspace root = its parent
        workspaceRoot := filepath.Dir(project.SourcePath)
        ws := &database.Workspace{
            Root:     workspaceRoot,
            Source:   project.SourcePath,
            Prepared: filepath.Join(workspaceRoot, "prepared"),
            Build:    filepath.Join(workspaceRoot, "build"),
        }
        adapter := &generator.UploadSourceAdapter{Project: project, Workspace: ws}
        prepared, err := adapter.Prepare(ctx)
        if err != nil {
            return nil, fmt.Errorf("failed to prepare upload source: %w", err)
        }
        gen = generator.NewGeneratorWithSource(project, prepared)
    } else {
        gen = generator.NewGenerator(project)
    }
    
    generated, err := gen.Generate()
    if err != nil {
        return nil, err
    }

    // Generate dynamic workflow YAML based on project profile
    workflowParams := WorkflowParams{
        SourceType:   "url",
        ProjectType:  "static",
        BuildSystem:  "none",
        NeedsBuild:   false,
        SourceDir:    ".",
        WebOutputDir: "www",
        NeedsRuntime: project.RuntimeType == "go-sqlite",
		NeedsNativeRuntime: project.RuntimeType == "go-native",
    }
    
    if project.SourceType == "upload" && project.SourcePath != "" {
        detector := &generator.ProjectDetector{SourceDir: project.SourcePath}
        profile, err := detector.Detect()
        if err == nil {
            workflowParams.SourceType = "upload"
            workflowParams.ProjectType = profile.Type
            workflowParams.BuildSystem = profile.Builder
            workflowParams.NeedsBuild = profile.NeedsBuild()
            if profile.Type == "node" {
                workflowParams.SourceDir = "source"
                workflowParams.WebOutputDir = "www"
            }
        }
    }
    
    workflowFile := GenerateWorkflowYAML(workflowParams, build.BuildType == "aab")

    files := map[string]string{
        "capacitor.config.json":                           generated.CapacitorConfig,
        "package.json":                                    generated.PackageJSON,
        "www/index.html":                                  generated.WebIndex,
        "www/app_bridge.js":                               generated.AppBridgeJS,
        "www/app/logo.png":                                string(generated.LauncherIconPNG),
        ".github/workflows/build-apk.yml":                 workflowFile,
        "android-custom/app_config.json":                  generated.AppConfigJSON,
        "www/app_config.json":                             generated.AppConfigJSON,
        "android-custom/icon_base64.txt":                    generated.IconBase64,
        "android-custom/google-services.json":             string(generated.GoogleServicesJSON),
    }

	// WEB-SOURCE-004 / WEB-SOURCE-005: when a runtime is enabled, push the
	// runtime package source files and build script into the ephemeral repo.
	switch project.RuntimeType {
	case "go-sqlite":
		for path, content := range RuntimeSourcesForRepo() {
			files[path] = content
		}
		log.Printf("DEBUG: Pushed go-sqlite runtime files to ephemeral repo")
	case "go-native":
		for path, content := range NativeRuntimeSourcesForRepo() {
			files[path] = content
		}
		log.Printf("DEBUG: Pushed go-native runtime files to ephemeral repo")
	}

    
    // For upload projects, push prepared/ assets to www/app/
    if project.SourceType == "upload" && project.SourcePath != "" {
        // Determine project profile
        detector := &generator.ProjectDetector{SourceDir: project.SourcePath}
        profile, err := detector.Detect()
        if err != nil {
            log.Printf("DEBUG: Project detection failed: %v", err)
        } else {
            // Push prepared/ as www/app/ (for static projects)
            preparedDir := filepath.Join(filepath.Dir(project.SourcePath), "prepared")
            preparedFiles, err := CollectSourceFiles(preparedDir)
            if err != nil {
                log.Printf("DEBUG: Failed to collect prepared files: %v", err)
            } else {
                for relPath, content := range preparedFiles {
                    files["www/app/"+relPath] = content
                }
                log.Printf("DEBUG: Pushed %d prepared files to www/app/", len(preparedFiles))
            }

            // For Node projects, also push source/ for GitHub Actions build
            if profile.Type == "node" {
                sourceFiles, err := CollectSourceFiles(project.SourcePath)
                if err != nil {
                    return nil, fmt.Errorf("failed to collect source files: %w", err)
                }
                for relPath, content := range sourceFiles {
                    files["source/"+relPath] = content
                }
                log.Printf("DEBUG: Pushed %d source files for Node project", len(sourceFiles))
            }
        }
    }

    // 2. Create repo and push files
    repoFull, err := p.PushProject(ctx, build.ProjectID, files)
    if err != nil {
        return nil, err
    }

    // Update client repo to the new repo
    p.client.Repo = repoFull

    // 3. Dispatch workflow
    // Only pass build_id to avoid "inputs are too large" error
    inputs := map[string]interface{}{
        "build_id": fmt.Sprintf("%d", build.ID),
    }
    log.Printf("DEBUG: Starting 90s sleep before dispatch")
    time.Sleep(90 * time.Second)
    log.Printf("DEBUG: Sleep complete, dispatching workflow")
    if err := p.client.DispatchWorkflow(ctx, p.workflowID, inputs); err != nil {
        log.Printf("DEBUG: Dispatch failed: %v", err)
        return nil, err
    }
    log.Printf("DEBUG: Dispatch succeeded")

    // 4. Find run ID
    runID, err := p.findRunByBuildID(ctx, build.ID)
    if err != nil {
        return &BuildResult{ProviderName: "github", Status: StatusSubmitted}, nil
    }

    return &BuildResult{
        ProviderName: "github",
        Status:       StatusSubmitted,
        ProviderRunID: runID,
    }, nil
}

func (p *GitHubActionsProvider) findRunByBuildID(ctx context.Context, buildID int64) (string, error) {
    runs, err := p.client.ListWorkflowRuns(ctx, p.workflowID)
    if err != nil {
        return "", err
    }
    for _, run := range runs {
        // Check if the run's payload contains our build_id
        // In real GitHub API, inputs are not directly accessible from run object.
        // We can use run name or correlation by created_at.
        // For now, we return the latest run's ID as best effort.
        // A more robust method would use a unique workflow input and fetch runs.
        id, _ := run["id"].(float64)
        if id > 0 {
            return fmt.Sprintf("%.0f", id), nil
        }
    }
    return "", fmt.Errorf("no run found")
}

func (p *GitHubActionsProvider) GetStatus(ctx context.Context, providerRunID string) (*BuildResult, error) {
    run, err := p.client.GetRun(ctx, providerRunID)
    if err != nil {
        return nil, err
    }
    status := runStatusToString(run)
    return &BuildResult{
        ProviderName: "github",
        Status:       status,
    }, nil
}

func (p *GitHubActionsProvider) DownloadArtifact(ctx context.Context, providerRunID string) ([]byte, error) {
    artifacts, err := p.client.ListArtifacts(ctx, providerRunID)
    if err != nil {
        return nil, err
    }
    targetName := p.artifactName
    for _, art := range artifacts {
        name, _ := art["name"].(string)
        if name == p.aabArtifactName {
            targetName = p.aabArtifactName
        }
        if name == targetName {
            url, _ := art["archive_download_url"].(string)
            if url == "" {
                return nil, fmt.Errorf("artifact has no download URL")
            }
            return p.client.DownloadArtifact(ctx, url)
        }
    }
    return nil, fmt.Errorf("artifact %s not found", p.artifactName)
}

// ExtractAPK extracts APK from downloaded ZIP artifact, returns path, size, hash
func (p *GitHubActionsProvider) ExtractAPK(zipData []byte) (string, int64, string, error) {
    reader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
    if err != nil {
        return "", 0, "", fmt.Errorf("invalid ZIP: %w", err)
    }

    var apkFile *zip.File
    for _, f := range reader.File {
        if (strings.HasSuffix(f.Name, ".apk") || strings.HasSuffix(f.Name, ".aab")) && !strings.HasPrefix(f.Name, "__MACOSX/") {
            apkFile = f
            break
        }
    }
    if apkFile == nil {
        return "", 0, "", fmt.Errorf("no APK found in artifact")
    }

    rc, err := apkFile.Open()
    if err != nil {
        return "", 0, "", err
    }
    defer rc.Close()

    apkData, err := io.ReadAll(rc)
    if err != nil {
        return "", 0, "", err
    }
    if len(apkData) == 0 {
        return "", 0, "", fmt.Errorf("APK is empty")
    }

    hash := sha256.Sum256(apkData)
    return string(apkData), int64(len(apkData)), hex.EncodeToString(hash[:]), nil
}

func runStatusToString(run map[string]interface{}) string {
    status, _ := run["status"].(string)
    conclusion, _ := run["conclusion"].(string)
    switch status {
    case "queued":
        return StatusQueued
    case "in_progress":
        return StatusBuilding
    case "completed":
        switch conclusion {
        case "success":
            return StatusSuccess
        case "failure", "cancelled", "timed_out":
            return StatusFailed
        }
    }
    return StatusSubmitted
}


// PushProject creates a repo (or uses existing) and pushes generated files
func (p *GitHubActionsProvider) PushProject(ctx context.Context, projectID int64, files map[string]string) (string, error) {
    // Use unique repo name with timestamp to avoid conflicts
    repoName := fmt.Sprintf("app-builder-%d-%d", projectID, time.Now().Unix())
    repo, err := p.client.CreateRepo(ctx, repoName, "Generated by Local APK Builder")
    if err != nil {
        // If repo exists, try to use existing
        if strings.Contains(err.Error(), "name already exists") {
            return "", fmt.Errorf("repo already exists, using existing not yet implemented")
        }
        return "", err
    }
    owner := ""
    if ownerVal, ok := repo["owner"].(map[string]interface{}); ok {
        if login, ok := ownerVal["login"].(string); ok {
            owner = login
        }
    }
    if owner == "" {
        // Fallback: use current token's username from repo full_name
        if fullName, ok := repo["full_name"].(string); ok {
            parts := strings.SplitN(fullName, "/", 2)
            if len(parts) == 2 {
                owner = parts[0]
            }
        }
    }
    repoFull := fmt.Sprintf("%s/%s", owner, repoName)

    for path, content := range files {
        log.Printf("DEBUG: Pushing file: %s (size: %d)", path, len(content))
        if err := p.client.PushFile(ctx, repoFull, path, content, "Generate project"); err != nil {
            log.Printf("DEBUG: Push failed for %s: %v", path, err)
            return repoFull, err
        }
    }
    return repoFull, nil
}
