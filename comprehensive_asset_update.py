#!/usr/bin/env python3
"""
Comprehensive Asset Generation Pipeline Integration
Updates ALL character JSON files for gif-generator compatibility
"""

import json
import os
import sys
from pathlib import Path
from typing import Dict, List, Any

def get_comprehensive_asset_config(char_name: str, char_data: Dict, file_path: Path) -> Dict:
    """Generate complete asset generation configuration based on character personality and existing animations"""
    
    # Base configuration template according to task requirements
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
    
    # Determine character archetype from name, description, and file path
    archetype = determine_character_archetype(char_name, char_data, file_path)
    
    # Character-specific base prompts
    base_prompts = {
        "default": "A friendly anime character with short brown hair and warm brown eyes, cute casual clothing, cheerful and approachable appearance, digital art, transparent background, high quality character design suitable for desktop companion",
        "tsundere": "A cute anime girl with twin tails and orange/red hair, bright eyes, school uniform or casual dress, tsundere expression with slight blush, arms crossed or hands on hips, digital art, transparent background, high quality anime character design",
        "romance_tsundere": "A beautiful anime girl with flowing hair and expressive eyes, elegant clothing with romantic touches, tsundere personality showing subtle romantic interest, digital art, transparent background, high quality romantic character design",
        "flirty": "A charming anime girl with vibrant hair and sparkling eyes, stylish outfit with playful accessories, confident and flirty expression, welcoming pose, digital art, transparent background, high quality character design for romance companion",
        "romance_flirty": "A stunning anime girl with flowing colorful hair and bright eyes, fashionable romantic outfit, confident flirty smile and pose, romantic accessories, digital art, transparent background, high quality romantic character design",
        "slow_burn": "A gentle anime character with soft features and calm eyes, modest comfortable clothing, thoughtful and reserved expression, peaceful demeanor, digital art, transparent background, high quality character design for slow romance",
        "romance_slowburn": "A graceful anime character with elegant features and deep eyes, sophisticated clothing with subtle romantic touches, contemplative and gentle expression, digital art, transparent background, high quality romantic character design",
        "romance_supportive": "A warm anime character with kind eyes and soft smile, comfortable caring outfit, supportive and nurturing expression, open welcoming pose, digital art, transparent background, high quality character design for supportive romance",
        "klippy": "A stylized anime version of a paperclip character with anthropomorphic features, metallic silver-blue coloring, expressive eyes, slightly sarcastic but helpful expression, digital art, transparent background, unique character design",
        "aria_luna": "A beautiful anime girl with long flowing silver hair and bright purple eyes, wearing a flowing celestial robe with star patterns, ethereal and mystical appearance, digital art, transparent background, high quality, detailed character design suitable for desktop companion, magical aura, soft lighting",
        "easy": "A sweet anime character with soft features and bright eyes, simple comfortable clothing, happy and easy-going expression, relaxed pose, digital art, transparent background, high quality character design for beginner-friendly companion",
        "normal": "A balanced anime character with pleasant features and friendly eyes, normal casual clothing, moderate expression showing contentment, standard pose, digital art, transparent background, high quality character design for balanced experience",
        "hard": "A sophisticated anime character with sharp features and intense eyes, formal or complex clothing, demanding or high-maintenance expression, confident pose, digital art, transparent background, high quality character design for challenging experience",
        "challenge": "An elite anime character with striking features and piercing eyes, luxurious or complex outfit, proud and challenging expression, commanding pose, digital art, transparent background, high quality character design for expert-level experience",
        "specialist": "A sleepy anime character with drowsy features and tired but cute eyes, comfortable pajamas or cozy clothing, sleepy expression with slight smile, relaxed sleepy pose, digital art, transparent background, high quality character design for energy-focused gameplay",
        "romance": "A romantic anime character with beautiful features and loving eyes, elegant romantic outfit with soft colors, gentle romantic expression, graceful pose, digital art, transparent background, high quality character design for romance experience",
        "multiplayer": "A social anime character with expressive features and bright eyes, casual social outfit with fun accessories, friendly communicative expression, open social pose, digital art, transparent background, high quality character design for multiplayer interaction",
        "markov_example": "An intelligent anime character with thoughtful features and curious eyes, smart casual outfit, contemplative expression showing intelligence, digital art, transparent background, high quality character design for AI-powered dialog system",
        "llm_example": "A futuristic anime character with tech-savvy features and bright eyes, modern outfit with tech accessories, intelligent expression, digital art, transparent background, high quality character design for LLM integration",
        "news_example": "A knowledgeable anime character with sharp features and attentive eyes, professional casual outfit, informed and alert expression, digital art, transparent background, high quality character design for news and information features",
        "helper_bot": "A helpful anime character with kind features and bright eyes, assistant-style outfit, eager-to-help expression, supportive pose, digital art, transparent background, high quality character design for helpful multiplayer bot",
        "social_bot": "A social anime character with animated features and sparkling eyes, trendy social outfit, chatty and friendly expression, welcoming pose, digital art, transparent background, high quality character design for social multiplayer bot",
        "group_moderator": "An organized anime character with confident features and attentive eyes, moderator-style outfit, responsible and coordinating expression, leadership pose, digital art, transparent background, high quality character design for group management bot",
        "shy_companion": "A quiet anime character with gentle features and soft eyes, modest comfortable clothing, shy but warm expression, reserved pose, digital art, transparent background, high quality character design for introverted companion"
    }
    
    # Get existing animations from character data
    existing_animations = char_data.get("animations", {})
    
    # Create comprehensive animation mappings for ALL existing animations
    animation_mappings = {}
    
    # Standard animation mappings
    standard_mappings = {
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
        },
        # Character-specific animations
        "magical": {
            "promptModifier": "casting magic spell, hands glowing with celestial energy, intense concentration, robes flowing dramatically, magical circles and symbols",
            "stateDescription": "Using magical powers",
            "frameCount": 10,
            "customSettings": {
                "qualitySettings": {
                    "steps": 30,
                    "cfgScale": 8.0
                }
            }
        },
        "sleeping": {
            "promptModifier": "peaceful sleeping pose, eyes closed, serene expression, sitting or reclining position, soft glow",
            "stateDescription": "Resting or sleeping state",
            "frameCount": 4
        },
        "thinking": {
            "promptModifier": "thoughtful expression, hand near chin, contemplative pose, focused look",
            "stateDescription": "Thinking or processing information",
            "frameCount": 5
        },
        "excited": {
            "promptModifier": "energetic bouncing motion, wide smile, enthusiastic pose, dynamic expression",
            "stateDescription": "Excited and energetic state",
            "frameCount": 8
        },
        "reading": {
            "promptModifier": "reading a book or document, focused expression, holding reading material, intellectual pose",
            "stateDescription": "Reading or studying information",
            "frameCount": 6
        },
        "critical": {
            "promptModifier": "stern or disapproving expression, arms crossed, serious demeanor, demanding pose",
            "stateDescription": "Critical or demanding state",
            "frameCount": 5
        },
        "demanding": {
            "promptModifier": "authoritative pose, pointing gesture, expectant expression, commanding presence",
            "stateDescription": "Making demands or requests",
            "frameCount": 6
        },
        "boost": {
            "promptModifier": "energized and motivated expression, fist pump or celebratory pose, boosted confidence",
            "stateDescription": "Boosted or energized state",
            "frameCount": 7
        },
        "comforting": {
            "promptModifier": "gentle caring expression, open arms, nurturing pose, warm and supportive demeanor",
            "stateDescription": "Providing comfort and support",
            "frameCount": 6
        },
        "caring": {
            "promptModifier": "attentive and loving expression, reaching out gesture, caring pose, protective stance",
            "stateDescription": "Showing care and concern",
            "frameCount": 5
        },
        # Battle animations
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
        },
        "victory": {
            "promptModifier": "triumphant pose, victory sign or raised fist, celebratory expression, winner stance",
            "stateDescription": "Celebrating victory",
            "frameCount": 8
        },
        "defeat": {
            "promptModifier": "disappointed but graceful expression, hand on heart, accepting pose, respectful demeanor",
            "stateDescription": "Accepting defeat gracefully",
            "frameCount": 5
        },
        "special": {
            "promptModifier": "charging special ability, concentrated energy gathering, dramatic pose, power buildup",
            "stateDescription": "Preparing special attack",
            "frameCount": 10
        }
    }
    
    # Map all existing animations
    for anim_name in existing_animations.keys():
        if anim_name in standard_mappings:
            animation_mappings[anim_name] = standard_mappings[anim_name]
        else:
            # Create generic mapping for unknown animations
            animation_mappings[anim_name] = {
                "promptModifier": f"expressive pose and animation for {anim_name} state, character showing {anim_name} emotion or action",
                "stateDescription": f"Character in {anim_name} state",
                "frameCount": 6
            }
    
    # Ensure core animations are present
    core_animations = ["idle", "talking", "happy", "sad", "hungry", "eating"]
    for core_anim in core_animations:
        if core_anim not in animation_mappings:
            animation_mappings[core_anim] = standard_mappings[core_anim]
    
    # Character-specific customizations
    if archetype == "aria_luna" and "magical" not in animation_mappings:
        animation_mappings["magical"] = standard_mappings["magical"]
    if archetype in ["specialist", "aria_luna"] and "sleeping" not in animation_mappings:
        animation_mappings["sleeping"] = standard_mappings["sleeping"]
    
    # Combine everything
    asset_config = {
        "basePrompt": base_prompts.get(archetype, base_prompts["default"]),
        "animationMappings": animation_mappings,
        **base_config
    }
    
    return asset_config

def determine_character_archetype(char_name: str, char_data: Dict, file_path: Path) -> str:
    """Determine character archetype from various sources"""
    
    # Check file path for clues
    path_str = str(file_path).lower()
    
    if "aria_luna" in path_str:
        return "aria_luna"
    elif "tsundere" in path_str:
        if "romance" in path_str:
            return "romance_tsundere"
        return "tsundere"
    elif "flirty" in path_str:
        if "romance" in path_str:
            return "romance_flirty"
        return "flirty"
    elif "slow_burn" in path_str or "slowburn" in path_str:
        if "romance" in path_str:
            return "romance_slowburn"
        return "slow_burn"
    elif "supportive" in path_str:
        return "romance_supportive"
    elif "romance" in path_str:
        return "romance"
    elif "klippy" in path_str:
        return "klippy"
    elif "easy" in path_str:
        return "easy"
    elif "normal" in path_str:
        return "normal"
    elif "hard" in path_str:
        return "hard"
    elif "challenge" in path_str:
        return "challenge"
    elif "specialist" in path_str:
        return "specialist"
    elif "multiplayer" in path_str:
        if "helper" in path_str:
            return "helper_bot"
        elif "social" in path_str:
            return "social_bot"
        elif "group" in path_str or "moderator" in path_str:
            return "group_moderator"
        elif "shy" in path_str:
            return "shy_companion"
        return "multiplayer"
    elif "markov" in path_str:
        return "markov_example"
    elif "llm" in path_str:
        return "llm_example"
    elif "news" in path_str:
        return "news_example"
    
    # Check character name and description
    name_desc = f"{char_name} {char_data.get('description', '')}".lower()
    
    if "tsundere" in name_desc:
        return "tsundere"
    elif "flirty" in name_desc:
        return "flirty"
    elif "shy" in name_desc:
        return "shy_companion"
    elif "romance" in name_desc:
        return "romance"
    elif "multiplayer" in name_desc or "social" in name_desc:
        return "multiplayer"
    elif "news" in name_desc:
        return "news_example"
    elif "helper" in name_desc:
        return "helper_bot"
    
    return "default"

def update_character_asset_generation(file_path: Path):
    """Update a single character file with complete asset generation config"""
    
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            char_data = json.load(f)
        
        char_name = char_data.get("name", file_path.stem)
        
        # Generate new asset generation config
        new_asset_config = get_comprehensive_asset_config(char_name, char_data, file_path)
        
        # Update or add assetGeneration
        char_data["assetGeneration"] = new_asset_config
        
        # Write back with proper formatting
        with open(file_path, 'w', encoding='utf-8') as f:
            json.dump(char_data, f, indent=2, ensure_ascii=False)
        
        return True, None
        
    except Exception as e:
        return False, str(e)

def main():
    if len(sys.argv) != 2:
        print("Usage: python comprehensive_asset_update.py <characters_directory>")
        sys.exit(1)
    
    characters_dir = Path(sys.argv[1])
    if not characters_dir.exists():
        print(f"Directory not found: {characters_dir}")
        sys.exit(1)
    
    print("=== COMPREHENSIVE ASSET GENERATION PIPELINE INTEGRATION ===\\n")
    
    # Find all character JSON files (excluding templates)
    json_files = []
    for root, dirs, files in os.walk(characters_dir):
        for file in files:
            if file.endswith('.json') and 'templates' not in root:
                full_path = Path(root) / file
                json_files.append(full_path)
    
    print(f"Found {len(json_files)} character files to process\\n")
    
    successful = 0
    failed = 0
    
    for json_file in sorted(json_files):
        relative_path = json_file.relative_to(characters_dir)
        print(f"Processing {relative_path}...", end=" ")
        
        success, error = update_character_asset_generation(json_file)
        
        if success:
            print("✅ Updated")
            successful += 1
        else:
            print(f"❌ Failed: {error}")
            failed += 1
    
    print(f"\\n=== RESULTS ===")
    print(f"Successfully updated: {successful}")
    print(f"Failed: {failed}")
    print(f"Total files: {len(json_files)}")
    
    if failed == 0:
        print("\\n🎉 All character files now have complete asset generation configurations!")
        print("\\n📋 Next steps:")
        print("• Test with: go run cmd/gif-generator/main.go batch --input assets/characters/ --model flux1d --dry-run")
        print("• Validate: go run tools/validate_characters.go assets/characters/*/character.json")
        print("• Generate assets: go run cmd/gif-generator/main.go batch --input assets/characters/ --model flux1d")

if __name__ == "__main__":
    main()
