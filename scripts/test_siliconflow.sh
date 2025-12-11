curl --request POST \
  --url https://api.siliconflow.com/v1/chat/completions \
  --header 'Authorization: Bearer API_KEY_GOES_HERE' \
  --header 'Content-Type: application/json' \
  --data '{
    "model": "Qwen/Qwen2.5-7B-Instruct",
    "messages": [
      {
        "role": "user",
        "content": "Say hello."
      }
    ],
    "stream": false
  }'
