# Snyk Security Scanner Container for HelixAgent
FROM snyk/snyk-cli:latest

RUN apt-get update && apt-get install -y curl jq git \
    && rm -rf /var/lib/apt/lists/*

RUN curl -L https://go.dev/dl/go1.24.0.linux-amd64.tar.gz | tar -C /usr/local -xzf -
ENV PATH=$PATH:/usr/local/go/bin
ENV GOPATH=/go
ENV GOBIN=$GOPATH/bin
ENV PATH=$PATH:$GOBIN

WORKDIR /app
COPY . .

RUN mkdir -p /scripts /reports

RUN cat > /scripts/scan-all.sh << 'SCRIPT_EOF'
#!/bin/bash
echo "🔒 HelixAgent Security Scan with Snyk"
mkdir -p /reports
echo "Scanning dependencies..."
snyk test --json > /reports/snyk-deps.json 2>&1 || true
echo "Scanning code..."
snyk code test --json > /reports/snyk-code.json 2>&1 || true
echo "✅ Scan complete"
SCRIPT_EOF

RUN chmod +x /scripts/*.sh
CMD ["/scripts/scan-all.sh"]
