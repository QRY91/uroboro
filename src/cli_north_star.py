#!/usr/bin/env python3
"""
uroboro - The Self-Documenting Content Pipeline
North Star CLI - 3 sacred commands only
"""

import argparse
import sys
from pathlib import Path
import json
from datetime import datetime

# Import existing modules
from .aggregator import ContentAggregator
from .processors.content_generator import ContentGenerator
from .usage_tracker import get_tracker
from .git_integration import GitIntegration


def track_command_usage(command: str, subcommand: str = None, success: bool = True):
    """Track command usage if enabled"""
    try:
        tracker = get_tracker()
        tracker.track_command(command, subcommand, success)
    except Exception:
        # Silently fail if tracking has issues - don't break user workflow
        pass


def cmd_capture(args):
    """Handle capture - the 10-second insight capture"""
    try:
        aggregator = ContentAggregator()
        content = " ".join(args.content)
        
        # Auto git integration - capture recent commits if in git repo
        if args.auto_git:
            try:
                git = GitIntegration()
                recent_commits = git.get_recent_commits(days=1)
                if recent_commits:
                    print(f"🔗 Auto-captured {len(recent_commits)} recent commits")
                    for commit in recent_commits:
                        git_content = f"Commit: {commit['message']} (files: {', '.join(commit['files'][:3])})"
                        aggregator.quick_capture(git_content, project=args.project, tags=['git'])
            except Exception:
                pass  # Git integration is optional
        
        aggregator.quick_capture(content, project=args.project, tags=args.tags)
        print(f"✅ Captured: {content[:60]}{'...' if len(content) > 60 else ''}")
        track_command_usage("capture", success=True)
        
    except Exception as e:
        track_command_usage("capture", success=False)
        print(f"❌ Capture failed: {e}")


def cmd_publish(args):
    """Handle publish - generate and publish professional content"""
    try:
        print(f"🔍 Collecting activity from last {args.days} day(s)...")
        aggregator = ContentAggregator()
        activity = aggregator.collect_recent_activity(days=args.days)
        
        # Check if we have any content
        total_items = len(activity.get("projects", {})) + len(activity.get("daily_notes", []))
        if total_items == 0:
            print("❌ No recent activity found to process")
            print("💡 Try: uro capture 'your development insight' first")
            track_command_usage("publish", args.type, success=False)
            return
        
        print(f"✅ Found activity from {len(activity.get('projects', {}))} projects")
        
        generator = ContentGenerator()
        
        # Auto-mine knowledge if requested or if deep mode
        if args.deep or args.mine:
            print("🧠 Mining knowledge base for deeper insights...")
            try:
                knowledge_analysis = generator.mine_knowledge_base(
                    path=None,  # Use default path
                    deep_analysis=args.deep,
                    privacy_filter=True
                )
                # Integrate knowledge mining into activity
                activity['knowledge_insights'] = knowledge_analysis
            except Exception as e:
                print(f"⚠️ Knowledge mining failed: {e}")
        
        # Auto-detect voice if not specified
        if not args.voice:
            print("🎤 Auto-detecting your writing voice...")
            try:
                from .voice_analyzer import VoiceAnalyzer
                analyzer = VoiceAnalyzer(".")  # Analyze current directory
                voice_profile = analyzer.analyze_notes()
                args.voice = "personal_detected"
                print(f"✅ Detected voice: {voice_profile.get('style', 'professional')}")
            except Exception:
                args.voice = "professional"  # Fallback
        
        # Generate content based on type
        if args.type == "blog":
            print("📝 Generating blog post...")
            content = generator.generate_blog_post(
                activity, 
                title=args.title, 
                tags=args.tags, 
                format=args.format, 
                voice=args.voice
            )
            
            if args.preview:
                generator.enhanced_preview_content(content, "blog")
            else:
                saved_path = generator.save_blog_post(content, format=args.format)
                print(f"✅ Blog post saved to: {saved_path}")
        
        elif args.type == "social":
            print("📱 Generating social media content...")
            social_hooks = generator.create_social_hooks(activity, voice=args.voice)
            
            print("--- SOCIAL HOOKS ---")
            for i, hook in enumerate(social_hooks, 1):
                print(f"{i}. {hook}")
            print("--- END SOCIAL HOOKS ---")
        
        elif args.type == "devlog":
            print("📋 Generating development log...")
            devlog = generator.generate_devlog_summary(activity)
            
            if args.preview:
                generator.enhanced_preview_content(devlog, "devlog")
            else:
                print("--- DEVLOG SUMMARY ---")
                print(devlog)
                print("--- END DEVLOG ---")
        
        track_command_usage("publish", args.type, success=True)
        
    except Exception as e:
        track_command_usage("publish", args.type, success=False)
        print(f"❌ Publish failed: {e}")


def cmd_status(args):
    """Show uroboro status - everything you need to know"""
    try:
        aggregator = ContentAggregator()
        
        # Core status
        print("🐍 uroboro status")
        print(f"Notes root: {aggregator.notes_root}")
        print(f"Active projects: {len([p for p in aggregator.projects.values() if p.get('active', False)])}")
        
        # Recent activity
        activity = aggregator.collect_recent_activity(days=args.days)
        total_items = len(activity.get("projects", {})) + len(activity.get("daily_notes", []))
        print(f"Recent activity ({args.days} days): {total_items} items")
        
        if args.verbose:
            for project, data in activity.get("projects", {}).items():
                print(f"  📁 {project}: {len(data.get('devlog', []))} devlog entries")
        
        # Git status (if in git repo)
        try:
            git = GitIntegration()
            recent_commits = git.get_recent_commits(days=args.days)
            print(f"Git commits ({args.days} days): {len(recent_commits)}")
            if args.verbose and recent_commits:
                for commit in recent_commits[:3]:  # Show latest 3
                    print(f"  🔗 {commit['hash'][:8]}: {commit['message'][:50]}")
        except Exception:
            print("Git status: Not a git repository")
        
        # Voice status
        print("\n🎤 Voice Profile Status:")
        try:
            from .voice_analyzer import VoiceAnalyzer
            analyzer = VoiceAnalyzer(".")
            voice_profile = analyzer.analyze_notes()
            print(f"  Style: {voice_profile.get('style', 'Not analyzed')}")
            print(f"  Sentences analyzed: {voice_profile.get('sentence_count', 0)}")
            print(f"  Avg sentence length: {voice_profile.get('avg_sentence_length', 0):.1f} words")
        except Exception:
            print("  Status: Voice not analyzed yet (run with --analyze-voice)")
        
        # Usage tracking (if enabled)
        try:
            tracker = get_tracker()
            if tracker._check_enabled():
                stats = tracker.get_stats()
                print(f"\n📊 Usage Stats:")
                print(f"  Total commands: {stats.get('total_commands', 0)}")
                print(f"  Most used: {stats.get('most_used_command', 'None')}")
                if args.verbose:
                    print("  Command breakdown:")
                    for cmd, count in stats.get('command_breakdown', {}).items():
                        print(f"    {cmd}: {count}")
        except Exception:
            print("\n📊 Usage tracking: Disabled")
        
        # Analyze voice if requested
        if args.analyze_voice:
            print("\n🎤 Analyzing voice patterns...")
            try:
                from .voice_analyzer import VoiceAnalyzer
                analyzer = VoiceAnalyzer(".")
                voice_profile = analyzer.analyze_notes()
                analyzer.save_profile()
                print("✅ Voice analysis complete and saved")
            except Exception as e:
                print(f"❌ Voice analysis failed: {e}")
        
        track_command_usage("status", success=True)
        
    except Exception as e:
        track_command_usage("status", success=False)
        print(f"❌ Status check failed: {e}")


def main():
    """Main CLI entry point - North Star simplicity"""
    parser = argparse.ArgumentParser(
        prog="uroboro",
        description="The Self-Documenting Content Pipeline",
        epilog="Three commands. That's it. 🎯"
    )
    
    subparsers = parser.add_subparsers(dest='command', help='Sacred commands')
    
    # CAPTURE - 10-second insight capture
    capture_parser = subparsers.add_parser('capture', 
                                         help='Capture development insights (10 seconds)')
    capture_parser.add_argument('content', nargs='+', help='Your development insight')
    capture_parser.add_argument('--project', '-p', help='Project name')
    capture_parser.add_argument('--tags', '-t', nargs='+', help='Tags for categorization')
    capture_parser.add_argument('--auto-git', action='store_true', 
                               help='Auto-capture recent git commits')
    
    # PUBLISH - generate professional content 
    publish_parser = subparsers.add_parser('publish',
                                         help='Generate professional content (2 minutes)')
    publish_parser.add_argument('--type', choices=["blog", "social", "devlog"], 
                               default="blog", help='Content type to generate')
    publish_parser.add_argument('--days', '-d', type=int, default=7, 
                               help='Days of activity to include')
    publish_parser.add_argument('--title', help='Custom title for content')
    publish_parser.add_argument('--tags', nargs='+', help='Tags for the content')
    publish_parser.add_argument('--format', choices=["mdx", "markdown", "text"], 
                               default="markdown", help='Output format')
    publish_parser.add_argument('--voice', help='Writing voice (auto-detected if not specified)')
    publish_parser.add_argument('--preview', action='store_true', 
                               help='Preview without saving')
    publish_parser.add_argument('--deep', action='store_true', 
                               help='Deep analysis with knowledge mining')
    publish_parser.add_argument('--mine', action='store_true',
                               help='Include knowledge base insights')
    
    # STATUS - see everything important
    status_parser = subparsers.add_parser('status', 
                                        help='Show status and recent activity')
    status_parser.add_argument('--days', '-d', type=int, default=7, 
                              help='Days to check for activity')
    status_parser.add_argument('--verbose', '-v', action='store_true', 
                              help='Show detailed information')
    status_parser.add_argument('--analyze-voice', action='store_true',
                              help='Analyze and update voice profile')
    
    # Parse and dispatch
    args = parser.parse_args()
    
    if not args.command:
        print("🐍 uroboro - The Self-Documenting Content Pipeline")
        print("")
        print("🎯 North Star Workflow (3 commands, that's it):")
        print("  uro capture 'Fixed database timeout - cut query time from 3s to 200ms'")
        print("  uro publish --blog")
        print("  uro status")
        print("")
        print("Get acknowledged for your actual work. 🔥")
        parser.print_help()
        return
    
    # Sacred command dispatch
    commands = {
        'capture': cmd_capture,
        'publish': cmd_publish,
        'status': cmd_status,
    }
    
    if args.command in commands:
        try:
            commands[args.command](args)
        except KeyboardInterrupt:
            print("\n👋 Interrupted by user")
            sys.exit(1)
        except Exception as e:
            print(f"❌ Error: {e}")
            if "--verbose" in sys.argv or "-v" in sys.argv:
                import traceback
                traceback.print_exc()
            sys.exit(1)
    else:
        print(f"❌ Unknown command: {args.command}")
        print("🎯 Valid commands: capture, publish, status")
        sys.exit(1)


if __name__ == "__main__":
    main() 