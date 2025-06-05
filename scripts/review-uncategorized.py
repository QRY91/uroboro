#!/usr/bin/env python3
"""
Review uncategorized captures in batches for manual categorization
"""

import os
import re
import glob
import sys
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
    """Same classification logic as organize-captures.py"""
    PROJECT_RULES = {
        'uroboro': [
            r'uroboro', r'landing page', r'cross-platform', r'XDG', r'patch.notes',
            r'CLI.*uro', r'internal/', r'GetDataDir', r'common/dirs', r'go build.*uroboro',
            r'README.*uro', r'three commands', r'north star', r'more gun principle',
            r'capture.*publish.*status', r'devlog', r'WebSocket.*uro',
            r'content pipeline', r'capture script', r'testing capture',
            r'unified CLI interface', r'privacy-first tracking', r'VHS demo', r'uro short alias', r'git hook', r'system installation'
        ],
        'payment-system': [
            r'payment', r'billing', r'stripe', r'paypal', r'invoice', r'subscription',
            r'checkout', r'credit card', r'transaction'
        ],
        'collaboration': [
            r'real-time collaboration', r'WebSocket(?!.*uro)', r'socket\.io',
            r'collaboration features', r'shared workspace', r'team features'
        ],
        'auth-system': [
            r'OAuth2', r'JWT', r'authentication', r'login', r'session', r'user auth',
            r'security', r'token'
        ],
        'general-dev': [
            r'memory leak', r'performance', r'optimization', r'bug fix',
            r'connection pooling(?!.*uro)', r'rate limiting(?!.*uro)'
        ],
        'qryzone': [
            r'qry\.zone', r'qryzone', r'Cloudflare', r'DNS migration', r'nameserver', 
            r'Vercel records', r'email routing', r'deployment successful', r'site live'
        ],
        'doggowoof': [
            r'DoggoWoof', r'doggowoof', r'DOGGOWOOF', r'Genesis of DoggoWoof', 
            r'AI Oracle', r'BIG LABRADO', r'REBRAND COMPLETE', r'Cobra CLI'
        ],
        'slopsquid': [
            r'slopsquid', r'SlopSquid', r'SLOPSQUID',
            r'browser extension', r'AI detection', r'SlopSquid', r'squid', r'hot pink', r'manifest V3', r'chrome extension', r'AI slop', r'detection algorithm'
        ],
        'panopticron-report': [
            r'panopticron', r'academic-style', r'report structure', r'SDG reflection', 
            r'voice-ai', r'academic paper', r'final-version', r'Q&A system', r'presentation-prep',
            r'academic-integrity', r'psychology', r'voice-to-writing', r'deadline'
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

def get_uncategorized_captures():
    """Get all uncategorized captures"""
    data_dir = os.path.expanduser("~/.local/share/uroboro/daily")
    md_files = glob.glob(os.path.join(data_dir, "*.md"))
    
    uncategorized = []
    for md_file in md_files:
        captures = extract_captures_from_file(md_file)
        for capture in captures:
            if classify_capture(capture['content']) == 'uncategorized':
                uncategorized.append(capture)
    
    return sorted(uncategorized, key=lambda x: x['timestamp'])

def show_batch(captures, start_idx, batch_size=10):
    """Show a batch of captures for review"""
    end_idx = min(start_idx + batch_size, len(captures))
    
    print(f"\n📋 Captures {start_idx + 1}-{end_idx} of {len(captures)}")
    print("=" * 60)
    
    for i in range(start_idx, end_idx):
        capture = captures[i]
        date = capture['timestamp'][:10]
        preview = capture['content'][:100].replace('\n', ' ').strip()
        
        print(f"\n[{i + 1:2d}] {date} ({capture['file']})")
        print(f"    {preview}...")
        
        # Show more context if needed
        if len(capture['content']) > 100:
            print(f"    [...{len(capture['content']) - 100} more chars]")
    
    return end_idx

def main():
    if len(sys.argv) > 1:
        batch_size = int(sys.argv[1])
    else:
        batch_size = 10
    
    print("🔍 Loading uncategorized captures...")
    uncategorized = get_uncategorized_captures()
    
    if not uncategorized:
        print("✅ No uncategorized captures found!")
        return
    
    print(f"📊 Found {len(uncategorized)} uncategorized captures")
    print(f"📄 Showing in batches of {batch_size}")
    
    current_idx = 0
    
    while current_idx < len(uncategorized):
        current_idx = show_batch(uncategorized, current_idx, batch_size)
        
        if current_idx < len(uncategorized):
            print(f"\n⏭️  Continue to next batch? (y/N/q/jump to index): ", end="")
            response = input().strip().lower()
            
            if response in ['q', 'quit']:
                break
            elif response.isdigit():
                jump_to = int(response) - 1
                if 0 <= jump_to < len(uncategorized):
                    current_idx = jump_to
                else:
                    print(f"❌ Invalid index. Must be 1-{len(uncategorized)}")
            elif response not in ['y', 'yes', '']:
                break
        else:
            print(f"\n✅ Finished reviewing all {len(uncategorized)} uncategorized captures")
    
    print(f"\n🎯 To categorize captures:")
    print(f"1. Note which ones belong to specific projects")
    print(f"2. Add keywords to PROJECT_RULES in scripts/organize-captures.py")
    print(f"3. Or manually edit the daily files to add 'Project: name' lines")
    print(f"4. Re-run: python3 scripts/organize-captures.py")

if __name__ == "__main__":
    main() 