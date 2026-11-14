# Dockerfile for valheimctl development and testing
# Runs natively on linux/arm64 (Apple Silicon) and linux/amd64
FROM debian:bookworm-slim

ENV DEBIAN_FRONTEND=noninteractive
ENV PATH=/usr/local/go/bin:$PATH

# Install base dependencies + Valheim server prerequisites
RUN apt-get update && apt-get install -y \
    curl \
    git \
    build-essential \
    ca-certificates \
    sudo \
    libatomic1 \
    libpulse-dev \
    libpulse0 \
    && rm -rf /var/lib/apt/lists/*

# Install steamcmd from Valve's CDN (not in Debian arm64 repos)
RUN mkdir -p /usr/local/steamcmd && \
    curl -fsSL https://steamcdn-a.akamaihd.net/client/installer/steamcmd_linux.tar.gz -o /tmp/steamcmd.tar.gz && \
    tar -C /usr/local/steamcmd -xzf /tmp/steamcmd.tar.gz && \
    rm -f /tmp/steamcmd.tar.gz && \
    ln -sf /usr/local/steamcmd/steamcmd.sh /usr/local/bin/steamcmd

# Install Go 1.25.7 (architecture-aware)
ARG TARGETARCH
RUN if [ "$TARGETARCH" = "arm64" ]; then \
        GO_TARBALL="go1.25.7.linux-arm64.tar.gz"; \
    else \
        GO_TARBALL="go1.25.7.linux-amd64.tar.gz"; \
    fi && \
    curl -fsSL "https://go.dev/dl/${GO_TARBALL}" -o /tmp/go.tar.gz && \
    tar -C /usr/local -xzf /tmp/go.tar.gz && \
    rm -f /tmp/go.tar.gz

# Create a non-root user with sudo (valheimctl needs sudo for systemctl)
RUN useradd -m -s /bin/bash valheim && \
    echo "valheim ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/valheim && \
    chmod 0440 /etc/sudoers.d/valheim

USER valheim
WORKDIR /home/valheim

# Copy the source from the build context (local edits are picked up on rebuild)
COPY . /home/valheim/valheimctl/

# Build and install valheimctl
RUN cd /home/valheim/valheimctl && \
    go build -buildvcs=false -o /tmp/valheimctl . && \
    sudo cp /tmp/valheimctl /usr/local/bin/valheimctl && \
    sudo chmod +x /usr/local/bin/valheimctl

# Run unit tests
RUN cd /home/valheim/valheimctl && go test -buildvcs=false ./...

# Default command
CMD ["valheimctl", "--help"]
