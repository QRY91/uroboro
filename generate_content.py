#!/usr/bin/env python3
"""
Content Generation Script

Aggregates recent development activity and generates blog posts, social content, etc.
"""

import json
import argparse
from pathlib import Path
from src.aggregator import ContentAggregator
from src.processors.content_generator import ContentGenerator

def main():
    parser = argparse.ArgumentParser(description="Generate content from development activity")
    parser.add_argument("--days", "-d", type=int, default=1, help="Days of activity to collect")
    parser.add_argument("--output", "-o", choices=["blog", "devlog", "social", "all", "knowledge"], 
                       default="all", help="Type of content to generate")
    parser.add_argument("--title", "-t", help="Custom title for blog post")
    parser.add_argument("--tags", nargs="+", help="Tags for the content")
    parser.add_argument("--dry-run", action="store_true", help="Show output without saving")
    parser.add_argument("--format", "-f", choices=["mdx", "markdown", "text"], 
                       default="mdx", help="Output format for blog posts")
    parser.add_argument("--voice", "-v", help="Writing voice/style (e.g., professional_conversational, technical_deep, storytelling, minimalist, thought_leadership)")
    parser.add_argument("--preview", action="store_true", help="Show content preview instead of saving")
    parser.add_argument("--notes-path", help="Path to notes directory (for knowledge mining)")
    parser.add_argument("--mega-mining", action="store_true", help="Deep archaeological analysis of knowledge base")
    parser.add_argument("--no-privacy-filter", action="store_true", help="Disable privacy filter (include personal notes)")
    
    args = parser.parse_args()
    
    # Initialize content generator
    generator = ContentGenerator()
    
    # Handle knowledge mining separately
    if args.output == "knowledge":
        if args.mega_mining:
            print("🏺 MEGA MINING MODE: Archaeological expedition through your knowledge base...")
        else:
            print("🧠 Mining knowledge base for themes and insights...")
            
        knowledge_analysis = generator.mine_knowledge_base(
            args.notes_path, 
            deep_analysis=args.mega_mining,
            privacy_filter=not args.no_privacy_filter
        )
        
        if args.preview or args.dry_run:
            generator.preview_content(knowledge_analysis, "knowledge analysis")
        else:
            # Save to output directory with unique timestamp
            output_dir = Path("output") / "knowledge-mining"
            output_dir.mkdir(parents=True, exist_ok=True)
            
            from datetime import datetime
            timestamp = datetime.now().strftime("%Y-%m-%d_%H-%M-%S")
            
            # Include scope in filename if analyzing specific path
            if args.notes_path:
                scope = Path(args.notes_path).name
                # Sanitize scope for filename (replace spaces, special chars)
                scope_clean = scope.replace(" ", "-").replace("/", "-").replace("\\", "-")
                scope_clean = "".join(c for c in scope_clean if c.isalnum() or c in "-_")
                filename = f"archaeology-{scope_clean}-{timestamp}.md"
            else:
                filename = f"knowledge-analysis-{timestamp}.md"
            
            output_path = output_dir / filename
            
            with open(output_path, 'w', encoding='utf-8') as f:
                f.write(knowledge_analysis)
            
            print(f"✅ Knowledge archaeology saved to: {output_path}")
        return
    
    # Collect recent activity
    print(f"🔍 Collecting activity from last {args.days} day(s)...")
    aggregator = ContentAggregator()
    activity = aggregator.collect_recent_activity(days=args.days)
    
    # Check if we have any content
    total_items = len(activity.get("projects", {})) + len(activity.get("daily_notes", []))
    if total_items == 0:
        print("❌ No recent activity found to process")
        return
    
    print(f"✅ Found activity from {len(activity.get('projects', {}))} projects and {len(activity.get('daily_notes', []))} daily notes")
    
    # Generate content based on requested type
    if args.output in ["devlog", "all"]:
        print("\n📝 Generating devlog summary...")
        devlog = generator.generate_devlog_summary(activity)
        if args.preview or args.dry_run:
            generator.preview_content(devlog, "devlog")
        else:
            print("--- DEVLOG SUMMARY ---")
            print(devlog)
            print("--- END DEVLOG ---\n")
    
    if args.output in ["blog", "all"]:
        print("📝 Generating blog post...")
        blog_post = generator.generate_blog_post(activity, title=args.title, tags=args.tags, format=args.format, voice=args.voice)
        
        if args.preview:
            generator.preview_content(blog_post, "blog")
        elif args.dry_run:
            print("--- BLOG POST (DRY RUN) ---")
            print(blog_post)
            print("--- END BLOG POST ---\n")
        else:
            saved_path = generator.save_blog_post(blog_post, format=args.format)
            print(f"✅ Blog post saved to: {saved_path}")
    
    if args.output in ["social", "all"]:
        print("📱 Generating social media hooks...")
        social_hooks = generator.create_social_hooks(activity)
        print("--- SOCIAL HOOKS ---")
        for i, hook in enumerate(social_hooks, 1):
            print(f"{i}. {hook}")
        print("--- END SOCIAL HOOKS ---\n")
    
    # Save raw activity for reference
    if not args.dry_run and not args.preview:
        output_dir = Path("output") / "daily-runs"
        output_dir.mkdir(parents=True, exist_ok=True)
        
        from datetime import datetime
        timestamp = datetime.now().strftime("%Y-%m-%d_%H-%M-%S")
        activity_file = output_dir / f"activity_{timestamp}.json"
        
        with open(activity_file, 'w') as f:
            json.dump(activity, f, indent=2)
        
        print(f"📄 Raw activity saved to: {activity_file}")

if __name__ == "__main__":
    main() 