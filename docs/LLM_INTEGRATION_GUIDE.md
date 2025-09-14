# LLM Dialog Backend Integration

This document describes the integration of optional LLM (Large Language Model) dialog capabilities into the Desktop Companion application using the miniLM service.

## Overview

The LLM dialog backend provides AI-powered conversational abilities while maintaining full backward compatibility with existing character configurations. The integration follows the "lazy programmer" philosophy by leveraging the existing miniLM service rather than implementing LLM functionality from scratch.

## Key Features

- **Optional Integration**: LLM features are completely optional and can be disabled
- **Automatic Fallback**: Automatic fallback to Markov chains when LLM backend fails  
- **Personality-Aware**: Extracts personality from character traits and Markov training data
- **Context-Aware**: Uses character mood, relationship level, and situation context
- **Backward Compatible**: Existing character configurations work unchanged
- **Error Resilient**: Comprehensive error handling with graceful degradation

## Architecture

### Components

1. **LLMDialogBackend**: Adapter implementing the `DialogBackend` interface
2. **PersonalityExtractor**: Converts character data to LLM-compatible prompts
3. **Backend Registry**: Integration with existing dialog system registration
4. **Configuration Utilities**: Tools for adding LLM config to existing characters

### Integration Flow

```
Character Configuration → Dialog Manager → LLM Backend → miniLM Service
                                      ↓ (on failure)
                                   Fallback Chain → Markov/Simple Random
```

## Configuration

### Character JSON Schema Extension

Add an optional `dialogBackend` section to character JSON files:

```json
{
  "dialogBackend": {
    "enabled": true,
    "defaultBackend": "llm",
    "fallbackChain": ["markov_chain", "simple_random"],
    "backends": {
      "llm": {
        "modelPath": "models/companion-chat.gguf",
        "maxTokens": 50,
        "temperature": 0.8,
        "personalityWeight": 1.5,
        "moodInfluence": 1.0,
        "enabled": true,
        "mockMode": true
      }
    }
  }
}
```

### LLM Backend Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | `false` | Master enable/disable switch |
| `mockMode` | bool | `true` | Use mock responses for development |
| `modelPath` | string | `""` | Path to GGUF model file |
| `maxTokens` | int | `50` | Maximum tokens to generate |
| `temperature` | float64 | `0.8` | Response randomness (0.0-2.0) |
| `personalityWeight` | float64 | `1.0` | Personality influence (0.0-2.0) |
| `moodInfluence` | float64 | `0.7` | Mood influence (0.0-2.0) |
| `systemPrompt` | string | `""` | Base system prompt template |
| `personalityPrompt` | string | `""` | Additional personality context |
| `maxGenerationTime` | int | `30` | Timeout in seconds |

## Usage

### Enabling LLM for Existing Characters

Use the provided configuration generator:

```bash
go run tools/llm-config-generator/main.go \
  -input assets/characters \
  -archetype romance \
  -enable \
  -mock
```

This adds LLM configuration to existing character files while preserving all existing functionality.

### Character Archetypes

The system provides pre-configured archetypes:

- **default**: Balanced companion personality
- **romance**: Emotionally expressive with romantic undertones  
- **tsundere**: Defensive personality hiding affection
- **flirty**: Playful and charming interactions

### Manual Configuration

1. **Add Dialog Backend Section**: Include `dialogBackend` in character JSON
2. **Configure LLM Backend**: Set model path and generation parameters
3. **Set Fallback Chain**: Configure fallback to existing backends
4. **Test Integration**: Verify functionality with mock mode

## Implementation Details

### Personality Extraction

The system automatically extracts personality information from:

- **Character Traits**: Numerical personality scores
- **Markov Training Data**: Analysis of existing dialog patterns  
- **Character Description**: Natural language personality descriptions

### Error Handling

Comprehensive error handling ensures system stability:

1. **Initialization Errors**: LLM failures disable backend without affecting app
2. **Generation Timeouts**: Automatic fallback after configurable timeout
3. **Quality Validation**: Low-quality responses trigger fallback
4. **Health Monitoring**: Continuous health checking with recovery attempts

### Fallback Mechanisms

Three levels of fallback ensure responses are always available:

1. **LLM Backend**: Primary AI-powered responses
2. **Markov Chain**: Statistical text generation from training data
3. **Simple Random**: Basic random selection from predefined responses

## Development

### Testing

Run comprehensive tests:

```bash
# Unit tests for LLM backend
go test ./lib/dialog/ -run TestLLMDialogBackend -v

# Integration tests for full system
go test ./lib/dialog/ -run TestLLMDialogSystem -v

# Backward compatibility tests
go test ./lib/dialog/ -run BackwardCompatibility -v
```

### Mock Mode

For development without model files:

```json
{
  "llm": {
    "enabled": true,
    "mockMode": true,
    "debug": true
  }
}
```

Mock mode provides realistic response generation for testing dialog flows.

### Debug Mode

Enable detailed logging:

```json
{
  "dialogBackend": {
    "debugMode": true,
    "backends": {
      "llm": {
        "debug": true
      }
    }
  }
}
```

## Production Deployment

### Model Requirements

- **Format**: GGUF model files
- **Size**: Recommend models under 1GB for desktop deployment
- **Type**: Chat/instruction-tuned models work best
- **Quantization**: 4-bit or 8-bit quantization for performance

### Performance Optimization

- **Concurrent Requests**: Limit via `concurrentRequests` setting
- **Context Management**: Use appropriate `contextSize` for memory usage
- **Generation Timeout**: Set reasonable `maxGenerationTime` for UX
- **Health Checking**: Configure `healthCheckInterval` for monitoring

### Security Considerations

- **Model Validation**: Verify model integrity before loading
- **Input Sanitization**: miniLM handles input validation
- **Resource Limits**: Enforce memory and CPU limits
- **Network Security**: No network access required for local models

## Migration Guide

### From Markov-Only Characters

1. **Backup Existing Config**: Create backups before modification
2. **Run Config Generator**: Use provided tool with appropriate archetype
3. **Test Functionality**: Verify both LLM and fallback work correctly
4. **Gradual Rollout**: Enable LLM for subset of characters initially

### From External LLM Services

1. **Model Conversion**: Convert models to GGUF format if needed
2. **Prompt Translation**: Adapt existing prompts to new format
3. **Parameter Tuning**: Adjust temperature and token limits
4. **Fallback Configuration**: Set up Markov chains as backup

## Troubleshooting

### Common Issues

**LLM Backend Not Responding**
- Check model path exists and is readable
- Verify sufficient system memory
- Enable debug mode for detailed logging
- Test with mock mode first

**Poor Response Quality**
- Adjust temperature and token settings
- Improve personality prompts
- Enhance Markov training data for fallbacks
- Verify model is appropriate for chat/dialog

**Performance Issues**
- Reduce concurrent request limits
- Optimize model size (use quantized models)
- Adjust generation timeout
- Monitor system resource usage

### Log Analysis

Key log messages to monitor:

```
INFO  LLM dialog backend initialized successfully
DEBUG LLM response generated successfully
WARN  LLM backend health check failed, triggering fallback
ERROR Critical LLM error detected, disabling backend
```

## API Reference

### LLMDialogBackend Methods

```go
// Initialize configures the backend with JSON configuration
func (llm *LLMDialogBackend) Initialize(config json.RawMessage) error

// GenerateResponse produces AI-powered dialog responses
func (llm *LLMDialogBackend) GenerateResponse(context DialogContext) (DialogResponse, error)

// CanHandle checks if backend can process the request
func (llm *LLMDialogBackend) CanHandle(context DialogContext) bool

// UpdateMemory records interaction outcomes for learning
func (llm *LLMDialogBackend) UpdateMemory(context DialogContext, response DialogResponse, feedback *UserFeedback) error

// IsHealthy checks backend health status
func (llm *LLMDialogBackend) IsHealthy() bool

// GetModelInfo returns model and configuration information
func (llm *LLMDialogBackend) GetModelInfo() map[string]interface{}
```

### PersonalityExtractor Methods

```go
// ExtractFromTrainingData analyzes Markov data for personality
func (pe *PersonalityExtractor) ExtractFromTrainingData(trainingData []string) PersonalityPrompt

// ExtractFromTraits converts character traits to prompts
func (pe *PersonalityExtractor) ExtractFromTraits(traits map[string]float64) PersonalityPrompt

// CombinePrompts merges multiple personality sources
func (pe *PersonalityExtractor) CombinePrompts(prompts ...PersonalityPrompt) PersonalityPrompt
```

## License and Dependencies

- **Desktop Companion**: MIT License
- **miniLM Service**: Compatible with existing project licensing
- **Go Dependencies**: All use permissive licenses (MIT, Apache 2.0, BSD-3-Clause)

No additional license requirements introduced by LLM integration.

## Support

For issues related to LLM integration:

1. **Check Configuration**: Verify JSON syntax and required fields
2. **Test with Mock Mode**: Isolate configuration vs. model issues  
3. **Review Logs**: Enable debug mode for detailed diagnostics
4. **Fallback Testing**: Ensure Markov chains work independently
5. **Performance Monitoring**: Check resource usage and timing

The integration is designed to fail gracefully - if LLM features don't work, the companion should continue functioning with existing dialog systems.