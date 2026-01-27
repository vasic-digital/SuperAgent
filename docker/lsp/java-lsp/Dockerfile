# docker/lsp/java-lsp/Dockerfile
# Java Language Server (Eclipse JDT.LS) - Standalone
FROM eclipse-temurin:17-jdk-jammy

RUN apt-get update && apt-get install -y \
    socat \
    curl \
    wget \
    && rm -rf /var/lib/apt/lists/*

# Download and install Eclipse JDT.LS
RUN mkdir -p /opt/jdtls && \
    wget -q "https://www.eclipse.org/downloads/download.php?file=/jdtls/snapshots/jdt-language-server-latest.tar.gz" -O /tmp/jdtls.tar.gz && \
    tar -xzf /tmp/jdtls.tar.gz -C /opt/jdtls && \
    rm /tmp/jdtls.tar.gz

# Create workspace directory
RUN mkdir -p /workspace

ENV JAVA_HOME=/opt/java/openjdk

EXPOSE 5006

# Startup script
RUN echo '#!/bin/bash\n\
socat TCP-LISTEN:5006,reuseaddr,fork EXEC:"java \
    -Declipse.application=org.eclipse.jdt.ls.core.id1 \
    -Dosgi.bundles.defaultStartLevel=4 \
    -Declipse.product=org.eclipse.jdt.ls.core.product \
    -noverify \
    -Xmx1G \
    -jar /opt/jdtls/plugins/org.eclipse.equinox.launcher_*.jar \
    -configuration /opt/jdtls/config_linux \
    -data /workspace/.jdtls"' > /start.sh && chmod +x /start.sh

CMD ["/start.sh"]
