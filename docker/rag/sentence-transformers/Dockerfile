# docker/rag/sentence-transformers/Dockerfile
# Sentence Transformers Embedding Service
FROM python:3.11-slim-bookworm

RUN apt-get update && apt-get install -y \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Install sentence-transformers and dependencies
RUN pip install --no-cache-dir \
    sentence-transformers \
    flask \
    gunicorn \
    torch --index-url https://download.pytorch.org/whl/cpu

# Create directories
RUN mkdir -p /models /app

WORKDIR /app

# Create the server
COPY server.py .

# Pre-download default model
RUN python -c "from sentence_transformers import SentenceTransformer; SentenceTransformer('all-mpnet-base-v2')"

EXPOSE 8016

ENV TRANSFORMERS_CACHE=/models
ENV HF_HOME=/models

CMD ["gunicorn", "-w", "2", "-b", "0.0.0.0:8016", "--timeout", "300", "server:app"]
