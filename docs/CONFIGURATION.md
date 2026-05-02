# Configuration Reference (refinery.toml)

The `refinery.toml` file is the central authority for Refinery's orchestration. It is strictly validated before any execution.

## Global Settings

- **`refinery_version`** (string, required): The version of the Refinery binary to use in CI. Supports semantic aliases (`latest`, `1`, `2.0`, etc.).
- **`output_dir`** (string): Directory where built artifacts and packages will be stored. Defaults to `"dist"`.

## [project]
Metadata describing the overall project.

- **`name`** (string, required): Technical name of the project.
- **`lang`** (string, required): Programming language (e.g., `"rust"`).
- **`description`** (string): Short project description for package metadata.
- **`author`** (string): Project maintainer or author.
- **`license`** (string): SPDX license identifier (e.g., `"MIT"`, `"Apache-2.0"`).
- **`homepage`** (string): URL to the project website or repository.

## [artifacts.<name>]
Defines a specific build unit. The `<name>` key must match the component name defined in your language manifest (e.g., the `name` field in `Cargo.toml` for `[[bin]]` or `[lib]`). Refinery uses this name to locate the compiled files.

- **`type`** (string, required): Either `"bin"` (executable) or `"lib"` (library).
- **`source`** (string, required): Path to the entry point (e.g., `"src/main.rs"`).
- **`library_types`** (list of strings): Only for `type = "lib"`. Formats like `"cdylib"`, `"staticlib"`, `"rlib"`.
- **`packages`** (list of strings): Distribution formats: `"deb"`, `"rpm"`, `"msi"`, `"tar.gz"`, `"zip"`.
- **`headers`** (boolean): If `true`, includes `.h` and `.hpp` files in archive packages.

### [artifacts.<name>.targets.<id>]
Platform-specific build configurations.

- **`os`** (string, required): Target OS (`"linux"`, `"windows"`, `"darwin"`, `"wasm"`, `"wasi"`).
- **`archs`** (list of strings, required): Target architectures (e.g., `["x86_64", "aarch64"]`).
- **`abis`** (list of strings): Target ABIs (e.g., `["gnu", "musl", "msvc"]`).
- **`runner`** (string): Custom GitHub runner image for this target (optional).
- **`packages`** (list of strings): System packages to install before building.
- **`lang_opts`** (map): Language-specific options (see below).

## Language Options (`lang_opts`)

### Rust Engines
- **`profile`**: Cargo profile (`"release"`, `"dev"`, or custom).
- **`features`**: List or comma-separated string of Cargo features.
- **`tags`**: Alias for `features`.
- **`all-features`**: Boolean to enable all features.
- **`no-default-features`**: Boolean to disable default features.
- **`linker`**: Manual override for the linker executable.
- **`deployment_target`**: (macOS) Minimum OS version (default `"11.0"`).

## [naming]
Templates for output filenames. If not specified, the system applies standard defaults.

- **`binary`**: Format for raw binaries (default: `"{artifact}-{os}-{arch}{abi}"`).
- **`package`**: Format for distribution packages (default: `"{artifact}-{version}-{os}-{arch}{abi}.{ext}"`).

### Template Variables
- **`{artifact}`**: The technical name of the artifact.
- **`{os}`**: Target operating system (e.g., `linux`, `windows`).
- **`{arch}`**: Target architecture (e.g., `x86_64`, `aarch64`).
- **`{version}`**: Project version extracted automatically from the language manifest (e.g., `Cargo.toml`).
- **`{abi}`**: Target ABI. If present, it is automatically prefixed with a hyphen (e.g., `-musl`).
- **`{ext}`**: The appropriate file extension for the format and OS (e.g., `exe`, `deb`, `tar.gz`).

## Validation Rules
Refinery performs strict validation before any execution:
1. **Uniqueness**: Artifact names must be unique keys within the `artifacts` map.
2. **Mandatory Fields**: `project.name`, `project.lang`, and `os` in targets are non-optional.
3. **Target Integrity**: Each target must define at least one architecture in `archs`.
4. **Type Consistency**: `library_types` is only validated and used when `type = "lib"`.

## [metadata]
A key-value map of arbitrary strings. This data is not used by the core engine but is available for hooks or external post-processing tools.
