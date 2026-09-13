package build

import (
	"fmt"
	"strings"
)

// WorkflowParams contains parameters for workflow generation
type WorkflowParams struct {
	SourceType   string // "url" | "upload"
	ProjectType  string // "static" | "node"
	BuildSystem  string // "vite" | "none"
	NeedsBuild   bool
	SourceDir    string // "source" or "."
	WebOutputDir string // "www" or "dist"
	NeedsRuntime bool   // when true, inject gomobile + NDK steps for go-sqlite
	NeedsNativeRuntime bool // when true, inject gomobile + NDK steps for go-native
}

// GenerateWorkflowYAML creates the workflow YAML with the given params
func GenerateWorkflowYAML(params WorkflowParams, isAAB bool) string {
	needsBuild := "false"
	if params.NeedsBuild {
		needsBuild = "true"
	}

	workflowName := "Build Android APK"
	artifactName := "local-apk-builder-debug"
	if isAAB {
		workflowName = "Build Android AAB"
		artifactName = "local-apk-builder-aab"
	}

	result := fmt.Sprintf(`name: %s

on:
  push:
    branches: [main]
  workflow_dispatch:
    inputs:
      build_id:
        description: 'Build ID'
        required: true

permissions:
  contents: read

env:
  SOURCE_TYPE: '%s'
  PROJECT_TYPE: '%s'
  BUILD_SYSTEM: '%s'
  NEEDS_BUILD: '%s'
  NEEDS_RUNTIME: 'false'
  SOURCE_DIR: '%s'
  WEB_OUTPUT_DIR: '%s'

jobs:
  build:
    name: %s
    runs-on: ubuntu-latest
    timeout-minutes: 60

    steps:
      - name: Checkout repository
        uses: actions/checkout@v4

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
          echo "Source dir: $SOURCE_DIR"

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

          echo "Copying dist/ to www/app/"
          mkdir -p ../www/app
          cp -r dist/* ../www/app/

          # Fix Vite output for Capacitor WebView
          if [ -f "../www/app/index.html" ]; then
            echo "Fixing Vite index.html for Capacitor..."
            sed -i 's/ crossorigin//g' ../www/app/index.html
            sed -i 's|src="/assets/|src="./assets/|g' ../www/app/index.html
            sed -i 's|href="/assets/|href="./assets/|g' ../www/app/index.html
            echo "Fixed absolute asset paths"

            # Inject app_bridge.js before </body>
            if ! grep -q "app_bridge.js" ../www/app/index.html; then
              sed -i 's|</body>|<script src="../app_bridge.js"></script></body>|' ../www/app/index.html
              echo "Injected app_bridge.js"
            fi

            # Inject safe-area CSS before </head>
            if ! grep -q "kin-safe-area" ../www/app/index.html; then
              sed -i 's|</head>|<style id="kin-safe-area">body{padding-top:env(safe-area-inset-top);padding-bottom:env(safe-area-inset-bottom);}html,body{overscroll-behavior:none;-webkit-tap-highlight-color:transparent;}</style></head>|' ../www/app/index.html
              echo "Injected safe-area CSS"
            fi

            echo "=== Final index.html ==="
            head -30 ../www/app/index.html
          fi

          echo "Build output copied to www/app/"

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
          else
            echo "android/ directory already exists"
          fi

          echo "Step 3: Running npx cap sync android..."
          npx cap sync android || { echo "ERROR: npx cap sync android failed"; exit 1; }
          echo "Android platform setup complete"

          echo "Injecting native navigation files..."
          mkdir -p android/app/src/main/assets
          cp android-custom/app_config.json android/app/src/main/assets/app_config.json

          # Decode icon from base64
          if [ -f android-custom/icon_base64.txt ]; then
            mkdir -p assets
            base64 -d android-custom/icon_base64.txt > assets/icon.png
            echo "Icon decoded to assets/icon.png"
          fi

          echo "Generating launcher icons with @capacitor/assets..."
          npx capacitor-assets generate --android

          if [ -f android-custom/google-services.json ]; then
            cp android-custom/google-services.json android/app/google-services.json
          fi

      - name: Setup Java 21
        uses: actions/setup-java@v4
        with:
          distribution: 'temurin'
          java-version: '21'

      - name: Setup Android SDK
        uses: android-actions/setup-android@v3

      # RUNTIME_BLOCK_HERE

      - name: Build release %s
        working-directory: android
        run: |
          chmod +x gradlew
          %s --no-daemon

      - name: Verify APK exists
        working-directory: android
        run: |
          echo "Searching for APK..."
          find app/build/outputs/apk -name "*.apk" 2>/dev/null || echo "No APK directory"
          APK_PATH=""
          for candidate in app-release.apk app-release-unsigned.apk app-debug.apk; do
            if [ -f "app/build/outputs/apk/release/$candidate" ]; then
              APK_PATH="app/build/outputs/apk/release/$candidate"
              break
            fi
            if [ -f "app/build/outputs/apk/debug/$candidate" ]; then
              APK_PATH="app/build/outputs/apk/debug/$candidate"
              break
            fi
          done
          if [ -n "$APK_PATH" ] && [ -f "$APK_PATH" ]; then
            echo "APK found: $APK_PATH"
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

          APKSIGNER=$(find ${ANDROID_HOME}/build-tools -name apksigner -type f | sort -r | head -1)
          echo "Using apksigner: $APKSIGNER"

          # Remove old signature first
          zip -d app-release.apk "META-INF/*.SF" "META-INF/*.RSA" "META-INF/*.DSA" 2>/dev/null || true

          $APKSIGNER sign --v1-signing-enabled true --v2-signing-enabled true --v3-signing-enabled true --ks debug.keystore --ks-key-alias debug --ks-pass pass:android --key-pass pass:android --out app-release-signed.apk app-release.apk
          if [ -f app-release-signed.apk ]; then
            mv app-release-signed.apk app-release.apk
            echo "APK signed with v2+v3 signature"
            $APKSIGNER verify --verbose app-release.apk | head -5
          else
            echo "ERROR: signing failed"
            exit 1
          fi

      - name: Upload %s artifact
        uses: actions/upload-artifact@v4
        with:
          name: %s
          path: android/app-release.%s
          if-no-files-found: error
`,
		workflowName,
		params.SourceType,
		params.ProjectType,
		params.BuildSystem,
		needsBuild,
		params.SourceDir,
		params.WebOutputDir,
		workflowName,
		strings.ToUpper(isAABtoExt(isAAB)),
		isAABtoGradle(isAAB),
		strings.ToUpper(isAABtoExt(isAAB)),
		artifactName,
		isAABtoExt(isAAB),
	)

	switch {
	case params.NeedsRuntime:
		result = strings.Replace(result, "# RUNTIME_BLOCK_HERE", runtimeSetupBlock, 1)
		result = strings.Replace(result, "NEEDS_RUNTIME: 'false'", "NEEDS_RUNTIME: 'true'", 1)
	case params.NeedsNativeRuntime:
		result = strings.Replace(result, "# RUNTIME_BLOCK_HERE", nativeRuntimeSetupBlock, 1)
		result = strings.Replace(result, "NEEDS_RUNTIME: 'false'", "NEEDS_RUNTIME: 'true'", 1)
	default:
		result = strings.Replace(result, "# RUNTIME_BLOCK_HERE", "", 1)
	}

	return result
}

func isAABtoExt(isAAB bool) string {
	if isAAB {
		return "aab"
	}
	return "apk"
}

func isAABtoGradle(isAAB bool) string {
	if isAAB {
		return "./gradlew bundleRelease"
	}
	return "./gradlew assembleRelease"
}

