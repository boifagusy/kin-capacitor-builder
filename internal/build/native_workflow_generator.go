package build

// nativeRuntimeSetupBlock is injected into the workflow YAML when
// a project has RuntimeType "go-native". It provides:
//   - Go toolchain
//   - Android NDK (for gomobile bind)
//   - gomobile + gobind tools
//   - runtime.aar build (arm64 + armv7)
//   - gradle patch to add the AAR dependency
//   - cleartext manifest for localhost (if used)
const nativeRuntimeSetupBlock = `
      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Install Android NDK
        run: |
          yes | sdkmanager --install "ndk;26.1.10909125" || true
          echo "ANDROID_NDK_HOME=$ANDROID_HOME/ndk/26.1.10909125" >> $GITHUB_ENV
          echo "ANDROID_NDK_ROOT=$ANDROID_HOME/ndk/26.1.10909125" >> $GITHUB_ENV

      - name: Install gomobile
        run: |
          go install golang.org/x/mobile/cmd/gomobile@latest
          go install golang.org/x/mobile/cmd/gobind@latest
          echo "$(go env GOPATH)/bin" >> $GITHUB_PATH

      - name: Initialize gomobile
        run: gomobile init

      - name: Build native AAR
        run: |
          chmod +x build-go.sh
          ./build-go.sh

      - name: Add AAR dependency to build.gradle
        run: |
          GRADLE=android/app/build.gradle
          if ! grep -q "runtime.aar" $GRADLE; then
            echo "" >> $GRADLE
            echo "dependencies {" >> $GRADLE
            echo "    implementation files('libs/runtime.aar')" >> $GRADLE
            echo "}" >> $GRADLE
          fi

      - name: Enable cleartext
        run: |
          MANIFEST=android/app/src/main/AndroidManifest.xml
          if ! grep -q usesCleartextTraffic $MANIFEST; then
            sed -i 's|<application|<application android:usesCleartextTraffic="true"|' $MANIFEST
          fi
`
