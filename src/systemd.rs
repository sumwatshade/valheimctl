
use anyhow::{bail, Context, Result};
use std::process::Command;

#[derive(Debug, Clone)]
pub struct ServiceStatus {
    pub active: bool,
    pub state: String,
    pub description: String,
}

fn validate_unit(unit: &str) -> Result<()> {
    anyhow::ensure!(!unit.is_empty(), "Empty systemd unit");
    anyhow::ensure!(
        !unit.starts_with('-')
            && unit.chars().all(|c| {
                c.is_ascii_alphanumeric()
                    || matches!(c, '@' | '.' | '_' | ':' | '-')
            }),
        "Invalid systemd unit name"
    );
    Ok(())
}

fn run(args: &[&str]) -> Result<String> {
    let output = Command::new("/usr/bin/systemctl")
        .args(args)
        .output()
        .context("Could not execute systemctl")?;

    if !output.status.success() {
        bail!(
            "systemctl failed: {}",
            String::from_utf8_lossy(&output.stderr).trim()
        );
    }

    Ok(String::from_utf8_lossy(&output.stdout)
        .trim()
        .to_string())
}

pub fn start(unit: &str) -> Result<()> {
    validate_unit(unit)?;
    run(&["start", unit])?;
    Ok(())
}

pub fn stop(unit: &str) -> Result<()> {
    validate_unit(unit)?;
    run(&["stop", unit])?;
    Ok(())
}

pub fn restart(unit: &str) -> Result<()> {
    validate_unit(unit)?;
    run(&["restart", unit])?;
    Ok(())
}

pub fn status(unit: &str) -> Result<ServiceStatus> {
    validate_unit(unit)?;

    let state = run(&[
        "show",
        unit,
        "--property=ActiveState",
        "--value",
    ])?;

    let description = run(&[
        "show",
        unit,
        "--property=Description",
        "--value",
    ])?;

    Ok(ServiceStatus {
        active: state == "active",
        state,
        description,
    })
}
