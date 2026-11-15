
use clap::{Parser, Subcommand};

#[derive(Parser)]
#[command(
    name = "valheimctl",
    version,
    about = "Manage a Valheim dedicated server"
)]
pub struct Cli {
    #[arg(
        short,
        long,
        default_value = "/etc/valheimctl.toml"
    )]
    pub config: String,

    #[command(subcommand)]
    pub command: Command,
}

#[derive(Subcommand)]
pub enum Command {
    Status,
    Start,
    Stop,
    Restart,
    Worlds,
    Backup,
    Web,
}
