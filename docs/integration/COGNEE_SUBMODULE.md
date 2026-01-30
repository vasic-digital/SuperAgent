# Cognee Integration via Git Submodule

## Overview

Cognee is integrated into HelixAgent as a **Git submodule** located at `external/cognee`. This allows us to:
- Track specific Cognee versions
- Apply custom bugfixes
- Pull latest updates from upstream
- Maintain reproducible builds

## Submodule Details

- **Repository**: https://github.com/topoteretes/cognee
- **Local Path**: `external/cognee/`
- **Default Branch**: `helixagent-bugfix` (contains critical fixes)
- **Upstream Branch**: `main`

## Quick Reference

### Update to Latest Cognee

```bash
cd external/cognee
git fetch origin
git checkout main
git pull origin main
cd ../..
git add external/cognee
git commit -m "Update Cognee to latest version"
```

### Switch to Specific Version

```bash
cd external/cognee
git checkout v0.5.1  # or any tag
cd ../..
git add external/cognee
git commit -m "Pin Cognee to v0.5.1"
```

### Return to Bugfix Branch

```bash
cd external/cognee
git checkout helixagent-bugfix
cd ../..
```

### Rebuild Cognee Container

```bash
# Stop existing container
podman stop helixagent-cognee
podman rm helixagent-cognee

# Rebuild from submodule
podman build -t helixagent-cognee:latest -f external/cognee/Dockerfile external/cognee

# Or use docker-compose
podman-compose build cognee
podman-compose up -d cognee
```

## HelixAgent Bugfix Branch

The `helixagent-bugfix` branch contains critical fixes for HelixAgent integration:

### Fix 1: extract_subgraph_chunks Type Handling

**File**: `cognee/tasks/memify/extract_subgraph_chunks.py`

**Problem**: Original code assumed all inputs were `CogneeGraph` objects, but sometimes receives raw strings, causing:
```python
AttributeError: 'str' object has no attribute 'nodes'
```

**Solution**: Added type checking to handle both `CogneeGraph` objects AND raw strings:

```python
async def extract_subgraph_chunks(subgraphs):
    for subgraph in subgraphs:
        if isinstance(subgraph, str):
            yield subgraph
        elif isinstance(subgraph, CogneeGraph):
            for node in subgraph.nodes.values():
                if node.attributes["type"] == "DocumentChunk":
                    yield node.attributes["text"]
        else:
            logging.warning(f"Unexpected subgraph type: {type(subgraph)}")
            yield str(subgraph)
```

**Impact**: Prevents API timeouts and enables proper integration with HelixAgent.

## Testing After Updates

After updating Cognee or switching versions, run comprehensive tests:

```bash
# 1. Rebuild and start
podman-compose build cognee
podman-compose up -d cognee

# 2. Wait for container to be healthy
sleep 30

# 3. Run verification tests
./challenges/scripts/cognee_integration_challenge.sh

# 4. Test API response time
python3 tests/manual/test_cognee_response_time.py
```

## Version History

| Version | Date | Status | Notes |
|---------|------|--------|-------|
| `helixagent-bugfix` | 2026-01-30 | ✅ WORKING | Custom bugfix branch |
| `v0.5.1` | 2026-01-XX | ⚠️ BUGGY | Has extract_subgraph_chunks bug |
| `v0.5.0` | 2026-01-XX | ⚠️ BUGGY | Has extract_subgraph_chunks bug |
| `v0.4.1` | 2025-12-XX | ⚠️ BUGGY | Has extract_subgraph_chunks bug |
| `main` (latest) | 2026-01-30 | ⚠️ BUGGY | Upstream bug not yet fixed |

## Troubleshooting

### Container Fails to Start

```bash
# Check build logs
podman logs helixagent-cognee

# Rebuild from scratch
podman rmi helixagent-cognee:latest
podman build -t helixagent-cognee:latest -f external/cognee/Dockerfile external/cognee
```

### API Returns 401 Unauthorized

Cognee requires authentication. Ensure these environment variables are set:

```bash
COGNEE_AUTH_EMAIL=admin@helixagent.ai
COGNEE_AUTH_PASSWORD=HelixAgentPass123
```

### API Timeouts (30+ seconds)

This indicates the `extract_subgraph_chunks` bug. Ensure you're using `helixagent-bugfix` branch:

```bash
cd external/cognee
git branch  # Should show * helixagent-bugfix
git checkout helixagent-bugfix
cd ../..
podman-compose build cognee
podman-compose up -d cognee
```

### Submodule Not Initialized

If `external/cognee` is empty:

```bash
git submodule init
git submodule update
```

## Contributing Fixes Upstream

If you create additional fixes in `helixagent-bugfix` that should go upstream:

```bash
cd external/cognee

# Create feature branch from main
git checkout main
git pull origin main
git checkout -b fix/extract-subgraph-chunks

# Cherry-pick your commits
git cherry-pick <commit-hash>

# Push and create PR
git push origin fix/extract-subgraph-chunks
# Then create PR on GitHub
```

## References

- Cognee Repository: https://github.com/topoteretes/cognee
- Cognee Documentation: https://docs.cognee.ai
- Bug Report: `docs/COGNEE_BUG.md`
- Integration Tests: `challenges/scripts/cognee_integration_challenge.sh`
