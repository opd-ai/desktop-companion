#!/usr/bin/env python3
"""
Character Feature Audit Script
Analyzes all character.json files for feature coverage and compatibility
"""

import json
import os
import sys
from pathlib import Path
from typing import Dict, List, Set, Any
from collections import defaultdict

def load_character_files(characters_dir: str) -> Dict[str, Dict]:
    """Load all character.json files from the characters directory"""
    characters = {}
    characters_path = Path(characters_dir)
    
    # Find all character.json files
    for character_dir in characters_path.iterdir():
        if character_dir.is_dir():
            character_file = character_dir / "character.json"
            if character_file.exists():
                try:
                    with open(character_file, 'r', encoding='utf-8') as f:
                        data = json.load(f)
                        characters[character_dir.name] = data
                except json.JSONDecodeError as e:
                    print(f"Error loading {character_file}: {e}")
                except Exception as e:
                    print(f"Unexpected error loading {character_file}: {e}")
    
    return characters

def analyze_feature_coverage(characters: Dict[str, Dict]) -> Dict[str, Any]:
    """Analyze feature coverage across all characters"""
    
    # Define all possible features
    CORE_FEATURES = [
        "dialogBackend", "giftSystem", "multiplayer", "newsFeatures", 
        "battleSystem", "assetGeneration"
    ]
    
    ROMANCE_FEATURES = [
        "personality", "romanceDialogs", "romanceEvents"
    ]
    
    INTERACTIVE_FEATURES = [
        "generalEvents", "randomEvents", "progression", "interactions"
    ]
    
    BASIC_FEATURES = [
        "animations", "dialogs", "behavior", "stats", "gameRules"
    ]
    
    ALL_FEATURES = CORE_FEATURES + ROMANCE_FEATURES + INTERACTIVE_FEATURES + BASIC_FEATURES
    
    # Analyze each character
    feature_matrix = {}
    missing_features = defaultdict(list)
    
    for char_name, char_data in characters.items():
        feature_matrix[char_name] = {}
        
        for feature in ALL_FEATURES:
            has_feature = feature in char_data and char_data[feature]
            feature_matrix[char_name][feature] = has_feature
            
            if not has_feature:
                missing_features[char_name].append(feature)
    
    # Calculate statistics
    feature_stats = {}
    for feature in ALL_FEATURES:
        count = sum(1 for char_features in feature_matrix.values() if char_features.get(feature, False))
        total = len(characters)
        feature_stats[feature] = {
            "count": count,
            "total": total,
            "percentage": (count / total) * 100 if total > 0 else 0
        }
    
    return {
        "characters": list(characters.keys()),
        "feature_matrix": feature_matrix,
        "missing_features": dict(missing_features),
        "feature_stats": feature_stats,
        "total_characters": len(characters)
    }

def analyze_asset_generation(characters: Dict[str, Dict]) -> Dict[str, Any]:
    """Analyze asset generation configuration across characters"""
    
    asset_analysis = {
        "has_asset_generation": [],
        "missing_asset_generation": [],
        "incomplete_configurations": [],
        "animation_mappings": defaultdict(list)
    }
    
    for char_name, char_data in characters.items():
        if "assetGeneration" in char_data and char_data["assetGeneration"]:
            asset_gen = char_data["assetGeneration"]
            asset_analysis["has_asset_generation"].append(char_name)
            
            # Check for complete configuration
            required_fields = ["basePrompt", "animationMappings", "generationSettings"]
            missing_fields = [field for field in required_fields if field not in asset_gen]
            
            if missing_fields:
                asset_analysis["incomplete_configurations"].append({
                    "character": char_name,
                    "missing_fields": missing_fields
                })
            
            # Analyze animation mappings
            if "animationMappings" in asset_gen:
                for anim_name in asset_gen["animationMappings"].keys():
                    asset_analysis["animation_mappings"][anim_name].append(char_name)
                    
        else:
            asset_analysis["missing_asset_generation"].append(char_name)
    
    return asset_analysis

def generate_report(analysis: Dict[str, Any], asset_analysis: Dict[str, Any]) -> str:
    """Generate a comprehensive audit report"""
    
    report = []
    report.append("=" * 80)
    report.append("CHARACTER FEATURE AUDIT REPORT")
    report.append("=" * 80)
    report.append(f"Total Characters Analyzed: {analysis['total_characters']}")
    report.append("")
    
    # Feature Coverage Summary
    report.append("FEATURE COVERAGE SUMMARY")
    report.append("-" * 40)
    
    for feature, stats in sorted(analysis['feature_stats'].items(), key=lambda x: x[1]['percentage'], reverse=True):
        count = stats['count']
        total = stats['total']
        percentage = stats['percentage']
        report.append(f"{feature:20} {count:3}/{total:3} ({percentage:5.1f}%)")
    
    report.append("")
    
    # Missing Features by Character
    report.append("MISSING FEATURES BY CHARACTER")
    report.append("-" * 40)
    
    for char_name, missing in analysis['missing_features'].items():
        if missing:
            report.append(f"{char_name}:")
            for feature in missing:
                report.append(f"  - {feature}")
            report.append("")
    
    # Asset Generation Analysis
    report.append("ASSET GENERATION ANALYSIS")
    report.append("-" * 40)
    report.append(f"Characters with Asset Generation: {len(asset_analysis['has_asset_generation'])}")
    for char in asset_analysis['has_asset_generation']:
        report.append(f"  ✓ {char}")
    
    report.append(f"\nCharacters Missing Asset Generation: {len(asset_analysis['missing_asset_generation'])}")
    for char in asset_analysis['missing_asset_generation']:
        report.append(f"  ✗ {char}")
    
    if asset_analysis['incomplete_configurations']:
        report.append(f"\nIncomplete Asset Generation Configurations:")
        for config in asset_analysis['incomplete_configurations']:
            report.append(f"  ⚠ {config['character']}: missing {', '.join(config['missing_fields'])}")
    
    # Most Common Animation Mappings
    report.append(f"\nAnimation Mapping Coverage:")
    for anim_name, chars in sorted(asset_analysis['animation_mappings'].items(), key=lambda x: len(x[1]), reverse=True):
        report.append(f"  {anim_name:15} {len(chars):2} characters")
    
    report.append("")
    report.append("=" * 80)
    
    return "\n".join(report)

def main():
    if len(sys.argv) != 2:
        print("Usage: python character_audit.py <characters_directory>")
        sys.exit(1)
    
    characters_dir = sys.argv[1]
    
    if not os.path.exists(characters_dir):
        print(f"Error: Directory {characters_dir} does not exist")
        sys.exit(1)
    
    print("Loading character files...")
    characters = load_character_files(characters_dir)
    
    if not characters:
        print("No character files found!")
        sys.exit(1)
    
    print(f"Analyzing {len(characters)} characters...")
    
    # Perform analysis
    analysis = analyze_feature_coverage(characters)
    asset_analysis = analyze_asset_generation(characters)
    
    # Generate and save report
    report = generate_report(analysis, asset_analysis)
    
    # Save to file
    with open("CHARACTER_FEATURE_AUDIT_REPORT.md", "w") as f:
        f.write("# Character Feature Audit Report\n\n")
        f.write("```\n")
        f.write(report)
        f.write("\n```\n")
        f.write("\n## Recommendations\n\n")
        f.write("### Priority 1: Asset Generation Configuration\n")
        f.write("Configure `assetGeneration` for all characters missing this feature to enable gif-generator compatibility.\n\n")
        f.write("### Priority 2: Core Feature Standardization\n")
        f.write("Add missing core features (dialogBackend, giftSystem, multiplayer, newsFeatures, battleSystem) to achieve feature parity.\n\n")
        f.write("### Priority 3: Enhanced Interactivity\n")
        f.write("Ensure all characters have generalEvents, interactions, and progression systems configured.\n\n")
    
    print(report)
    print(f"\nReport saved to CHARACTER_FEATURE_AUDIT_REPORT.md")

if __name__ == "__main__":
    main()
