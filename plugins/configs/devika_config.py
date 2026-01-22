"""
HelixAgent configuration for Devika

Add to your Devika configuration or import this module.
"""

HELIXAGENT_CONFIG = {
    "endpoint": "https://localhost:7061",
    "transport": {
        "prefer_http3": True,
        "enable_toon": True,
        "enable_brotli": True,
        "timeout": 30,
    },
    "events": {
        "transport": "sse",
        "subscribe_to_debates": True,
        "subscribe_to_tasks": True,
    },
    "debate": {
        "show_phase_indicators": True,
        "show_confidence_scores": True,
        "enable_multi_pass_validation": True,
    },
}

# LiteLLM-compatible provider configuration
LITELLM_HELIXAGENT = {
    "model": "openai/helix-debate-ensemble",
    "api_base": "https://localhost:7061/v1",
    "api_key": "helixagent",  # Use actual key if required
}
