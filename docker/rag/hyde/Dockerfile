# docker/rag/hyde/Dockerfile
# HyDE - Hypothetical Document Embeddings Service
FROM python:3.11-slim-bookworm

RUN apt-get update && apt-get install -y \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Install dependencies
RUN pip install --no-cache-dir \
    flask \
    gunicorn \
    requests \
    numpy

# Create app directory
RUN mkdir -p /app

WORKDIR /app

# Create the server
COPY server.py .

EXPOSE 8019

CMD ["gunicorn", "-w", "2", "-b", "0.0.0.0:8019", "--timeout", "120", "server:app"]
