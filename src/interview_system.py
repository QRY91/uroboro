"""
Interactive Interview System for Context Extraction
Helps extract narrative and decision-making context from users, 
especially useful for academic documentation and reflection.
"""

import json
import subprocess
import re
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Any, Optional
import sys


class InteractiveInterviewer:
    """Interactive interview system for extracting context and narrative"""
    
    def __init__(self, llm_model: str = "mistral:latest"):
        self.llm_model = llm_model
        self.session_data = {
            "timestamp": datetime.now().isoformat(),
            "questions_asked": [],
            "responses": [],
            "analysis_insights": [],
            "extracted_context": {}
        }
    
    def analyze_materials_for_gaps(self, materials_path: str) -> Dict[str, Any]:
        """Analyze existing materials to identify context gaps and interesting patterns"""
        
        materials_path = Path(materials_path)
        analysis = {
            "file_types_found": [],
            "key_themes": [],
            "potential_gaps": [],
            "interesting_patterns": [],
            "suggested_questions": []
        }
        
        print(f"🔍 Analyzing materials at: {materials_path}")
        
        if not materials_path.exists():
            analysis["error"] = f"Path does not exist: {materials_path}"
            return analysis
        
        # Collect all relevant files
        file_patterns = ["*.md", "*.txt", "*.json", "*.py", "*.ts", "*.js"]
        files_found = []
        
        for pattern in file_patterns:
            files_found.extend(materials_path.rglob(pattern))
        
        analysis["total_files"] = len(files_found)
        
        # Quick analysis of file types and content
        for file_path in files_found[:20]:  # Limit for performance
            try:
                if file_path.suffix.lower() in ['.md', '.txt']:
                    content = file_path.read_text(encoding='utf-8')
                    
                    # Look for interesting patterns
                    if any(keyword in content.lower() for keyword in 
                           ['originally planned', 'changed', 'decided', 'instead', 'problem', 'challenge']):
                        analysis["interesting_patterns"].append({
                            "file": str(file_path.relative_to(materials_path)),
                            "pattern": "Decision/change language detected",
                            "snippet": content[:200] + "..."
                        })
                    
                    # Look for technical decision indicators
                    tech_indicators = ['chose', 'selected', 'technology', 'framework', 'decided to use']
                    if any(indicator in content.lower() for indicator in tech_indicators):
                        analysis["key_themes"].append("Technical decision-making")
                    
                    # Look for AI collaboration indicators
                    ai_indicators = ['ai', 'claude', 'gpt', 'assistant', 'conversation']
                    if any(indicator in content.lower() for indicator in ai_indicators):
                        analysis["key_themes"].append("AI-assisted development")
                        
            except Exception as e:
                continue
        
        # Generate suggested questions based on patterns
        if "Technical decision-making" in analysis["key_themes"]:
            analysis["suggested_questions"].append({
                "category": "technical_decisions",
                "question": "I notice several references to technology choices. What factors most influenced your key technical decisions?",
                "focus": "Understanding decision-making rationale"
            })
        
        if "AI-assisted development" in analysis["key_themes"]:
            analysis["suggested_questions"].append({
                "category": "ai_collaboration", 
                "question": "How did AI assistance actually work in practice? What surprised you about the collaboration?",
                "focus": "Real AI collaboration experience"
            })
        
        if analysis["interesting_patterns"]:
            analysis["suggested_questions"].append({
                "category": "changes_and_adaptations",
                "question": "I see evidence of plans changing during development. What were the key pivots and what drove them?",
                "focus": "Adaptation and learning process"
            })
        
        return analysis
    
    def generate_contextual_questions(self, analysis: Dict[str, Any], interview_type: str = "postmortem") -> List[Dict[str, str]]:
        """Generate targeted questions based on material analysis and interview type"""
        
        questions = []
        
        if interview_type == "postmortem":
            # Questions about completed project reflection
            questions.extend([
                {
                    "id": "overview",
                    "question": "Looking back at this project, what's the story you'd tell about how it unfolded?",
                    "purpose": "Get the narrative overview and emotional journey",
                    "follow_up": "What aspects would be hard for someone to understand just from the files?"
                },
                {
                    "id": "biggest_learning",
                    "question": "What was the biggest thing you learned that wasn't planned?",
                    "purpose": "Capture unexpected insights and growth",
                    "follow_up": "How did that change your approach as you went?"
                }
            ])
        
        elif interview_type == "deviation_analysis":
            # Questions about plan vs reality
            questions.extend([
                {
                    "id": "original_plan",
                    "question": "When you first started this project, what did you think it would look like?",
                    "purpose": "Establish baseline expectations",
                    "follow_up": "What were you most confident about in that original vision?"
                },
                {
                    "id": "major_pivots",
                    "question": "What are the biggest ways the project ended up different from your original plan?",
                    "purpose": "Identify key deviations and turning points",
                    "follow_up": "Which of those changes do you think were most important for the project's success?"
                }
            ])
        
        elif interview_type == "gap_analysis":
            # Questions to fill identified gaps
            questions.extend([
                {
                    "id": "missing_context",
                    "question": "What context about this project would be completely invisible to someone reading the documentation?",
                    "purpose": "Extract tacit knowledge and unspoken context",
                    "follow_up": "What assumptions did you make that someone else might not?"
                }
            ])
        
        # Add questions based on material analysis
        for suggested in analysis.get("suggested_questions", []):
            questions.append({
                "id": suggested["category"],
                "question": suggested["question"],
                "purpose": suggested["focus"],
                "follow_up": "Can you give me a specific example of how that played out?"
            })
        
        return questions
    
    def conduct_interview(self, questions: List[Dict[str, str]], save_responses: bool = True) -> Dict[str, Any]:
        """Conduct interactive interview with the user"""
        
        print("\n🎤 UROBORO INTERVIEW SESSION")
        print("=" * 50)
        print("This interview will help extract the full story behind your project.")
        print("I'll ask targeted questions based on your materials to capture context")
        print("that's often lost in documentation. Feel free to be conversational!")
        print("\nType 'skip' to skip a question, 'done' to finish early, or 'pause' to save and continue later.")
        print("=" * 50)
        
        responses = {}
        
        for i, q in enumerate(questions, 1):
            print(f"\n📝 Question {i}/{len(questions)} [{q['id']}]")
            print(f"🎯 Purpose: {q['purpose']}")
            print("─" * 40)
            print(f"❓ {q['question']}")
            
            # Get user response
            response_lines = []
            print("\n💬 Your response (press Enter twice when done, or type control commands):")
            
            while True:
                try:
                    line = input()
                    if line.lower() in ['skip', 'done', 'pause']:
                        if line.lower() == 'skip':
                            print("⏭️  Skipping this question...")
                            break
                        elif line.lower() == 'done':
                            print("✅ Interview completed early.")
                            self.session_data["completed_early"] = True
                            return self._finalize_interview(responses)
                        elif line.lower() == 'pause':
                            print("⏸️  Pausing interview. You can resume later.")
                            self.session_data["paused"] = True
                            return self._save_partial_interview(responses)
                    
                    if line == "" and response_lines and response_lines[-1] == "":
                        # Two empty lines = end of response
                        break
                    
                    response_lines.append(line)
                        
                except KeyboardInterrupt:
                    print("\n\n🛑 Interview interrupted. Saving partial responses...")
                    self.session_data["interrupted"] = True
                    return self._save_partial_interview(responses)
            
            if response_lines and response_lines[0].lower() != 'skip':
                response_text = "\n".join(response_lines).strip()
                if response_text:
                    responses[q['id']] = {
                        "question": q['question'],
                        "response": response_text,
                        "timestamp": datetime.now().isoformat(),
                        "purpose": q['purpose']
                    }
                    
                    # Ask follow-up if response is substantial
                    if len(response_text) > 50 and q.get('follow_up'):
                        print(f"\n🔍 Follow-up: {q['follow_up']}")
                        follow_up_response = input("💬 Quick follow-up response: ").strip()
                        if follow_up_response:
                            responses[q['id']]["follow_up_response"] = follow_up_response
                    
                    print("✅ Response recorded!")
        
        print("\n🎉 Interview completed!")
        return self._finalize_interview(responses)
    
    def _finalize_interview(self, responses: Dict[str, Any]) -> Dict[str, Any]:
        """Finalize interview session and generate summary"""
        
        self.session_data["responses"] = responses
        self.session_data["completed_at"] = datetime.now().isoformat()
        self.session_data["total_responses"] = len(responses)
        
        # Generate quick summary
        summary = self._generate_interview_summary(responses)
        self.session_data["summary"] = summary
        
        return {
            "session_data": self.session_data,
            "responses": responses,
            "summary": summary
        }
    
    def _save_partial_interview(self, responses: Dict[str, Any]) -> Dict[str, Any]:
        """Save partial interview for resuming later"""
        
        self.session_data["responses"] = responses
        self.session_data["partial_save_at"] = datetime.now().isoformat()
        
        return {
            "session_data": self.session_data,
            "responses": responses,
            "status": "partial"
        }
    
    def _generate_interview_summary(self, responses: Dict[str, Any]) -> str:
        """Generate a summary of interview insights"""
        
        if not responses:
            return "No responses collected."
        
        summary = f"# Interview Summary - {datetime.now().strftime('%Y-%m-%d %H:%M')}\n\n"
        summary += f"**Total Responses:** {len(responses)}\n\n"
        
        summary += "## Key Insights Extracted\n\n"
        
        for response_id, data in responses.items():
            summary += f"### {response_id.replace('_', ' ').title()}\n"
            summary += f"**Question:** {data['question']}\n\n"
            summary += f"**Response:** {data['response'][:300]}{'...' if len(data['response']) > 300 else ''}\n\n"
            
            if data.get('follow_up_response'):
                summary += f"**Follow-up:** {data['follow_up_response']}\n\n"
            
            summary += "---\n\n"
        
        return summary
    
    def save_interview_results(self, results: Dict[str, Any], output_dir: str = None) -> str:
        """Save interview results to files"""
        
        if not output_dir:
            output_dir = Path("output") / "interviews"
        else:
            output_dir = Path(output_dir)
        
        output_dir.mkdir(parents=True, exist_ok=True)
        
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        
        # Save raw session data
        session_file = output_dir / f"interview_session_{timestamp}.json"
        with open(session_file, 'w', encoding='utf-8') as f:
            json.dump(results.get("session_data", {}), f, indent=2, ensure_ascii=False)
        
        # Save formatted summary
        summary_file = output_dir / f"interview_summary_{timestamp}.md"
        with open(summary_file, 'w', encoding='utf-8') as f:
            f.write(results.get("summary", ""))
        
        # Save detailed responses
        responses_file = output_dir / f"interview_responses_{timestamp}.md"
        detailed_content = self._generate_detailed_responses(results.get("responses", {}))
        with open(responses_file, 'w', encoding='utf-8') as f:
            f.write(detailed_content)
        
        print(f"\n💾 Interview results saved:")
        print(f"  📄 Summary: {summary_file}")
        print(f"  📝 Detailed responses: {responses_file}")
        print(f"  🗃️ Raw session data: {session_file}")
        
        return str(summary_file)
    
    def _generate_detailed_responses(self, responses: Dict[str, Any]) -> str:
        """Generate detailed markdown document with all responses"""
        
        content = f"# Detailed Interview Responses - {datetime.now().strftime('%Y-%m-%d %H:%M')}\n\n"
        content += "This document contains the complete interview responses for academic documentation.\n\n"
        
        content += "## Interview Purpose\n"
        content += "This interview was conducted to extract narrative context and decision-making rationale "
        content += "that might not be captured in technical documentation. The goal is to provide rich "
        content += "contextual information for academic analysis and reflection.\n\n"
        
        content += "## Responses\n\n"
        
        for i, (response_id, data) in enumerate(responses.items(), 1):
            content += f"### {i}. {response_id.replace('_', ' ').title()}\n\n"
            content += f"**Purpose:** {data.get('purpose', 'Not specified')}\n\n"
            content += f"**Question:** {data['question']}\n\n"
            content += f"**Response:**\n{data['response']}\n\n"
            
            if data.get('follow_up_response'):
                content += f"**Follow-up Response:** {data['follow_up_response']}\n\n"
            
            content += f"**Timestamp:** {data.get('timestamp', 'Not recorded')}\n\n"
            content += "---\n\n"
        
        content += "## Academic Use\n"
        content += "These responses provide qualitative insights into the development process and "
        content += "decision-making that complement the technical documentation. They can be used "
        content += "for academic analysis of methodology, lessons learned, and reflection.\n\n"
        
        return content
    
    def conduct_full_interview_session(self, materials_path: str, interview_type: str = "postmortem", 
                                     output_dir: str = None) -> str:
        """Complete interview workflow: analyze, question, conduct, save"""
        
        print(f"🎤 Starting {interview_type} interview session...")
        
        # Step 1: Analyze materials
        print("📊 Analyzing existing materials for context gaps...")
        analysis = self.analyze_materials_for_gaps(materials_path)
        
        if "error" in analysis:
            print(f"❌ Error analyzing materials: {analysis['error']}")
            return None
        
        print(f"✅ Analysis complete - found {analysis['total_files']} files")
        print(f"🎯 Key themes: {', '.join(analysis.get('key_themes', []))}")
        
        # Step 2: Generate questions
        print("❓ Generating targeted questions...")
        questions = self.generate_contextual_questions(analysis, interview_type)
        print(f"✅ Generated {len(questions)} questions")
        
        # Step 3: Conduct interview
        print("🎙️ Starting interactive interview...")
        results = self.conduct_interview(questions)
        
        # Step 4: Save results
        summary_file = self.save_interview_results(results, output_dir)
        
        print(f"\n🎉 Interview session complete!")
        print(f"📄 Summary available at: {summary_file}")
        
        return summary_file


def main():
    """CLI entry point for testing the interview system"""
    import argparse
    
    parser = argparse.ArgumentParser(description="Interactive Interview System for Context Extraction")
    parser.add_argument("materials_path", help="Path to project materials to analyze")
    parser.add_argument("--type", choices=["postmortem", "deviation_analysis", "gap_analysis"], 
                       default="postmortem", help="Type of interview to conduct")
    parser.add_argument("--output-dir", help="Output directory for results")
    
    args = parser.parse_args()
    
    interviewer = InteractiveInterviewer()
    interviewer.conduct_full_interview_session(
        args.materials_path,
        args.type,
        args.output_dir
    )


if __name__ == "__main__":
    main() 