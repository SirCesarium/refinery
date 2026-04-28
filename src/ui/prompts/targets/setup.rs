use crate::core::schema::{Abi, Arch, Binary, Library, OS, TargetMatrix, Targets};
use crate::ui::{Result, get_render_config, prompt_confirm};
use inquire::MultiSelect;

/// # Errors
/// Returns error if interactive prompts fail.
pub fn configure_targets(
    targets: &mut Targets,
    binaries: &[Binary],
    libraries: &[Library],
) -> Result<()> {
    let os_options = OS::all().to_vec();
    let selected_oss = MultiSelect::new("Select Target OS:", os_options)
        .with_render_config(get_render_config())
        .prompt()?;

    let artifacts_all: Vec<String> = binaries
        .iter()
        .map(|b| b.name.clone())
        .chain(libraries.iter().map(|l| l.name.clone()))
        .collect();

    for os in selected_oss {
        println!("\nConfiguring {os} Matrix:");

        let abis = Abi::for_os(os);
        let selected_abis = if abis.is_empty() {
            vec![None]
        } else {
            let chosen = MultiSelect::new(&format!("Select {os} ABI variants:"), abis.to_vec())
                .with_render_config(get_render_config())
                .prompt()?;
            if chosen.is_empty() {
                println!("Warning: No ABI selected for {os}. Skipping.");
                continue;
            }
            chosen.into_iter().map(Some).collect()
        };

        for abi in selected_abis {
            if let Some(a) = abi {
                println!("  - ABI: {a}");
            }
            let matrix = prompt_target_matrix(os, &artifacts_all)?;
            apply_matrix(targets, os, abi, matrix);
        }
    }

    Ok(())
}

fn prompt_target_matrix(os: OS, artifacts_all: &[String]) -> Result<TargetMatrix> {
    let arch_opts = Arch::for_os(os).to_vec();
    let archs = MultiSelect::new("Select Architectures:", arch_opts)
        .with_default(&[0]) // Default to first (usually x86_64)
        .with_render_config(get_render_config())
        .prompt()?;

    let artifacts = MultiSelect::new("Select Artifacts to build:", artifacts_all.to_vec())
        .with_default(&(0..artifacts_all.len()).collect::<Vec<_>>())
        .with_render_config(get_render_config())
        .prompt()?;

    let pkg_opts = match os {
        OS::Linux => vec!["deb", "rpm", "tar.gz"],
        OS::Windows => vec!["msi", "zip"],
        OS::Macos => vec!["tar.gz", "zip"],
    };
    let pkg = MultiSelect::new("Select Packaging Formats:", pkg_opts)
        .with_render_config(get_render_config())
        .prompt()?
        .into_iter()
        .map(String::from)
        .collect();

    let strip = if os == OS::Macos {
        false
    } else {
        prompt_confirm("Strip symbols from binaries?", false)?
    };

    let ext = if os == OS::Windows {
        let val: String = inquire::Text::new("Override windows extension:")
            .with_default(".exe")
            .with_render_config(get_render_config())
            .prompt()?;
        if val == ".exe" || val.is_empty() {
            None
        } else {
            Some(val)
        }
    } else {
        None
    };

    Ok(TargetMatrix {
        archs,
        artifacts,
        pkg,
        ext,
        cross_image: None,
        strip,
        overrides: None,
    })
}

fn apply_matrix(targets: &mut Targets, os: OS, abi: Option<Abi>, matrix: TargetMatrix) {
    match (os, abi) {
        (OS::Linux, Some(Abi::Gnu)) => {
            let mut l = targets.linux.take().unwrap_or_default();
            l.gnu = Some(matrix);
            targets.linux = Some(l);
        }
        (OS::Linux, Some(Abi::Musl)) => {
            let mut l = targets.linux.take().unwrap_or_default();
            l.musl = Some(matrix);
            targets.linux = Some(l);
        }
        (OS::Windows, Some(Abi::Msvc) | None) => {
            let mut w = targets.windows.take().unwrap_or_default();
            w.msvc = Some(matrix);
            targets.windows = Some(w);
        }
        (OS::Windows, Some(Abi::Gnu)) => {
            let mut w = targets.windows.take().unwrap_or_default();
            w.gnu = Some(matrix);
            targets.windows = Some(w);
        }
        (OS::Macos, _) => {
            targets.macos = Some(matrix);
        }
        _ => {}
    }
}
