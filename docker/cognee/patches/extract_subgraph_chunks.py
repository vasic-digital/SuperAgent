from cognee.modules.graph.cognee_graph.CogneeGraph import CogneeGraph


async def extract_subgraph_chunks(subgraphs: list):
    """
    Get all Document Chunks from subgraphs and forward to next task in pipeline.
    Handles both CogneeGraph objects and raw string data gracefully.

    When non-OpenAI LLMs are used with instructor's structured output,
    the response may be a raw string instead of a CogneeGraph object.
    This patched version handles that case by yielding the string directly.
    """
    for subgraph in subgraphs:
        if isinstance(subgraph, str):
            # Raw text data passed directly - yield as-is
            yield subgraph
        elif isinstance(subgraph, CogneeGraph):
            for node in subgraph.nodes.values():
                if node.attributes.get("type") == "DocumentChunk":
                    yield node.attributes.get("text", "")
        elif hasattr(subgraph, "nodes"):
            for node in subgraph.nodes.values():
                if hasattr(node, "attributes") and node.attributes.get("type") == "DocumentChunk":
                    yield node.attributes.get("text", "")
        else:
            # Fallback: convert to string
            yield str(subgraph)
