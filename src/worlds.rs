
use anyhow::{Context, Result};
use std::{
    fs,
    path::{Path, PathBuf},
};
use walkdir::WalkDir;

#[derive(Debug, Clone)]
pub struct SaveEntry {
    pub name: String,
    pub path: PathBuf,
    pub is_directory: bool,
    pub files: usize,
    pub bytes: u64,
}

pub fn inspect(root: impl AsRef<Path>) -> Result<Vec<SaveEntry>> {
    let root = root.as_ref();

    if !root.exists() {
        anyhow::bail!(
            "Save directory does not exist: {}",
            root.display()
        );
    }

    anyhow::ensure!(
        root.is_dir(),
        "Save path is not a directory: {}",
        root.display()
    );

    let mut entries = Vec::new();

    for item in fs::read_dir(root)
        .context("Could not read save directory")?
    {
        let item = item?;
        let path = item.path();
        let metadata = fs::symlink_metadata(&path)?;

        let is_directory = metadata.is_dir();
        let mut files = 0usize;
        let mut bytes = 0u64;

        if is_directory {
            for child in WalkDir::new(&path)
                .follow_links(false)
                .into_iter()
                .filter_map(Result::ok)
            {
                if child.file_type().is_file() {
                    files += 1;
                    bytes += child.metadata()?.len();
                }
            }
        } else if metadata.is_file() {
            files = 1;
            bytes = metadata.len();
        }

        entries.push(SaveEntry {
            name: item.file_name().to_string_lossy().to_string(),
            path,
            is_directory,
            files,
            bytes,
        });
    }

    entries.sort_by(|a, b| a.name.cmp(&b.name));
    Ok(entries)
}
