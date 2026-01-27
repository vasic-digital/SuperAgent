"""BGE-M3 Multilingual Embedding Server"""
import os
from flask import Flask, request, jsonify

app = Flask(__name__)

# Lazy load model
model = None

def get_model():
    """Lazy load the BGE-M3 model"""
    global model
    if model is None:
        try:
            from FlagEmbedding import BGEM3FlagModel
            model_name = os.environ.get('BGE_MODEL', 'BAAI/bge-m3')
            model = BGEM3FlagModel(model_name, use_fp16=False)
        except Exception as e:
            print(f"Failed to load BGE-M3: {e}")
            # Fallback to sentence-transformers
            from sentence_transformers import SentenceTransformer
            model = SentenceTransformer('all-mpnet-base-v2')
    return model

@app.route('/health', methods=['GET'])
def health():
    return jsonify({
        "status": "healthy",
        "service": "bge-m3",
        "model": os.environ.get('BGE_MODEL', 'BAAI/bge-m3'),
        "loaded": model is not None
    })

@app.route('/encode', methods=['POST'])
def encode():
    data = request.json
    texts = data.get('texts', [])

    if not texts:
        return jsonify({"error": "No texts provided"}), 400

    if isinstance(texts, str):
        texts = [texts]

    m = get_model()

    # Check if it's BGE-M3 or fallback
    if hasattr(m, 'encode'):
        if hasattr(m, 'encode_queries'):
            # BGE-M3 model
            embeddings = m.encode(texts)
            if isinstance(embeddings, dict):
                embeddings = embeddings.get('dense_vecs', embeddings)
        else:
            # Sentence transformer fallback
            embeddings = m.encode(texts)

        if hasattr(embeddings, 'tolist'):
            embeddings = embeddings.tolist()
    else:
        return jsonify({"error": "Model not properly loaded"}), 500

    return jsonify({
        "embeddings": embeddings,
        "model": os.environ.get('BGE_MODEL', 'BAAI/bge-m3'),
        "dimension": len(embeddings[0]) if embeddings else 0
    })

@app.route('/encode_multi', methods=['POST'])
def encode_multi():
    """Encode with multiple representations (dense, sparse, colbert)"""
    data = request.json
    texts = data.get('texts', [])
    return_dense = data.get('return_dense', True)
    return_sparse = data.get('return_sparse', False)
    return_colbert = data.get('return_colbert', False)

    if not texts:
        return jsonify({"error": "No texts provided"}), 400

    if isinstance(texts, str):
        texts = [texts]

    m = get_model()

    result = {"texts_count": len(texts)}

    if hasattr(m, 'encode'):
        if hasattr(m, 'encode_queries'):
            # Full BGE-M3
            output = m.encode(
                texts,
                return_dense=return_dense,
                return_sparse=return_sparse,
                return_colbert_vecs=return_colbert
            )
            if return_dense and 'dense_vecs' in output:
                result['dense'] = output['dense_vecs'].tolist()
            if return_sparse and 'lexical_weights' in output:
                result['sparse'] = output['lexical_weights']
            if return_colbert and 'colbert_vecs' in output:
                result['colbert'] = [v.tolist() for v in output['colbert_vecs']]
        else:
            # Fallback - only dense
            embeddings = m.encode(texts)
            result['dense'] = embeddings.tolist() if hasattr(embeddings, 'tolist') else embeddings

    return jsonify(result)

if __name__ == '__main__':
    port = int(os.environ.get('BGE_PORT', 8017))
    app.run(host='0.0.0.0', port=port)
