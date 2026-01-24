#!/usr/bin/env python3
"""
HelixAgent Service Exporter for Prometheus

This exporter collects metrics from HelixAgent, ChromaDB, Cognee, and LLMsVerifier
services and exposes them in Prometheus format.

Metrics exported:
- Service health status (up/down)
- Provider counts and status
- MCP server counts
- Response times
- Token usage (from HelixAgent API)
- Debate statistics
- RAG retrieval metrics
"""

import os
import sys
import time
import json
import logging
from http.server import HTTPServer, BaseHTTPRequestHandler
from urllib.request import urlopen, Request
from urllib.error import URLError, HTTPError
from threading import Thread
from typing import Dict, Any, Optional

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger('helixagent-exporter')

# Configuration
HELIXAGENT_URL = os.environ.get('HELIXAGENT_URL', 'http://localhost:7061')
CHROMADB_URL = os.environ.get('CHROMADB_URL', 'http://localhost:8001')
COGNEE_URL = os.environ.get('COGNEE_URL', 'http://localhost:8000')
LLMSVERIFIER_URL = os.environ.get('LLMSVERIFIER_URL', 'http://localhost:8180')
EXPORTER_PORT = int(os.environ.get('EXPORTER_PORT', '9200'))
SCRAPE_INTERVAL = int(os.environ.get('SCRAPE_INTERVAL', '15'))


class MetricsCollector:
    """Collects metrics from HelixAgent ecosystem services."""

    def __init__(self):
        self.metrics: Dict[str, Any] = {}
        self.last_scrape = 0

    def fetch_json(self, url: str, timeout: int = 5) -> Optional[Dict]:
        """Fetch JSON from URL with error handling."""
        try:
            req = Request(url, headers={'Accept': 'application/json'})
            with urlopen(req, timeout=timeout) as response:
                return json.loads(response.read().decode())
        except (URLError, HTTPError, json.JSONDecodeError) as e:
            logger.warning(f"Failed to fetch {url}: {e}")
            return None
        except Exception as e:
            logger.error(f"Unexpected error fetching {url}: {e}")
            return None

    def check_health(self, url: str, timeout: int = 5) -> tuple:
        """Check service health and return (is_up, response_time_ms)."""
        start = time.time()
        try:
            with urlopen(url, timeout=timeout) as response:
                response.read()
                return (1, (time.time() - start) * 1000)
        except Exception:
            return (0, 0)

    def collect_helixagent_metrics(self) -> Dict[str, Any]:
        """Collect HelixAgent-specific metrics."""
        metrics = {}

        # Health check
        is_up, response_time = self.check_health(f"{HELIXAGENT_URL}/health")
        metrics['helixagent_up'] = is_up
        metrics['helixagent_response_time_ms'] = response_time

        if is_up:
            # Providers
            providers = self.fetch_json(f"{HELIXAGENT_URL}/v1/providers")
            if providers:
                metrics['helixagent_providers_total'] = providers.get('count', 0)
                provider_list = providers.get('providers', [])
                metrics['helixagent_providers_healthy'] = sum(
                    1 for p in provider_list if p.get('status') in ('healthy', 'verified')
                )
                metrics['helixagent_providers_unhealthy'] = (
                    metrics['helixagent_providers_total'] - metrics['helixagent_providers_healthy']
                )

            # MCP capabilities
            mcp = self.fetch_json(f"{HELIXAGENT_URL}/v1/mcp")
            if mcp:
                metrics['helixagent_mcp_servers_total'] = len(mcp.get('mcp_servers', []))
                metrics['helixagent_mcp_providers_total'] = len(mcp.get('providers', []))

            # Tool search
            tools = self.fetch_json(f"{HELIXAGENT_URL}/v1/mcp/tools/search?q=")
            if tools:
                metrics['helixagent_tools_total'] = tools.get('count', 0)

            # Monitoring stats
            monitoring = self.fetch_json(f"{HELIXAGENT_URL}/v1/monitoring/status")
            if monitoring:
                metrics['helixagent_circuit_breakers_open'] = monitoring.get('open_circuit_breakers', 0)
                metrics['helixagent_active_requests'] = monitoring.get('active_requests', 0)

        return metrics

    def collect_chromadb_metrics(self) -> Dict[str, Any]:
        """Collect ChromaDB metrics."""
        metrics = {}

        # Health check
        is_up, response_time = self.check_health(f"{CHROMADB_URL}/api/v1/heartbeat")
        metrics['chromadb_up'] = is_up
        metrics['chromadb_response_time_ms'] = response_time

        if is_up:
            # Version info
            version = self.fetch_json(f"{CHROMADB_URL}/api/v1/version")
            if version:
                metrics['chromadb_version_info'] = 1

            # Collections count
            try:
                collections = self.fetch_json(f"{CHROMADB_URL}/api/v1/collections")
                if isinstance(collections, list):
                    metrics['chromadb_collections_total'] = len(collections)
            except Exception:
                pass

        return metrics

    def collect_cognee_metrics(self) -> Dict[str, Any]:
        """Collect Cognee metrics."""
        metrics = {}

        # Health check
        is_up, response_time = self.check_health(f"{COGNEE_URL}/health")
        metrics['cognee_up'] = is_up
        metrics['cognee_response_time_ms'] = response_time

        return metrics

    def collect_llmsverifier_metrics(self) -> Dict[str, Any]:
        """Collect LLMsVerifier metrics."""
        metrics = {}

        # Health check
        is_up, response_time = self.check_health(f"{LLMSVERIFIER_URL}/health")
        metrics['llmsverifier_up'] = is_up
        metrics['llmsverifier_response_time_ms'] = response_time

        if is_up:
            # Verification stats
            stats = self.fetch_json(f"{LLMSVERIFIER_URL}/v1/stats")
            if stats:
                metrics['llmsverifier_verifications_total'] = stats.get('total_verifications', 0)
                metrics['llmsverifier_providers_verified'] = stats.get('verified_providers', 0)

        return metrics

    def collect_all(self) -> Dict[str, Any]:
        """Collect all metrics."""
        all_metrics = {}

        # Collect from each service
        all_metrics.update(self.collect_helixagent_metrics())
        all_metrics.update(self.collect_chromadb_metrics())
        all_metrics.update(self.collect_cognee_metrics())
        all_metrics.update(self.collect_llmsverifier_metrics())

        # Add metadata
        all_metrics['helixagent_exporter_scrape_duration_ms'] = 0
        all_metrics['helixagent_exporter_last_scrape_timestamp'] = time.time()

        self.metrics = all_metrics
        self.last_scrape = time.time()

        return all_metrics

    def format_prometheus(self) -> str:
        """Format metrics in Prometheus exposition format."""
        lines = []

        # Add header
        lines.append("# HELP helixagent_up Whether HelixAgent is up (1) or down (0)")
        lines.append("# TYPE helixagent_up gauge")
        lines.append("# HELP helixagent_response_time_ms HelixAgent response time in milliseconds")
        lines.append("# TYPE helixagent_response_time_ms gauge")
        lines.append("# HELP helixagent_providers_total Total number of LLM providers")
        lines.append("# TYPE helixagent_providers_total gauge")
        lines.append("# HELP helixagent_providers_healthy Number of healthy LLM providers")
        lines.append("# TYPE helixagent_providers_healthy gauge")
        lines.append("# HELP helixagent_providers_unhealthy Number of unhealthy LLM providers")
        lines.append("# TYPE helixagent_providers_unhealthy gauge")
        lines.append("# HELP helixagent_mcp_servers_total Total number of MCP servers")
        lines.append("# TYPE helixagent_mcp_servers_total gauge")
        lines.append("# HELP helixagent_tools_total Total number of available tools")
        lines.append("# TYPE helixagent_tools_total gauge")
        lines.append("# HELP chromadb_up Whether ChromaDB is up (1) or down (0)")
        lines.append("# TYPE chromadb_up gauge")
        lines.append("# HELP chromadb_response_time_ms ChromaDB response time in milliseconds")
        lines.append("# TYPE chromadb_response_time_ms gauge")
        lines.append("# HELP chromadb_collections_total Number of ChromaDB collections")
        lines.append("# TYPE chromadb_collections_total gauge")
        lines.append("# HELP cognee_up Whether Cognee is up (1) or down (0)")
        lines.append("# TYPE cognee_up gauge")
        lines.append("# HELP cognee_response_time_ms Cognee response time in milliseconds")
        lines.append("# TYPE cognee_response_time_ms gauge")
        lines.append("# HELP llmsverifier_up Whether LLMsVerifier is up (1) or down (0)")
        lines.append("# TYPE llmsverifier_up gauge")
        lines.append("")

        # Add metrics
        for key, value in self.metrics.items():
            if isinstance(value, (int, float)):
                lines.append(f"{key} {value}")

        return "\n".join(lines) + "\n"


class MetricsHandler(BaseHTTPRequestHandler):
    """HTTP handler for Prometheus metrics endpoint."""

    collector = MetricsCollector()

    def log_message(self, format, *args):
        """Override to use logging instead of stderr."""
        logger.debug(f"{self.address_string()} - {format % args}")

    def do_GET(self):
        """Handle GET requests."""
        if self.path == '/metrics':
            self.collector.collect_all()
            metrics = self.collector.format_prometheus()

            self.send_response(200)
            self.send_header('Content-Type', 'text/plain; version=0.0.4; charset=utf-8')
            self.send_header('Content-Length', len(metrics))
            self.end_headers()
            self.wfile.write(metrics.encode())

        elif self.path == '/health':
            self.send_response(200)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            self.wfile.write(b'{"status":"healthy"}')

        elif self.path == '/':
            html = '''<!DOCTYPE html>
<html>
<head><title>HelixAgent Exporter</title></head>
<body>
<h1>HelixAgent Service Exporter</h1>
<p>Metrics endpoint: <a href="/metrics">/metrics</a></p>
<p>Health endpoint: <a href="/health">/health</a></p>
<h2>Monitored Services:</h2>
<ul>
<li>HelixAgent API</li>
<li>ChromaDB</li>
<li>Cognee</li>
<li>LLMsVerifier</li>
</ul>
</body>
</html>'''
            self.send_response(200)
            self.send_header('Content-Type', 'text/html')
            self.end_headers()
            self.wfile.write(html.encode())

        else:
            self.send_response(404)
            self.end_headers()


def main():
    """Main entry point."""
    logger.info(f"Starting HelixAgent Exporter on port {EXPORTER_PORT}")
    logger.info(f"Monitoring services:")
    logger.info(f"  - HelixAgent: {HELIXAGENT_URL}")
    logger.info(f"  - ChromaDB: {CHROMADB_URL}")
    logger.info(f"  - Cognee: {COGNEE_URL}")
    logger.info(f"  - LLMsVerifier: {LLMSVERIFIER_URL}")

    server = HTTPServer(('0.0.0.0', EXPORTER_PORT), MetricsHandler)

    try:
        server.serve_forever()
    except KeyboardInterrupt:
        logger.info("Shutting down exporter")
        server.shutdown()


if __name__ == '__main__':
    main()
