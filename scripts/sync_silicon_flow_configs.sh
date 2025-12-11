#!/bin/bash

# Sync SiliconFlow config between Crush and OpenCode

API_KEY=$(grep -o '"api_key": "[^"]*' ~/.config/crush/crush.json | head -1 | cut -d'"' -f4)

if [ -n "$API_KEY" ]; then
    
    # Update OpenCode config with same key
    sed -i "s/\"apiKey\": \".*\"/\"apiKey\": \"$API_KEY\"/g" ~/.config/opencode/config.json
    echo "✅ Synced API key to OpenCode"
    
    # Also update base URL
    sed -i "s|https://api\..*|https://api.siliconflow.com/v1|g" ~/.config/opencode/config.json
    echo "✅ Updated base URL"
    
else

    echo "⚠️  No API key found in Crush config"
fi

