#!/usr/bin/env python3
"""
uroboro - The Self-Documenting Content Pipeline
North Star CLI - 3 core commands only
"""

import argparse
import sys
from pathlib import Path
import json
from datetime import datetime

# Import existing modules
from .aggregator import ContentAggregator
from .processors.content_generator import ContentGenerator
from .git_integration import GitIntegration


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
        
        # EXPERIMENTAL: Interactive QA enhancement
        enhanced_content = content
        if hasattr(args, 'qa') and args.qa:
            enhanced_content = _enhance_capture_with_qa(content, args.qa)
        
        aggregator.quick_capture(enhanced_content, project=args.project, tags=args.tags)
        print(f"✅ Captured: {enhanced_content[:60]}{'...' if len(enhanced_content) > 60 else ''}")
        
    except Exception as e:
        print(f"❌ Capture failed: {e}")


def _enhance_capture_with_qa(content: str, num_questions: int) -> str:
    """EXPERIMENTAL: Enhance capture with interactive follow-up questions"""
    if num_questions < 1 or num_questions > 3:
        print("⚠️  QA questions limited to 1-3")
        return content
    
    print(f"\n🤔 {num_questions} quick question(s) to enhance your capture:")
    print(f"📝 Original: {content}")
    
    # Simple question templates based on content analysis
    questions = _get_smart_questions(content, num_questions)
    
    enhanced_parts = [content]
    
    for i, question in enumerate(questions, 1):
        try:
            answer = input(f"\n{i}. {question}\n> ").strip()
            if answer:
                enhanced_parts.append(answer)
            else:
                print("   (skipped)")
        except KeyboardInterrupt:
            print("\n⏭️  Skipping remaining questions")
            break
    
    if len(enhanced_parts) > 1:
        enhanced_content = enhanced_parts[0] + "\n\n" + "\n".join(f"• {part}" for part in enhanced_parts[1:])
        return enhanced_content
    
    return content


def _get_smart_questions(content: str, num_questions: int) -> list:
    """Generate smart follow-up questions based on capture content"""
    content_lower = content.lower()
    available_questions = []
    
    # Check for conventional commit patterns first
    commit_type = _detect_commit_type(content)
    if commit_type:
        available_questions.extend(_get_commit_type_questions(commit_type))
    
    # Context-aware question suggestions (fallback if no commit type)
    if not available_questions:
        if any(word in content_lower for word in ['fix', 'bug', 'error', 'issue']):
            available_questions.extend([
                "What was the root cause?",
                "How did you discover this?",
                "What's the impact/scope?"
            ])
        
        if any(word in content_lower for word in ['implement', 'add', 'create', 'build']):
            available_questions.extend([
                "What problem does this solve?",
                "What was the key insight?",
                "What's the next step?"
            ])
        
        if any(word in content_lower for word in ['optimize', 'improve', 'performance']):
            available_questions.extend([
                "What metrics improved?",
                "What was the bottleneck?", 
                "How much faster/better?"
            ])
    
    # Fallback general questions
    if not available_questions:
        available_questions = [
            "What was the main challenge?",
            "What did you learn?",
            "What's the impact?",
            "What would you do differently?",
            "What's next?"
        ]
    
    # Return up to num_questions, avoiding duplicates
    return list(dict.fromkeys(available_questions))[:num_questions]


def _detect_commit_type(content: str) -> str:
    """Detect conventional commit type from content"""
    content_lower = content.lower()
    
    # Direct conventional commit patterns
    if content_lower.startswith('feat:') or content_lower.startswith('feature:'):
        return 'feat'
    elif content_lower.startswith('fix:'):
        return 'fix'
    elif content_lower.startswith('refactor:'):
        return 'refactor'
    elif content_lower.startswith('docs:'):
        return 'docs'
    elif content_lower.startswith('test:'):
        return 'test'
    elif content_lower.startswith('style:'):
        return 'style'
    elif content_lower.startswith('perf:'):
        return 'perf'
    
    # Infer from git commit context
    if 'git commit:' in content_lower:
        # Extract the actual commit message
        if 'git commit:' in content_lower:
            commit_msg = content.split('git commit:', 1)[1].strip()
            return _detect_commit_type(commit_msg)
    
    return None


def _get_commit_type_questions(commit_type: str) -> list:
    """Get targeted questions based on conventional commit type"""
    questions_by_type = {
        'feat': [
            "What user problem does this solve?",
            "What's the key user benefit?",
            "What's the next enhancement?",
            "How will users discover this?"
        ],
        'fix': [
            "What was the root cause?",
            "How did you discover this bug?",
            "What's the user impact?",
            "How can this be prevented?"
        ],
        'refactor': [
            "What was the technical debt?",
            "What's cleaner/better now?",
            "What risk did this reduce?",
            "What does this enable next?"
        ],
        'docs': [
            "What was unclear before?",
            "Who will benefit from this?",
            "What examples help most?",
            "What questions does this answer?"
        ],
        'test': [
            "What edge case does this cover?",
            "What bug could this catch?",
            "How much confidence does this add?",
            "What scenario worried you?"
        ],
        'perf': [
            "What metrics improved?",
            "What was the bottleneck?",
            "How much faster/better?",
            "What's the user impact?"
        ],
        'style': [
            "What's more consistent now?",
            "What standard does this follow?",
            "What's easier to read?",
            "What confusion does this prevent?"
        ]
    }
    
    return questions_by_type.get(commit_type, [])


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
        
    except Exception as e:
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
        
        # Show recent capture content if requested
        if hasattr(args, 'recent') and args.recent:
            print(f"\n📝 Recent Captures (last {args.days} days):")
            captures_shown = 0
            
            # Parse daily notes for captures
            for daily_note in activity.get("daily_notes", []):
                content = daily_note.get("content", "")
                lines = content.split('\n')
                
                for i, line in enumerate(lines):
                    if line.startswith('## ') and 'T' in line:  # Timestamp marker
                        # Extract timestamp and capture content
                        timestamp = line.replace('## ', '').strip()
                        capture_content = []
                        j = i + 1
                        
                        # Collect content until next timestamp or end
                        while j < len(lines) and not lines[j].startswith('## '):
                            if lines[j].strip():  # Skip empty lines
                                capture_content.append(lines[j])
                            j += 1
                        
                        if capture_content:
                            # Format timestamp for display
                            try:
                                dt = datetime.fromisoformat(timestamp)
                                friendly_time = dt.strftime("%b %d, %H:%M")
                            except:
                                friendly_time = timestamp
                            
                            print(f"  🕐 {friendly_time}")
                            for content_line in capture_content[:3]:  # Show first 3 lines max
                                if content_line.startswith('Tags:'):
                                    print(f"    🏷️  {content_line}")
                                else:
                                    print(f"    📄 {content_line.strip()}")
                            if len(capture_content) > 3:
                                print(f"    ... ({len(capture_content) - 3} more lines)")
                            print()
                            
                            captures_shown += 1
                            if captures_shown >= 10:  # Limit to 10 most recent
                                break
                
                if captures_shown >= 10:
                    break
            
            if captures_shown == 0:
                print("  No recent captures found")
            elif captures_shown == 10:
                print(f"  (Showing 10 most recent captures)")
        
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
        
    except Exception as e:
        print(f"❌ Status check failed: {e}")


def main():
    """Main CLI entry point - North Star simplicity"""
    parser = argparse.ArgumentParser(
        prog="uroboro",
        description="The Self-Documenting Content Pipeline",
        epilog="Three commands. That's it. 🎯"
    )
    
    subparsers = parser.add_subparsers(dest='command', help='Core commands')
    
    # CAPTURE - 10-second insight capture
    capture_parser = subparsers.add_parser('capture', 
                                         help='Capture development insights (10 seconds)')
    capture_parser.add_argument('content', nargs='+', help='Your development insight')
    capture_parser.add_argument('--project', '-p', help='Project name')
    capture_parser.add_argument('--tags', '-t', nargs='+', help='Tags for categorization')
    capture_parser.add_argument('--auto-git', action='store_true', 
                               help='Auto-capture recent git commits')
    capture_parser.add_argument('--qa', type=int, metavar='N', 
                               help='[EXPERIMENTAL] Ask N follow-up questions (1-3) to enhance capture')
    
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
    status_parser.add_argument('--recent', action='store_true',
                              help='Show recent captures')
    
    # Parse and dispatch
    args = parser.parse_args()
    
    if not args.command:
        print("🐍 uroboro - The Self-Documenting Content Pipeline")
        print("🧪 EXPERIMENTAL BRANCH - Features being tested")
        print("")
        print("🎯 North Star Workflow (3 commands, that's it):")
        print("  uro capture 'Fixed database timeout - cut query time from 3s to 200ms'")
        print("  uro publish --blog")
        print("  uro status")
        print("")
        print("🧪 Experimental features:")
        print("  uro capture 'content' --qa 2  # Interactive follow-up questions")
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