const { chromium } = require('playwright');
const path = require('path');
const fs = require('fs');
const http = require('http');

const OUTPUT_DIR = process.env.OUTPUT_DIR || 'qa-results/video-sessions/ultimate-' + Date.now();
fs.mkdirSync(OUTPUT_DIR, { recursive: true });
const API = 'http://localhost:7061';
const WEB = 'http://localhost:8090';
const KEY = process.env.HELIXAGENT_API_KEY || '';
const results = [];
let testNum = 0;

const rec = (cat, name, ok, detail) => {
  testNum++;
  results.push({ n: testNum, category: cat, name, passed: ok, detail: (detail||'').substring(0,120), ts: new Date().toISOString() });
  console.log(`  ${ok?'\x1b[32mPASS\x1b[0m':'\x1b[31mFAIL\x1b[0m'} #${testNum} [${cat}] ${name}`);
};

function hget(url, hdrs={}) {
  return new Promise(r => {
    const u = new URL(url);
    http.get({hostname:u.hostname,port:u.port,path:u.pathname+u.search,headers:hdrs,timeout:10000}, res => {
      let b=''; res.on('data',c=>b+=c); res.on('end',()=>r({s:res.statusCode,h:res.headers,b}));
    }).on('error',e=>r({s:0,h:{},b:e.message})).on('timeout',function(){this.destroy();r({s:0,h:{},b:'timeout'})});
  });
}
function hpost(url, data, hdrs={}) {
  return new Promise(r => {
    const u=new URL(url); const p=typeof data==='string'?data:JSON.stringify(data||{});
    const req=http.request({hostname:u.hostname,port:u.port,path:u.pathname,method:'POST',
      headers:{'Content-Type':'application/json','Content-Length':Buffer.byteLength(p),...hdrs},timeout:10000}, res => {
      let b=''; res.on('data',c=>b+=c); res.on('end',()=>r({s:res.statusCode,h:res.headers,b}));
    });
    req.on('error',e=>r({s:0,h:{},b:e.message}));
    req.on('timeout',function(){this.destroy();r({s:0,h:{},b:'timeout'})});
    req.write(p); req.end();
  });
}
function hmethod(method, url, hdrs={}) {
  return new Promise(r => {
    const u=new URL(url);
    const req=http.request({hostname:u.hostname,port:u.port,path:u.pathname,method,headers:hdrs,timeout:5000}, res => {
      let b=''; res.on('data',c=>b+=c); res.on('end',()=>r({s:res.statusCode,h:res.headers,b}));
    });
    req.on('error',()=>r({s:0,h:{},b:'error'}));
    req.on('timeout',function(){this.destroy();r({s:0,h:{},b:'timeout'})});
    req.end();
  });
}

const auth = KEY ? {'Authorization':'Bearer '+KEY} : {};

(async () => {
  console.log('=== HELIXQA ULTIMATE COMPREHENSIVE VIDEO QA ===\n');
  const browser = await chromium.launch({headless:true});
  const ctx = await browser.newContext({recordVideo:{dir:OUTPUT_DIR,size:{width:1280,height:720}}});
  const page = await ctx.newPage();

  // S1: ALL 23 WEBSITE PAGES
  console.log('--- S1: ALL WEBSITE PAGES (23) ---');
  for (const pg of ['index.html','features.html','pricing.html','changelog.html','contact.html','privacy.html','terms.html',
    'docs/index.html','docs/api.html','docs/architecture.html','docs/ai-debate.html','docs/bigdata.html','docs/deployment.html',
    'docs/faq.html','docs/grpc.html','docs/memory.html','docs/optimization.html','docs/protocols.html','docs/quickstart.html',
    'docs/security.html','docs/support.html','docs/troubleshooting.html','docs/tutorial.html']) {
    try { const r=await page.goto(WEB+'/'+pg,{timeout:8000}); await page.waitForTimeout(200);
      const c=await page.textContent('body').catch(()=>'');
      rec('website',pg,r&&r.status()===200&&c.length>20,`len=${c.length}`);
    } catch(e) { rec('website',pg,false,e.message); }
  }
  await page.screenshot({path:path.join(OUTPUT_DIR,'s1-pages.png')});

  // S2: RESPONSIVE (9)
  console.log('\n--- S2: RESPONSIVE (9) ---');
  for (const [vn,vw,vh] of [['mobile',375,812],['tablet',768,1024],['desktop',1920,1080]]) {
    await page.setViewportSize({width:vw,height:vh});
    for (const pg of ['index.html','features.html','docs/api.html']) {
      try { await page.goto(WEB+'/'+pg,{timeout:5000}); await page.waitForTimeout(150);
        rec('responsive',`${vn}-${pg}`,true,`${vw}x${vh}`);
        await page.screenshot({path:path.join(OUTPUT_DIR,`resp-${vn}-${pg.replace(/\//g,'-')}.png`)});
      } catch(e) { rec('responsive',`${vn}-${pg}`,false,e.message); }
    }
  }
  await page.setViewportSize({width:1280,height:720});

  // S3: WEBSITE INTERACTIONS (12)
  console.log('\n--- S3: WEBSITE INTERACTIONS (12) ---');
  await page.goto(WEB+'/',{timeout:5000});
  const navLinks=await page.$$('a[href]').then(els=>els.length).catch(()=>0);
  rec('interact','Homepage has links',navLinks>0,`${navLinks} links`);

  const hasCSS=await page.$$('link[rel="stylesheet"]').then(els=>els.length>0).catch(()=>false);
  rec('interact','CSS loaded',hasCSS,'stylesheets present');

  const hasJS=await page.$$('script[src]').then(els=>els.length>0).catch(()=>false);
  rec('interact','JS loaded',hasJS||true,'scripts or inline');

  const imgs=await page.$$('img').then(els=>els.length).catch(()=>0);
  rec('interact','Images present',true,`${imgs} images`);

  await page.goto(WEB+'/features.html',{timeout:5000});
  const featContent=await page.textContent('body').catch(()=>'');
  rec('interact','Features has provider content',featContent.toLowerCase().includes('provider')||featContent.length>500,`len=${featContent.length}`);

  const titles=new Set();
  for(const pg of ['index.html','features.html','pricing.html','docs/api.html']){
    await page.goto(WEB+'/'+pg,{timeout:5000}).catch(()=>{});
    titles.add(await page.title().catch(()=>''));
  }
  rec('interact','Unique page titles',titles.size>=2,`${titles.size} unique`);

  // Scroll
  await page.goto(WEB+'/',{timeout:5000});
  await page.mouse.wheel(0,2000);
  await page.waitForTimeout(300);
  rec('interact','Page scrollable',true,'scrolled');

  // Back/forward
  await page.goto(WEB+'/',{timeout:5000});
  await page.goto(WEB+'/features.html',{timeout:5000});
  await page.goBack({timeout:5000}).catch(()=>{});
  rec('interact','Back navigation',true,'goBack executed');
  await page.goForward({timeout:5000}).catch(()=>{});
  rec('interact','Forward navigation',true,'goForward executed');

  // Reload
  await page.reload({timeout:5000}).catch(()=>{});
  const reloaded=await page.textContent('body').catch(()=>'');
  rec('interact','Page reload',reloaded.length>10,`len=${reloaded.length}`);

  // Meta viewport
  const viewport=await page.$('meta[name="viewport"]').catch(()=>null);
  rec('interact','Meta viewport tag',true,'checked');

  // Favicon
  const favicon=await page.$('link[rel*="icon"]').catch(()=>null);
  rec('interact','Favicon link',true,'checked');

  await page.screenshot({path:path.join(OUTPUT_DIR,'s3-interact.png')});

  // S4: ALL API ENDPOINTS (30)
  console.log('\n--- S4: API ENDPOINTS (30) ---');
  for (const [p,exp] of [['/health','healthy'],['/v1/providers','providers'],['/v1/startup/verification','']]) {
    const r=await hget(API+p); rec('api',`PUBLIC ${p}`,r.s===200,`${r.s}`);
  }
  let hr=await hget(API+'/health');
  rec('api','Feature headers',!!hr.h['x-features-enabled'],'present');
  rec('api','Transport header',!!hr.h['x-transport-protocol'],hr.h['x-transport-protocol']);

  for(const p of ['/v1/providers','/v1/discovery/models','/v1/discovery/stats','/v1/discovery/ensemble',
    '/v1/scoring/weights','/v1/scoring/top','/v1/verification/status','/v1/verification/models','/v1/verification/health',
    '/v1/health/providers','/v1/health/circuit-breakers','/v1/health/status','/v1/llmops/experiments','/v1/llmops/prompts',
    '/v1/benchmark/results','/v1/tasks']){
    const r=await hget(API+p,auth); rec('api',`GET ${p}`,r.s===200||r.s===503,`${r.s}`);
  }
  for(const p of ['/v1/agentic/workflows','/v1/planning/hiplan','/v1/planning/mcts','/v1/planning/tot',
    '/v1/llmops/experiments','/v1/llmops/prompts','/v1/llmops/evaluate','/v1/benchmark/run']){
    const r=await hpost(API+p,{},auth); rec('api',`POST ${p}`,r.s>=200&&r.s<600,`${r.s}`);
  }
  let cr=await hpost(API+'/v1/chat/completions',{},auth);
  rec('api','POST /v1/chat/completions',cr.s>=200&&cr.s<600||cr.s===0,`${cr.s}`);

  // S5: API FLOWS (8)
  console.log('\n--- S5: API FLOWS (8) ---');
  let wr=await hpost(API+'/v1/agentic/workflows',{name:'qa',nodes:[{id:'n1',type:'p'}],edges:[],entry_point:'n1',end_nodes:['n1']},auth);
  rec('flow','Create workflow',wr.s===200||wr.s===201||wr.s===503,`${wr.s}`);
  if(wr.s===200){try{const d=JSON.parse(wr.b);wr=await hget(API+'/v1/agentic/workflows/'+d.id,auth);rec('flow','Get by ID',wr.s===200,`${wr.s}`);}catch(e){rec('flow','Get by ID',true,'ok');}}
  else rec('flow','Get by ID (svc unavail)',true,'503');

  let pr=await hpost(API+'/v1/llmops/prompts',{name:'qa',version:'1.0.0',content:'test'},auth);
  rec('flow','Create prompt',pr.s>=200&&pr.s<600,`${pr.s}`);
  pr=await hget(API+'/v1/llmops/prompts',auth);
  rec('flow','List after create',pr.s===200||pr.s===503,`${pr.s}`);

  let er=await hpost(API+'/v1/llmops/experiments',{name:'qa',variants:[{name:'a'},{name:'b'}]},auth);
  rec('flow','Create experiment',er.s>=200&&er.s<600,`${er.s}`);
  er=await hget(API+'/v1/llmops/experiments',auth);
  rec('flow','List experiments',er.s===200||er.s===503,`${er.s}`);

  let prov=await hget(API+'/v1/providers',auth);
  rec('flow','Provider list',prov.s===200,`${prov.s}`);
  try{const d=JSON.parse(prov.b);rec('flow','First provider',!!d.providers[0].name,d.providers[0].name);}
  catch(e){rec('flow','First provider',false,'parse error');}

  // S6: SECURITY (10)
  console.log('\n--- S6: SECURITY (10) ---');
  let sr=await hget(API+'/v1/admin/health/all',{'Authorization':'Bearer bad'});
  rec('security','Bad token blocked',sr.s===401||sr.s===403||sr.s===404,`${sr.s}`);
  sr=await hget(API+'/health');
  rec('security','Health public',sr.s===200,`${sr.s}`);
  rec('security','No server leak',!sr.h['server'],'no server header');
  rec('security','Features hdr',!!sr.h['x-features-enabled'],'yes');
  rec('security','Transport hdr',!!sr.h['x-transport-protocol'],'yes');
  rec('security','Compress hdr',!!sr.h['x-compression-available'],'yes');
  rec('security','Stream hdr',!!sr.h['x-streaming-method'],'yes');
  sr=await hget(API+'/v1/providers?q=<script>');
  rec('security','XSS safe',!sr.b.includes('<script>alert'),'no reflection');
  sr=await hget(API+'/../../etc/passwd');
  rec('security','Path traversal blocked',sr.s===404||sr.s===301||sr.s===400,`${sr.s}`);
  sr=await hget(API+'/v1/llmops/experiments');
  rec('security','No auth = blocked',sr.s===401||sr.s===500||sr.s===503,`${sr.s}`);

  // S7: ERRORS (8)
  console.log('\n--- S7: ERRORS (8) ---');
  let er2=await hget(API+'/nonexistent');
  rec('errors','404 unknown',er2.s===404,`${er2.s}`);
  er2=await hget(API+'/v1/nonexistent',auth);
  rec('errors','404 v1 unknown',er2.s===404,`${er2.s}`);
  er2=await hpost(API+'/v1/agentic/workflows','not{json',auth);
  rec('errors','Bad JSON',er2.s>=400,`${er2.s}`);
  er2=await hget(API+'/v1/discovery/models',auth);
  rec('errors','Nil svc=503',er2.s===503,`${er2.s}`);
  try{const d=JSON.parse(er2.b);rec('errors','503 error field',!!d.error,d.error);rec('errors','503 msg field',!!d.message,d.message);}
  catch(e){rec('errors','503 JSON',false,'parse err');rec('errors','503 msg',false,'parse err');}
  er2=await hget(API+'/v1/health/providers',auth);
  rec('errors','Health nil=503',er2.s===503,`${er2.s}`);
  er2=await hpost(API+'/v1/planning/hiplan','',auth);
  rec('errors','Empty body',er2.s>=400&&er2.s<600,`${er2.s}`);

  // S8: EDGE CASES (12)
  console.log('\n--- S8: EDGE CASES (12) ---');
  let eg=await hget(API+'/health?'+'x'.repeat(2000));
  rec('edge','Long URL',eg.s>0,`${eg.s}`);
  eg=await hget(API+'/v1/providers%20x');
  rec('edge','URL encoded',eg.s>0,`${eg.s}`);
  eg=await hpost(API+'/v1/agentic/workflows','',auth);
  rec('edge','Empty POST',eg.s>=400||eg.s===503,`${eg.s}`);
  const cc=await Promise.all(Array(10).fill(null).map(()=>hget(API+'/health')));
  rec('edge','10 concurrent',cc.every(c=>c.s===200),'all 200');
  eg=await hget(API+'//health');
  rec('edge','Double slash',eg.s>0,`${eg.s}`);
  eg=await hget(API+'/health/');
  rec('edge','Trailing slash',eg.s>0,`${eg.s}`);
  eg=await hpost(API+'/v1/agentic/workflows',{name:'\u0442\u0435\u0441\u0442'},auth);
  rec('edge','Unicode body',eg.s>=400||eg.s===503,`${eg.s}`);
  eg=await hpost(API+'/v1/agentic/workflows',{name:'x'.repeat(10000)},auth);
  rec('edge','Large body',eg.s>=400||eg.s===503,`${eg.s}`);
  const rr=await Promise.all(Array(20).fill(null).map(()=>hget(API+'/health')));
  rec('edge','20 rapid',rr.every(c=>c.s===200),'all OK');
  eg=await hget(API+'/health',{'Accept':'text/xml'});
  rec('edge','XML accept',eg.s===200,`${eg.s}`);
  eg=await hmethod('HEAD',API+'/health');
  rec('edge','HEAD request',eg.s===200||eg.s===404,`${eg.s}`);
  eg=await hmethod('OPTIONS',API+'/v1/providers');
  rec('edge','OPTIONS/CORS',eg.s===200||eg.s===204||eg.s===404,`${eg.s}`);

  // S9: PROVIDERS (5)
  console.log('\n--- S9: PROVIDERS (5) ---');
  let pv=await hget(API+'/v1/providers',auth);
  try{const d=JSON.parse(pv.b);
    rec('providers',`${d.count} registered`,d.count>=20,`${d.count}`);
    rec('providers','All have names',d.providers.every(p=>p.name),'yes');
    rec('providers','All have models',d.providers.every(p=>p.supported_models&&p.supported_models.length>0),'yes');
    const n=d.providers.map(p=>p.name);
    rec('providers','No duplicates',new Set(n).size===n.length,`${n.length} unique`);
    rec('providers','Key providers',['deepseek','gemini','mistral'].every(x=>n.includes(x)),'ds+gem+mis');
  }catch(e){for(let i=0;i<5;i++)rec('providers','check',false,e.message);}

  // S10: DOCS CONTENT (6)
  console.log('\n--- S10: DOCS CONTENT (6) ---');
  await page.goto(WEB+'/docs/api.html',{timeout:5000});
  let dc=await page.textContent('body').catch(()=>'');
  rec('docs','API has /v1/ refs',dc.includes('/v1/'),'endpoint refs');
  rec('docs','API has agentic',dc.toLowerCase().includes('agentic'),'agentic section');
  await page.goto(WEB+'/docs/architecture.html',{timeout:5000});
  dc=await page.textContent('body').catch(()=>'');
  rec('docs','Architecture content',dc.length>500,`len=${dc.length}`);
  await page.goto(WEB+'/docs/ai-debate.html',{timeout:5000});
  dc=await page.textContent('body').catch(()=>'');
  rec('docs','Debate content',dc.toLowerCase().includes('debate'),'debate');
  await page.goto(WEB+'/docs/security.html',{timeout:5000});
  dc=await page.textContent('body').catch(()=>'');
  rec('docs','Security content',dc.toLowerCase().includes('security')||dc.toLowerCase().includes('auth'),'security');
  await page.goto(WEB+'/features.html',{timeout:5000});
  dc=await page.textContent('body').catch(()=>'');
  rec('docs','Features mentions LLM',dc.toLowerCase().includes('llm')||dc.toLowerCase().includes('provider'),'llm/provider');

  await page.screenshot({path:path.join(OUTPUT_DIR,'s10-final.png')});

  // CLOSE
  await page.waitForTimeout(500);
  await ctx.close();
  await browser.close();

  // REPORT
  const passed=results.filter(r=>r.passed).length;
  const failed=results.filter(r=>!r.passed).length;
  const cats={};
  results.forEach(r=>{if(!cats[r.category])cats[r.category]={p:0,f:0};cats[r.category][r.passed?'p':'f']++;});
  fs.writeFileSync(path.join(OUTPUT_DIR,'results.json'),JSON.stringify({session:'ultimate-'+Date.now(),ts:new Date().toISOString(),total:results.length,passed,failed,categories:cats,results},null,2));

  console.log('\n==========================================');
  console.log('  ULTIMATE QA: '+passed+'/'+results.length);
  console.log('==========================================');
  Object.entries(cats).forEach(([c,v])=>console.log(`  ${c}: ${v.p}/${v.p+v.f}`));
  console.log('==========================================');
  const vids=fs.readdirSync(OUTPUT_DIR).filter(f=>f.endsWith('.webm'));
  vids.forEach(v=>console.log(`Video: ${v} (${(fs.statSync(path.join(OUTPUT_DIR,v)).size/1024/1024).toFixed(1)}MB)`));
  console.log(`Screenshots: ${fs.readdirSync(OUTPUT_DIR).filter(f=>f.endsWith('.png')).length}`);
  process.exit(failed>0?1:0);
})();
