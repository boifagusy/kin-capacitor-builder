package build

// runtimeSetupBlock is injected into the workflow YAML when a project has RuntimeType "go-sqlite".
const runtimeSetupBlock = `
      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: 1.22

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

      - name: Build runtime AAR
        run: |
          chmod +x build-go.sh
          ./build-go.sh

      - name: Copy AAR into Android project
        run: |
          mkdir -p android/app/libs
          cp app/libs/runtime.aar android/app/libs/runtime.aar

      - name: Patch Capacitor gradle for runtime
        run: |
          GRADLE=android/app/build.gradle
          if ! grep -q "flatDir" $GRADLE; then
            echo "" >> $GRADLE
            echo "repositories { flatDir { dirs libs } }" >> $GRADLE
            echo "dependencies { implementation fileTree(dir: libs, include: [*.aar]) }" >> $GRADLE
          fi

      - name: Enable cleartext for localhost runtime
        run: |
          MANIFEST=android/app/src/main/AndroidManifest.xml
          if ! grep -q usesCleartextTraffic $MANIFEST; then
            sed -i "s|<application|<application android:usesCleartextTraffic=\"true\"|" $MANIFEST
          fi
`
