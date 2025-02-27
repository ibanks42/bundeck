# BunDeck

<img src="https://raw.githubusercontent.com/ibanks42/bundeck/refs/heads/main/logo.png" width="256" />

[![GitHub release](https://img.shields.io/github/v/release/ibanks42/bundeck)](https://github.com/ibanks42/bundeck/releases)
[![GitHub issues](https://img.shields.io/github/issues/ibanks42/bundeck)](https://github.com/ibanks42/bundeck/issues)
[![GitHub pull requests](https://img.shields.io/github/issues-pr/ibanks42/bundeck)](https://github.com/ibanks42/bundeck/pulls)
[![GitHub license](https://img.shields.io/github/license/ibanks42/bundeck)](https://github.com/ibanks42/bundeck/blob/main/LICENSE.md)

BunDeck is an open source StreamDeck alternative that allows you to create and run customizable plugins using the Bun JavaScript runtime. It provides a modern, flexible way to automate tasks, control applications, and enhance your workflow from a single interface.

## Overview

BunDeck is a cross-platform desktop application that creates a system tray interface where you can create, manage, and run small JavaScript/TypeScript plugins. Each plugin can perform custom tasks, from controlling OBS scenes to sending keystrokes to your operating system.

### Key Features

- **Plugin System**: Create and run JavaScript/TypeScript plugins to automate various tasks
- **Built with Bun**: Leverages the speed and simplicity of the Bun JavaScript runtime
- **Cross-Platform**: Works on Windows, macOS, and Linux (including WSL)
- **System Tray Integration**: Runs in your system tray for easy access
- **Template Library**: Comes with pre-built templates for common tasks
- **Mobile Access**: Access your deck from mobile devices via QR code
- **Drag-and-Drop Interface**: Easily reorganize your plugins with drag-and-drop
- **Customizable Appearance**: Add custom images to your plugin buttons
- **Code Editor**: Built-in Monaco editor for creating and editing plugins
- **SQLite Storage**: Efficient local storage of plugins and settings

## Technology Stack

### Backend

- **Go**: Core application built with Go for performance and cross-platform support
- **Fiber**: Fast HTTP server for the API
- **SQLite**: Local database storage using modernc.org/sqlite
- **Systray**: System tray integration via fyne.io/systray

### Frontend

- **React 19**: Modern React for the web interface
- **TanStack Router & Query**: For routing and data fetching
- **Tailwind CSS**: For styling and responsive design
- **Monaco Editor**: Code editing with @monaco-editor/react
- **DnD Kit**: Drag-and-drop functionality for reordering plugins
- **Shadcn UI**: Accessible UI components

### Plugin Runtime

- **Bun**: Fast JavaScript/TypeScript runtime for executing plugins
- **TypeScript**: Type safety for plugin development

## Getting Started

### Prerequisites

- [Bun](https://bun.sh/) must be installed on your system (for running plugins)
- A modern web browser

### Installation

1. Download the latest release for your platform from the releases page
2. Run the executable file
3. BunDeck will start in your system tray

### Usage

1. Click the BunDeck icon in your system tray and select "Open App"
2. The web interface will open in your browser
3. Click "Edit" to enter edit mode
4. Add plugins from templates or create your own
5. Click a plugin to run it

## Plugin Development

Plugins in BunDeck are JavaScript/TypeScript files that can:

- Control external applications via their APIs
- Send keystrokes to your operating system
- Call web services
- Process and display information
- Automate repetitive tasks

### Example Plugin

```typescript
// Simple example plugin
import { v4 as uuidv4 } from 'uuid';

// Output appears in the BunDeck UI
console.log(uuidv4());
```

### Available Plugin Templates

BunDeck comes with several plugin templates:

- **OBS Scene Control**: Toggle webcam visibility across OBS scenes
- **Keystroke Sender**: Send keyboard shortcuts to your operating system
- More templates are going to be added regularly

## Contributing

For bugs, features, and discussion please use [GitHub Issues](https://github.com/ibanks42/bundeck/issues).

If you're opening a pull request, please make sure to run `make format` first.

## macOS App Distribution

### Code Signing and Notarization

For proper distribution of the macOS app bundle, the application needs to be code signed and notarized with Apple. This ensures users can run the app without security warnings.

#### Requirements for Code Signing:

1. **Apple Developer Account**: You need to be enrolled in the [Apple Developer Program](https://developer.apple.com/programs/) ($99/year).
2. **Developer ID Certificate**: Create a Developer ID Application certificate through the Apple Developer portal.
3. **Notarization Access**: Set up an app-specific password for your Apple ID to allow for automated notarization.

#### Setting up GitHub Secrets:

For automated building and signing in GitHub Actions, add these secrets to your repository:

- `MACOS_CERTIFICATE`: The base64-encoded .p12 certificate file
- `MACOS_CERTIFICATE_PWD`: The password for your .p12 certificate
- `MACOS_CERTIFICATE_NAME`: The name of your certificate (e.g., "Developer ID Application: Your Name (TEAMID)")
- `MACOS_NOTARIZATION_APPLE_ID`: Your Apple ID email
- `MACOS_NOTARIZATION_TEAM_ID`: Your Apple Developer Team ID
- `MACOS_NOTARIZATION_PWD`: An app-specific password (create in Apple ID settings)
- `MACOS_CI_KEYCHAIN_PWD`: A random password for the temporary keychain

#### Local Testing:

To test code signing locally:

```bash
# Code sign the app (replace with your certificate name)
codesign --force --options runtime --sign "Developer ID Application: Your Name" ./dist/Bundeck.app

# Verify the signature
codesign -dv --verbose=4 ./dist/Bundeck.app
```

For more information, see Apple's [Code Signing Guide](https://developer.apple.com/library/archive/documentation/Security/Conceptual/CodeSigningGuide/Introduction/Introduction.html).

## License

See [LICENSE](LICENSE) for details.
