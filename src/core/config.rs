use crate::core::project;
use anyhow::{Context, Result};
use serde::{Deserialize, Serialize};
use std::env;
use std::fs;
use std::path::{Path, PathBuf};
use toml_edit::de::from_str;
use toml_edit::ser::to_string_pretty;

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct Config {
    pub main_branch: String,
    pub preferred_editor: Option<String>,
    pub auto_check_cargo: bool,
}

impl Default for Config {
    fn default() -> Self {
        Self {
            main_branch: "main".to_string(),
            preferred_editor: None,
            auto_check_cargo: true,
        }
    }
}

#[derive(Debug, Deserialize, Default)]
struct LocalConfig {
    main_branch: Option<String>,
    preferred_editor: Option<String>,
    auto_check_cargo: Option<bool>,
}

#[derive(Debug, Deserialize)]
struct LocalConfigWrapper {
    release: Option<LocalConfig>,
}

impl Config {
    /// Loads the configuration by merging global and local settings.
    ///
    /// # Errors
    /// Returns an error if the configuration files exist but are invalid TOML.
    pub fn load() -> Result<Self> {
        let mut config = Self::load_global().unwrap_or_default();

        let local_config = Self::load_local()?;
        if let Some(branch) = local_config.main_branch {
            config.main_branch = branch;
        }
        if let Some(editor) = local_config.preferred_editor {
            config.preferred_editor = Some(editor);
        }
        if let Some(check) = local_config.auto_check_cargo {
            config.auto_check_cargo = check;
        }

        Ok(config)
    }

    fn load_global() -> Result<Self> {
        let path = global_config_path()?;
        if !path.exists() {
            return Ok(Self::default());
        }
        let content = fs::read_to_string(path)?;
        from_str(&content).context("Failed to parse global config")
    }

    fn load_local() -> Result<LocalConfig> {
        let path = Path::new("refinery.toml");
        if !path.exists() {
            return Ok(LocalConfig::default());
        }
        let content = fs::read_to_string(path)?;

        let wrapper: LocalConfigWrapper =
            from_str(&content).context("Failed to parse local refinery.toml")?;
        Ok(wrapper.release.unwrap_or_default())
    }

    /// Saves the current configuration to the global configuration file.
    ///
    /// # Errors
    /// Returns an error if the configuration directory cannot be created or the file cannot be written.
    pub fn save_global(&self) -> Result<()> {
        let path = global_config_path()?;
        if let Some(parent) = path.parent() {
            fs::create_dir_all(parent)?;
        }
        let content = to_string_pretty(self)?;
        fs::write(path, content).context("Failed to save global config")
    }

    /// Returns the preferred editor, either from config, environment variables, or auto-detection.
    #[must_use]
    pub fn get_editor(&self) -> String {
        self.preferred_editor
            .clone()
            .or_else(|| env::var("VISUAL").ok())
            .or_else(|| env::var("EDITOR").ok())
            .unwrap_or_else(|| {
                if project::check_command("nvim") {
                    "nvim".to_string()
                } else if project::check_command("vim") {
                    "vim".to_string()
                } else if project::check_command("code") {
                    "code --wait".to_string()
                } else if project::check_command("nano") {
                    "nano".to_string()
                } else {
                    "vi".to_string()
                }
            })
    }
}

fn global_config_path() -> Result<PathBuf> {
    let home = env::var("HOME").context("Could not find HOME directory")?;
    Ok(PathBuf::from(home).join(".config/refinery/config.toml"))
}
