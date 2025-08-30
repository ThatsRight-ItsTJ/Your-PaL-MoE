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

## Requirements

### System Requirements
- **Go**: Version 1.21 or higher (tested with Go 1.21+)
- **Operating System**: Linux, macOS, or Windows
- **Memory**: Minimum 512MB RAM
- **Network**: Internet connection for external AI providers

### Go Installation
If Go is not installed on your system:

#### Linux/macOS:
```bash
# Download and install Go 1.21+
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

#### Windows:
Download the installer from [https://golang.org/dl/](https://golang.org/dl/) and follow the installation wizard.

#### Verify Installation:
```bash
go version
# Should output: go version go1.21.x
```

## Installation

### 1. Clone the Repository
```bash
git clone https://github.com/ThatsRight-ItsTJ/Your-PaL-MoE.git
cd Your-PaL-MoE
```

### 2. Install Dependencies
```bash
# Download and install Go modules
go mod download
go mod tidy
```

### 3. Build the Enhanced Server
```bash
# Build the main enhanced server
go build -o enhanced-server cmd/enhanced-server/main.go

# Verify the build
ls -la enhanced-server
```

## Configuration

### 1. Create Provider Configuration
```bash
# Copy the template to create your providers configuration
cp providers.csv.template providers.csv
```

### 2. Edit Provider Settings
Edit `providers.csv` with your AI provider credentials:

**CSV Format (6 columns):**
```csv
Name,Tier,Base_URL,APIKey,Model(s),Other
OpenAI,official,https://api.openai.com/v1,sk-your-key-here,gpt-3.5-turbo|gpt-4|gpt-4-turbo,Premium service with high rate limits
Anthropic,official,https://api.anthropic.com/v1,your-api-key,claude-3-5-sonnet|claude-3-haiku,High quality responses
Local_Ollama,unofficial,http://localhost:11434,none,/api/tags,Local deployment with full privacy
```

**Column Descriptions:**
1. **Name**: Human-readable provider name  
2. **Tier**: `official`, `community`, or `unofficial`
3. **Base_URL**: API endpoint URL
4. **APIKey**: Authentication key (or "none" for no auth)
5. **Model(s)**: Can be a URL endpoint (e.g. /models) or a pipe-delimited list 
6. **Other**: Additional information (Rate Limits, descriptions, etc.)

### 3. Optional: Create Agents Configuration
```bash
# The system will auto-create agents.csv with defaults, or you can customize it
# agents.csv defines specialized agents for different task types
```

## Basic Usage

### 1. Start the Enhanced Server
```bash
# Start with default providers.csv
./enhanced-server

# Or specify a custom providers file
./enhanced-server custom-providers.csv
```

The server will start on port 8080 and display:
```
INFO[0000] Enhanced Your-PaL-MoE system initialized successfully
INFO[0000] Starting Enhanced Your PaL MoE server on :8080
```

### 2. Verify Server Health
```bash
# Check if the server is running
curl http://localhost:8080/health

# Expected response:
# {"status":"healthy","timestamp":1234567890,"version":"enhanced-1.0.0"}
```

### 3. Process Your First Request
```bash
# Simple request
curl -X POST http://localhost:8080/api/v1/process \
  -H "Content-Type: application/json" \
  -d '{
    "content": "Hello, how are you?",
    "context": {"domain": "general"}
  }'

# Complex analysis request
curl -X POST http://localhost:8080/api/v1/process \
  -H "Content-Type: application/json" \
  -d '{
    "content": "Analyze machine learning algorithms for recommendation systems",
    "context": {"domain": "ai", "output_format": "markdown"},
    "constraints": {"max_tokens": 1000, "quality_threshold": 0.8}
  }'
```

### 4. View Available Providers
```bash
# List all configured providers with metrics
curl http://localhost:8080/api/v1/providers | jq '.'
```

### 5. Generate YAML Configurations
```bash
# Generate YAML for a specific provider
curl http://localhost:8080/api/v1/providers/openai/yaml

# Generate YAML configs for all providers
curl -X POST http://localhost:8080/api/v1/providers/yaml/generate-all
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

# Server health check
GET /health
```

### Enhanced YAML Generation
```bash
# Generate YAML config for specific provider
GET /api/v1/providers/{id}/yaml

# Generate YAML configs for all providers
POST /api/v1/providers/yaml/generate-all
```

## Testing

### Run Built-in Tests
```bash
# Test dynamic YAML generation
./test_dynamic_yaml.sh

# Run end-to-end tests
./test_end_to_end_simple.sh

# Test multi-model functionality
./test_multi_model.sh
```

### Manual Testing
```bash
# Test different complexity levels
curl -X POST http://localhost:8080/api/v1/process -d '{"content": "Simple task"}'
curl -X POST http://localhost:8080/api/v1/process -d '{"content": "Complex analysis requiring deep reasoning"}'

# Monitor system metrics
curl http://localhost:8080/api/v1/metrics
```

## Troubleshooting

### Common Issues

#### 1. "go: command not found"
```bash
# Install Go following the Requirements section above
go version  # Should show go1.21.x or higher
```

#### 2. "providers.csv not found"
```bash
# Create from template
cp providers.csv.template providers.csv
# Edit with your API keys
```

#### 3. "failed to initialize selector"
```bash
# Check providers.csv format
head -5 providers.csv
# Ensure 6 columns: Name,Tier,Base_URL,APIKey,Model(s),Other
```

#### 4. Port 8080 already in use
```bash
# Check what's using the port
lsof -i :8080
# Kill the process or modify the server code to use a different port
```

#### 5. API key authentication errors
```bash
# Verify your API keys in providers.csv
# Check provider documentation for correct key format
# Test with curl directly to the provider's API
```

## Performance Metrics

### Projected Improvements
- **Task Success Rate**: 85% → 95% (+10%)
- **Execution Efficiency**: 70% → 90% (+20%)
- **Quality Score**: 75% → 88% (+13%)
- **Cost Savings**: Maintains 70-90% optimization

### Real-Time Monitoring
```bash
# Get comprehensive system metrics
curl http://localhost:8080/api/v1/metrics | jq '.'
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
