use serde::{Deserialize, Serialize};
use std::fmt;

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct Binary {
    pub name: String,
    #[serde(default = "default_binary_path")]
    pub path: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub out_name: Option<String>,
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub features: Vec<String>,
    #[serde(default, skip_serializing_if = "std::ops::Not::not")]
    pub no_default_features: bool,
}

fn default_binary_path() -> String {
    "src/main.rs".to_string()
}

#[derive(Debug, Serialize, Deserialize, Clone, Default)]
pub struct Library {
    pub name: String,
    #[serde(default = "default_library_path")]
    pub path: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub out_name: Option<String>,
    pub types: Vec<LibType>,
    #[serde(default, skip_serializing_if = "std::ops::Not::not")]
    pub headers: bool,
}

fn default_library_path() -> String {
    "src/lib.rs".to_string()
}

#[derive(Debug, Serialize, Deserialize, PartialEq, Eq, Clone)]
#[serde(rename_all = "lowercase")]
pub enum LibType {
    Static,
    Dynamic,
}

impl fmt::Display for LibType {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::Static => write!(f, "static"),
            Self::Dynamic => write!(f, "dynamic"),
        }
    }
}

impl Default for Binary {
    fn default() -> Self {
        Self {
            name: "my-project".into(),
            path: "src/main.rs".into(),
            out_name: None,
            features: vec![],
            no_default_features: false,
        }
    }
}
