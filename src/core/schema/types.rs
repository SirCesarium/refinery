// @swt-disable max-repetition
use crate::errors::TargetError;
use serde::{Deserialize, Serialize};
use std::fmt;

#[derive(Debug, Serialize, Deserialize, Clone, Copy, PartialEq, Eq, Hash, Default)]
#[serde(rename_all = "lowercase")]
pub enum OS {
    #[default]
    Linux,
    Windows,
    Macos,
}

impl OS {
    #[must_use]
    pub const fn all() -> &'static [Self] {
        &[Self::Linux, Self::Windows, Self::Macos]
    }

    #[must_use]
    pub const fn to_triple_part(&self) -> &'static str {
        match self {
            Self::Linux => "unknown-linux",
            Self::Windows => "pc-windows",
            Self::Macos => "apple-darwin",
        }
    }
}

impl fmt::Display for OS {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::Linux => write!(f, "Linux"),
            Self::Windows => write!(f, "Windows"),
            Self::Macos => write!(f, "macOS"),
        }
    }
}

#[derive(Debug, Serialize, Deserialize, Clone, Copy, PartialEq, Eq, Hash)]
#[serde(rename_all = "lowercase")]
pub enum Abi {
    Gnu,
    Musl,
    Msvc,
}

impl Abi {
    #[must_use]
    pub const fn all() -> &'static [Self] {
        &[Self::Gnu, Self::Musl, Self::Msvc]
    }

    #[must_use]
    pub const fn for_os(os: OS) -> &'static [Self] {
        match os {
            OS::Linux => &[Self::Gnu, Self::Musl],
            OS::Windows => &[Self::Msvc, Self::Gnu],
            OS::Macos => &[],
        }
    }

    #[must_use]
    pub const fn to_triple_part(&self) -> &'static str {
        match self {
            Self::Gnu => "gnu",
            Self::Musl => "musl",
            Self::Msvc => "msvc",
        }
    }
}

impl fmt::Display for Abi {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}", self.to_triple_part())
    }
}

#[derive(Debug, Serialize, Deserialize, Clone, Copy, PartialEq, Eq, Hash)]
#[serde(rename_all = "lowercase")]
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
    pub const fn all() -> &'static [Self] {
        &[Self::X86_64, Self::I686, Self::Aarch64]
    }

    #[must_use]
    pub const fn for_os(os: OS) -> &'static [Self] {
        match os {
            OS::Macos => &[Self::X86_64, Self::Aarch64],
            _ => &[Self::X86_64, Self::I686, Self::Aarch64],
        }
    }

    #[must_use]
    pub const fn to_triple_part(&self) -> &'static str {
        match self {
            Self::X86_64 => "x86_64",
            Self::I686 => "i686",
            Self::Aarch64 => "aarch64",
        }
    }
}

impl fmt::Display for Arch {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}", self.to_triple_part())
    }
}

impl OS {
    /// Builds a standard Rust target triple.
    ///
    /// # Errors
    /// Returns a `TargetError` if the combination of parameters is not a supported Rust target.
    pub fn try_to_triple(&self, arch: Arch, abi: Option<Abi>) -> Result<String, TargetError> {
        let arch_part = arch.to_triple_part();

        match self {
            Self::Linux => {
                let abi = abi.ok_or_else(|| TargetError::MissingAbi(self.to_string()))?;
                match abi {
                    Abi::Gnu | Abi::Musl => Ok(format!("{arch_part}-unknown-linux-{abi}")),
                    Abi::Msvc => Err(TargetError::UnsupportedAbi(
                        abi.to_string(),
                        self.to_string(),
                    )),
                }
            }
            Self::Windows => {
                let abi = abi.ok_or_else(|| TargetError::MissingAbi(self.to_string()))?;
                match abi {
                    Abi::Msvc | Abi::Gnu => Ok(format!("{arch_part}-pc-windows-{abi}")),
                    Abi::Musl => Err(TargetError::UnsupportedAbi(
                        abi.to_string(),
                        self.to_string(),
                    )),
                }
            }
            Self::Macos => {
                if arch == Arch::I686 {
                    return Err(TargetError::UnsupportedArch(
                        arch.to_string(),
                        self.to_string(),
                    ));
                }
                // macOS triples (e.g. x86_64-apple-darwin) don't typically have an ABI suffix
                Ok(format!("{arch_part}-apple-darwin"))
            }
        }
    }
}
