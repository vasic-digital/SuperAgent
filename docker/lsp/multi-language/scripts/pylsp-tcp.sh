#!/bin/bash
# Python LSP TCP wrapper
PORT=${PYLSP_PORT:-5003}
socat TCP-LISTEN:${PORT},reuseaddr,fork EXEC:"/usr/bin/pylsp"
