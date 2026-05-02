# Build Hooks

Refinery allows the execution of custom system commands before and after the build process through the `hooks` configuration. This is essential for tasks like code generation, asset processing, or post-compilation cleanup.

## Configuration

Hooks are defined at the artifact level:

```toml
[artifacts.my_app.hooks]
pre_build = ["./scripts/generate_version.sh", "echo 'Starting build...'"]
post_build = ["cp {binary} ./backups/"]
```

## Execution Logic

1.  **Environment Preparation**: Refinery sets up all target-specific variables (linkers, paths).
2.  **Pre-Build Hooks**: Commands are executed sequentially. If any hook returns a non-zero exit code, the build is aborted (Fail-Fast).
3.  **Compilation**: The underlying compiler is invoked.
4.  **Artifact Retrieval**: Files are moved to the distribution directory.
5.  **Post-Build Hooks**: Commands are executed. Failures at this stage will still mark the build as failed.

## Variable Substitution

Commands defined in hooks can use dynamic placeholders that Refinery resolves at runtime based on the current build target:

- **`{artifact}`**: The name of the artifact.
- **`{os}`**: Target operating system.
- **`{arch}`**: Target architecture.
- **`{abi}`**: Target ABI.
- **`{version}`**: Project version.
- **`{binary}`**: The full path to the compiled binary before it is moved to the final output directory.

### Example: Version Header Generation

```toml
[artifacts.core.hooks]
pre_build = ["sed -i 's/VERSION/%v/g' version.h" ]
```
(Note: substitution occurs in the Go engine before the command is sent to the shell).
