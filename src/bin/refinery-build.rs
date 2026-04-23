#[path = "../commands/mod.rs"]
mod commands;

use crate::commands::{Cli, handle_command};
use clap::Parser;
use refinery_rs::ui::error;
use std::process;

#[tokio::main]
async fn main() {
    let cli = Cli::parse();

    if let Err(err) = handle_command(cli).await {
        error(&err);
        process::exit(1);
    }
}
