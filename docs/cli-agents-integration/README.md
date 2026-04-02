# CLI Agents Integration Guide

This directory contains comprehensive integration documentation for all CLI agents with HelixAgent.

## Overview

All 47 CLI agents can be configured to use HelixAgent as their AI provider, enabling:
- **AI Debate Ensemble**: Multi-LLM consensus via HelixAgent
- **MCP Integration**: 45+ MCP servers (filesystem, browser, git, etc.)
- **LSP Support**: Language Server Protocol integration
- **ACP Support**: Agent Communication Protocol
- **Code Formatters**: 32+ formatters
- **Vision**: Image analysis capabilities
- **Embeddings**: Semantic search and RAG

## Configuration Structure

Each agent has a JSON configuration file in `cli_agents_configs/` that defines:
1. Provider settings (HelixAgent endpoint)
2. MCP server configurations
3. Model capabilities
4. Agent profiles
5. Formatter preferences
6. Extension settings

## Quick Links

- [Architecture Overview](./ARCHITECTURE.md)
- [MCP Servers Reference](./MCP_SERVERS.md)
- [HTTP Endpoints Reference](./HTTP_ENDPOINTS.md)
- [Configuration Schema](./CONFIG_SCHEMA.md)

## Supported Agents

47 CLI agents supported.

