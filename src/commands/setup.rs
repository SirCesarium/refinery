use anyhow::Result;
use clap::Args;
use refinery_rs::core::project;
use refinery_rs::core::schema::RefineryConfig;
use refinery_rs::ui::prompts::{configure_libraries, install, installers};
use refinery_rs::ui::{icons, prompt_confirm, success, warn};
use refinery_rs::{log_step, prompt, prompt_multi};

#[derive(Args, Debug)]
pub struct SetupArgs {
    #[arg(short, long)]
    pub force: bool,
}

pub fn run(args: &SetupArgs) -> Result<()> {
    let mut config = RefineryConfig::load("refinery.toml")?;

    let options = vec![
        "Pipeline (refinery.yml)".to_string(),
        "Quality Gate (ci.yml)".to_string(),
        "Installers (WiX, deb, rpm)".to_string(),
        "Library Setup".to_string(),
    ];
    let selections: Vec<String> = prompt_multi!("What would you like to setup?", options)?;

    for selection in selections {
        match selection.as_str() {
            "Pipeline (refinery.yml)" => setup_pipeline(&config)?,
            "Quality Gate (ci.yml)" => setup_quality_gate()?,
            "Installers (WiX, deb, rpm)" => installers::setup_installers(&config, args.force)?,
            "Library Setup" => setup_lib(&mut config)?,
            _ => unreachable!(),
        }
    }

    Ok(())
}

fn setup_pipeline(config: &RefineryConfig) -> Result<()> {
    log_step!(
        icons::TICK,
        Green,
        "Configuring unified refinery pipeline..."
    );
    project::generate_workflow(config)?;
    success("Unified pipeline generated at .github/workflows/refinery.yml");
    Ok(())
}

fn setup_quality_gate() -> Result<()> {
    println!("\n{} CI Quality Gate Setup", icons::SETUP);
    println!("  - Sweet: Maintainability analysis (nesting, length, duplication)");
    println!("  - Format: Ensures code follows rustfmt standards");
    println!("  - Clippy: Lints for common mistakes and idiomatic improvements");
    println!("  - Tests: Executes your full test suite\n");

    let options = vec![
        "Sweet (Maintainability)".into(),
        "Format (rustfmt)".into(),
        "Clippy (Lints)".into(),
        "Tests (cargo test)".into(),
    ];
    let checks: Vec<String> = prompt_multi!("Select checks to include in ci.yml:", options)?;

    if checks.is_empty() {
        warn("No checks selected. Quality Gate skipped.");
        return Ok(());
    }

    let clippy_flags: String = if checks.iter().any(|c| c.contains("Clippy")) {
        let f: String = prompt!("Clippy flags (default: -- -D warnings):")?;
        if f.trim().is_empty() {
            "-- -D warnings".to_string()
        } else {
            f.trim().to_string()
        }
    } else {
        String::new()
    };

    project::generate_quality_gate(&checks, &clippy_flags)?;
    success("Custom Quality Gate generated at .github/workflows/ci.yml");
    Ok(())
}

fn setup_lib(config: &mut RefineryConfig) -> Result<()> {
    if config.libraries.is_empty() {
        warn("No libraries defined in refinery.toml.");
        if prompt_confirm("Add a library configuration now?", true)? {
            let default_name = RefineryConfig::try_get_default_project_name()?;
            configure_libraries(config, &default_name)?;
            config.save("refinery.toml")?;
        } else {
            return Ok(());
        }
    }

    for lib in config.libraries.clone() {
        log_step!(icons::LIB, Yellow, "Configuring library: {}...", lib.name);

        project::sync_metadata(config)?;
        success("Cargo.toml configured for library export.");

        project::generate_lib_boilerplate(&lib)?;
        success(&format!("Boilerplate created at {}", lib.path));

        if lib.headers {
            install::check_and_install("cbindgen", "cbindgen")?;
            success("cbindgen ready for header generation.");
        }
    }

    Ok(())
}
