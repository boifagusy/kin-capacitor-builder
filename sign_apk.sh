#!/data/data/com.termux/files/usr/bin/bash
# Local APK Signing Script for Termux
# Signs APK with v2 signature using local apksigner

set -e

APK_PATH="${1:-builds/build-309.apk}"
OUTPUT_PATH="${2:-builds/build-309-signed.apk}"

echo "=== Signing APK with v2 signature ==="
echo "Input: $APK_PATH"
echo "Output: $OUTPUT_PATH"

# Check if debug.keystore exists
if [ ! -f debug.keystore ]; then
    echo "Generating debug keystore..."
    keytool -genkeypair -v \
      -keystore debug.keystore \
      -alias debug \
      -keyalg RSA \
      -keysize 2048 \
      -validity 10000 \
      -storepass android \
      -keypass android \
      -dname "CN=Debug, OU=Debug, O=Debug, L=Debug, S=Debug, C=US"
fi

# Find apksigner
APKSIGNER="/data/data/com.termux/files/usr/lib/android-sdk/build-tools/34.0.0/apksigner"

# Sign with apksigner (v2+v3)
echo "Signing with apksigner..."
$APKSIGNER sign \
  --align-file-size \
  --ks debug.keystore \
  --ks-key-alias debug \
  --ks-pass pass:android \
  --key-pass pass:android \
  --out "$OUTPUT_PATH" \
  "$APK_PATH"

echo "Verifying signature..."
$APKSIGNER verify --verbose "$OUTPUT_PATH" 2>&1 | grep -E "Verifies|v2|v3"

echo ""
echo "=== SIGNED APK READY ==="
echo "$OUTPUT_PATH"
echo "Size: $(stat -c%s "$OUTPUT_PATH") bytes"
