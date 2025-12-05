# AGENTS.md - Cross-Platform Distribution Packaging

## Overview
Cross-platform packaging configurations for distribution via various package managers.

## Package Formats
- **Chocolatey**: Windows package manager (`.nuspec` + PowerShell scripts)
- **Debian**: Linux DEB packages (control, install, postinst scripts)
- **Homebrew**: macOS package manager (Ruby formula)
- **Scoop**: Windows package manager (JSON manifest)
- **Winget**: Microsoft Windows Package Manager (YAML manifests)

## Build Scripts
```bash
# Debian packaging
./scripts/build-deb.sh    # Build DEB package

# Cross-platform builds
./build.bat              # Multi-platform binaries via Go
```

## Code Patterns
- **Platform-Specific**: Separate configurations for each package manager
- **Installation Scripts**: Pre/post install hooks for setup
- **Metadata**: Version, dependencies, and maintainer information
- **Distribution**: Ready for automated publishing pipelines

## Dependencies
- Platform-specific package manager tools
- Go build tools for binary creation