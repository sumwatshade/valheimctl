
mod backups;
mod cli;
mod config;
mod systemd;
mod web;
mod worlds;

use anyhow::Result;
use clap::Parser;
use cli::{Cli, Command};
use config::Config;

#[tokio::main]
async fn main() -> Result<()> {
    tracing_subscriber::fmt()
        .with_env_filter(
            tracing_subscriber::EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| "info".into())
        )
        .init();

    let cli = Cli::parse();
    let config = Config::load(&cli.config)?;

    match cli.command {
        Command::Status => {
            let status = systemd::status(
                &config.server.systemd_unit
            )?;

            println!("Server: {}", config.server.name);
            println!("State: {}", status.state);
            println!("Description: {}", status.description);
        }

        Command::Start => {
            systemd::start(&config.server.systemd_unit)?;
            println!("Start requested.");
        }

        Command::Stop => {
            systemd::stop(&config.server.systemd_unit)?;
            println!("Stop requested.");
        }

        Command::Restart => {
            systemd::restart(&config.server.systemd_unit)?;
            println!("Restart requested.");
        }

        Command::Worlds => {
            for entry in worlds::inspect(
                &config.worlds.directory
            )? {
                println!(
                    "{} ({}) — {} files — {} bytes{}",
                    entry.name,
                    entry.path.display(),
                    entry.files,
                    entry.bytes,
                    if entry.is_directory {
                        " (directory)"
                    } else {
                        ""
                    }
                );
            }
        }

        Command::Backup => {
            let path = backups::create_consistent_backup(
                &config.server.systemd_unit,
                &config.worlds.directory,
                &config.backups.directory,
            )?;

            println!("Backup created: {}", path.display());
        }

        Command::Web => {
            web::serve(config).await?;
        }
    }

    Ok(())
}
