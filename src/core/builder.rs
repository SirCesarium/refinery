// @swt-disable max-lines
use crate::core::project;
use crate::core::schema::{Abi, OS, RefineryConfig, TargetMatrix};
use crate::errors::{RefineryError, Result};
use std::fs;
use std::path::Path;
use std::process::Command;
use toml_edit::DocumentMut;

/// Information about a build target.
pub struct TargetInfo {
    /// The target triple (e.g., x86_64-unknown-linux-gnu).
    pub triple: String,
    /// The target operating system.
    pub os: OS,
    /// The target ABI (if applicable).
    pub abi: Option<Abi>,
    /// The packaging matrix for this target.
    pub matrix: TargetMatrix,
}

/// Manages the build process for different targets.
pub struct BuildManager<'a> {
    /// The project configuration.
    pub config: &'a RefineryConfig,
    /// Whether to build in release mode.
    pub release: bool,
}

impl<'a> BuildManager<'a> {
    /// Creates a new `BuildManager`.
    #[must_use]
    pub const fn new(config: &'a RefineryConfig, release: bool) -> Self {
        Self { config, release }
    }

    /// Builds all configured targets.
    ///
    /// # Errors
    /// Returns an error if any target fails to build or package.
    pub fn build_all(&self) -> Result<()> {
        let targets = self.collect_targets_info()?;
        for info in targets {
            self.build_target(&info)?;
        }
        Ok(())
    }

    /// Builds a single target specified by its triple.
    ///
    /// # Errors
    /// Returns an error if the target is not found or fails to build.
    pub fn build_single(&self, target_triple: &str) -> Result<()> {
        let targets = self.collect_targets_info()?;
        let info = targets
            .into_iter()
            .find(|t| t.triple == target_triple)
            .ok_or_else(|| anyhow::anyhow!("Target {target_triple} not found in configuration"))?;
        self.build_target(&info)
    }

    /// Generates C headers for libraries that have them enabled.
    ///
    /// # Errors
    /// Returns an error if `cbindgen` fails to generate headers.
    pub fn generate_headers(&self) -> Result<()> {
        for lib in &self.config.libraries {
            if lib.headers {
                let mut cmd = Command::new("cbindgen");
                cmd.arg("--output").arg(format!("{}.h", lib.name));
                cmd.arg(&lib.path);

                if Path::new("cbindgen.toml").exists() {
                    cmd.arg("--config").arg("cbindgen.toml");
                } else {
                    cmd.arg("--lang").arg("c");
                }

                let status = cmd.status()?;
                if !status.success() {
                    return Err(RefineryError::Generic(anyhow::anyhow!(
                        "Failed to generate headers for {}",
                        lib.name
                    )));
                }
            }
        }
        Ok(())
    }

    fn collect_targets_info(&self) -> Result<Vec<TargetInfo>> {
        let mut infos = Vec::new();

        if let Some(linux) = &self.config.targets.linux {
            if let Some(gnu) = &linux.gnu {
                for triple in gnu.get_triples(OS::Linux, Some(Abi::Gnu))? {
                    infos.push(TargetInfo {
                        triple,
                        os: OS::Linux,
                        abi: Some(Abi::Gnu),
                        matrix: gnu.clone(),
                    });
                }
            }
            if let Some(musl) = &linux.musl {
                for triple in musl.get_triples(OS::Linux, Some(Abi::Musl))? {
                    infos.push(TargetInfo {
                        triple,
                        os: OS::Linux,
                        abi: Some(Abi::Musl),
                        matrix: musl.clone(),
                    });
                }
            }
        }

        if let Some(windows) = &self.config.targets.windows {
            if let Some(msvc) = &windows.msvc {
                for triple in msvc.get_triples(OS::Windows, Some(Abi::Msvc))? {
                    infos.push(TargetInfo {
                        triple,
                        os: OS::Windows,
                        abi: Some(Abi::Msvc),
                        matrix: msvc.clone(),
                    });
                }
            }
            if let Some(gnu) = &windows.gnu {
                for triple in gnu.get_triples(OS::Windows, Some(Abi::Gnu))? {
                    infos.push(TargetInfo {
                        triple,
                        os: OS::Windows,
                        abi: Some(Abi::Gnu),
                        matrix: gnu.clone(),
                    });
                }
            }
        }

        if let Some(macos) = &self.config.targets.macos {
            for triple in macos.get_triples(OS::Macos, None)? {
                infos.push(TargetInfo {
                    triple,
                    os: OS::Macos,
                    abi: None,
                    matrix: macos.clone(),
                });
            }
        }

        Ok(infos)
    }

    fn build_target(&self, info: &TargetInfo) -> Result<()> {
        project::setup_toolchain(&info.triple)?;

        let tool = info.matrix.tool.as_deref().unwrap_or_else(|| {
            if info.triple.contains("linux")
                && (info.triple.contains("musl") || !info.triple.starts_with("x86_64"))
            {
                "cross"
            } else {
                "cargo"
            }
        });

        for artifact in &info.matrix.artifacts {
            let mut cmd = match tool {
                "cross" => {
                    let mut c = Command::new("cross");
                    c.arg("build");
                    if let Some(image) = &info.matrix.cross_image {
                        c.env("CROSS_IMAGE", image);
                    }
                    c
                }
                "zigbuild" => {
                    let mut c = Command::new("cargo");
                    c.arg("zigbuild");
                    c
                }
                _ => {
                    let mut c = Command::new("cargo");
                    c.arg("build");
                    c
                }
            };

            cmd.arg("--target").arg(&info.triple);
            cmd.arg("--bin").arg(artifact);

            if self.release {
                cmd.arg("--release");
            }

            let status = cmd.status().map_err(RefineryError::Io)?;
            if !status.success() {
                return Err(RefineryError::Generic(anyhow::anyhow!(
                    "Failed to build artifact {artifact} for target: {}",
                    info.triple
                )));
            }
        }

        for pkg_format in &info.matrix.pkg {
            Self::validate_packaging_metadata(pkg_format)?;
            self.package_target(info, pkg_format)?;
            Self::rename_linux_packages(info, pkg_format)?;
        }

        Ok(())
    }

    fn rename_linux_packages(info: &TargetInfo, format: &str) -> Result<()> {
        if info.os != OS::Linux || (format != "deb" && format != "rpm") {
            return Ok(());
        }

        let abi_suffix = if info.triple.contains("musl") {
            "musl"
        } else {
            "gnu"
        };

        if let Ok(entries) = fs::read_dir("target") {
            for entry in entries.flatten() {
                let path = entry.path();
                if let Some(ext) = path.extension()
                    && ext.to_str() == Some(format)
                {
                    let file_name = path.file_stem().and_then(|s| s.to_str()).unwrap_or("");
                    if !file_name.contains(abi_suffix) {
                        let new_name = format!("{file_name}-{abi_suffix}.{format}");
                        let mut new_path = path.clone();
                        new_path.set_file_name(new_name);
                        fs::rename(&path, &new_path).map_err(RefineryError::Io)?;
                    }
                }
            }
        }
        Ok(())
    }

    fn validate_packaging_metadata(format: &str) -> Result<()> {
        let cargo_content = fs::read_to_string("Cargo.toml").map_err(RefineryError::Io)?;
        let cargo_toml = cargo_content
            .parse::<DocumentMut>()
            .map_err(RefineryError::Toml)?;

        let metadata_key = match format {
            "deb" => "deb",
            "rpm" => "generate-rpm",
            _ => return Ok(()),
        };

        if cargo_toml
            .get("package")
            .and_then(|p| p.get("metadata"))
            .and_then(|m| m.get(metadata_key))
            .is_none()
        {
            return Err(RefineryError::Generic(anyhow::anyhow!(
                "Missing [package.metadata.{metadata_key}] in Cargo.toml. Run 'refinery setup' first."
            )));
        }
        Ok(())
    }

    fn package_target(&self, info: &TargetInfo, format: &str) -> Result<()> {
        match format {
            "deb" => Self::run_cargo_deb(info),
            "msi" => Self::run_cargo_wix(info),
            "rpm" => Self::run_cargo_generate_rpm(info),
            "tar.gz" => self.run_archive_tar(info),
            "zip" => self.run_archive_zip(info),
            _ => Ok(()),
        }
    }

    fn run_cargo_deb(info: &TargetInfo) -> Result<()> {
        let mut cargo_toml_modified = false;
        let original_toml = fs::read_to_string("Cargo.toml").map_err(RefineryError::Io)?;

        if info.triple.contains("musl") {
            let profile = "release";
            let mut lib_found = false;
            if let Ok(entries) = fs::read_dir(format!("target/{}/{}", info.triple, profile)) {
                for entry in entries.flatten() {
                    if let Some(name) = entry.file_name().to_str()
                        && Path::new(name)
                            .extension()
                            .is_some_and(|ext| ext.eq_ignore_ascii_case("so"))
                    {
                        lib_found = true;
                        break;
                    }
                }
            }
            if !lib_found {
                let mut doc = original_toml
                    .parse::<DocumentMut>()
                    .map_err(RefineryError::Toml)?;
                if let Some(lib) = doc.remove("lib") {
                    doc.insert("lib-disabled", lib);
                    fs::write("Cargo.toml", doc.to_string()).map_err(RefineryError::Io)?;
                    cargo_toml_modified = true;
                }
            }
        }

        let status = Command::new("cargo")
            .arg("deb")
            .arg("--target")
            .arg(&info.triple)
            .arg("--no-build")
            .arg("--no-strip")
            .status()
            .map_err(RefineryError::Io);

        if cargo_toml_modified {
            fs::write("Cargo.toml", original_toml).map_err(RefineryError::Io)?;
        }
        let s = status?;
        if !s.success() {
            return Err(RefineryError::Generic(anyhow::anyhow!(
                "Failed to generate .deb for {}",
                info.triple
            )));
        }
        Ok(())
    }

    fn run_cargo_wix(info: &TargetInfo) -> Result<()> {
        let original_toml = fs::read_to_string("Cargo.toml").map_err(RefineryError::Io)?;
        let mut cargo_toml = original_toml
            .parse::<DocumentMut>()
            .map_err(RefineryError::Toml)?;

        let version_str = cargo_toml["package"]["version"]
            .as_str()
            .ok_or_else(|| anyhow::anyhow!("Missing version in Cargo.toml"))?
            .to_string();

        // WiX requires numeric versions and Cargo metadata requires 3-part SemVer.
        // Map 1.0.0-rc.2 -> 1.0.2 temporarily.
        let modified = if version_str.contains("-rc.") {
            let parts: Vec<&str> = version_str.split('-').collect();
            let rc_part = parts[1].replace("rc.", "");
            let base_parts: Vec<&str> = parts[0].split('.').collect();

            if base_parts.len() >= 2 {
                let wix_version = format!("{}.{}.{}", base_parts[0], base_parts[1], rc_part);
                cargo_toml["package"]["version"] = toml_edit::value(wix_version);
                fs::write("Cargo.toml", cargo_toml.to_string()).map_err(RefineryError::Io)?;
                true
            } else {
                false
            }
        } else {
            false
        };

        let status = Command::new("cargo")
            .arg("wix")
            .arg("--target")
            .arg(&info.triple)
            .status();

        if modified {
            fs::write("Cargo.toml", original_toml).map_err(RefineryError::Io)?;
        }

        let s = status.map_err(RefineryError::Io)?;
        if !s.success() {
            return Err(RefineryError::Generic(anyhow::anyhow!(
                "Failed to generate .msi for {}",
                info.triple
            )));
        }
        Ok(())
    }

    fn run_cargo_generate_rpm(info: &TargetInfo) -> Result<()> {
        let status = Command::new("cargo")
            .arg("generate-rpm")
            .arg("--target")
            .arg(&info.triple)
            .status()
            .map_err(RefineryError::Io)?;
        if !status.success() {
            return Err(RefineryError::Generic(anyhow::anyhow!(
                "Failed to generate .rpm for {}",
                info.triple
            )));
        }
        Ok(())
    }

    fn run_archive_tar(&self, info: &TargetInfo) -> Result<()> {
        let profile = if self.release { "release" } else { "debug" };
        let base_path = format!("target/{}/{}", info.triple, profile);
        let archive_name = format!("target/{}/{}.tar.gz", info.triple, info.triple);
        let mut args = vec!["-czf", &archive_name, "-C", &base_path];
        for bin in &self.config.binaries {
            if info.matrix.artifacts.contains(&bin.name) {
                args.push(&bin.name);
            }
        }
        let status = Command::new("tar")
            .args(&args)
            .status()
            .map_err(RefineryError::Io)?;
        if !status.success() {
            return Err(RefineryError::Generic(anyhow::anyhow!(
                "Failed to create .tar.gz for {}",
                info.triple
            )));
        }
        Ok(())
    }

    fn run_archive_zip(&self, info: &TargetInfo) -> Result<()> {
        let profile = if self.release { "release" } else { "debug" };
        let base_path = format!("target/{}/{}", info.triple, profile);
        let archive_name = format!("{}.zip", info.triple);

        let mut cmd = if cfg!(windows) {
            let mut c = Command::new("tar");
            c.arg("-a")
                .arg("-c")
                .arg("-f")
                .arg(format!("../../{archive_name}"));
            c
        } else {
            let mut c = Command::new("zip");
            c.arg("-j").arg(format!("../../{archive_name}"));
            c
        };

        for bin in &self.config.binaries {
            if info.matrix.artifacts.contains(&bin.name) {
                cmd.arg(if info.triple.contains("windows") {
                    format!("{}.exe", bin.name)
                } else {
                    bin.name.clone()
                });
            }
        }

        let status = cmd
            .current_dir(Path::new(&base_path))
            .status()
            .map_err(RefineryError::Io)?;
        if !status.success() {
            return Err(RefineryError::Generic(anyhow::anyhow!(
                "Failed to create .zip for {}",
                info.triple
            )));
        }
        fs::rename(
            format!("target/{archive_name}"),
            format!("target/{}/{archive_name}", info.triple),
        )
        .map_err(RefineryError::Io)?;
        Ok(())
    }
}
