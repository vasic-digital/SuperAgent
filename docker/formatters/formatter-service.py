#!/usr/bin/env python3
"""
Universal HTTP service wrapper for formatter binaries.
Wraps any CLI formatter (autopep8, yapf, sqlfluff, etc.) as an HTTP service.
"""

import argparse
import json
import subprocess
import sys
import tempfile
from http.server import BaseHTTPRequestHandler, HTTPServer
from pathlib import Path
from typing import Dict, Any, Optional


class FormatterServiceHandler(BaseHTTPRequestHandler):
    """HTTP handler for formatter service."""

    formatter_name: str = ""
    formatter_binary: str = ""

    def do_POST(self):
        """Handle POST /format request."""
        if self.path != "/format":
            self.send_error(404, "Not Found")
            return

        # Read request body
        content_length = int(self.headers.get('Content-Length', 0))
        body = self.rfile.read(content_length)

        try:
            request = json.loads(body)
            content = request.get('content', '')
            options = request.get('options', {})

            if not content:
                self.send_json_response(400, {
                    'success': False,
                    'error': 'content field is required'
                })
                return

            # Format the code
            result = self.format_code(content, options)
            self.send_json_response(200, result)

        except json.JSONDecodeError:
            self.send_json_response(400, {
                'success': False,
                'error': 'Invalid JSON'
            })
        except Exception as e:
            self.send_json_response(500, {
                'success': False,
                'error': str(e)
            })

    def do_GET(self):
        """Handle GET /health request."""
        if self.path == "/health":
            try:
                # Check if formatter binary is available
                result = subprocess.run(
                    [self.formatter_binary, "--version"],
                    capture_output=True,
                    text=True,
                    timeout=5
                )

                self.send_json_response(200, {
                    'status': 'healthy',
                    'formatter': self.formatter_name,
                    'version': result.stdout.strip() if result.returncode == 0 else 'unknown'
                })
            except Exception as e:
                self.send_json_response(503, {
                    'status': 'unhealthy',
                    'error': str(e)
                })
        else:
            self.send_error(404, "Not Found")

    def format_code(self, content: str, options: Dict[str, Any]) -> Dict[str, Any]:
        """Format code using the configured formatter."""
        try:
            # Build formatter command
            cmd = [self.formatter_binary]

            # Formatter-specific arguments
            if self.formatter_name == "autopep8":
                cmd.extend(["--aggressive", "--aggressive", "-"])
            elif self.formatter_name == "yapf":
                cmd.append("-")
            elif self.formatter_name == "sqlfluff":
                cmd.extend(["format", "-"])
            elif self.formatter_name == "rubocop":
                cmd.extend(["--auto-correct", "--stdin", "temp.rb"])
            elif self.formatter_name == "php-cs-fixer":
                # php-cs-fixer requires a temp file
                with tempfile.NamedTemporaryFile(mode='w', suffix='.php', delete=False) as f:
                    f.write(content)
                    temp_path = f.name

                result = subprocess.run(
                    [self.formatter_binary, "fix", temp_path],
                    capture_output=True,
                    text=True,
                    timeout=30
                )

                with open(temp_path, 'r') as f:
                    formatted = f.read()

                Path(temp_path).unlink()

                return {
                    'success': True,
                    'content': formatted,
                    'changed': content != formatted,
                    'formatter': self.formatter_name
                }
            elif self.formatter_name == "perltidy":
                cmd.extend(["-st", "-se"])
            elif self.formatter_name == "cljfmt":
                cmd = ["clojure", "-Tcljfmt", "fix"]
            else:
                cmd.append("-")

            # Execute formatter
            result = subprocess.run(
                cmd,
                input=content,
                capture_output=True,
                text=True,
                timeout=30
            )

            if result.returncode != 0 and result.stderr:
                return {
                    'success': False,
                    'error': result.stderr,
                    'formatter': self.formatter_name
                }

            formatted = result.stdout

            return {
                'success': True,
                'content': formatted,
                'changed': content != formatted,
                'formatter': self.formatter_name
            }

        except subprocess.TimeoutExpired:
            return {
                'success': False,
                'error': 'Formatting timeout (30s)',
                'formatter': self.formatter_name
            }
        except Exception as e:
            return {
                'success': False,
                'error': str(e),
                'formatter': self.formatter_name
            }

    def send_json_response(self, status_code: int, data: Dict[str, Any]):
        """Send JSON response."""
        self.send_response(status_code)
        self.send_header('Content-Type', 'application/json')
        self.end_headers()
        self.wfile.write(json.dumps(data).encode('utf-8'))

    def log_message(self, format, *args):
        """Log HTTP requests."""
        sys.stderr.write(f"[{self.formatter_name}] {format % args}\n")


def main():
    parser = argparse.ArgumentParser(description='HTTP service wrapper for code formatters')
    parser.add_argument('--formatter', required=True, help='Formatter name (autopep8, yapf, etc.)')
    parser.add_argument('--port', type=int, required=True, help='HTTP port')
    parser.add_argument('--host', default='0.0.0.0', help='HTTP host (default: 0.0.0.0)')
    parser.add_argument('--binary', help='Formatter binary path (defaults to formatter name)')

    args = parser.parse_args()

    # Set class variables
    FormatterServiceHandler.formatter_name = args.formatter
    FormatterServiceHandler.formatter_binary = args.binary or args.formatter

    # Start HTTP server
    server = HTTPServer((args.host, args.port), FormatterServiceHandler)

    print(f"🚀 {args.formatter} formatter service started on {args.host}:{args.port}")
    print(f"   Health: http://{args.host}:{args.port}/health")
    print(f"   Format: POST http://{args.host}:{args.port}/format")

    try:
        server.serve_forever()
    except KeyboardInterrupt:
        print(f"\n✋ {args.formatter} service stopped")
        server.shutdown()


if __name__ == '__main__':
    main()
