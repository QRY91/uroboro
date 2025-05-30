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
    
    def generate_blog_post(self, activity_data: Dict, title: str = None, tags: List[str] = None) -> str:
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
        
        prompt = f"""
        Write a engaging blog post about recent development work. Use this activity data:

        {all_activity}

        Structure the post as:
        1. Brief introduction setting context
        2. Main development highlights organized by project/topic
        3. Technical insights or lessons learned  
        4. What's coming next

        Write in first person, conversational but professional tone.
        Make it interesting for both technical and non-technical readers.
        Include specific details but explain technical concepts clearly.
        Aim for 300-500 words.
        """
        
        content = self._call_ollama(prompt)
        
        # Generate frontmatter
        frontmatter = self._generate_frontmatter(title, tags)
        
        return f"{frontmatter}\n\n{content}"
    
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
    
    def save_blog_post(self, content: str, filename: str = None, output_dir: str = None) -> str:
        """Save blog post to qryzone content directory"""
        
        if not filename:
            timestamp = datetime.now().strftime("%Y-%m-%d")
            filename = f"dev-update-{timestamp}.mdx"
        
        if not output_dir:
            # Default to qryzone blog directory
            output_dir = Path("../qryzone/content/blog")
        else:
            output_dir = Path(output_dir)
        
        output_dir.mkdir(parents=True, exist_ok=True)
        output_path = output_dir / filename
        
        with open(output_path, 'w', encoding='utf-8') as f:
            f.write(content)
        
        return str(output_path)
    
    def create_social_hooks(self, activity_data: Dict) -> List[str]:
        """Generate social media hooks from activity"""
        
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
        
        prompt = f"""
        Create 3-5 engaging social media hooks from this development work:

        {combined_work}

        Format each as a tweet-style post (under 280 chars):
        - Make them engaging and accessible
        - Include relevant hashtags
        - Show progress/momentum
        - Mix technical and human interest angles

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