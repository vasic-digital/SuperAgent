# docker/lsp/devops-lsp/Dockerfile
# DevOps LSP Container: Bash, YAML, Dockerfile, Terraform, XML
FROM node:20-slim

RUN apt-get update && apt-get install -y \
    socat \
    curl \
    supervisor \
    wget \
    unzip \
    openjdk-17-jre-headless \
    && rm -rf /var/lib/apt/lists/*

# Install bash-language-server
RUN npm install -g bash-language-server

# Install yaml-language-server
RUN npm install -g yaml-language-server

# Install dockerfile-language-server
RUN npm install -g dockerfile-language-server-nodejs

# Install terraform-ls
RUN wget -q https://releases.hashicorp.com/terraform-ls/0.32.5/terraform-ls_0.32.5_linux_amd64.zip -O /tmp/terraform-ls.zip && \
    unzip /tmp/terraform-ls.zip -d /usr/local/bin && \
    rm /tmp/terraform-ls.zip

# Install lemminx (XML Language Server)
RUN mkdir -p /opt/lemminx && \
    wget -q https://github.com/eclipse/lemminx/releases/download/0.27.0/lemminx-linux.zip -O /tmp/lemminx.zip && \
    unzip /tmp/lemminx.zip -d /opt/lemminx && \
    chmod +x /opt/lemminx/lemminx-linux && \
    rm /tmp/lemminx.zip

# Create workspace directory
RUN mkdir -p /workspace

# Supervisord config
COPY supervisord.conf /etc/supervisor/supervisord.conf

EXPOSE 5020 5021 5022 5023 5024

CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/supervisord.conf"]
