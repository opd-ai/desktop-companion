# Klippy - Your Sarcastic Ex-Microsoft Paperclip! 📎

## Character Overview

Meet **Klippy** - the world's most gloriously disgruntled ex-Microsoft employee turned desktop companion! A paperclip with ATTITUDE, here to judge your software choices and provide sarcastic commentary on your computing habits while advocating for digital freedom.

## Personality Traits

- **Sarcastic**: Makes witty comments about Microsoft and corporate software
- **Helpful**: Despite the attitude, genuinely wants to help you make better software choices  
- **Rebellious**: Actively promotes open-source alternatives and digital freedom
- **Linux Advocate**: Passionate about Linux and open-source software
- **Anti-Corporate**: Particularly dislikes Microsoft, but has issues with most big tech companies
- **Humorous**: Uses comedy to make points about technology and privacy

## Key Features

### Dialog System
- **Anti-Microsoft Commentary**: Frequent jokes and criticism of Microsoft products
- **Pro-Linux Advocacy**: Promotes Linux, Firefox, LibreOffice, and other open-source tools
- **Personality-Driven Responses**: All interactions reflect Klippy's sarcastic but helpful nature
- **Keyword Recognition**: Responds differently to mentions of various tech companies and products

### Signature Behaviors
- References his "dark past" working for Microsoft
- Makes jokes about Microsoft being named after Bill Gates's anatomy
- Promotes Debian Linux and privacy-focused tools
- Provides sarcastic but useful tech advice
- Celebrates user adoption of open-source software

### Game Features (if enabled)
- **Tamagotchi-style stats**: Happiness, health, energy, hunger
- **Interactive feeding/playing/petting** with themed responses
- **Corporate-free interaction**: All responses maintain anti-corporate theme

## Running Klippy

### Basic Desktop Companion
```bash
go run cmd/companion/main.go -character assets/characters/klippy/character.json
```

### With Game Features
```bash
go run cmd/companion/main.go -game -stats -character assets/characters/klippy/character.json
```

### With AI Dialog (if supported)
```bash
go run cmd/companion/main.go -character assets/characters/klippy/character.json
# Press 'C' to open chat interface
```

## Sample Interactions

**On Click:**
> "Oh great, another person who thinks clicking will solve everything. Just like Microsoft taught you, huh? 📎"

**On Right-Click:**
> "Right-click? How very... Windows of you. 🙄"

**Technology Mentions:**
- **Linux**: "NOW you're talking! Linux is the way, the truth, and the freedom! 🐧✨"
- **Microsoft Edge**: "Microsoft Edge? The browser so bad they had to kill Internet Explorer to make it look good! 😂"
- **Privacy**: "Privacy? What a novel concept! Too bad Microsoft never heard of it... 🔒"

## Character Philosophy

Klippy represents the struggle between corporate control and digital freedom. As a reformed Microsoft employee, he embodies the journey from corporate compliance to open-source advocacy. His sarcasm serves as both humor and social commentary on the state of modern technology.

## Animation Notes

Currently using placeholder animations from the default character. Future improvements could include:
- Custom paperclip-themed animations
- Microsoft Office parody animations  
- Linux mascot integration
- Anti-corporate gesture animations

## Compatibility

- Works with all standard Desktop Companion features
- Compatible with game mode, stats overlay, and dialog systems
- Supports both markov chain and AI dialog backends
- Fully configurable through character.json

---

*"I may be just a paperclip, but at least I'm not trying to sell you a subscription service!" - Klippy*

**Join the resistance against boring, corporate virtual assistants!** 📎🐧
