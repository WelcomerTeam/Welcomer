#!/usr/bin/env python3
"""
Generate fonts.json from Google Fonts metadata API.
"""

import json
import urllib.request
from collections import OrderedDict

def fetch_google_fonts_metadata():
    """Fetch the Google Fonts metadata from the official API."""
    url = "https://fonts.google.com/metadata/fonts"
    print(f"Fetching Google Fonts metadata from {url}...")
    
    try:
        with urllib.request.urlopen(url, timeout=30) as response:
            data = json.loads(response.read().decode('utf-8'))
        return data
    except Exception as e:
        print(f"Error fetching metadata: {e}")
        return None

def normalize_weight(weight_str):
    """Convert weight string to standardized format."""
    if weight_str == "400":
        return "regular"
    elif weight_str == "700":
        return "bold"
    else:
        return weight_str

def generate_fonts_dict(metadata):
    """Generate the fonts dictionary from metadata."""
    fonts_dict = OrderedDict()
    
    if not metadata or 'familyMetadataList' not in metadata:
        print("Invalid metadata structure")
        return fonts_dict
    
    family_list = metadata['familyMetadataList']
    print(f"Processing {len(family_list)} font families...")
    
    for family_meta in family_list:
        family_name = family_meta.get('family')
        if not family_name:
            continue
        
        # Extract available weights
        weights = {}
        font_variants = family_meta.get('fonts', {})
        
        for weight_key in sorted(font_variants.keys()):
            # Remove italic suffix if present
            base_weight = weight_key.rstrip('i')
            
            # Add the weight if not already added
            if base_weight not in weights:
                normalized = normalize_weight(base_weight)
                weights[normalized] = base_weight
        
        # Ensure there's at least a regular weight
        if not weights:
            weights['regular'] = '400'
        
        # Determine default weight
        default_weight = 'regular' if 'regular' in weights else list(weights.keys())[0]
        
        # Create font entry
        fonts_dict[family_name] = {
            "name": family_name,
            "defaultWeight": default_weight,
            "weights": weights
        }
    
    return fonts_dict

def format_fonts_json(fonts_dict):
    """Format the fonts dictionary as valid JSON."""
    # Use json.dumps for proper JSON formatting
    return json.dumps(fonts_dict, indent=4)

def format_fonts_go(fonts_dict):
    """Format the fonts dictionary as Go code."""
    lines = ["package service", "", "var Fonts = map[string]Font{"]
    
    for family_name, font_data in fonts_dict.items():
        # Escape quotes in family name for Go string literal
        escaped_name = family_name.replace('"', '\\"')
        
        lines.append(f'\t"{escaped_name}": {{')
        lines.append(f'\t\tname:          "{escaped_name}",')
        lines.append(f'\t\tdefaultWeight: "{font_data["defaultWeight"]}",')
        lines.append('\t\tweights: map[string]string{')
        
        # Sort weights for consistent output
        weights = font_data['weights']
        for weight_name in sorted(weights.keys(), key=lambda x: (x != 'regular', x)):
            weight_value = weights[weight_name]
            lines.append(f'\t\t\t"{weight_name}": "{weight_value}",')
        
        lines.append('\t\t},')
        lines.append('\t},')
        lines.append('')
    
    # Add web-safe fonts
    web_safe_fonts = [
        # ("Arial", "regular", {"regular": "400", "bold": "700"}),
        # ("Verdana", "regular", {"regular": "400", "bold": "700"}),
        # ("Tahoma", "regular", {"regular": "400", "bold": "700"}),
        # ("Trebuchet MS", "regular", {"regular": "400", "bold": "700"}),
        # ("Times New Roman", "regular", {"regular": "400", "bold": "700"}),
        # ("Georgia", "regular", {"regular": "400", "bold": "700"}),
        # ("Garamond", "regular", {"regular": "400", "bold": "700"}),
        # ("Courier New", "regular", {"regular": "400", "bold": "700"}),
    ]
    
    lines.append('\t// web safe fonts')
    for name, default_weight, weights in web_safe_fonts:
        lines.append(f'\t"{name}": {{')
        lines.append(f'\t\tname:          "{name}",')
        lines.append('\t\twebsafe:       true,')
        lines.append(f'\t\tdefaultWeight: "{default_weight}",')
        lines.append('\t\tweights: map[string]string{')
        for weight_name, weight_value in weights.items():
            lines.append(f'\t\t\t"{weight_name}": "{weight_value}",')
        lines.append('\t\t},')
        lines.append('\t},')
        lines.append('')
    
    lines.append("}")
    return '\n'.join(lines)

def main():
    import os
    
    # Get script directory
    script_dir = os.path.dirname(os.path.abspath(__file__))
    
    # Fetch metadata
    metadata = fetch_google_fonts_metadata()
    if not metadata:
        return
    
    # Generate fonts dictionary
    fonts_dict = generate_fonts_dict(metadata)
    print(f"Generated {len(fonts_dict)} font entries")

    # Write JSON file
    json_output = format_fonts_json(fonts_dict)
    json_path = os.path.join(script_dir, "fonts.json")
    try:
        with open(json_path, 'w') as f:
            f.write(json_output)
        print(f"Successfully wrote {json_path}")
    except Exception as e:
        print(f"Error writing JSON file: {e}")
        return

    # Write JSON file
    json_output = format_fonts_json(fonts_dict)
    json_path = os.path.join(script_dir, "website/app/src/fonts.json")
    try:
        with open(json_path, 'w') as f:
            f.write(json_output)
        print(f"Successfully wrote {json_path}")
    except Exception as e:
        print(f"Error writing JSON file: {e}")
        return
    
    # Write Go file
    go_output = format_fonts_go(fonts_dict)
    go_path = os.path.join(script_dir, "welcomer-images-next/service/fonts_generated.go")
    try:
        with open(go_path, 'w') as f:
            f.write(go_output)
        print(f"Successfully wrote {go_path}")
    except Exception as e:
        print(f"Error writing Go file: {e}")
        return

if __name__ == "__main__":
    main()
