# Refinery-RS Build & Quality Actions

Surgical Rust tools for automated matrix builds and code quality (using `sweet`).

## Features

- **Refinery-RS Release**: Multi-target builds for Windows, Linux, macOS, and Android processors.
- **Refinery-RS CI**: Integrated code quality gate with [`sweet`](https://github.com/SirCesarium/sweet) (swt), `clippy`, and `rustfmt`.
- **Docker Ready**: Automatic binary-to-container pipeline for GitHub Packages (GHCR).

---

## 🚀 CI Workflow (Quality Gate)

```yaml
jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v5
      - uses: sircesarium/refinery-rs/ci@main
        with:
          enable-sweet: true
          enable-fmt: true
          enable-clippy: true
```

### CI Inputs

| Input             | Description                        | Default |
| ----------------- | ---------------------------------- | ------- |
| `enable-sweet`    | Run `swt` maintainability analysis | `true`  |
| `enable-clippy`   | Run Clippy lints                   | `true`  |
| `enable-fmt`      | Check code formatting              | `true`  |
| `sweet-threshold` | Custom directory/file for `swt`    | `.`     |

---

## 📦 Release Workflow (Matrix Build)

```yaml
jobs:
  release:
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        include:
          - target: x86_64-unknown-linux-gnu
            os: ubuntu-latest
          - target: x86_64-pc-windows-msvc
            os: windows-latest
    steps:
      - uses: actions/checkout@v5
      - uses: sircesarium/refinery-rs@main
        with:
          target: ${{ matrix.target }}
          publish-docker: true
          github-token: ${{ secrets.GITHUB_TOKEN }}
```

### Release & Docker Inputs

| Input               | Description                            | Default             |
| ------------------- | -------------------------------------- | ------------------- |
| `target`            | **Required**. The Rust target to build | -                   |
| `export-bins`       | Export binary executables              | `true`              |
| `export-libs`       | Export `.so`, `.dll`, `.dylib`, `.a`   | `true`              |
| `use-cross`         | Use `cross-rs` for building            | `false`             |
| `publish-docker`    | Build and push Docker image to GHCR    | `false`             |
| `docker-image-name` | Custom name for the docker image       | `github.repository` |
| `github-token`      | Token for GHCR authentication          | -                   |

---

## 🍬 [Sweet](https://github.com/SirCesarium/sweet) Integration

The CI action automatically installs and runs `swt`. If `swt` finds maintenance risks, it will emit a GitHub Action warning:
`🍬 Sweet identified potential maintainability issues in your code!`

---

## 🐳 Docker Deployment

If `publish-docker` is enabled and no `Dockerfile` is found in the repository, Refinery-RS generates a surgical **Debian-slim** image:

- **Base**: `debian:bookworm-slim`
- **Path**: `/usr/local/bin/app`
- **Tags**: `latest` and `github.sha`

---

## License

MIT / Apache-2.0
