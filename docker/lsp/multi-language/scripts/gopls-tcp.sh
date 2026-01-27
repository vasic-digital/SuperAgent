#!/bin/bash
# gopls TCP wrapper - exposes gopls over TCP
PORT=${GOPLS_PORT:-5001}
socat TCP-LISTEN:${PORT},reuseaddr,fork EXEC:"/root/go/bin/gopls serve"
