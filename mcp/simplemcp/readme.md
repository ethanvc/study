本地mcp，使用stdio通信。
```json
{
  "mcpServers": {
    "weather": {
      "command": "uv",
      "args": [
        "--directory",
        "/Users/ethan/Downloads/weather",
        "run",
        "weather.py"
      ]
    }
  }
}
```

remote mcp：
```json
{
    "github": {
      "httpUrl": "https://api.githubcopilot.com/mcp",
      "headers": {
              "Authorization": "<YOUR_GITHUB_PAT>"
            },
            "timeout": 5000
     }
}
```