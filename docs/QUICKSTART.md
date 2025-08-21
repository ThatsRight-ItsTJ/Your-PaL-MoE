# Intelligent AI Gateway - Quick Start Guide

## Prerequisites
- Docker and Docker Compose installed
- API keys for any official providers you want to use

## Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/intelligent-ai-gateway.git
cd intelligent-ai-gateway
```

2. Copy the CSV template:
```bash
cp providers.csv.template providers.csv
```

3. Edit providers.csv with your providers:
```csv
Name,Tier,Endpoint,Model(s)
OpenAI,official,https://api.openai.com/v1,https://api.openai.com/v1/models
Pollinations,community,https://text.pollinations.ai,https://text.pollinations.ai/models
```

4. Set environment variables:
```bash
cp .env.example .env
# Edit .env with your API keys
```

5. Start the gateway:
```bash
docker-compose up -d
```

## Adding Providers

### Via CSV (Recommended)

1. Edit providers.csv
2. Add a new line with: Name,Tier,Endpoint,Model(s)
3. Save the file
4. The gateway will auto-reload

### Via Admin UI

1. Navigate to http://localhost:3000/admin/ui
2. Click on "CSV Config" tab
3. Edit the CSV content
4. Click "Save & Reload"

## Testing Your Setup

### Test a provider:
```bash
curl -X POST http://localhost:3000/v1/chat/completions \
  -H "Authorization: Bearer your-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

### Check provider health:
```bash
curl http://localhost:3000/admin/providers \
  -H "Authorization: Bearer admin-key"
```

## Unofficial API Setup

For unofficial APIs, create a script in the scripts/ directory:

1. Create your script (e.g., scripts/my-api.py)
2. Add to CSV: MyAPI,unofficial,./scripts/my-api.py,my-model
3. The gateway will generate a template if the script doesn't exist
4. Edit the template to add your API integration

## Cost Optimization

The gateway automatically optimizes costs by:

- Routing to free providers first (Pollinations, unofficial APIs)
- Falling back to paid providers only when needed
- Using parallel execution for faster responses
- Caching common requests

To adjust optimization:
```json
{
  "cost_optimization": "aggressive",  // or "balanced", "quality"
  "provider_preference": ["unofficial", "community", "official"]
}
```

## Monitoring

Access the admin dashboard at: http://localhost:3000/admin/ui

Features:

- Real-time cost tracking
- Provider health monitoring
- Performance analytics
- Optimization recommendations

## Troubleshooting

### Provider not working?

- Check health status in admin UI
- Test the provider: /admin/providers/{name}/test
- Check logs: docker-compose logs gateway

### High costs?

- Enable aggressive cost optimization
- Add more community/unofficial providers
- Check optimization recommendations in admin UI

### Need help?

- Documentation: https://github.com/yourusername/intelligent-ai-gateway/docs
- Issues: https://github.com/yourusername/intelligent-ai-gateway/issues