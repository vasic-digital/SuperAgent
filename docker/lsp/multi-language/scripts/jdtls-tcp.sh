#!/bin/bash
# Eclipse JDT.LS TCP wrapper
PORT=${JDTLS_PORT:-5006}
WORKSPACE=${LSP_WORKSPACE:-/workspace}

socat TCP-LISTEN:${PORT},reuseaddr,fork EXEC:"java \
    -Declipse.application=org.eclipse.jdt.ls.core.id1 \
    -Dosgi.bundles.defaultStartLevel=4 \
    -Declipse.product=org.eclipse.jdt.ls.core.product \
    -Dlog.level=ALL \
    -noverify \
    -Xmx1G \
    -jar /opt/jdtls/plugins/org.eclipse.equinox.launcher_*.jar \
    -configuration /opt/jdtls/config_linux \
    -data ${WORKSPACE}/.jdtls"
