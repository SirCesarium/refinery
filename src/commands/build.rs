use clap::Args;
use refinery_rs::core::builder::BuildManager;
use refinery_rs::core::project;
use refinery_rs::core::schema::RefineryConfig;

#[derive(Args, Debug)]
pub struct BuildArgs {
    #[arg(short, long)]
    pub target: Option<String>,
    #[arg(long)]
    pub release: bool,
    #[arg(long)]
    pub headers_only: bool,
}

pub fn run(args: &BuildArgs) -> anyhow::Result<()> {
    let config = RefineryConfig::load("refinery.toml")?;
    let builder = BuildManager::new(&config, args.release);

    project::sync_metadata(&config)?;

    if args.headers_only {
        builder.generate_headers()?;
        return Ok(());
    }

    if let Some(target_triple) = &args.target {
        builder.build_single(target_triple)?;
    } else {
        builder.build_all()?;
    }

    Ok(())
}
