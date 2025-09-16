#!/usr/bin/env python3
"""
Asset Generation Configuration Script
Adds assetGeneration configurations to all characters based on their personality and archetype
"""

import json
import os
import sys
from pathlib import Path
from typing import Dict, List, Any

def get_character_asset_config(char_name: str, char_data: Dict) -> Dict:
    """Generate asset generation configuration based on character personality"""
    
    # Base configuration template
    base_config = {
        "generationSettings": {
            "model": "flux1d",
            "artStyle": "anime",
            "resolution": {
                "width": 128,
                "height": 128
            },
            "qualitySettings": {
                "steps": 25,
                "cfgScale": 7.5,
                "seed": -1,
                "sampler": "euler_a",
                "scheduler": "normal"
            },
            "animationSettings": {
                "frameRate": 12,
                "duration": 2.5,
                "loopType": "seamless",
                "optimization": "balanced",
                "maxFileSize": 450,
                "transparencyEnabled": True,
                "colorPalette": "adaptive"
            }
        },
        "assetMetadata": {
            "version": "1.0.0",
            "generatedAt": "2024-12-19T12:00:00Z",
            "generatedBy": "gif-generator v1.0.0"
        },
        "backupSettings": {
            "enabled": True,
            "backupPath": "backups",
            "maxBackups": 5,
            "compressBackups": True
        }
    }
    
    # Character-specific configurations
    character_configs = {
        "default": {
            "basePrompt": "A friendly anime character with short brown hair and warm brown eyes, cute casual clothing, cheerful and approachable appearance, digital art, transparent background, high quality character design suitable for desktop companion",
            "personality_traits": ["friendly", "cheerful", "approachable", "casual"]
        },
        "tsundere": {
            "basePrompt": "A cute anime girl with twin tails and orange/red hair, bright eyes, school uniform or casual dress, tsundere expression with slight blush, arms crossed or hands on hips, digital art, transparent background, high quality anime character design",
            "personality_traits": ["tsundere", "proud", "defensive", "cute"]
        },
        "romance_tsundere": {
            "basePrompt": "A beautiful anime girl with flowing hair and expressive eyes, elegant clothing with romantic touches, tsundere personality showing subtle romantic interest, digital art, transparent background, high quality romantic character design",
            "personality_traits": ["romantic", "tsundere", "elegant", "expressive"]
        },
        "flirty": {
            "basePrompt": "A charming anime girl with vibrant hair and sparkling eyes, stylish outfit with playful accessories, confident and flirty expression, welcoming pose, digital art, transparent background, high quality character design for romance companion",
            "personality_traits": ["flirty", "confident", "charming", "playful"]
        },
        "romance_flirty": {
            "basePrompt": "A stunning anime girl with flowing colorful hair and bright eyes, fashionable romantic outfit, confident flirty smile and pose, romantic accessories, digital art, transparent background, high quality romantic character design",
            "personality_traits": ["romantic", "flirty", "confident", "stunning"]
        },
        "slow_burn": {
            "basePrompt": "A gentle anime character with soft features and calm eyes, modest comfortable clothing, thoughtful and reserved expression, peaceful demeanor, digital art, transparent background, high quality character design for slow romance",
            "personality_traits": ["gentle", "thoughtful", "reserved", "peaceful"]
        },
        "romance_slowburn": {
            "basePrompt": "A graceful anime character with elegant features and deep eyes, sophisticated clothing with subtle romantic touches, contemplative and gentle expression, digital art, transparent background, high quality romantic character design",
            "personality_traits": ["graceful", "elegant", "contemplative", "romantic"]
        },
        "romance_supportive": {
            "basePrompt": "A warm anime character with kind eyes and soft smile, comfortable caring outfit, supportive and nurturing expression, open welcoming pose, digital art, transparent background, high quality character design for supportive romance",
            "personality_traits": ["supportive", "nurturing", "kind", "warm"]
        },
        "klippy": {
            "basePrompt": "A stylized anime version of a paperclip character with anthropomorphic features, metallic silver-blue coloring, expressive eyes, slightly sarcastic but helpful expression, digital art, transparent background, unique character design",
            "personality_traits": ["sarcastic", "helpful", "unique", "metallic"]
        },
        "easy": {
            "basePrompt": "A sweet anime character with soft features and bright eyes, simple comfortable clothing, happy and easy-going expression, relaxed pose, digital art, transparent background, high quality character design for beginner-friendly companion",
            "personality_traits": ["sweet", "easy-going", "relaxed", "simple"]
        },
        "normal": {
            "basePrompt": "A balanced anime character with pleasant features and friendly eyes, normal casual clothing, moderate expression showing contentment, standard pose, digital art, transparent background, high quality character design for balanced experience",
            "personality_traits": ["balanced", "pleasant", "moderate", "content"]
        },
        "hard": {
            "basePrompt": "A sophisticated anime character with sharp features and intense eyes, formal or complex clothing, demanding or high-maintenance expression, confident pose, digital art, transparent background, high quality character design for challenging experience",
            "personality_traits": ["sophisticated", "demanding", "intense", "challenging"]
        },
        "challenge": {
            "basePrompt": "An elite anime character with striking features and piercing eyes, luxurious or complex outfit, proud and challenging expression, commanding pose, digital art, transparent background, high quality character design for expert-level experience",
            "personality_traits": ["elite", "challenging", "commanding", "proud"]
        },
        "specialist": {
            "basePrompt": "A sleepy anime character with drowsy features and tired but cute eyes, comfortable pajamas or cozy clothing, sleepy expression with slight smile, relaxed sleepy pose, digital art, transparent background, high quality character design for energy-focused gameplay",
            "personality_traits": ["sleepy", "cozy", "tired", "cute"]
        },
        "romance": {
            "basePrompt": "A romantic anime character with beautiful features and loving eyes, elegant romantic outfit with soft colors, gentle romantic expression, graceful pose, digital art, transparent background, high quality character design for romance experience",
            "personality_traits": ["romantic", "beautiful", "loving", "elegant"]
        },
        "multiplayer": {
            "basePrompt": "A social anime character with expressive features and bright eyes, casual social outfit with fun accessories, friendly communicative expression, open social pose, digital art, transparent background, high quality character design for multiplayer interaction",
            "personality_traits": ["social", "communicative", "friendly", "expressive"]
        },
        "markov_example": {
            "basePrompt": "An intelligent anime character with thoughtful features and curious eyes, smart casual outfit, contemplative expression showing intelligence, digital art, transparent background, high quality character design for AI-powered dialog system",
            "personality_traits": ["intelligent", "thoughtful", "curious", "analytical"]
        },
        "llm_example": {
            "basePrompt": "A futuristic anime character with tech-savvy features and bright eyes, modern outfit with tech accessories, intelligent expression, digital art, transparent background, high quality character design for LLM integration",
            "personality_traits": ["futuristic", "tech-savvy", "intelligent", "modern"]
        },
        "news_example": {
            "basePrompt": "A knowledgeable anime character with sharp features and attentive eyes, professional casual outfit, informed and alert expression, digital art, transparent background, high quality character design for news and information features",
            "personality_traits": ["knowledgeable", "professional", "informed", "alert"]
        }
    }
    
    # Get character-specific config or use default
    char_config = character_configs.get(char_name, character_configs["default"])
    
    # Standard animation mappings for all characters
    standard_animations = {
        "idle": {
            "promptModifier": "standing calmly with arms at sides, peaceful neutral expression, slight smile, relaxed pose",
            "negativePrompt": "angry, aggressive, dark, scary, low quality, blurry",
            "stateDescription": "Default calm state",
            "frameCount": 6
        },
        "talking": {
            "promptModifier": "speaking with hand gestures, mouth slightly open, expressive face, animated pose, welcoming expression",
            "negativePrompt": "silent, static, angry expression, dark mood",
            "stateDescription": "Speaking or interacting with user",
            "frameCount": 8
        },
        "happy": {
            "promptModifier": "bright cheerful smile, eyes sparkling with joy, hands clasped together or raised, radiant expression",
            "negativePrompt": "sad, angry, neutral expression, dark colors",
            "stateDescription": "Joyful and excited state",
            "frameCount": 6
        },
        "sad": {
            "promptModifier": "downcast eyes, gentle frown, hand touching cheek or covering face, melancholic but still beautiful",
            "negativePrompt": "happy, cheerful, bright colors, aggressive",
            "stateDescription": "Sad or disappointed state",
            "frameCount": 4
        },
        "hungry": {
            "promptModifier": "looking longingly at food, hand on stomach, slightly droopy expression, cute hungry pose",
            "negativePrompt": "full, satisfied, eating, aggressive",
            "stateDescription": "Hungry and wanting food",
            "frameCount": 5
        },
        "eating": {
            "promptModifier": "eating food happily, content expression, food in hands or near mouth, satisfied pose",
            "negativePrompt": "hungry, sad, empty hands, aggressive",
            "stateDescription": "Eating and satisfied",
            "frameCount": 6
        },
        "blushing": {
            "promptModifier": "soft pink blush on cheeks, shy smile, one hand near face, averting gaze slightly, cute embarrassed expression",
            "negativePrompt": "confident, bold, angry, dark mood",
            "stateDescription": "Shy and blushing romantic state",
            "frameCount": 5
        },
        "heart_eyes": {
            "promptModifier": "heart-shaped pupils or sparkles in eyes, loving expression, hands near heart, surrounded by floating hearts",
            "negativePrompt": "normal eyes, angry, sad, dark mood",
            "stateDescription": "In love or adoring state",
            "frameCount": 6
        },
        "shy": {
            "promptModifier": "looking down shyly, hands clasped behind back or in front, timid expression, cute shy pose",
            "negativePrompt": "confident, bold, outgoing, aggressive",
            "stateDescription": "Timid and shy state",
            "frameCount": 4
        },
        "flirty": {
            "promptModifier": "playful wink or flirty smile, confident pose, one hand on hip or touching hair, charming expression",
            "negativePrompt": "shy, timid, serious, angry",
            "stateDescription": "Flirty and charming state",
            "frameCount": 7
        },
        "romantic_idle": {
            "promptModifier": "gentle romantic expression, soft smile, dreamy eyes, peaceful romantic pose",
            "negativePrompt": "aggressive, angry, rushed, unromantic",
            "stateDescription": "Peaceful romantic state",
            "frameCount": 5
        },
        "jealous": {
            "promptModifier": "slightly pouting expression, arms crossed, looking away with subtle jealous expression",
            "negativePrompt": "happy, content, peaceful, aggressive",
            "stateDescription": "Jealous or envious state",
            "frameCount": 4
        },
        "excited_romance": {
            "promptModifier": "excited happy expression with romantic sparkles, jumping or energetic pose, love-struck appearance",
            "negativePrompt": "calm, sad, angry, static",
            "stateDescription": "Excited romantic state",
            "frameCount": 8
        }
    }
    
    # Add battle animations if battleSystem is enabled
    if char_data.get("battleSystem", {}).get("enabled", False):
        standard_animations.update({
            "attack": {
                "promptModifier": "dynamic attack pose, concentrated expression, action stance, power effects",
                "stateDescription": "Attacking in battle",
                "frameCount": 8
            },
            "defend": {
                "promptModifier": "defensive stance, protective pose, focused expression, shield or guard position",
                "stateDescription": "Defending in battle",
                "frameCount": 6
            },
            "heal": {
                "promptModifier": "gentle healing pose, hands glowing with soft light, caring expression, restorative energy",
                "stateDescription": "Healing or supporting",
                "frameCount": 7
            }
        })
    
    # Character-specific animation additions
    if char_name == "aria_luna":
        standard_animations["magical"] = {
            "promptModifier": "casting magic spell, hands glowing with celestial energy, intense concentration, robes flowing dramatically, magical circles and symbols",
            "stateDescription": "Using magical powers",
            "frameCount": 10,
            "customSettings": {
                "qualitySettings": {
                    "steps": 30,
                    "cfgScale": 8.0
                }
            }
        }
        standard_animations["sleeping"] = {
            "promptModifier": "peaceful sleeping pose, eyes closed, serene expression, sitting or reclining position, soft glow",
            "stateDescription": "Resting or sleeping state",
            "frameCount": 4
        }
    elif char_name == "specialist":
        standard_animations["sleeping"] = {
            "promptModifier": "very sleepy expression, eyes drooping or closed, yawning, tired but content pose",
            "stateDescription": "Sleepy and tired state",
            "frameCount": 4
        }
    
    # Combine everything
    asset_config = {
        "basePrompt": char_config["basePrompt"],
        "animationMappings": standard_animations,
        **base_config
    }
    
    return asset_config

def add_missing_features_to_character(char_name: str, char_data: Dict) -> Dict:
    """Add missing standard features to a character while preserving personality"""
    
    # If character already has assetGeneration, skip it
    if "assetGeneration" not in char_data:
        char_data["assetGeneration"] = get_character_asset_config(char_name, char_data)
    
    # Add randomEvents if missing
    if "randomEvents" not in char_data or not char_data["randomEvents"]:
        char_data["randomEvents"] = [
            {
                "name": "spontaneous_moment",
                "description": "A delightful spontaneous interaction",
                "probability": 0.05,
                "effects": {"happiness": 10, "affection": 5},
                "animations": ["happy", "talking"],
                "responses": [
                    "What a lovely surprise! I'm so happy to share this moment with you! 😊",
                    "Life is full of wonderful unexpected moments like this! ✨",
                    "These spontaneous times together are what I treasure most! 💕"
                ],
                "cooldown": 1800
            }
        ]
    
    # Add missing core features with minimal configurations
    if "dialogBackend" not in char_data:
        char_data["dialogBackend"] = {
            "enabled": True,
            "defaultBackend": "markov_chain",
            "fallbackChain": ["simple_random"],
            "confidenceThreshold": 0.6,
            "backends": {
                "markov_chain": {
                    "chainOrder": 2,
                    "minWords": 3,
                    "maxWords": 12,
                    "temperatureMin": 0.4,
                    "temperatureMax": 0.7,
                    "usePersonality": True,
                    "trainingData": [
                        "Hello! I'm so happy to see you again!",
                        "How are you doing today? You look wonderful!",
                        "Thanks for visiting me! I love spending time with you.",
                        "Your presence always brightens my day!",
                        "What would you like to talk about today?",
                        "I'm here if you need someone to chat with!"
                    ]
                }
            }
        }
    
    if "giftSystem" not in char_data:
        char_data["giftSystem"] = {
            "enabled": True,
            "inventorySettings": {
                "maxSlots": 8,
                "autoSort": True,
                "stackSimilar": True
            },
            "preferences": {
                "favoriteCategories": ["food", "flowers", "books"],
                "personalityModifiers": {
                    "food": 1.3,
                    "flowers": 1.5,
                    "books": 1.2
                }
            },
            "memorySettings": {
                "rememberGifts": True,
                "trackPreferences": True,
                "learningEnabled": True
            }
        }
    
    if "multiplayer" not in char_data:
        char_data["multiplayer"] = {
            "enabled": True,
            "botCapable": False,
            "networkID": f"{char_name}_companion_v1",
            "maxPeers": 5,
            "socialLevel": "moderate",
            "networkPersonality": "friendly"
        }
    
    if "newsFeatures" not in char_data:
        char_data["newsFeatures"] = {
            "enabled": True,
            "updateInterval": 1800,
            "maxStoredItems": 20,
            "readingPersonality": "casual",
            "preferredCategories": ["general", "lifestyle"],
            "feeds": []
        }
    
    if "battleSystem" not in char_data:
        char_data["battleSystem"] = {
            "enabled": True,
            "aiDifficulty": "balanced",
            "battleStats": {
                "hp": {"base": 75, "growth": 2.5},
                "attack": {"base": 12, "growth": 1.8},
                "defense": {"base": 10, "growth": 2.0},
                "speed": {"base": 8, "growth": 1.5}
            },
            "availableActions": ["attack", "defend", "heal"]
        }
    
    if "generalEvents" not in char_data:
        char_data["generalEvents"] = [
            {
                "name": "friendly_chat",
                "description": "A casual conversation moment",
                "responses": [
                    "I've been thinking about our friendship today! 😊",
                    "What's been on your mind lately?",
                    "I love our little conversations!"
                ],
                "choices": [
                    {
                        "text": "Share your thoughts",
                        "effects": {"happiness": 5, "affection": 3},
                        "responses": ["Thank you for sharing! I really appreciate that! 💕"],
                        "animation": "happy"
                    }
                ],
                "cooldown": 3600,
                "category": "conversation"
            }
        ]
    
    if "progression" not in char_data:
        char_data["progression"] = {
            "levels": [
                {
                    "name": "New Friend",
                    "requirement": {"age": 0},
                    "size": 128
                },
                {
                    "name": "Good Friend", 
                    "requirement": {"age": 86400, "affection": 20},
                    "size": 132
                },
                {
                    "name": "Close Friend",
                    "requirement": {"age": 259200, "affection": 45, "trust": 30},
                    "size": 136
                }
            ]
        }
    
    if "behavior" not in char_data:
        char_data["behavior"] = {
            "idleTimeout": 30,
            "movementEnabled": True,
            "defaultSize": 128
        }
    
    return char_data

def process_character_file(char_path: Path):
    """Process a single character file"""
    try:
        # Load character data
        with open(char_path, 'r', encoding='utf-8') as f:
            char_data = json.load(f)
        
        char_name = char_path.parent.name
        print(f"Processing {char_name}...")
        
        # Add missing features
        updated_data = add_missing_features_to_character(char_name, char_data)
        
        # Write back to file
        with open(char_path, 'w', encoding='utf-8') as f:
            json.dump(updated_data, f, indent=2, ensure_ascii=False)
        
        print(f"  ✓ Updated {char_name}")
        return True
        
    except Exception as e:
        print(f"  ✗ Error processing {char_path}: {e}")
        return False

def main():
    if len(sys.argv) != 2:
        print("Usage: python add_asset_generation.py <characters_directory>")
        sys.exit(1)
    
    characters_dir = Path(sys.argv[1])
    
    if not characters_dir.exists():
        print(f"Error: Directory {characters_dir} does not exist")
        sys.exit(1)
    
    # Process all character files
    processed = 0
    failed = 0
    
    for char_dir in characters_dir.iterdir():
        if char_dir.is_dir():
            char_file = char_dir / "character.json"
            if char_file.exists():
                if process_character_file(char_file):
                    processed += 1
                else:
                    failed += 1
    
    print(f"\nCompleted: {processed} characters processed, {failed} failed")

if __name__ == "__main__":
    main()
