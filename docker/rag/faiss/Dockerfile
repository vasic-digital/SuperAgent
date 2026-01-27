# docker/rag/faiss/Dockerfile
# FAISS Vector Search Server
FROM python:3.11-slim-bookworm

RUN apt-get update && apt-get install -y \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Install FAISS and dependencies
RUN pip install --no-cache-dir \
    faiss-cpu \
    flask \
    numpy \
    gunicorn

# Create directories
RUN mkdir -p /data/indexes /app

WORKDIR /app

# Create the FAISS server
COPY server.py .

EXPOSE 8015

CMD ["gunicorn", "-w", "4", "-b", "0.0.0.0:8015", "server:app"]
