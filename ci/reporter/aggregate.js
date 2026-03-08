#!/usr/bin/env node
"use strict";

const fs = require("fs");
const path = require("path");
const { execFileSync } = require("child_process");
const { XMLParser } = require("fast-xml-parser");
const { globSync } = require("glob");

const WORKSPACE = process.env.WORKSPACE || "/workspace";
const REPORTS_DIR = path.join(WORKSPACE, "reports");
const OUTPUT_JSON = path.join(REPORTS_DIR, "results.json");
const OUTPUT_HTML = path.join(REPORTS_DIR, "summary.html");
const TEMPLATE = path.join(__dirname, "dashboard-template.html");

const parser = new XMLParser({
  ignoreAttributes: false,
  attributeNamePrefix: "@_",
});

function parseJunitXml(filePath) {
  try {
    const xml = fs.readFileSync(filePath, "utf8");
    const parsed = parser.parse(xml);
    const suites = parsed.testsuites || parsed.testsuite || {};
    const attrs = suites["@_tests"]
      ? suites
      : (Array.isArray(suites.testsuite)
          ? suites.testsuite[0]
          : suites.testsuite) || {};

    return {
      tests: parseInt(attrs["@_tests"] || "0", 10),
      failures: parseInt(attrs["@_failures"] || "0", 10),
      errors: parseInt(attrs["@_errors"] || "0", 10),
      skipped: parseInt(attrs["@_skipped"] || "0", 10),
      time: parseFloat(attrs["@_time"] || "0"),
    };
  } catch {
    return { tests: 0, failures: 0, errors: 0, skipped: 0, time: 0 };
  }
}

function parseCoverageFile(filePath) {
  try {
    const content = fs.readFileSync(filePath, "utf8");
    const match = content.match(/([\d.]+)%/g);
    if (match && match.length > 0) {
      return parseFloat(match[match.length - 1]);
    }
    return 0;
  } catch {
    return 0;
  }
}

function collectPhaseResults(phase) {
  const dir = path.join(REPORTS_DIR, phase);
  if (!fs.existsSync(dir)) return null;

  const xmlFiles = globSync("**/*.xml", { cwd: dir });
  let totalTests = 0;
  let totalFail = 0;
  let totalErr = 0;
  let totalSkip = 0;
  let totalTime = 0;

  const testSuites = {};
  for (const f of xmlFiles) {
    const result = parseJunitXml(path.join(dir, f));
    testSuites[f] = result;
    totalTests += result.tests;
    totalFail += result.failures;
    totalErr += result.errors;
    totalSkip += result.skipped;
    totalTime += result.time;
  }

  // Coverage
  let coverage = 0;
  const coverageFiles = globSync("**/*coverage*summary*.txt", { cwd: dir });
  if (coverageFiles.length > 0) {
    coverage = parseCoverageFile(path.join(dir, coverageFiles[0]));
  }

  // Artifacts
  const artifacts = [];
  const relDir = path.join(WORKSPACE, "releases");
  if (fs.existsSync(relDir)) {
    const pattern =
      phase === "go" ? "**/build-info.json" : `${phase}/**/build-info.json`;
    const infoFiles = globSync(pattern, { cwd: relDir });
    for (const f of infoFiles) {
      try {
        artifacts.push(
          JSON.parse(fs.readFileSync(path.join(relDir, f), "utf8"))
        );
      } catch {
        /* skip invalid */
      }
    }
  }

  return {
    status: totalFail === 0 && totalErr === 0 ? "pass" : "fail",
    duration_s: Math.round(totalTime),
    tests: {
      total: totalTests,
      passed: totalTests - totalFail - totalErr - totalSkip,
      failed: totalFail + totalErr,
      skipped: totalSkip,
    },
    coverage: { percent: coverage },
    test_suites: testSuites,
    artifacts: artifacts,
  };
}

function collectFalsePositiveChecks() {
  const checks = [];
  const fpFiles = globSync("**/false-positive-checks.json", {
    cwd: REPORTS_DIR,
  });
  for (const f of fpFiles) {
    try {
      const data = JSON.parse(
        fs.readFileSync(path.join(REPORTS_DIR, f), "utf8")
      );
      if (data.checks) checks.push(...data.checks);
    } catch {
      /* skip invalid */
    }
  }
  return checks;
}

function getGitInfo(field) {
  try {
    return execFileSync("git", [field === "commit" ? "rev-parse" : "rev-parse", ...(field === "branch" ? ["--abbrev-ref"] : []), "HEAD"], {
      cwd: WORKSPACE,
      encoding: "utf8",
    }).trim();
  } catch {
    return "unknown";
  }
}

// Collect all results
const results = {
  timestamp: new Date().toISOString(),
  git: {
    commit: getGitInfo("commit"),
    branch: getGitInfo("branch"),
    dirty: false,
  },
  resource_limit: process.env.CI_RESOURCE_LIMIT || "low",
  phases: {
    go: collectPhaseResults("go"),
    mobile: collectPhaseResults("mobile"),
    web: collectPhaseResults("web"),
  },
  totals: {
    tests_total: 0,
    tests_passed: 0,
    tests_failed: 0,
    coverage_avg: 0,
  },
  false_positive_checks: collectFalsePositiveChecks(),
  signing: {},
  lighthouse: {},
};

// Calculate totals
let coverageSum = 0;
let coverageCount = 0;
for (const phase of Object.values(results.phases)) {
  if (!phase) continue;
  results.totals.tests_total += phase.tests.total;
  results.totals.tests_passed += phase.tests.passed;
  results.totals.tests_failed += phase.tests.failed;
  if (phase.coverage.percent > 0) {
    coverageSum += phase.coverage.percent;
    coverageCount++;
  }
}
results.totals.coverage_avg =
  coverageCount > 0
    ? Math.round((coverageSum / coverageCount) * 100) / 100
    : 0;

// Signing status
try {
  const sigFile = path.join(REPORTS_DIR, "mobile", "signing-verification.json");
  if (fs.existsSync(sigFile)) {
    results.signing = JSON.parse(fs.readFileSync(sigFile, "utf8"));
  }
} catch {
  /* no signing data */
}

// Write JSON
fs.mkdirSync(path.dirname(OUTPUT_JSON), { recursive: true });
fs.writeFileSync(OUTPUT_JSON, JSON.stringify(results, null, 2));
console.log("Results written to: " + OUTPUT_JSON);

// Generate HTML dashboard
let html = fs.readFileSync(TEMPLATE, "utf8");
html = html.replace("{{RESULTS_JSON}}", JSON.stringify(results));
html = html.replace("{{TIMESTAMP}}", results.timestamp);
html = html.replace(
  "{{GIT_COMMIT}}",
  results.git.commit.substring(0, 8)
);
html = html.replace("{{GIT_BRANCH}}", results.git.branch);
html = html.replace("{{RESOURCE_LIMIT}}", results.resource_limit);
html = html.replace("{{TOTAL_TESTS}}", String(results.totals.tests_total));
html = html.replace("{{TOTAL_PASSED}}", String(results.totals.tests_passed));
html = html.replace("{{TOTAL_FAILED}}", String(results.totals.tests_failed));
html = html.replace("{{COVERAGE_AVG}}", String(results.totals.coverage_avg));
fs.writeFileSync(OUTPUT_HTML, html);
console.log("Dashboard written to: " + OUTPUT_HTML);

// Exit with error if any failures
if (results.totals.tests_failed > 0) {
  console.error("FAILED: " + results.totals.tests_failed + " test(s) failed");
  process.exit(1);
}

console.log("All phases passed.");
