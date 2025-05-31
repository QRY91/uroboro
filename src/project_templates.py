#!/usr/bin/env python3
"""
Project Templates for uroboro
Sets up new projects with optimal uroboro integration
"""

import json
import os
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Optional


class ProjectTemplates:
    def __init__(self):
        self.templates_dir = Path(__file__).parent.parent / "templates"
        self.builtin_templates = {
            "web": self._web_project_template,
            "api": self._api_project_template,
            "game": self._game_project_template,
            "tool": self._tool_project_template,
            "research": self._research_project_template,
            "mobile": self._mobile_project_template,
        }
    
    def list_templates(self) -> List[str]:
        """List available project templates"""
        return list(self.builtin_templates.keys())
    
    def create_project(self, 
                      project_path: str, 
                      template: str,
                      project_name: str = None,
                      description: str = None,
                      tech_stack: List[str] = None,
                      context: str = None) -> bool:
        """Create a new project from template"""
        
        project_dir = Path(project_path).expanduser()
        
        if not project_name:
            project_name = project_dir.name
        
        # Create project directory if it doesn't exist
        project_dir.mkdir(parents=True, exist_ok=True)
        
        # Create .devlog directory
        devlog_dir = project_dir / ".devlog"
        devlog_dir.mkdir(exist_ok=True)
        
        # Generate template content
        if template in self.builtin_templates:
            template_content = self.builtin_templates[template](
                project_name=project_name,
                description=description,
                tech_stack=tech_stack or [],
                context=context
            )
        else:
            print(f"❌ Unknown template: {template}")
            return False
        
        # Write devlog README
        readme_path = devlog_dir / "README.md"
        with open(readme_path, 'w', encoding='utf-8') as f:
            f.write(template_content)
        
        # Create initial capture file
        today = datetime.now().date()
        capture_path = devlog_dir / f"{today}-capture.md"
        
        initial_capture = f"""## {datetime.now().isoformat()}
Project initialized with uroboro template: {template}

{description or f"New {template} project: {project_name}"}

Ready for development tracking!
"""
        
        with open(capture_path, 'w', encoding='utf-8') as f:
            f.write(initial_capture)
        
        # Create project-specific .gitignore if needed
        gitignore_path = project_dir / ".gitignore"
        if not gitignore_path.exists():
            gitignore_content = self._get_gitignore_for_template(template)
            if gitignore_content:
                with open(gitignore_path, 'w', encoding='utf-8') as f:
                    f.write(gitignore_content)
        
        print(f"✅ Project created: {project_dir}")
        print(f"📁 Devlog ready: {devlog_dir}")
        print(f"📄 Context file: {readme_path}")
        print(f"🎯 First capture: {capture_path}")
        
        return True
    
    def _web_project_template(self, project_name: str, description: str, tech_stack: List[str], context: str) -> str:
        """Template for web applications"""
        return f"""# Project: {project_name}

## Purpose
{description or "A web application project focused on user experience and modern web technologies."}

## Current Focus
- Frontend user interface development
- Backend API implementation
- Database design and optimization
- User authentication and security
- Performance optimization
- Responsive design

## Technical Stack
{self._format_tech_stack(tech_stack or ["React", "Node.js", "Express", "PostgreSQL"])}

## Development Context
{context or "Building a modern web application with focus on clean architecture, user experience, and scalability. Emphasis on both frontend polish and backend reliability."}

## AI Instructions
When generating content about this project, emphasize:
- User experience improvements and design decisions
- Technical architecture choices and trade-offs
- Performance optimization strategies
- Development workflow and best practices
- Integration challenges and solutions

Focus on the journey of building a complete web application rather than just individual features.
"""

    def _api_project_template(self, project_name: str, description: str, tech_stack: List[str], context: str) -> str:
        """Template for API/backend projects"""
        return f"""# Project: {project_name}

## Purpose
{description or "A robust API service designed for scalability, reliability, and developer experience."}

## Current Focus
- API design and endpoint architecture
- Database schema optimization
- Authentication and authorization
- Rate limiting and security
- Documentation and testing
- Deployment and monitoring

## Technical Stack
{self._format_tech_stack(tech_stack or ["Python", "FastAPI", "PostgreSQL", "Redis", "Docker"])}

## Development Context
{context or "Building a production-ready API with emphasis on clean architecture, comprehensive testing, and excellent developer experience. Focus on reliability and performance."}

## AI Instructions
When generating content about this project, emphasize:
- API design decisions and architectural patterns
- Database optimization and query performance
- Security implementation and best practices
- Testing strategies and coverage
- Deployment automation and monitoring
- Developer experience improvements

Focus on the technical depth and production-readiness aspects of API development.
"""

    def _game_project_template(self, project_name: str, description: str, tech_stack: List[str], context: str) -> str:
        """Template for game development projects"""
        return f"""# Project: {project_name}

## Purpose
{description or "An interactive game focusing on engaging gameplay mechanics and polished user experience."}

## Current Focus
- Game mechanics implementation
- Graphics and visual effects
- Audio and sound design
- User interface and controls
- Performance optimization
- Playtesting and balancing

## Technical Stack
{self._format_tech_stack(tech_stack or ["Unity", "C#", "Blender", "FMOD"])}

## Development Context
{context or "Creating an engaging game experience with focus on innovative mechanics, polished visuals, and smooth gameplay. Balancing technical implementation with creative design."}

## AI Instructions
When generating content about this project, emphasize:
- Game design decisions and player experience
- Technical implementation of gameplay mechanics
- Performance optimization for smooth gameplay
- Creative problem-solving in game development
- Iterative design and playtesting insights
- Visual and audio design integration

Focus on the creative and technical aspects of game development rather than just coding.
"""

    def _tool_project_template(self, project_name: str, description: str, tech_stack: List[str], context: str) -> str:
        """Template for developer tools and utilities"""
        return f"""# Project: {project_name}

## Purpose
{description or "A developer tool designed to improve productivity and solve specific development workflow challenges."}

## Current Focus
- Core functionality implementation
- Command-line interface design
- Configuration and customization
- Documentation and examples
- Testing and reliability
- Distribution and packaging

## Technical Stack
{self._format_tech_stack(tech_stack or ["Python", "Click", "PyTest", "Poetry"])}

## Development Context
{context or "Building a reliable developer tool with focus on usability, performance, and solving real workflow problems. Emphasis on clean APIs and excellent documentation."}

## AI Instructions
When generating content about this project, emphasize:
- Problem-solving approach and user needs
- API design and interface decisions
- Development workflow improvements
- Technical implementation details
- Testing and reliability strategies
- Documentation and user experience

Focus on how the tool solves real problems and improves developer productivity.
"""

    def _research_project_template(self, project_name: str, description: str, tech_stack: List[str], context: str) -> str:
        """Template for research and experimental projects"""
        return f"""# Project: {project_name}

## Purpose
{description or "A research project exploring new concepts, technologies, or methodologies with experimental implementation."}

## Current Focus
- Literature review and background research
- Proof of concept implementation
- Data collection and analysis
- Experimental validation
- Documentation and findings
- Future research directions

## Technical Stack
{self._format_tech_stack(tech_stack or ["Python", "Jupyter", "NumPy", "Pandas", "Matplotlib"])}

## Development Context
{context or "Exploring new ideas through experimental implementation and rigorous testing. Focus on learning, discovery, and documenting insights for future work."}

## AI Instructions
When generating content about this project, emphasize:
- Research methodology and experimental approach
- Learning and discovery process
- Technical challenges and solutions
- Data analysis and insights
- Theoretical implications and applications
- Future research opportunities

Focus on the investigative and learning aspects rather than just implementation details.
"""

    def _mobile_project_template(self, project_name: str, description: str, tech_stack: List[str], context: str) -> str:
        """Template for mobile application projects"""
        return f"""# Project: {project_name}

## Purpose
{description or "A mobile application designed for excellent user experience across different devices and platforms."}

## Current Focus
- Mobile UI/UX implementation
- Cross-platform compatibility
- Performance optimization
- Offline functionality
- App store guidelines compliance
- User testing and feedback

## Technical Stack
{self._format_tech_stack(tech_stack or ["React Native", "TypeScript", "AsyncStorage", "Firebase"])}

## Development Context
{context or "Building a mobile app with focus on native performance, intuitive user interface, and cross-platform compatibility. Emphasis on mobile-specific design patterns."}

## AI Instructions
When generating content about this project, emphasize:
- Mobile-specific design decisions and constraints
- Cross-platform development challenges
- Performance optimization for mobile devices
- User experience and interface design
- App store submission and compliance
- Mobile testing and debugging strategies

Focus on mobile-specific development challenges and user experience considerations.
"""

    def _format_tech_stack(self, tech_stack: List[str]) -> str:
        """Format technology stack as markdown list"""
        if not tech_stack:
            return "- Technology stack to be determined"
        
        return "\n".join(f"- {tech}" for tech in tech_stack)
    
    def _get_gitignore_for_template(self, template: str) -> Optional[str]:
        """Get appropriate .gitignore content for template type"""
        gitignores = {
            "web": """# Dependencies
node_modules/
npm-debug.log*
yarn-debug.log*
yarn-error.log*

# Production builds
build/
dist/
.next/

# Environment variables
.env
.env.local
.env.production

# IDE
.vscode/
.idea/

# OS
.DS_Store
Thumbs.db
""",
            "api": """# Python
__pycache__/
*.py[cod]
*$py.class
*.so
.Python
env/
venv/
.venv/
.env

# Database
*.db
*.sqlite3

# IDE
.vscode/
.idea/

# OS
.DS_Store
Thumbs.db
""",
            "tool": """# Python
__pycache__/
*.py[cod]
*$py.class
*.so
.Python
build/
develop-eggs/
dist/
downloads/
eggs/
.eggs/
lib/
lib64/
parts/
sdist/
var/
wheels/
*.egg-info/
.installed.cfg
*.egg

# Virtual environments
env/
venv/
.venv/

# IDE
.vscode/
.idea/

# OS
.DS_Store
Thumbs.db
""",
            "game": """# Unity
[Ll]ibrary/
[Tt]emp/
[Oo]bj/
[Bb]uild/
[Bb]uilds/
Assets/AssetStoreTools*

# Visual Studio
.vs/
*.sln
*.csproj
*.userprefs

# IDE
.vscode/
.idea/

# OS
.DS_Store
Thumbs.db
""",
            "research": """# Python
__pycache__/
*.py[cod]
*$py.class

# Jupyter
.ipynb_checkpoints/

# Data
*.csv
*.json
*.xlsx
data/
datasets/

# Models
models/
*.pkl
*.model

# Virtual environments
env/
venv/
.venv/

# IDE
.vscode/
.idea/

# OS
.DS_Store
Thumbs.db
""",
            "mobile": """# React Native
node_modules/
npm-debug.log*
yarn-debug.log*
yarn-error.log*

# iOS
ios/build/
ios/Pods/
*.xcworkspace
*.xcuserstate

# Android
android/app/build/
android/build/
android/.gradle/

# Metro
.metro-health-check*

# Environment
.env

# IDE
.vscode/
.idea/

# OS
.DS_Store
Thumbs.db
"""
        }
        
        return gitignores.get(template)