# Submodule Fixes

## Issue Summary

The `cli_agents/bridle` submodule contains a nested submodule `plugins/skill-enhancers/axiom` that is missing its URL configuration in `.gitmodules`.

### Root Cause
The bridle repository (github.com:jeremylongshore/claude-code-plugins-plus-skills) has a git submodule reference for `axiom` but the URL is not configured in `.gitmodules`.

### Error Message
```
fatal: No url found for submodule path 'cli_agents/bridle/plugins/skill-enhancers/axiom' in .gitmodules
fatal: Failed to recurse into submodule path 'cli_agents/bridle'
```

### Impact
- Cannot run `git submodule update --init --recursive` without errors
- CI/CD pipelines may fail
- Fresh clones cannot fully initialize all submodules

### Workaround

Until the upstream bridle repository fixes this issue, use the following workaround:

```bash
# Update submodules but skip bridle's nested submodules
git submodule update --init --recursive -- cli_agents/bridle
git submodule update --init --recursive
# The error will appear but other submodules will update
```

Or use the provided script:
```bash
./scripts/update_submodules.sh
```

### Permanent Fix Options

1. **Fork bridle repository** (Recommended for production)
   - Fork `jeremylongshore/claude-code-plugins-plus-skills` to our organization
   - Fix the submodule URL in the fork
   - Update main repo to point to our fork

2. **Remove bridle submodule** (If not critical)
   - Remove from `.gitmodules`
   - Document as optional dependency

3. **Contact upstream** (Long-term)
   - Open issue on bridle repository
   - Wait for fix, then update submodule reference

### Status
- **Identified**: 2026-04-02
- **Severity**: Medium (workaround available)
- **Blocking**: No (CI can use workaround script)
