#!/usr/bin/env python3
"""
Git Integration for uroboro
Automatically captures development insights from git commits
"""

import subprocess
import json
import os
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Optional


class GitIntegration:
    def __init__(self, repo_path: str = "."):
        self.repo_path = Path(repo_path)
        self.is_git_repo = self._check_git_repo()
    
    def _check_git_repo(self) -> bool:
        """Check if current directory is a git repository"""
        try:
            result = subprocess.run(
                ["git", "rev-parse", "--git-dir"],
                cwd=self.repo_path,
                capture_output=True,
                text=True,
                check=True
            )
            return True
        except (subprocess.CalledProcessError, FileNotFoundError):
            return False
    
    def get_recent_commits(self, days: int = 7, author: Optional[str] = None) -> List[Dict]:
        """Get recent commits with details"""
        if not self.is_git_repo:
            return []
        
        # Build git log command
        cmd = [
            "git", "log",
            f"--since={days} days ago",
            "--pretty=format:%H|%s|%an|%ae|%ad|%B",
            "--date=iso"
        ]
        
        if author:
            cmd.extend(["--author", author])
        
        try:
            result = subprocess.run(
                cmd,
                cwd=self.repo_path,
                capture_output=True,
                text=True,
                check=True
            )
            
            commits = []
            for line in result.stdout.strip().split('\n'):
                if not line:
                    continue
                    
                parts = line.split('|', 5)
                if len(parts) >= 6:
                    commit = {
                        "hash": parts[0],
                        "subject": parts[1],
                        "author_name": parts[2],
                        "author_email": parts[3],
                        "date": parts[4],
                        "message": parts[5].strip()
                    }
                    commits.append(commit)
            
            return commits
            
        except subprocess.CalledProcessError:
            return []
    
    def get_commit_diff(self, commit_hash: str, context_lines: int = 3) -> str:
        """Get diff for a specific commit"""
        if not self.is_git_repo:
            return ""
        
        try:
            result = subprocess.run(
                ["git", "show", f"--unified={context_lines}", commit_hash],
                cwd=self.repo_path,
                capture_output=True,
                text=True,
                check=True
            )
            return result.stdout
        except subprocess.CalledProcessError:
            return ""
    
    def get_changed_files(self, commit_hash: str) -> List[str]:
        """Get list of files changed in a commit"""
        if not self.is_git_repo:
            return []
        
        try:
            result = subprocess.run(
                ["git", "diff-tree", "--no-commit-id", "--name-only", "-r", commit_hash],
                cwd=self.repo_path,
                capture_output=True,
                text=True,
                check=True
            )
            return [f.strip() for f in result.stdout.strip().split('\n') if f.strip()]
        except subprocess.CalledProcessError:
            return []
    
    def analyze_commit_patterns(self, commits: List[Dict]) -> Dict:
        """Analyze patterns in commit messages and changes"""
        if not commits:
            return {}
        
        # Analyze commit message patterns
        message_keywords = {}
        file_changes = {}
        commit_frequency = {}
        
        for commit in commits:
            # Extract keywords from commit messages
            message = commit["message"].lower()
            words = message.split()
            
            for word in words:
                if len(word) > 3 and word.isalpha():  # Filter meaningful words
                    message_keywords[word] = message_keywords.get(word, 0) + 1
            
            # Track file changes
            changed_files = self.get_changed_files(commit["hash"])
            for file_path in changed_files:
                file_ext = Path(file_path).suffix
                if file_ext:
                    file_changes[file_ext] = file_changes.get(file_ext, 0) + 1
            
            # Track commit frequency by day
            date = commit["date"][:10]  # YYYY-MM-DD
            commit_frequency[date] = commit_frequency.get(date, 0) + 1
        
        return {
            "total_commits": len(commits),
            "message_keywords": dict(sorted(message_keywords.items(), key=lambda x: x[1], reverse=True)[:20]),
            "file_changes": dict(sorted(file_changes.items(), key=lambda x: x[1], reverse=True)),
            "commit_frequency": commit_frequency,
            "analysis_date": datetime.now().isoformat()
        }
    
    def auto_capture_commits(self, days: int = 1, author: Optional[str] = None) -> List[str]:
        """Automatically capture recent commits as uroboro insights"""
        commits = self.get_recent_commits(days, author)
        captured_files = []
        
        from .aggregator import ContentAggregator
        aggregator = ContentAggregator()
        
        for commit in commits:
            # Create capture content from commit
            content = f"Git commit: {commit['subject']}\n\n{commit['message']}"
            
            # Add file change context
            changed_files = self.get_changed_files(commit["hash"])
            if changed_files:
                content += f"\n\nChanged files: {', '.join(changed_files[:10])}"
                if len(changed_files) > 10:
                    content += f" (and {len(changed_files) - 10} more)"
            
            # Capture with git-specific tags
            tags = ["git-commit", "auto-captured"]
            if "fix" in commit["subject"].lower() or "bug" in commit["subject"].lower():
                tags.append("bugfix")
            if "feat" in commit["subject"].lower() or "add" in commit["subject"].lower():
                tags.append("feature")
            
            try:
                captured_file = aggregator.quick_capture(
                    content, 
                    project=None,  # Will determine from current directory
                    tags=tags
                )
                captured_files.append(captured_file)
            except Exception as e:
                print(f"Warning: Could not capture commit {commit['hash'][:8]}: {e}")
        
        return captured_files
    
    def setup_git_hooks(self, hook_type: str = "post-commit") -> bool:
        """Setup git hooks to automatically capture commits"""
        if not self.is_git_repo:
            return False
        
        hooks_dir = self.repo_path / ".git" / "hooks"
        hook_file = hooks_dir / hook_type
        
        # Create hook script content
        hook_content = f"""#!/bin/bash
# uroboro git integration - auto-capture commits
# Generated by uroboro on {datetime.now().isoformat()}

# Check if uroboro is available
if command -v uro &> /dev/null; then
    # Get the latest commit message
    commit_msg=$(git log -1 --pretty=%B)
    
    # Capture to uroboro with git context
    uro capture "Git commit: $commit_msg" --tags git-commit auto-captured
else
    echo "uroboro not found in PATH - skipping auto-capture"
fi
"""
        
        try:
            # Write hook file
            with open(hook_file, 'w') as f:
                f.write(hook_content)
            
            # Make it executable
            os.chmod(hook_file, 0o755)
            
            print(f"✅ Git hook installed: {hook_file}")
            print("Commits will now be automatically captured to uroboro!")
            return True
            
        except Exception as e:
            print(f"❌ Failed to install git hook: {e}")
            return False
    
    def remove_git_hooks(self, hook_type: str = "post-commit") -> bool:
        """Remove uroboro git hooks"""
        if not self.is_git_repo:
            return False
        
        hooks_dir = self.repo_path / ".git" / "hooks"
        hook_file = hooks_dir / hook_type
        
        try:
            if hook_file.exists():
                # Check if it's our hook
                with open(hook_file, 'r') as f:
                    content = f.read()
                
                if "uroboro git integration" in content:
                    hook_file.unlink()
                    print(f"✅ Removed uroboro git hook: {hook_file}")
                    return True
                else:
                    print(f"❌ Hook exists but not created by uroboro: {hook_file}")
                    return False
            else:
                print(f"No hook found: {hook_file}")
                return False
                
        except Exception as e:
            print(f"❌ Failed to remove git hook: {e}")
            return False 