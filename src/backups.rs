
use crate::systemd;
use anyhow::{Context, Result};
use chrono::Utc;
use flate2::{
    read::GzDecoder,
    write::GzEncoder,
    Compression,
};
use std::{
    fs::{self, File},
    path::{Path, PathBuf},
};
use tar::{Archive, Builder};
use uuid::Uuid;

pub fn create_consistent_backup(
    unit: &str,
    worlds_dir: impl AsRef<Path>,
    backup_dir: impl AsRef<Path>,
) -> Result<PathBuf> {
    let worlds_dir = worlds_dir.as_ref();
    let backup_dir = backup_dir.as_ref();

    anyhow::ensure!(
        worlds_dir.is_dir(),
        "Save directory does not exist: {}",
        worlds_dir.display()
    );

    let original_status = systemd::status(unit)
        .context("Could not determine server state; refusing backup")?;

    fs::create_dir_all(backup_dir)
        .context("Could not create backup directory")?;

    if original_status.active {
        systemd::stop(unit)
            .context("Could not stop server for backup")?;
    }

    let result = create_archive(worlds_dir, backup_dir);

    let restart_result = if original_status.active {
        systemd::start(unit)
    } else {
        Ok(())
    };

    match (result, restart_result) {
        (Ok(path), Ok(())) => Ok(path),
        (Err(error), Ok(())) => Err(error),
        (Ok(_), Err(error)) => {
            anyhow::bail!(
                "Backup was created, but the server could not be restarted: {error:#}"
            )
        }
        (Err(backup_error), Err(restart_error)) => {
            anyhow::bail!(
                "Backup failed: {backup_error:#}; server restart also failed: {restart_error:#}"
            )
        }
    }
}

fn create_archive(
    worlds_dir: &Path,
    backup_dir: &Path,
) -> Result<PathBuf> {
    let filename = format!(
        "valheim-worlds-{}-{}.tar.gz",
        Utc::now().format("%Y%m%d-%H%M%S"),
        Uuid::new_v4().simple()
    );

    let destination = backup_dir.join(filename);
    let file = File::create(&destination)?;

    let encoder = GzEncoder::new(file, Compression::default());
    let mut archive = Builder::new(encoder);

    archive
        .append_dir_all("worlds_local", worlds_dir)
        .context("Could not add save directory to archive")?;

    let encoder = archive.into_inner()?;
    encoder.finish()?;

    verify_archive(&destination)?;

    Ok(destination)
}

pub fn verify_archive(path: impl AsRef<Path>) -> Result<()> {
    let file = File::open(path.as_ref())
        .context("Could not open backup archive")?;

    let decoder = GzDecoder::new(file);
    let mut archive = Archive::new(decoder);
    let mut found_entries = 0usize;

    for entry in archive.entries()? {
        let entry = entry?;
        let entry_path = entry.path()?;

        anyhow::ensure!(
            !entry_path.is_absolute()
                && !entry_path.components().any(|component| {
                    matches!(
                        component,
                        std::path::Component::ParentDir
                    )
                }),
            "Backup contains an unsafe path"
        );

        found_entries += 1;
    }

    anyhow::ensure!(
        found_entries > 0,
        "Backup archive is empty"
    );

    Ok(())
}
