#[must_use]
pub const fn get_icon(nf: &'static str, plain: &'static str) -> &'static str {
    if cfg!(feature = "nerd-fonts") {
        nf
    } else {
        plain
    }
}

pub const TICK: &str = get_icon("󰄬", "✓");
pub const WARN: &str = get_icon("󰀪", "!");
pub const INFO: &str = get_icon("󰋼", "i");
pub const SETUP: &str = get_icon("󰒓", "*");
pub const PROMPT: &str = get_icon("󱩔", ">");
pub const CHECKBOX_ON: &str = get_icon("󰄬 ", "[x] ");
pub const CHECKBOX_OFF: &str = get_icon("󰄱 ", "[ ] ");
pub const PACKAGE: &str = get_icon("󰏖", "p");
pub const LIB: &str = get_icon("󰈚", "L");
pub const ROCKET: &str = get_icon("󰓅", "^");
pub const FIRE: &str = get_icon("󰈸", "f");
pub const CRITICAL: &str = get_icon("󰅙", "X");
pub const BRANCH: &str = get_icon("󰘬", "b");
pub const FOLDER: &str = get_icon("󱆃", "/");
