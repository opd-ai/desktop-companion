# ComfyUI GIF Asset Generation Pipeline Plan

## Overview

This document outlines the design and implementation plan for a Go-based pipeline that integrates with a local ComfyUI instance to automatically generate GIF assets for the desktop-companion virtual pet application.

## Pipeline Architecture

### 1. Core Components

#### ComfyUI Client (`lib/comfyui/`)
- **HTTP Client**: Interface with ComfyUI WebAPI for workflow submission
- **WebSocket Client**: Real-time monitoring of generation progress
- **Workflow Manager**: Template-based workflow creation and customization
- **Queue Manager**: Batch processing and job queue management

#### Asset Generator (`lib/assets/`)
- **Character Generator**: Base character image generation from descriptions
- **Variant Generator**: Mood/activity-specific image variations
- **Animation Processor**: Multi-frame animation sequence creation
- **GIF Optimizer**: Compression and optimization for desktop overlay use

#### Pipeline Controller (`cmd/gif-generator/`)
- **Batch Processor**: Orchestrates multi-character generation pipelines
- **Configuration Manager**: Character archetype and generation settings
- **Quality Validator**: Automated output validation and retry logic
- **File System Manager**: Asset organization and deployment

### 2. Generation Workflow

#### Phase 1: Base Character Generation
1. **Character Description Processing**
   - Parse character archetype from existing JSON configurations
   - Extract visual traits, personality markers, and style preferences
   - Generate ComfyUI prompts with consistent styling parameters

2. **Base Image Creation**
   - Submit base character workflow to ComfyUI
   - Generate 4-8 base poses/angles for animation consistency
   - Apply transparent background processing
   - Validate image quality and style consistency

#### Phase 2: State Variant Generation
1. **Animation State Mapping**
   ```go
   // Required animation states per character
   requiredStates := []string{
       "idle",     // Default desktop companion state
       "talking",  // Dialog/interaction state
       "happy",    // Positive mood/interaction
       "sad",      // Negative mood/error state
       "hungry",   // Game mechanic state
       "eating",   // Game interaction state
   }
   
   // Romance-specific states (for romance archetypes)
   romanceStates := []string{
       "shy",      // Tsundere/slow_burn personalities
       "flirty",   // Flirty/romance_flirty personalities
       "loving",   // Deep relationship states
       "jealous",  // Crisis/conflict states
   }
   
   // Game-specific states (for characters with game features)
   gameStates := []string{
       "sick",     // Health mechanic
       "tired",    // Energy mechanic
       "excited",  // Achievement/event state
   }
   ```

2. **Mood/Activity Variant Generation**
   - Apply state-specific prompt modifications
   - Maintain character consistency across variants
   - Generate multiple angle variations per state
   - Validate emotional expression accuracy

#### Phase 3: GIF Animation Creation
1. **Frame Sequence Generation**
   - Create 4-8 frame sequences for smooth looping
   - Apply micro-movements (breathing, blinking, idle motion)
   - Ensure seamless loop points
   - Optimize frame timing for natural animation

2. **GIF Processing Pipeline**
   ```go
   type GIFConfig struct {
       Width         int           // Target width (64-256px)
       Height        int           // Target height (64-256px)
       FrameCount    int           // 4-8 frames per animation
       FrameRate     int           // 10-15 FPS
       Colors        int           // Indexed color count (256 max)
       Transparency  bool          // Enable alpha channel
       Optimization  string        // "size" or "quality"
       MaxFileSize   int           // <500KB target
   }
   ```

3. **Quality Validation**
   - File size compliance (<500KB per GIF)
   - Transparency preservation
   - Loop continuity verification
   - Animation timing validation

## Implementation Plan

### Package Structure
```
lib/
├── comfyui/          # ComfyUI integration
│   ├── client.go     # HTTP/WebSocket client
│   ├── workflow.go   # Workflow management
│   ├── queue.go      # Job queue processing
│   └── monitor.go    # Progress monitoring
├── assets/           # Asset generation
│   ├── generator.go  # Character generation logic
│   ├── processor.go  # Image processing pipeline
│   ├── animator.go   # GIF animation creation
│   └── optimizer.go  # GIF optimization
└── pipeline/         # Pipeline orchestration
    ├── controller.go # Main pipeline controller
    ├── config.go     # Configuration management
    ├── validator.go  # Quality validation
    └── deployer.go   # Asset deployment

cmd/
└── gif-generator/    # CLI application
    ├── main.go       # Entry point
    ├── batch.go      # Batch processing commands
    ├── single.go     # Single character commands
    └── validate.go   # Validation commands
```

### Core Interfaces

#### ComfyUI Integration
```go
package comfyui

// Client interface for ComfyUI API interaction
type Client interface {
    // SubmitWorkflow submits a workflow for processing
    SubmitWorkflow(ctx context.Context, workflow *Workflow) (*Job, error)
    
    // MonitorJob tracks job progress via WebSocket
    MonitorJob(ctx context.Context, jobID string) (<-chan JobProgress, error)
    
    // GetResult retrieves completed job output
    GetResult(ctx context.Context, jobID string) (*JobResult, error)
    
    // GetQueueStatus returns current queue information
    GetQueueStatus(ctx context.Context) (*QueueStatus, error)
}

// Workflow represents a ComfyUI workflow configuration
type Workflow struct {
    ID          string                 `json:"id"`
    Nodes       map[string]interface{} `json:"nodes"`
    Connections []Connection          `json:"connections"`
    Metadata    WorkflowMetadata      `json:"metadata"`
}

// WorkflowTemplate for generating character variations
type WorkflowTemplate struct {
    BaseWorkflow  *Workflow
    PromptSlots   []PromptSlot
    Parameters    map[string]Parameter
    OutputNodes   []string
}
```

#### Asset Generation
```go
package assets

// Generator interface for character asset creation
type Generator interface {
    // GenerateBaseCharacter creates base character images
    GenerateBaseCharacter(ctx context.Context, req *CharacterRequest) (*GenerationResult, error)
    
    // GenerateVariants creates mood/activity variations
    GenerateVariants(ctx context.Context, baseResult *GenerationResult, states []string) (*VariantResult, error)
    
    // CreateAnimations generates GIF animations from image sequences
    CreateAnimations(ctx context.Context, variants *VariantResult) (*AnimationResult, error)
    
    // OptimizeAssets applies compression and validation
    OptimizeAssets(ctx context.Context, animations *AnimationResult) (*OptimizedResult, error)
}

// CharacterRequest defines character generation parameters
type CharacterRequest struct {
    Archetype     string            `json:"archetype"`     // e.g., "romance_tsundere"
    Description   string            `json:"description"`   // Character description
    Style         string            `json:"style"`         // Art style (pixel, anime, etc.)
    Traits        map[string]string `json:"traits"`        // Visual traits
    OutputConfig  *OutputConfig     `json:"output_config"` // Size, format settings
}
```

#### Pipeline Control
```go
package pipeline

// Controller orchestrates the complete generation pipeline
type Controller interface {
    // ProcessCharacter generates complete asset set for one character
    ProcessCharacter(ctx context.Context, config *CharacterConfig) (*ProcessResult, error)
    
    // ProcessBatch generates assets for multiple characters
    ProcessBatch(ctx context.Context, configs []*CharacterConfig) (*BatchResult, error)
    
    // ValidateAssets checks generated assets for compliance
    ValidateAssets(ctx context.Context, assetPath string) (*ValidationResult, error)
    
    // DeployAssets moves validated assets to target locations
    DeployAssets(ctx context.Context, result *ProcessResult) error
}

// CharacterConfig defines complete character processing configuration
type CharacterConfig struct {
    Character     *CharacterRequest     `json:"character"`
    States        []string             `json:"states"`        // Required animation states
    GIFConfig     *GIFConfig           `json:"gif_config"`    // GIF generation settings
    Validation    *ValidationConfig    `json:"validation"`    // Quality requirements
    Deployment    *DeploymentConfig    `json:"deployment"`    // Output configuration
}
```

### Implementation Phases

#### Phase 1: ComfyUI Integration (1-2 weeks)
1. **HTTP Client Implementation**
   - REST API client for workflow submission
   - Authentication and session management
   - Error handling and retry logic

2. **WebSocket Monitoring**
   - Real-time progress tracking
   - Job status updates
   - Result notification handling

3. **Workflow Management**
   - Template loading and customization
   - Dynamic prompt injection
   - Parameter validation

#### Phase 2: Asset Generation Pipeline (2-3 weeks)
1. **Character Generator**
   - Archetype-based prompt generation
   - Style consistency management
   - Multi-angle base generation

2. **Variant Generator**
   - State-specific modifications
   - Mood expression application
   - Activity pose generation

3. **Animation Processor**
   - Frame sequence creation
   - Motion interpolation
   - Loop optimization

#### Phase 3: GIF Processing & Optimization (1-2 weeks)
1. **GIF Creation**
   - Frame-to-GIF conversion
   - Transparency preservation
   - Color palette optimization

2. **Quality Validation**
   - File size compliance
   - Animation quality metrics
   - Loop continuity checks

3. **Batch Processing**
   - Multi-character workflows
   - Progress tracking
   - Error recovery

#### Phase 4: Integration & Deployment (1 week)
1. **CLI Application**
   - Command-line interface
   - Configuration management
   - Batch operation support

2. **Asset Deployment**
   - Automated file placement
   - Existing asset backup
   - Character.json updates

3. **Validation Tools**
   - Asset verification
   - Compatibility testing
   - Performance benchmarks

## Technical Requirements

### ComfyUI Integration Requirements
```go
// ComfyUI API Configuration
type ComfyUIConfig struct {
    ServerURL     string        `json:"server_url"`     // "http://localhost:8188"
    APIKey        string        `json:"api_key"`        // Optional authentication
    Timeout       time.Duration `json:"timeout"`        // Request timeout
    RetryAttempts int           `json:"retry_attempts"` // Failed request retries
    QueueLimit    int           `json:"queue_limit"`    // Max concurrent jobs
}

// Workflow Template Configuration
type WorkflowConfig struct {
    TemplatesPath string                    `json:"templates_path"` // Workflow JSON files
    Models        map[string]ModelConfig    `json:"models"`         // Model configurations
    Styles        map[string]StyleConfig    `json:"styles"`         // Style presets
    Quality       QualityConfig             `json:"quality"`        // Output quality settings
}
```

### Asset Specifications
```go
// GIF Output Requirements
type GIFRequirements struct {
    MaxWidth      int     `json:"max_width"`      // 256px maximum
    MaxHeight     int     `json:"max_height"`     // 256px maximum
    MinWidth      int     `json:"min_width"`      // 64px minimum
    MinHeight     int     `json:"min_height"`     // 64px minimum
    MaxFileSize   int     `json:"max_file_size"`  // 500KB maximum
    MaxFrames     int     `json:"max_frames"`     // 8 frames maximum
    MinFrames     int     `json:"min_frames"`     // 4 frames minimum
    FrameRate     int     `json:"frame_rate"`     // 10-15 FPS
    Transparency  bool    `json:"transparency"`   // Required for overlay
    LoopValidation bool   `json:"loop_validation"` // Seamless loop check
}

// Character Asset Requirements
type AssetRequirements struct {
    RequiredStates []string `json:"required_states"` // Core animation states
    OptionalStates []string `json:"optional_states"` // Archetype-specific states
    StyleConsistency bool   `json:"style_consistency"` // Cross-state consistency
    ArchetypeCompliance bool `json:"archetype_compliance"` // Personality accuracy
}
```

### Error Handling Strategy
```go
// Error types for pipeline processing
type PipelineError struct {
    Stage     string    `json:"stage"`     // Generation stage
    Type      string    `json:"type"`      // Error category
    Message   string    `json:"message"`   // Human-readable description
    Retryable bool      `json:"retryable"` // Whether retry is possible
    Timestamp time.Time `json:"timestamp"` // Error occurrence time
}

// Recovery strategies for different error types
type RecoveryStrategy struct {
    MaxRetries    int           `json:"max_retries"`    // Maximum retry attempts
    BackoffDelay  time.Duration `json:"backoff_delay"`  // Delay between retries
    FallbackMode  string        `json:"fallback_mode"`  // Fallback generation mode
    SkipOnFailure bool          `json:"skip_on_failure"` // Continue with other assets
}
```

## Integration with Existing System

### Character System Integration
The pipeline will integrate with the existing character system through:

1. **Configuration Integration**
   - Read existing character.json files for archetype information
   - Extract personality traits and visual preferences
   - Preserve existing animation mappings

2. **Asset Replacement**
   - Backup existing GIF files before replacement
   - Maintain file naming conventions
   - Update character.json with new asset metadata

3. **Validation Integration**
   - Use existing character validation system
   - Ensure compatibility with animation manager
   - Test with existing UI components

### File System Integration
```bash
# Generated asset structure (matches existing layout)
assets/characters/{archetype}/animations/
├── idle.gif           # Base idle animation
├── talking.gif        # Dialog state animation
├── happy.gif          # Positive mood animation
├── sad.gif            # Negative mood animation
├── hungry.gif         # Game state animation (if applicable)
├── eating.gif         # Game interaction animation (if applicable)
└── generated/         # Pipeline metadata and backups
    ├── metadata.json  # Generation parameters and history
    ├── backups/       # Previous versions
    └── frames/        # Individual frame sources
```

## Performance Considerations

### Concurrent Processing
- **Parallel Character Generation**: Process multiple characters simultaneously
- **Batch State Generation**: Generate all states for a character in parallel
- **Frame Processing**: Concurrent frame generation and optimization
- **Resource Management**: CPU/GPU utilization monitoring and throttling

### Memory Management
```go
// Resource monitoring for pipeline processes
type ResourceMonitor struct {
    MaxMemoryUsage int64         `json:"max_memory_mb"`    // Memory limit
    GPUMemoryLimit int64         `json:"gpu_memory_mb"`    // VRAM limit
    ConcurrentJobs int           `json:"concurrent_jobs"`  // Parallel job limit
    TempDirCleanup time.Duration `json:"cleanup_interval"` // Temp file cleanup
}
```

### Optimization Strategies
- **Template Caching**: Cache ComfyUI workflow templates
- **Asset Streaming**: Stream large assets instead of loading fully into memory
- **Progressive Quality**: Generate low-quality previews first, then high-quality finals
- **Incremental Processing**: Resume interrupted batch operations

## Testing Strategy

### Unit Testing
- ComfyUI client API interactions
- Asset generation algorithms
- GIF processing and optimization
- Configuration validation

### Integration Testing
- End-to-end pipeline execution
- Character system compatibility
- Asset deployment verification
- Performance benchmarking

### Validation Testing
```go
// Asset validation test suite
type ValidationSuite struct {
    FileFormatTests     []FileFormatTest     `json:"file_format_tests"`
    AnimationTests      []AnimationTest      `json:"animation_tests"`
    QualityTests        []QualityTest        `json:"quality_tests"`
    CompatibilityTests  []CompatibilityTest  `json:"compatibility_tests"`
}
```

## Deployment and Operations

### CLI Commands
```bash
# Generate assets for single character
gif-generator character --archetype romance_tsundere --style anime

# Batch generate for all archetypes
gif-generator batch --config batch_config.json --parallel 4

# Validate existing assets
gif-generator validate --path assets/characters/

# Deploy generated assets
gif-generator deploy --source generated/ --target assets/characters/
```

### Configuration Management
```json
{
  "comfyui": {
    "server_url": "http://localhost:8188",
    "timeout": "300s",
    "retry_attempts": 3
  },
  "generation": {
    "default_style": "pixel_art",
    "base_resolution": [128, 128],
    "frame_count": 6,
    "animation_duration": "1s"
  },
  "quality": {
    "max_file_size": 500000,
    "min_frame_rate": 10,
    "transparency_required": true
  },
  "deployment": {
    "backup_existing": true,
    "validate_before_deploy": true,
    "update_character_json": true
  }
}
```

## Future Enhancements

### Advanced Features
1. **Style Transfer**: Apply consistent art styles across character sets
2. **Facial Expression AI**: Automated emotion-appropriate facial expressions
3. **Pose Variation**: Dynamic pose generation for enhanced animation variety
4. **Quality Enhancement**: AI upscaling and detail enhancement
5. **Custom Model Training**: Character-specific model fine-tuning

### Workflow Extensions
1. **Interactive Preview**: Real-time generation preview and adjustment
2. **Version Control**: Asset versioning and rollback capabilities
3. **A/B Testing**: Generate multiple variants for quality comparison
4. **Community Integration**: Share and download community-created assets

### Performance Optimizations
1. **Distributed Processing**: Multi-machine ComfyUI clusters
2. **Caching System**: Intelligent asset caching and reuse
3. **Progressive Generation**: Adaptive quality based on use case
4. **Resource Prediction**: ML-based resource requirement prediction

## Conclusion

This pipeline design provides a comprehensive solution for automated GIF asset generation that integrates seamlessly with the existing desktop-companion architecture. The modular design ensures maintainability while the interface-based approach supports future extensions and improvements.

The implementation follows the project's "lazy programmer" philosophy by leveraging ComfyUI's existing capabilities while writing minimal glue code for integration. The pipeline supports all 19 character archetypes and maintains compatibility with the existing character system.

By automating the asset generation process, this pipeline will significantly reduce the manual effort required to create and maintain character animations while ensuring consistent quality and style across the entire character library.