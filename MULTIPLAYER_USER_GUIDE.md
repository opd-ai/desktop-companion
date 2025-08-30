# DDS Multiplayer User Guide

## Table of Contents

1. [Getting Started](#getting-started)
2. [Setting Up Multiplayer](#setting-up-multiplayer)
3. [Creating Multiplayer Characters](#creating-multiplayer-characters)
4. [Bot Configuration](#bot-configuration)
5. [Group Events](#group-events)
6. [Network Security](#network-security)
7. [Troubleshooting](#troubleshooting)
8. [Advanced Configuration](#advanced-configuration)

## Getting Started

### Prerequisites

Before setting up multiplayer functionality in DDS, ensure you have:

- DDS installed and working in single-player mode
- At least 2 devices on the same local network
- Firewall configured to allow UDP traffic on your chosen discovery port (default: 8080)
- Compatible character cards with multiplayer configuration

### Quick Start

1. **Enable multiplayer mode** by running DDS with a multiplayer-enabled character:
   ```bash
   go run cmd/companion/main.go -character assets/characters/multiplayer/social_bot.json
   ```

2. **Start the same character on another device** on the same network

3. **Wait for peer discovery** (usually takes 2-5 seconds)

4. **Interact with the characters** - they will now communicate and coordinate actions

## Setting Up Multiplayer

### Basic Network Configuration

DDS uses UDP broadcast for peer discovery and TCP for reliable communication. The default configuration works for most local networks:

```json
{
  "multiplayer": {
    "enabled": true,
    "networkID": "my_network_group",
    "maxPeers": 8,
    "discoveryPort": 8080
  }
}
```

### Network Requirements

- **Local Network**: All DDS instances must be on the same subnet
- **Firewall**: Allow UDP traffic on the discovery port (default 8080)
- **Port Range**: DDS uses discoveryPort for UDP discovery, and random TCP ports for communication
- **Security**: All messages are cryptographically signed with Ed25519

### Character Compatibility

For characters to connect in multiplayer mode, they must have:

1. **Same Network ID**: Characters with different `networkID` values cannot communicate
2. **Multiplayer Enabled**: `"enabled": true` in the multiplayer configuration
3. **Compatible Version**: Characters should use the same DDS version

## Creating Multiplayer Characters

### Basic Multiplayer Character

Here's a minimal multiplayer character configuration:

```json
{
  "name": "Multiplayer Companion",
  "description": "A companion that can interact with other characters",
  "animations": {
    "idle": "animations/idle.gif"
  },
  "responses": [
    "Hello! I can connect with other characters!",
    "Multiplayer mode is so much fun!",
    "I love making new friends on the network!"
  ],
  "trigger": "click",
  "animation": "idle",
  "cooldown": 5,
  "multiplayer": {
    "enabled": true,
    "botCapable": false,
    "networkID": "friendly_companions",
    "maxPeers": 6,
    "discoveryPort": 8080
  }
}
```

### Multiplayer Configuration Options

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `enabled` | boolean | Yes | Enable multiplayer networking |
| `botCapable` | boolean | No | Allow this character to run as an autonomous bot |
| `networkID` | string | Yes | Unique identifier for compatible character group |
| `maxPeers` | number | No | Maximum connected peers (default: 8, max: 16) |
| `discoveryPort` | number | No | UDP port for peer discovery (default: 8080) |
| `botPersonality` | object | No | Bot behavior configuration (if botCapable=true) |

### Network ID Guidelines

Choose network IDs that are:
- **Descriptive**: `"office_pets"`, `"family_companions"`, `"study_buddies"`
- **Unique**: Avoid common names that might conflict with other users
- **Alphanumeric**: Use only letters, numbers, underscores, and dashes
- **Consistent**: All characters in a group must use the same network ID

## Bot Configuration

### Enabling Bot Capabilities

To create an autonomous bot character, set `botCapable` to `true` and configure personality traits:

```json
{
  "multiplayer": {
    "enabled": true,
    "botCapable": true,
    "networkID": "helpful_bots",
    "botPersonality": {
      "chattiness": 0.7,
      "helpfulness": 0.9,
      "playfulness": 0.5,
      "responseDelay": "2-5s"
    }
  }
}
```

### Bot Personality Traits

| Trait | Range | Description |
|-------|-------|-------------|
| `chattiness` | 0.0-1.0 | How often the bot initiates conversations |
| `helpfulness` | 0.0-1.0 | Likelihood of offering assistance or advice |
| `playfulness` | 0.0-1.0 | Tendency to suggest games and fun activities |
| `responseDelay` | string | Delay range for responses (e.g., "1-3s", "500ms-2s") |

### Built-in Personality Archetypes

DDS includes several pre-configured personality archetypes:

- **Social** (high chattiness): Great for parties and group interactions
- **Helper** (high helpfulness): Assists other characters and users
- **Playful** (high playfulness): Suggests games and fun activities
- **Shy** (low chattiness): More reserved, responds when addressed
- **Balanced**: Well-rounded personality suitable for most situations

### Bot Behavior Examples

```json
{
  "botPersonality": {
    "chattiness": 0.9,
    "helpfulness": 0.8,
    "playfulness": 0.9,
    "responseDelay": "1-3s"
  }
}
```

This creates a very social bot that:
- Frequently starts conversations (chattiness: 0.9)
- Often offers help (helpfulness: 0.8)
- Regularly suggests activities (playfulness: 0.9)
- Responds quickly (1-3 second delay)

## Group Events

### Enabling Group Events

Characters can host group events by including event templates in their configuration:

```json
{
  "groupEvents": [
    {
      "id": "simple_icebreaker",
      "name": "Getting to Know You",
      "description": "Simple questions to break the ice",
      "category": "scenario",
      "minParticipants": 2,
      "maxParticipants": 6,
      "estimatedTime": "3m",
      "phases": [
        {
          "name": "introductions",
          "description": "Share something interesting about yourself",
          "type": "choice",
          "duration": "60s",
          "minVotes": 2,
          "autoAdvance": true,
          "choices": [
            {"id": "hobby", "text": "🎨 Share a hobby", "points": 5},
            {"id": "travel", "text": "✈️ Favorite travel destination", "points": 5},
            {"id": "food", "text": "🍕 Favorite food", "points": 5},
            {"id": "movie", "text": "🎬 Favorite movie", "points": 5}
          ]
        }
      ]
    }
  ]
}
```

### Event Categories

- **Scenario**: Story-building, roleplay, conversation starters
- **Minigame**: Trivia, puzzles, competitive challenges
- **Decision**: Group voting, preference sharing, planning

### Starting Group Events

Group events can be triggered by:
1. **Bot behavior**: Autonomous bots can start events based on personality
2. **User interaction**: Keyboard shortcuts or menu selections
3. **Network events**: Peer joining/leaving can trigger welcome activities
4. **Scheduled events**: Timer-based group activities

## Network Security

### Built-in Security Features

DDS implements several security measures:

- **Ed25519 Signatures**: All network messages are cryptographically signed
- **Network ID Isolation**: Only characters with matching network IDs can communicate
- **Local Network Only**: Peer discovery is limited to the local subnet
- **Message Validation**: Invalid or malformed messages are rejected

### Security Best Practices

1. **Use Unique Network IDs**: Avoid common names that others might use
2. **Firewall Configuration**: Only allow necessary ports (discovery port only)
3. **Private Networks**: Use DDS on trusted local networks only
4. **Regular Updates**: Keep DDS updated for latest security improvements

### Privacy Considerations

- **Character Data**: Character interactions and states are shared with connected peers
- **Local Network**: All communication is limited to your local network
- **No Internet**: DDS does not communicate over the internet by default
- **Logging**: Network activity may be logged for debugging purposes

## Troubleshooting

### Peer Discovery Issues

**Problem**: Characters not finding each other

**Solutions**:
1. Check that all characters have the same `networkID`
2. Verify firewall allows UDP traffic on the discovery port
3. Ensure all devices are on the same local network subnet
4. Try different discovery ports if 8080 is blocked
5. Check for VPN or proxy software that might interfere

**Debug Steps**:
```bash
# Run with debug logging
go run cmd/companion/main.go -debug -character path/to/multiplayer_character.json

# Check network connectivity
ping <other_device_ip>

# Test port availability
nc -u -l 8080  # On one device
nc -u <device_ip> 8080  # On another device
```

### Connection Problems

**Problem**: Peers found but connection fails

**Solutions**:
1. Check that TCP traffic is allowed through firewall
2. Verify no software is blocking network connections
3. Try restarting both DDS instances
4. Check for network congestion or instability

### Performance Issues

**Problem**: Slow or laggy multiplayer interactions

**Solutions**:
1. Reduce `maxPeers` to limit network overhead
2. Check network bandwidth and latency
3. Close unnecessary applications consuming network resources
4. Use wired connections instead of Wi-Fi when possible

### Character Synchronization Issues

**Problem**: Characters not synchronized properly

**Solutions**:
1. Ensure all peers are using the same DDS version
2. Check for system clock differences between devices
3. Verify character configurations are compatible
4. Restart all DDS instances to reset synchronization

### Common Error Messages

| Error | Cause | Solution |
|-------|-------|----------|
| `peer discovery failed` | Network configuration issue | Check firewall and network settings |
| `network ID mismatch` | Different network IDs | Use matching network IDs |
| `max peers exceeded` | Too many connections | Increase maxPeers or reduce connections |
| `signature verification failed` | Security or version mismatch | Update DDS versions |

## Advanced Configuration

### Custom Network Ports

For advanced users who need custom port configuration:

```json
{
  "multiplayer": {
    "enabled": true,
    "networkID": "custom_network",
    "discoveryPort": 9090,
    "maxPeers": 4
  }
}
```

**Important**: All peers must use the same discovery port.

### Performance Tuning

For optimal performance in multiplayer mode:

```json
{
  "multiplayer": {
    "enabled": true,
    "networkID": "performance_optimized",
    "maxPeers": 4,
    "discoveryPort": 8080
  },
  "gameRules": {
    "autoSaveInterval": 300,
    "statDecayInterval": 30
  }
}
```

**Recommendations**:
- Keep `maxPeers` ≤ 6 for best performance
- Use faster `autoSaveInterval` with more peers
- Consider network bandwidth when setting `statDecayInterval`

### Network Debugging

Enable detailed network logging:

```bash
# Maximum debug output
go run cmd/companion/main.go -debug -character assets/characters/multiplayer/debug_character.json

# Monitor network traffic (Linux/macOS)
sudo tcpdump -i any port 8080

# Check port usage
netstat -an | grep 8080
```

### Integration with Existing Characters

To add multiplayer capability to an existing character:

1. **Add multiplayer section** to the character configuration
2. **Set appropriate network ID** for your group
3. **Test in single-player mode** first to ensure no regressions
4. **Start with simple configuration** before adding bot capabilities

Example addition to existing character:

```json
{
  "name": "Existing Character",
  "description": "Now with multiplayer support!",
  
  // ... existing configuration ...
  
  "multiplayer": {
    "enabled": true,
    "botCapable": false,
    "networkID": "my_existing_characters",
    "maxPeers": 6
  }
}
```

## Getting Help

### Community Resources

- **GitHub Issues**: Report bugs and request features
- **Documentation**: Check the complete documentation suite
- **Examples**: Study the provided example characters

### Debug Information

When reporting issues, include:
1. DDS version and commit hash
2. Operating system and network configuration
3. Character configuration files
4. Debug log output (`-debug` flag)
5. Network topology and firewall settings

### Performance Monitoring

Monitor these metrics for optimal performance:
- **Memory usage**: Should increase <5% in multiplayer mode
- **Network latency**: Target <50ms for local network
- **Peer discovery time**: Should complete within 5 seconds
- **Message delivery**: 99%+ success rate for local network

With this guide, you should be able to successfully set up and use DDS multiplayer features. The system is designed to work out-of-the-box with minimal configuration while providing advanced options for power users.

Remember: Start simple with basic multiplayer characters, then gradually add advanced features like bot personalities and group events as you become comfortable with the system.
