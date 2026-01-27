"""HyDE - Hypothetical Document Embeddings Server

HyDE improves retrieval by:
1. Taking a query
2. Generating a hypothetical document that would answer the query
3. Embedding the hypothetical document instead of the query
4. Using that embedding for retrieval
"""
import os
import requests
from flask import Flask, request, jsonify

app = Flask(__name__)

LLM_URL = os.environ.get('HYDE_LLM_URL', 'http://helixagent:7061/v1/chat/completions')
EMBEDDING_URL = os.environ.get('HYDE_EMBEDDING_URL', 'http://sentence-transformers:8016/encode')

def generate_hypothetical_document(query: str, num_documents: int = 1) -> list:
    """Generate hypothetical documents using LLM"""
    prompt = f"""Please write a short passage that would answer the following question.
The passage should be informative and directly address the question.

Question: {query}

Passage:"""

    try:
        response = requests.post(
            LLM_URL,
            json={
                "model": "ensemble",
                "messages": [{"role": "user", "content": prompt}],
                "max_tokens": 256,
                "temperature": 0.7,
                "n": num_documents
            },
            timeout=60
        )
        response.raise_for_status()
        data = response.json()

        documents = []
        for choice in data.get('choices', []):
            content = choice.get('message', {}).get('content', '')
            if content:
                documents.append(content.strip())

        return documents if documents else [query]
    except Exception as e:
        print(f"LLM generation failed: {e}")
        return [query]

def get_embeddings(texts: list) -> list:
    """Get embeddings from embedding service"""
    try:
        response = requests.post(
            EMBEDDING_URL,
            json={"texts": texts},
            timeout=30
        )
        response.raise_for_status()
        data = response.json()
        return data.get('embeddings', [])
    except Exception as e:
        print(f"Embedding failed: {e}")
        return []

@app.route('/health', methods=['GET'])
def health():
    return jsonify({
        "status": "healthy",
        "service": "hyde",
        "llm_url": LLM_URL,
        "embedding_url": EMBEDDING_URL
    })

@app.route('/transform', methods=['POST'])
def transform():
    """Transform a query using HyDE"""
    data = request.json
    query = data.get('query', '')
    num_documents = data.get('num_documents', 1)
    return_documents = data.get('return_documents', False)

    if not query:
        return jsonify({"error": "Query required"}), 400

    # Generate hypothetical documents
    hypothetical_docs = generate_hypothetical_document(query, num_documents)

    # Get embeddings of hypothetical documents
    embeddings = get_embeddings(hypothetical_docs)

    if not embeddings:
        return jsonify({"error": "Failed to generate embeddings"}), 500

    # Average embeddings if multiple documents
    if len(embeddings) > 1:
        import numpy as np
        avg_embedding = np.mean(embeddings, axis=0).tolist()
    else:
        avg_embedding = embeddings[0]

    result = {
        "query": query,
        "embedding": avg_embedding,
        "dimension": len(avg_embedding)
    }

    if return_documents:
        result["hypothetical_documents"] = hypothetical_docs

    return jsonify(result)

@app.route('/batch_transform', methods=['POST'])
def batch_transform():
    """Transform multiple queries using HyDE"""
    data = request.json
    queries = data.get('queries', [])
    num_documents = data.get('num_documents', 1)

    if not queries:
        return jsonify({"error": "Queries required"}), 400

    results = []
    for query in queries:
        hypothetical_docs = generate_hypothetical_document(query, num_documents)
        embeddings = get_embeddings(hypothetical_docs)

        if embeddings:
            import numpy as np
            if len(embeddings) > 1:
                avg_embedding = np.mean(embeddings, axis=0).tolist()
            else:
                avg_embedding = embeddings[0]
            results.append({
                "query": query,
                "embedding": avg_embedding
            })
        else:
            results.append({
                "query": query,
                "error": "Failed to generate embedding"
            })

    return jsonify({"results": results})

if __name__ == '__main__':
    port = int(os.environ.get('HYDE_PORT', 8019))
    app.run(host='0.0.0.0', port=port)
