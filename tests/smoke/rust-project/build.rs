extern crate cbindgen;

use std::env;
use std::path::PathBuf;

fn main() {
    let crate_dir = match env::var("CARGO_MANIFEST_DIR") {
        Ok(v) => v,
        Err(_) => return,
    };

    let package_name = match env::var("CARGO_PKG_NAME") {
        Ok(v) => v,
        Err(_) => "smoke_lib".to_string(),
    };

    let output_file = PathBuf::from(&crate_dir).join(format!("{}.h", package_name));

    let config = cbindgen::Config {
        language: cbindgen::Language::C,
        ..Default::default()
    };

    match cbindgen::generate_with_config(crate_dir, config) {
        Ok(bindings) => {
            bindings.write_to_file(output_file);
        }
        Err(e) => {
            eprintln!("Error generating bindings: {}", e);
            std::process::exit(1);
        }
    }
}
