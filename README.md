
# valheimctl

A Rust CLI and LAN web interface for managing an existing Valheim dedicated server.

## Current features

- CLI status, start, stop, and restart
- Read-only inspection of the configured save directory
- Full save-directory archive backups
- Backup verification by reading the gzip and tar archive
- Backup operation stops the server temporarily if it was running
- LAN web dashboard
- HTTP Basic Authentication
- CSRF token validation on web actions
- Web start, stop, restart, and backup actions
- Fixed systemd unit operations; no arbitrary shell command endpoint

## Not implemented yet

- Restore and rollback
- SteamCMD updates
- Live log streaming
- Backup retention and scheduled backups
- Persistent audit history
- HTTPS termination

## Build

Install Rust and Cargo, then:

    cargo check
    cargo test
    cargo build --release

## Configuration

Copy the example configuration:

    sudo install -o root -g valheimctl -m 0640 \
      config/valheimctl.example.toml \
      /etc/valheimctl.toml

Edit the configuration and set the actual worlds directory, backup directory,
systemd unit, web bind address, username, password, and CSRF token.

Use long, unique random values for both web secrets.

## CLI

    valheimctl --config /etc/valheimctl.toml status
    valheimctl --config /etc/valheimctl.toml start
    valheimctl --config /etc/valheimctl.toml stop
    valheimctl --config /etc/valheimctl.toml restart
    valheimctl --config /etc/valheimctl.toml worlds
    valheimctl --config /etc/valheimctl.toml backup

## Web

For initial testing, bind to localhost:

    valheimctl --config /etc/valheimctl.toml web

Use an SSH tunnel to access it remotely:

    ssh -L 8080:127.0.0.1:8080 your-user@your-valheim-vm

Then browse to:

    http://127.0.0.1:8080

For LAN access, set `web.bind` to the VM's LAN address and port.
Restrict that port with the VM firewall to your trusted LAN/VLAN.

HTTP Basic Authentication over plain HTTP does not encrypt credentials.
For routine LAN use, HTTPS through a reverse proxy is recommended.
Do not expose this service directly to the internet.

## Save safety

The application treats the configured save directory as opaque data.
It does not rename, convert, repair, or delete save files.

Backups archive the entire directory tree. A backup temporarily stops the
Valheim service if it was running and attempts to start it again afterward.

The current backup implementation does not yet provide restore, rollback,
retention, or a persistent operation log. Keep independent backups until
those features have been implemented and tested.
