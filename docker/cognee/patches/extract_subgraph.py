from cognee.modules.graph.cognee_graph.CogneeGraph import CogneeGraph


async def extract_subgraph(subgraphs: list):
    """
    Extract edges from subgraphs. Handles both CogneeGraph objects and raw data gracefully.

    When non-OpenAI LLMs are used with instructor's structured output,
    the response may be a raw string instead of a CogneeGraph object.
    This patched version handles that case by skipping string inputs.
    """
    for subgraph in subgraphs:
        if isinstance(subgraph, str):
            # Raw text data - no edges to extract, skip
            continue
        elif isinstance(subgraph, CogneeGraph):
            for edge in subgraph.edges:
                yield edge
        elif hasattr(subgraph, "edges"):
            for edge in subgraph.edges:
                yield edge
