# Your PaL MoE (Mixture of Experts)

A cost-optimized AI router that automatically selects the best provider for each task using a simple CSV-based configuration system with enhanced intelligence features.

## Features

### 🎯 **Original Core Features**
- **Multi-tier Provider System**: Official → Community → Unofficial fallback routing
- **Cost Optimization**: Achieve 70-90% cost savings through intelligent provider selection  
- **Go Concurrency**: True parallel processing with goroutines and channels
- **Simple CSV Configuration**: Easy provider management via `providers.csv`
- **RESTful API**: Clean HTTP endpoints for integration

### 🧠 **Enhanced Intelligence Features**
- **Task Complexity Analysis**: Multi-dimensional scoring (reasoning, knowledge, computation, coordination)
- **Self-Supervised Prompt Optimization (SPO)**: Automatic prompt enhancement with caching
- **Adaptive Provider Selection**: ML-inspired multi-criteria scoring with real-time learning
- **Performance Tracking**: Continuous metrics collection and provider health monitoring
- **AI-Powered YAML Generation**: Automated provider configuration using Pollinations API

## Quick Start

### Installation
```bash
git clone https://github.com/yourusername/Your-PaL-MoE.git
cd Your-PaL-MoE
go mod tidy
```

### Configuration

Create `providers.csv` with your AI providers:
```bash
cp providers.csv.template providers.csv
```

**CSV Format (6 columns):**
1. **Name**: Human-readable provider name  
2. **Tier**: `official`, `community`, or `unofficial`
4. **Endpoint**: API endpoint URL
5. **APIKey**: Authentication key (or "none" for no auth)
6. **Model**: Can be a url endpoint (e.g. /models) or a delimited list 
7. **Other**: Any other relevant information (Rate Limits, etc.)
   
### Basic Usage

```bash
# Start the enhanced server
go build -o enhanced-server cmd/enhanced-server/main.go
./enhanced-server

# Process a request
curl -X POST http://localhost:8080/api/v1/process \
  -H "Content-Type: application/json" \
  -d '{
    "content": "Analyze machine learning algorithms for recommendation systems",
    "context": {"domain": "ai", "output_format": "markdown"},
    "constraints": {"max_tokens": 1000}
  }'
```

## Enhanced Pipeline Flow

```
📝 Initial Prompt → 🧠 SPO Analysis → 🎯 Provider Selection → 🚀 Task Execution
       ↓                ↓                    ↓                    ↓
   Complexity      Optimization        Intelligent           Parallel
    Analysis        & Caching          Selection             Processing
```

### Step-by-Step Process

1. **Task Reasoning**: Analyzes complexity across 4 dimensions (reasoning, knowledge, computation, coordination)
2. **SPO Optimization**: Enhances prompts using 5 strategies (clarification, structure, context, constraints, examples)
3. **Provider Selection**: Multi-criteria scoring considering cost, performance, latency, and reliability
4. **Execution**: Parallel processing with performance monitoring and learning

## API Endpoints

### Core Endpoints
```bash
# Process a request with enhanced pipeline
POST /api/v1/process

# Get request status and results
GET /api/v1/requests/{id}

# List all providers with metrics
GET /api/v1/providers

# Get system performance metrics
GET /api/v1/metrics
```

### Enhanced YAML Generation
```bash
# Generate YAML config for specific provider
GET /api/v1/providers/{id}/yaml

# Generate YAML configs for all providers
POST /api/v1/providers/yaml/generate-all
```

### Example Response
```json
{
  "id": "req_1234567890",
  "complexity": {
    "reasoning": 2, "knowledge": 3, "computation": 1, "coordination": 1,
    "overall": 2, "score": 0.58
  },
  "optimized_prompt": {
    "original": "Analyze machine learning algorithms",
    "optimized": "Analyze machine learning algorithms\n\nPlease be specific and detailed...",
    "improvements": ["Added detailed guidance", "Enhanced clarity"],
    "confidence": 0.85, "cost_savings": 0.255
  },
  "assignment": {
    "provider_id": "openai_gpt4", "confidence": 0.92,
    "estimated_cost": 0.045, "reasoning": "Selected for high complexity analysis"
  },
  "status": "completed", "total_cost": 0.042, "total_duration": "2.3s"
}
```

## AI-Powered YAML Generation

The enhanced system can automatically generate YAML configurations from CSV entries using the **Pollinations API** (no-auth required):

### Automatic Generation
```bash
# Generate YAML for a specific provider
curl http://localhost:8080/api/v1/providers/openai_gpt4/yaml

# Batch generate all provider YAMLs  
curl -X POST http://localhost:8080/api/v1/providers/yaml/generate-all
```

### Example Generated YAML
```yaml
# OpenAI GPT-4 Provider Configuration
provider:
  id: openai_gpt4
  name: "OpenAI GPT-4"
  tier: official
  
api:
  endpoint: "https://api.openai.com/v1"
  authentication:
    type: "bearer_token"
    key: "${OPENAI_API_KEY}"
    
model:
  name: "gpt-4"
  temperature: 0.7
  max_tokens: 8192
  
rate_limiting:
  requests_per_minute: 10000
  tier: "premium"
  
cost_tracking:
  cost_per_token: 0.00003
  billing_model: "per_token"
  
retry_config:
  max_retries: 3
  backoff_multiplier: 2
  timeout_seconds: 30
```

## Performance Metrics

### Projected Improvements
- **Task Success Rate**: 85% → 95% (+10%)
- **Execution Efficiency**: 70% → 90% (+20%)
- **Quality Score**: 75% → 88% (+13%)
- **Cost Savings**: Maintains 70-90% optimization

### Real-Time Monitoring
```bash
# Get system metrics
curl http://localhost:8080/api/v1/metrics
```

Returns:
- Total/successful/failed requests
- Average response time and cost savings
- Provider health scores and performance trends
- Active request count and system status

## Architecture

### Core Components
- **TaskReasoningEngine**: Rule-based complexity analysis using regex patterns (no local AI models)
- **SPOOptimizer**: Self-supervised prompt optimization with LRU caching
- **AdaptiveProviderSelector**: Multi-criteria provider scoring with real-time adaptation
- **YAMLGenerator**: AI-powered configuration generation via Pollinations API
- **EnhancedSystem**: Main orchestrator with metrics tracking and performance monitoring

### Technology Stack
- **Language**: Go 1.21+
- **Concurrency**: Goroutines and channels for parallel processing
- **HTTP Framework**: Gorilla Mux for REST API
- **AI Integration**: Pollinations API for YAML generation
- **Configuration**: CSV-based provider management
- **Logging**: Structured logging with Logrus

## Cost Optimization Strategy

### Multi-Tier Routing
1. **Official Providers** (OpenAI, Anthropic): High-quality, higher cost
2. **Community Providers** (Pollinations, Together): Balanced quality/cost
3. **Unofficial/Local**: Lowest cost, variable quality

### Intelligent Selection
- **Complexity Matching**: Route complex tasks to higher-tier providers
- **Cost Efficiency**: Balance quality requirements with budget constraints  
- **Performance Learning**: Adapt selection based on historical performance
- **Fallback Strategy**: Automatic failover to alternative providers

## Use Cases

### Development & Prototyping
- **Local Testing**: Use unofficial/local providers for development
- **Cost Control**: Automatic cost-aware provider selection
- **Quality Assurance**: Multi-tier validation for production workloads

### Production Deployment
- **High Availability**: Multi-provider redundancy with automatic failover
- **Performance Optimization**: ML-inspired provider selection
- **Cost Management**: Achieve 70-90% cost savings while maintaining quality

### Enterprise Integration
- **Scalable Architecture**: Go concurrency handles high request volumes
- **Monitoring & Analytics**: Comprehensive metrics and performance tracking
- **Configuration Management**: Simple CSV-based provider administration

## Contributing

We welcome contributions! The enhanced system maintains the modular architecture of the original while adding intelligent capabilities.

### Development Guidelines
- Maintain cost optimization principles
- Preserve Go concurrency advantages
- Extend CSV compatibility for easy provider management
- Focus on intelligence without complexity

### Key Areas for Contribution
1. **Advanced Task Decomposition**: Operator-based workflow systems
2. **Quality Assurance**: Multi-tier validation frameworks  
3. **Machine Learning**: Predictive provider selection models
4. **Real-time Adaptation**: Dynamic provider performance tuning

## License

MIT License - see [LICENSE](LICENSE) for details.

## Support

- **Documentation**: See `/docs` for detailed guides
- **Issues**: Report bugs and feature requests via GitHub Issues
- **Discussions**: Join community discussions for questions and ideas

---

**Your PaL MoE**: Cost-optimized AI routing with enhanced intelligence, maintaining simplicity while adding powerful features for task complexity analysis, self-supervised optimization, and adaptive provider selection.
