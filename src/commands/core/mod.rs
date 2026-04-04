// @swt-disable max-repetition
use clap::{Args, Subcommand};

use crate::auto_dispatch;
use crate::cmd;

#[derive(Args)]
pub struct CoreArgs {
    #[command(subcommand)]
    pub action: CoreAction,
}

#[derive(Subcommand)]
pub enum CoreAction {
    Build,
}

pub mod build;

cmd!(core(args: CoreArgs) {
    auto_dispatch!(args.action, CoreAction, {
        Build
    })
});
