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
        
    def _call_ollama(self, prompt: str, model: str = None) -> str:
        """Call ollama with the given prompt"""
        model = model or self.llm_model
        
        try:
            result = subprocess.run([
                "ollama", "run", model
            ], input=prompt, capture_output=True, text=True, timeout=60)
            
            if result.returncode == 0:
                return result.stdout.strip()
            else:
                print(f"Ollama error: {result.stderr}")
                return "Error: Could not generate content"
                
        except subprocess.TimeoutExpired:
            return "Error: LLM request timed out"
        except FileNotFoundError:
            return "Error: Ollama not found. Please install ollama first."
    
    def generate_devlog_summary(self, activity_data: Dict) -> str:
        """Generate a development log summary from aggregated activity"""
        
        # Extract content for summarization
        project_content = []
        for project_name, project_data in activity_data.get("projects", {}).items():
            if "devlog" in project_data:
                for entry in project_data["devlog"]:
                    project_content.append(f"**{project_name}**: {entry['content']}")
        
        daily_content = []
        for note in activity_data.get("daily_notes", []):
            daily_content.append(note['content'])
        
        all_content = "\n".join(project_content + daily_content)
        
        prompt = f"""
        Analyze this development activity and create a concise development log summary:

        {all_content}

        Create a structured summary with:
        ## Technical Work
        - Brief bullet points of what was accomplished

        ## Key Insights  
        - Important discoveries or decisions made

        ## Next Steps
        - What should be tackled next

        Keep it professional but conversational. Focus on the most significant items.
        """
        
        return self._call_ollama(prompt)
    
    def generate_blog_post(self, activity_data: Dict, title: str = None, tags: List[str] = None, format: str = "mdx", voice: str = None) -> str:
        """Generate a full blog post from aggregated activity"""
        
        # Auto-generate title if not provided
        if not title:
            title = f"Development Update - {datetime.now().strftime('%B %d, %Y')}"
        
        # Extract project highlights
        project_highlights = []
        for project_name, project_data in activity_data.get("projects", {}).items():
            if "devlog" in project_data:
                content = "\n".join([entry['content'] for entry in project_data["devlog"]])
                project_highlights.append(f"**{project_name}**: {content}")
        
        all_activity = "\n".join(project_highlights)
        if activity_data.get("daily_notes"):
            daily_content = "\n".join([note['content'] for note in activity_data["daily_notes"]])
            all_activity += f"\n\n**General Notes**: {daily_content}"
        
        # Get style configuration
        style_config = self.config.get("style_config", {})
        voice_name = voice or style_config.get("default_voice", "professional_conversational")
        voice_config = style_config.get("voices", {}).get(voice_name, {})
        
        # Build style instructions
        style_instructions = voice_config.get("prompt_additions", "Write in first person, conversational but professional tone.")
        
        # Add brand voice preferences
        brand_voice = style_config.get("brand_voice", {})
        if brand_voice.get("avoid_phrases"):
            style_instructions += f"\n\nAvoid these phrases: {', '.join(brand_voice['avoid_phrases'])}"
        if brand_voice.get("preferred_phrases"):
            style_instructions += f"\nPrefer phrases like: {', '.join(brand_voice['preferred_phrases'])}"
        if brand_voice.get("personality_traits"):
            style_instructions += f"\nWrite with personality traits: {', '.join(brand_voice['personality_traits'])}"
        
        # Add custom instructions if any
        custom_instructions = style_config.get("custom_instructions", "")
        if custom_instructions:
            style_instructions += f"\n\nAdditional instructions: {custom_instructions}"

        prompt = f"""
        Write a engaging blog post about recent development work. Use this activity data:

        {all_activity}

        Structure the post as:
        1. Brief introduction setting context
        2. Main development highlights organized by project/topic
        3. Technical insights or lessons learned  
        4. What's coming next

        STYLE INSTRUCTIONS:
        {style_instructions}
        
        Aim for 300-500 words.
        """
        
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
    
    def save_blog_post(self, content: str, filename: str = None, output_dir: str = None, format: str = "mdx") -> str:
        """Save blog post to appropriate directory based on format"""
        
        if not filename:
            timestamp = datetime.now().strftime("%Y-%m-%d")
            if format == "mdx":
                filename = f"dev-update-{timestamp}.mdx"
            elif format == "markdown":
                filename = f"dev-update-{timestamp}.md"
            else:
                filename = f"dev-update-{timestamp}.txt"
        
        if not output_dir:
            if format == "mdx":
                # Default to qryzone blog directory for MDX
                output_dir = Path("../qryzone/content/blog")
            else:
                # Plain formats go to local output
                output_dir = Path("output/posts")
        else:
            output_dir = Path(output_dir)
        
        output_dir.mkdir(parents=True, exist_ok=True)
        output_path = output_dir / filename
        
        with open(output_path, 'w', encoding='utf-8') as f:
            f.write(content)
        
        return str(output_path)
    
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