#!/usr/bin/env bash
# test-valheimctl-docker.sh - Build the Docker image and run the valheimctl
# smoke test inside a container.
#
# Usage: ./scripts/test-valheimctl-docker.sh [ROOT_DIR]
#
# ROOT_DIR defaults to /home/valheim/valheim-test-root inside the container.
# It is wiped and recreated each run so you can iterate repeatedly.
set -euo pipefail

ROOT_DIR="${1:-/home/valheim/valheim-test-root}"
CONTAINER_NAME="valheimctl-test"
IMAGE_NAME="valheimctl:latest"

cleanup() {
  docker rm -f "$CONTAINER_NAME" 2>/dev/null || true
}
trap cleanup EXIT

echo "==> Building Docker image: $IMAGE_NAME"
docker build -t "$IMAGE_NAME" .

echo "==> Removing any existing container"
docker rm -f "$CONTAINER_NAME" 2>/dev/null || true

echo "==> Starting container"
docker run -d \
  --name "$CONTAINER_NAME" \
  --hostname valheimctl-test \
  "$IMAGE_NAME" \
  bash -c "while true; do sleep 1; done"

echo "==> Copying smoke test into container"
docker cp scripts/test-valheimctl.sh "$CONTAINER_NAME:/home/valheim/test-valheimctl.sh"
docker exec --user root "$CONTAINER_NAME" chmod +x /home/valheim/test-valheimctl.sh

echo "==> Running smoke test inside container (ROOT_DIR=$ROOT_DIR)"
# register is expected to fail in Docker (no systemd); the smoke test handles
# that gracefully with || true and reports it as a FAIL.
# Run as root so init's apt-get prerequisite check can succeed.
docker exec --user root "$CONTAINER_NAME" bash /home/valheim/test-valheimctl.sh "$ROOT_DIR"

echo ""
echo "==> Done. Container '$CONTAINER_NAME' has been removed."
