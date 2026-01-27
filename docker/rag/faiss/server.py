"""FAISS Vector Search Server"""
import os
import json
import numpy as np
from flask import Flask, request, jsonify

try:
    import faiss
except ImportError:
    faiss = None

app = Flask(__name__)

# In-memory index storage
indexes = {}

def get_or_create_index(name: str, dimension: int = 768):
    """Get or create a FAISS index"""
    if name not in indexes:
        if faiss:
            indexes[name] = faiss.IndexFlatL2(dimension)
        else:
            indexes[name] = {"vectors": [], "dimension": dimension}
    return indexes[name]

@app.route('/health', methods=['GET'])
def health():
    return jsonify({
        "status": "healthy",
        "service": "faiss-server",
        "faiss_available": faiss is not None,
        "indexes": len(indexes)
    })

@app.route('/index', methods=['POST'])
def create_index():
    data = request.json
    name = data.get('name', 'default')
    dimension = data.get('dimension', 768)
    get_or_create_index(name, dimension)
    return jsonify({"status": "created", "name": name, "dimension": dimension})

@app.route('/add', methods=['POST'])
def add_vectors():
    data = request.json
    name = data.get('index', 'default')
    vectors = np.array(data.get('vectors', []), dtype=np.float32)
    ids = data.get('ids', list(range(len(vectors))))

    if len(vectors) == 0:
        return jsonify({"error": "No vectors provided"}), 400

    index = get_or_create_index(name, vectors.shape[1])

    if faiss:
        index.add(vectors)
    else:
        for v in vectors:
            index["vectors"].append(v.tolist())

    return jsonify({"status": "added", "count": len(vectors)})

@app.route('/search', methods=['POST'])
def search():
    data = request.json
    name = data.get('index', 'default')
    query = np.array([data.get('query', [])], dtype=np.float32)
    k = data.get('k', 10)

    if name not in indexes:
        return jsonify({"error": "Index not found"}), 404

    index = indexes[name]

    if faiss:
        distances, indices = index.search(query, k)
        results = [{"id": int(idx), "distance": float(dist)}
                   for idx, dist in zip(indices[0], distances[0]) if idx >= 0]
    else:
        # Simple fallback for when FAISS is not available
        vectors = index["vectors"]
        if not vectors:
            results = []
        else:
            distances = []
            for i, v in enumerate(vectors):
                dist = np.linalg.norm(query[0] - np.array(v))
                distances.append((i, dist))
            distances.sort(key=lambda x: x[1])
            results = [{"id": idx, "distance": float(dist)} for idx, dist in distances[:k]]

    return jsonify({"results": results})

@app.route('/delete', methods=['POST'])
def delete_index():
    data = request.json
    name = data.get('name', 'default')
    if name in indexes:
        del indexes[name]
        return jsonify({"status": "deleted", "name": name})
    return jsonify({"error": "Index not found"}), 404

if __name__ == '__main__':
    port = int(os.environ.get('FAISS_PORT', 8015))
    app.run(host='0.0.0.0', port=port)
