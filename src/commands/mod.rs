pub mod config;
pub mod edit;
pub mod init;
pub mod release;

use anyhow::Result;
use clap::{Parser, Subcommand};

#[derive(Parser)]
#[command(name = "refinery")]
#[command(about = "🦀 Refining Rust into universal artifacts", long_about = None)]
pub struct Cli {
    #[command(subcommand)]
    pub command: Commands,
}

#[derive(Subcommand)]
pub enum Commands {
    /// Initialize a new refinery project
    Init(init::InitArgs),
    /// Edit an existing refinery project
    Edit(edit::EditArgs),
    /// Manage project releases
    Release(release::ReleaseArgs),
    /// Manage global configuration
    Config(config::ConfigArgs),
}

pub async fn handle_command(cli: Cli) -> Result<()> {
    match cli.command {
        Commands::Init(args) => init::run(&args),
        Commands::Edit(args) => edit::run(&args),
        Commands::Release(args) => release::run(&args),
        Commands::Config(args) => config::run(&args),
    }
}
