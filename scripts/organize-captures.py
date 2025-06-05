#!/usr/bin/env python3
"""
Retroactively organize uroboro captures by project using content analysis
"""

import os
import re
import glob
from pathlib import Path
from datetime import datetime

# Project classification rules
PROJECT_RULES = {
    'uroboro': [
        r'uroboro', r'landing page', r'cross-platform', r'XDG', r'patch.notes',
        r'CLI.*uro', r'internal/', r'GetDataDir', r'common/dirs', r'go build.*uroboro',
        r'README.*uro', r'three commands', r'north star', r'more gun principle',
        r'capture.*publish.*status', r'devlog', r'WebSocket.*uro',
        r'content pipeline', r'capture script', r'testing capture',
        r'unified CLI interface', r'privacy-first tracking', r'VHS demo', r'uro short alias', r'git hook', r'system installation',
        r'git commit', r'auto-captured', r'font test', r'demo tapes', r'tickertape', r'thumbnail readability',
        r'Add core functionality', r'Initial commit', r'Test auto-capture hook'],
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
        r'browser extension', r'AI detection', r'SlopSquid', r'squid', r'hot pink', r'manifest V3', r'chrome extension', r'AI slop', r'detection algorithm'],
    'panopticron-report': [
        r'panopticron', r'academic-style', r'report structure', r'SDG reflection', 
        r'voice-ai', r'academic paper', r'final-version', r'Q&A system', r'presentation-prep',
        r'academic-integrity', r'psychology', r'voice-to-writing', r'deadline'
    ]
}

def classify_capture(content):
    """Classify a capture based on its content"""
    content_lower = content.lower()
    
    # Score each project based on keyword matches
    scores = {}
    for project, keywords in PROJECT_RULES.items():
        score = 0
        for keyword in keywords:
            matches = len(re.findall(keyword, content_lower, re.IGNORECASE))
            score += matches
        scores[project] = score
    
    # Return the highest scoring project (if score > 0)
    if max(scores.values()) > 0:
        return max(scores, key=scores.get)
    return 'uncategorized'

def extract_captures_from_file(filepath):
    """Extract individual captures from a daily file"""
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()
    
    # Split by timestamp headers
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
                    'file': filepath
                })
    
    return captures

def organize_captures():
    """Main function to organize all captures"""
    data_dir = os.path.expanduser("~/.local/share/uroboro/daily")
    
    if not os.path.exists(data_dir):
        print(f"❌ Data directory not found: {data_dir}")
        return
    
    print(f"🔍 Analyzing captures in {data_dir}")
    
    # Find all markdown files
    md_files = glob.glob(os.path.join(data_dir, "*.md"))
    
    if not md_files:
        print("❌ No capture files found")
        return
    
    # Organize captures by project
    projects = {}
    total_captures = 0
    
    for md_file in md_files:
        print(f"📄 Processing {os.path.basename(md_file)}")
        captures = extract_captures_from_file(md_file)
        
        for capture in captures:
            project = classify_capture(capture['content'])
            
            if project not in projects:
                projects[project] = []
            
            projects[project].append(capture)
            total_captures += 1
    
    # Print summary
    print(f"\n📊 Analysis Results ({total_captures} total captures):")
    print("=" * 50)
    
    for project, captures in sorted(projects.items()):
        count = len(captures)
        percentage = (count / total_captures) * 100
        print(f"{project:15} {count:3d} captures ({percentage:5.1f}%)")
    
    print("\n🔧 Uroboro Project Captures:")
    print("=" * 30)
    
    if 'uroboro' in projects:
        uroboro_captures = projects['uroboro']
        for capture in sorted(uroboro_captures, key=lambda x: x['timestamp'])[-10:]:
            date = capture['timestamp'][:10]  # Just the date part
            preview = capture['content'][:80].replace('\n', ' ')
            print(f"{date} | {preview}...")
    
    return projects

def create_project_specific_files(projects):
    """Create separate files for each project"""
    output_dir = os.path.expanduser("~/.local/share/uroboro/by-project")
    os.makedirs(output_dir, exist_ok=True)
    
    for project, captures in projects.items():
        if project == 'uncategorized':
            continue
            
        output_file = os.path.join(output_dir, f"{project}.md")
        
        with open(output_file, 'w', encoding='utf-8') as f:
            f.write(f"# {project.title()} Development Log\n\n")
            
            for capture in sorted(captures, key=lambda x: x['timestamp']):
                f.write(f"## {capture['timestamp']}\n\n")
                f.write(f"{capture['content']}\n\n")
        
        print(f"📝 Created {output_file} with {len(captures)} captures")

if __name__ == "__main__":
    print("🚀 Uroboro Capture Organizer")
    print("Retroactively categorizing captures by project...\n")
    
    projects = organize_captures()
    
    if projects:
        print(f"\n📁 Create project-specific files? (y/N): ", end="")
        response = input().strip().lower()
        
        if response in ['y', 'yes']:
            create_project_specific_files(projects)
            print(f"\n✅ Project files created in ~/.local/share/uroboro/by-project/")
        
        print(f"\n🎯 Next steps:")
        print(f"1. Review the categorization above")
        print(f"2. Use project-specific files for focused devlogs")
        print(f"3. Going forward, use: uroboro capture 'work' --project 'uroboro'") 