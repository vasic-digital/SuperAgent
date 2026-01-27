#!/bin/bash
# rust-analyzer TCP wrapper
PORT=${RUST_ANALYZER_PORT:-5002}
socat TCP-LISTEN:${PORT},reuseaddr,fork EXEC:"/root/.cargo/bin/rust-analyzer"
