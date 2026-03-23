const { chromium } = require('playwright');
const path = require('path');
const fs = require('fs');

const OUTPUT_DIR = process.env.OUTPUT_DIR || 'qa-results/video-sessions/session-' + Date.now();
fs.mkdirSync(OUTPUT_DIR, { recursive: true });

(async () => {
  console.log('Starting HelixQA Video-Recorded Browser Session');
  console.log('Output: ' + OUTPUT_DIR);
  
  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext({
    recordVideo: {
      dir: OUTPUT_DIR,
      size: { width: 1280, height: 720 }
    }
  });
  
  const page = await context.newPage();
  const results = [];
  
  const record = (name, passed, detail) => {
    results.push({ name, passed, detail, timestamp: new Date().toISOString() });
    console.log(`  ${passed ? 'PASS' : 'FAIL'}: ${name}${detail ? ' - ' + detail : ''}`);
  };

  // ============ TEST 1: Website Homepage ============
  console.log('\n--- Test 1: Website Homepage ---');
  try {
    await page.goto('http://localhost:8090/', { timeout: 10000 });
    await page.waitForTimeout(1000);
    const title = await page.title();
    record('Website loads', true, 'Title: ' + title);
    await page.screenshot({ path: path.join(OUTPUT_DIR, 'homepage.png') });
  } catch (e) {
    record('Website loads', false, e.message);
  }

  // ============ TEST 2: Features Page ============
  console.log('\n--- Test 2: Features Page ---');
  try {
    await page.goto('http://localhost:8090/features.html', { timeout: 10000 });
    await page.waitForTimeout(1000);
    const content = await page.textContent('body');
    record('Features page loads', content.length > 100, 'Content length: ' + content.length);
    await page.screenshot({ path: path.join(OUTPUT_DIR, 'features.png') });
  } catch (e) {
    record('Features page loads', false, e.message);
  }

  // ============ TEST 3: API Health Check ============
  console.log('\n--- Test 3: API Health via Browser ---');
  try {
    await page.goto('http://localhost:7061/health', { timeout: 10000 });
    await page.waitForTimeout(500);
    const body = await page.textContent('body');
    const healthy = body.includes('healthy');
    record('API health endpoint', healthy, body.substring(0, 100));
    await page.screenshot({ path: path.join(OUTPUT_DIR, 'api-health.png') });
  } catch (e) {
    record('API health endpoint', false, e.message);
  }

  // ============ TEST 4: API Providers ============
  console.log('\n--- Test 4: API Providers via Browser ---');
  try {
    await page.goto('http://localhost:7061/v1/providers', { timeout: 15000 });
    await page.waitForTimeout(500);
    const body = await page.textContent('body');
    const hasProviders = body.includes('providers') && body.includes('count');
    record('API providers endpoint', hasProviders, 'Has provider data: ' + hasProviders);
    await page.screenshot({ path: path.join(OUTPUT_DIR, 'api-providers.png') });
  } catch (e) {
    record('API providers endpoint', false, e.message);
  }

  // ============ TEST 5: Website Navigation ============
  console.log('\n--- Test 5: Website Navigation ---');
  try {
    await page.goto('http://localhost:8090/', { timeout: 10000 });
    await page.waitForTimeout(500);
    
    // Check all pages load
    const pages = ['features.html', 'pricing.html', 'changelog.html', 'contact.html'];
    let allLoaded = true;
    for (const p of pages) {
      await page.goto('http://localhost:8090/' + p, { timeout: 5000 });
      await page.waitForTimeout(300);
      const ok = (await page.textContent('body')).length > 50;
      if (!ok) allLoaded = false;
    }
    record('All website pages load', allLoaded, pages.length + ' pages tested');
    await page.screenshot({ path: path.join(OUTPUT_DIR, 'navigation.png') });
  } catch (e) {
    record('All website pages load', false, e.message);
  }

  // ============ TEST 6: API Docs Page ============
  console.log('\n--- Test 6: API Documentation Page ---');
  try {
    await page.goto('http://localhost:8090/docs/api.html', { timeout: 10000 });
    await page.waitForTimeout(1000);
    const body = await page.textContent('body');
    const hasEndpoints = body.includes('/v1/') || body.includes('endpoint');
    record('API docs page', hasEndpoints, 'Contains endpoint docs: ' + hasEndpoints);
    await page.screenshot({ path: path.join(OUTPUT_DIR, 'api-docs.png') });
  } catch (e) {
    record('API docs page', false, e.message);
  }

  // ============ TEST 7: Responsive Check ============
  console.log('\n--- Test 7: Responsive Design ---');
  try {
    await page.setViewportSize({ width: 375, height: 812 }); // iPhone
    await page.goto('http://localhost:8090/', { timeout: 10000 });
    await page.waitForTimeout(500);
    await page.screenshot({ path: path.join(OUTPUT_DIR, 'mobile-view.png') });
    
    await page.setViewportSize({ width: 1920, height: 1080 }); // Desktop
    await page.goto('http://localhost:8090/', { timeout: 10000 });
    await page.waitForTimeout(500);
    await page.screenshot({ path: path.join(OUTPUT_DIR, 'desktop-view.png') });
    record('Responsive design check', true, 'Mobile + Desktop screenshots captured');
  } catch (e) {
    record('Responsive design check', false, e.message);
  }

  // ============ TEST 8: Error Pages ============
  console.log('\n--- Test 8: Error Handling ---');
  try {
    const resp = await page.goto('http://localhost:7061/nonexistent', { timeout: 5000 });
    record('404 error handling', resp.status() === 404, 'Status: ' + resp.status());
    await page.screenshot({ path: path.join(OUTPUT_DIR, 'error-404.png') });
  } catch (e) {
    record('404 error handling', false, e.message);
  }

  // Close and save video
  await page.waitForTimeout(1000);
  await context.close();
  await browser.close();

  // Write results
  const report = {
    session_id: 'video-qa-' + Date.now(),
    timestamp: new Date().toISOString(),
    total: results.length,
    passed: results.filter(r => r.passed).length,
    failed: results.filter(r => !r.passed).length,
    results: results
  };
  
  fs.writeFileSync(path.join(OUTPUT_DIR, 'results.json'), JSON.stringify(report, null, 2));
  
  console.log('\n========================================');
  console.log(`  Results: ${report.passed}/${report.total} passed`);
  console.log('========================================');
  
  // List output files
  const files = fs.readdirSync(OUTPUT_DIR);
  console.log('\nOutput files:');
  files.forEach(f => {
    const stat = fs.statSync(path.join(OUTPUT_DIR, f));
    console.log(`  ${f} (${(stat.size / 1024).toFixed(1)}KB)`);
  });

  process.exit(report.failed > 0 ? 1 : 0);
})();
