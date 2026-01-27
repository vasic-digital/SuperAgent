#!/bin/bash
# TypeScript Language Server TCP wrapper
PORT=${TSSERVER_PORT:-5004}
socat TCP-LISTEN:${PORT},reuseaddr,fork EXEC:"/usr/bin/typescript-language-server --stdio"
