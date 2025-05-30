import json
import os
from datetime import datetime, timedelta
from pathlib import Path
from typing import Dict, List, Any
import argparse

class ContentAggregator:
    def __init__(self, config_path="config/settings.json"):
        self.config = self._load_config(config_path)
        self.notes_root = Path(self.config.get("notes_root", "~/notes")).expanduser()
        self.projects = self._load_projects()
    
    def _load_config(self, path: str) -> Dict:
        try:
            with open(path, 'r') as f:
                return json.load(f)
        except FileNotFoundError:
            return self._create_default_config(path)
    
    def _create_default_config(self, path: str) -> Dict:
        default = {
            "notes_root": "~/notes",
            "output_dir": "output",
            "projects": {
                "quantum-dice": {
                    "path": "~/stuff/projects/quantum_dice",
                    "type": "game",
                    "active": True
                }
            },
            "content_types": ["devlog", "blog", "social"]
        }
        os.makedirs(os.path.dirname(path), exist_ok=True)
        with open(path, 'w') as f:
            json.dump(default, f, indent=2)
        return default
    
    def _load_projects(self) -> Dict:
        return self.config.get("projects", {})
    
    def quick_capture(self, content: str, project: str = None, tags: List[str] = None):
        """Quick capture from command line or Cursor terminal"""
        timestamp = datetime.now().isoformat()
        
        # Determine capture location
        if project and project in self.projects:
            project_path = Path(self.projects[project]["path"]).expanduser()
            capture_file = project_path / ".devlog" / f"{datetime.now().date()}-capture.md"
        else:
            capture_file = self.notes_root / "daily" / f"{datetime.now().date()}.md"
        
        # Ensure directory exists
        capture_file.parent.mkdir(parents=True, exist_ok=True)
        
        # Append content
        with open(capture_file, 'a', encoding='utf-8') as f:
            f.write(f"\n## {timestamp}\n")
            if tags:
                f.write(f"Tags: {', '.join(tags)}\n")
            f.write(f"{content}\n")
        
        print(f"✅ Captured to {capture_file}")
        return str(capture_file)

    def collect_recent_activity(self, days: int = 1) -> Dict[str, Any]:
        """Collect recent activity across all monitored locations"""
        cutoff = datetime.now() - timedelta(days=days)
        activity = {
            "timestamp": datetime.now().isoformat(),
            "projects": {},
            "daily_notes": [],
            "raw_captures": []
        }
        
        # Collect from active projects
        for project_name, project_config in self.projects.items():
            if not project_config.get("active", False):
                continue
                
            project_path = Path(project_config["path"]).expanduser()
            project_activity = self._collect_project_activity(project_path, cutoff)
            if project_activity:
                activity["projects"][project_name] = project_activity
        
        # Collect daily notes
        daily_dir = self.notes_root / "daily"
        if daily_dir.exists():
            for note_file in daily_dir.glob("*.md"):
                if datetime.fromtimestamp(note_file.stat().st_mtime) > cutoff:
                    activity["daily_notes"].append({
                        "file": str(note_file),
                        "content": note_file.read_text(encoding='utf-8')
                    })
        
        return activity
    
    def _collect_project_activity(self, project_path: Path, cutoff: datetime) -> Dict:
        """Collect activity from a specific project"""
        activity = {}
        
        # Check for devlog entries
        devlog_dir = project_path / ".devlog"
        if devlog_dir.exists():
            recent_logs = []
            for log_file in devlog_dir.glob("*.md"):
                if log_file.name == "README.md":
                    # Skip README, handle separately
                    continue
                if datetime.fromtimestamp(log_file.stat().st_mtime) > cutoff:
                    recent_logs.append({
                        "file": str(log_file),
                        "content": log_file.read_text(encoding='utf-8')
                    })
            if recent_logs:
                activity["devlog"] = recent_logs
            
            # Load project context from README if available
            readme_file = devlog_dir / "README.md"
            if readme_file.exists():
                activity["context"] = {
                    "file": str(readme_file),
                    "content": readme_file.read_text(encoding='utf-8')
                }
        
        # TODO: Add git activity collection
        # TODO: Add file change detection
        
        return activity

# CLI interface for quick testing
if __name__ == "__main__":
    import sys
    
    parser = argparse.ArgumentParser(description="Content Pipeline Aggregator")
    subparsers = parser.add_subparsers(dest='command', help='Available commands')
    
    # Capture command
    capture_parser = subparsers.add_parser('capture', help='Capture content')
    capture_parser.add_argument('content', nargs='+', help='Content to capture')
    capture_parser.add_argument('--project', '-p', help='Project name')
    capture_parser.add_argument('--tags', '-t', nargs='+', help='Tags')
    
    # Collect command
    collect_parser = subparsers.add_parser('collect', help='Collect recent activity')
    collect_parser.add_argument('--days', '-d', type=int, default=1, help='Days to look back')
    
    if len(sys.argv) == 1:
        print("Content Pipeline Aggregator")
        print("Usage: python aggregator.py [capture|collect] ...")
        parser.print_help()
        sys.exit(0)
    
    args = parser.parse_args()
    aggregator = ContentAggregator()
    
    if args.command == "capture":
        content = " ".join(args.content)
        aggregator.quick_capture(content, project=args.project, tags=args.tags)
    
    elif args.command == "collect":
        activity = aggregator.collect_recent_activity(days=args.days)
        print(json.dumps(activity, indent=2)) 