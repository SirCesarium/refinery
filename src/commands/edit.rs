use anyhow::{Result, anyhow};
use clap::Args;
use refinery_rs::core::schema::RefineryConfig;
use refinery_rs::ui::prompts::{edit_binaries, edit_libraries, edit_targets, select_main_action};
use refinery_rs::ui::{print_banner, print_highlighted_toml, prompt_confirm, success};
use std::path::Path;

#[derive(Args)]
pub struct EditArgs {}

pub fn run(_args: &EditArgs) -> Result<()> {
    print_banner();
    let config_path = Path::new("refinery.toml");
    if !config_path.exists() {
        return Err(anyhow!("No refinery.toml found. Run 'init' first."));
    }
    let mut config = RefineryConfig::load(config_path)?;
    let mut saved = true;

    loop {
        let action = select_main_action()?;

        match action.as_str() {
            "Binaries" => {
                if edit_binaries(&mut config)? {
                    saved = false;
                }
            }
            "Libraries" => {
                if edit_libraries(&mut config)? {
                    saved = false;
                }
            }
            "Targets" => {
                if edit_targets(&mut config)? {
                    saved = false;
                }
            }
            "Review & Save" => {
                config.validate()?;
                let toml = config.to_toml()?;
                println!("\n--- Preview ---");
                print_highlighted_toml(&toml);
                if prompt_confirm("Overwrite refinery.toml?", true)? {
                    config.save(config_path)?;
                    success("Configuration updated.");
                    break;
                }
            }
            "Exit" => {
                if saved || prompt_confirm("Exit without saving changes?", false)? {
                    break;
                }
            }
            _ => break,
        }
    }
    Ok(())
}
