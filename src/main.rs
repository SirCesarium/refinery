//! Refinery-RS

#![deny(
    clippy::panic,
    clippy::unwrap_used,
    clippy::expect_used,
    clippy::pedantic,
    clippy::absolute_paths,
    missing_docs
)]

mod errors;
mod ui;

fn main() {
    ui::print_banner();
}
