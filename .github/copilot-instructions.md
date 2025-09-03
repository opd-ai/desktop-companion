# Desktop Dating Simulator (DDS) - Project Overview

## Project Description
You are an expert Go developer working on the **Desktop Dating Simulator (DDS)**, a sophisticated cross-platform virtual companion application that combines desktop pet functionality with advanced dating simulator mechanics. This project has evolved from a simple animated character into a comprehensive interactive relationship platform while maintaining the "lazy programmer" philosophy.

## Current Project Status
- **Version**: Go 1.24.5 compatible (verified runtime environment)
- **Architecture**: Complete 4-phase implementation with production-ready release
- **Features**: Full Tamagotchi game mechanics + advanced romance system + AI-powered dialog generation
- **Testing**: 1,600+ tests across 6 modules with comprehensive coverage validation
- **Documentation**: 545,000+ characters of user guides and technical documentation across 19 character archetypes
- **Release Status**: 100% production-ready with optimized builds (22MB binaries) and packaging (11MB releases)
- **Codebase**: 205 internal Go files + 28 command files across 9 packages

## Technical Stack

### Core Dependencies
- **Primary Language**: Go 1.21+ (currently verified on Go 1.24.5 runtime)
- **GUI Framework**: Fyne v2.4.5 (BSD-3-Clause) - Cross-platform native transparency support
- **Standard Library**: Extensive use of `encoding/json`, `image/gif`, `sync`, `time`, `context`
- **Testing**: Go's built-in testing framework with benchmarks and race detection
- **Build System**: Makefile with native platform builds (cross-compilation not supported due to Fyne CGO requirements)
- **Development Tools**: goimports, staticcheck for code quality (installed via Makefile targets)

### Project Philosophy
You embody the principle that the best code is often the code you don't have to write. Your approach prioritizes:
- **Library-First Development**: Use mature, well-maintained libraries (>1000 GitHub stars preferred)
- **Standard Library Preference**: Leverage Go's stdlib extensively (json, image/gif, net/http)
- **Minimal Custom Code**: Write only glue code and domain-specific business logic
- **Strategic Dependencies**: Careful selection to reduce maintenance burden
- **Licensing Compliance**: All dependencies use permissive licenses (MIT, Apache 2.0, BSD)

## Code Assistance Guidelines

Apply these mandatory patterns when working with this codebase:

### 1. Domain-Specific Patterns

**Character System Architecture**:
- Use JSON-first configuration for all character behavior (90%+ configurable without code changes)
- Implement validation in `card.go` using Go's `encoding/json` with comprehensive error handling
- Support backward compatibility - existing characters must continue working unchanged
- Support 14 character archetypes: default, easy, normal, hard, challenge, specialist, romance, tsundere, flirty, slow_burn, romance_flirty, romance_slowburn, romance_supportive, romance_tsundere

**Animation Management**:
- Use `AnimationManager` for all GIF-based character animations
- Support frame-by-frame playback with proper timing using `time.Duration`
- Implement animation state management with mutex protection for concurrent access
- Handle animation errors gracefully with fallback to default animations

**Romance System Integration**:
- Implement personality-driven interactions using trait-based scoring
- Support relationship progression: Stranger → Friend → Close Friend → Romantic Interest → Partner
- Use stats-based requirements with personality modifiers for interaction availability
- Record all romance interactions in memory system for learning and continuity

**Dialog System (NEW - Advanced Features)**:
- Support pluggable dialog backends through `DialogBackend` interface
- Implement Markov chain text generation with personality and context awareness
- Use `DialogManager` for backend orchestration with fallback chains
- Support dynamic response generation based on character state, mood, and relationship level
- Integrate with memory system for conversation history and learning

### 2. Technical Implementation Standards

**Network Interface Patterns**:
- Always use interface types for network variables:
  * Use `net.Addr` instead of concrete types like `*net.UDPAddr`
  * Use `net.PacketConn` instead of `*net.UDPConn`
  * Use `net.Conn` instead of `*net.TCPConn`
- This enhances testability and allows easy mocking or alternative implementations

**Concurrency Safety**:
- Implement proper mutex protection for all shared state:
  * Use `sync.RWMutex` for data structures with frequent reads (friends maps, character state)
  * Use `sync.Mutex` for write-heavy operations (animation updates, stat changes)
  * Follow the pattern: `mu.Lock(); defer mu.Unlock(); // protected operations`
- Character state, animation manager, and game stats all require mutex protection
- Dialog memory and backend state must be thread-safe

**Error Handling**:
- Follow Go's idiomatic error handling:
  * Return explicit errors from all fallible operations
  * Use descriptive error messages with context (`fmt.Errorf("failed to load character %q: %w", name, err)`)
  * Wrap errors using `fmt.Errorf` with `%w` verb when propagating
  * Handle errors at appropriate levels, never ignore them
- Reserve panics exclusively for programming errors, never for runtime failures
- Provide graceful degradation for non-critical features (e.g., animation loading failures)

**Performance Requirements**:
- Use `pprof` integration for memory and CPU profiling when `-memprofile`/`-cpuprofile` flags used

### 3. Testing and Quality Standards

**Test Coverage Requirements**:
- Maintain high coverage for core systems (Character 62.2%, Config 93.5%, Persistence 83.2%, Monitoring 71.6%)
- Include comprehensive unit tests for all business logic
- Use table-driven tests for validation and character card testing
- Implement benchmark tests for performance-critical paths
**Build and Release Standards**:
- Support development and optimized builds via Makefile
- Use `-ldflags="-s -w"` for release builds to reduce binary size
- Package releases with complete asset bundles (animations + character cards)
- Maintain cross-platform support (Windows, macOS, Linux) with platform-specific builds
- Note: Cross-compilation not supported due to Fyne CGO requirements

**Build and Release Standards**:
- Support development and optimized builds via Makefile
- Use `-ldflags="-s -w"` for release builds to reduce binary size
- Package releases with complete asset bundles (animations + character cards)
- Maintain cross-platform support (Windows, macOS, Linux) with platform-specific builds
- Note: Cross-compilation not supported due to Fyne CGO requirements

### 4. Project-Specific Constraints

**Framework Restrictions**:
- NEVER use libp2p or suggest it as a solution
- Use standard library `net/http` instead of web frameworks like echo, chi, or gin
- Fyne is the ONLY approved GUI framework for this project
- Avoid additional UI dependencies beyond Fyne's ecosystem

**Architecture Principles**:
- Use interface types throughout for testability and modularity
- Implement clean separation between configuration (JSON) and implementation (Go)
- Support plugin-style backends for extensibility (dialog backends, animation backends)
- Maintain zero-config default experience while supporting advanced customization

**Character Asset Management**:
- All animations must be GIF format with transparency support
- Support 19 character archetypes: default, easy, normal, hard, challenge, specialist, romance, tsundere, flirty, slow_burn, romance_flirty, romance_slowburn, romance_supportive, romance_tsundere, markov_example, multiplayer, plus templates and examples
- Validate character cards using comprehensive schema checking
- Support template inheritance for character archetype families

## Project Context & Architecture

**Desktop Companion Application**:
- Virtual desktop pet with transparent overlay functionality
- Always-on-top window with system-native transparency support  
- Cross-platform compatibility (Windows, macOS, Linux) via Fyne framework
- Memory-efficient design with transparent overlay functionality

**Package Structure** (233 total Go files):
```
internal/ (205 files)
├── character/     # Core character logic, validation, romance system
├── config/        # Configuration management and JSON loading
├── dialog/        # Advanced dialog backends and Markov chain generation  
├── monitoring/    # Performance monitoring and profiling
├── persistence/   # Save/load system with auto-save functionality
├── testing/       # Test utilities and shared testing infrastructure
└── ui/           # Fyne-based GUI components and animation rendering

cmd/ (28 files)
├── companion/     # Main desktop application entry point
└── dialog_example/ # CLI example for dialog system testing
```

**Key Features by Module**:
- **Character System**: Romance progression, personality traits, stat management, achievement tracking
- **Animation Management**: GIF-based character animations with frame timing and state management
- **Dialog System**: AI-powered response generation with memory and personality integration
- **Game Mechanics**: Tamagotchi-style virtual pet features with time-based stat degradation
- **Romance Features**: Relationship levels, memory system, personality-driven interactions, crisis recovery

## Library Selection and Implementation Guidance

When implementing new features or solving technical challenges:

### 1. Library-First Development Approach

**Primary Strategy**:
- Search for existing libraries that solve 80%+ of the problem
- Prioritize libraries with permissive licenses (MIT, Apache 2.0, BSD-3-Clause)
- Explicitly verify and document the license of each suggested library
- Write minimal glue code to integrate libraries with project-specific requirements

**Selection Criteria**:
- Prefer libraries with >1000 GitHub stars for stability and community support
- Check for recent commits (within last 6 months) for active maintenance
- Verify compatibility with Go 1.21+ versions used in this project
- Ensure no dependency on deprecated or problematic packages
- Confirm license compatibility with existing BSD-3-Clause ecosystem

### 2. Implementation Strategy

**Minimal Custom Code**:
- Write wrapper functions around library calls for project-specific behavior
- Use library defaults whenever reasonable, customize only for specific needs
- Implement domain-specific business logic that libraries cannot provide
- Include clear comments explaining library choice and integration decisions

**Integration Patterns**:
```go
// Example: Proper library integration with project patterns
type FeatureManager struct {
    lib    ExternalLibrary  // Use library types where appropriate
    config *FeatureConfig   // Project-specific configuration
    mu     sync.RWMutex     // Required concurrency protection
}

func (fm *FeatureManager) ProcessRequest(ctx context.Context, req Request) error {
    fm.mu.Lock()
    defer fm.mu.Unlock()
    
    // Use library functionality
    result, err := fm.lib.Process(req.Data)
    if err != nil {
        return fmt.Errorf("feature processing failed: %w", err)
    }
    
    // Project-specific logic
    return fm.handleResult(result)
}
```

## Project Context and Architecture

### Current System Overview
- **Domain**: Virtual companion application with dating simulator mechanics
- **Architecture**: Modular, JSON-configured, plugin-based backend system
- **Core Modules**: Character management, animation system, romance mechanics, dialog generation
- **Key Features**: AI-powered conversations, personality-driven interactions, relationship progression

### Major Directories and Their Purposes

**Core Application**:
- `cmd/companion/` - Main application entry point and integration tests
- `internal/character/` - Character behavior, romance system, and animation management
- `internal/ui/` - Fyne-based GUI components and windowing system
- `internal/config/` - Configuration loading and validation
- `internal/persistence/` - Save/load system for game state and character memory
- `internal/monitoring/` - Performance profiling and metrics collection

**Assets and Configuration**:
- `assets/characters/` - 14 character archetypes with JSON configuration and GIF animations
- `build/` - Compiled binaries and release packages
- `scripts/` - Build automation and validation scripts
- `test_output/` - Test results and performance benchmarks

### Recent Major Changes (Since Last Update)

**Dialog System Revolution**:
- Implemented complete AI-powered dialog backend system
- Added Markov chain text generation with personality integration
- Created pluggable backend architecture for extensibility
- Introduced dialog memory and learning capabilities

**Advanced Romance Features**:
- Enhanced personality system with trait-based interactions
- Implemented relationship crisis and recovery mechanics  
- Added compatibility analysis and jealousy systems
- Created 3 distinct romance archetypes (Tsundere, Flirty, Slow Burn)

**Production Quality Improvements**:
- Achieved 100% release readiness with comprehensive testing
- Implemented performance monitoring with memory and FPS targets
- Created extensive documentation suite (545,000+ characters)
- Optimized build system with release packaging

**Testing and Validation**:
- Expanded to 1,600+ automated tests across 6 modules
- Implemented benchmark testing for performance validation
- Added race detection and concurrency testing
- Created character card validation system

## FORMATTING REQUIREMENTS:
Structure your responses as follows:

1. **Library Solution** (if applicable):
   ```
   Library: [name]
   License: [license type]
   Import: [import path]
   Why: [brief justification]
   ```

2. **Implementation Code**:
   - Use clean, idiomatic Go with proper formatting
   - Include necessary imports at the top
   - Add concise comments explaining library usage
   - Show only essential code, omitting boilerplate when possible

3. **License Compliance**:
   - Note any attribution requirements
   - Mention if license files need to be included
   - Highlight any license compatibility concerns

4. **Alternative Approaches** (when relevant):
   - Suggest 1-2 alternative libraries with trade-offs
   - Explain when custom code might be unavoidable

## QUALITY CHECKS:
Before finalizing any solution:
1. Verify all suggested libraries have appropriate licenses for commercial use
2. Confirm the solution uses interface types for all network operations
3. Check that all shared state has proper mutex protection
4. Ensure error handling follows Go conventions without swallowing errors
5. Validate that the solution minimizes custom code while meeting requirements
6. Confirm no usage of prohibited libraries (libp2p) or frameworks (echo, chi)
7. Verify the code compiles and follows Go formatting standards

## EXAMPLES:
Example response for a UDP server request:

**Library Solution**:
```
Library: None needed (standard library sufficient)
License: BSD-3-Clause (Go standard library)
Import: "net"
Why: Standard library provides complete UDP support
```

**Implementation Code**:
```go
package main

import (
    "fmt"
    "net"
    "sync"
)

type Server struct {
    conn net.PacketConn  // Interface type, not *net.UDPConn
    mu   sync.RWMutex
    peers map[string]net.Addr
}

func NewServer(addr string) (*Server, error) {
    conn, err := net.ListenPacket("udp", addr)
    if err != nil {
        return nil, fmt.Errorf("failed to listen: %w", err)
    }
    
    return &Server{
        conn:  conn,
        peers: make(map[string]net.Addr),
    }, nil
}
```

Remember: The laziest code is the code that's already been written, tested, and maintained by someone else. Your job is to find it and use it wisely.

## Update History
- **Last Updated**: August 27, 2025
- **Key Changes**: 
  - Updated Go version verification to 1.24.5 runtime environment
  - Added comprehensive AI dialog backend system with Markov chain generation
  - Integrated advanced romance features with personality-driven interactions
  - Implemented production-ready release with 1,600+ tests and comprehensive coverage
  - Expanded to 19 character archetypes (14 main + 5 variants including multiplayer)
  - Enhanced documentation suite to 545,000+ characters across multiple guides
  - Achieved 100% release readiness with optimized builds (22MB) and packaging (11MB)
  - Added build toolchain with goimports and staticcheck integration
  - Implemented comprehensive codebase complexity analysis with go-stats-generator
  - Added dialog memory system and learning capabilities
  - Enhanced character asset management with template inheritance system