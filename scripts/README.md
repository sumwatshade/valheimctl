# valheimctl VM setup & test scripts

This directory contains everything you need to spin up a Debian environment,
install `valheimctl`, and run repeatable end-to-end tests against it.

> **Note for Apple Silicon (ARM) Macs:** VirtualBox cannot run x86_64 guests on
> ARM. Use the Docker-based setup below instead, which runs natively on
> `linux/arm64`.

## Files

| File | Purpose |
|------|---------|
| `../Dockerfile` | Debian 12 container with Go, steamcmd, and valheimctl pre-installed. |
| `../docker-compose.yml` | Compose config for interactive container use. |
| `provision-vm.sh` | One-shot provisioning script for an existing Debian VM (kept for reference). |
| `test-valheimctl.sh` | End-to-end smoke test: init, start, stop, status, register, backup create/list/apply. |
| `test-valheimctl-docker.sh` | Builds the Docker image and runs the smoke test inside a container. |
| `rebuild-and-test-docker.sh` | Rebuilds the image, runs unit tests, then runs the smoke test. |

## Quick start (Docker — recommended for Apple Silicon)

From the repo root on your host:

```bash
# Build the image and run the smoke test
./scripts/test-valheimctl-docker.sh
```

This builds a Debian 12 container, installs Go + steamcmd + valheimctl, runs
`go test ./...`, then executes the smoke test inside the container.

## Iterating on the CLI

1. Edit source on your host.
2. Rebuild and re-test:

   ```bash
   ./scripts/rebuild-and-test-docker.sh
   ```

   This rebuilds the image, runs `go test ./...`, and runs the smoke test.

3. Repeat until the CLI works for your use case.

## Interactive container

```bash
docker-compose up
```

Then inside another terminal:

```bash
docker exec -it valheimctl-test bash
```

Inside the container, `valheimctl` is at `/usr/local/bin/valheimctl` and the
source is at `/home/valheim/valheimctl`.

## Notes

- `valheimctl` is Linux-only and targets Debian/Ubuntu with systemd. The Docker
  image uses `debian:bookworm-slim`.
- `start`/`stop`/`status` currently manipulate JSON state files under
  `.valheimctl/` — they do not launch or kill a real Valheim process.
- `register` writes a placeholder systemd unit and calls `systemctl enable`.
  Inside Docker, systemd is not running by default, so `register` will fail on
  the `systemctl` calls. To test `register`, you'd need a container with
  systemd (e.g. `jrei/systemd-ubuntu`) or a real VM.
- The smoke test wipes and recreates its root directory each run, so it is safe
  to run repeatedly.
