#!/bin/bash
# clangd TCP wrapper
PORT=${CLANGD_PORT:-5005}
socat TCP-LISTEN:${PORT},reuseaddr,fork EXEC:"/usr/bin/clangd"
