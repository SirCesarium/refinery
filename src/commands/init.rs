use anyhow::Result;
use clap::Args;
use refinery_rs::core::project;
use refinery_rs::core::schema::{RefineryConfig, Targets};
use refinery_rs::ui::prompts::{
    configure_binaries, configure_libraries, configure_targets, select_init_action,
};
use refinery_rs::ui::{print_banner, print_highlighted_toml, prompt_confirm, success, warn};
use std::path::{Path, PathBuf};

#[derive(Args)]
pub struct InitArgs {
    #[arg(short, long)]
    pub force: bool,
}

pub fn run(args: &InitArgs) -> Result<()> {
    print_banner();
    let config_path = PathBuf::from("refinery.toml");
    if config_path.exists() && !args.force {
        anyhow::bail!(
            "Configuration file 'refinery.toml' already exists. Use --force to overwrite."
        );
    }

    let default_name = RefineryConfig::try_get_default_project_name()?;
    let mut config = RefineryConfig {
        refinery_version: Some(env!("CARGO_PKG_VERSION").to_string()),
        binaries: vec![],
        libraries: vec![],
        targets: Targets::default(),
        fail_fast: true,
    };

    loop {
        let selection = select_init_action()?;

        match selection.as_str() {
            "Add Binaries" => configure_binaries(&mut config, &default_name)?,
            "Add Libraries" => configure_libraries(&mut config, &default_name)?,
            "Configure Targets" => {
                if config.binaries.is_empty() && config.libraries.is_empty() {
                    println!();
                    warn("Add an artifact (binary or library) before configuring targets.");
                } else {
                    configure_targets(&mut config.targets, &config.binaries, &config.libraries)?;
                }
            }
            "Review & Save" => {
                if handle_save(&config, &config_path)? {
                    break;
                }
            }
            _ => unreachable!(),
        }
    }

    Ok(())
}

fn handle_save(config: &RefineryConfig, path: &Path) -> Result<bool> {
    config.validate()?;
    let toml = config.to_toml()?;
    println!("\n--- Preview ---");
    print_highlighted_toml(&toml);
    println!();
    if prompt_confirm("Save to refinery.toml?", true)? {
        config.save(path)?;
        println!();
        success("Project initialized successfully.");

        if !config.targets.is_empty() && prompt_confirm("Generate GitHub Actions workflow?", true)?
        {
            project::generate_workflow(config)?;
            success("GitHub Actions workflow generated at .github/workflows/refinery.yml");
        }
        return Ok(true);
    }
    Ok(false)
}
