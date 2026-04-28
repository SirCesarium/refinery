use super::overrides::manage_overrides;
use super::setup::configure_targets;
use crate::core::schema::{Abi, Arch, OS, RefineryConfig};
use crate::prompt;
use crate::ui::{Result, get_render_config, prompt_confirm};

/// # Errors
/// Returns error if interactive prompts fail.
pub fn edit_targets(config: &mut RefineryConfig) -> Result<bool> {
    let mut changed = false;
    loop {
        let mut options = vec!["Add/Edit Target OS".to_string()];
        if let Some(ref l) = config.targets.linux {
            if l.gnu.is_some() {
                options.push("Linux (gnu)".into());
            }
            if l.musl.is_some() {
                options.push("Linux (musl)".into());
            }
        }
        if let Some(ref w) = config.targets.windows {
            if w.msvc.is_some() {
                options.push("Windows (msvc)".into());
            }
            if w.gnu.is_some() {
                options.push("Windows (gnu)".into());
            }
        }
        if config.targets.macos.is_some() {
            options.push("macOS".into());
        }
        options.push("Done".into());

        let choice = inquire::Select::new("Edit targets:", options)
            .with_render_config(get_render_config())
            .prompt()?;

        if choice == "Done" {
            break;
        }

        if choice.contains("Add/Edit Target OS") {
            configure_targets(&mut config.targets, &config.binaries, &config.libraries)?;
            changed = true;
        } else {
            let (os, abi) = if choice.contains("Linux (gnu)") {
                (OS::Linux, Some(Abi::Gnu))
            } else if choice.contains("Linux (musl)") {
                (OS::Linux, Some(Abi::Musl))
            } else if choice.contains("Windows (msvc)") {
                (OS::Windows, Some(Abi::Msvc))
            } else if choice.contains("Windows (gnu)") {
                (OS::Windows, Some(Abi::Gnu))
            } else if choice.contains("macOS") {
                (OS::Macos, None)
            } else {
                continue;
            };

            if edit_target_matrix_ui(config, os, abi)? {
                changed = true;
            }
        }
    }
    Ok(changed)
}

#[allow(clippy::too_many_lines)]
fn edit_target_matrix_ui(config: &mut RefineryConfig, os: OS, abi: Option<Abi>) -> Result<bool> {
    loop {
        let target = match (os, abi) {
            (OS::Linux, Some(Abi::Gnu)) => {
                config.targets.linux.as_ref().and_then(|l| l.gnu.as_ref())
            }
            (OS::Linux, Some(Abi::Musl)) => {
                config.targets.linux.as_ref().and_then(|l| l.musl.as_ref())
            }
            (OS::Windows, Some(Abi::Msvc)) => config
                .targets
                .windows
                .as_ref()
                .and_then(|w| w.msvc.as_ref()),
            (OS::Windows, Some(Abi::Gnu)) => {
                config.targets.windows.as_ref().and_then(|w| w.gnu.as_ref())
            }
            (OS::Macos, _) => config.targets.macos.as_ref(),
            _ => None,
        };

        let Some(matrix) = target else {
            return Ok(false);
        };

        let archs_label = format!("Architectures: {:?}", matrix.archs);
        let artifacts_label = format!("Artifacts: {:?}", matrix.artifacts);
        let pkg_label = format!("Packaging: {:?}", matrix.pkg);
        let ext_label = format!("Extension: {}", matrix.ext.as_deref().unwrap_or("auto"));
        let strip_label = format!("Strip symbols: {}", matrix.strip);
        let overrides_label = "Manage Name Overrides".to_string();
        let remove_label = "!!! Remove this Target !!!".to_string();
        let done_label = "Back".to_string();

        let fields = vec![
            archs_label.clone(),
            artifacts_label,
            pkg_label,
            ext_label,
            strip_label,
            overrides_label.clone(),
            remove_label.clone(),
            done_label.clone(),
        ];

        let field = inquire::Select::new(&format!("Edit {os} matrix:"), fields)
            .with_render_config(get_render_config())
            .prompt()?;

        if field == done_label {
            return Ok(true);
        }
        if field == remove_label {
            if prompt_confirm("Delete this target matrix?", false)? {
                match (os, abi) {
                    (OS::Linux, Some(Abi::Gnu)) => {
                        if let Some(mut l) = config.targets.linux.take() {
                            l.gnu = None;
                            config.targets.linux = Some(l);
                        }
                    }
                    (OS::Linux, Some(Abi::Musl)) => {
                        if let Some(mut l) = config.targets.linux.take() {
                            l.musl = None;
                            config.targets.linux = Some(l);
                        }
                    }
                    (OS::Windows, Some(Abi::Msvc)) => {
                        if let Some(mut w) = config.targets.windows.take() {
                            w.msvc = None;
                            config.targets.windows = Some(w);
                        }
                    }
                    (OS::Windows, Some(Abi::Gnu)) => {
                        if let Some(mut w) = config.targets.windows.take() {
                            w.gnu = None;
                            config.targets.windows = Some(w);
                        }
                    }
                    (OS::Macos, _) => config.targets.macos = None,
                    _ => {}
                }
                return Ok(true);
            }
            continue;
        }

        let target_mut = match (os, abi) {
            (OS::Linux, Some(Abi::Gnu)) => {
                config.targets.linux.as_mut().and_then(|l| l.gnu.as_mut())
            }
            (OS::Linux, Some(Abi::Musl)) => {
                config.targets.linux.as_mut().and_then(|l| l.musl.as_mut())
            }
            (OS::Windows, Some(Abi::Msvc)) => config
                .targets
                .windows
                .as_mut()
                .and_then(|w| w.msvc.as_mut()),
            (OS::Windows, Some(Abi::Gnu)) => {
                config.targets.windows.as_mut().and_then(|w| w.gnu.as_mut())
            }
            (OS::Macos, _) => config.targets.macos.as_mut(),
            _ => None,
        };

        let Some(matrix) = target_mut else {
            return Ok(false);
        };

        if field == archs_label {
            let options = Arch::for_os(os).to_vec();
            let defaults: Vec<usize> = matrix
                .archs
                .iter()
                .filter_map(|a| options.iter().position(|o| o == a))
                .collect();

            matrix.archs = inquire::MultiSelect::new("Select architectures:", options)
                .with_default(&defaults)
                .with_render_config(get_render_config())
                .prompt()?;
        } else if field.contains("Artifacts") {
            let mut options: Vec<String> = config.binaries.iter().map(|b| b.name.clone()).collect();
            options.extend(config.libraries.iter().map(|l| l.name.clone()));

            let defaults: Vec<usize> = matrix
                .artifacts
                .iter()
                .filter_map(|a| options.iter().position(|o| o == a))
                .collect();

            matrix.artifacts = inquire::MultiSelect::new("Select artifacts to build:", options)
                .with_default(&defaults)
                .with_render_config(get_render_config())
                .prompt()?;
        } else if field.contains("Packaging") {
            let options = match os {
                OS::Linux => vec!["deb", "rpm", "tar.gz"],
                OS::Windows => vec!["msi", "zip"],
                OS::Macos => vec!["tar.gz", "zip"],
            };
            let defaults: Vec<usize> = matrix
                .pkg
                .iter()
                .filter_map(|p| options.iter().position(|o| o == p))
                .collect();

            matrix.pkg = inquire::MultiSelect::new("Select packaging formats:", options)
                .with_default(&defaults)
                .with_render_config(get_render_config())
                .prompt()?
                .into_iter()
                .map(String::from)
                .collect();
        } else if field.contains("Extension") {
            let ext: String = prompt!("Override file extension (empty for default):")?;
            matrix.ext = if ext.trim().is_empty() {
                None
            } else {
                Some(ext.trim().to_string())
            };
        } else if field.contains("Strip") {
            matrix.strip = prompt_confirm("Strip symbols from binaries?", matrix.strip)?;
        } else if field == overrides_label {
            manage_overrides(matrix, &config.binaries, &config.libraries)?;
        }
    }
}
