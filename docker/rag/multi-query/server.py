"""Multi-Query Retrieval Server

Improves retrieval by:
1. Generating multiple query variations
2. Retrieving documents for each variation
3. Combining results with deduplication and re-ranking
"""
import os
import requests
from flask import Flask, request, jsonify

app = Flask(__name__)

LLM_URL = os.environ.get('MQ_LLM_URL', 'http://helixagent:7061/v1/chat/completions')
VECTOR_URL = os.environ.get('MQ_VECTOR_URL', 'http://qdrant:6333')

def generate_query_variations(query: str, num_variations: int = 3) -> list:
    """Generate variations of the query using LLM"""
    prompt = f"""Generate {num_variations} different variations of the following search query.
Each variation should capture the same intent but use different wording or perspective.

Original query: {query}

Output only the variations, one per line, without numbering or explanations."""

    try:
        response = requests.post(
            LLM_URL,
            json={
                "model": "ensemble",
                "messages": [{"role": "user", "content": prompt}],
                "max_tokens": 256,
                "temperature": 0.7
            },
            timeout=60
        )
        response.raise_for_status()
        data = response.json()

        content = data.get('choices', [{}])[0].get('message', {}).get('content', '')
        variations = [v.strip() for v in content.strip().split('\n') if v.strip()]

        # Always include original query
        if query not in variations:
            variations = [query] + variations

        return variations[:num_variations + 1]
    except Exception as e:
        print(f"Query variation generation failed: {e}")
        return [query]

def search_vector_db(query: str, collection: str, limit: int = 10):
    """Search vector database"""
    try:
        # This assumes Qdrant API
        response = requests.post(
            f"{VECTOR_URL}/collections/{collection}/points/search",
            json={
                "query": query,
                "limit": limit,
                "with_payload": True
            },
            timeout=30
        )
        if response.status_code == 200:
            return response.json().get('result', [])
    except Exception as e:
        print(f"Vector search failed: {e}")
    return []

@app.route('/health', methods=['GET'])
def health():
    return jsonify({
        "status": "healthy",
        "service": "multi-query",
        "llm_url": LLM_URL,
        "vector_url": VECTOR_URL
    })

@app.route('/expand', methods=['POST'])
def expand():
    """Expand a query into multiple variations"""
    data = request.json
    query = data.get('query', '')
    num_variations = data.get('num_variations', 3)

    if not query:
        return jsonify({"error": "Query required"}), 400

    variations = generate_query_variations(query, num_variations)

    return jsonify({
        "original": query,
        "variations": variations,
        "count": len(variations)
    })

@app.route('/search', methods=['POST'])
def search():
    """Search with multiple query variations and combine results"""
    data = request.json
    query = data.get('query', '')
    collection = data.get('collection', 'documents')
    num_variations = data.get('num_variations', 3)
    limit = data.get('limit', 10)

    if not query:
        return jsonify({"error": "Query required"}), 400

    # Generate query variations
    variations = generate_query_variations(query, num_variations)

    # Search with each variation
    all_results = []
    seen_ids = set()

    for i, variation in enumerate(variations):
        results = search_vector_db(variation, collection, limit)
        for result in results:
            result_id = result.get('id')
            if result_id and result_id not in seen_ids:
                seen_ids.add(result_id)
                # Add source query info
                result['source_query'] = variation
                result['query_index'] = i
                all_results.append(result)

    # Sort by score
    all_results.sort(key=lambda x: x.get('score', 0), reverse=True)

    return jsonify({
        "query": query,
        "variations_used": variations,
        "results": all_results[:limit],
        "total_found": len(all_results)
    })

@app.route('/fuse', methods=['POST'])
def fuse():
    """Reciprocal Rank Fusion of multiple result sets"""
    data = request.json
    result_sets = data.get('result_sets', [])
    k = data.get('k', 60)  # RRF constant

    if not result_sets:
        return jsonify({"error": "Result sets required"}), 400

    # Calculate RRF scores
    rrf_scores = {}
    for results in result_sets:
        for rank, result in enumerate(results, 1):
            doc_id = result.get('id', str(rank))
            if doc_id not in rrf_scores:
                rrf_scores[doc_id] = {"doc": result, "score": 0}
            rrf_scores[doc_id]["score"] += 1 / (k + rank)

    # Sort by RRF score
    fused = sorted(rrf_scores.values(), key=lambda x: x["score"], reverse=True)

    return jsonify({
        "results": [{"doc": item["doc"], "rrf_score": item["score"]} for item in fused],
        "count": len(fused)
    })

if __name__ == '__main__':
    port = int(os.environ.get('MQ_PORT', 8020))
    app.run(host='0.0.0.0', port=port)
