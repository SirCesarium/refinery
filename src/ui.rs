//! User Interface and Interaction module.
//! Handles terminal output and interactive prompts.
#![allow(dead_code)]

use crate::errors::Result;
use console::Style;
use inquire::{Confirm, Text};

/// Internal macro to handle stylized printing.
#[macro_export]
macro_rules! log_step {
    ($icon:expr, $color:ident, $($arg:tt)*) => {
        println!("{} {}", $icon, console::Style::new().$color().bold().apply_to(format!($($arg)*)));
    };
}

/// Prints a success message with a green checkmark icon.
///
/// # Arguments
///
/// * `msg` - The success message string to display.
///
/// # Examples
///
/// ```rust
/// // Assuming ui module is in scope and imported correctly
/// // ui::success("Operation completed successfully!");
/// ```
pub fn success(msg: &str) {
    println!("[+] {}", Style::new().green().bold().apply_to(msg));
}

/// Prints an informational message with a cyan info icon.
///
/// # Arguments
///
/// * `msg` - The informational message string to display.
///
/// # Examples
///
/// ```rust
/// // Assuming ui module is in scope and imported correctly
/// // ui::info("Processing data...");
/// ```
pub fn info(msg: &str) {
    println!("[i] {}", Style::new().cyan().apply_to(msg));
}

/// Prints a warning message with a yellow exclamation mark icon.
///
/// # Arguments
///
/// * `msg` - The warning message string to display.
///
/// # Examples
///
/// ```rust
/// // Assuming ui module is in scope and imported correctly
/// // ui::warn("Configuration file not found, using defaults.");
/// ```
pub fn warn(msg: &str) {
    println!("[!] {}", Style::new().yellow().apply_to(msg));
}

/// Prompts the user for a confirmation with a default value.
///
/// # Arguments
///
/// * `msg` - The confirmation question string to display.
/// * `default` - The default boolean value (true or false) for the prompt.
///
/// # Returns
///
/// A `Result` containing the user's boolean answer, or an `inquire::InquireError` if
/// the prompt fails.
///
/// # Examples
///
/// ```rust
/// // Assuming ui module is in scope and imported correctly
/// // let confirm = ui::prompt_confirm("Are you sure?", true)?;
/// ```
pub fn prompt_confirm(msg: &str, default: bool) -> Result<bool> {
    Ok(Confirm::new(msg).with_default(default).prompt()?)
}

/// Prompts the user for text input with a default value.
///
/// # Arguments
///
/// * `msg` - The prompt question string to display.
/// * `default` - The default string value to pre-fill the input with.
///
/// # Returns
///
/// A `Result` containing the user's input string, or an `inquire::InquireError` if
/// the prompt fails.
///
/// # Examples
///
/// ```rust
/// // Assuming ui module is in scope and imported correctly
/// // let name = ui::prompt_text("Enter your name:", "Guest")?;
/// ```
pub fn prompt_text(msg: &str, default: &str) -> Result<String> {
    Ok(Text::new(msg).with_default(default).prompt()?)
}

/// Prints the ASCII banner for the CLI.
///
/// This function displays the project's logo and a tagline.
pub fn print_banner() {
    let orange = Style::new().color256(208);
    let banner = r"
    ____       _____                      
   / __ \___  / __(_)___  ___  _______  __
  / /_/ / _ \/ /_/ / __ \/ _ \/ ___/ / / /
 / _, _/  __/ __/ / / / /  __/ /  / /_/ / 
/_/ |_|\___/_/ /_/_/ /_/\___/_/   \__, /  
                                 /____/  
    ";
    println!("{}", orange.apply_to(banner));
    println!(
        "🦀 Refining Rust into universal artifacts.
"
    );
}
