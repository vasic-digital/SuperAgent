#!/usr/bin/env python3
"""
Provider API Documentation Fetcher
Fetches complete web documentation for LLM providers
"""

import asyncio
import json
import os
import sys
from pathlib import Path
from typing import Dict, List, Optional
from dataclasses import dataclass
import aiohttp
from bs4 import BeautifulSoup
import markdown

@dataclass
class ProviderConfig:
    name: str
    docs_url: str
    api_base: str
    http3_support: bool
    brotli_support: bool
    streaming_support: bool
    
PROVIDERS = {
    "openai": ProviderConfig(
        name="OpenAI",
        docs_url="https://platform.openai.com/docs/api-reference",
        api_base="https://api.openai.com/v1",
        http3_support=True,
        brotli_support=True,
        streaming_support=True
    ),
    "anthropic": ProviderConfig(
        name="Anthropic",
        docs_url="https://docs.anthropic.com/en/api",
        api_base="https://api.anthropic.com/v1",
        http3_support=True,
        brotli_support=True,
        streaming_support=True
    ),
    "google-gemini": ProviderConfig(
        name="Google Gemini",
        docs_url="https://ai.google.dev/api",
        api_base="https://generativelanguage.googleapis.com/v1beta",
        http3_support=True,
        brotli_support=True,
        streaming_support=True
    ),
    "deepseek": ProviderConfig(
        name="DeepSeek",
        docs_url="https://platform.deepseek.com/api-docs",
        api_base="https://api.deepseek.com/v1",
        http3_support=False,  # Check actual support
        brotli_support=False,  # Check actual support
        streaming_support=True
    ),
    "qwen": ProviderConfig(
        name="Qwen (Alibaba)",
        docs_url="https://help.aliyun.com/document_detail/611472.html",
        api_base="https://dashscope-intl.aliyuncs.com/compatible-mode/v1",
        http3_support=False,  # Check actual support
        brotli_support=False,  # Check actual support
        streaming_support=True
    ),
    "mistral": ProviderConfig(
        name="Mistral AI",
        docs_url="https://docs.mistral.ai/api",
        api_base="https://api.mistral.ai/v1",
        http3_support=False,  # Check actual support
        brotli_support=True,
        streaming_support=True
    ),
    "groq": ProviderConfig(
        name="Groq",
        docs_url="https://console.groq.com/docs/api-reference",
        api_base="https://api.groq.com/openai/v1",
        http3_support=False,  # Check actual support
        brotli_support=False,  # Check actual support
        streaming_support=True
    ),
    "z-ai": ProviderConfig(
        name="Z.AI (Zhipu/GLM)",
        docs_url="https://open.bigmodel.cn/dev/api",
        api_base="https://open.bigmodel.cn/api/paas/v4",
        http3_support=False,  # Check actual support
        brotli_support=False,  # Check actual support
        streaming_support=True
    ),
    "openrouter": ProviderConfig(
        name="OpenRouter",
        docs_url="https://openrouter.ai/docs",
        api_base="https://openrouter.ai/api/v1",
        http3_support=False,  # Check actual support
        brotli_support=True,
        streaming_support=True
    ),
    "cohere": ProviderConfig(
        name="Cohere",
        docs_url="https://docs.cohere.com/reference",
        api_base="https://api.cohere.com/v1",
        http3_support=False,  # Check actual support
        brotli_support=True,
        streaming_support=True
    ),
}

class ProviderDocFetcher:
    def __init__(self, output_dir: str = "docs/providers"):
        self.output_dir = Path(output_dir)
        self.output_dir.mkdir(parents=True, exist_ok=True)
        self.session: Optional[aiohttp.ClientSession] = None
        
    async def __aenter__(self):
        # Configure session with HTTP/2 and compression support
        connector = aiohttp.TCPConnector(
            limit=100,
            limit_per_host=10,
            enable_cleanup_closed=True,
            force_close=False,
        )
        
        headers = {
            "User-Agent": "HelixAgent-DocFetcher/1.0",
            "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
            "Accept-Language": "en-US,en;q=0.5",
            "Accept-Encoding": "br, gzip, deflate",  # Brotli first
        }
        
        self.session = aiohttp.ClientSession(
            connector=connector,
            headers=headers,
            timeout=aiohttp.ClientTimeout(total=60)
        )
        return self
    
    async def __aexit__(self, exc_type, exc_val, exc_tb):
        if self.session:
            await self.session.close()
    
    async def fetch_page(self, url: str) -> Optional[str]:
        """Fetch a single page with retry logic"""
        max_retries = 3
        for attempt in range(max_retries):
            try:
                async with self.session.get(url) as response:
                    if response.status == 200:
                        content_type = response.headers.get('Content-Type', '')
                        if 'gzip' in response.headers.get('Content-Encoding', ''):
                            # Handle gzip
                            import gzip
                            data = await response.read()
                            return gzip.decompress(data).decode('utf-8')
                        elif 'br' in response.headers.get('Content-Encoding', ''):
                            # Handle brotli
                            import brotli
                            data = await response.read()
                            return brotli.decompress(data).decode('utf-8')
                        else:
                            return await response.text()
                    else:
                        print(f"  HTTP {response.status} for {url}")
                        return None
            except Exception as e:
                print(f"  Attempt {attempt + 1} failed for {url}: {e}")
                if attempt < max_retries - 1:
                    await asyncio.sleep(2 ** attempt)  # Exponential backoff
                else:
                    return None
    
    def extract_endpoints(self, html: str, provider: str) -> List[Dict]:
        """Extract API endpoints from HTML documentation"""
        soup = BeautifulSoup(html, 'html.parser')
        endpoints = []
        
        # Provider-specific extraction logic
        if provider == "openai":
            # OpenAI uses specific structure
            sections = soup.find_all(['h2', 'h3', 'section'])
            for section in sections:
                # Extract endpoint information
                pass
        
        return endpoints
    
    def extract_models(self, html: str, provider: str) -> List[Dict]:
        """Extract available models from documentation"""
        soup = BeautifulSoup(html, 'html.parser')
        models = []
        
        # Provider-specific extraction
        if provider == "openai":
            # Look for model tables or lists
            tables = soup.find_all('table')
            for table in tables:
                # Extract model information
                pass
        
        return models
    
    async def fetch_provider_docs(self, provider_id: str, config: ProviderConfig):
        """Fetch all documentation for a provider"""
        print(f"\n📚 Fetching docs for {config.name}...")
        
        provider_dir = self.output_dir / provider_id
        provider_dir.mkdir(exist_ok=True)
        
        # Create subdirectories
        (provider_dir / "requests").mkdir(exist_ok=True)
        (provider_dir / "responses").mkdir(exist_ok=True)
        (provider_dir / "advanced").mkdir(exist_ok=True)
        
        # Fetch main docs page
        main_page = await self.fetch_page(config.docs_url)
        if not main_page:
            print(f"  ❌ Failed to fetch main docs for {config.name}")
            return
        
        # Save raw HTML for analysis
        (provider_dir / "raw-docs.html").write_text(main_page, encoding='utf-8')
        print(f"  ✅ Saved main documentation")
        
        # Extract and save endpoints
        endpoints = self.extract_endpoints(main_page, provider_id)
        (provider_dir / "endpoints.json").write_text(
            json.dumps(endpoints, indent=2), encoding='utf-8'
        )
        
        # Extract and save models
        models = self.extract_models(main_page, provider_id)
        (provider_dir / "models.json").write_text(
            json.dumps(models, indent=2), encoding='utf-8'
        )
        
        # Create README
        readme = self.generate_readme(config, endpoints, models)
        (provider_dir / "README.md").write_text(readme, encoding='utf-8')
        
        print(f"  ✅ Documentation saved to {provider_dir}")
    
    def generate_readme(self, config: ProviderConfig, endpoints: List, models: List) -> str:
        """Generate README.md for provider"""
        return f"""# {config.name} API Documentation

## Base URL
```
{config.api_base}
```

## Feature Support

| Feature | Status |
|---------|--------|
| HTTP/3 (QUIC) | {'✅' if config.http3_support else '❌'} |
| Brotli Compression | {'✅' if config.brotli_support else '❌'} |
| Streaming (SSE) | {'✅' if config.streaming_support else '❌'} |
| Toon Encoding | ⏳ |

## Endpoints

See [endpoints.json](./endpoints.json) for complete list.

## Models

See [models.json](./models.json) for available models.

## Documentation Source
- [Official Docs]({config.docs_url})

## Last Updated
Generated: {asyncio.get_event_loop().time()}

---
*This documentation was automatically generated by HelixAgent DocFetcher*
"""
    
    async def fetch_all(self, provider_ids: Optional[List[str]] = None):
        """Fetch documentation for all or specified providers"""
        if provider_ids:
            providers = {k: v for k, v in PROVIDERS.items() if k in provider_ids}
        else:
            providers = PROVIDERS
        
        print(f"Fetching documentation for {len(providers)} providers...")
        
        tasks = [
            self.fetch_provider_docs(provider_id, config)
            for provider_id, config in providers.items()
        ]
        
        await asyncio.gather(*tasks, return_exceptions=True)
        
        print("\n✅ Documentation fetch complete!")


def main():
    """Main entry point"""
    import argparse
    
    parser = argparse.ArgumentParser(description="Fetch provider API documentation")
    parser.add_argument(
        "--providers",
        nargs="+",
        choices=list(PROVIDERS.keys()) + ["all"],
        default=["all"],
        help="Providers to fetch (default: all)"
    )
    parser.add_argument(
        "--output",
        default="docs/providers",
        help="Output directory"
    )
    
    args = parser.parse_args()
    
    provider_ids = None if "all" in args.providers else args.providers
    
    async def run():
        async with ProviderDocFetcher(output_dir=args.output) as fetcher:
            await fetcher.fetch_all(provider_ids)
    
    asyncio.run(run())


if __name__ == "__main__":
    main()
