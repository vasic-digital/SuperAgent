"""RAGatouille ColBERT Retrieval Server"""
import os
from flask import Flask, request, jsonify

app = Flask(__name__)

# Lazy load model
rag = None
indexes = {}

def get_rag():
    """Lazy load the RAGatouille model"""
    global rag
    if rag is None:
        try:
            from ragatouille import RAGPretrainedModel
            model_name = os.environ.get('RAGATOUILLE_MODEL', 'colbert-ir/colbertv2.0')
            rag = RAGPretrainedModel.from_pretrained(model_name)
        except Exception as e:
            print(f"Failed to load RAGatouille: {e}")
            rag = "failed"
    return rag if rag != "failed" else None

@app.route('/health', methods=['GET'])
def health():
    return jsonify({
        "status": "healthy",
        "service": "ragatouille",
        "model": os.environ.get('RAGATOUILLE_MODEL', 'colbert-ir/colbertv2.0'),
        "indexes": list(indexes.keys())
    })

@app.route('/index', methods=['POST'])
def create_index():
    """Create a new index from documents"""
    data = request.json
    name = data.get('name', 'default')
    documents = data.get('documents', [])

    if not documents:
        return jsonify({"error": "No documents provided"}), 400

    model = get_rag()
    if model is None:
        return jsonify({"error": "Model not available"}), 500

    try:
        index_path = f"/data/indexes/{name}"
        model.index(
            collection=documents,
            index_name=name,
            max_document_length=512,
            split_documents=True
        )
        indexes[name] = {"path": index_path, "doc_count": len(documents)}
        return jsonify({"status": "created", "name": name, "documents": len(documents)})
    except Exception as e:
        return jsonify({"error": str(e)}), 500

@app.route('/search', methods=['POST'])
def search():
    """Search the index"""
    data = request.json
    name = data.get('index', 'default')
    query = data.get('query', '')
    k = data.get('k', 10)

    if not query:
        return jsonify({"error": "Query required"}), 400

    model = get_rag()
    if model is None:
        return jsonify({"error": "Model not available"}), 500

    try:
        results = model.search(query, k=k)
        return jsonify({
            "results": results,
            "query": query,
            "k": k
        })
    except Exception as e:
        return jsonify({"error": str(e)}), 500

@app.route('/encode', methods=['POST'])
def encode():
    """Encode queries/documents with ColBERT"""
    data = request.json
    texts = data.get('texts', [])

    if not texts:
        return jsonify({"error": "No texts provided"}), 400

    if isinstance(texts, str):
        texts = [texts]

    model = get_rag()
    if model is None:
        return jsonify({"error": "Model not available"}), 500

    try:
        embeddings = model.encode(texts)
        return jsonify({
            "embeddings": [e.tolist() if hasattr(e, 'tolist') else e for e in embeddings],
            "count": len(texts)
        })
    except Exception as e:
        return jsonify({"error": str(e)}), 500

if __name__ == '__main__':
    port = int(os.environ.get('RAGATOUILLE_PORT', 8018))
    app.run(host='0.0.0.0', port=port)
