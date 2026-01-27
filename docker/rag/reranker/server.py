"""Cross-Encoder Reranking Server

Reranks retrieval results using cross-encoder models that consider
both query and document together for more accurate relevance scoring.
"""
import os
from flask import Flask, request, jsonify
from sentence_transformers import CrossEncoder

app = Flask(__name__)

# Model cache
models = {}

def get_model(model_name: str = None):
    """Get or load a cross-encoder model"""
    if model_name is None:
        model_name = os.environ.get('RERANKER_MODEL', 'cross-encoder/ms-marco-MiniLM-L-6-v2')

    if model_name not in models:
        models[model_name] = CrossEncoder(model_name)
    return models[model_name]

@app.route('/health', methods=['GET'])
def health():
    return jsonify({
        "status": "healthy",
        "service": "reranker",
        "default_model": os.environ.get('RERANKER_MODEL', 'cross-encoder/ms-marco-MiniLM-L-6-v2'),
        "loaded_models": list(models.keys())
    })

@app.route('/rerank', methods=['POST'])
def rerank():
    """Rerank documents based on relevance to query"""
    data = request.json
    query = data.get('query', '')
    documents = data.get('documents', [])
    model_name = data.get('model')
    top_k = data.get('top_k')

    if not query:
        return jsonify({"error": "Query required"}), 400

    if not documents:
        return jsonify({"error": "Documents required"}), 400

    model = get_model(model_name)

    # Create query-document pairs
    pairs = [[query, doc if isinstance(doc, str) else doc.get('text', doc.get('content', str(doc)))]
             for doc in documents]

    # Get scores
    scores = model.predict(pairs)

    # Create results with original documents and scores
    results = []
    for i, (doc, score) in enumerate(zip(documents, scores)):
        result = {
            "index": i,
            "score": float(score),
            "document": doc
        }
        results.append(result)

    # Sort by score descending
    results.sort(key=lambda x: x["score"], reverse=True)

    # Apply top_k if specified
    if top_k:
        results = results[:top_k]

    return jsonify({
        "query": query,
        "results": results,
        "model": model_name or os.environ.get('RERANKER_MODEL', 'cross-encoder/ms-marco-MiniLM-L-6-v2')
    })

@app.route('/score', methods=['POST'])
def score():
    """Score a single query-document pair"""
    data = request.json
    query = data.get('query', '')
    document = data.get('document', '')
    model_name = data.get('model')

    if not query or not document:
        return jsonify({"error": "Query and document required"}), 400

    model = get_model(model_name)
    score = model.predict([[query, document]])[0]

    return jsonify({
        "query": query,
        "document": document,
        "score": float(score),
        "model": model_name or os.environ.get('RERANKER_MODEL', 'cross-encoder/ms-marco-MiniLM-L-6-v2')
    })

@app.route('/batch_score', methods=['POST'])
def batch_score():
    """Score multiple query-document pairs"""
    data = request.json
    pairs = data.get('pairs', [])
    model_name = data.get('model')

    if not pairs:
        return jsonify({"error": "Pairs required"}), 400

    model = get_model(model_name)

    # Convert pairs to list format
    pair_list = [[p.get('query', p[0] if isinstance(p, list) else ''),
                  p.get('document', p[1] if isinstance(p, list) else '')]
                 for p in pairs]

    scores = model.predict(pair_list)

    results = [{"query": p[0], "document": p[1], "score": float(s)}
               for p, s in zip(pair_list, scores)]

    return jsonify({
        "results": results,
        "model": model_name or os.environ.get('RERANKER_MODEL', 'cross-encoder/ms-marco-MiniLM-L-6-v2')
    })

@app.route('/models', methods=['GET'])
def list_models():
    available = [
        "cross-encoder/ms-marco-MiniLM-L-6-v2",
        "cross-encoder/ms-marco-MiniLM-L-12-v2",
        "cross-encoder/ms-marco-TinyBERT-L-2-v2",
        "cross-encoder/stsb-roberta-base",
        "BAAI/bge-reranker-base",
        "BAAI/bge-reranker-large"
    ]
    return jsonify({
        "available": available,
        "loaded": list(models.keys()),
        "default": os.environ.get('RERANKER_MODEL', 'cross-encoder/ms-marco-MiniLM-L-6-v2')
    })

if __name__ == '__main__':
    port = int(os.environ.get('RERANKER_PORT', 8021))
    # Pre-load default model
    get_model()
    app.run(host='0.0.0.0', port=port)
