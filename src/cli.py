#!/usr/bin/env python3
"""
uroboro - The Self-Documenting Content Pipeline
Unified CLI interface for all uroboro functionality
"""

import argparse
import sys
from pathlib import Path

# Import existing modules
from .aggregator import ContentAggregator
from .processors.content_generator import ContentGenerator
from .usage_tracker import get_tracker
from .git_integration import GitIntegration
from .project_templates import ProjectTemplates


def track_command_usage(command: str, subcommand: str = None, success: bool = True):
    """Track command usage if enabled"""
    try:
        tracker = get_tracker()
        tracker.track_command(command, subcommand, success)
    except Exception:
        # Silently fail if tracking has issues - don't break user workflow
        pass


def cmd_capture(args):
    """Handle capture subcommand"""
    try:
        aggregator = ContentAggregator()
        content = " ".join(args.content)
        aggregator.quick_capture(content, project=args.project, tags=args.tags)
        track_command_usage("capture", success=True)
    except Exception as e:
        track_command_usage("capture", success=False)
        raise


def cmd_generate(args):
    """Handle generate subcommand - simplified from generate_content.py"""
    try:
        print(f"🔍 Collecting activity from last {args.days} day(s)...")
        aggregator = ContentAggregator()
        activity = aggregator.collect_recent_activity(days=args.days)
        
        # Check if we have any content
        total_items = len(activity.get("projects", {})) + len(activity.get("daily_notes", []))
        if total_items == 0:
            print("❌ No recent activity found to process")
            track_command_usage("generate", args.output, success=False)
            return
        
        print(f"✅ Found activity from {len(activity.get('projects', {}))} projects and {len(activity.get('daily_notes', []))} daily notes")
        
        generator = ContentGenerator()
        
        if args.output in ["devlog", "all"]:
            print("\n📝 Generating devlog summary...")
            devlog = generator.generate_devlog_summary(activity)
            if args.preview:
                generator.preview_content(devlog, "devlog")
            else:
                print("--- DEVLOG SUMMARY ---")
                print(devlog)
                print("--- END DEVLOG ---\n")
        
        if args.output in ["blog", "all"]:
            print("📝 Generating blog post...")
            blog_post = generator.generate_blog_post(
                activity, 
                title=args.title, 
                tags=args.tags, 
                format=args.format, 
                voice=args.voice
            )
            
            if args.preview:
                generator.preview_content(blog_post, "blog")
            else:
                saved_path = generator.save_blog_post(blog_post, format=args.format)
                print(f"✅ Blog post saved to: {saved_path}")
        
        if args.output in ["social", "all"]:
            print("📱 Generating social media hooks...")
            social_hooks = generator.create_social_hooks(activity, voice=args.voice)
            print("--- SOCIAL HOOKS ---")
            for i, hook in enumerate(social_hooks, 1):
                print(f"{i}. {hook}")
            print("--- END SOCIAL HOOKS ---\n")
        
        track_command_usage("generate", args.output, success=True)
        
    except Exception as e:
        track_command_usage("generate", args.output, success=False)
        raise


def cmd_voice(args):
    """Handle voice analysis subcommand"""
    try:
        from .voice_analyzer import main as voice_main
        # Call the existing voice analyzer
        voice_main()
        track_command_usage("voice", success=True)
    except ImportError:
        print("❌ Voice analyzer not available. Make sure voice_analyzer.py exists.")
        track_command_usage("voice", success=False)
    except Exception as e:
        track_command_usage("voice", success=False)
        raise


def cmd_mine(args):
    """Handle knowledge mining subcommand"""
    try:
        generator = ContentGenerator()
        
        if args.mega:
            print("🏺 MEGA MINING MODE: Archaeological expedition through your knowledge base...")
        else:
            print("🧠 Mining knowledge base for themes and insights...")
        
        knowledge_analysis = generator.mine_knowledge_base(
            args.path, 
            deep_analysis=args.mega,
            privacy_filter=not args.no_privacy_filter
        )
        
        if args.preview:
            generator.preview_content(knowledge_analysis, "knowledge analysis")
        else:
            # Save to output directory
            output_dir = Path("output") / "knowledge-mining"
            output_dir.mkdir(parents=True, exist_ok=True)
            
            from datetime import datetime
            timestamp = datetime.now().strftime("%Y-%m-%d_%H-%M-%S")
            
            if args.path:
                scope = Path(args.path).name
                scope_clean = "".join(c for c in scope if c.isalnum() or c in "-_")
                filename = f"archaeology-{scope_clean}-{timestamp}.md"
            else:
                filename = f"knowledge-analysis-{timestamp}.md"
            
            output_path = output_dir / filename
            
            with open(output_path, 'w', encoding='utf-8') as f:
                f.write(knowledge_analysis)
            
            print(f"✅ Knowledge archaeology saved to: {output_path}")
        
        track_command_usage("mine", "mega" if args.mega else "basic", success=True)
        
    except Exception as e:
        track_command_usage("mine", "mega" if args.mega else "basic", success=False)
        raise


def cmd_status(args):
    """Show uroboro status and recent activity"""
    try:
        aggregator = ContentAggregator()
        
        # Show config info
        print("🐍 uroboro status")
        print(f"Notes root: {aggregator.notes_root}")
        print(f"Active projects: {len([p for p in aggregator.projects.values() if p.get('active', False)])}")
        
        # Show recent captures
        activity = aggregator.collect_recent_activity(days=args.days)
        total_items = len(activity.get("projects", {})) + len(activity.get("daily_notes", []))
        print(f"Recent activity ({args.days} days): {total_items} items")
        
        if args.verbose:
            for project, data in activity.get("projects", {}).items():
                print(f"  📁 {project}: {len(data.get('devlog', []))} devlog entries")
        
        track_command_usage("status", success=True)
        
    except Exception as e:
        track_command_usage("status", success=False)
        raise


def cmd_git(args):
    """Handle git integration subcommand"""
    try:
        git = GitIntegration()
        
        if not git.is_git_repo:
            print("❌ Not in a git repository")
            track_command_usage("git", "not_repo", success=False)
            return
        
        if args.hook_install:
            success = git.setup_git_hooks()
            track_command_usage("git", "hook_install", success=success)
        
        elif args.hook_remove:
            success = git.remove_git_hooks()
            track_command_usage("git", "hook_remove", success=success)
        
        elif args.capture_commits:
            print(f"📦 Capturing git commits from last {args.days} days...")
            captured = git.auto_capture_commits(days=args.days, author=args.author)
            print(f"✅ Captured {len(captured)} commits")
            for file_path in captured:
                print(f"  📄 {file_path}")
            track_command_usage("git", "capture_commits", success=True)
        
        elif args.analyze:
            print(f"🔍 Analyzing git commits from last {args.days} days...")
            commits = git.get_recent_commits(days=args.days, author=args.author)
            analysis = git.analyze_commit_patterns(commits)
            
            print(f"\n📊 Git Analysis Results:")
            print(f"  Total commits: {analysis.get('total_commits', 0)}")
            
            if analysis.get('message_keywords'):
                print(f"\n  Top commit keywords:")
                for word, count in list(analysis['message_keywords'].items())[:10]:
                    print(f"    {word}: {count} times")
            
            if analysis.get('file_changes'):
                print(f"\n  File types changed:")
                for ext, count in analysis['file_changes'].items():
                    print(f"    {ext}: {count} files")
            
            track_command_usage("git", "analyze", success=True)
        
        else:
            # Show git status
            commits = git.get_recent_commits(days=7)
            print(f"🔗 Git Integration Status")
            print(f"Repository: {git.repo_path}")
            print(f"Recent commits (7 days): {len(commits)}")
            
            # Check for hooks
            hooks_dir = git.repo_path / ".git" / "hooks"
            hook_file = hooks_dir / "post-commit"
            if hook_file.exists():
                with open(hook_file, 'r') as f:
                    content = f.read()
                if "uroboro git integration" in content:
                    print("✅ Git hook installed (auto-capture enabled)")
                else:
                    print("⚠️ Git hook exists but not from uroboro")
            else:
                print("❌ No git hook installed")
            
            print("\nAvailable actions:")
            print("  uro git --hook-install        Install git hook for auto-capture")
            print("  uro git --capture-commits     Capture recent commits manually")
            print("  uro git --analyze            Analyze commit patterns")
            
            track_command_usage("git", "status", success=True)
    
    except Exception as e:
        track_command_usage("git", "error", success=False)
        raise


def cmd_project(args):
    """Handle project template subcommand"""
    try:
        templates = ProjectTemplates()
        
        if args.list:
            print("📋 Available project templates:")
            for template in templates.list_templates():
                print(f"  • {template}")
            print("\nUsage: uro project create <path> --template <type>")
            track_command_usage("project", "list", success=True)
        
        elif args.create:
            if not args.template:
                print("❌ Template type required. Use --template <type>")
                print("Available templates:", ", ".join(templates.list_templates()))
                track_command_usage("project", "create", success=False)
                return
            
            success = templates.create_project(
                project_path=args.create,
                template=args.template,
                project_name=args.name,
                description=args.description,
                tech_stack=args.tech_stack,
                context=args.context
            )
            
            track_command_usage("project", "create", success=success)
        
        else:
            print("📁 Project Templates")
            print("Available commands:")
            print("  uro project --list                    List available templates")
            print("  uro project create <path> --template <type>  Create new project")
            print("\nExample:")
            print("  uro project create my-web-app --template web --name 'My Web App'")
            
            track_command_usage("project", "help", success=True)
    
    except Exception as e:
        track_command_usage("project", "error", success=False)
        raise


def cmd_tracking(args):
    """Handle usage tracking management"""
    tracker = get_tracker()
    
    if args.enable:
        # Show consent dialog
        print("📊 Enable Usage Tracking")
        print()
        print("uroboro can collect anonymous usage statistics to improve the tool:")
        print("  • Which commands you use most")
        print("  • How often commands succeed/fail")
        print("  • Daily usage patterns")
        print()
        print("🔒 PRIVACY GUARANTEE:")
        print("  • Data NEVER leaves your machine")
        print("  • No personal content is tracked")
        print("  • No network requests are made")
        print("  • You can disable/clear data anytime")
        print()
        
        response = input("Enable local usage tracking? [y/N]: ").lower().strip()
        if response in ['y', 'yes']:
            tracker.enable_tracking()
        else:
            print("Usage tracking remains disabled.")
    
    elif args.disable:
        tracker.disable_tracking()
        
        response = input("Also clear existing usage data? [y/N]: ").lower().strip()
        if response in ['y', 'yes']:
            tracker.clear_data()
    
    elif args.clear:
        tracker.clear_data()
    
    elif args.stats:
        tracker.show_stats()
    
    else:
        # Show current status
        stats = tracker.get_stats()
        print("📊 Usage Tracking Status")
        print(f"Enabled: {stats['enabled']}")
        
        if stats['enabled']:
            total_commands = sum(cmd_data['count'] for cmd_data in stats['commands'].values())
            print(f"Commands tracked: {total_commands}")
            print(f"Data location: {tracker.usage_file}")
        
        print()
        print("Available actions:")
        print("  uroboro tracking --enable    Enable tracking")
        print("  uroboro tracking --disable   Disable tracking")
        print("  uroboro tracking --stats     Show statistics")
        print("  uroboro tracking --clear     Clear all data")


def show_first_run_notice():
    """Show privacy notice on first run"""
    tracker = get_tracker()
    notice_file = Path.home() / ".uroboro" / "first_run_complete"
    
    if not notice_file.exists():
        print("🐍 Welcome to uroboro!")
        print()
        print("uroboro respects your privacy:")
        print("  • All AI processing is local-only")
        print("  • No data is sent to external servers")
        print("  • Usage tracking is OPT-IN and local-only")
        print()
        print("To enable anonymous usage tracking to help improve uroboro:")
        print("  uroboro tracking --enable")
        print()
        
        # Create first run marker
        notice_file.parent.mkdir(exist_ok=True)
        notice_file.touch()


def main():
    """Main CLI entry point"""
    parser = argparse.ArgumentParser(
        prog="uroboro",
        description="The Self-Documenting Content Pipeline",
        epilog="Use 'uroboro <command> --help' for command-specific help"
    )
    
    subparsers = parser.add_subparsers(dest='command', help='Available commands')
    
    # Capture command
    capture_parser = subparsers.add_parser('capture', aliases=['c'], help='Capture development insights')
    capture_parser.add_argument('content', nargs='+', help='Content to capture')
    capture_parser.add_argument('--project', '-p', help='Project name')
    capture_parser.add_argument('--tags', '-t', nargs='+', help='Tags for categorization')
    
    # Generate command
    gen_parser = subparsers.add_parser('generate', help='Generate content from recent activity')
    gen_parser.add_argument('--days', '-d', type=int, default=1, help='Days of activity to collect')
    gen_parser.add_argument('--output', '-o', choices=["blog", "devlog", "social", "all"], 
                           default="all", help='Type of content to generate')
    gen_parser.add_argument('--title', '-t', help='Custom title for blog post')
    gen_parser.add_argument('--tags', nargs='+', help='Tags for the content')
    gen_parser.add_argument('--format', '-f', choices=["mdx", "markdown", "text"], 
                           default="mdx", help='Output format')
    gen_parser.add_argument('--voice', '-v', help='Writing voice/style')
    gen_parser.add_argument('--preview', action='store_true', help='Preview without saving')
    
    # Voice analysis command
    voice_parser = subparsers.add_parser('voice', help='Analyze and train your writing voice')
    
    # Knowledge mining command
    mine_parser = subparsers.add_parser('mine', help='Mine knowledge base for insights')
    mine_parser.add_argument('--path', help='Path to analyze (default: all notes)')
    mine_parser.add_argument('--mega', action='store_true', help='Deep archaeological analysis')
    mine_parser.add_argument('--no-privacy-filter', action='store_true', help='Include personal notes')
    mine_parser.add_argument('--preview', action='store_true', help='Preview without saving')
    
    # Status command
    status_parser = subparsers.add_parser('status', help='Show uroboro status and recent activity')
    status_parser.add_argument('--days', '-d', type=int, default=7, help='Days to check for activity')
    status_parser.add_argument('--verbose', '-v', action='store_true', help='Show detailed info')
    
    # Git integration command
    git_parser = subparsers.add_parser('git', help='Git integration for automatic commit capture')
    git_parser.add_argument('--hook-install', action='store_true', help='Install git hook for auto-capture')
    git_parser.add_argument('--hook-remove', action='store_true', help='Remove git hook')
    git_parser.add_argument('--capture-commits', action='store_true', help='Capture recent commits')
    git_parser.add_argument('--analyze', action='store_true', help='Analyze commit patterns')
    git_parser.add_argument('--days', '-d', type=int, default=7, help='Days to look back')
    git_parser.add_argument('--author', help='Filter commits by author')
    
    # Project templates command
    project_parser = subparsers.add_parser('project', help='Project template management')
    project_parser.add_argument('--list', action='store_true', help='List available templates')
    project_parser.add_argument('create', nargs='?', help='Create project at path')
    project_parser.add_argument('--template', help='Template type to use')
    project_parser.add_argument('--name', help='Project name')
    project_parser.add_argument('--description', help='Project description')
    project_parser.add_argument('--tech-stack', nargs='+', help='Technology stack')
    project_parser.add_argument('--context', help='Additional context for AI')
    
    # Usage tracking command
    tracking_parser = subparsers.add_parser('tracking', help='Manage usage tracking (local-only)')
    tracking_parser.add_argument('--enable', action='store_true', help='Enable usage tracking')
    tracking_parser.add_argument('--disable', action='store_true', help='Disable usage tracking')
    tracking_parser.add_argument('--stats', action='store_true', help='Show usage statistics')
    tracking_parser.add_argument('--clear', action='store_true', help='Clear usage data')
    
    # Parse args and dispatch
    args = parser.parse_args()
    
    if not args.command:
        show_first_run_notice()
        print("🐍 uroboro - The Self-Documenting Content Pipeline")
        print("")
        print("Quick start:")
        print("  uroboro capture 'Fixed authentication bug in login flow'")
        print("  uroboro generate --blog --voice storytelling")
        print("  uroboro status")
        print("")
        parser.print_help()
        return
    
    # Dispatch to command handlers
    commands = {
        'capture': cmd_capture,
        'c': cmd_capture,  # Shorthand for capture
        'generate': cmd_generate,
        'voice': cmd_voice,
        'mine': cmd_mine,
        'status': cmd_status,
        'git': cmd_git,
        'project': cmd_project,
        'tracking': cmd_tracking,
    }
    
    if args.command in commands:
        try:
            commands[args.command](args)
        except KeyboardInterrupt:
            print("\n👋 Interrupted by user")
            sys.exit(1)
        except Exception as e:
            print(f"❌ Error: {e}")
            sys.exit(1)
    else:
        print(f"❌ Unknown command: {args.command}")
        parser.print_help()
        sys.exit(1)


if __name__ == "__main__":
    main() 