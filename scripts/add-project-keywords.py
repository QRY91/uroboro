#!/usr/bin/env python3
"""
Helper to add new keywords to project classification
"""

import sys
import re

def add_keywords_to_rules(project_name, keywords):
    """Add keywords to the PROJECT_RULES in organize-captures.py"""
    
    organize_file = 'scripts/organize-captures.py'
    
    with open(organize_file, 'r') as f:
        content = f.read()
    
    # Find the project rules section
    if project_name not in content:
        print(f"❌ Project '{project_name}' not found in rules")
        print("Available projects: uroboro, payment-system, collaboration, auth-system, general-dev")
        return False
    
    # Add keywords to the specific project
    pattern = f"'{project_name}': \\[(.*?)\\]"
    match = re.search(pattern, content, re.DOTALL)
    
    if not match:
        print(f"❌ Could not find {project_name} rules pattern")
        return False
    
    current_rules = match.group(1)
    
    # Add new keywords
    new_keywords = [f"r'{keyword}'" for keyword in keywords]
    keywords_str = ', '.join(new_keywords)
    
    # Insert before the closing bracket
    updated_rules = current_rules.rstrip() + f',\n        {keywords_str}'
    
    # Replace in content
    new_content = content.replace(
        f"'{project_name}': [{current_rules}]",
        f"'{project_name}': [{updated_rules}]"
    )
    
    # Write back
    with open(organize_file, 'w') as f:
        f.write(new_content)
    
    print(f"✅ Added keywords to {project_name}: {keywords}")
    return True

def main():
    if len(sys.argv) < 3:
        print("Usage: python3 scripts/add-project-keywords.py PROJECT_NAME keyword1 keyword2 ...")
        print("Example: python3 scripts/add-project-keywords.py uroboro 'VSCode extension' 'demo recording'")
        print("Projects: uroboro, payment-system, collaboration, auth-system, general-dev")
        return
    
    project = sys.argv[1]
    keywords = sys.argv[2:]
    
    if add_keywords_to_rules(project, keywords):
        print("\n🔄 Re-run organization:")
        print("python3 scripts/organize-captures.py")

if __name__ == "__main__":
    main() 