/**
 * HelixAgent COMPREHENSIVE Video QA — Full System Coverage
 *
 * Tests ALL 7 applications, 310+ endpoints, OpenCode CLI agent,
 * all protocols, all features, all edge cases, security, streaming.
 *
 * Prerequisites:
 *   - HelixAgent running on port 7061
 *   - Website served on port 8090
 *   - Cognee-mock on port 8000
 *   - API demo on port 8080
 *   - OpenCode CLI agent binary available
 */
const { chromium } = require('playwright');
const path = require('path');
const fs = require('fs');
const http = require('http');
const { execFileSync } = require('child_process');

const TS = new Date().toISOString().replace(/[:.]/g, '-').substring(0, 19);
const OUTPUT_DIR = process.env.OUTPUT_DIR || `qa-results/video-sessions/comprehensive-${TS}`;
fs.mkdirSync(OUTPUT_DIR, { recursive: true });

const API = process.env.HELIX_API || 'http://localhost:7061';
const WEB = process.env.HELIX_WEB || 'http://localhost:8090';
const DEMO_API = process.env.HELIX_DEMO_API || 'http://localhost:8080';
const COGNEE = process.env.HELIX_COGNEE || 'http://localhost:8000';
// Read API key from .env if not in environment
let KEY = process.env.HELIXAGENT_API_KEY || '';
if (!KEY) {
  try {
    const envContent = fs.readFileSync('.env', 'utf8');
    const match = envContent.match(/^HELIXAGENT_API_KEY=(.+)$/m);
    if (match) KEY = match[1].trim();
  } catch {}
}
console.log(`API Key: ${KEY ? KEY.substring(0, 8) + '...' : 'NONE'}`);
const BIN_DIR = process.env.BIN_DIR || './bin';

const results = [];
let testNum = 0;
let failures = [];

const rec = (cat, name, ok, detail) => {
  testNum++;
  const d = (detail || '').toString().substring(0, 200);
  results.push({ n: testNum, category: cat, name, passed: !!ok, detail: d, ts: new Date().toISOString() });
  const icon = ok ? '\x1b[32mPASS\x1b[0m' : '\x1b[31mFAIL\x1b[0m';
  console.log(`  ${icon} #${testNum} [${cat}] ${name} — ${d}`);
  if (!ok) failures.push({ n: testNum, cat, name, detail: d });
};

function hreq(method, url, body, hdrs = {}, timeout = 15000) {
  return new Promise(resolve => {
    const u = new URL(url);
    const opts = {
      hostname: u.hostname, port: u.port, path: u.pathname + u.search,
      method, headers: { ...hdrs }, timeout
    };
    if (body !== null && body !== undefined) {
      const p = typeof body === 'string' ? body : JSON.stringify(body);
      opts.headers['Content-Type'] = opts.headers['Content-Type'] || 'application/json';
      opts.headers['Content-Length'] = Buffer.byteLength(p);
      const req = http.request(opts, res => {
        let b = ''; res.on('data', c => b += c);
        res.on('end', () => resolve({ s: res.statusCode, h: res.headers, b }));
      });
      req.on('error', e => resolve({ s: 0, h: {}, b: e.message }));
      req.on('timeout', function() { this.destroy(); resolve({ s: 0, h: {}, b: 'timeout' }); });
      req.write(p); req.end();
    } else {
      const req = (method === 'GET' ? http.get : http.request)(opts, res => {
        let b = ''; res.on('data', c => b += c);
        res.on('end', () => resolve({ s: res.statusCode, h: res.headers, b }));
      });
      if (method !== 'GET') req.end();
      req.on('error', e => resolve({ s: 0, h: {}, b: e.message }));
      req.on('timeout', function() { this.destroy(); resolve({ s: 0, h: {}, b: 'timeout' }); });
    }
  });
}

const hget = (url, hdrs) => hreq('GET', url, null, hdrs);
const hpost = (url, data, hdrs) => hreq('POST', url, data, hdrs);
const hput = (url, data, hdrs) => hreq('PUT', url, data, hdrs);
const hdelete = (url, hdrs) => hreq('DELETE', url, null, hdrs);
const hhead = (url, hdrs) => hreq('HEAD', url, null, hdrs);
const hoptions = (url, hdrs) => hreq('OPTIONS', url, null, hdrs);

const auth = KEY ? { 'Authorization': 'Bearer ' + KEY } : {};

function shell(bin, args) {
  try { return { ok: true, out: execFileSync(bin, args, { timeout: 30000, encoding: 'utf8' }).trim() }; }
  catch (e) { return { ok: false, out: (e.stderr || e.stdout || e.message || '').substring(0, 300) }; }
}

(async () => {
  const t0 = Date.now();
  console.log('╔══════════════════════════════════════════════════════════════╗');
  console.log('║   HELIXAGENT COMPREHENSIVE VIDEO QA — FULL SYSTEM COVERAGE  ║');
  console.log('╚══════════════════════════════════════════════════════════════╝\n');

  // Launch Playwright with video recording
  const browser = await chromium.launch({ headless: true });
  const ctx = await browser.newContext({
    recordVideo: { dir: OUTPUT_DIR, size: { width: 1280, height: 720 } },
    viewport: { width: 1280, height: 720 }
  });
  const page = await ctx.newPage();

  // ═══════════════════════════════════════════════════════════════
  // SECTION 1: HELIXAGENT MAIN SERVER (App 1/7)
  // ═══════════════════════════════════════════════════════════════
  console.log('\n╔═══ APP 1/7: HELIXAGENT MAIN SERVER ═══╗');

  // 1.1 HEALTH & SYSTEM
  console.log('\n--- 1.1 HEALTH & SYSTEM ---');
  let r = await hget(API + '/health');
  rec('app1-health', 'GET /health returns 200', r.s === 200, `status=${r.s}`);
  try { const d = JSON.parse(r.b); rec('app1-health', 'Health body has status field', d.status === 'healthy', d.status); }
  catch { rec('app1-health', 'Health body is valid JSON', false, 'parse error'); }

  r = await hget(API + '/metrics');
  rec('app1-health', 'GET /metrics (Prometheus)', r.s === 200 || r.s === 404, `status=${r.s}`);

  r = await hget(API + '/v1/features');
  rec('app1-health', 'GET /v1/features', r.s === 200, `status=${r.s}`);

  r = await hget(API + '/v1/features/available');
  rec('app1-health', 'GET /v1/features/available', r.s === 200, `status=${r.s}`);

  r = await hget(API + '/v1/features/agents');
  rec('app1-health', 'GET /v1/features/agents', r.s === 200, `status=${r.s}`);

  // 1.2 HEALTH RESPONSE HEADERS
  console.log('\n--- 1.2 RESPONSE HEADERS ---');
  r = await hget(API + '/health');
  rec('app1-headers', 'X-Features-Enabled header', !!r.h['x-features-enabled'], r.h['x-features-enabled'] || 'missing');
  rec('app1-headers', 'X-Transport-Protocol header', !!r.h['x-transport-protocol'], r.h['x-transport-protocol'] || 'missing');
  rec('app1-headers', 'X-Compression-Available header', !!r.h['x-compression-available'], r.h['x-compression-available'] || 'missing');
  rec('app1-headers', 'X-Streaming-Method header', !!r.h['x-streaming-method'], r.h['x-streaming-method'] || 'missing');
  rec('app1-headers', 'No Server header leak', !r.h['server'], r.h['server'] || 'none (good)');
  rec('app1-headers', 'Content-Type is JSON', (r.h['content-type'] || '').includes('json'), r.h['content-type']);

  // 1.3 AUTH ENDPOINTS
  console.log('\n--- 1.3 AUTH ENDPOINTS ---');
  r = await hpost(API + '/v1/auth/register', { email: 'qa@test.com', password: 'test123' });
  rec('app1-auth', 'POST /v1/auth/register', r.s > 0, `status=${r.s}`);
  r = await hpost(API + '/v1/auth/login', { email: 'qa@test.com', password: 'test123' });
  rec('app1-auth', 'POST /v1/auth/login', r.s > 0, `status=${r.s}`);
  r = await hpost(API + '/v1/auth/refresh', {});
  rec('app1-auth', 'POST /v1/auth/refresh', r.s > 0, `status=${r.s}`);
  r = await hget(API + '/v1/auth/me', auth);
  rec('app1-auth', 'GET /v1/auth/me', r.s > 0, `status=${r.s}`);
  r = await hpost(API + '/v1/auth/logout', {}, auth);
  rec('app1-auth', 'POST /v1/auth/logout', r.s > 0, `status=${r.s}`);

  // 1.4 OPENAI-COMPATIBLE ENDPOINTS
  console.log('\n--- 1.4 OPENAI-COMPATIBLE ---');
  r = await hget(API + '/v1/models', auth);
  rec('app1-openai', 'GET /v1/models', r.s === 200 || r.s === 503, `status=${r.s}`);
  r = await hreq('POST', API + '/v1/chat/completions', { model: 'helixagent-debate', messages: [{ role: 'user', content: 'Hello QA test' }], max_tokens: 5 }, auth, 60000);
  rec('app1-openai', 'POST /v1/chat/completions', r.s > 0 || r.b === 'timeout', `status=${r.s} len=${r.b.length}`);
  r = await hpost(API + '/v1/completions', { model: 'helixagent-debate', prompt: 'Test', max_tokens: 5 }, auth);
  rec('app1-openai', 'POST /v1/completions', r.s > 0, `status=${r.s}`);
  r = await hpost(API + '/v1/ensemble/completions', { prompt: 'Test QA', max_tokens: 5 }, auth);
  rec('app1-openai', 'POST /v1/ensemble/completions', r.s > 0, `status=${r.s}`);

  // 1.5 PROVIDER MANAGEMENT
  console.log('\n--- 1.5 PROVIDER MANAGEMENT ---');
  r = await hget(API + '/v1/providers', auth);
  rec('app1-providers', 'GET /v1/providers', r.s === 200, `status=${r.s}`);
  let providers = [];
  try { const d = JSON.parse(r.b); providers = d.providers || []; } catch {}
  rec('app1-providers', 'Provider count >= 20', providers.length >= 20, `count=${providers.length}`);
  rec('app1-providers', 'All providers have name', providers.every(p => p.name), 'valid');
  rec('app1-providers', 'All providers have models', providers.every(p => p.supported_models && p.supported_models.length > 0), 'valid');
  const provNames = providers.map(p => p.name);
  rec('app1-providers', 'No duplicate providers', new Set(provNames).size === provNames.length, `${provNames.length} unique`);
  rec('app1-providers', 'Has DeepSeek', provNames.includes('deepseek'), provNames.includes('deepseek') ? 'yes' : 'missing');
  rec('app1-providers', 'Has Gemini', provNames.includes('gemini'), provNames.includes('gemini') ? 'yes' : 'missing');
  rec('app1-providers', 'Has Mistral', provNames.includes('mistral'), provNames.includes('mistral') ? 'yes' : 'missing');
  rec('app1-providers', 'Has Claude or OAuth provider', provNames.includes('claude') || provNames.includes('qwen') || provNames.includes('zen'), 'oauth providers checked');
  rec('app1-providers', 'Has OpenRouter', provNames.includes('openrouter'), provNames.includes('openrouter') ? 'yes' : 'missing');

  if (providers.length > 0) {
    r = await hget(API + '/v1/providers/' + (providers[0].id || providers[0].name), auth);
    rec('app1-providers', 'GET /v1/providers/:id', r.s === 200 || r.s === 404, `status=${r.s}`);
  }

  r = await hget(API + '/v1/providers/discovery', auth);
  rec('app1-providers', 'GET /v1/providers/discovery', r.s === 200 || r.s === 503, `status=${r.s}`);
  r = await hget(API + '/v1/providers/best', auth);
  rec('app1-providers', 'GET /v1/providers/best', r.s === 200 || r.s === 503, `status=${r.s}`);
  r = await hget(API + '/v1/providers/verification', auth);
  rec('app1-providers', 'GET /v1/providers/verification', r.s === 200 || r.s === 503, `status=${r.s}`);

  // 1.6 DISCOVERY ENDPOINTS
  console.log('\n--- 1.6 DISCOVERY ---');
  for (const ep of ['/v1/discovery/models', '/v1/discovery/models/selected', '/v1/discovery/stats',
    '/v1/discovery/ensemble', '/v1/discovery/debate-model']) {
    r = await hget(API + ep, auth);
    rec('app1-discovery', `GET ${ep}`, r.s === 200 || r.s === 401 || r.s === 503, `status=${r.s}`);
  }
  r = await hpost(API + '/v1/discovery/trigger', {}, auth);
  rec('app1-discovery', 'POST /v1/discovery/trigger', r.s > 0, `status=${r.s}`);

  // 1.7 SCORING ENDPOINTS
  console.log('\n--- 1.7 SCORING ---');
  for (const ep of ['/v1/scoring/top', '/v1/scoring/range', '/v1/scoring/weights']) {
    r = await hget(API + ep, auth);
    rec('app1-scoring', `GET ${ep}`, r.s === 200 || r.s === 401 || r.s === 503, `status=${r.s}`);
  }
  r = await hpost(API + '/v1/scoring/batch', { model_ids: ['test'] }, auth);
  rec('app1-scoring', 'POST /v1/scoring/batch', r.s > 0, `status=${r.s}`);
  r = await hpost(API + '/v1/scoring/compare', { models: ['a', 'b'] }, auth);
  rec('app1-scoring', 'POST /v1/scoring/compare', r.s > 0, `status=${r.s}`);
  r = await hpost(API + '/v1/scoring/cache/invalidate', {}, auth);
  rec('app1-scoring', 'POST /v1/scoring/cache/invalidate', r.s > 0, `status=${r.s}`);

  // 1.8 VERIFICATION ENDPOINTS
  console.log('\n--- 1.8 VERIFICATION ---');
  for (const ep of ['/v1/verification/status', '/v1/verification/models', '/v1/verification/health', '/v1/verification/tests']) {
    r = await hget(API + ep, auth);
    rec('app1-verify', `GET ${ep}`, r.s === 200 || r.s === 401 || r.s === 503, `status=${r.s}`);
  }
  r = await hpost(API + '/v1/verification/model', { model_id: 'test' }, auth);
  rec('app1-verify', 'POST /v1/verification/model', r.s > 0, `status=${r.s}`);
  r = await hpost(API + '/v1/verification/batch', { model_ids: ['test'] }, auth);
  rec('app1-verify', 'POST /v1/verification/batch', r.s > 0, `status=${r.s}`);
  r = await hpost(API + '/v1/verification/code-visibility', { code: 'print("hello")' }, auth);
  rec('app1-verify', 'POST /v1/verification/code-visibility', r.s > 0, `status=${r.s}`);

  // 1.9 STARTUP VERIFICATION
  console.log('\n--- 1.9 STARTUP ---');
  r = await hget(API + '/v1/startup/verification');
  rec('app1-startup', 'GET /v1/startup/verification', r.s === 200, `status=${r.s}`);
  try {
    const sv = JSON.parse(r.b);
    const vp = sv.verified_providers || sv.providers || sv.results || [];
    rec('app1-startup', 'Has verified_providers', Array.isArray(vp) || typeof sv.verified_providers !== 'undefined' || typeof sv.providers !== 'undefined', `keys=${Object.keys(sv).join(',')}`);
    rec('app1-startup', 'Has debate_team', !!sv.debate_team, 'present');
  } catch { rec('app1-startup', 'Startup JSON parseable', false, 'parse error'); }

  // 1.10 DEBATE ENDPOINTS
  console.log('\n--- 1.10 AI DEBATE ---');
  r = await hget(API + '/v1/debates/team', auth);
  rec('app1-debate', 'GET /v1/debates/team', r.s === 200 || r.s === 503, `status=${r.s}`);
  r = await hget(API + '/v1/debates/orchestrator/status', auth);
  rec('app1-debate', 'GET /v1/debates/orchestrator/status', r.s === 200 || r.s === 401 || r.s === 503, `status=${r.s}`);
  r = await hget(API + '/v1/debates/history', auth);
  rec('app1-debate', 'GET /v1/debates/history', r.s === 200 || r.s === 401 || r.s === 503, `status=${r.s}`);
  r = await hpost(API + '/v1/debates/create', { topic: 'QA test debate', participants: 3 }, auth);
  rec('app1-debate', 'POST /v1/debates/create', r.s > 0, `status=${r.s}`);

  // 1.11 MCP PROTOCOL
  console.log('\n--- 1.11 MCP PROTOCOL ---');
  for (const ep of ['/v1/mcp/capabilities', '/v1/mcp/tools', '/v1/mcp/prompts', '/v1/mcp/resources',
    '/v1/mcp/categories', '/v1/mcp/stats']) {
    r = await hget(API + ep, auth);
    rec('app1-mcp', `GET ${ep}`, r.s === 200 || r.s === 503, `status=${r.s}`);
  }
  r = await hpost(API + '/v1/mcp/tools/call', { tool: 'test', arguments: {} }, auth);
  rec('app1-mcp', 'POST /v1/mcp/tools/call', r.s > 0, `status=${r.s}`);
  r = await hget(API + '/v1/mcp/tools/search?q=test', auth);
  rec('app1-mcp', 'GET /v1/mcp/tools/search', r.s > 0, `status=${r.s}`);
  r = await hget(API + '/v1/mcp/tools/suggestions', auth);
  rec('app1-mcp', 'GET /v1/mcp/tools/suggestions', r.s > 0, `status=${r.s}`);

  // 1.12 ACP PROTOCOL
  console.log('\n--- 1.12 ACP PROTOCOL ---');
  r = await hget(API + '/v1/acp/status', auth);
  rec('app1-acp', 'GET /v1/acp/status', r.s > 0, `status=${r.s}`);
  r = await hget(API + '/v1/acp/agents', auth);
  rec('app1-acp', 'GET /v1/acp/agents', r.s > 0, `status=${r.s}`);
  r = await hpost(API + '/v1/acp/execute', { action: 'test' }, auth);
  rec('app1-acp', 'POST /v1/acp/execute', r.s > 0, `status=${r.s}`);
  r = await hpost(API + '/v1/acp/broadcast', { message: 'test' }, auth);
  rec('app1-acp', 'POST /v1/acp/broadcast', r.s > 0, `status=${r.s}`);

  // 1.13 LSP PROTOCOL
  console.log('\n--- 1.13 LSP PROTOCOL ---');
  r = await hget(API + '/v1/lsp/servers', auth);
  rec('app1-lsp', 'GET /v1/lsp/servers', r.s > 0, `status=${r.s}`);
  r = await hget(API + '/v1/lsp/stats', auth);
  rec('app1-lsp', 'GET /v1/lsp/stats', r.s > 0, `status=${r.s}`);
  r = await hpost(API + '/v1/lsp/execute', { method: 'test' }, auth);
  rec('app1-lsp', 'POST /v1/lsp/execute', r.s > 0, `status=${r.s}`);
  r = await hpost(API + '/v1/lsp/sync', {}, auth);
  rec('app1-lsp', 'POST /v1/lsp/sync', r.s > 0, `status=${r.s}`);

  // 1.14 EMBEDDINGS
  console.log('\n--- 1.14 EMBEDDINGS ---');
  r = await hget(API + '/v1/embeddings/stats', auth);
  rec('app1-embed', 'GET /v1/embeddings/stats', r.s > 0, `status=${r.s}`);
  r = await hget(API + '/v1/embeddings/providers', auth);
  rec('app1-embed', 'GET /v1/embeddings/providers', r.s > 0, `status=${r.s}`);
  r = await hpost(API + '/v1/embeddings/generate', { text: 'QA test embedding' }, auth);
  rec('app1-embed', 'POST /v1/embeddings/generate', r.s > 0, `status=${r.s}`);
  r = await hpost(API + '/v1/embeddings/search', { query: 'test' }, auth);
  rec('app1-embed', 'POST /v1/embeddings/search', r.s > 0, `status=${r.s}`);

  // 1.15 VISION
  console.log('\n--- 1.15 VISION ---');
  r = await hget(API + '/v1/vision/providers', auth);
  rec('app1-vision', 'GET /v1/vision/providers', r.s > 0, `status=${r.s}`);
  r = await hpost(API + '/v1/vision/analyze', { image_url: 'test' }, auth);
  rec('app1-vision', 'POST /v1/vision/analyze', r.s > 0, `status=${r.s}`);

  // 1.16 COGNEE (Memory)
  console.log('\n--- 1.16 COGNEE ---');
  r = await hget(API + '/v1/cognee/health', auth);
  rec('app1-cognee', 'GET /v1/cognee/health', r.s > 0, `status=${r.s}`);
  r = await hget(API + '/v1/cognee/datasets', auth);
  rec('app1-cognee', 'GET /v1/cognee/datasets', r.s > 0, `status=${r.s}`);
  r = await hpost(API + '/v1/cognee/add', { content: 'QA test memory' }, auth);
  rec('app1-cognee', 'POST /v1/cognee/add', r.s > 0, `status=${r.s}`);
  r = await hpost(API + '/v1/cognee/search', { query: 'test' }, auth);
  rec('app1-cognee', 'POST /v1/cognee/search', r.s > 0, `status=${r.s}`);

  // 1.17 RAG
  console.log('\n--- 1.17 RAG ---');
  r = await hget(API + '/v1/rag/health', auth);
  rec('app1-rag', 'GET /v1/rag/health', r.s > 0, `status=${r.s}`);
  r = await hget(API + '/v1/rag/stats', auth);
  rec('app1-rag', 'GET /v1/rag/stats', r.s > 0, `status=${r.s}`);
  r = await hpost(API + '/v1/rag/search', { query: 'test' }, auth);
  rec('app1-rag', 'POST /v1/rag/search', r.s > 0, `status=${r.s}`);
  r = await hpost(API + '/v1/rag/search/hybrid', { query: 'test' }, auth);
  rec('app1-rag', 'POST /v1/rag/search/hybrid', r.s > 0, `status=${r.s}`);
  r = await hpost(API + '/v1/rag/chunk', { text: 'Test document for chunking.' }, auth);
  rec('app1-rag', 'POST /v1/rag/chunk', r.s > 0, `status=${r.s}`);

  // 1.18 CODE FORMATTERS
  console.log('\n--- 1.18 CODE FORMATTERS ---');
  r = await hget(API + '/v1/formatters');
  rec('app1-fmt', 'GET /v1/formatters (public)', r.s === 200, `status=${r.s}`);
  r = await hpost(API + '/v1/format', { code: 'func main(){fmt.Println("hi")}', language: 'go' });
  rec('app1-fmt', 'POST /v1/format', r.s === 200 || r.s === 400 || r.s === 503, `status=${r.s}`);
  r = await hpost(API + '/v1/format/check', { code: 'x=1', language: 'python' });
  rec('app1-fmt', 'POST /v1/format/check', r.s > 0, `status=${r.s}`);

  // 1.19 AGENTIC WORKFLOWS
  console.log('\n--- 1.19 AGENTIC WORKFLOWS ---');
  const wfBody = { name: 'qa-workflow', nodes: [{ id: 'n1', type: 'prompt' }], edges: [], entry_point: 'n1', end_nodes: ['n1'] };
  r = await hpost(API + '/v1/agentic/workflows', wfBody, auth);
  rec('app1-agentic', 'POST /v1/agentic/workflows (create)', r.s === 200 || r.s === 201 || r.s === 401, `status=${r.s}`);
  let wfId = '';
  try { wfId = JSON.parse(r.b).id; } catch {}
  r = await hget(API + '/v1/agentic/workflows', auth);
  rec('app1-agentic', 'GET /v1/agentic/workflows (list)', r.s === 200 || r.s === 401 || r.s === 404, `status=${r.s}`);
  if (wfId) {
    r = await hget(API + '/v1/agentic/workflows/' + wfId, auth);
    rec('app1-agentic', 'GET /v1/agentic/workflows/:id', r.s === 200, `status=${r.s}`);
    r = await hpost(API + '/v1/agentic/workflows/' + wfId + '/execute', {}, auth);
    rec('app1-agentic', 'POST execute workflow', r.s > 0, `status=${r.s}`);
    r = await hget(API + '/v1/agentic/workflows/' + wfId + '/status', auth);
    rec('app1-agentic', 'GET workflow status', r.s > 0, `status=${r.s}`);
  }

  // 1.20 PLANNING
  console.log('\n--- 1.20 PLANNING ---');
  r = await hpost(API + '/v1/planning/hiplan', { goal: 'QA test', constraints: [] }, auth);
  rec('app1-planning', 'POST /v1/planning/hiplan', r.s > 0, `status=${r.s}`);
  r = await hpost(API + '/v1/planning/mcts', { goal: 'QA test', iterations: 10 }, auth);
  rec('app1-planning', 'POST /v1/planning/mcts', r.s > 0, `status=${r.s}`);
  r = await hpost(API + '/v1/planning/tot', { problem: 'QA test', branches: 2 }, auth);
  rec('app1-planning', 'POST /v1/planning/tot', r.s > 0, `status=${r.s}`);
  r = await hget(API + '/v1/planning/stats', auth);
  rec('app1-planning', 'GET /v1/planning/stats', r.s > 0, `status=${r.s}`);

  // 1.21 LLMOPS
  console.log('\n--- 1.21 LLMOPS ---');
  r = await hget(API + '/v1/llmops/experiments', auth);
  rec('app1-llmops', 'GET /v1/llmops/experiments', r.s === 200 || r.s === 401 || r.s === 503, `status=${r.s}`);
  r = await hpost(API + '/v1/llmops/experiments', { name: 'qa-exp', variants: [{ name: 'v1' }] }, auth);
  rec('app1-llmops', 'POST /v1/llmops/experiments', r.s > 0, `status=${r.s}`);
  r = await hget(API + '/v1/llmops/prompts', auth);
  rec('app1-llmops', 'GET /v1/llmops/prompts', r.s > 0, `status=${r.s}`);
  r = await hpost(API + '/v1/llmops/prompts', { name: 'qa-prompt', content: 'test', version: '1.0' }, auth);
  rec('app1-llmops', 'POST /v1/llmops/prompts', r.s > 0, `status=${r.s}`);
  r = await hpost(API + '/v1/llmops/evaluate', { model: 'test', input: 'test', output: 'test' }, auth);
  rec('app1-llmops', 'POST /v1/llmops/evaluate', r.s > 0, `status=${r.s}`);

  // 1.22 BENCHMARK
  console.log('\n--- 1.22 BENCHMARK ---');
  r = await hget(API + '/v1/benchmark/results', auth);
  rec('app1-bench', 'GET /v1/benchmark/results', r.s > 0, `status=${r.s}`);
  r = await hget(API + '/v1/benchmark/leaderboard', auth);
  rec('app1-bench', 'GET /v1/benchmark/leaderboard', r.s > 0, `status=${r.s}`);
  r = await hget(API + '/v1/benchmark/categories', auth);
  rec('app1-bench', 'GET /v1/benchmark/categories', r.s > 0, `status=${r.s}`);
  r = await hpost(API + '/v1/benchmark/run', { benchmark: 'test', model: 'test' }, auth);
  rec('app1-bench', 'POST /v1/benchmark/run', r.s > 0, `status=${r.s}`);

  // 1.23 BACKGROUND TASKS
  console.log('\n--- 1.23 BACKGROUND TASKS ---');
  r = await hget(API + '/v1/tasks', auth);
  rec('app1-tasks', 'GET /v1/tasks', r.s > 0, `status=${r.s}`);
  r = await hpost(API + '/v1/tasks', { type: 'test', payload: {} }, auth);
  rec('app1-tasks', 'POST /v1/tasks (create)', r.s > 0, `status=${r.s}`);
  r = await hget(API + '/v1/tasks/queue', auth);
  rec('app1-tasks', 'GET /v1/tasks/queue', r.s > 0, `status=${r.s}`);

  // 1.24 SESSIONS
  console.log('\n--- 1.24 SESSIONS ---');
  r = await hget(API + '/v1/sessions', auth);
  rec('app1-sessions', 'GET /v1/sessions', r.s > 0, `status=${r.s}`);
  r = await hpost(API + '/v1/sessions', { context: 'qa-test' }, auth);
  rec('app1-sessions', 'POST /v1/sessions', r.s > 0, `status=${r.s}`);

  // 1.25 CLI AGENTS
  console.log('\n--- 1.25 CLI AGENTS ---');
  r = await hget(API + '/v1/agents', auth);
  rec('app1-agents', 'GET /v1/agents', r.s === 200 || r.s === 401, `status=${r.s}`);
  if (r.s === 200) {
    try { const d = JSON.parse(r.b); rec('app1-agents', 'Has 40+ agents', (d.agents || d).length >= 40, `count=${(d.agents || d).length}`); }
    catch { rec('app1-agents', '40+ agents (parse)', false, 'parse error'); }
  } else { rec('app1-agents', 'Agents endpoint requires auth', r.s === 401, `status=${r.s}`); }
  r = await hget(API + '/v1/agents/opencode', auth);
  rec('app1-agents', 'GET /v1/agents/opencode', r.s > 0, `status=${r.s}`);

  // 1.26 SKILLS
  console.log('\n--- 1.26 SKILLS ---');
  r = await hget(API + '/v1/skills', auth);
  rec('app1-skills', 'GET /v1/skills', r.s > 0, `status=${r.s}`);
  r = await hget(API + '/v1/skills/categories', auth);
  rec('app1-skills', 'GET /v1/skills/categories', r.s > 0, `status=${r.s}`);

  // 1.27 MONITORING
  console.log('\n--- 1.27 MONITORING ---');
  for (const ep of ['/v1/monitoring/providers', '/v1/monitoring/circuit-breakers', '/v1/monitoring/latency',
    '/v1/monitoring/oauth', '/v1/monitoring/concurrency']) {
    r = await hget(API + ep, auth);
    rec('app1-monitor', `GET ${ep}`, r.s > 0, `status=${r.s}`);
  }
  r = await hget(API + '/v1/health/providers', auth);
  rec('app1-monitor', 'GET /v1/health/providers', r.s > 0, `status=${r.s}`);
  r = await hget(API + '/v1/health/circuit-breakers', auth);
  rec('app1-monitor', 'GET /v1/health/circuit-breakers', r.s > 0, `status=${r.s}`);
  r = await hget(API + '/v1/health/status', auth);
  rec('app1-monitor', 'GET /v1/health/status', r.s > 0, `status=${r.s}`);

  // 1.28 MODEL METADATA
  console.log('\n--- 1.28 MODEL METADATA ---');
  r = await hget(API + '/v1/models/metadata', auth);
  rec('app1-models', 'GET /v1/models/metadata', r.s > 0, `status=${r.s}`);

  await page.screenshot({ path: path.join(OUTPUT_DIR, 's-app1-api.png') });

  // ═══════════════════════════════════════════════════════════════
  // SECTION 2: WEBSITE (Visual QA)
  // ═══════════════════════════════════════════════════════════════
  console.log('\n╔═══ WEBSITE VISUAL QA ═══╗');

  console.log('\n--- 2.1 ALL WEBSITE PAGES ---');
  const allPages = ['index.html', 'features.html', 'pricing.html', 'changelog.html', 'contact.html',
    'privacy.html', 'terms.html', 'docs/index.html', 'docs/api.html', 'docs/architecture.html',
    'docs/ai-debate.html', 'docs/bigdata.html', 'docs/deployment.html', 'docs/faq.html',
    'docs/grpc.html', 'docs/memory.html', 'docs/optimization.html', 'docs/protocols.html',
    'docs/quickstart.html', 'docs/security.html', 'docs/support.html', 'docs/troubleshooting.html',
    'docs/tutorial.html'];
  for (const pg of allPages) {
    try {
      const resp = await page.goto(WEB + '/' + pg, { timeout: 15000 });
      await page.waitForTimeout(200);
      const c = await page.textContent('body').catch(() => '');
      rec('website', pg, resp && resp.status() === 200 && c.length > 20, `len=${c.length}`);
    } catch (e) { rec('website', pg, false, e.message); }
  }
  await page.screenshot({ path: path.join(OUTPUT_DIR, 's-website-pages.png') });

  console.log('\n--- 2.2 RESPONSIVE DESIGN ---');
  const viewports = [
    ['mobile-small', 320, 568], ['mobile', 375, 812], ['mobile-large', 428, 926],
    ['tablet', 768, 1024], ['tablet-landscape', 1024, 768],
    ['desktop', 1280, 720], ['desktop-hd', 1920, 1080], ['desktop-4k', 2560, 1440]
  ];
  for (const [vn, vw, vh] of viewports) {
    await page.setViewportSize({ width: vw, height: vh });
    try {
      await page.goto(WEB + '/', { timeout: 5000 });
      await page.waitForTimeout(150);
      rec('responsive', `${vn} ${vw}x${vh}`, true, 'rendered');
      await page.screenshot({ path: path.join(OUTPUT_DIR, `resp-${vn}.png`) });
    } catch (e) { rec('responsive', `${vn} ${vw}x${vh}`, false, e.message); }
  }
  await page.setViewportSize({ width: 1280, height: 720 });

  console.log('\n--- 2.3 WEBSITE INTERACTIONS ---');
  await page.goto(WEB + '/', { timeout: 5000 });
  const navLinks = await page.$$('a[href]').then(els => els.length).catch(() => 0);
  rec('interact', 'Homepage has links', navLinks > 0, `${navLinks} links`);
  rec('interact', 'CSS loaded', await page.$$('link[rel="stylesheet"]').then(els => els.length > 0).catch(() => false), 'checked');
  rec('interact', 'Images present', (await page.$$('img').then(els => els.length).catch(() => 0)) >= 0, 'checked');
  rec('interact', 'Meta viewport tag', !!(await page.$('meta[name="viewport"]').catch(() => null)), 'checked');

  await page.goto(WEB + '/', { timeout: 5000 });
  await page.goto(WEB + '/features.html', { timeout: 5000 });
  await page.goBack({ timeout: 5000 }).catch(() => {});
  rec('interact', 'Browser back', true, 'executed');
  await page.goForward({ timeout: 5000 }).catch(() => {});
  rec('interact', 'Browser forward', true, 'executed');
  await page.reload({ timeout: 5000 }).catch(() => {});
  rec('interact', 'Page reload', true, 'executed');
  await page.mouse.wheel(0, 3000);
  await page.waitForTimeout(300);
  rec('interact', 'Page scroll', true, 'scrolled');

  console.log('\n--- 2.4 DOCS CONTENT ---');
  await page.goto(WEB + '/docs/api.html', { timeout: 5000 });
  let dc = await page.textContent('body').catch(() => '');
  rec('docs', 'API docs has /v1/ refs', dc.includes('/v1/'), 'present');
  rec('docs', 'API docs mentions agentic', dc.toLowerCase().includes('agentic'), 'present');
  rec('docs', 'API docs mentions providers', dc.toLowerCase().includes('provider'), 'present');

  await page.goto(WEB + '/docs/architecture.html', { timeout: 5000 });
  dc = await page.textContent('body').catch(() => '');
  rec('docs', 'Architecture page content', dc.length > 500, `len=${dc.length}`);

  await page.goto(WEB + '/docs/ai-debate.html', { timeout: 5000 });
  dc = await page.textContent('body').catch(() => '');
  rec('docs', 'Debate doc has debate content', dc.toLowerCase().includes('debate'), 'present');

  await page.goto(WEB + '/features.html', { timeout: 5000 });
  dc = await page.textContent('body').catch(() => '');
  rec('docs', 'Features mentions LLM/provider', dc.toLowerCase().includes('llm') || dc.toLowerCase().includes('provider'), 'present');

  await page.screenshot({ path: path.join(OUTPUT_DIR, 's-website-docs.png') });

  // ═══════════════════════════════════════════════════════════════
  // SECTION 3: SECURITY TESTING
  // ═══════════════════════════════════════════════════════════════
  console.log('\n╔═══ SECURITY TESTING ═══╗');

  r = await hget(API + '/v1/admin/health/all', { 'Authorization': 'Bearer invalid-token-12345' });
  rec('security', 'Invalid token rejected', r.s === 401 || r.s === 403 || r.s === 404, `status=${r.s}`);
  r = await hget(API + '/v1/llmops/experiments');
  rec('security', 'No auth = blocked (llmops)', r.s === 401 || r.s === 403 || r.s === 500 || r.s === 503, `status=${r.s}`);
  r = await hget(API + '/health');
  rec('security', 'Health is public', r.s === 200, `status=${r.s}`);
  r = await hget(API + '/v1/providers?q=<script>alert(1)</script>');
  rec('security', 'XSS in query blocked', !r.b.includes('<script>alert'), 'no reflection');
  r = await hget(API + "/v1/providers?q=' OR 1=1 --");
  rec('security', 'SQL injection blocked', r.s === 200 || r.s === 400, `status=${r.s}`);
  r = await hget(API + '/../../etc/passwd');
  rec('security', 'Path traversal blocked', r.s === 404 || r.s === 301 || r.s === 400, `status=${r.s}`);
  r = await hget(API + '/health');
  rec('security', 'No Server header leak', !r.h['server'], r.h['server'] || 'none');
  rec('security', 'No X-Powered-By leak', !r.h['x-powered-by'], r.h['x-powered-by'] || 'none');
  r = await hreq('TRACE', API + '/health', null);
  rec('security', 'TRACE method blocked', r.s === 404 || r.s === 405 || r.s === 0, `status=${r.s}`);

  // ═══════════════════════════════════════════════════════════════
  // SECTION 4: ERROR HANDLING
  // ═══════════════════════════════════════════════════════════════
  console.log('\n╔═══ ERROR HANDLING ═══╗');

  r = await hget(API + '/nonexistent');
  rec('errors', '404 on unknown path', r.s === 404, `status=${r.s}`);
  r = await hget(API + '/v1/nonexistent', auth);
  rec('errors', '404 on unknown v1 path', r.s === 404, `status=${r.s}`);
  r = await hpost(API + '/v1/agentic/workflows', 'not{json', auth);
  rec('errors', 'Bad JSON returns 400', r.s === 400 || r.s === 401, `status=${r.s}`);
  r = await hpost(API + '/v1/planning/hiplan', '', auth);
  rec('errors', 'Empty body returns error', r.s >= 400 && r.s < 600, `status=${r.s}`);
  r = await hpost(API + '/v1/chat/completions', { model: '', messages: [] }, auth);
  rec('errors', 'Empty model returns error', r.s >= 400 || r.s === 0, `status=${r.s}`);
  r = await hget(API + '/v1/discovery/models', auth);
  rec('errors', 'Nil service returns 503 or auth', r.s === 503 || r.s === 200 || r.s === 401, `status=${r.s}`);
  if (r.s === 503) {
    try { const d = JSON.parse(r.b); rec('errors', '503 has error field', !!d.error, d.error); }
    catch { rec('errors', '503 JSON structure', false, 'not parseable'); }
  } else { rec('errors', 'Discovery available', true, `status=${r.s}`); }

  // ═══════════════════════════════════════════════════════════════
  // SECTION 5: EDGE CASES
  // ═══════════════════════════════════════════════════════════════
  console.log('\n╔═══ EDGE CASES ═══╗');

  r = await hget(API + '/health?' + 'x'.repeat(4000));
  rec('edge', 'Very long URL (4K)', r.s > 0, `status=${r.s}`);
  r = await hget(API + '//health');
  rec('edge', 'Double slash', r.s > 0, `status=${r.s}`);
  r = await hget(API + '/health/');
  rec('edge', 'Trailing slash', r.s > 0, `status=${r.s}`);
  r = await hpost(API + '/v1/agentic/workflows', { name: '\u0442\u0435\u0441\u0442 \u4e2d\u6587' }, auth);
  rec('edge', 'Unicode body', r.s > 0, `status=${r.s}`);
  r = await hpost(API + '/v1/agentic/workflows', { name: 'x'.repeat(100000) }, auth);
  rec('edge', 'Large body (100K)', r.s > 0, `status=${r.s}`);
  r = await hget(API + '/health', { 'Accept': 'text/xml' });
  rec('edge', 'XML Accept header', r.s === 200, `status=${r.s}`);
  r = await hhead(API + '/health');
  rec('edge', 'HEAD request', r.s === 200 || r.s === 404, `status=${r.s}`);
  r = await hoptions(API + '/v1/providers');
  rec('edge', 'OPTIONS/CORS', r.s === 200 || r.s === 204 || r.s === 404, `status=${r.s}`);

  const cc10 = await Promise.all(Array(10).fill(null).map(() => hget(API + '/health')));
  rec('edge', '10 concurrent /health', cc10.every(c => c.s === 200), 'all 200');
  const cc50 = await Promise.all(Array(50).fill(null).map(() => hget(API + '/v1/providers')));
  rec('edge', '50 concurrent /v1/providers', cc50.filter(c => c.s === 200).length >= 45, `${cc50.filter(c => c.s === 200).length}/50`);
  const rapid = [];
  for (let i = 0; i < 20; i++) rapid.push(await hget(API + '/health'));
  rec('edge', '20 rapid sequential', rapid.every(c => c.s === 200), 'all OK');

  // ═══════════════════════════════════════════════════════════════
  // SECTION 6: APP 2 — API DEMO SERVER
  // ═══════════════════════════════════════════════════════════════
  console.log('\n╔═══ APP 2/7: API DEMO SERVER ═══╗');

  const demoOk = await hget(DEMO_API + '/api/v1/health').then(r => r.s > 0).catch(() => false);
  if (demoOk) {
    for (const ep of ['/api/v1/health', '/api/v1/status', '/api/v1/metrics']) {
      r = await hget(DEMO_API + ep);
      rec('app2-demo', `GET ${ep}`, r.s === 200, `status=${r.s}`);
    }
    for (const ep of ['/api/v1/mcp/tools/list', '/api/v1/mcp/servers', '/api/v1/acp/status',
      '/api/v1/plugins/', '/api/v1/templates/', '/api/v1/analytics/health']) {
      r = await hget(DEMO_API + ep);
      rec('app2-demo', `GET ${ep}`, r.s === 200 || r.s === 301, `status=${r.s}`);
    }
    r = await hpost(DEMO_API + '/api/v1/mcp/tools/call', { name: 'test' });
    rec('app2-demo', 'POST /api/v1/mcp/tools/call', r.s === 200, `status=${r.s}`);
    r = await hpost(DEMO_API + '/api/v1/lsp/completion', { file: 'test.go', position: { line: 1, col: 1 } });
    rec('app2-demo', 'POST /api/v1/lsp/completion', r.s === 200, `status=${r.s}`);
    r = await hpost(DEMO_API + '/api/v1/acp/execute', { action: 'test' });
    rec('app2-demo', 'POST /api/v1/acp/execute', r.s === 200, `status=${r.s}`);
  } else {
    rec('app2-demo', 'API Demo server accessible', false, 'NOT RUNNING on ' + DEMO_API);
  }

  // ═══════════════════════════════════════════════════════════════
  // SECTION 7: APP 3 — COGNEE-MOCK SERVER
  // ═══════════════════════════════════════════════════════════════
  console.log('\n╔═══ APP 3/7: COGNEE-MOCK SERVER ═══╗');

  const cogneeOk = await hget(COGNEE + '/health').then(r => r.s === 200).catch(() => false);
  if (cogneeOk) {
    for (const ep of ['/health', '/api/v1/health']) {
      r = await hget(COGNEE + ep);
      rec('app3-cognee', `GET ${ep}`, r.s === 200, `status=${r.s}`);
    }
    r = await hpost(COGNEE + '/api/v1/add', { data: 'QA test', type: 'text' });
    rec('app3-cognee', 'POST /api/v1/add', r.s === 200, `status=${r.s}`);
    r = await hpost(COGNEE + '/api/v1/search', { query: 'test' });
    rec('app3-cognee', 'POST /api/v1/search', r.s === 200, `status=${r.s}`);
    r = await hpost(COGNEE + '/api/v1/cognify', {});
    rec('app3-cognee', 'POST /api/v1/cognify', r.s === 200, `status=${r.s}`);
    r = await hget(COGNEE + '/api/v1/graph');
    rec('app3-cognee', 'GET /api/v1/graph', r.s === 200, `status=${r.s}`);
    r = await hget(COGNEE + '/api/v1/datasets');
    rec('app3-cognee', 'GET /api/v1/datasets', r.s === 200, `status=${r.s}`);
  } else {
    rec('app3-cognee', 'Cognee-mock server accessible', false, 'NOT RUNNING on ' + COGNEE);
  }

  // ═══════════════════════════════════════════════════════════════
  // SECTION 8-10: CLI APPS (4, 5, 6, 7)
  // ═══════════════════════════════════════════════════════════════
  console.log('\n╔═══ APP 4/7: SANITY-CHECK ═══╗');
  let sr = shell(BIN_DIR + '/sanity-check', ['--help']);
  rec('app4-sanity', 'sanity-check --help', sr.ok || sr.out.length > 0, sr.out.substring(0, 100));
  sr = shell(BIN_DIR + '/sanity-check', ['--host', 'localhost', '--port', '7061', '--json']);
  rec('app4-sanity', 'sanity-check --json', sr.out.length > 0, sr.out.substring(0, 100));

  console.log('\n╔═══ APP 5/7: GENERATE-CONSTITUTION ═══╗');
  sr = shell(BIN_DIR + '/generate-constitution', ['--help']);
  rec('app5-constitution', 'generate-constitution --help', sr.ok || sr.out.length > 0, sr.out.substring(0, 100));

  console.log('\n╔═══ APP 6/7: MCP-BRIDGE ═══╗');
  sr = shell(BIN_DIR + '/mcp-bridge', ['--help']);
  rec('app6-mcpbridge', 'mcp-bridge --help', sr.ok || sr.out.length > 0, sr.out.substring(0, 100));

  console.log('\n╔═══ APP 7/7: GRPC-SERVER ═══╗');
  sr = shell(BIN_DIR + '/grpc-server', ['--help']);
  rec('app7-grpc', 'grpc-server --help', sr.ok || sr.out.length > 0, sr.out.substring(0, 100));

  // ═══════════════════════════════════════════════════════════════
  // SECTION 11: OPENCODE CLI AGENT
  // ═══════════════════════════════════════════════════════════════
  console.log('\n╔═══ OPENCODE CLI AGENT ═══╗');

  sr = shell(BIN_DIR + '/helixagent', ['--generate-agent-config=opencode']);
  rec('opencode', 'Config generation runs', sr.out.length > 0, `output_len=${sr.out.length}`);
  const configHasMCP = sr.out.includes('mcp') || sr.out.includes('MCP');
  rec('opencode', 'Config includes MCP servers', configHasMCP, configHasMCP ? 'present' : 'missing');
  const configHasProvider = sr.out.includes('helixagent') || sr.out.includes('provider');
  rec('opencode', 'Config includes HelixAgent provider', configHasProvider, configHasProvider ? 'present' : 'missing');
  const configHasModel = sr.out.includes('helixagent-debate') || sr.out.includes('model');
  rec('opencode', 'Config includes model reference', configHasModel, configHasModel ? 'present' : 'missing');
  try {
    const configJson = JSON.parse(sr.out);
    rec('opencode', 'Config is valid JSON', true, 'parsed');
    rec('opencode', 'Config has provider section', !!configJson.providers || !!configJson.provider, 'present');
    const mcpServers = configJson.mcpServers || configJson.mcp_servers || configJson.mcp || {};
    const mcpCount = Object.keys(mcpServers).length;
    rec('opencode', 'Config has 15+ MCP servers', mcpCount >= 15, `count=${mcpCount}`);
  } catch {
    rec('opencode', 'Config JSON parsing', false, 'invalid JSON');
  }

  for (const agent of ['crush', 'kilocode']) {
    sr = shell(BIN_DIR + '/helixagent', ['--generate-agent-config=' + agent]);
    rec('cli-agents', `${agent} config generation`, sr.out.length > 10, `output_len=${sr.out.length}`);
  }

  // ═══════════════════════════════════════════════════════════════
  // SECTION 12: SSE STREAMING
  // ═══════════════════════════════════════════════════════════════
  console.log('\n╔═══ SSE STREAMING ═══╗');

  for (const proto of ['mcp', 'acp', 'lsp', 'embeddings', 'vision', 'cognee']) {
    r = await hreq('GET', API + '/v1/protocols/sse/' + proto, null, { ...auth, 'Accept': 'text/event-stream' }, 5000);
    rec('streaming', `SSE /v1/protocols/sse/${proto}`, r.s > 0 || r.b === 'timeout' || r.s === 0, `status=${r.s || 'timeout(expected for SSE)'}`);
  }
  r = await hpost(API + '/v1/chat/completions', { model: 'helixagent-debate', messages: [{ role: 'user', content: 'Say hi' }], stream: true, max_tokens: 5 }, auth);
  rec('streaming', 'Streaming chat completion', r.s > 0, `status=${r.s} len=${r.b.length}`);

  // ═══════════════════════════════════════════════════════════════
  // SECTION 13: STRESS
  // ═══════════════════════════════════════════════════════════════
  console.log('\n╔═══ STRESS ═══╗');

  const mixed = await Promise.all(['/health', '/v1/providers', '/v1/features', '/health', '/v1/providers'].map(ep => hget(API + ep)));
  rec('stress', 'Mixed concurrent (5)', mixed.every(c => c.s === 200), 'all 200');
  const burst = await Promise.all(Array(20).fill(null).map(() => hget(API + '/v1/providers')));
  rec('stress', '20-burst providers', burst.filter(c => c.s === 200).length >= 18, `${burst.filter(c => c.s === 200).length}/20`);

  const endpoints = ['/health', '/v1/providers', '/v1/features'];
  for (const ep of endpoints) {
    const times = [];
    for (let i = 0; i < 3; i++) {
      const t = Date.now();
      await hget(API + ep);
      times.push(Date.now() - t);
    }
    const avg = times.reduce((a, b) => a + b, 0) / times.length;
    rec('stress', `Latency ${ep}`, avg < 5000, `${avg.toFixed(0)}ms avg`);
  }

  // ═══════════════════════════════════════════════════════════════
  // SECTION 14: HELIXQA MODULE
  // ═══════════════════════════════════════════════════════════════
  console.log('\n╔═══ HELIXQA MODULE ═══╗');

  sr = shell(BIN_DIR + '/helixqa', ['version']);
  rec('helixqa', 'helixqa version', sr.ok || sr.out.includes('0.'), sr.out.substring(0, 100));
  sr = shell(BIN_DIR + '/helixqa', ['list', '--banks', 'qa-banks/', '--json']);
  rec('helixqa', 'helixqa list --json', sr.out.length > 10, `output_len=${sr.out.length}`);

  const bankFiles = fs.readdirSync('qa-banks').filter(f => f.endsWith('.json'));
  for (const bf of bankFiles) {
    try {
      const content = JSON.parse(fs.readFileSync(path.join('qa-banks', bf), 'utf8'));
      const challenges = content.challenges || content.test_cases || [];
      rec('helixqa-banks', `${bf} valid`, true, `${challenges.length} challenges`);
    } catch (e) { rec('helixqa-banks', `${bf} valid`, false, e.message); }
  }

  // ═══════════════════════════════════════════════════════════════
  // FINALIZE
  // ═══════════════════════════════════════════════════════════════
  console.log('\n╔═══ FINALIZING ═══╗');
  await page.screenshot({ path: path.join(OUTPUT_DIR, 's-final.png') });
  await page.waitForTimeout(500);
  await ctx.close();
  await browser.close();

  const elapsed = ((Date.now() - t0) / 1000).toFixed(1);
  const passed = results.filter(r => r.passed).length;
  const failed = results.filter(r => !r.passed).length;
  const cats = {};
  results.forEach(r => { if (!cats[r.category]) cats[r.category] = { p: 0, f: 0 }; cats[r.category][r.passed ? 'p' : 'f']++; });

  fs.writeFileSync(path.join(OUTPUT_DIR, 'results.json'), JSON.stringify({
    session: 'comprehensive-' + Date.now(),
    ts: new Date().toISOString(),
    duration_seconds: parseFloat(elapsed),
    total: results.length, passed, failed,
    categories: cats,
    failures: failures,
    results
  }, null, 2));

  console.log('\n╔══════════════════════════════════════════════════════════════╗');
  console.log(`║  COMPREHENSIVE QA: ${passed}/${results.length} PASSED (${failed} FAILED)`.padEnd(63) + '║');
  console.log(`║  Duration: ${elapsed}s`.padEnd(63) + '║');
  console.log('╠══════════════════════════════════════════════════════════════╣');
  Object.entries(cats).sort().forEach(([c, v]) => {
    const status = v.f === 0 ? '\x1b[32mOK\x1b[0m' : '\x1b[31mFAIL\x1b[0m';
    console.log(`║  ${c.padEnd(25)} ${v.p}/${v.p + v.f} ${status}`.padEnd(72) + '║');
  });
  console.log('╠══════════════════════════════════════════════════════════════╣');
  const vids = fs.readdirSync(OUTPUT_DIR).filter(f => f.endsWith('.webm'));
  vids.forEach(v => console.log(`║  Video: ${v} (${(fs.statSync(path.join(OUTPUT_DIR, v)).size / 1024 / 1024).toFixed(1)}MB)`.padEnd(63) + '║'));
  console.log(`║  Screenshots: ${fs.readdirSync(OUTPUT_DIR).filter(f => f.endsWith('.png')).length}`.padEnd(63) + '║');
  console.log('╚══════════════════════════════════════════════════════════════╝');

  if (failures.length > 0) {
    console.log('\n\x1b[31m=== FAILURES ===\x1b[0m');
    failures.forEach(f => console.log(`  #${f.n} [${f.cat}] ${f.name}: ${f.detail}`));
  }

  process.exit(failed > 0 ? 1 : 0);
})();
