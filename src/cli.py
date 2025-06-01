#!/usr/bin/env python3
"""
uroboro - The Self-Documenting Content Pipeline
Unified CLI interface for all uroboro functionality
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
from .project_templates import ProjectTemplates
from .research_organizer import ResearchOrganizer
from .academic_voice import AcademicVoiceGenerator
from .interview_system import InteractiveInterviewer
from .qa_system import InteractiveQASession, QASystemModes


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
        
        # Enable final file display if requested
        if hasattr(args, 'show_final') and args.show_final:
            generator.set_show_final_file(True)
        
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


def cmd_research(args):
    """Handle research organization subcommand"""
    try:
        organizer = ResearchOrganizer()
        
        if args.init:
            # Initialize academic research project structure
            project_path = organizer.initialize_academic_project(
                args.project or "academic-research-project",
                args.path
            )
            track_command_usage("research", "init", success=True)
            return
            
        if args.setup_staging:
            # Setup import staging areas
            staging_dir = organizer.setup_import_staging(args.project)
            print(f"\n🎯 Import staging ready!")
            print(f"📁 Place materials in:")
            print(f"  🎨 Figma designs: {staging_dir}/staging/figma-designs/")
            print(f"  📝 Obsidian notes: {staging_dir}/staging/obsidian-notes/")
            print(f"  🤖 AI conversations: {staging_dir}/staging/ai-conversations/")
            print(f"  📄 Reference docs: {staging_dir}/staging/reference-docs/")
            track_command_usage("research", "setup_staging", success=True)
            return
            
        if args.import_obsidian:
            # Import Obsidian vault with research filtering
            print(f"📝 Importing Obsidian vault: {args.import_obsidian}")
            imported_notes = organizer.import_obsidian_vault_research(
                args.import_obsidian, 
                args.project,
                args.filter_patterns
            )
            print(f"✅ Successfully imported {len(imported_notes)} research-relevant notes")
            track_command_usage("research", "import_obsidian", success=True)
            return
            
        if args.import_figma:
            # Import Figma designs with academic documentation
            print(f"🎨 Importing Figma designs: {args.import_figma}")
            imported_designs = organizer.import_figma_designs_research(
                args.import_figma,
                args.project
            )
            print(f"✅ Successfully imported {len(imported_designs)} design artifacts")
            track_command_usage("research", "import_figma", success=True)
            return
            
        if args.import_conversations:
            # Import AI conversation logs
            print(f"🤖 Importing AI conversations: {args.import_conversations}")
            imported_convs = organizer.import_ai_conversations_research(
                args.import_conversations,
                args.project
            )
            print(f"✅ Successfully imported {len(imported_convs)} conversation logs")
            track_command_usage("research", "import_conversations", success=True)
            return
            
        if args.create_index:
            # Create comprehensive research materials index
            index_file = organizer.create_research_materials_index(args.project)
            print(f"✅ Research materials index created: {index_file}")
            track_command_usage("research", "create_index", success=True)
            return
            
        if args.analyze:
            # Analyze source project for development metrics
            print(f"🔍 Analyzing project: {args.source}")
            metrics = organizer.extract_development_metrics(args.source, args.days)
            
            if "error" in metrics:
                print(f"❌ Analysis failed: {metrics['error']}")
                track_command_usage("research", "analyze_error", success=False)
                return
                
            # Save metrics to research project
            if args.project:
                research_dir = Path(args.project)
            else:
                research_dir = Path.cwd()
                
            metrics_dir = research_dir / "research" / "performance-metrics"
            metrics_dir.mkdir(parents=True, exist_ok=True)
            
            timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
            metrics_file = metrics_dir / f"analysis_{timestamp}.json"
            
            with open(metrics_file, 'w', encoding='utf-8') as f:
                json.dump(metrics, f, indent=2)
                
            print(f"✅ Analysis saved to: {metrics_file}")
            print(f"📊 Found {metrics['git_activity'].get('total_commits', 0)} commits")
            print(f"🔧 Tech stack: {list(metrics['technical_stack']['languages'].keys())}")
            
            track_command_usage("research", "analyze", success=True)
            
        elif args.note:
            # Organize research notes
            if not args.category:
                print("❌ --category required when using --note")
                return
                
            research_path = args.project or Path.cwd()
            note_file = organizer.organize_research_notes(
                research_path, 
                args.category, 
                args.content or ""
            )
            
            print(f"✅ Research note saved to: {note_file}")
            track_command_usage("research", f"note_{args.category}", success=True)
            
        elif args.summary:
            # Generate development summary
            research_path = args.project or Path.cwd()
            summary = organizer.generate_development_summary(research_path)
            
            if args.save:
                output_dir = Path(research_path) / "output" / "documentation"
                output_dir.mkdir(parents=True, exist_ok=True)
                
                timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
                summary_file = output_dir / f"development_summary_{timestamp}.md"
                
                with open(summary_file, 'w', encoding='utf-8') as f:
                    f.write(summary)
                    
                print(f"✅ Development summary saved to: {summary_file}")
            else:
                print("--- DEVELOPMENT RESEARCH SUMMARY ---")
                print(summary)
                print("--- END SUMMARY ---")
                
            track_command_usage("research", "summary", success=True)
            
        else:
            # Show research command help
            print("🔬 Uroboro Research Organization System")
            print("=====================================")
            print()
            print("🎯 Academic Research Commands:")
            print("  uro research --init [name] [--path dir]    Initialize academic research project")
            print("  uro research --setup-staging               Setup import staging areas")
            print("  uro research --create-index                Create research materials index")
            print()
            print("📥 Material Import Commands:")
            print("  uro research --import-obsidian <path>      Import Obsidian vault with filtering")
            print("  uro research --import-figma <path>         Import Figma designs with documentation") 
            print("  uro research --import-conversations <path> Import AI conversation logs")
            print()
            print("📊 Analysis Commands:")
            print("  uro research --analyze --source <path>     Analyze project for metrics")
            print("  uro research --note --category <type>      Add research note")
            print("  uro research --summary [--save]            Generate development summary")
            print()
            print("🚀 Quick Start Workflow:")
            print("  1. uro research --init my-academic-project")
            print("  2. uro research --setup-staging --project my-academic-project")
            print("  3. uro research --import-obsidian ~/notes --project my-academic-project")
            print("  4. uro research --import-figma ~/figma-exports --project my-academic-project")
            print("  5. uro research --create-index --project my-academic-project")
            print()
            print("💡 Pro Tip: Combine with monkey dump for ultimate flexibility!")
            print("    uro dump ~/any-materials --output-dir my-project/imports/staging/")
            
            track_command_usage("research", "help", success=True)
            
    except Exception as e:
        track_command_usage("research", "error", success=False)
        raise


def cmd_academic(args):
    """Handle academic content generation subcommand"""
    try:
        generator = AcademicVoiceGenerator()
        
        if args.devlog:
            print("📚 Generating exhaustive academic development log...")
            
            # Collect recent activity for context
            aggregator = ContentAggregator()
            activity = aggregator.collect_recent_activity(days=args.days)
            
            # Extract development metrics if analyzing a project
            research_materials = {}
            if args.project_path:
                organizer = ResearchOrganizer()
                research_materials = organizer.extract_development_metrics(args.project_path, args.days)
            
            academic_devlog = generator.generate_academic_devlog(research_materials, args.focus)
            
            if args.save:
                output_dir = Path("output") / "academic"
                output_dir.mkdir(parents=True, exist_ok=True)
                
                timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
                filename = f"academic_devlog_{timestamp}.md"
                output_path = output_dir / filename
                
                with open(output_path, 'w', encoding='utf-8') as f:
                    f.write(academic_devlog)
                    
                print(f"✅ Academic devlog saved to: {output_path}")
            else:
                print("--- ACADEMIC DEVELOPMENT LOG ---")
                print(academic_devlog)
                print("--- END ACADEMIC LOG ---\n")
                
        elif args.bullets:
            print("📝 Generating exhaustive academic bullet points...")
            
            # Collect activity for bullet generation
            aggregator = ContentAggregator()
            activity = aggregator.collect_recent_activity(days=args.days)
            
            bullets = generator.generate_exhaustive_bullets(activity, "academic")
            
            print("--- EXHAUSTIVE ACADEMIC BULLETS ---")
            for i, bullet in enumerate(bullets, 1):
                print(f"{i:2d}. {bullet}")
            print("--- END BULLETS ---\n")
            
        elif args.synthesis:
            print("🔬 Generating research materials synthesis...")
            
            # Look for imported materials
            materials_path = Path(args.research_path or ".")
            imported_materials = list(materials_path.rglob("*.md"))
            
            synthesis = generator.generate_research_synthesis(imported_materials, args.focus)
            
            if args.save:
                output_dir = Path("output") / "academic"
                output_dir.mkdir(parents=True, exist_ok=True)
                
                timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
                filename = f"research_synthesis_{timestamp}.md"
                output_path = output_dir / filename
                
                with open(output_path, 'w', encoding='utf-8') as f:
                    f.write(synthesis)
                    
                print(f"✅ Research synthesis saved to: {output_path}")
            else:
                print("--- RESEARCH SYNTHESIS ---")
                print(synthesis)
                print("--- END SYNTHESIS ---\n")
                
        elif args.section:
            print(f"📖 Generating academic report section: {args.section}")
            
            # Collect source materials
            source_materials = {}
            if args.project_path:
                organizer = ResearchOrganizer()
                source_materials = organizer.extract_development_metrics(args.project_path, args.days)
                
            section_content = generator.generate_academic_report_section(args.section, source_materials)
            
            if args.save:
                output_dir = Path("output") / "academic" / "sections"
                output_dir.mkdir(parents=True, exist_ok=True)
                
                timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
                filename = f"{args.section}_section_{timestamp}.md"
                output_path = output_dir / filename
                
                with open(output_path, 'w', encoding='utf-8') as f:
                    f.write(section_content)
                    
                print(f"✅ {args.section.title()} section saved to: {output_path}")
            else:
                print(f"--- {args.section.upper()} SECTION ---")
                print(section_content)
                print("--- END SECTION ---\n")
                
        else:
            print("📚 Academic Voice Generator")
            print("Available commands:")
            print("  uro academic --devlog           Generate exhaustive academic development log")
            print("  uro academic --bullets          Generate exhaustive academic bullet points")  
            print("  uro academic --synthesis        Generate research materials synthesis")
            print("  uro academic --section <type>   Generate specific academic report section")
            print("\nExample:")
            print("  uro academic --devlog --project-path ../panopticron --save")
            print("  uro academic --section methodology --save")
            
        track_command_usage("academic", args.devlog or args.bullets or args.synthesis or args.section or "help", success=True)
        
    except Exception as e:
        track_command_usage("academic", "error", success=False)
        raise


def cmd_import(args):
    """Handle import subcommand for external data sources"""
    try:
        from .processors.content_generator import ContentGenerator
        
        generator = ContentGenerator()
        
        # Enable final file display if requested
        if hasattr(args, 'show_final') and args.show_final:
            generator.set_show_final_file(True)
        
        if args.obsidian:
            vault_path = args.obsidian
            print(f"🧠 Importing Obsidian vault from: {vault_path}")
            
            result = generator.import_obsidian_vault(
                vault_path=vault_path,
                output_dir=args.output_dir,
                include_private=args.include_private
            )
            
            if "error" in result:
                print(f"❌ Import failed: {result['error']}")
                track_command_usage("import", "obsidian", success=False)
                return
            
            summary = result["summary"]
            print(f"\n📊 Import Summary:")
            print(f"  • Processed: {summary['processed_files']} files")
            print(f"  • Skipped: {summary['skipped_files']} files")
            print(f"  • Links: {summary['total_links']}")
            print(f"  • Tags: {summary['total_tags']}")
            
            track_command_usage("import", "obsidian", success=True)
            
        else:
            print("📥 Import command")
            print("Available sources:")
            print("  --obsidian <path>    Import Obsidian vault")
            print("\nExample:")
            print("  uro import --obsidian ~/notes --show-final")
            
    except Exception as e:
        track_command_usage("import", "error", success=False)
        raise


def cmd_dump(args):
    """Handle dump subcommand - universal file ingestion"""
    try:
        from .processors.content_generator import ContentGenerator
        
        generator = ContentGenerator()
        
        # Enable final file display if requested
        if hasattr(args, 'show_final') and args.show_final:
            generator.set_show_final_file(True)
        
        source_path = args.path
        print(f"🐒 Starting universal ingestion from: {source_path}")
        
        result = generator.universal_ingest(
            source_path=source_path,
            output_dir=args.output_dir,
            project_name=args.project_name
        )
        
        if "error" in result:
            print(f"❌ Ingestion failed: {result['error']}")
            track_command_usage("dump", "error", success=False)
            return
        
        summary = result["summary"]
        print(f"\n🎉 MONKEY DUMP COMPLETE!")
        print(f"  📊 Processed: {summary['processed_files']} files")
        print(f"  ⏭️  Skipped: {summary['skipped_files']} files")
        print(f"  📁 Categories: {len(summary['categories'])} types")
        
        # Show breakdown by category
        for category, count in summary["categories"].items():
            print(f"    • {category}: {count} files")
        
        print(f"\n💾 Data: {result['data_file']}")
        print(f"📄 Report: {result['report_file']}")
        
        track_command_usage("dump", "success", success=True)
        
    except Exception as e:
        track_command_usage("dump", "error", success=False)
        raise


def cmd_interview(args):
    """Handle interactive interview subcommand"""
    try:
        interviewer = InteractiveInterviewer()
        
        # Determine materials path
        materials_path = args.path or "."
        
        print(f"🎤 Initializing {args.type} interview session...")
        print(f"📁 Analyzing materials at: {materials_path}")
        
        # Run the full interview session
        summary_file = interviewer.conduct_full_interview_session(
            materials_path=materials_path,
            interview_type=args.type,
            output_dir=args.output_dir
        )
        
        if summary_file:
            print(f"\n✅ Interview completed successfully!")
            print(f"📄 Results saved to: {summary_file}")
            print(f"\n💡 Use this extracted context to enhance your academic documentation!")
            
            # Offer to integrate with academic generation
            if args.integrate_academic:
                print("\n🔗 Integrating interview results with academic generation...")
                # Future: automatically incorporate interview insights into academic reports
                print("📝 Integration feature coming soon!")
        else:
            print("❌ Interview session failed or was cancelled.")
            track_command_usage("interview", args.type, success=False)
            return
        
        track_command_usage("interview", args.type, success=True)
        
    except KeyboardInterrupt:
        print("\n\n👋 Interview interrupted by user")
        track_command_usage("interview", f"{args.type}_interrupted", success=False)
    except Exception as e:
        print(f"❌ Interview error: {e}")
        track_command_usage("interview", args.type, success=False)
        raise


def cmd_qa(args):
    """Handle Q&A preparation subcommand for presentation practice"""
    try:
        qa_session = InteractiveQASession()
        
        # Determine materials path
        materials_path = args.path or "."
        
        print(f"🎭 Initializing Q&A session...")
        print(f"📁 Analyzing materials at: {materials_path}")
        print(f"👥 Audience types: {', '.join(args.audiences)}")
        print(f"🎯 Mode: {args.mode}")
        
        # Conduct Q&A session
        results = qa_session.prepare_for_presentation(
            materials_path=materials_path,
            audience_types=args.audiences,
            questions_per_audience=args.questions_per_audience,
            mode=args.mode
        )
        
        # Save results if not preview mode
        if args.mode != "preview" and "responses" in results:
            output_dir = args.output_dir or "./output/qa-sessions"
            report_file = qa_session.save_session_results(results, output_dir)
            print(f"\n📋 Q&A session completed!")
            print(f"💾 Report saved: {report_file}")
        elif args.mode == "preview":
            print(f"\n📋 Preview: Generated {len(results.get('questions', []))} questions")
            print("Use --mode practice or --mode quiz to conduct actual session")
        
        return True
        
    except Exception as e:
        print(f"❌ Q&A session failed: {e}")
        return False


def cmd_dojo(args):
    """Handle dojo subcommand - skill gap analysis and training preparation"""
    try:
        print("🥋 Entering the dojo - analyzing skill gaps...")
        
        if args.stats:
            from .sensei_system import DojoAnalyzer
            analyzer = DojoAnalyzer()
            
            print("📊 Generating uroboro skills assessment...")
            stats_report = analyzer.display_skill_stats()
            print(stats_report)
        
        elif args.analyze_gap:
            from .sensei_system import DojoAnalyzer
            analyzer = DojoAnalyzer()
            
            if args.skill:
                gap_analysis = analyzer.analyze_skill_gap(args.skill, materials_path=args.materials)
                print(f"📊 Skill gap analysis for '{args.skill}':")
                print(gap_analysis)
            else:
                print("🎯 Available skills to analyze:")
                skills = analyzer.list_available_skills()
                for skill in skills:
                    print(f"  • {skill}")
        
        elif args.prepare_training:
            from .sensei_system import DojoAnalyzer
            analyzer = DojoAnalyzer()
            training_plan = analyzer.prepare_training_materials(args.skill, args.materials)
            print(f"📚 Training materials prepared for '{args.skill}'")
            print(training_plan)
        
        else:
            # Show dojo help with stats option
            print("🥋 **Dojo - Skill Development Center**")
            print()
            print("Available commands:")
            print("  uro dojo --stats                     Show uroboro's skill assessment")
            print("  uro dojo --analyze-gap [--skill X]   Analyze skill gaps") 
            print("  uro dojo --prepare-training --skill X Prepare training materials")
            print()
            print("Example usage:")
            print("  uro dojo --stats                     # See all skills and proficiency")
            print("  uro dojo --analyze-gap --skill ieee_formatting")
            print("  uro dojo --prepare-training --skill academic_writing")
        
        track_command_usage("dojo", args.stats or args.analyze_gap or args.prepare_training or "help", success=True)
        
    except Exception as e:
        track_command_usage("dojo", success=False)
        print(f"❌ Dojo error: {e}")
        raise


def cmd_sensei(args):
    """Handle sensei subcommand - teaching and knowledge transfer"""
    try:
        print("🎓 Sensei mode activated - preparing to teach...")
        
        if args.teach:
            from .sensei_system import SenseiTeacher
            sensei = SenseiTeacher()
            
            teaching_session = sensei.teach_skill(
                skill=args.skill,
                student_type=args.student_type or "user",
                materials_path=args.materials,
                voice_profile=args.voice
            )
            
            print(f"🧠 Teaching session for '{args.skill}':")
            print(teaching_session)
            
            # Save teaching session for apprentice to learn from
            if args.save_session:
                session_path = sensei.save_teaching_session(teaching_session, args.skill)
                print(f"💾 Teaching session saved to: {session_path}")
        
        elif args.respond_to_query:
            from .sensei_system import SenseiTeacher
            sensei = SenseiTeacher()
            
            query_path = Path(args.query_file)
            if not query_path.exists():
                print(f"❌ Query file not found: {args.query_file}")
                return
            
            response = sensei.respond_to_query(query_path)
            print("📝 Sensei response:")
            print(response)
        
        track_command_usage("sensei", args.teach or args.respond_to_query, success=True)
        
    except Exception as e:
        track_command_usage("sensei", success=False)
        print(f"❌ Sensei error: {e}")
        raise


def cmd_apprentice(args):
    """Handle apprentice subcommand - learning through observation"""
    try:
        print("👁️ Apprentice mode - observing and learning...")
        
        if args.observe_session:
            from .sensei_system import ApprenticeObserver
            apprentice = ApprenticeObserver()
            
            session_path = Path(args.session_file)
            if not session_path.exists():
                print(f"❌ Session file not found: {args.session_file}")
                return
            
            learning_insights = apprentice.observe_teaching_session(session_path)
            print("🧠 Learning insights from session:")
            print(learning_insights)
            
            if args.integrate_voice:
                voice_updates = apprentice.integrate_into_voice_profile(learning_insights, args.voice_profile)
                print(f"🎯 Voice profile updated with new insights")
        
        elif args.practice_skill:
            from .sensei_system import ApprenticeObserver
            apprentice = ApprenticeObserver()
            
            practice_session = apprentice.practice_skill(
                skill=args.skill,
                practice_materials=args.materials,
                voice_profile=args.voice_profile
            )
            
            print(f"🎯 Practice session for '{args.skill}':")
            print(practice_session)
        
        track_command_usage("apprentice", args.observe_session or args.practice_skill, success=True)
        
    except Exception as e:
        track_command_usage("apprentice", success=False)
        print(f"❌ Apprentice error: {e}")
        raise


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
    gen_parser.add_argument('--show-final', action='store_true', help='Show final file')
    
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
    
    # Research organization command
    research_parser = subparsers.add_parser('research', help='Academic research organization and material import')
    research_parser.add_argument('--init', action='store_true', help='Initialize academic research project structure')
    research_parser.add_argument('--setup-staging', action='store_true', help='Setup import staging areas')
    research_parser.add_argument('--create-index', action='store_true', help='Create research materials index')
    
    # Material import arguments
    research_parser.add_argument('--import-obsidian', help='Import Obsidian vault with research filtering')
    research_parser.add_argument('--import-figma', help='Import Figma designs with academic documentation')
    research_parser.add_argument('--import-conversations', help='Import AI conversation logs for analysis')
    research_parser.add_argument('--filter-patterns', nargs='+', help='Custom filter patterns for imports')
    
    # Existing arguments
    research_parser.add_argument('--project', '-p', help='Research project path')
    research_parser.add_argument('--path', help='Custom path for project creation')
    research_parser.add_argument('--analyze', action='store_true', help='Analyze source project for metrics')
    research_parser.add_argument('--source', help='Source project to analyze')
    research_parser.add_argument('--days', '-d', type=int, default=90, help='Days of history to analyze')
    research_parser.add_argument('--note', action='store_true', help='Add research note')
    research_parser.add_argument('--category', choices=['technical-analysis', 'performance-metrics', 'implementation-notes', 'system-overview', 'deployment-process', 'user-feedback'], help='Research note category')
    research_parser.add_argument('--content', help='Note content')
    research_parser.add_argument('--summary', action='store_true', help='Generate development summary')
    research_parser.add_argument('--save', action='store_true', help='Save summary to file')
    
    # Academic content generation command
    academic_parser = subparsers.add_parser('academic', help='Generate exhaustive academic-style documentation')
    academic_parser.add_argument('--devlog', action='store_true', help='Generate academic development log')
    academic_parser.add_argument('--bullets', action='store_true', help='Generate exhaustive academic bullet points')
    academic_parser.add_argument('--synthesis', action='store_true', help='Generate research materials synthesis')
    academic_parser.add_argument('--section', choices=['methodology', 'implementation', 'results', 'analysis', 'conclusion'], help='Generate specific academic report section')
    academic_parser.add_argument('--days', '-d', type=int, default=30, help='Days of activity to analyze')
    academic_parser.add_argument('--focus', choices=['comprehensive', 'technical', 'user-experience'], default='comprehensive', help='Focus area for analysis')
    academic_parser.add_argument('--project-path', help='Path to project for technical analysis')
    academic_parser.add_argument('--research-path', help='Path to research materials')
    academic_parser.add_argument('--save', action='store_true', help='Save output to file')
    
    # Import command
    import_parser = subparsers.add_parser('import', help='Import external data sources')
    import_parser.add_argument('--obsidian', help='Path to Obsidian vault')
    import_parser.add_argument('--output-dir', help='Output directory for imported files')
    import_parser.add_argument('--include-private', action='store_true', help='Include private notes')
    import_parser.add_argument('--show-final', action='store_true', help='Show final file')
    
    # Dump command
    dump_parser = subparsers.add_parser('dump', help='Universal file ingestion - the monkey dump bucket')
    dump_parser.add_argument('path', help='Path to source data (required)')
    dump_parser.add_argument('--output-dir', help='Output directory for ingested files')
    dump_parser.add_argument('--project-name', help='Project name for ingestion')
    dump_parser.add_argument('--show-final', action='store_true', help='Show final file')
    
    # Interview command
    interview_parser = subparsers.add_parser('interview', help='Conduct an interactive interview to extract context and narrative')
    interview_parser.add_argument('--type', choices=['postmortem', 'deviation_analysis', 'gap_analysis'], 
                                 default='postmortem', help='Type of interview to conduct')
    interview_parser.add_argument('--path', help='Path to project materials to analyze (default: current directory)')
    interview_parser.add_argument('--output-dir', help='Output directory for interview results')
    interview_parser.add_argument('--integrate-academic', action='store_true', 
                                 help='Integrate interview results with academic generation (future feature)')
    
    # Q&A command
    qa_parser = subparsers.add_parser('qa', help='Interactive Q&A practice for presentation preparation')
    qa_parser.add_argument('--path', help='Path to project materials to analyze (default: current directory)')
    qa_parser.add_argument('--audiences', nargs='+', 
                          choices=['expert', 'layman', 'grandma', 'skeptical_academic', 
                                  'business_executive', 'technical_peer', 'concerned_citizen', 'investor'],
                          default=['expert', 'layman', 'grandma'],
                          help='Audience types for question generation')
    qa_parser.add_argument('--questions-per-audience', type=int, default=2,
                          help='Number of questions per audience type (default: 2)')
    qa_parser.add_argument('--mode', choices=['preview', 'practice', 'quiz'], default='practice',
                          help='Q&A session mode: preview (show questions), practice (full session), quiz (rapid fire)')
    qa_parser.add_argument('--output-dir', help='Output directory for Q&A session results')
    
    # Dojo command - skill gap analysis and training preparation
    dojo_parser = subparsers.add_parser('dojo', help='🥋 Skill gap analysis and training preparation')
    dojo_parser.add_argument('--stats', action='store_true', help='Show uroboro skills assessment')
    dojo_parser.add_argument('--analyze-gap', action='store_true', help='Analyze skill gaps')
    dojo_parser.add_argument('--prepare-training', action='store_true', help='Prepare training materials')
    dojo_parser.add_argument('--skill', help='Specific skill to analyze or train (e.g., ieee_formatting, academic_writing)')
    dojo_parser.add_argument('--materials', help='Path to materials for analysis (default: current directory)')
    
    # Sensei command - teaching and knowledge transfer
    sensei_parser = subparsers.add_parser('sensei', help='🎓 AI teaching and knowledge transfer')
    sensei_parser.add_argument('--teach', action='store_true', help='Teach a specific skill')
    sensei_parser.add_argument('--respond-to-query', action='store_true', help='Respond to a structured query')
    sensei_parser.add_argument('--skill', help='Skill to teach (e.g., ieee_formatting, academic_writing)')
    sensei_parser.add_argument('--student-type', choices=['user', 'ai', 'apprentice'], default='user', help='Type of student receiving teaching')
    sensei_parser.add_argument('--materials', help='Path to materials for teaching context')
    sensei_parser.add_argument('--voice', help='Voice profile to use for teaching')
    sensei_parser.add_argument('--save-session', action='store_true', help='Save teaching session for apprentice learning')
    sensei_parser.add_argument('--query-file', help='JSON file containing query for sensei response')
    
    # Apprentice command - learning through observation
    apprentice_parser = subparsers.add_parser('apprentice', help='👁️ Learning through observation and practice')
    apprentice_parser.add_argument('--observe-session', action='store_true', help='Observe and learn from a teaching session')
    apprentice_parser.add_argument('--practice-skill', action='store_true', help='Practice a specific skill')
    apprentice_parser.add_argument('--session-file', help='Teaching session file to observe')
    apprentice_parser.add_argument('--skill', help='Skill to practice')
    apprentice_parser.add_argument('--materials', help='Materials to practice with')
    apprentice_parser.add_argument('--voice-profile', help='Voice profile to update with learning')
    apprentice_parser.add_argument('--integrate-voice', action='store_true', help='Integrate insights into voice profile')
    
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
        'research': cmd_research,
        'academic': cmd_academic,
        'import': cmd_import,
        'dump': cmd_dump,
        'interview': cmd_interview,
        'qa': cmd_qa,
        'dojo': cmd_dojo,
        'sensei': cmd_sensei,
        'apprentice': cmd_apprentice,
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