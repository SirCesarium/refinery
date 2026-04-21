use anyhow::Result;
use clap::{Args, Subcommand};
use refinery_rs::core::config::Config;
use refinery_rs::ui::{info, success};

#[derive(Args, Debug)]
pub struct ConfigArgs {
    #[command(subcommand)]
    pub action: ConfigAction,
}

#[derive(Subcommand, Debug)]
pub enum ConfigAction {
    /// Show current configuration
    Show,
    /// Set a configuration value
    Set {
        /// Key to set (`main_branch`, `preferred_editor`, `auto_check_cargo`)
        key: String,
        /// Value to set
        value: String,
    },
    /// Reset configuration to defaults
    Reset,
}

pub fn run(args: &ConfigArgs) -> Result<()> {
    let mut config = Config::load()?;

    match &args.action {
        ConfigAction::Show => {
            info("Current Global Configuration:");
            println!("main_branch = \"{}\"", config.main_branch);
            println!("preferred_editor = {:?}", config.preferred_editor);
            println!("auto_check_cargo = {}", config.auto_check_cargo);
            println!("\nEffective Editor: {}", config.get_editor());
        }
        ConfigAction::Set { key, value } => {
            match key.as_str() {
                "main_branch" => config.main_branch.clone_from(value),
                "preferred_editor" => config.preferred_editor = Some(value.clone()),
                "auto_check_cargo" => {
                    config.auto_check_cargo = value
                        .parse()
                        .map_err(|_| anyhow::anyhow!("Invalid boolean value"))?;
                }
                _ => anyhow::bail!("Unknown configuration key: {key}"),
            }
            config.save_global()?;
            success(&format!("Successfully set {key} to {value}"));
        }
        ConfigAction::Reset => {
            let default_config = Config::default();
            default_config.save_global()?;
            success("Configuration reset to defaults.");
        }
    }

    Ok(())
}
