use std::collections::HashMap;

pub const CHECKOUT: &str = "actions/checkout@v6";
pub const RUST_TOOLCHAIN: &str = "dtolnay/rust-toolchain@stable";
pub const RUST_CACHE: &str = "Swatinem/rust-cache@v2";
pub const SOFTPROPS_RELEASE: &str = "softprops/action-gh-release@v3";

pub const UPLOAD_ARTIFACT: &str = "actions/upload-artifact@v7";
pub const DOWNLOAD_ARTIFACT: &str = "actions/download-artifact@v8";

pub const CROSS_GITHUB: &str = "https://github.com/cross-rs/cross";

pub const SWEET_REPO: &str = "https://github.com/SirCesarium/sweet";
pub const SWEET_DEFAULT_VERSION: &str = "v5.0.0-rc.4";
pub const SWEET_BINARY: &str = "swt_x86_64-unknown-linux-musl";

#[must_use]
pub fn default_permissions() -> HashMap<String, String> {
    let mut p = HashMap::new();
    p.insert("contents".into(), "write".into());
    p.insert("pull-requests".into(), "write".into());
    p
}
