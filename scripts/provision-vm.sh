#!/usr/bin/env bash
# provision-vm.sh - One-shot provisioning for a Debian VM.
# Run this inside a Debian VM as a user with sudo.
set -euo pipefail

echo "==> Updating package index"
sudo apt-get update
DEBIAN_FRONTEND=noninteractive sudo apt-get install -y \
  curl git build-essential ca-certificates sudo \
  steamcmd libatomic1 libpulse-dev libpulse0

echo "==> Installing Go 1.25.7"
if ! command -v go >/dev/null 2>&1; then
  ARCH=$(uname -m)
  if [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
    GO_TARBALL="go1.25.7.linux-arm64.tar.gz"
  else
    GO_TARBALL="go1.25.7.linux-amd64.tar.gz"
  fi
  curl -fsSL "https://go.dev/dl/${GO_TARBALL}" -o /tmp/go.tar.gz
  sudo tar -C /usr/local -xzf /tmp/go.tar.gz
  rm -f /tmp/go.tar.gz
  echo 'export PATH=$PATH:/usr/local/go/bin' | sudo tee /etc/profile.d/go.sh >/dev/null
  sudo chmod +x /etc/profile.d/go.sh
fi
export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin
echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' >> ~/.bashrc

echo "==> Cloning valheimctl"
if [ ! -d ~/valheimctl ]; then
  git clone https://github.com/sumwatshade/valheimctl.git ~/valheimctl
else
  cd ~/valheimctl && git pull --ff-only
fi

echo "==> Building and installing valheimctl"
cd ~/valheimctl
go build -o /tmp/valheimctl .
sudo cp /tmp/valheimctl /usr/local/bin/valheimctl
sudo chmod +x /usr/local/bin/valheimctl

echo "==> valheimctl installed:"
valheimctl --help

echo "==> Running test suite"
go test ./...

echo "==> Done. Run 'valheimctl --help' to see available commands."
