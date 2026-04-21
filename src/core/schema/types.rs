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

impl fmt::Display for OS {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::Linux => write!(f, "linux"),
            Self::Windows => write!(f, "windows"),
            Self::Macos => write!(f, "macos"),
        }
    }
}

#[derive(Debug, Serialize, Deserialize, Clone, Copy, PartialEq, Eq)]
#[serde(rename_all = "lowercase")]
pub enum LibC {
    Gnu,
    Musl,
}

impl fmt::Display for LibC {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::Gnu => write!(f, "gnu"),
            Self::Musl => write!(f, "musl"),
        }
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

impl fmt::Display for Arch {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::X86_64 => write!(f, "x86_64"),
            Self::I686 => write!(f, "i686"),
            Self::Aarch64 => write!(f, "aarch64"),
        }
    }
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
