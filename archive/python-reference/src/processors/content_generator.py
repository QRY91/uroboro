import json
import subprocess
import re
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Any, Optional

class ContentGenerator:
    def __init__(self, config: Dict = None):
        self.config = config or {}
        # Load from aggregator config if available
        if not config:
            try:
                with open("config/settings.json", 'r') as f:
                    self.config = json.load(f)
            except FileNotFoundError:
                pass
        
        self.llm_model = self.config.get("llm_model", "mistral:latest")
        self.templates_dir = Path("templates")
        self.show_final_file = False  # Control flag for final file display
        
    def set_show_final_file(self, enabled: bool = True):
        """Enable or disable final file display"""
        self.show_final_file = enabled
    
    def display_final_file(self, file_path: str, content_type: str = "file") -> None:
        """Display the final file content for following along"""
        if not self.show_final_file:
            return
            
        try:
            file_path = Path(file_path)
            if file_path.exists():
                print(f"\n🔍 FINAL {content_type.upper()}: {file_path}")
                print("=" * (len(str(file_path)) + len(content_type) + 15))
                with open(file_path, 'r', encoding='utf-8') as f:
                    content = f.read()
                print(content)
                print("=" * (len(str(file_path)) + len(content_type) + 15))
                print(f"📁 Saved to: {file_path.absolute()}\n")
            else:
                print(f"⚠️  Final file not found: {file_path}")
        except Exception as e:
            print(f"❌ Error displaying final file: {e}")
    
    def enhanced_preview_content(self, content: str, content_type: str = "content", file_path: str = None) -> None:
        """Enhanced preview with optional final file display"""
        print(f"\n--- {content_type.upper()} PREVIEW ---")
        if file_path:
            print(f"📄 Will be saved to: {file_path}")
        print("-" * 50)
        print(content)
        print("-" * 50)
        print(f"--- END {content_type.upper()} PREVIEW ---\n")
        
        # Show final file if enabled and path provided
        if file_path and self.show_final_file:
            # Wait a moment for user to read preview
            try:
                input("Press Enter to see final file content...")
                self.display_final_file(file_path, content_type)
            except KeyboardInterrupt:
                print("\n⏭️  Skipping final file display")
    
    def _call_ollama(self, prompt: str, model: str = None) -> str:
        """Call Ollama to generate content"""
        model = model or self.llm_model
        
        try:
            print(f"[DEBUG] Calling ollama with model: {model}, prompt length: {len(prompt)} chars")
            result = subprocess.run([
                "ollama", "run", model
            ], input=prompt, capture_output=True, text=True, timeout=15)
            
            if result.returncode != 0:
                print(f"[DEBUG] Ollama error: {result.stderr}")
                return f"Error calling LLM: {result.stderr}"
            
            print(f"[DEBUG] Ollama success, response length: {len(result.stdout)} chars")
            return result.stdout.strip()
            
        except subprocess.TimeoutExpired:
            print("[DEBUG] Ollama timeout after 15 seconds")
            return "Error: LLM request timed out"
        except Exception as e:
            print(f"[DEBUG] Ollama exception: {e}")
            return f"Error calling LLM: {str(e)}"
    
    def generate_devlog_summary(self, activity_data: Dict) -> str:
        """Generate a development log summary from aggregated activity"""
        
        # PRIORITY 1: Extract daily notes FIRST (most recent captures)
        daily_content = []
        for note in activity_data.get("daily_notes", []):
            if note.get("priority") == "high" or note.get("type") == "recent_capture":
                daily_content.append(f"**Recent Capture**: {note['content']}")
            else:
                daily_content.append(note['content'])
        
        # PRIORITY 2: Extract project content (older devlogs)
        project_content = []
        for project_name, project_data in activity_data.get("projects", {}).items():
            if "devlog" in project_data:
                for entry in project_data["devlog"]:
                    project_content.append(f"**{project_name}**: {entry['content']}")
        
        # Prioritize daily notes by putting them first
        all_content = "\n".join(daily_content + project_content)
        
        prompt = f"""
        Analyze this development activity and create a concise development log summary.
        
        IMPORTANT: The "Recent Capture" entries are today's most important work - prioritize these heavily.

        {all_content}

        Create a structured summary with:
        ## Technical Work
        - Brief bullet points of what was accomplished (prioritize recent captures)

        ## Key Insights  
        - Important discoveries or decisions made

        ## Next Steps
        - What should be tackled next

        Keep it professional but conversational. Focus on the most significant items from today's captures first.
        """
        
        return self._call_ollama(prompt)
    
    def generate_blog_post(self, activity_data: Dict, title: str = None, tags: List[str] = None, format: str = "mdx", voice: str = None) -> str:
        """Generate a blog post from aggregated activity data"""
        
        # Get style configuration
        style_config = self.config.get("style_config", {})
        voice_name = voice or style_config.get("default_voice", "professional_conversational")
        voice_config = style_config.get("voices", {}).get(voice_name, {})
        
        # Build style instructions
        style_instructions = voice_config.get("prompt_additions", "Write in engaging, professional tone.")
        
        # Add brand voice preferences
        brand_voice = style_config.get("brand_voice", {})
        if brand_voice.get("avoid_phrases"):
            style_instructions += f"\n\nAvoid these phrases: {', '.join(brand_voice['avoid_phrases'])}"
        if brand_voice.get("preferred_phrases"):
            style_instructions += f"\nPrefer phrases like: {', '.join(brand_voice['preferred_phrases'])}"
        
        # PRIORITY: Focus on daily captures (recent work)
        daily_captures = []
        for note in activity_data.get("daily_notes", []):
            if note.get("priority") == "high" or note.get("type") == "recent_capture":
                # Extract actual capture content after timestamps
                content_lines = note['content'].split('\n')
                for line in content_lines:
                    if line.startswith('## 2025-06-02T'):  # Today's captures
                        # Skip the timestamp line, get the actual content
                        continue
                    elif line.strip() and not line.startswith('##'):
                        # This is the actual capture content
                        daily_captures.append(line.strip())
                        
        # Build focused content summary  
        today_work = "\n".join(daily_captures[:5]) if daily_captures else "No recent captures found"
        
        if not title:
            title = f"Dev Update - {datetime.now().strftime('%B %d, %Y')}"
        
        # Create focused prompt with today's work only
        prompt = f"""Write an engaging blog post about today's development work:

TODAY'S CAPTURES:
{today_work}

Structure:
1. Brief intro (1-2 sentences)
2. Main highlights from today's work
3. Brief technical insight or lesson learned
4. Quick note on what's next

STYLE: {style_instructions}

Keep it 200-300 words, focused on today's specific achievements."""
        
        content = self._call_ollama(prompt)
        
        if format == "mdx":
            # Generate frontmatter for MDX
            frontmatter = self._generate_frontmatter(title, tags)
            return f"{frontmatter}\n\n{content}"
        elif format == "markdown":
            # Plain markdown with simple header
            return f"# {title}\n\n*{datetime.now().strftime('%B %d, %Y')}*\n\n{content}"
        else:
            # Plain text
            return f"{title}\n{'=' * len(title)}\n{datetime.now().strftime('%B %d, %Y')}\n\n{content}"
    
    def _generate_frontmatter(self, title: str, tags: List[str] = None) -> str:
        """Generate MDX frontmatter for qryzone blog"""
        
        if not tags:
            tags = ["development", "update"]
        
        date = datetime.now().strftime("%Y-%m-%d")
        
        return f"""---
title: "{title}"
date: "{date}"
author: "Q"
tags: {json.dumps(tags)}
excerpt: "Recent development progress and insights"
---"""
    
    def save_blog_post(self, content: str, filename: str = None, output_dir: str = None, format: str = "markdown") -> str:
        """Save blog post to appropriate directory based on format"""
        
        if not filename:
            timestamp = datetime.now().strftime("%Y-%m-%d")
            if format == "html":
                filename = f"dev-update-{timestamp}.html"
            elif format == "markdown":
                filename = f"dev-update-{timestamp}.md"
            else:
                filename = f"dev-update-{timestamp}.txt"
        
        if not output_dir:
            # All formats go to local output by default
            output_dir = Path("output/posts")
        else:
            output_dir = Path(output_dir)
        
        output_dir.mkdir(parents=True, exist_ok=True)
        output_path = output_dir / filename
        
        # Convert markdown to HTML if HTML format requested
        if format == "html" and not content.strip().startswith('<!DOCTYPE'):
            # Simple markdown to HTML conversion
            html_content = self._markdown_to_html(content)
            content = html_content
        
        with open(output_path, 'w', encoding='utf-8') as f:
            f.write(content)
        
        # Display final file if enabled
        if self.show_final_file:
            self.display_final_file(str(output_path), "blog post")
        
        return str(output_path)
    
    def save_content_with_preview(self, content: str, file_path: str, content_type: str = "content") -> str:
        """Save content with enhanced preview and final file display"""
        file_path = Path(file_path)
        file_path.parent.mkdir(parents=True, exist_ok=True)
        
        # Show enhanced preview first
        self.enhanced_preview_content(content, content_type, str(file_path))
        
        # Save the file
        with open(file_path, 'w', encoding='utf-8') as f:
            f.write(content)
        
        return str(file_path)
    
    def preview_content(self, content: str, content_type: str = "blog") -> None:
        """Display content in a readable format for preview"""
        print(f"\n--- {content_type.upper()} PREVIEW ---")
        print(content)
        print(f"--- END {content_type.upper()} PREVIEW ---\n")
    
    def create_social_hooks(self, activity_data: Dict, voice: str = None) -> List[str]:
        """Generate social media hooks from activity"""
        
        # Get style configuration
        style_config = self.config.get("style_config", {})
        voice_name = voice or style_config.get("default_voice", "professional_conversational")
        voice_config = style_config.get("voices", {}).get(voice_name, {})
        
        # Build style instructions for social content
        style_instructions = voice_config.get("prompt_additions", "Write in conversational tone.")
        
        # Add brand voice preferences
        brand_voice = style_config.get("brand_voice", {})
        if brand_voice.get("avoid_phrases"):
            style_instructions += f"\n\nAvoid these phrases: {', '.join(brand_voice['avoid_phrases'])}"
        if brand_voice.get("preferred_phrases"):
            style_instructions += f"\nPrefer phrases like: {', '.join(brand_voice['preferred_phrases'])}"
        
        # Extract key accomplishments
        accomplishments = []
        for project_name, project_data in activity_data.get("projects", {}).items():
            if "devlog" in project_data:
                for entry in project_data["devlog"]:
                    # Extract meaningful content (skip timestamps)
                    content_lines = [line for line in entry['content'].split('\n') 
                                   if line.strip() and not line.startswith('##')]
                    accomplishments.extend(content_lines)
        
        combined_work = "\n".join(accomplishments[:10])  # Limit context
        
        # Adjust social approach based on voice
        if voice_name in ["matter_of_fact", "personal_excavated"]:
            social_approach = """Create 3-5 concise social media posts from this development work:
            
            - Report what was built/learned without promotional language
            - Focus on technical substance
            - Avoid exclamation points and engagement tactics
            - Include relevant hashtags but keep them minimal
            - Write for people who already care about the technology"""
        else:
            social_approach = """Create 3-5 engaging social media hooks from this development work:
            
            - Make them engaging and accessible
            - Include relevant hashtags
            - Show progress/momentum
            - Mix technical and human interest angles"""
        
        prompt = f"""
        {social_approach}

        {combined_work}

        STYLE INSTRUCTIONS:
        {style_instructions}

        Format each as a tweet-style post (under 280 chars):

        Output format:
        1. [hook text]
        2. [hook text]
        etc.
        """
        
        response = self._call_ollama(prompt)
        
        # Parse numbered list
        hooks = []
        for line in response.split('\n'):
            line = line.strip()
            if re.match(r'^\d+\.\s+', line):
                hook = re.sub(r'^\d+\.\s+', '', line)
                hooks.append(hook)
        
        return hooks
    
    def mine_knowledge_base(self, notes_path: str = None, deep_analysis: bool = False, privacy_filter: bool = True) -> str:
        """Mine the notes knowledge base for themes, insights, and patterns"""
        
        if not notes_path:
            # Use notes path from config
            notes_path = Path(self.config.get("projects", {}).get("notes", {}).get("path", "~/notes")).expanduser()
        else:
            notes_path = Path(notes_path).expanduser()
        
        if not notes_path.exists():
            return "Error: Notes directory not found"
        
        # Collect markdown files
        md_files = list(notes_path.glob("*.md"))
        txt_files = list(notes_path.glob("*.txt"))
        
        # For deep analysis, also recursively search subdirectories
        if deep_analysis:
            md_files.extend(list(notes_path.rglob("*.md")))
            txt_files.extend(list(notes_path.rglob("*.txt")))
            # Also include other text-like files
            other_files = list(notes_path.rglob("*.html"))
            other_files.extend(list(notes_path.rglob("*.json")))
            other_files.extend(list(notes_path.rglob("*.log")))
        else:
            other_files = []
        
        all_files = md_files + txt_files + other_files
        
        if not all_files:
            return "Error: No markdown or text files found in notes directory"
        
        # Privacy filter keywords (things to skip or anonymize)
        privacy_keywords = [
            "password", "secret", "private", "personal", "embarrassing", 
            "diary", "journal", "confession", "vent", "rant", "therapy",
            "relationship", "dating", "crush", "anxiety", "depression"
        ] if privacy_filter else []
        
        # Read and aggregate content
        knowledge_content = []
        file_summaries = []
        source_mapping = {}
        
        for file_path in all_files:
            try:
                content = file_path.read_text(encoding='utf-8')
                if len(content.strip()) < 50:  # Skip very short files
                    continue
                
                # Privacy filter check
                if privacy_filter and any(keyword in content.lower() for keyword in privacy_keywords):
                    file_summaries.append(f"- {file_path.name} (PRIVATE - {len(content)} chars)")
                    continue
                
                # Store full content for deep analysis
                if deep_analysis:
                    content_chunk = content[:2000]  # More content for deep analysis
                else:
                    content_chunk = content[:1000]
                
                knowledge_content.append(f"**{file_path.name}**:\n{content_chunk}...")
                file_summaries.append(f"- {file_path.name} ({len(content)} chars)")
                source_mapping[file_path.name] = {
                    "path": str(file_path),
                    "size": len(content),
                    "modified": file_path.stat().st_mtime if file_path.exists() else 0
                }
                
            except Exception as e:
                print(f"Error reading {file_path}: {e}")
        
        combined_content = "\n\n".join(knowledge_content)
        files_summary = "\n".join(file_summaries)
        source_list = "\n".join([f"- **{name}**: {info['path']}" for name, info in source_mapping.items()])
        
        if deep_analysis:
            prompt = f"""
            Perform DEEP ARCHAEOLOGICAL ANALYSIS of this knowledge base. This is a personal reflection for the author to rediscover forgotten insights and patterns in their own thinking.

            FILES ANALYZED ({len(source_mapping)} files):
            {files_summary}

            CONTENT FOR ANALYSIS:
            {combined_content}

            SOURCE MAPPING:
            {source_list}

            Create a comprehensive self-reflection analysis in markdown format:

            # 🧭 Personal Knowledge Archaeology Report

            ## 📚 Knowledge Inventory
            - What domains of knowledge are represented?
            - What's the balance between technical and non-technical content?
            - What timeframes are covered?

            ## 🧠 Your Mental Models
            - What frameworks for thinking appear repeatedly?
            - How do you approach problem-solving across different domains?
            - What cognitive patterns emerge?

            ## 🎯 Core Values & Philosophy
            - What principles guide your decisions?
            - What matters most to you professionally and personally?
            - How has your thinking evolved?

            ## 🔧 Technical Evolution
            - What's your learning trajectory across technologies?
            - Where do you see knowledge gaps vs strengths?
            - What tools/languages/concepts keep appearing?

            ## 🌊 Recurring Themes & Obsessions
            - What topics do you return to repeatedly?
            - What problems fascinate you?
            - What questions keep coming up?

            ## 🔗 Hidden Connections
            - What unexpected links exist between seemingly unrelated notes?
            - How do your interests intersect and inform each other?
            - What patterns might you have missed?

            ## 💡 Forgotten Gold
            - What insights seem particularly valuable but might be forgotten?
            - What ideas deserve revisiting?
            - What knowledge could be "freshened up" for current projects?

            ## 🚀 Future Content Seeds
            - What stories are hiding in this knowledge?
            - What would make compelling blog posts or projects?
            - What expertise could you share with others?

            ## 🎭 Your Intellectual Personality
            - How would you describe your thinking style?
            - What makes your approach unique?
            - What's your "intellectual signature"?

            ## 📍 Knowledge Map
            - Which files contain your most valuable insights?
            - What should you revisit first?
            - Where are the treasure troves?

            Be specific, quote examples, cite sources by filename. Make this a mirror for self-discovery.
            """
        else:
            prompt = f"""
            Analyze this knowledge base from markdown files and extract key insights:

            FILES ANALYZED:
            {files_summary}

            CONTENT SAMPLE:
            {combined_content}

            Please provide a comprehensive analysis in markdown format covering:

            # Knowledge Base Analysis

            ## Main Themes
            - What are the 3-5 central themes across all content?

            ## Technical Interests
            - What technologies, tools, and languages appear most frequently?
            - What patterns in technical learning/exploration exist?

            ## Professional Philosophy 
            - What approaches to software development emerge?
            - What values or principles are emphasized?

            ## Current Challenges
            - What problems or obstacles are being worked through?
            - What areas need more development?

            ## Learning Patterns
            - How does the author approach learning new topics?
            - What knowledge gaps are identified?

            ## Project Insights
            - What connections exist between different projects?
            - What development patterns emerge?

            ## Key Takeaways
            - What are the most valuable insights for future development?
            - What themes could inform content creation?

            Be specific and quote relevant examples where possible. Aim for actionable insights.
            """
        
        return self._call_ollama(prompt)
    
    def import_obsidian_vault(self, vault_path: str, output_dir: str = None, include_private: bool = False) -> Dict[str, Any]:
        """Import Obsidian vault structure and content for context building"""
        
        vault_path = Path(vault_path).expanduser()
        if not vault_path.exists():
            return {"error": "Vault path does not exist"}
        
        if not output_dir:
            output_dir = Path("output") / "obsidian-import"
        else:
            output_dir = Path(output_dir)
        
        output_dir.mkdir(parents=True, exist_ok=True)
        
        # Check for Obsidian config
        obsidian_config = vault_path / ".obsidian"
        has_obsidian_config = obsidian_config.exists()
        
        print(f"🧠 Importing Obsidian vault: {vault_path}")
        print(f"📁 Output directory: {output_dir}")
        print(f"⚙️  Obsidian config detected: {has_obsidian_config}")
        
        # Collect all markdown files
        md_files = list(vault_path.rglob("*.md"))
        
        # Privacy filtering
        privacy_keywords = [
            "private", "personal", "diary", "journal", "secret",
            "password", "sensitive", "confidential", "vent", "rant"
        ] if not include_private else []
        
        # Parse vault structure
        vault_data = {
            "metadata": {
                "vault_path": str(vault_path),
                "import_timestamp": datetime.now().isoformat(),
                "total_files": len(md_files),
                "has_obsidian_config": has_obsidian_config
            },
            "files": {},
            "links": {},
            "tags": set(),
            "structure": {}
        }
        
        # Process each markdown file
        processed_files = []
        skipped_files = []
        
        for md_file in md_files:
            try:
                content = md_file.read_text(encoding='utf-8')
                
                # Privacy filtering
                if privacy_keywords and any(keyword in content.lower() or keyword in md_file.name.lower() 
                                         for keyword in privacy_keywords):
                    skipped_files.append(str(md_file.relative_to(vault_path)))
                    continue
                
                # Extract file metadata
                file_info = self._parse_obsidian_file(md_file, content, vault_path)
                vault_data["files"][str(md_file.relative_to(vault_path))] = file_info
                
                # Extract links
                links = self._extract_obsidian_links(content)
                if links:
                    vault_data["links"][str(md_file.relative_to(vault_path))] = links
                
                # Extract tags
                tags = self._extract_obsidian_tags(content)
                vault_data["tags"].update(tags)
                
                processed_files.append(str(md_file.relative_to(vault_path)))
                
            except Exception as e:
                print(f"⚠️  Error processing {md_file.name}: {e}")
                skipped_files.append(str(md_file.relative_to(vault_path)))
        
        # Convert tags set to list for JSON serialization
        vault_data["tags"] = sorted(list(vault_data["tags"]))
        
        # Create import summary
        summary = {
            "processed_files": len(processed_files),
            "skipped_files": len(skipped_files),
            "total_links": sum(len(links) for links in vault_data["links"].values()),
            "total_tags": len(vault_data["tags"]),
            "privacy_filtered": not include_private and len(skipped_files) > 0
        }
        
        vault_data["import_summary"] = summary
        
        # Save vault data
        import_file = output_dir / f"vault-import-{datetime.now().strftime('%Y%m%d-%H%M%S')}.json"
        with open(import_file, 'w', encoding='utf-8') as f:
            json.dump(vault_data, f, indent=2, ensure_ascii=False)
        
        # Generate import report
        report = self._generate_vault_import_report(vault_data, processed_files, skipped_files)
        report_file = output_dir / f"import-report-{datetime.now().strftime('%Y%m%d-%H%M%S')}.md"
        
        if self.show_final_file:
            self.save_content_with_preview(report, str(report_file), "import report")
        else:
            with open(report_file, 'w', encoding='utf-8') as f:
                f.write(report)
        
        print(f"✅ Vault import complete!")
        print(f"📊 Processed: {len(processed_files)} files")
        print(f"⏭️  Skipped: {len(skipped_files)} files")
        print(f"💾 Data saved to: {import_file}")
        print(f"📄 Report saved to: {report_file}")
        
        return {
            "vault_data": vault_data,
            "import_file": str(import_file),
            "report_file": str(report_file),
            "summary": summary
        }
    
    def _parse_obsidian_file(self, file_path: Path, content: str, vault_root: Path) -> Dict[str, Any]:
        """Parse individual Obsidian file for metadata"""
        
        lines = content.split('\n')
        
        # Extract frontmatter if present
        frontmatter = {}
        if content.startswith('---'):
            try:
                end_idx = content.find('---', 3)
                if end_idx > 0:
                    import yaml
                    frontmatter_text = content[3:end_idx].strip()
                    frontmatter = yaml.safe_load(frontmatter_text) or {}
            except:
                pass  # Ignore frontmatter parsing errors
        
        # Basic file info
        file_info = {
            "name": file_path.name,
            "path": str(file_path.relative_to(vault_root)),
            "size": len(content),
            "line_count": len(lines),
            "modified": datetime.fromtimestamp(file_path.stat().st_mtime).isoformat(),
            "frontmatter": frontmatter
        }
        
        # Extract first heading as title
        for line in lines:
            if line.startswith('# '):
                file_info["title"] = line[2:].strip()
                break
        
        # Count content types
        file_info["content_stats"] = {
            "headings": len([line for line in lines if line.startswith('#')]),
            "bullet_points": len([line for line in lines if line.strip().startswith('-') or line.strip().startswith('*')]),
            "code_blocks": content.count('```'),
            "links": len(self._extract_obsidian_links(content)),
            "tags": len(self._extract_obsidian_tags(content))
        }
        
        return file_info
    
    def _extract_obsidian_links(self, content: str) -> List[str]:
        """Extract Obsidian-style links [[link]]"""
        import re
        links = re.findall(r'\[\[([^\]]+)\]\]', content)
        return [link.split('|')[0] for link in links]  # Remove display text
    
    def _extract_obsidian_tags(self, content: str) -> List[str]:
        """Extract Obsidian-style tags #tag"""
        import re
        tags = re.findall(r'#([a-zA-Z0-9_/-]+)', content)
        return tags
    
    def _generate_vault_import_report(self, vault_data: Dict, processed_files: List[str], skipped_files: List[str]) -> str:
        """Generate a markdown report of the vault import"""
        
        metadata = vault_data["metadata"]
        summary = vault_data["import_summary"]
        
        report = f"""# Obsidian Vault Import Report
        
## Import Summary
- **Vault Path**: `{metadata['vault_path']}`
- **Import Date**: {metadata['import_timestamp']}
- **Total Files Found**: {metadata['total_files']}
- **Files Processed**: {summary['processed_files']}
- **Files Skipped**: {summary['skipped_files']}
- **Total Links**: {summary['total_links']}
- **Unique Tags**: {summary['total_tags']}
- **Privacy Filtering**: {'Enabled' if summary['privacy_filtered'] else 'Disabled'}

## File Structure
"""
        
        # Add file list
        if processed_files:
            report += "\n### Processed Files\n"
            for file_path in sorted(processed_files):
                file_info = vault_data["files"].get(file_path, {})
                size = file_info.get("size", 0)
                links = len(vault_data["links"].get(file_path, []))
                report += f"- `{file_path}` ({size} chars, {links} links)\n"
        
        if skipped_files:
            report += "\n### Skipped Files\n"
            for file_path in sorted(skipped_files):
                report += f"- `{file_path}` (privacy filtered)\n"
        
        # Add tags summary
        if vault_data["tags"]:
            report += f"\n## Tags Found ({len(vault_data['tags'])})\n"
            for tag in sorted(vault_data["tags"][:20]):  # Limit to first 20
                report += f"- #{tag}\n"
            if len(vault_data["tags"]) > 20:
                report += f"- ... and {len(vault_data['tags']) - 20} more\n"
        
        # Add most connected files
        link_counts = {file: len(links) for file, links in vault_data["links"].items()}
        if link_counts:
            report += "\n## Most Connected Files\n"
            sorted_files = sorted(link_counts.items(), key=lambda x: x[1], reverse=True)[:10]
            for file_path, link_count in sorted_files:
                report += f"- `{file_path}` ({link_count} outgoing links)\n"
        
        report += f"\n## Usage\n"
        report += f"This import can now be used with uroboro's context system to enhance content generation with your Obsidian knowledge base.\n"
        
        return report
    
    def universal_ingest(self, source_path: str, output_dir: str = None, project_name: str = None) -> Dict[str, Any]:
        """Universal file ingestion - the 'monkey dump stuff here' bucket"""
        
        source_path = Path(source_path).expanduser()
        if not source_path.exists():
            return {"error": f"Source path does not exist: {source_path}"}
        
        if not output_dir:
            output_dir = Path("output") / "universal-ingest"
        else:
            output_dir = Path(output_dir)
        
        if not project_name:
            project_name = source_path.name if source_path.is_dir() else source_path.stem
            
        # Create organized output structure
        ingest_dir = output_dir / f"ingest-{project_name}-{datetime.now().strftime('%Y%m%d-%H%M%S')}"
        ingest_dir.mkdir(parents=True, exist_ok=True)
        
        print(f"🐒 MONKEY DUMP MODE: Ingesting everything from {source_path}")
        print(f"📁 Organizing into: {ingest_dir}")
        
        # Collect all files recursively
        if source_path.is_file():
            all_files = [source_path]
        else:
            all_files = []
            for pattern in ["**/*"]:
                all_files.extend(source_path.glob(pattern))
            all_files = [f for f in all_files if f.is_file()]
        
        print(f"🔍 Found {len(all_files)} files to process")
        
        # Initialize ingestion results
        ingest_data = {
            "metadata": {
                "source_path": str(source_path),
                "project_name": project_name,
                "ingest_timestamp": datetime.now().isoformat(),
                "total_files_found": len(all_files)
            },
            "processed_files": {},
            "file_types": {},
            "parsing_results": {},
            "organization": {
                "markdown": [],
                "text": [],
                "data": [],
                "images": [],
                "documents": [],
                "design": [],
                "code": [],
                "other": []
            }
        }
        
        # Process each file
        processed_count = 0
        skipped_count = 0
        
        for file_path in all_files:
            try:
                # Skip hidden files and system files
                if file_path.name.startswith('.') and file_path.name not in ['.md', '.txt']:
                    skipped_count += 1
                    continue
                
                file_info = self._analyze_and_ingest_file(file_path, source_path, ingest_dir)
                
                if file_info:
                    relative_path = str(file_path.relative_to(source_path))
                    ingest_data["processed_files"][relative_path] = file_info
                    
                    # Track file types
                    file_type = file_info.get("file_type", "unknown")
                    ingest_data["file_types"][file_type] = ingest_data["file_types"].get(file_type, 0) + 1
                    
                    # Organize by category
                    category = file_info.get("category", "other")
                    ingest_data["organization"][category].append(relative_path)
                    
                    processed_count += 1
                    
                    if processed_count % 10 == 0:
                        print(f"  📄 Processed {processed_count} files...")
                        
            except Exception as e:
                print(f"⚠️  Error processing {file_path.name}: {e}")
                skipped_count += 1
        
        # Generate summary
        summary = {
            "processed_files": processed_count,
            "skipped_files": skipped_count,
            "file_types": dict(ingest_data["file_types"]),
            "total_content_size": sum(info.get("size", 0) for info in ingest_data["processed_files"].values()),
            "categories": {cat: len(files) for cat, files in ingest_data["organization"].items() if files}
        }
        
        ingest_data["summary"] = summary
        
        # Save ingestion data
        data_file = ingest_dir / "ingestion-data.json"
        with open(data_file, 'w', encoding='utf-8') as f:
            json.dump(ingest_data, f, indent=2, ensure_ascii=False)
        
        # Generate ingestion report
        report = self._generate_ingestion_report(ingest_data)
        report_file = ingest_dir / "ingestion-report.md"
        
        if self.show_final_file:
            self.save_content_with_preview(report, str(report_file), "ingestion report")
        else:
            with open(report_file, 'w', encoding='utf-8') as f:
                f.write(report)
        
        print(f"\n✅ Universal ingestion complete!")
        print(f"📊 Processed: {processed_count} files")
        print(f"⏭️  Skipped: {skipped_count} files")
        print(f"💾 Data saved to: {data_file}")
        print(f"📄 Report saved to: {report_file}")
        
        return {
            "ingest_data": ingest_data,
            "data_file": str(data_file),
            "report_file": str(report_file),
            "summary": summary
        }
    
    def _analyze_and_ingest_file(self, file_path: Path, source_root: Path, output_dir: Path) -> Dict[str, Any]:
        """Analyze and ingest a single file with smart parsing"""
        
        file_info = {
            "name": file_path.name,
            "path": str(file_path.relative_to(source_root)),
            "size": file_path.stat().st_size,
            "modified": datetime.fromtimestamp(file_path.stat().st_mtime).isoformat(),
            "extension": file_path.suffix.lower()
        }
        
        # Determine file type and category
        ext = file_path.suffix.lower()
        
        # Smart file type detection
        if ext in ['.md', '.markdown']:
            file_info.update(self._parse_markdown_file(file_path))
            file_info["file_type"] = "markdown"
            file_info["category"] = "markdown"
            
        elif ext in ['.txt', '.text']:
            file_info.update(self._parse_text_file(file_path))
            file_info["file_type"] = "text"
            file_info["category"] = "text"
            
        elif ext in ['.csv', '.tsv']:
            file_info.update(self._parse_data_file(file_path))
            file_info["file_type"] = "data"
            file_info["category"] = "data"
            
        elif ext in ['.json', '.yaml', '.yml']:
            file_info.update(self._parse_structured_file(file_path))
            file_info["file_type"] = "structured_data"
            file_info["category"] = "data"
            
        elif ext in ['.png', '.jpg', '.jpeg', '.gif', '.svg', '.webp']:
            file_info.update(self._parse_image_file(file_path))
            file_info["file_type"] = "image"
            file_info["category"] = "images"
            
        elif ext in ['.pdf', '.doc', '.docx']:
            file_info.update(self._parse_document_file(file_path))
            file_info["file_type"] = "document"
            file_info["category"] = "documents"
            
        elif ext in ['.fig', '.sketch', '.xd']:  # Design files (Figma, Sketch, XD)
            file_info.update(self._parse_design_file(file_path))
            file_info["file_type"] = "design"
            file_info["category"] = "design"
            
        elif ext in ['.py', '.js', '.ts', '.html', '.css', '.java', '.cpp', '.rs', '.go']:
            file_info.update(self._parse_code_file(file_path))
            file_info["file_type"] = "code"
            file_info["category"] = "code"
            
        else:
            file_info["file_type"] = "unknown"
            file_info["category"] = "other"
            file_info["parsing_notes"] = "Unknown file type - basic metadata only"
        
        return file_info
    
    def _parse_markdown_file(self, file_path: Path) -> Dict[str, Any]:
        """Parse markdown file for content and structure"""
        try:
            content = file_path.read_text(encoding='utf-8')
            lines = content.split('\n')
            
            info = {
                "content_preview": content[:500] + "..." if len(content) > 500 else content,
                "line_count": len(lines),
                "char_count": len(content)
            }
            
            # Extract frontmatter
            if content.startswith('---'):
                try:
                    end_idx = content.find('---', 3)
                    if end_idx > 0:
                        import yaml
                        frontmatter_text = content[3:end_idx].strip()
                        info["frontmatter"] = yaml.safe_load(frontmatter_text) or {}
                except:
                    pass
            
            # Extract headings
            headings = [line for line in lines if line.startswith('#')]
            info["headings"] = headings[:10]  # First 10 headings
            info["heading_count"] = len(headings)
            
            # Extract links (both markdown and Obsidian style)
            import re
            md_links = re.findall(r'\[([^\]]+)\]\(([^)]+)\)', content)
            obsidian_links = re.findall(r'\[\[([^\]]+)\]\]', content)
            info["markdown_links"] = [{"text": text, "url": url} for text, url in md_links]
            info["obsidian_links"] = obsidian_links
            info["total_links"] = len(md_links) + len(obsidian_links)
            
            # Extract tags
            tags = re.findall(r'#([a-zA-Z0-9_/-]+)', content)
            info["tags"] = list(set(tags))  # Remove duplicates
            
            # Extract code blocks
            code_blocks = re.findall(r'```(\w+)?\n(.*?)```', content, re.DOTALL)
            info["code_blocks"] = [{"language": lang or "unknown", "preview": code[:100]} for lang, code in code_blocks]
            
            return info
            
        except Exception as e:
            return {"parsing_error": str(e)}
    
    def _parse_text_file(self, file_path: Path) -> Dict[str, Any]:
        """Parse plain text file"""
        try:
            content = file_path.read_text(encoding='utf-8')
            lines = content.split('\n')
            
            return {
                "content_preview": content[:300] + "..." if len(content) > 300 else content,
                "line_count": len(lines),
                "char_count": len(content),
                "avg_line_length": sum(len(line) for line in lines) / len(lines) if lines else 0
            }
        except Exception as e:
            return {"parsing_error": str(e)}
    
    def _parse_data_file(self, file_path: Path) -> Dict[str, Any]:
        """Parse CSV/TSV and other data files"""
        try:
            import csv
            
            with open(file_path, 'r', encoding='utf-8') as f:
                # Detect delimiter
                sample = f.read(1024)
                f.seek(0)
                
                delimiter = ','
                if '\t' in sample:
                    delimiter = '\t'
                elif ';' in sample:
                    delimiter = ';'
                
                reader = csv.reader(f, delimiter=delimiter)
                rows = list(reader)
                
                if rows:
                    headers = rows[0] if rows else []
                    data_rows = rows[1:] if len(rows) > 1 else []
                    
                    return {
                        "delimiter": delimiter,
                        "columns": headers,
                        "column_count": len(headers),
                        "row_count": len(data_rows),
                        "preview_rows": data_rows[:3],  # First 3 data rows
                        "total_size": len(rows)
                    }
                
        except Exception as e:
            return {"parsing_error": str(e)}
    
    def _parse_structured_file(self, file_path: Path) -> Dict[str, Any]:
        """Parse JSON/YAML files"""
        try:
            content = file_path.read_text(encoding='utf-8')
            
            if file_path.suffix.lower() == '.json':
                data = json.loads(content)
                return {
                    "format": "json",
                    "keys": list(data.keys()) if isinstance(data, dict) else None,
                    "type": type(data).__name__,
                    "preview": str(data)[:200] + "..." if len(str(data)) > 200 else str(data)
                }
            else:  # YAML
                import yaml
                data = yaml.safe_load(content)
                return {
                    "format": "yaml",
                    "keys": list(data.keys()) if isinstance(data, dict) else None,
                    "type": type(data).__name__,
                    "preview": str(data)[:200] + "..." if len(str(data)) > 200 else str(data)
                }
                
        except Exception as e:
            return {"parsing_error": str(e)}
    
    def _parse_image_file(self, file_path: Path) -> Dict[str, Any]:
        """Parse image files (basic metadata)"""
        try:
            # Basic image info - would need PIL/Pillow for dimensions
            return {
                "image_type": file_path.suffix.lower(),
                "size_bytes": file_path.stat().st_size,
                "parsing_notes": "Image detected - install Pillow for dimensions"
            }
        except Exception as e:
            return {"parsing_error": str(e)}
    
    def _parse_document_file(self, file_path: Path) -> Dict[str, Any]:
        """Parse document files (PDF, Word, etc.)"""
        return {
            "document_type": file_path.suffix.lower(),
            "size_bytes": file_path.stat().st_size,
            "parsing_notes": "Document detected - would need specialized libraries for content extraction"
        }
    
    def _parse_design_file(self, file_path: Path) -> Dict[str, Any]:
        """Parse design files (Figma, Sketch, etc.)"""
        return {
            "design_type": file_path.suffix.lower(),
            "size_bytes": file_path.stat().st_size,
            "parsing_notes": "Design file detected - Figma API integration planned for future versions"
        }
    
    def _parse_code_file(self, file_path: Path) -> Dict[str, Any]:
        """Parse code files"""
        try:
            content = file_path.read_text(encoding='utf-8')
            lines = content.split('\n')
            
            return {
                "language": file_path.suffix.lower(),
                "line_count": len(lines),
                "char_count": len(content),
                "content_preview": content[:300] + "..." if len(content) > 300 else content,
                "functions": len([line for line in lines if 'def ' in line or 'function ' in line])
            }
        except Exception as e:
            return {"parsing_error": str(e)}
    
    def _generate_ingestion_report(self, ingest_data: Dict) -> str:
        """Generate a comprehensive ingestion report"""
        
        metadata = ingest_data["metadata"]
        summary = ingest_data["summary"]
        organization = ingest_data["organization"]
        
        report = f"""# Universal Ingestion Report: {metadata['project_name']}

## 🐒 Monkey Dump Summary
- **Source**: `{metadata['source_path']}`
- **Ingested**: {metadata['ingest_timestamp']}
- **Files Found**: {metadata['total_files_found']}
- **Files Processed**: {summary['processed_files']}
- **Files Skipped**: {summary['skipped_files']}
- **Total Content Size**: {summary['total_content_size']:,} bytes

## 📊 File Type Distribution
"""
        
        for file_type, count in summary["file_types"].items():
            report += f"- **{file_type}**: {count} files\n"
        
        report += f"\n## 🗂️ Content Organization\n"
        
        for category, files in organization.items():
            if files:
                report += f"\n### {category.title()} ({len(files)} files)\n"
                for file_path in sorted(files[:20]):  # Show first 20 files per category
                    file_info = ingest_data["processed_files"].get(file_path, {})
                    size = file_info.get("size", 0)
                    report += f"- `{file_path}` ({size:,} bytes)\n"
                
                if len(files) > 20:
                    report += f"- ... and {len(files) - 20} more files\n"
        
        # Add intelligent insights
        report += f"\n## 🧠 Intelligent Insights\n"
        
        # Markdown insights
        md_files = organization.get("markdown", [])
        if md_files:
            total_headings = sum(ingest_data["processed_files"][f].get("heading_count", 0) for f in md_files)
            total_links = sum(ingest_data["processed_files"][f].get("total_links", 0) for f in md_files)
            report += f"- **Markdown Knowledge Base**: {len(md_files)} files with {total_headings} headings and {total_links} links\n"
        
        # Tag analysis
        all_tags = set()
        for file_path, file_info in ingest_data["processed_files"].items():
            if "tags" in file_info:
                all_tags.update(file_info["tags"])
        
        if all_tags:
            report += f"- **Tags Found**: {len(all_tags)} unique tags across all files\n"
            top_tags = sorted(list(all_tags))[:10]
            report += f"  - Popular tags: {', '.join(f'#{tag}' for tag in top_tags)}\n"
        
        # Data insights
        data_files = organization.get("data", [])
        if data_files:
            total_rows = sum(ingest_data["processed_files"][f].get("row_count", 0) for f in data_files)
            report += f"- **Data Files**: {len(data_files)} files with ~{total_rows} total rows of data\n"
        
        report += f"\n## 🚀 Next Steps\n"
        report += f"- Use `uro generate` to create content from this ingested knowledge\n"
        report += f"- Use `uro mine --path .` to analyze patterns across all content\n"
        report += f"- Reference specific files in prompts for targeted content generation\n"
        
        if organization.get("design"):
            report += f"- Design files detected - Figma API integration coming soon!\n"
        
        report += f"\n## 📁 File Structure\n"
        report += f"All processed files have been analyzed and organized. "
        report += f"Raw data available in `ingestion-data.json` for programmatic access.\n"
        
        return report
    
    def _markdown_to_html(self, markdown_content: str) -> str:
        """Simple markdown to HTML conversion"""
        html = markdown_content
        
        # Convert headers
        html = re.sub(r'^# (.+)$', r'<h1>\1</h1>', html, flags=re.MULTILINE)
        html = re.sub(r'^## (.+)$', r'<h2>\1</h2>', html, flags=re.MULTILINE)
        html = re.sub(r'^### (.+)$', r'<h3>\1</h3>', html, flags=re.MULTILINE)
        
        # Convert bold and italic
        html = re.sub(r'\*\*(.+?)\*\*', r'<strong>\1</strong>', html)
        html = re.sub(r'\*(.+?)\*', r'<em>\1</em>', html)
        
        # Convert code blocks
        html = re.sub(r'```(\w+)?\n(.*?)\n```', r'<pre><code>\2</code></pre>', html, flags=re.DOTALL)
        html = re.sub(r'`(.+?)`', r'<code>\1</code>', html)
        
        # Convert links
        html = re.sub(r'\[(.+?)\]\((.+?)\)', r'<a href="\2">\1</a>', html)
        
        # Convert line breaks to paragraphs
        paragraphs = html.split('\n\n')
        html_paragraphs = []
        for p in paragraphs:
            p = p.strip()
            if p and not p.startswith('<'):
                p = f'<p>{p}</p>'
            html_paragraphs.append(p)
        
        html = '\n\n'.join(html_paragraphs)
        
        # Wrap in basic HTML structure
        html = f"""<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Development Update</title>
    <style>
        body {{ font-family: Arial, sans-serif; max-width: 800px; margin: 0 auto; padding: 20px; }}
        code {{ background: #f4f4f4; padding: 2px 4px; }}
        pre {{ background: #f4f4f4; padding: 10px; overflow-x: auto; }}
    </style>
</head>
<body>
{html}
</body>
</html>"""
        
        return html 