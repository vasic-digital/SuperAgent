"""
HelixAgent Python SDK

A Python client for the HelixAgent AI orchestration platform.
Compatible with OpenAI-style API calls.
"""

from .client import (
    HelixAgent,
    Debates,
    Protocols,
    Analytics,
    Plugins,
    Templates,
)
from .exceptions import (
    HelixAgentError,
    AuthenticationError,
    RateLimitError,
    APIError,
    ConnectionError,
    ValidationError,
    TimeoutError,
)
from .types import (
    ChatMessage,
    ChatCompletionResponse,
    ChatCompletionChoice,
    ChatCompletionChunk,
    StreamChoice,
    StreamDelta,
    Usage,
    Model,
    EnsembleConfig,
    ParticipantConfig,
    DebateConfig,
    DebateResponse,
    DebateStatus,
    DebateResult,
    LSPPosition,
    PluginInfo,
    TemplateInfo,
)

__version__ = "0.1.0"
__all__ = [
    # Main client
    "HelixAgent",
    # API classes
    "Debates",
    "Protocols",
    "Analytics",
    "Plugins",
    "Templates",
    # Exceptions
    "HelixAgentError",
    "AuthenticationError",
    "RateLimitError",
    "APIError",
    "ConnectionError",
    "ValidationError",
    "TimeoutError",
    # Chat types
    "ChatMessage",
    "ChatCompletionResponse",
    "ChatCompletionChoice",
    "ChatCompletionChunk",
    "StreamChoice",
    "StreamDelta",
    "Usage",
    # Model types
    "Model",
    "EnsembleConfig",
    # Debate types
    "ParticipantConfig",
    "DebateConfig",
    "DebateResponse",
    "DebateStatus",
    "DebateResult",
    # Protocol types
    "LSPPosition",
    # Plugin/Template types
    "PluginInfo",
    "TemplateInfo",
]
