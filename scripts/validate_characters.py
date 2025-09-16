#!/usr/bin/env python3
"""
Character Configuration Validator
Validates JSON syntax, feature configurations, and compatibility
"""

import json
import os
import sys
from pathlib import Path
from typing import Dict, List, Any, Tuple
import re

def validate_json_syntax(file_path: Path) -> Tuple[bool, str]:
    """Validate JSON syntax and structure"""
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            json.load(f)
        return True, "Valid JSON syntax"
    except json.JSONDecodeError as e:
        return False, f"JSON syntax error: {e}"
    except Exception as e:
        return False, f"File error: {e}"

def validate_required_fields(char_data: Dict) -> List[str]:
    """Validate that required fields are present"""
    errors = []
    
    required_fields = ["name", "description", "animations"]
    for field in required_fields:
        if field not in char_data:
            errors.append(f"Missing required field: {field}")
    
    # Validate required animations
    if "animations" in char_data:
        required_animations = ["idle", "talking", "happy", "sad"]
        for anim in required_animations:
            if anim not in char_data["animations"]:
                errors.append(f"Missing required animation: {anim}")
    
    return errors

def validate_asset_generation(char_data: Dict) -> List[str]:
    """Validate asset generation configuration"""
    errors = []
    
    if "assetGeneration" not in char_data:
        errors.append("Missing assetGeneration configuration")
        return errors
    
    asset_gen = char_data["assetGeneration"]
    
    # Required fields
    required_fields = ["basePrompt", "animationMappings", "generationSettings"]
    for field in required_fields:
        if field not in asset_gen:
            errors.append(f"assetGeneration missing required field: {field}")
    
    # Validate basePrompt
    if "basePrompt" in asset_gen:
        prompt = asset_gen["basePrompt"]
        if not prompt or len(prompt) < 50:
            errors.append("basePrompt should be at least 50 characters for quality generation")
        if "transparent background" not in prompt.lower():
            errors.append("basePrompt should include 'transparent background' for proper asset generation")
    
    # Validate animation mappings
    if "animationMappings" in asset_gen:
        mappings = asset_gen["animationMappings"]
        required_anims = ["idle", "talking", "happy", "sad"]
        
        for anim in required_anims:
            if anim not in mappings:
                errors.append(f"assetGeneration missing required animation mapping: {anim}")
            else:
                mapping = mappings[anim]
                if "promptModifier" not in mapping:
                    errors.append(f"Animation {anim} missing promptModifier")
                if "stateDescription" not in mapping:
                    errors.append(f"Animation {anim} missing stateDescription")
    
    # Validate generation settings
    if "generationSettings" in asset_gen:
        settings = asset_gen["generationSettings"]
        
        if "model" not in settings:
            errors.append("generationSettings missing model specification")
        elif settings["model"] not in ["flux1d", "flux1s", "sdxl"]:
            errors.append(f"Unknown model: {settings['model']}")
        
        if "artStyle" not in settings:
            errors.append("generationSettings missing artStyle")
        elif settings["artStyle"] not in ["anime", "pixel_art", "realistic", "cartoon", "chibi"]:
            errors.append(f"Unknown artStyle: {settings['artStyle']}")
        
        if "resolution" in settings:
            resolution = settings["resolution"]
            if "width" not in resolution or "height" not in resolution:
                errors.append("resolution missing width or height")
            elif resolution.get("width") != 128 or resolution.get("height") != 128:
                errors.append("resolution should be 128x128 for desktop companion compatibility")
    
    return errors

def validate_feature_consistency(char_data: Dict) -> List[str]:
    """Validate feature configuration consistency"""
    errors = []
    
    # Check dialog backend configuration
    if "dialogBackend" in char_data and char_data["dialogBackend"].get("enabled"):
        backend = char_data["dialogBackend"]
        if "backends" not in backend:
            errors.append("dialogBackend enabled but no backends configured")
        elif "defaultBackend" not in backend:
            errors.append("dialogBackend missing defaultBackend specification")
        elif backend["defaultBackend"] not in backend.get("backends", {}):
            errors.append("dialogBackend defaultBackend not found in backends configuration")
    
    # Check gift system configuration
    if "giftSystem" in char_data and char_data["giftSystem"].get("enabled"):
        gift_system = char_data["giftSystem"]
        if "preferences" not in gift_system:
            errors.append("giftSystem enabled but preferences not configured")
        if "inventorySettings" not in gift_system:
            errors.append("giftSystem enabled but inventorySettings not configured")
    
    # Check multiplayer configuration
    if "multiplayer" in char_data and char_data["multiplayer"].get("enabled"):
        multiplayer = char_data["multiplayer"]
        required_fields = ["networkID", "networkPersonality"]
        for field in required_fields:
            if field not in multiplayer:
                errors.append(f"multiplayer enabled but missing {field}")
    
    # Check battle system configuration
    if "battleSystem" in char_data and char_data["battleSystem"].get("enabled"):
        battle = char_data["battleSystem"]
        if "battleStats" not in battle:
            errors.append("battleSystem enabled but battleStats not configured")
        if "availableActions" not in battle:
            errors.append("battleSystem enabled but availableActions not configured")
    
    # Check stats consistency
    if "stats" in char_data:
        stats = char_data["stats"]
        for stat_name, stat_config in stats.items():
            if isinstance(stat_config, dict):
                if "max" in stat_config and "initial" in stat_config:
                    if stat_config["initial"] > stat_config["max"]:
                        errors.append(f"Stat {stat_name}: initial value exceeds maximum")
                
                if "criticalThreshold" in stat_config and "max" in stat_config:
                    if stat_config["criticalThreshold"] >= stat_config["max"]:
                        errors.append(f"Stat {stat_name}: criticalThreshold should be less than maximum")
    
    return errors

def validate_animation_paths(char_data: Dict, char_dir: Path) -> List[str]:
    """Validate that animation files exist (if not using asset generation)"""
    errors = []
    
    if "animations" not in char_data:
        return errors
    
    animations = char_data["animations"]
    
    for anim_name, anim_path in animations.items():
        if isinstance(anim_path, str):
            # Convert relative path to absolute
            if not anim_path.startswith('/'):
                full_path = char_dir / anim_path
            else:
                full_path = Path(anim_path)
            
            # Only check if asset generation is not configured or this animation isn't mapped
            asset_gen = char_data.get("assetGeneration", {})
            if not asset_gen or anim_name not in asset_gen.get("animationMappings", {}):
                if not full_path.exists():
                    errors.append(f"Animation file not found: {anim_path} (for {anim_name})")
                elif not full_path.suffix.lower() == '.gif':
                    errors.append(f"Animation file should be GIF format: {anim_path}")
    
    return errors

def validate_character_file(char_path: Path) -> Dict[str, Any]:
    """Validate a single character file"""
    result = {
        "character": char_path.parent.name,
        "path": str(char_path),
        "valid": True,
        "errors": [],
        "warnings": []
    }
    
    # JSON syntax validation
    syntax_valid, syntax_msg = validate_json_syntax(char_path)
    if not syntax_valid:
        result["valid"] = False
        result["errors"].append(syntax_msg)
        return result
    
    # Load character data
    try:
        with open(char_path, 'r', encoding='utf-8') as f:
            char_data = json.load(f)
    except Exception as e:
        result["valid"] = False
        result["errors"].append(f"Failed to load character data: {e}")
        return result
    
    # Required fields validation
    required_errors = validate_required_fields(char_data)
    result["errors"].extend(required_errors)
    
    # Asset generation validation
    asset_errors = validate_asset_generation(char_data)
    result["errors"].extend(asset_errors)
    
    # Feature consistency validation
    consistency_errors = validate_feature_consistency(char_data)
    result["errors"].extend(consistency_errors)
    
    # Animation path validation
    animation_errors = validate_animation_paths(char_data, char_path.parent)
    result["warnings"].extend(animation_errors)  # These are warnings since asset generation can create missing files
    
    # Additional validations
    if len(char_data.get("name", "")) == 0:
        result["errors"].append("Character name cannot be empty")
    
    if len(char_data.get("description", "")) < 10:
        result["warnings"].append("Character description should be more descriptive (at least 10 characters)")
    
    # Check for personality configuration
    if "personality" not in char_data or not char_data["personality"]:
        result["warnings"].append("Character missing personality configuration")
    
    # Set overall validity
    if result["errors"]:
        result["valid"] = False
    
    return result

def generate_validation_report(results: List[Dict[str, Any]]) -> str:
    """Generate validation report"""
    valid_count = sum(1 for r in results if r["valid"])
    total_count = len(results)
    
    report = []
    report.append("=" * 80)
    report.append("CHARACTER VALIDATION REPORT")
    report.append("=" * 80)
    report.append(f"Total Characters: {total_count}")
    report.append(f"Valid Characters: {valid_count}")
    report.append(f"Invalid Characters: {total_count - valid_count}")
    report.append(f"Validation Success Rate: {(valid_count/total_count)*100:.1f}%")
    report.append("")
    
    # Valid characters
    valid_chars = [r for r in results if r["valid"]]
    if valid_chars:
        report.append("VALID CHARACTERS")
        report.append("-" * 40)
        for result in valid_chars:
            report.append(f"✓ {result['character']}")
            if result["warnings"]:
                for warning in result["warnings"]:
                    report.append(f"  ⚠ {warning}")
        report.append("")
    
    # Invalid characters
    invalid_chars = [r for r in results if not r["valid"]]
    if invalid_chars:
        report.append("INVALID CHARACTERS")
        report.append("-" * 40)
        for result in invalid_chars:
            report.append(f"✗ {result['character']}")
            for error in result["errors"]:
                report.append(f"  ✗ {error}")
            for warning in result["warnings"]:
                report.append(f"  ⚠ {warning}")
            report.append("")
    
    # Summary by error type
    all_errors = []
    all_warnings = []
    for result in results:
        all_errors.extend(result["errors"])
        all_warnings.extend(result["warnings"])
    
    if all_errors:
        report.append("ERROR SUMMARY")
        report.append("-" * 40)
        error_counts = {}
        for error in all_errors:
            error_type = error.split(":")[0] if ":" in error else error
            error_counts[error_type] = error_counts.get(error_type, 0) + 1
        
        for error_type, count in sorted(error_counts.items(), key=lambda x: x[1], reverse=True):
            report.append(f"{error_type}: {count} occurrences")
        report.append("")
    
    if all_warnings:
        report.append("WARNING SUMMARY")
        report.append("-" * 40)
        warning_counts = {}
        for warning in all_warnings:
            warning_type = warning.split(":")[0] if ":" in warning else warning
            warning_counts[warning_type] = warning_counts.get(warning_type, 0) + 1
        
        for warning_type, count in sorted(warning_counts.items(), key=lambda x: x[1], reverse=True):
            report.append(f"{warning_type}: {count} occurrences")
        report.append("")
    
    report.append("=" * 80)
    
    return "\n".join(report)

def main():
    if len(sys.argv) != 2:
        print("Usage: python validate_characters.py <characters_directory>")
        sys.exit(1)
    
    characters_dir = Path(sys.argv[1])
    
    if not characters_dir.exists():
        print(f"Error: Directory {characters_dir} does not exist")
        sys.exit(1)
    
    print("Validating character configurations...")
    
    results = []
    
    # Process all character files
    for char_dir in characters_dir.iterdir():
        if char_dir.is_dir():
            char_file = char_dir / "character.json"
            if char_file.exists():
                result = validate_character_file(char_file)
                results.append(result)
                
                # Print real-time progress
                if result["valid"]:
                    print(f"✓ {result['character']}")
                else:
                    print(f"✗ {result['character']} ({len(result['errors'])} errors)")
    
    # Generate and save report
    report = generate_validation_report(results)
    
    with open("CHARACTER_VALIDATION_REPORT.md", "w") as f:
        f.write("# Character Validation Report\n\n")
        f.write("```\n")
        f.write(report)
        f.write("\n```\n")
    
    print("\n" + report)
    print(f"\nValidation report saved to CHARACTER_VALIDATION_REPORT.md")
    
    # Exit with error code if any characters are invalid
    invalid_count = len([r for r in results if not r["valid"]])
    if invalid_count > 0:
        sys.exit(1)

if __name__ == "__main__":
    main()
