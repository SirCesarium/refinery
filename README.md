# 🦀🏭 Refinery-RS (Legacy)

> Refining Rust into universal artifacts.

`refinery` is an orchestration tool designed to automate GitHub Actions workflows, manage `Cargo.toml` metadata, and streamline cross-compilation for Rust projects.

> [!IMPORTANT]
> This Rust implementation is preserved under the `v1.0-rust-legacy` tag. The project is migrating to **Go** for better YAML ecosystem stability and maintenance.
> The project has migrated to Go to escape the "dead-crate hell" of the Rust YAML ecosystem (e.g., the deprecation of serde_yaml). Moving to Go ensures long-term stability by using industry-standard tooling (yaml.v3) for orchestration, avoiding the maintenance burden of abandoned low-level dependencies

## Features

-   **`init`**: Interactive project setup and `refinery.toml` generation.
-   **`forge`**: Generates unified CI/CD pipelines (`refinery.yml`) for Linux, Windows, and macOS.
-   **`build`**: Orchestrates local builds and header generation via `cbindgen`.
-   **`setup`**: Configures environments for `deb`, `rpm`, and `msi` (WiX) installers.
-   **`release`**: Manages SemVer bumping and automated Git tagging.

## Why the Migration?

The transition to **Go** addresses:
*   **YAML Stability**: Leveraging `yaml.v3` to avoid the fragmentation of the Rust YAML ecosystem.
*   **Orchestration Focus**: Prioritizing simplicity in pipeline generation over low-level language constraints.

## License

This project is licensed under MIT.
