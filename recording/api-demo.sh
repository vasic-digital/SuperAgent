#!/bin/bash
echo "🌐 SuperAgent API Demo"
echo "===================="

echo "1. Making simple completion request..."
curl -X POST http://localhost:8080/api/v1/completion \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "Explain the benefits of multi-provider AI orchestration in 3 sentences",
    "providers": ["claude", "gemini"],
    "max_tokens": 150
  }'

echo -e "\n\n2. Making AI debate request..."
curl -X POST http://localhost:8080/api/v1/completion \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "What are the most important security considerations for AI infrastructure?",
    "providers": ["claude", "gemini", "deepseek"],
    "ai_debate": true,
    "max_tokens": 200
  }'

echo -e "\n\n3. Checking health status..."
curl http://localhost:8080/health

echo -e "\n\n✅ API demo complete!"
