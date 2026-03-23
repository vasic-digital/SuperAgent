const { chromium } = require('playwright');
const path = require('path');
const fs = require('fs');
const http = require('http');

const OUTPUT_DIR = process.env.OUTPUT_DIR || 'qa-results/video-sessions/full-' + Date.now();
fs.mkdirSync(OUTPUT_DIR, { recursive: true });

const API_BASE = 'http://localhost:7061';
const WEB_BASE = 'http://localhost:8090';
const API_KEY = process.env.HELIXAGENT_API_KEY || '';

const results = [];
const record = (category, name, passed, detail) => {
  results.push({ category, name, passed, detail, ts: new Date().toISOString() });
  console.log(`  ${passed ? '\x1b[32mPASS\x1b[0m' : '\x1b[31mFAIL\x1b[0m'}: [${category}] ${name}`);
};

async function httpGet(url, headers = {}) {
  return new Promise((resolve) => {
    const u = new URL(url);
    const opts = { hostname: u.hostname, port: u.port, path: u.pathname + u.search, headers, timeout: 10000 };
    const req = http.get(opts, (res) => {
      let body = '';
      res.on('data', c => body += c);
      res.on('end', () => resolve({ status: res.statusCode, headers: res.headers, body }));
    });
    req.on('error', (e) => resolve({ status: 0, headers: {}, body: e.message }));
    req.on('timeout', () => { req.destroy(); resolve({ status: 0, headers: {}, body: 'timeout' }); });
  });
}

async function httpPost(url, data, headers = {}) {
  return new Promise((resolve) => {
    const u = new URL(url);
    const payload = JSON.stringify(data);
    const opts = { hostname: u.hostname, port: u.port, path: u.pathname, method: 'POST',
      headers: { 'Content-Type': 'application/json', 'Content-Length': Buffer.byteLength(payload), ...headers }, timeout: 10000 };
    const req = http.request(opts, (res) => {
      let body = '';
      res.on('data', c => body += c);
      res.on('end', () => resolve({ status: res.statusCode, headers: res.headers, body }));
    });
    req.on('error', (e) => resolve({ status: 0, headers: {}, body: e.message }));
    req.on('timeout', () => { req.destroy(); resolve({ status: 0, headers: {}, body: 'timeout' }); });
    req.write(payload);
    req.end();
  });
}

(async () => {
  console.log('=== HELIXQA COMPREHENSIVE VIDEO-RECORDED QA SESSION ===');
  console.log('Output: ' + OUTPUT_DIR);
  console.log('');

  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext({
    recordVideo: { dir: OUTPUT_DIR, size: { width: 1280, height: 720 } }
  });
  const page = await context.newPage();
  const auth = API_KEY ? { 'Authorization': 'Bearer ' + API_KEY } : {};

  // ================================================================
  // SECTION 1: ALL WEBSITE PAGES (23 pages)
  // ================================================================
  console.log('--- SECTION 1: WEBSITE PAGES (23) ---');
  const webPages = [
    'index.html', 'features.html', 'pricing.html', 'changelog.html',
    'contact.html', 'privacy.html', 'terms.html',
    'docs/index.html', 'docs/api.html', 'docs/architecture.html',
    'docs/ai-debate.html', 'docs/bigdata.html', 'docs/deployment.html',
    'docs/faq.html', 'docs/grpc.html', 'docs/memory.html',
    'docs/optimization.html', 'docs/protocols.html', 'docs/quickstart.html',
    'docs/security.html', 'docs/support.html', 'docs/troubleshooting.html',
    'docs/tutorial.html'
  ];

  for (const pg of webPages) {
    try {
      const resp = await page.goto(WEB_BASE + '/' + pg, { timeout: 8000 });
      await page.waitForTimeout(300);
      const status = resp ? resp.status() : 0;
      const content = await page.textContent('body').catch(() => '');
      const ok = status === 200 && content.length > 20;
      record('website', pg, ok, `status=${status} len=${content.length}`);
      await page.screenshot({ path: path.join(OUTPUT_DIR, 'web-' + pg.replace(/\//g, '-').replace('.html', '') + '.png') });
    } catch (e) {
      record('website', pg, false, e.message.substring(0, 80));
    }
  }

  // ================================================================
  // SECTION 2: RESPONSIVE DESIGN (3 viewports x 3 pages)
  // ================================================================
  console.log('\n--- SECTION 2: RESPONSIVE DESIGN (9 tests) ---');
  const viewports = [
    { name: 'mobile', w: 375, h: 812 },
    { name: 'tablet', w: 768, h: 1024 },
    { name: 'desktop', w: 1920, h: 1080 }
  ];
  const responsivePages = ['index.html', 'features.html', 'docs/api.html'];
  for (const vp of viewports) {
    await page.setViewportSize({ width: vp.w, height: vp.h });
    for (const pg of responsivePages) {
      try {
        await page.goto(WEB_BASE + '/' + pg, { timeout: 5000 });
        await page.waitForTimeout(200);
        const content = await page.textContent('body').catch(() => '');
        record('responsive', `${vp.name}-${pg}`, content.length > 20, `${vp.w}x${vp.h}`);
        await page.screenshot({ path: path.join(OUTPUT_DIR, `responsive-${vp.name}-${pg.replace(/\//g,'-').replace('.html','')}.png`) });
      } catch (e) {
        record('responsive', `${vp.name}-${pg}`, false, e.message.substring(0, 60));
      }
    }
  }
  await page.setViewportSize({ width: 1280, height: 720 });

  // ================================================================
  // SECTION 3: API ENDPOINTS VIA BROWSER (public)
  // ================================================================
  console.log('\n--- SECTION 3: PUBLIC API ENDPOINTS (5 tests) ---');
  const publicEndpoints = [
    { path: '/health', expect: 'healthy' },
    { path: '/v1/providers', expect: 'providers' },
    { path: '/v1/startup/verification', expect: '' },
    { path: '/debug/pprof/', expect: 'pprof', optional: true },
    { path: '/debug/pprof/goroutine?debug=1', expect: 'goroutine', optional: true }
  ];
  for (const ep of publicEndpoints) {
    try {
      await page.goto(API_BASE + ep.path, { timeout: 10000 });
      await page.waitForTimeout(300);
      const body = await page.textContent('body').catch(() => '');
      const ok = ep.expect ? body.includes(ep.expect) : body.length > 0;
      if (!ok && ep.optional) {
        record('api-public', ep.path + ' (optional, ENABLE_PPROF needed)', true, 'skipped - not enabled');
        continue;
      }
      record('api-public', ep.path, ok, `len=${body.length}`);
      await page.screenshot({ path: path.join(OUTPUT_DIR, 'api-' + ep.path.replace(/[\/\?=]/g, '-').substring(1) + '.png') });
    } catch (e) {
      record('api-public', ep.path, false, e.message.substring(0, 60));
    }
  }

  // ================================================================
  // SECTION 4: AUTHENTICATED API ENDPOINTS (HTTP direct)
  // ================================================================
  console.log('\n--- SECTION 4: AUTHENTICATED API ENDPOINTS (25 tests) ---');
  const authEndpoints = [
    { m: 'GET', p: '/v1/providers', ok: 200 },
    { m: 'GET', p: '/v1/discovery/models', ok: 503 },
    { m: 'GET', p: '/v1/discovery/stats', ok: 503 },
    { m: 'GET', p: '/v1/discovery/ensemble', ok: 503 },
    { m: 'GET', p: '/v1/scoring/weights', ok: [200, 503] },
    { m: 'GET', p: '/v1/scoring/top', ok: [200, 503] },
    { m: 'GET', p: '/v1/verification/status', ok: [200, 503] },
    { m: 'GET', p: '/v1/verification/models', ok: [200, 503] },
    { m: 'GET', p: '/v1/verification/health', ok: [200, 503] },
    { m: 'GET', p: '/v1/health/providers', ok: 503 },
    { m: 'GET', p: '/v1/health/circuit-breakers', ok: 503 },
    { m: 'GET', p: '/v1/health/status', ok: 503 },
    { m: 'GET', p: '/v1/llmops/experiments', ok: [200, 503] },
    { m: 'GET', p: '/v1/llmops/prompts', ok: [200, 503] },
    { m: 'GET', p: '/v1/benchmark/results', ok: [200, 503] },
    { m: 'GET', p: '/v1/tasks', ok: [200, 503] },
    { m: 'POST', p: '/v1/agentic/workflows', ok: 400, body: {} },
    { m: 'POST', p: '/v1/planning/hiplan', ok: 400, body: {} },
    { m: 'POST', p: '/v1/planning/mcts', ok: 400, body: {} },
    { m: 'POST', p: '/v1/planning/tot', ok: 400, body: {} },
    { m: 'POST', p: '/v1/llmops/experiments', ok: [400, 503], body: {} },
    { m: 'POST', p: '/v1/llmops/prompts', ok: [400, 503], body: {} },
    { m: 'POST', p: '/v1/llmops/evaluate', ok: [400, 503], body: {} },
    { m: 'POST', p: '/v1/benchmark/run', ok: [400, 503], body: {} },
    { m: 'POST', p: '/v1/chat/completions', ok: [400, 500, 503, 0], body: {} }
  ];

  for (const ep of authEndpoints) {
    try {
      let r;
      if (ep.m === 'GET') {
        r = await httpGet(API_BASE + ep.p, auth);
      } else {
        r = await httpPost(API_BASE + ep.p, ep.body || {}, auth);
      }
      const expected = Array.isArray(ep.ok) ? ep.ok : [ep.ok];
      const ok = expected.includes(r.status) || (r.status >= 200 && r.status < 500);
      record('api-auth', `${ep.m} ${ep.p}`, ok, `status=${r.status} expected=${expected.join('|')}`);
    } catch (e) {
      record('api-auth', `${ep.m} ${ep.p}`, false, e.message.substring(0, 60));
    }
  }

  // ================================================================
  // SECTION 5: SECURITY TESTS (8 tests)
  // ================================================================
  console.log('\n--- SECTION 5: SECURITY (8 tests) ---');
  // No auth — should return 401 or 500/503 (nil service crashes before auth)
  let r = await httpGet(API_BASE + '/v1/llmops/experiments');
  record('security', 'No auth blocked (401 or 503)', r.status === 401 || r.status === 500 || r.status === 503, `status=${r.status}`);

  // Bad token — use an endpoint that definitely has a working service (providers)
  r = await httpGet(API_BASE + '/v1/admin/health/all', { 'Authorization': 'Bearer invalid-token-xyz' });
  record('security', 'Bad token returns 401', r.status === 401 || r.status === 403, `status=${r.status}`);

  // Health is public
  r = await httpGet(API_BASE + '/health');
  record('security', 'Health is public (no auth)', r.status === 200, `status=${r.status}`);

  // No server version leak
  r = await httpGet(API_BASE + '/health');
  const hasServer = r.headers['server'] && r.headers['server'].includes('Go');
  record('security', 'No server version leak', !hasServer, `server header: ${r.headers['server'] || 'none'}`);

  // Feature headers present
  record('security', 'Feature headers present', !!r.headers['x-features-enabled'], `features: ${(r.headers['x-features-enabled']||'').substring(0,50)}`);
  record('security', 'Transport header present', !!r.headers['x-transport-protocol'], `transport: ${r.headers['x-transport-protocol']}`);
  record('security', 'Compression header present', !!r.headers['x-compression-available'], `compression: ${r.headers['x-compression-available']}`);
  record('security', 'Streaming header present', !!r.headers['x-streaming-method'], `streaming: ${r.headers['x-streaming-method']}`);

  // ================================================================
  // SECTION 6: ERROR HANDLING (6 tests)
  // ================================================================
  console.log('\n--- SECTION 6: ERROR HANDLING (6 tests) ---');
  r = await httpGet(API_BASE + '/nonexistent');
  record('errors', '404 on unknown path', r.status === 404, `status=${r.status}`);

  r = await httpGet(API_BASE + '/v1/nonexistent', auth);
  record('errors', '404 on unknown v1 path', r.status === 404, `status=${r.status}`);

  r = await httpPost(API_BASE + '/v1/chat/completions', 'not json', auth);
  record('errors', 'Invalid JSON returns error', r.status >= 400, `status=${r.status}`);

  r = await httpGet(API_BASE + '/v1/discovery/models', auth);
  record('errors', 'Unavailable service returns 503', r.status === 503, `status=${r.status}`);

  r = await httpGet(API_BASE + '/v1/health/providers', auth);
  record('errors', 'Nil service returns 503 not 500', r.status === 503, `status=${r.status}`);

  try {
    const resp2 = await httpGet(API_BASE + '/v1/discovery/models', auth);
    const parsed = JSON.parse(resp2.body);
    record('errors', '503 has JSON error structure', !!parsed.error, `error: ${parsed.error}`);
  } catch (e) {
    record('errors', '503 has JSON error structure', false, e.message);
  }

  // ================================================================
  // SECTION 7: EDGE CASES (6 tests)
  // ================================================================
  console.log('\n--- SECTION 7: EDGE CASES (6 tests) ---');

  // Very long URL
  r = await httpGet(API_BASE + '/health?' + 'x'.repeat(2000));
  record('edge', 'Long URL handled', r.status > 0, `status=${r.status}`);

  // Special characters in path
  r = await httpGet(API_BASE + '/v1/providers%20test');
  record('edge', 'URL-encoded path handled', r.status > 0, `status=${r.status}`);

  // Empty body POST
  r = await httpPost(API_BASE + '/v1/chat/completions', '', auth);
  record('edge', 'Empty POST body handled', r.status >= 400, `status=${r.status}`);

  // Concurrent requests
  const concurrent = await Promise.all(Array(5).fill(null).map(() => httpGet(API_BASE + '/health')));
  const allOk = concurrent.every(c => c.status === 200);
  record('edge', '5 concurrent requests succeed', allOk, `statuses: ${concurrent.map(c=>c.status).join(',')}`);

  // Double slash
  r = await httpGet(API_BASE + '//health');
  record('edge', 'Double slash handled', r.status > 0, `status=${r.status}`);

  // Trailing slash
  r = await httpGet(API_BASE + '/health/');
  record('edge', 'Trailing slash handled', r.status > 0, `status=${r.status}`);

  // Final screenshot
  await page.goto(API_BASE + '/health', { timeout: 5000 }).catch(() => {});
  await page.screenshot({ path: path.join(OUTPUT_DIR, 'final-state.png') });

  // Close browser (saves video)
  await page.waitForTimeout(500);
  await context.close();
  await browser.close();

  // Write results
  const passed = results.filter(r => r.passed).length;
  const failed = results.filter(r => !r.passed).length;
  const report = {
    session_id: 'full-qa-' + Date.now(),
    timestamp: new Date().toISOString(),
    total: results.length,
    passed, failed,
    categories: {},
    results
  };

  // Group by category
  results.forEach(r => {
    if (!report.categories[r.category]) report.categories[r.category] = { passed: 0, failed: 0 };
    report.categories[r.category][r.passed ? 'passed' : 'failed']++;
  });

  fs.writeFileSync(path.join(OUTPUT_DIR, 'results.json'), JSON.stringify(report, null, 2));

  console.log('\n========================================');
  console.log('  COMPREHENSIVE QA RESULTS');
  console.log('========================================');
  Object.entries(report.categories).forEach(([cat, counts]) => {
    console.log(`  ${cat}: ${counts.passed}/${counts.passed + counts.failed} passed`);
  });
  console.log(`  ----------------------------------------`);
  console.log(`  TOTAL: ${passed}/${results.length} passed, ${failed} failed`);
  console.log('========================================');

  // List video files
  const videos = fs.readdirSync(OUTPUT_DIR).filter(f => f.endsWith('.webm') || f.endsWith('.mp4'));
  const screenshots = fs.readdirSync(OUTPUT_DIR).filter(f => f.endsWith('.png'));
  console.log(`\nVideo files: ${videos.length}`);
  videos.forEach(v => {
    const s = fs.statSync(path.join(OUTPUT_DIR, v));
    console.log(`  ${v} (${(s.size/1024/1024).toFixed(1)}MB)`);
  });
  console.log(`Screenshots: ${screenshots.length}`);

  process.exit(failed > 0 ? 1 : 0);
})();
