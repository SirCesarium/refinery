# Getting Started

This guide will walk you through the initial setup of Refinery and your first build and deployment cycle.

## Installation

### Binary Download
The most efficient way to use Refinery is by downloading a pre-compiled binary for your system from the [Releases](https://github.com/SirCesarium/refinery/releases/latest) page.

### From Source
If you have the Go toolchain installed (v1.26.2 or higher):

```bash
go install github.com/SirCesarium/refinery/cmd/refinery@latest
```

## 1. Initializing a Project

To start, run the initialization command in your project root:

```bash
refinery init
```

This creates a `refinery.toml` file with a baseline configuration.

## 2. Configuring Artifacts

Edit `refinery.toml` to define your project structure. For a standard Rust project with a binary and a library:

```toml
refinery_version = "latest"
output_dir = "dist"

[project]
name = "my-app"
lang = "rust"

[artifacts.cli]
type = "bin"
source = "src/main.rs"
packages = ["tar.gz"]

[artifacts.cli.targets.linux]
os = "linux"
archs = ["x86_64", "aarch64"]

[artifacts.core]
type = "lib"
library_types = ["cdylib", "staticlib"]
packages = ["zip"]

[artifacts.core.targets.windows]
os = "windows"
archs = ["x86_64"]
abis = ["msvc"]
```

## 3. Local Build and Packaging

To build a specific artifact for a target architecture locally:

```bash
refinery build --artifact cli --os linux --arch x86_64
```

Refinery will:
1. Validate the configuration against your `Cargo.toml`.
2. Orchestrate the compiler using the correct target triple.
3. Package the resulting binary into a `.tar.gz` in the `dist/` directory.

## 4. Automating CI/CD

Generate a GitHub Actions workflow to automate multi-platform builds:

```bash
refinery migrate github
```

This command outputs `.github/workflows/refinery-build.yml`, pre-configured with a build matrix and automatic binary delivery.
