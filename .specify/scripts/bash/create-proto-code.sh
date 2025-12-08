#!/usr/bin/env bash
set -euo pipefail

PROTOC=${PROTOC:-protoc}
PROTO_GEN_GO=${PROTO_GEN_GO:-protoc-gen-go}

if ! command -v "$PROTOC" >/dev/null 2>&1; then
  echo "protoc (Protocol Buffers compiler) not found. Please install to generate code." >&2
  exit 0
fi
if ! command -v "$PROTO_GEN_GO" >/dev/null 2>&1; then
  echo "protoc-gen-go not found. Please install to generate Go bindings." >&2
  exit 0
fi
PROJ_ROOT=$(pwd)
PROTO_PATH="$PROJ_ROOT/specs/001-super-agent/contracts/llm-facade.proto"
OUT_DIR="$PROJ_ROOT/specs/001-super-agent/contracts/"

echo "Generating Go bindings from $PROTO_PATH to $OUT_DIR";
protoc --go_out="$OUT_DIR" "$PROTO_PATH" || true

echo "Proto code generation attempted. If protoc plugins are installed, generated *_pb.go files will appear alongside the proto." 
