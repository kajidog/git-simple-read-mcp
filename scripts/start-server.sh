#!/bin/bash

# Git Remote MCP Server の起動スクリプト

WORKSPACE_DIR="${1:-./workspace}"
PORT="${2:-8080}"
HOST="${3:-localhost}"

echo "Git Remote MCP Server を起動中..."
echo "ワークスペース: $WORKSPACE_DIR"
echo "ポート: $PORT"
echo "ホスト: $HOST"

# ワークスペースディレクトリを作成
mkdir -p "$WORKSPACE_DIR"

# サーバー起動
./git-remote-mcp mcp --transport http --port "$PORT" --host "$HOST" --workspace "$WORKSPACE_DIR"