# Refinery-RS Build & Quality Actions

Surgical Rust tools for automated matrix builds and code quality (using `sweet`).

## Features

- **Refinery-RS Release**: Build and package for Windows, Linux, macOS, and Android.
- **Refinery-RS CI**: Integrated code quality gate with `sweet` (swt), `clippy`, and `rustfmt`.

---

## 🚀 CI Workflow (Quality Gate)

Use this for your pull requests and pushes to `main`.

```yaml
jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: sircesarium/refinery-rs/ci@main
        with:
          enable-sweet: true
          enable-fmt: true
          enable-clippy: true
```

### CI Inputs

| Input | Description | Default |
|-------|-------------|---------|
| `enable-sweet` | Run `swt` maintainability analysis | `true` |
| `enable-clippy` | Run Clippy lints | `true` |
| `enable-fmt` | Check code formatting | `true` |
| `sweet-threshold` | Custom directory/file for `swt` | `.` |

---

## 📦 Release Workflow (Matrix Build)

Use this to generate artifacts and create GitHub releases.

```yaml
jobs:
  release:
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        target: [x86_64-unknown-linux-gnu, x86_64-pc-windows-msvc]
        include:
          - target: x86_64-unknown-linux-gnu
            os: ubuntu-latest
          - target: x86_64-pc-windows-msvc
            os: windows-latest
    steps:
      - uses: actions/checkout@v4
      - uses: sircesarium/refinery-rs@main
        with:
          target: ${{ matrix.target }}
          export-libs: true
```

### Release Inputs (Marketplace Action)

| Input | Description | Default |
|-------|-------------|---------|
| `target` | **Required**. The Rust target to build | - |
| `export-libs` | Export `.so`, `.dll`, `.dylib`, `.a` | `true` |
| `export-bins` | Export binary executables | `true` |
| `use-cross` | Use `cross-rs` for building | `false` |

---

## 🍬 Sweet Integration

The CI action automatically installs and runs `swt`. If `swt` finds maintenance risks, it will emit a GitHub Action warning:
`🍬 Sweet identified potential maintainability issues in your code!`

---

## License

MIT / Apache-2.0
