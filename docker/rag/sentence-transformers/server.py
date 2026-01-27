"""Sentence Transformers Embedding Server"""
import os
import json
from flask import Flask, request, jsonify
from sentence_transformers import SentenceTransformer

app = Flask(__name__)

# Model cache
models = {}

def get_model(model_name: str = None):
    """Get or load a sentence transformer model"""
    if model_name is None:
        model_name = os.environ.get('ST_DEFAULT_MODEL', 'all-mpnet-base-v2')

    if model_name not in models:
        models[model_name] = SentenceTransformer(model_name)
    return models[model_name]

@app.route('/health', methods=['GET'])
def health():
    return jsonify({
        "status": "healthy",
        "service": "sentence-transformers",
        "default_model": os.environ.get('ST_DEFAULT_MODEL', 'all-mpnet-base-v2'),
        "loaded_models": list(models.keys())
    })

@app.route('/encode', methods=['POST'])
def encode():
    data = request.json
    texts = data.get('texts', [])
    model_name = data.get('model')

    if not texts:
        return jsonify({"error": "No texts provided"}), 400

    if isinstance(texts, str):
        texts = [texts]

    model = get_model(model_name)
    embeddings = model.encode(texts, convert_to_numpy=True)

    return jsonify({
        "embeddings": embeddings.tolist(),
        "model": model_name or os.environ.get('ST_DEFAULT_MODEL', 'all-mpnet-base-v2'),
        "dimension": len(embeddings[0])
    })

@app.route('/models', methods=['GET'])
def list_models():
    available = [
        "all-mpnet-base-v2",
        "all-MiniLM-L6-v2",
        "all-MiniLM-L12-v2",
        "paraphrase-multilingual-mpnet-base-v2",
        "multi-qa-mpnet-base-dot-v1"
    ]
    return jsonify({
        "available": available,
        "loaded": list(models.keys()),
        "default": os.environ.get('ST_DEFAULT_MODEL', 'all-mpnet-base-v2')
    })

@app.route('/similarity', methods=['POST'])
def similarity():
    data = request.json
    text1 = data.get('text1', '')
    text2 = data.get('text2', '')
    model_name = data.get('model')

    if not text1 or not text2:
        return jsonify({"error": "Both text1 and text2 required"}), 400

    model = get_model(model_name)
    embeddings = model.encode([text1, text2])

    # Cosine similarity
    from numpy import dot
    from numpy.linalg import norm

    similarity = float(dot(embeddings[0], embeddings[1]) / (norm(embeddings[0]) * norm(embeddings[1])))

    return jsonify({
        "similarity": similarity,
        "model": model_name or os.environ.get('ST_DEFAULT_MODEL', 'all-mpnet-base-v2')
    })

if __name__ == '__main__':
    port = int(os.environ.get('ST_PORT', 8016))
    # Pre-load default model
    get_model()
    app.run(host='0.0.0.0', port=port)
