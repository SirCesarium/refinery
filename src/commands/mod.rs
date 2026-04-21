pub mod edit;
pub mod init;

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
    Init(init::InitArgs),
    Edit(edit::EditArgs),
}

pub async fn handle_command(cli: Cli) -> Result<()> {
    match cli.command {
        Commands::Init(args) => init::run(&args),
        Commands::Edit(args) => edit::run(&args),
    }
}
