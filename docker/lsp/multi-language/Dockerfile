# docker/lsp/multi-language/Dockerfile
# Multi-Language LSP Container: Go, Rust, Python, TypeScript, C++, Java
FROM ubuntu:22.04

ENV DEBIAN_FRONTEND=noninteractive

# Install base dependencies
RUN apt-get update && apt-get install -y \
    curl \
    wget \
    git \
    ca-certificates \
    gnupg \
    software-properties-common \
    supervisor \
    socat \
    netcat-openbsd \
    && rm -rf /var/lib/apt/lists/*

# ============= Go (gopls) =============
ENV GO_VERSION=1.24
RUN wget -q https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz && \
    rm go${GO_VERSION}.linux-amd64.tar.gz
ENV PATH="/usr/local/go/bin:/root/go/bin:${PATH}"
ENV GOPATH="/root/go"
RUN go install golang.org/x/tools/gopls@latest

# ============= Rust (rust-analyzer) =============
RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
ENV PATH="/root/.cargo/bin:${PATH}"
RUN rustup component add rust-analyzer

# ============= Python (pylsp + pyright) =============
RUN apt-get update && apt-get install -y python3 python3-pip python3-venv && \
    rm -rf /var/lib/apt/lists/*
RUN pip3 install --break-system-packages \
    python-lsp-server[all] \
    pyright \
    pylsp-mypy \
    python-lsp-black \
    pylsp-rope

# ============= Node.js + TypeScript LSP =============
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - && \
    apt-get install -y nodejs && \
    rm -rf /var/lib/apt/lists/*
RUN npm install -g typescript typescript-language-server

# ============= C/C++ (clangd) =============
RUN apt-get update && apt-get install -y clangd-15 && \
    ln -s /usr/bin/clangd-15 /usr/bin/clangd && \
    rm -rf /var/lib/apt/lists/*

# ============= Java (jdt.ls) =============
RUN apt-get update && apt-get install -y openjdk-17-jdk && \
    rm -rf /var/lib/apt/lists/*
ENV JAVA_HOME=/usr/lib/jvm/java-17-openjdk-amd64
RUN mkdir -p /opt/jdtls && \
    wget -q https://www.eclipse.org/downloads/download.php?file=/jdtls/snapshots/jdt-language-server-latest.tar.gz -O /tmp/jdtls.tar.gz && \
    tar -xzf /tmp/jdtls.tar.gz -C /opt/jdtls && \
    rm /tmp/jdtls.tar.gz

# Create supervisord config
RUN mkdir -p /etc/supervisor/conf.d
COPY supervisord.conf /etc/supervisor/supervisord.conf

# Create wrapper scripts for TCP access
RUN mkdir -p /scripts
COPY scripts/* /scripts/
RUN chmod +x /scripts/*.sh

EXPOSE 5001 5002 5003 5004 5005 5006

CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/supervisord.conf"]
