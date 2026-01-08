#!/bin/bash
echo "🚀 HelixAgent Demo - Multi-Provider AI Orchestration"
echo "==================================================="

echo "1. Installing HelixAgent..."
git clone https://github.com/helixagent/helixagent.git
cd helixagent
make build

echo "2. Creating configuration..."
cat > config.yaml << 'CONFIG'
providers:
  claude:
    api_key: ${CLAUDE_API_KEY}
    model: claude-3-opus-20240229
  
  gemini:
    api_key: ${GEMINI_API_KEY}
    model: gemini-pro
  
  deepseek:
    api_key: ${DEEPSEEK_API_KEY}
    model: deepseek-chat

ai_debate:
  enabled: true
  participants:
    - name: "Claude"
      role: "analyst"
      llms: ["claude"]
    - name: "Gemini"
      role: "critic"
      llms: ["gemini"]
CONFIG

echo "3. Starting HelixAgent..."
./helixagent --config config.yaml

echo "🎉 HelixAgent is running! Visit http://localhost:8080"
