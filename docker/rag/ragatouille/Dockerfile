# docker/rag/ragatouille/Dockerfile
# RAGatouille - ColBERT-based Retrieval Service
FROM python:3.11-slim-bookworm

RUN apt-get update && apt-get install -y \
    curl \
    git \
    && rm -rf /var/lib/apt/lists/*

# Install dependencies
RUN pip install --no-cache-dir \
    ragatouille \
    flask \
    gunicorn \
    torch --index-url https://download.pytorch.org/whl/cpu

# Create directories
RUN mkdir -p /data/indexes /models /app

WORKDIR /app

# Create the server
COPY server.py .

EXPOSE 8018

ENV TRANSFORMERS_CACHE=/models
ENV HF_HOME=/models

CMD ["gunicorn", "-w", "1", "-b", "0.0.0.0:8018", "--timeout", "600", "server:app"]
