use crate::errors::{RefineryError, Result};
use serde::Deserialize;

#[derive(Debug, Deserialize)]
pub struct RefineryConfig {
    pub binaries: Vec<Binary>,
    pub libraries: Vec<Library>,
    pub targets: Vec<TargetMatrix>,
}

#[derive(Debug, Deserialize)]
pub struct Binary {
    pub name: String,
    pub path: String,
    #[serde(default)]
    pub features: Vec<String>,
    #[serde(default)]
    pub no_default_features: bool,
}

#[derive(Debug, Deserialize)]
pub struct Library {
    pub name: String,
    pub path: String,
    pub types: Vec<LibType>,
    #[serde(default)]
    pub headers: bool,
}

#[derive(Debug, Deserialize, PartialEq, Eq)]
#[serde(rename_all = "lowercase")]
pub enum LibType {
    Static,
    Dynamic,
}

#[derive(Debug, Deserialize)]
pub struct TargetMatrix {
    pub os: OS,
    #[serde(default)]
    pub libc: Option<LibC>,
    pub archs: Vec<Arch>,
    pub artifacts: Vec<String>,
    #[serde(default)]
    pub pkg: Vec<String>,
    #[serde(default)]
    pub ext: Option<String>,
}

#[derive(Debug, Deserialize, Clone, Copy, PartialEq, Eq)]
#[serde(rename_all = "lowercase")]
pub enum OS {
    Linux,
    Windows,
    Macos,
}

#[derive(Debug, Deserialize, Clone, Copy, PartialEq, Eq)]
#[serde(rename_all = "lowercase")]
pub enum LibC {
    Gnu,
    Musl,
}

#[derive(Debug, Deserialize, Clone, Copy, PartialEq, Eq)]
pub enum Arch {
    #[serde(alias = "x64", alias = "amd64", alias = "x86_64")]
    X86_64,
    #[serde(alias = "x32", alias = "x86", alias = "i686")]
    I686,
    #[serde(alias = "arm64", alias = "silicon", alias = "aarch64")]
    Aarch64,
}

impl Arch {
    #[must_use]
    pub const fn to_triple_part(&self) -> &'static str {
        match self {
            Self::X86_64 => "x86_64",
            Self::I686 => "i686",
            Self::Aarch64 => "aarch64",
        }
    }
}

impl TargetMatrix {
    /// Generates a list of official Rust target triples based on the matrix configuration.
    ///
    /// This method maps the high-level OS and Architecture definitions from the TOML
    /// configuration into specific strings recognized by `cargo` and `cross`.
    ///
    /// # Errors
    ///
    /// Returns a [`RefineryError::Config`] in the following cases:
    /// * The target OS is Linux but no `libc` (gnu/musl) was specified.
    /// * The target configuration is physically incompatible (e.g., x32 architecture on macOS).
    ///
    /// # Examples
    ///
    /// ```
    /// # use refinery_rs::core::schema::{TargetMatrix, OS, LibC, Arch};
    /// # use refinery_rs::errors::Result;
    /// # fn main() -> Result<()> {
    /// let matrix = TargetMatrix {
    ///     os: OS::Linux,
    ///     libc: Some(LibC::Musl),
    ///     archs: vec![Arch::X86_64],
    ///     artifacts: vec!["my-app".into()],
    ///     pkg: vec![],
    ///     ext: None,
    /// };
    ///
    /// let triples = matrix.get_triples()?;
    /// assert_eq!(triples, vec!["x86_64-unknown-linux-musl"]);
    /// # Ok(())
    /// # }
    /// ```
    pub fn get_triples(&self) -> Result<Vec<String>> {
        let mut triples = Vec::new();
        for arch in &self.archs {
            let triple = match (self.os, self.libc, arch) {
                (OS::Linux, Some(LibC::Gnu), a) => {
                    format!("{}-unknown-linux-gnu", a.to_triple_part())
                }
                (OS::Linux, Some(LibC::Musl), a) => {
                    format!("{}-unknown-linux-musl", a.to_triple_part())
                }
                (OS::Linux, None, _) => {
                    return Err(RefineryError::Config(
                        "Linux target requires libc (gnu/musl)".into(),
                    ));
                }
                (OS::Windows, _, Arch::X86_64) => "x86_64-pc-windows-msvc".into(),
                (OS::Windows, _, Arch::I686) => "i686-pc-windows-msvc".into(),
                (OS::Windows, _, Arch::Aarch64) => "aarch64-pc-windows-msvc".into(),
                (OS::Macos, _, Arch::X86_64) => "x86_64-apple-darwin".into(),
                (OS::Macos, _, Arch::Aarch64) => "aarch64-apple-darwin".into(),
                (OS::Macos, _, Arch::I686) => {
                    return Err(RefineryError::Config(
                        "macOS does not support x32 (i686) architecture".into(),
                    ));
                }
            };
            triples.push(triple);
        }
        Ok(triples)
    }
}

impl Default for RefineryConfig {
    fn default() -> Self {
        Self {
            binaries: vec![Binary::default()],
            libraries: vec![],
            targets: vec![
                TargetMatrix::default_linux(),
                TargetMatrix::default_windows(),
            ],
        }
    }
}

impl Default for Binary {
    fn default() -> Self {
        Self {
            name: "my-project".into(),
            path: "src/main.rs".into(),
            features: vec![],
            no_default_features: false,
        }
    }
}

impl TargetMatrix {
    fn default_linux() -> Self {
        Self {
            os: OS::Linux,
            libc: Some(LibC::Gnu),
            archs: vec![Arch::X86_64],
            artifacts: vec!["my-project".into()],
            pkg: vec!["deb".into()],
            ext: None,
        }
    }

    fn default_windows() -> Self {
        Self {
            os: OS::Windows,
            libc: None,
            archs: vec![Arch::X86_64],
            artifacts: vec!["my-project".into()],
            pkg: vec!["msi".into()],
            ext: Some(".exe".into()),
        }
    }
}
