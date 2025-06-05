#!/usr/bin/env python3
"""
Generate patch notes from organized uroboro captures
"""

import os
import re
from datetime import datetime

def extract_key_captures():
    """Extract key uroboro development captures"""
    uroboro_file = os.path.expanduser("~/.local/share/uroboro/by-project/uroboro.md")
    
    if not os.path.exists(uroboro_file):
        print("❌ Uroboro capture file not found")
        return []
    
    with open(uroboro_file, 'r') as f:
        content = f.read()
    
    # Extract captures with meaningful content (filter out test commits)
    captures = []
    entries = re.split(r'\n## (\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})', content)
    
    for i in range(1, len(entries), 2):
        if i + 1 < len(entries):
            timestamp = entries[i]
            capture_content = entries[i + 1].strip()
            
            # Filter out noise (auto-commits, tests, etc.)
            if (capture_content and 
                not re.search(r'Initial commit|Add core functionality|Test auto-capture hook', capture_content) and
                not re.search(r'font test|Smaller font|Larger font', capture_content, re.IGNORECASE) and
                len(capture_content) > 50):  # Meaningful content only
                
                captures.append({
                    'timestamp': timestamp,
                    'content': capture_content,
                    'date': timestamp[:10]  # Just the date part
                })
    
    return sorted(captures, key=lambda x: x['timestamp'], reverse=True)

def categorize_captures(captures):
    """Categorize captures by type of work"""
    categories = {
        'cross-platform': [],
        'cli-interface': [],
        'git-integration': [],
        'demos-docs': [],
        'landing-page': [],
        'voice-features': [],
        'testing-ci': [],
        'templates': [],
        'other': []
    }
    
    for capture in captures:
        content = capture['content'].lower()
        
        if 'cross-platform' in content or 'windows' in content or 'xdg' in content:
            categories['cross-platform'].append(capture)
        elif 'cli' in content or 'unified' in content or 'command' in content:
            categories['cli-interface'].append(capture)
        elif 'git' in content or 'hook' in content or 'commit' in content:
            categories['git-integration'].append(capture)
        elif 'vhs' in content or 'demo' in content or 'documentation' in content:
            categories['demos-docs'].append(capture)
        elif 'landing page' in content or 'website' in content or 'uroboro.dev' in content:
            categories['landing-page'].append(capture)
        elif 'voice' in content or 'style' in content:
            categories['voice-features'].append(capture)
        elif 'test' in content or 'ci' in content or 'github' in content:
            categories['testing-ci'].append(capture)
        elif 'template' in content:
            categories['templates'].append(capture)
        else:
            categories['other'].append(capture)
    
    return categories

def format_for_patch_notes(categories):
    """Format categorized captures into patch notes format"""
    patch_notes = []
    
    category_titles = {
        'cross-platform': '🔧 Cross-Platform Support',
        'cli-interface': '⚡ CLI Interface Modernization', 
        'git-integration': '🔗 Git Integration',
        'demos-docs': '📹 Documentation & Demos',
        'landing-page': '🌐 Landing Page & Branding',
        'voice-features': '🎭 Voice & Style Features',
        'testing-ci': '🧪 Testing & CI Infrastructure',
        'templates': '📝 Project Templates',
        'other': '🔄 General Improvements'
    }
    
    for category, title in category_titles.items():
        if categories[category]:
            patch_notes.append(f"\n## {title}\n")
            
            # Take the most recent/important capture from each category
            key_capture = categories[category][0]
            content = key_capture['content']
            
            # Extract key points
            if 'Tags:' in content:
                content = content.split('Tags:')[1] if content.count('Tags:') == 1 else content.split('Tags:')[0]
            
            # Clean up content
            content = re.sub(r'Project: uroboro', '', content).strip()
            content = re.sub(r'\n+', ' ', content)
            
            # Format as patch note
            patch_notes.append(f"**{key_capture['date']}**: {content[:300]}...")
            
            if len(categories[category]) > 1:
                patch_notes.append(f"\n*Plus {len(categories[category])-1} more related updates*")
    
    return '\n'.join(patch_notes)

def main():
    print("🚀 Generating uroboro patch notes from captures...")
    
    captures = extract_key_captures()
    print(f"📊 Found {len(captures)} meaningful uroboro captures")
    
    if not captures:
        print("❌ No captures found")
        return
    
    categories = categorize_captures(captures)
    
    print("\n📋 Capture breakdown:")
    for category, items in categories.items():
        if items:
            print(f"  {category}: {len(items)} captures")
    
    patch_notes = format_for_patch_notes(categories)
    
    print("\n" + "="*60)
    print("GENERATED PATCH NOTES")
    print("="*60)
    print(patch_notes)
    print("\n" + "="*60)
    
    # Save to file
    with open('/tmp/uroboro-patch-notes.md', 'w') as f:
        f.write(f"# Uroboro Development Updates\n")
        f.write(f"*Generated from {len(captures)} uroboro development captures*\n")
        f.write(patch_notes)
        f.write(f"\n\n---\n*Generated by uroboro from its own development captures*")
    
    print(f"💾 Saved to /tmp/uroboro-patch-notes.md")

if __name__ == "__main__":
    main() 