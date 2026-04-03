#!/bin/bash
# Test Data Seeder for CLI Agents Validation
# Populates database with test instances and configurations

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[PASS]${NC} $1"; }
log_error() { echo -e "${RED}[FAIL]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }

# Database configuration
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-helixagent}"
DB_PASS="${DB_PASS:-helixagent123}"
DB_NAME="${DB_NAME:-helixagent_db}"

DATABASE_URL="postgres://${DB_USER}:${DB_PASS}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable"

# Check PostgreSQL
log_info "Checking PostgreSQL connection..."
if ! pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" > /dev/null 2>&1; then
    log_error "PostgreSQL is not running or not accessible"
    exit 1
fi
log_success "PostgreSQL is ready"

# Seed agent instances
log_info "Seeding agent instances..."

psql "$DATABASE_URL" << 'EOF'
-- Clear existing test data (be careful in production!)
TRUNCATE TABLE agent_instances CASCADE;
TRUNCATE TABLE ensemble_sessions CASCADE;
TRUNCATE TABLE feature_registry CASCADE;
TRUNCATE TABLE agent_communication_log CASCADE;
TRUNCATE TABLE distributed_locks CASCADE;
TRUNCATE TABLE crdt_state CASCADE;
TRUNCATE TABLE background_tasks CASCADE;

-- Insert test agent instances for different providers
INSERT INTO agent_instances (id, agent_type, instance_id, status, capabilities, last_heartbeat, created_at, updated_at) VALUES
-- Primary Tier
(gen_random_uuid(), 'claude', 'claude-primary-01', 'active', '{"vision": true, "tools": true, "streaming": true, "context_window": 200000}'::jsonb, NOW(), NOW(), NOW()),
(gen_random_uuid(), 'claude', 'claude-primary-02', 'idle', '{"vision": true, "tools": true, "streaming": true, "context_window": 200000}'::jsonb, NOW() - INTERVAL '5 minutes', NOW(), NOW()),
(gen_random_uuid(), 'openai-gpt4', 'gpt4-primary-01', 'active', '{"vision": true, "tools": true, "streaming": true, "json_mode": true}'::jsonb, NOW(), NOW(), NOW()),
(gen_random_uuid(), 'codex', 'codex-primary-01', 'active', '{"code": true, "streaming": true, "editor": true}'::jsonb, NOW(), NOW(), NOW()),

-- Secondary Tier
(gen_random_uuid(), 'gemini', 'gemini-secondary-01', 'active', '{"vision": true, "tools": true, "streaming": true, "context_window": 1000000}'::jsonb, NOW(), NOW(), NOW()),
(gen_random_uuid(), 'deepseek', 'deepseek-secondary-01', 'active', '{"code": true, "reasoning": true, "streaming": true}'::jsonb, NOW(), NOW(), NOW()),
(gen_random_uuid(), 'mistral', 'mistral-secondary-01', 'idle', '{"tools": true, "streaming": true, "json_mode": true}'::jsonb, NOW() - INTERVAL '10 minutes', NOW(), NOW()),
(gen_random_uuid(), 'groq', 'groq-secondary-01', 'active', '{"streaming": true, "fast": true, "low_latency": true}'::jsonb, NOW(), NOW(), NOW()),
(gen_random_uuid(), 'qwen', 'qwen-secondary-01', 'active', '{"vision": true, "tools": true, "streaming": true, "multilingual": true}'::jsonb, NOW(), NOW(), NOW()),
(gen_random_uuid(), 'xai', 'xai-secondary-01', 'background', '{"vision": true, "streaming": true, "real_time": true}'::jsonb, NOW() - INTERVAL '2 minutes', NOW(), NOW()),
(gen_random_uuid(), 'cohere', 'cohere-secondary-01', 'idle', '{"tools": true, "streaming": true, "rerank": true}'::jsonb, NOW() - INTERVAL '15 minutes', NOW(), NOW()),
(gen_random_uuid(), 'perplexity', 'perplexity-secondary-01', 'active', '{"search": true, "citations": true, "streaming": true}'::jsonb, NOW(), NOW(), NOW()),

-- Tertiary Tier (for testing various configurations)
(gen_random_uuid(), 'together', 'together-tertiary-01', 'idle', '{"streaming": true}'::jsonb, NOW() - INTERVAL '30 minutes', NOW(), NOW()),
(gen_random_uuid(), 'fireworks', 'fireworks-tertiary-01', 'active', '{"streaming": true, "fast": true}'::jsonb, NOW(), NOW(), NOW()),
(gen_random_uuid(), 'openrouter', 'openrouter-tertiary-01', 'active', '{"aggregated": true, "streaming": true}'::jsonb, NOW(), NOW(), NOW()),
(gen_random_uuid(), 'ai21', 'ai21-tertiary-01', 'idle', '{"streaming": true}'::jsonb, NOW() - INTERVAL '1 hour', NOW(), NOW()),
(gen_random_uuid(), 'azure', 'azure-tertiary-01', 'active', '{"vision": true, "tools": true, "enterprise": true}'::jsonb, NOW(), NOW(), NOW()),
(gen_random_uuid(), 'bedrock', 'bedrock-tertiary-01', 'active', '{"vision": true, "enterprise": true}'::jsonb, NOW(), NOW(), NOW()),
(gen_random_uuid(), 'ollama', 'ollama-local-01', 'active', '{"local": true, "privacy": true}'::jsonb, NOW(), NOW(), NOW());

-- Insert ensemble sessions for testing
INSERT INTO ensemble_sessions (id, strategy, status, agent_count, coordination_config, created_at, updated_at) VALUES
(gen_random_uuid(), 'voting', 'active', 3, '{"threshold": 0.66, "timeout": 30}'::jsonb, NOW(), NOW()),
(gen_random_uuid(), 'consensus', 'active', 5, '{"min_agreement": 0.8, "max_rounds": 3}'::jsonb, NOW(), NOW()),
(gen_random_uuid(), 'debate', 'paused', 4, '{"rounds": 2, "critique_depth": "deep"}'::jsonb, NOW() - INTERVAL '1 hour', NOW() - INTERVAL '30 minutes'),
(gen_random_uuid(), 'pipeline', 'completed', 3, '{"stages": ["analysis", "synthesis", "review"]}'::jsonb, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '1 hour'),
(gen_random_uuid(), 'parallel', 'active', 6, '{"max_concurrent": 4, "aggregation": "best_of_n"}'::jsonb, NOW(), NOW()),
(gen_random_uuid(), 'expert_panel', 'active', 4, '{"domain_experts": ["code", "math", "writing", "analysis"]}'::jsonb, NOW(), NOW());

-- Insert feature registry
INSERT INTO feature_registry (id, feature_name, porting_status, implementation_version, configuration_schema, created_at, updated_at) VALUES
(gen_random_uuid(), 'instance_management', 'completed', '1.0.0', '{"max_instances": 100, "pool_size": 10}'::jsonb, NOW(), NOW()),
(gen_random_uuid(), 'event_bus', 'completed', '1.0.0', '{"buffer_size": 1000, "workers": 4}'::jsonb, NOW(), NOW()),
(gen_random_uuid(), 'distributed_sync', 'completed', '1.0.0', '{"lock_timeout": 30, "retry_attempts": 3}'::jsonb, NOW(), NOW()),
(gen_random_uuid(), 'multi_instance_coordination', 'completed', '1.0.0', '{"strategies": ["voting", "consensus", "debate", "pipeline", "parallel", "sequential", "expert_panel"]}'::jsonb, NOW(), NOW()),
(gen_random_uuid(), 'load_balancer', 'completed', '1.0.0', '{"algorithms": ["round_robin", "least_connections", "weighted_response_time", "priority"]}'::jsonb, NOW(), NOW()),
(gen_random_uuid(), 'health_monitor', 'completed', '1.0.0', '{"check_interval": 30, "failure_threshold": 3}'::jsonb, NOW(), NOW()),
(gen_random_uuid(), 'background_workers', 'completed', '1.0.0', '{"task_types": ["health_check", "sync", "cleanup", "metrics", "report", "optimize", "backup"]}'::jsonb, NOW(), NOW()),
(gen_random_uuid(), 'aider_integration', 'completed', '1.0.0', '{"repo_map": true, "diff_format": true, "git_ops": true}'::jsonb, NOW(), NOW()),
(gen_random_uuid(), 'claude_code_integration', 'completed', '1.0.0', '{"terminal_ui": true, "interactive": true}'::jsonb, NOW(), NOW()),
(gen_random_uuid(), 'openhands_integration', 'completed', '1.0.0', '{"sandbox": true, "execution": true}'::jsonb, NOW(), NOW()),
(gen_random_uuid(), 'kiro_integration', 'completed', '1.0.0', '{"memory": true, "persistence": true}'::jsonb, NOW(), NOW()),
(gen_random_uuid(), 'continue_integration', 'completed', '1.0.0', '{"lsp_client": true, "inline": true}'::jsonb, NOW(), NOW()),
(gen_random_uuid(), 'streaming_pipeline', 'completed', '1.0.0', '{"buffer_size": 4096, "flush_interval": 100}'::jsonb, NOW(), NOW()),
(gen_random_uuid(), 'formatters', 'completed', '1.0.0', '{"formats": ["json", "xml", "yaml", "markdown", "code"]}'::jsonb, NOW(), NOW()),
(gen_random_uuid(), 'semantic_caching', 'completed', '1.0.0', '{"similarity_threshold": 0.95, "ttl": 3600}'::jsonb, NOW(), NOW()),
(gen_random_uuid(), 'http_handlers', 'completed', '1.0.0', '{"rate_limit": 1000, "timeout": 60}'::jsonb, NOW(), NOW()),
(gen_random_uuid(), 'rest_endpoints', 'completed', '1.0.0', '{"version": "v1", "cors": true}'::jsonb, NOW(), NOW()),
(gen_random_uuid(), 'openai_compatibility', 'completed', '1.0.0', '{"chat_completions": true, "embeddings": true}'::jsonb, NOW(), NOW()),
(gen_random_uuid(), 'ensemble_endpoints', 'completed', '1.0.0', '{"session_management": true, "real_time": true}'::jsonb, NOW(), NOW()),
(gen_random_uuid(), 'provider_registry', 'completed', '1.0.0', '{"providers": 47, "health_checks": true}'::jsonb, NOW(), NOW());

-- Insert sample communication logs
INSERT INTO agent_communication_log (id, sender_id, receiver_id, message_type, payload, timestamp, session_id) VALUES
(gen_random_uuid(), 'claude-primary-01', 'ensemble-session-1', 'request', '{"task": "analyze_code", "content": "func main() {}"}'::jsonb, NOW(), gen_random_uuid()),
(gen_random_uuid(), 'gpt4-primary-01', 'ensemble-session-1', 'response', '{"result": "analysis_complete", "findings": ["empty_function"]}'::jsonb, NOW(), gen_random_uuid()),
(gen_random_uuid(), 'deepseek-secondary-01', 'ensemble-session-2', 'request', '{"task": "optimize", "content": "for i := 0; i < len(arr); i++"}'::jsonb, NOW(), gen_random_uuid()),
(gen_random_uuid(), 'gemini-secondary-01', 'ensemble-session-2', 'response', '{"result": "optimization", "suggestion": "range loop"}'::jsonb, NOW(), gen_random_uuid());

-- Insert background tasks
INSERT INTO background_tasks (id, task_type, status, payload, priority, scheduled_at, started_at, completed_at, result, error_message, retry_count, max_retries, created_at, updated_at) VALUES
(gen_random_uuid(), 'health_check', 'completed', '{"target": 'claude-primary-01'}'::jsonb, 1, NOW() - INTERVAL '5 minutes', NOW() - INTERVAL '5 minutes', NOW() - INTERVAL '4 minutes', '{"status": "healthy"}'::jsonb, NULL, 0, 3, NOW(), NOW()),
(gen_random_uuid(), 'sync', 'running', '{"type": "crdt_merge"}'::jsonb, 2, NOW() - INTERVAL '2 minutes', NOW() - INTERVAL '2 minutes', NULL, NULL, NULL, 0, 3, NOW(), NOW()),
(gen_random_uuid(), 'cleanup', 'pending', '{"target": "old_logs"}'::jsonb, 3, NOW() + INTERVAL '1 hour', NULL, NULL, NULL, NULL, 0, 3, NOW(), NOW()),
(gen_random_uuid(), 'metrics', 'completed', '{"aggregation": "hourly"}'::jsonb, 1, NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour', NOW() - INTERVAL '55 minutes', '{"records_processed": 15000}'::jsonb, NULL, 0, 3, NOW(), NOW()),
(gen_random_uuid(), 'report', 'failed', '{"type": "daily_summary"}'::jsonb, 2, NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day', NOW() - INTERVAL '23 hours', NULL, 'Database connection timeout', 3, 3, NOW(), NOW()),
(gen_random_uuid(), 'optimize', 'pending', '{"target": "query_performance"}'::jsonb, 2, NOW() + INTERVAL '30 minutes', NULL, NULL, NULL, NULL, 0, 3, NOW(), NOW()),
(gen_random_uuid(), 'backup', 'completed', '{"destination": "s3"}'::jsonb, 3, NOW() - INTERVAL '6 hours', NOW() - INTERVAL '6 hours', NOW() - INTERVAL '5 hours', '{"size": "2.5GB", "files": 150}'::jsonb, NULL, 0, 3, NOW(), NOW());

EOF

log_success "Test data seeded successfully"

# Verify counts
log_info "Verifying seeded data..."

INSTANCE_COUNT=$(psql "$DATABASE_URL" -t -c "SELECT COUNT(*) FROM agent_instances;" | xargs)
SESSION_COUNT=$(psql "$DATABASE_URL" -t -c "SELECT COUNT(*) FROM ensemble_sessions;" | xargs)
FEATURE_COUNT=$(psql "$DATABASE_URL" -t -c "SELECT COUNT(*) FROM feature_registry;" | xargs)
TASK_COUNT=$(psql "$DATABASE_URL" -t -c "SELECT COUNT(*) FROM background_tasks;" | xargs)

echo ""
echo "Seeded Data Summary:"
echo "  📊 Agent Instances: $INSTANCE_COUNT"
echo "  📊 Ensemble Sessions: $SESSION_COUNT"
echo "  📊 Feature Registry: $FEATURE_COUNT"
echo "  📊 Background Tasks: $TASK_COUNT"
echo ""

log_success "Database is ready for testing!"
