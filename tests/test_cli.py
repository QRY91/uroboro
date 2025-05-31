#!/usr/bin/env python3
"""
Test suite for uroboro CLI functionality
"""

import pytest
import subprocess
import tempfile
import os
from pathlib import Path
import json


class TestUroboroCLI:
    """Test suite for uroboro CLI commands"""
    
    def setup_method(self):
        """Setup test environment for each test"""
        self.temp_dir = tempfile.mkdtemp()
        self.original_cwd = os.getcwd()
        os.chdir(self.temp_dir)
        
        # Create basic project structure
        os.makedirs(".devlog", exist_ok=True)
        
    def teardown_method(self):
        """Cleanup after each test"""
        os.chdir(self.original_cwd)
        
    def run_uro_command(self, args, input_text=None):
        """Helper to run uroboro commands"""
        cmd = ["python", "-m", "src.cli"] + args
        result = subprocess.run(
            cmd,
            capture_output=True,
            text=True,
            input=input_text,
            cwd=self.original_cwd
        )
        return result
    
    def test_cli_help(self):
        """Test that CLI help works"""
        result = self.run_uro_command(["--help"])
        assert result.returncode == 0
        assert "uroboro" in result.stdout
        assert "Self-Documenting Content Pipeline" in result.stdout
    
    def test_capture_command(self):
        """Test basic capture functionality"""
        result = self.run_uro_command([
            "capture", 
            "Test capture content for CLI testing"
        ])
        assert result.returncode == 0
        assert "✅ Captured" in result.stdout
    
    def test_capture_with_project_and_tags(self):
        """Test capture with project and tags"""
        result = self.run_uro_command([
            "capture", 
            "Test capture with metadata",
            "--project", "test-project",
            "--tags", "testing", "cli"
        ])
        assert result.returncode == 0
        assert "✅ Captured" in result.stdout
    
    def test_status_command(self):
        """Test status command"""
        result = self.run_uro_command(["status"])
        assert result.returncode == 0
        assert "uroboro status" in result.stdout
    
    def test_status_verbose(self):
        """Test status command with verbose flag"""
        result = self.run_uro_command(["status", "--verbose"])
        assert result.returncode == 0
        assert "uroboro status" in result.stdout
    
    def test_generate_preview(self):
        """Test generate command with preview"""
        # First capture some content
        self.run_uro_command([
            "capture", 
            "Sample content for generation testing"
        ])
        
        result = self.run_uro_command([
            "generate", 
            "--preview",
            "--output", "devlog"
        ])
        assert result.returncode == 0
    
    def test_tracking_status(self):
        """Test tracking status command"""
        result = self.run_uro_command(["tracking"])
        assert result.returncode == 0
        assert "Usage Tracking Status" in result.stdout
    
    def test_git_status_not_repo(self):
        """Test git command when not in a git repo"""
        result = self.run_uro_command(["git"])
        assert result.returncode == 0
        assert "Not in a git repository" in result.stdout
    
    def test_invalid_command(self):
        """Test handling of invalid commands"""
        result = self.run_uro_command(["nonexistent-command"])
        assert result.returncode == 1
        assert "Unknown command" in result.stdout


class TestContentAggregator:
    """Test the content aggregation functionality"""
    
    def setup_method(self):
        """Setup test environment"""
        self.temp_dir = tempfile.mkdtemp()
        self.original_cwd = os.getcwd()
        os.chdir(self.temp_dir)
        
        # Create test notes structure
        notes_dir = Path("notes")
        notes_dir.mkdir(exist_ok=True)
        
        daily_dir = notes_dir / "daily"
        daily_dir.mkdir(exist_ok=True)
        
        # Create test content
        with open(daily_dir / "2025-05-31.md", "w") as f:
            f.write("""## 2025-05-31T10:00:00.000000
Test capture for aggregator testing

## 2025-05-31T11:00:00.000000
Another test capture with different content
""")
    
    def teardown_method(self):
        """Cleanup after each test"""
        os.chdir(self.original_cwd)
    
    def test_aggregator_import(self):
        """Test that we can import the aggregator"""
        import sys
        sys.path.insert(0, self.original_cwd)
        
        from src.aggregator import ContentAggregator
        aggregator = ContentAggregator()
        assert aggregator is not None
    
    def test_collect_recent_activity(self):
        """Test collecting recent activity"""
        import sys
        sys.path.insert(0, self.original_cwd)
        
        from src.aggregator import ContentAggregator
        aggregator = ContentAggregator()
        
        activity = aggregator.collect_recent_activity(days=7)
        assert isinstance(activity, dict)
        assert "daily_notes" in activity


class TestGitIntegration:
    """Test git integration functionality"""
    
    def setup_method(self):
        """Setup test git repository"""
        self.temp_dir = tempfile.mkdtemp()
        self.original_cwd = os.getcwd()
        os.chdir(self.temp_dir)
        
        # Initialize git repo
        subprocess.run(["git", "init"], check=True, capture_output=True)
        subprocess.run(["git", "config", "user.name", "Test User"], check=True)
        subprocess.run(["git", "config", "user.email", "test@example.com"], check=True)
        
        # Create initial commit
        with open("README.md", "w") as f:
            f.write("# Test Repository")
        subprocess.run(["git", "add", "README.md"], check=True)
        subprocess.run(["git", "commit", "-m", "Initial commit"], check=True)
    
    def teardown_method(self):
        """Cleanup after each test"""
        os.chdir(self.original_cwd)
    
    def test_git_integration_import(self):
        """Test that we can import git integration"""
        import sys
        sys.path.insert(0, self.original_cwd)
        
        from src.git_integration import GitIntegration
        git = GitIntegration()
        assert git.is_git_repo is True
    
    def test_get_recent_commits(self):
        """Test getting recent commits"""
        import sys
        sys.path.insert(0, self.original_cwd)
        
        from src.git_integration import GitIntegration
        git = GitIntegration()
        
        commits = git.get_recent_commits(days=7)
        assert isinstance(commits, list)
        assert len(commits) >= 1  # Should have at least our initial commit


class TestProjectTemplates:
    """Test project template functionality"""
    
    def setup_method(self):
        """Setup test environment"""
        self.temp_dir = tempfile.mkdtemp()
        self.original_cwd = os.getcwd()
        os.chdir(self.temp_dir)
    
    def teardown_method(self):
        """Cleanup after each test"""
        os.chdir(self.original_cwd)
    
    def test_project_templates_import(self):
        """Test that we can import project templates"""
        import sys
        sys.path.insert(0, self.original_cwd)
        
        from src.project_templates import ProjectTemplates
        templates = ProjectTemplates()
        assert templates is not None
    
    def test_list_templates(self):
        """Test listing available templates"""
        import sys
        sys.path.insert(0, self.original_cwd)
        
        from src.project_templates import ProjectTemplates
        templates = ProjectTemplates()
        
        template_list = templates.list_templates()
        assert isinstance(template_list, list)
        assert "web" in template_list
        assert "api" in template_list
        assert "tool" in template_list
    
    def test_create_project(self):
        """Test creating a project from template"""
        import sys
        sys.path.insert(0, self.original_cwd)
        
        from src.project_templates import ProjectTemplates
        templates = ProjectTemplates()
        
        project_path = "test-web-project"
        success = templates.create_project(
            project_path=project_path,
            template="web",
            project_name="Test Web App",
            description="A test web application"
        )
        
        assert success is True
        assert Path(project_path).exists()
        assert Path(project_path, ".devlog").exists()
        assert Path(project_path, ".devlog", "README.md").exists()


if __name__ == "__main__":
    pytest.main([__file__, "-v"]) 