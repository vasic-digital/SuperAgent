FROM cognee/cognee:main

# Install additional packages required by LiteLLM for various providers:
# - google-auth, google-generativeai, protobuf: Gemini provider support
# - sentencepiece, mistral-common: Mistral tokenizer support
# Using --upgrade to merge namespace packages (google/) correctly
RUN pip install --no-cache-dir --upgrade --target=/app/.venv/lib/python3.12/site-packages \
    google-auth \
    google-auth-httplib2 \
    google-api-core \
    google-generativeai \
    protobuf \
    sentencepiece \
    "mistral-common[sentencepiece]"

# Patch extract_subgraph_chunks.py to handle raw string data gracefully
# When non-OpenAI LLMs are used, instructor's structured output may return
# raw strings instead of CogneeGraph objects. This patch adds defensive
# type checking to prevent AttributeError: 'str' has no attribute 'nodes'
COPY patches/extract_subgraph_chunks.py /app/cognee/tasks/memify/extract_subgraph_chunks.py
COPY patches/extract_subgraph.py /app/cognee/tasks/memify/extract_subgraph.py

# Clear bytecache so Python uses the patched source files
RUN rm -f /app/cognee/tasks/memify/__pycache__/extract_subgraph_chunks.cpython-*.pyc \
          /app/cognee/tasks/memify/__pycache__/extract_subgraph.cpython-*.pyc
