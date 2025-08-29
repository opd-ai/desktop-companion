# Integrated Chatbot Demo

This demo showcases the complete chatbot integration in the DesktopWindow system.

## Features Demonstrated

### Chatbot Interface Integration
- **Keyboard Shortcut**: Press `C` to toggle the chatbot interface
- **Enhanced Focus**: Input field automatically focuses when chatbot opens
- **Quick Close**: Press `ESC` to quickly close chatbot interface
- **Context Menu Access**: Right-click the character and select "Open Chat"/"Close Chat"
- **AI-Powered Responses**: Chatbot uses the dialog backend system for AI responses
- **Conversation History**: Multi-turn conversations with persistent history
- **Conditional Availability**: Chatbot only appears for characters with dialog backend enabled
- **Helpful Shortcuts**: Right-click → "Shortcuts" for quick keyboard reference

### User Interactions
1. **Basic Interaction**: Left-click the character for traditional dialog responses
2. **Context Menu**: Right-click for a context menu including chatbot access and helpful shortcuts
3. **Chatbot Chat**: Use keyboard shortcut `C` or context menu to open chat interface
4. **Enhanced Navigation**: 
   - Press `ESC` to quickly close chatbot interface
   - Input field automatically focuses when chatbot opens
   - Right-click → "Shortcuts" for quick reference
5. **Dragging**: Character can be moved around the desktop (if movement enabled)

## Running the Demo

```bash
cd examples/integrated_demo
go run .
```

## Demo Character

The demo creates an "AI Companion" character with:
- Dialog backend enabled for AI chat functionality
- Traditional click/right-click responses
- Multiple animation states (idle, talking, happy)
- Draggable movement capability
- Debug mode enabled to show keyboard shortcuts in console

## Technical Integration

This demo demonstrates the complete integration of the chatbot interface into the main desktop window system:

- **DesktopWindow**: Contains chatbot interface as part of its UI content
- **Keyboard Shortcuts**: 'C' key toggles chatbot, 'S' key toggles stats (if available)
- **Context Menu**: Dynamic menu items based on character capabilities
- **Character Cards**: Uses dialog backend configuration to enable AI chat
- **Rendering**: Chatbot interface rendered as overlay widget in window layout

The integration preserves the library-first philosophy and follows established Fyne patterns.
