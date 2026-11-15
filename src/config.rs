
use anyhow::{Context, Result};
use serde::Deserialize;
use std::{fs, path::Path};

#[derive(Debug, Clone, Deserialize)]
pub struct Config {
    pub server: ServerConfig,
    pub worlds: WorldsConfig,
    pub backups: BackupConfig,
    pub web: WebConfig,
}

#[derive(Debug, Clone, Deserialize)]
pub struct ServerConfig {
    pub name: String,
    pub systemd_unit: String,
}

#[derive(Debug, Clone, Deserialize)]
pub struct WorldsConfig {
    pub directory: String,
}

#[derive(Debug, Clone, Deserialize)]
pub struct BackupConfig {
    pub directory: String,
}

#[derive(Debug, Clone, Deserialize)]
pub struct WebConfig {
    pub bind: String,
    pub username: String,
    pub password: String,
    pub csrf_token: String,
}

impl Config {
    pub fn load(path: impl AsRef<Path>) -> Result<Self> {
        let contents = fs::read_to_string(path.as_ref())
            .context("Could not read configuration file")?;

        let config: Self = toml::from_str(&contents)
            .context("Invalid TOML configuration")?;

        anyhow::ensure!(
            !config.web.username.is_empty()
                && !config.web.password.is_empty()
                && !config.web.csrf_token.is_empty(),
            "Web username, password, and CSRF token must be configured"
        );

        anyhow::ensure!(
            !config.server.systemd_unit.starts_with('-'),
            "Invalid systemd unit name"
        );

        Ok(config)
    }
}
