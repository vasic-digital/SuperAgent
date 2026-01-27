# docker/rag/bge-m3/Dockerfile
# BGE-M3 Multilingual Embedding Service
FROM python:3.11-slim-bookworm

RUN apt-get update && apt-get install -y \
    curl \
    git \
    && rm -rf /var/lib/apt/lists/*

# Install dependencies
RUN pip install --no-cache-dir \
    FlagEmbedding \
    flask \
    gunicorn \
    torch --index-url https://download.pytorch.org/whl/cpu

# Create directories
RUN mkdir -p /models /app

WORKDIR /app

# Create the server
COPY server.py .

EXPOSE 8017

ENV TRANSFORMERS_CACHE=/models
ENV HF_HOME=/models

CMD ["gunicorn", "-w", "1", "-b", "0.0.0.0:8017", "--timeout", "600", "server:app"]
