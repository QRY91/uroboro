#!/usr/bin/env python3
"""
Local Usage Tracker for uroboro
Collects usage patterns locally to improve UX - never leaves your machine
"""

import json
import os
from datetime import datetime
from pathlib import Path
from typing import Dict, List


class LocalUsageTracker:
    def __init__(self):
        self.config_dir = Path.home() / ".uroboro"
        self.usage_file = self.config_dir / "usage.json"
        self.config_file = self.config_dir / "tracker_config.json"
        self.enabled = self._check_enabled()
    
    def _check_enabled(self) -> bool:
        """Check if usage tracking is enabled (opt-in)"""
        if not self.config_file.exists():
            return False  # Disabled by default
        
        try:
            with open(self.config_file, 'r') as f:
                config = json.load(f)
                return config.get("enabled", False)
        except:
            return False
    
    def enable_tracking(self):
        """Enable usage tracking with user consent"""
        self.config_dir.mkdir(exist_ok=True)
        
        config = {
            "enabled": True,
            "consent_date": datetime.now().isoformat(),
            "version": "1.0",
            "privacy_policy": "Data never leaves your machine. Used only to improve uroboro UX."
        }
        
        with open(self.config_file, 'w') as f:
            json.dump(config, f, indent=2)
        
        self.enabled = True
        print("✅ Usage tracking enabled. Data stays local, improves your uroboro experience.")
    
    def disable_tracking(self):
        """Disable usage tracking and optionally clear data"""
        if self.config_file.exists():
            config = {"enabled": False, "disabled_date": datetime.now().isoformat()}
            with open(self.config_file, 'w') as f:
                json.dump(config, f, indent=2)
        
        self.enabled = False
        print("✅ Usage tracking disabled.")
    
    def clear_data(self):
        """Clear all usage data"""
        if self.usage_file.exists():
            os.remove(self.usage_file)
        print("✅ Usage data cleared.")
    
    def track_command(self, command: str, subcommand: str = None, success: bool = True):
        """Track command usage (if enabled)"""
        if not self.enabled:
            return
        
        self.config_dir.mkdir(exist_ok=True)
        
        # Load existing data
        if self.usage_file.exists():
            try:
                with open(self.usage_file, 'r') as f:
                    data = json.load(f)
            except:
                data = {"commands": {}, "daily_stats": {}}
        else:
            data = {"commands": {}, "daily_stats": {}}
        
        # Track command usage
        cmd_key = f"{command}:{subcommand}" if subcommand else command
        if cmd_key not in data["commands"]:
            data["commands"][cmd_key] = {"count": 0, "last_used": None, "success_rate": []}
        
        data["commands"][cmd_key]["count"] += 1
        data["commands"][cmd_key]["last_used"] = datetime.now().isoformat()
        data["commands"][cmd_key]["success_rate"].append(success)
        
        # Keep only last 100 success/failure records per command
        if len(data["commands"][cmd_key]["success_rate"]) > 100:
            data["commands"][cmd_key]["success_rate"] = data["commands"][cmd_key]["success_rate"][-100:]
        
        # Track daily stats
        today = datetime.now().date().isoformat()
        if today not in data["daily_stats"]:
            data["daily_stats"][today] = 0
        data["daily_stats"][today] += 1
        
        # Save data
        with open(self.usage_file, 'w') as f:
            json.dump(data, f, indent=2)
    
    def get_stats(self) -> Dict:
        """Get usage statistics for user review"""
        if not self.usage_file.exists():
            return {"enabled": self.enabled, "commands": {}, "daily_stats": {}}
        
        try:
            with open(self.usage_file, 'r') as f:
                data = json.load(f)
                data["enabled"] = self.enabled
                return data
        except:
            return {"enabled": self.enabled, "commands": {}, "daily_stats": {}}
    
    def show_stats(self):
        """Display usage statistics to user"""
        stats = self.get_stats()
        
        if not stats["enabled"]:
            print("📊 Usage tracking is disabled.")
            print("   Enable with: uroboro tracking --enable")
            return
        
        print("📊 Local Usage Statistics")
        print(f"   Tracking enabled: {stats['enabled']}")
        print(f"   Data location: {self.usage_file}")
        print()
        
        if not stats["commands"]:
            print("   No commands tracked yet.")
            return
        
        print("Most used commands:")
        sorted_commands = sorted(
            stats["commands"].items(), 
            key=lambda x: x[1]["count"], 
            reverse=True
        )
        
        for cmd, data in sorted_commands[:10]:
            success_rate = sum(data["success_rate"]) / len(data["success_rate"]) * 100 if data["success_rate"] else 100
            print(f"   {cmd}: {data['count']} times (Success: {success_rate:.1f}%)")
        
        # Daily usage trend
        recent_days = sorted(stats["daily_stats"].items())[-7:]  # Last 7 days
        if recent_days:
            print("\nDaily usage (last 7 days):")
            for date, count in recent_days:
                print(f"   {date}: {count} commands")


# Global tracker instance
_tracker = None

def get_tracker() -> LocalUsageTracker:
    """Get the global tracker instance"""
    global _tracker
    if _tracker is None:
        _tracker = LocalUsageTracker()
    return _tracker 