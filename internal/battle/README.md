# DDS Battle System: Core Framework

This package implements the core turn-based battle system for Desktop Dating Simulator (DDS).

## WHY
- Provides a fair, extensible, and testable battle system for both AI and multiplayer
- Strictly enforces fairness and balance constraints
- Integrates with existing DDS character, animation, and network systems

## Key Files
- `manager.go`: BattleManager, state, and interface
- `actions.go`: Action processing and validation
- `fairness.go`: Fairness constraint enforcement
- `ai.go`: Timeout-driven AI decisions
- `manager_test.go`: Unit tests for all core logic

## Design Principles
- Standard library only (no external dependencies)
- All shared state protected by `sync.RWMutex`
- All errors handled explicitly
- Functions <30 lines, single responsibility
- Code is readable by junior developers

## Usage
See `manager_test.go` for usage and test cases.
