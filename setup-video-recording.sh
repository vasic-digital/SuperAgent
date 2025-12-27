#!/bin/bash

# SuperAgent Video Production Setup Script
# This script sets up everything needed for recording the first video

echo "🎬 SuperAgent Video Production Setup"
echo "===================================="
echo ""

echo "📁 Checking project structure..."
if [ ! -f "Website/package.json" ]; then
    echo "❌ Error: Website directory not found"
    exit 1
fi

echo "✅ Project structure verified"
echo ""

echo "🔧 Setting up recording environment..."

# Create recording directory
mkdir -p recording
cd recording

echo "📝 Creating video recording structure..."

# Create project files for demonstration
cat > superagent-demo.sh << 'EOF'
#!/bin/bash
echo "🚀 SuperAgent Demo - Multi-Provider AI Orchestration"
echo "==================================================="

echo "1. Installing SuperAgent..."
git clone https://github.com/superagent/superagent.git
cd superagent
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

echo "3. Starting SuperAgent..."
./superagent --config config.yaml

echo "🎉 SuperAgent is running! Visit http://localhost:8080"
EOF

chmod +x superagent-demo.sh

# Create demo API request
cat > api-demo.sh << 'EOF'
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
EOF

chmod +x api-demo.sh

# Create monitoring dashboard screenshot
cat > dashboard.html << 'EOF'
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>SuperAgent Monitoring Dashboard</title>
    <style>
        body { font-family: 'Inter', sans-serif; margin: 0; padding: 20px; background: #0f172a; color: #f8fafc; }
        .dashboard { max-width: 1200px; margin: 0 auto; }
        .header { background: linear-gradient(135deg, #2563eb, #3b82f6); padding: 30px; border-radius: 10px; margin-bottom: 20px; }
        .metrics { display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 20px; margin-bottom: 20px; }
        .metric-card { background: #1e293b; padding: 20px; border-radius: 8px; }
        .chart-container { background: #1e293b; padding: 20px; border-radius: 8px; margin-bottom: 20px; }
        .providers { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; }
        .provider-card { background: #1e293b; padding: 20px; border-radius: 8px; text-align: center; }
        .healthy { color: #10b981; }
        .warning { color: #f59e0b; }
        .critical { color: #ef4444; }
    </style>
</head>
<body>
    <div class="dashboard">
        <div class="header">
            <h1>🚀 SuperAgent Monitoring Dashboard</h1>
            <p>Real-time AI orchestration performance metrics</p>
        </div>
        
        <div class="metrics">
            <div class="metric-card">
                <h3>Total Requests</h3>
                <p class="stat">12,847</p>
                <p class="trend healthy">↑ 24% from last week</p>
            </div>
            <div class="metric-card">
                <h3>Average Response Time</h3>
                <p class="stat">142ms</p>
                <p class="trend healthy">↓ 18% improvement</p>
            </div>
            <div class="metric-card">
                <h3>Cost Savings</h3>
                <p class="stat">$3,847</p>
                <p class="trend healthy">35% vs single provider</p>
            </div>
            <div class="metric-card">
                <h3>Uptime</h3>
                <p class="stat">99.97%</p>
                <p class="trend healthy">Last 30 days</p>
            </div>
        </div>
        
        <div class="chart-container">
            <h3>Request Volume by Provider</h3>
            <div style="height: 300px; background: #0f172a; border-radius: 8px; padding: 20px;">
                <!-- Chart would be here -->
                <div style="display: flex; height: 250px; align-items: flex-end; gap: 30px; justify-content: center;">
                    <div style="background: #3b82f6; height: 80%; width: 50px; border-radius: 4px;">
                        <div style="text-align: center; margin-top: 10px;">Claude</div>
                    </div>
                    <div style="background: #10b981; height: 65%; width: 50px; border-radius: 4px;">
                        <div style="text-align: center; margin-top: 10px;">Gemini</div>
                    </div>
                    <div style="background: #f59e0b; height: 45%; width: 50px; border-radius: 4px;">
                        <div style="text-align: center; margin-top: 10px;">DeepSeek</div>
                    </div>
                </div>
            </div>
        </div>
        
        <div class="providers">
            <div class="provider-card">
                <h3>Claude</h3>
                <p class="status healthy">✓ Healthy</p>
                <p>Requests: 4,892</p>
                <p>Avg Time: 156ms</p>
            </div>
            <div class="provider-card">
                <h3>Gemini</h3>
                <p class="status healthy">✓ Healthy</p>
                <p>Requests: 3,847</p>
                <p>Avg Time: 128ms</p>
            </div>
            <div class="provider-card">
                <h3>DeepSeek</h3>
                <p class="status healthy">✓ Healthy</p>
                <p>Requests: 2,918</p>
                <p>Avg Time: 142ms</p>
            </div>
            <div class="provider-card">
                <h3>Ollama</h3>
                <p class="status warning">⚠️ Degraded</p>
                <p>Requests: 1,190</p>
                <p>Avg Time: 210ms</p>
            </div>
        </div>
    </div>
</body>
</html>
EOF

# Create terminal theme for recording
cat > terminal-theme.md << 'EOF'
# Terminal Recording Setup

## Recommended Theme:
- **Font**: JetBrains Mono, 14pt
- **Color Scheme**: Dracula or One Dark Pro
- **Background**: #1e1e1e (dark gray)
- **Opacity**: 95%
- **Padding**: 20px

## VS Code Settings:
```json
{
  "editor.fontSize": 16,
  "editor.fontFamily": "JetBrains Mono",
  "editor.lineHeight": 24,
  "workbench.colorTheme": "One Dark Pro",
  "editor.minimap.enabled": false,
  "editor.renderWhitespace": "none"
}
```

## Browser Setup:
- Clean bookmarks bar
- Dark theme enabled
- Zoom: 110%
- DevTools: Always docked to right

## Recording Checklist:
- [ ] Test audio levels (peak at -12dB to -6dB)
- [ ] Check lighting and framing
- [ ] Remove desktop clutter
- [ ] Close unnecessary applications
- [ ] Disable notifications
- [ ] Test recording before final take
EOF

echo "✅ Created recording environment files"
echo ""
echo "🎬 Video Recording Assets Created:"
echo "   - superagent-demo.sh: Installation and setup script"
echo "   - api-demo.sh: API demonstration script"
echo "   - dashboard.html: Monitoring dashboard example"
echo "   - terminal-theme.md: Terminal setup guidelines"
echo ""

echo "📁 Directory structure ready at: $(pwd)"
echo ""
echo "🎯 Next Steps for Recording:"
echo "1. Review VIDEO_SCRIPT_SUPERAGENT_5_MINUTES.md for script"
echo "2. Check VIDEO_PRODUCTION_SETUP_COMPLETE.md for equipment setup"
echo "3. Practice delivery with provided demo scripts"
echo "4. Record multiple takes, choose the best"
echo "5. Edit following post-production workflow"
echo ""
echo "🚀 Ready to create professional SuperAgent video content!"