#!/usr/bin/env python3
"""
Show uncategorized captures in batches for manual categorization
"""

import os
import re
import glob
from datetime import datetime

def extract_captures_from_file(filepath):
    """Extract individual captures from a daily file"""
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()
    
    captures = []
    entries = re.split(r'\n## (\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})', content)
    
    for i in range(1, len(entries), 2):
        if i + 1 < len(entries):
            timestamp = entries[i]
            capture_content = entries[i + 1].strip()
            if capture_content:
                captures.append({
                    'timestamp': timestamp,
                    'content': capture_content,
                    'file': os.path.basename(filepath)
                })
    
    return captures

def classify_capture(content):
    """Basic classification to identify already categorized captures"""
    PROJECT_RULES = {
        'uroboro': [
            r'uroboro', r'landing page', r'cross-platform', r'XDG', r'patch.notes',
            r'CLI.*uro', r'internal/', r'GetDataDir', r'common/dirs', r'go build.*uroboro',
            r'README.*uro', r'three commands', r'north star', r'more gun principle',
            r'capture.*publish.*status', r'devlog', r'WebSocket.*uro'
        ],
        'qryzone': [
            r'qry\.zone', r'qryzone', r'qry-zone', r'blog.*MDX', r'CSS grid.*responsive'
        ],
        'doggowoof': [
            r'doggowoof', r'DoggoWoof', r'DOGGOWOOF', r'BIG LABRADO', r'AI Oracle'
        ]
    }
    
    content_lower = content.lower()
    scores = {}
    for project, keywords in PROJECT_RULES.items():
        score = 0
        for keyword in keywords:
            matches = len(re.findall(keyword, content_lower, re.IGNORECASE))
            score += matches
        scores[project] = score
    
    if max(scores.values()) > 0:
        return max(scores, key=scores.get)
    return 'uncategorized'

def main():
    data_dir = os.path.expanduser("~/.local/share/uroboro/daily")
    md_files = glob.glob(os.path.join(data_dir, "*.md"))
    
    uncategorized = []
    for md_file in md_files:
        captures = extract_captures_from_file(md_file)
        for capture in captures:
            if classify_capture(capture['content']) == 'uncategorized':
                uncategorized.append(capture)
    
    uncategorized = sorted(uncategorized, key=lambda x: x['timestamp'])
    
    print(f"Found {len(uncategorized)} uncategorized captures")
    print("\nBATCH 1: Captures 1-25")
    print("=" * 60)
    
    for i in range(min(25, len(uncategorized))):
        capture = uncategorized[i]
        date = capture['timestamp'][:10]
        preview = capture['content'][:120].replace('\n', ' ').strip()
        print(f"[{i+1:2d}] {date} - {preview}...")

if __name__ == "__main__":
    main() 