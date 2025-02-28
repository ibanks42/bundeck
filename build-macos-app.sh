#!/bin/bash
set -e

# Build BunDeck macOS application bundle
# Comprehensive macOS app bundle creation script with proper systray support
# Works for Apple Silicon and Intel Macs

# Parameters
APP_NAME="BunDeck"
APP_VERSION=$(git describe --tags --always --dirty || echo "dev")
BUNDLE_IDENTIFIER="com.ibanks42.bundeck"
COPYRIGHT="©️ 2025 Isaiah Banks"

# Destination directories
DIST_DIR="./dist"
APP_DIR="${DIST_DIR}/${APP_NAME}.app"
CONTENTS_DIR="${APP_DIR}/Contents"
MACOS_DIR="${CONTENTS_DIR}/MacOS"
RESOURCES_DIR="${CONTENTS_DIR}/Resources"
FRAMEWORKS_DIR="${CONTENTS_DIR}/Frameworks"

# Clean the existing app if it exists
if [ -d "${DIST_DIR}/${APP_NAME}.app" ]; then
    echo "Cleaning existing app bundle..."
    rm -rf "${DIST_DIR}/${APP_NAME}.app"
fi

# Create the basic app bundle structure
mkdir -p "${MACOS_DIR}"
mkdir -p "${RESOURCES_DIR}"
mkdir -p "${FRAMEWORKS_DIR}"

# Copy app icon if it exists
if [ -f "logo.icns" ]; then
    cp "logo.icns" "${RESOURCES_DIR}/AppIcon.icns"
elif [ -f "icon.icns" ]; then
    cp "icon.icns" "${RESOURCES_DIR}/AppIcon.icns"
else
    # Create a blank icon file as a placeholder
    touch "${RESOURCES_DIR}/AppIcon.icns"
    echo "Warning: No logo.icns or icon.icns found. Using placeholder."
fi

# Build for both architectures and create universal binary
echo "Building for Apple Silicon (arm64)..."
GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 go build -ldflags "-s -w" -o "${APP_NAME}-arm64" .

echo "Building for Intel (amd64)..."
GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build -ldflags "-s -w" -o "${APP_NAME}-amd64" .

echo "Creating universal binary..."
lipo -create -output "${APP_NAME}" "${APP_NAME}-arm64" "${APP_NAME}-amd64"

# Check the resulting binary
lipo -info "${APP_NAME}"

# Copy binary to app bundle
cp "${APP_NAME}" "${MACOS_DIR}/${APP_NAME}-bin"

# Create a launcher script that properly handles macOS environment
cat > "${MACOS_DIR}/${APP_NAME}" << 'EOF'
#!/bin/bash
# Launcher script for BunDeck
# This handles proper launching of system tray applications

# Get the directory where this script is located
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
LOG_FILE=~/Library/Logs/BunDeck.log

# Create log directory if it doesn't exist
mkdir -p ~/Library/Logs

# Log start time and environment
echo "==============================================" >> "${LOG_FILE}"
echo "[BunDeck] Starting at $(date)" >> "${LOG_FILE}"
echo "[BunDeck] Running from $DIR" >> "${LOG_FILE}"
echo "[BunDeck] PATH: $PATH" >> "${LOG_FILE}"

# Remove quarantine attribute if it exists
echo "[BunDeck] Removing quarantine attribute if present" >> "${LOG_FILE}"
xattr -d com.apple.quarantine "$DIR/../.." 2>/dev/null || true

# Set up environment for macOS
export PATH="/usr/bin:/bin:/usr/sbin:/sbin:$PATH"

# Fix for Finder launching
if [[ -z "$TERM_PROGRAM" && "$TERM" == "dumb" ]]; then
    # This is being launched from Finder
    echo "[BunDeck] Detected launch from Finder" >> "${LOG_FILE}"
    # Force GUI mode and set proper working directory
    cd "$DIR"
    exec "$DIR/BunDeck-bin" >> "${LOG_FILE}" 2>&1
else
    # Standard launch (Terminal or direct execution)
    echo "[BunDeck] Standard launch" >> "${LOG_FILE}"
    cd "$DIR"
    exec "$DIR/BunDeck-bin" >> "${LOG_FILE}" 2>&1
fi
EOF

# Make the launcher script executable
chmod +x "${MACOS_DIR}/${APP_NAME}"

# Create a simple DMG background image
cat > "${DIST_DIR}/dmg-background.svg" << 'EOF'
<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<svg width="600" height="400" xmlns="http://www.w3.org/2000/svg">
  <rect width="600" height="400" fill="#f5f5f7"/>
  <path d="M280,190 L320,190 L320,180 L340,200 L320,220 L320,210 L280,210 Z" fill="#333"/>
</svg>
EOF

# Create Info.plist with full macOS menu bar app support
cat > "${CONTENTS_DIR}/Info.plist" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleDevelopmentRegion</key>
    <string>English</string>
    <key>CFBundleExecutable</key>
    <string>${APP_NAME}</string>
    <key>CFBundleIconFile</key>
    <string>AppIcon</string>
    <key>CFBundleIdentifier</key>
    <string>${BUNDLE_IDENTIFIER}</string>
    <key>CFBundleInfoDictionaryVersion</key>
    <string>6.0</string>
    <key>CFBundleName</key>
    <string>${APP_NAME}</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleShortVersionString</key>
    <string>${APP_VERSION}</string>
    <key>CFBundleVersion</key>
    <string>${APP_VERSION}</string>
    <key>NSHumanReadableCopyright</key>
    <string>${COPYRIGHT}</string>
    <key>NSHighResolutionCapable</key>
    <true/>
    <key>LSUIElement</key>
    <true/>
    <key>NSPrincipalClass</key>
    <string>NSApplication</string>
    <key>NSSupportsAutomaticGraphicsSwitching</key>
    <true/>
    <key>LSApplicationCategoryType</key>
    <string>public.app-category.utilities</string>
    <key>LSMinimumSystemVersion</key>
    <string>10.15</string>
    <key>NSRequiresAquaSystemAppearance</key>
    <false/>
    <key>LSMultipleInstancesProhibited</key>
    <true/>
    <key>LSBackgroundOnly</key>
    <false/>
    <key>NSAppTransportSecurity</key>
    <dict>
        <key>NSAllowsArbitraryLoads</key>
        <true/>
    </dict>
    <key>NSUserNotificationAlertStyle</key>
    <string>alert</string>
    <key>ITSAppUsesNonExemptEncryption</key>
    <false/>
</dict>
</plist>
EOF

# Create PkgInfo file
echo "APPL????" > "${CONTENTS_DIR}/PkgInfo"

# Clean up temporary files
rm -f "${APP_NAME}-arm64" "${APP_NAME}-amd64" "${APP_NAME}"

# Create zip archives for distribution
if [ "$1" == "intel" ] || [ "$1" == "apple" ]; then
    echo "Creating zip for $1..."
    (cd "${DIST_DIR}" && zip -r "${APP_NAME}-mac-$1.zip" "${APP_NAME}.app")
    mv "${DIST_DIR}/${APP_NAME}-mac-$1.zip" .

    echo "Created ${APP_NAME}-mac-$1.zip"
else
    # If no specific architecture is specified, just output the app bundle
    echo "Universal app bundle created (no specific architecture specified for zip)"
fi

echo "macOS app bundle created successfully!"
echo "You can run the app with: open ${APP_DIR}"
