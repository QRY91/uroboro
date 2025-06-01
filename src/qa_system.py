"""
Q&A System for Presentation Preparation and Understanding Validation
Extends the interview system with audience-specific questioning modes
Perfect for jury preparation and knowledge validation
"""

import json
import random
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Any, Optional
from .interview_system import InteractiveInterviewer


class QASystemModes:
    """Different audience types for question generation"""
    
    EXPERT = "expert"
    LAYMAN = "layman" 
    GRANDMA = "grandma"
    SKEPTICAL_ACADEMIC = "skeptical_academic"
    BUSINESS_EXECUTIVE = "business_executive"
    TECHNICAL_PEER = "technical_peer"
    CONCERNED_CITIZEN = "concerned_citizen"
    INVESTOR = "investor"


class QuestionGenerator:
    """Generate audience-specific questions from project materials"""
    
    def __init__(self):
        self.audience_profiles = {
            QASystemModes.EXPERT: {
                "name": "Domain Expert",
                "description": "Deep technical knowledge, asks about implementation details, edge cases, and theoretical foundations",
                "question_style": "technical_deep_dive",
                "complexity": "high",
                "focus": ["technical_implementation", "architecture", "performance", "scalability", "best_practices"]
            },
            QASystemModes.LAYMAN: {
                "name": "Educated Layperson", 
                "description": "General education but not domain-specific, needs clear explanations and practical context",
                "question_style": "practical_understanding",
                "complexity": "medium",
                "focus": ["problem_solved", "user_benefits", "real_world_impact", "practical_applications"]
            },
            QASystemModes.GRANDMA: {
                "name": "Your Grandmother",
                "description": "Intelligent but non-technical, asks about purpose, value, and human impact",
                "question_style": "human_centered",
                "complexity": "low",
                "focus": ["why_it_matters", "who_benefits", "simple_explanation", "practical_value"]
            },
            QASystemModes.SKEPTICAL_ACADEMIC: {
                "name": "Skeptical Academic Reviewer",
                "description": "Questions methodology, validity, limitations, and academic rigor",
                "question_style": "critical_analysis",
                "complexity": "high", 
                "focus": ["methodology", "limitations", "validity", "reproducibility", "academic_contribution"]
            },
            QASystemModes.BUSINESS_EXECUTIVE: {
                "name": "Business Executive",
                "description": "Cares about ROI, scalability, market potential, and business impact",
                "question_style": "business_focused",
                "complexity": "medium",
                "focus": ["business_value", "cost_benefit", "scalability", "market_potential", "competitive_advantage"]
            },
            QASystemModes.TECHNICAL_PEER: {
                "name": "Technical Peer",
                "description": "Fellow developer, asks about code quality, tools, and technical decisions",
                "question_style": "peer_review",
                "complexity": "high",
                "focus": ["code_quality", "tool_choices", "technical_decisions", "debugging", "maintenance"]
            },
            QASystemModes.CONCERNED_CITIZEN: {
                "name": "Concerned Citizen",
                "description": "Worried about privacy, security, ethical implications, and societal impact",
                "question_style": "ethical_safety",
                "complexity": "medium",
                "focus": ["privacy", "security", "ethical_implications", "societal_impact", "safety"]
            },
            QASystemModes.INVESTOR: {
                "name": "Potential Investor",
                "description": "Focused on market opportunity, growth potential, and financial projections",
                "question_style": "investment_focused",
                "complexity": "medium",
                "focus": ["market_size", "growth_potential", "monetization", "competitive_landscape", "team_capability"]
            }
        }
    
    def generate_questions_for_audience(self, materials_analysis: Dict[str, Any], 
                                      audience_type: str, num_questions: int = 5) -> List[Dict[str, str]]:
        """Generate audience-specific questions based on materials analysis"""
        
        if audience_type not in self.audience_profiles:
            raise ValueError(f"Unknown audience type: {audience_type}")
        
        profile = self.audience_profiles[audience_type]
        questions = []
        
        # Generate questions based on audience focus areas
        for focus_area in profile["focus"]:
            if len(questions) >= num_questions:
                break
                
            question = self._generate_focus_question(materials_analysis, focus_area, profile)
            if question:
                questions.append(question)
        
        # Fill remaining slots with general questions
        while len(questions) < num_questions:
            general_question = self._generate_general_question(materials_analysis, profile)
            if general_question and general_question not in questions:
                questions.append(general_question)
            else:
                break  # Avoid infinite loop
        
        return questions[:num_questions]
    
    def _generate_focus_question(self, materials: Dict[str, Any], focus_area: str, 
                                profile: Dict[str, str]) -> Optional[Dict[str, str]]:
        """Generate a question for a specific focus area"""
        
        audience_name = profile["name"]
        complexity = profile["complexity"]
        
        # Question templates by focus area and complexity
        question_templates = {
            "technical_implementation": {
                "high": [
                    f"As a {audience_name}, I'd like to understand your specific implementation choices. What trade-offs did you consider when selecting your technology stack?",
                    f"From a technical perspective, how does your system handle edge cases and error conditions?",
                    f"What specific design patterns did you implement, and why were they appropriate for your use case?"
                ],
                "medium": [
                    f"How did you decide which technologies to use for this project?",
                    f"What technical challenges did you encounter and how did you solve them?"
                ],
                "low": [
                    f"Can you explain how the technical parts of your project work in simple terms?"
                ]
            },
            "problem_solved": {
                "high": [
                    f"What specific problem does your system solve that existing solutions don't address?",
                    f"How do you quantify the problem you're solving, and what evidence supports its significance?"
                ],
                "medium": [
                    f"What problem were you trying to solve, and why was it important?",
                    f"How does your solution improve upon existing approaches?"
                ],
                "low": [
                    f"What problem does your project solve for people?",
                    f"Why was this problem worth solving?"
                ]
            },
            "methodology": {
                "high": [
                    f"As a {audience_name}, I have concerns about your research methodology. How do you address potential bias in your approach?",
                    f"What are the limitations of your methodology, and how do they affect your conclusions?",
                    f"How would you replicate this study, and what would you do differently?"
                ],
                "medium": [
                    f"Can you explain your research approach and why you chose it?",
                    f"What could someone criticize about your methodology?"
                ]
            },
            "business_value": {
                "high": [
                    f"What's the total addressable market for this solution, and how did you calculate it?",
                    f"What are the key business metrics you'd track to measure success?"
                ],
                "medium": [
                    f"How would this project make money or create business value?",
                    f"Who would pay for this, and why?"
                ],
                "low": [
                    f"How does this help businesses or organizations?"
                ]
            },
            "why_it_matters": {
                "low": [
                    f"Can you explain why this project is important in a way I can understand?",
                    f"How does this make life better for people?",
                    f"What would happen if this problem wasn't solved?"
                ]
            },
            "privacy": {
                "medium": [
                    f"As a {audience_name}, I'm worried about privacy. What data does your system collect, and how is it protected?",
                    f"Who has access to the information in your system, and how do you prevent misuse?"
                ],
                "low": [
                    f"Is my personal information safe with your project?"
                ]
            }
        }
        
        if focus_area in question_templates and complexity in question_templates[focus_area]:
            templates = question_templates[focus_area][complexity]
            selected_template = random.choice(templates)
            
            return {
                "id": f"{focus_area}_{complexity}",
                "question": selected_template,
                "audience": audience_name,
                "focus_area": focus_area,
                "complexity": complexity,
                "purpose": f"Test understanding of {focus_area} for {audience_name}"
            }
        
        return None
    
    def _generate_general_question(self, materials: Dict[str, Any], 
                                 profile: Dict[str, str]) -> Optional[Dict[str, str]]:
        """Generate a general question for the audience"""
        
        audience_name = profile["name"]
        complexity = profile["complexity"]
        
        general_templates = {
            "high": [
                f"What would you say to critics who question the validity of your approach?",
                f"How would you defend your work to someone who disagrees with your conclusions?",
                f"What's the weakest part of your project, and how do you address it?"
            ],
            "medium": [
                f"What surprised you most during this project?",
                f"If you had to start over, what would you do differently?",
                f"What's the most important thing people should understand about your work?"
            ],
            "low": [
                f"Can you tell me about your project in simple terms?",
                f"What are you most proud of in this work?",
                f"How does this project help people?"
            ]
        }
        
        if complexity in general_templates:
            template = random.choice(general_templates[complexity])
            return {
                "id": f"general_{complexity}",
                "question": template,
                "audience": audience_name,
                "focus_area": "general",
                "complexity": complexity,
                "purpose": f"General understanding check for {audience_name}"
            }
        
        return None


class InteractiveQASession:
    """Conduct interactive Q&A sessions for presentation preparation"""
    
    def __init__(self):
        self.question_generator = QuestionGenerator()
        self.session_data = {
            "timestamp": datetime.now().isoformat(),
            "questions_asked": [],
            "responses": [],
            "audience_types": [],
            "performance_metrics": {}
        }
    
    def prepare_for_presentation(self, materials_path: str, audience_types: List[str], 
                               questions_per_audience: int = 3, 
                               mode: str = "practice") -> Dict[str, Any]:
        """Prepare for presentation with mixed audience Q&A"""
        
        print(f"🎭 PRESENTATION PREP: Q&A Session")
        print(f"📁 Materials: {materials_path}")
        print(f"👥 Audiences: {', '.join(audience_types)}")
        print(f"❓ Questions per audience: {questions_per_audience}")
        print("=" * 50)
        
        # Analyze materials (simplified for now)
        materials_analysis = self._analyze_materials(materials_path)
        
        # Generate questions for each audience
        all_questions = []
        for audience_type in audience_types:
            audience_questions = self.question_generator.generate_questions_for_audience(
                materials_analysis, audience_type, questions_per_audience
            )
            all_questions.extend(audience_questions)
        
        # Shuffle for mixed audience simulation
        random.shuffle(all_questions)
        
        print(f"✅ Generated {len(all_questions)} questions from {len(audience_types)} audience types")
        
        if mode == "practice":
            return self._conduct_practice_session(all_questions)
        elif mode == "quiz":
            return self._conduct_quiz_session(all_questions)
        else:
            return {"questions": all_questions, "mode": "preview"}
    
    def _analyze_materials(self, materials_path: str) -> Dict[str, Any]:
        """Quick analysis of materials for question generation"""
        
        materials_path = Path(materials_path)
        analysis = {
            "has_technical_content": False,
            "has_business_content": False,
            "has_research_content": False,
            "key_topics": [],
            "complexity_level": "medium"
        }
        
        # Simple keyword-based analysis
        if materials_path.exists():
            for file_path in materials_path.rglob("*.md"):
                try:
                    content = file_path.read_text(encoding='utf-8').lower()
                    
                    if any(word in content for word in ['api', 'database', 'architecture', 'implementation']):
                        analysis["has_technical_content"] = True
                    
                    if any(word in content for word in ['business', 'roi', 'value', 'market', 'customer']):
                        analysis["has_business_content"] = True
                    
                    if any(word in content for word in ['research', 'methodology', 'study', 'analysis']):
                        analysis["has_research_content"] = True
                        
                except Exception:
                    continue
        
        return analysis
    
    def _conduct_practice_session(self, questions: List[Dict[str, str]]) -> Dict[str, Any]:
        """Conduct practice session with feedback"""
        
        print("\n🎤 PRACTICE MODE")
        print("Answer questions as you would in the real presentation.")
        print("Type 'skip' to skip, 'next' for next question, 'done' to finish")
        print("=" * 50)
        
        responses = []
        
        for i, question in enumerate(questions, 1):
            print(f"\n📝 Question {i}/{len(questions)}")
            print(f"👤 Audience: {question['audience']}")
            print(f"🎯 Focus: {question['focus_area']}")
            print("─" * 40)
            print(f"❓ {question['question']}")
            
            print("\n💬 Your answer (press Enter twice when done):")
            
            answer_lines = []
            while True:
                try:
                    line = input()
                    if line.lower() in ['skip', 'next', 'done']:
                        if line.lower() == 'done':
                            break
                        elif line.lower() in ['skip', 'next']:
                            print("⏭️  Moving to next question...")
                            break
                    
                    if line == "" and answer_lines and answer_lines[-1] == "":
                        break
                    
                    answer_lines.append(line)
                        
                except KeyboardInterrupt:
                    print("\n\n🛑 Session interrupted")
                    return self._finalize_session(responses)
            
            if line.lower() == 'done':
                break
                
            if answer_lines and answer_lines[0].lower() not in ['skip', 'next']:
                answer_text = "\n".join(answer_lines).strip()
                if answer_text:
                    responses.append({
                        "question": question,
                        "answer": answer_text,
                        "timestamp": datetime.now().isoformat()
                    })
                    print("✅ Answer recorded!")
        
        return self._finalize_session(responses)
    
    def _conduct_quiz_session(self, questions: List[Dict[str, str]]) -> Dict[str, Any]:
        """Quick quiz mode for rapid practice"""
        
        print("\n⚡ QUIZ MODE - Quick Fire Questions")
        print("Give brief answers. Focus on clarity and confidence.")
        print("=" * 50)
        
        responses = []
        
        for i, question in enumerate(questions, 1):
            print(f"\n❓ [{question['audience']}] {question['question']}")
            
            try:
                answer = input("💬 Quick answer: ").strip()
                if answer.lower() == 'done':
                    break
                    
                if answer and answer.lower() not in ['skip', 'next']:
                    responses.append({
                        "question": question,
                        "answer": answer,
                        "timestamp": datetime.now().isoformat(),
                        "mode": "quiz"
                    })
                    print("✅")
                    
            except KeyboardInterrupt:
                print("\n\n🛑 Quiz interrupted")
                break
        
        return self._finalize_session(responses)
    
    def _finalize_session(self, responses: List[Dict[str, Any]]) -> Dict[str, Any]:
        """Finalize Q&A session and provide feedback"""
        
        self.session_data["responses"] = responses
        self.session_data["total_questions"] = len(responses)
        self.session_data["completed_at"] = datetime.now().isoformat()
        
        # Generate performance feedback
        feedback = self._generate_feedback(responses)
        self.session_data["feedback"] = feedback
        
        return {
            "session_data": self.session_data,
            "responses": responses,
            "feedback": feedback
        }
    
    def _generate_feedback(self, responses: List[Dict[str, Any]]) -> Dict[str, str]:
        """Generate feedback on Q&A performance"""
        
        if not responses:
            return {"overall": "No responses to analyze"}
        
        # Analyze response patterns
        audience_coverage = {}
        avg_length = 0
        
        for response in responses:
            audience = response["question"]["audience"]
            audience_coverage[audience] = audience_coverage.get(audience, 0) + 1
            avg_length += len(response["answer"])
        
        avg_length = avg_length / len(responses) if responses else 0
        
        feedback = {
            "total_responses": len(responses),
            "audience_coverage": audience_coverage,
            "average_answer_length": f"{avg_length:.0f} characters",
            "strong_areas": [],
            "improvement_areas": []
        }
        
        # Simple feedback logic
        if avg_length < 50:
            feedback["improvement_areas"].append("Consider providing more detailed explanations")
        elif avg_length > 500:
            feedback["improvement_areas"].append("Try to be more concise in your answers")
        else:
            feedback["strong_areas"].append("Good balance of detail and conciseness")
        
        if len(audience_coverage) >= 3:
            feedback["strong_areas"].append("Good coverage across different audience types")
        
        return feedback
    
    def save_session_results(self, results: Dict[str, Any], output_dir: str = None) -> str:
        """Save Q&A session results"""
        
        if not output_dir:
            output_dir = Path("output") / "qa-sessions"
        else:
            output_dir = Path(output_dir)
        
        output_dir.mkdir(parents=True, exist_ok=True)
        
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        
        # Save session data
        session_file = output_dir / f"qa_session_{timestamp}.json"
        with open(session_file, 'w', encoding='utf-8') as f:
            json.dump(results, f, indent=2, ensure_ascii=False)
        
        # Generate summary report
        report = self._generate_session_report(results)
        report_file = output_dir / f"qa_report_{timestamp}.md"
        with open(report_file, 'w', encoding='utf-8') as f:
            f.write(report)
        
        print(f"\n💾 Q&A Session saved:")
        print(f"  📄 Report: {report_file}")
        print(f"  🗃️ Session data: {session_file}")
        
        return str(report_file)
    
    def _generate_session_report(self, results: Dict[str, Any]) -> str:
        """Generate Q&A session report"""
        
        session_data = results.get("session_data", {})
        responses = results.get("responses", [])
        feedback = results.get("feedback", {})
        
        report = f"# Q&A Session Report - {datetime.now().strftime('%Y-%m-%d %H:%M')}\n\n"
        report += "## Session Overview\n\n"
        report += f"- **Total Questions Answered:** {len(responses)}\n"
        report += f"- **Session Duration:** {session_data.get('completed_at', 'Unknown')}\n"
        report += f"- **Average Answer Length:** {feedback.get('average_answer_length', 'Unknown')}\n\n"
        
        # Audience coverage
        if "audience_coverage" in feedback:
            report += "## Audience Coverage\n\n"
            for audience, count in feedback["audience_coverage"].items():
                report += f"- **{audience}:** {count} questions\n"
            report += "\n"
        
        # Performance feedback
        if feedback.get("strong_areas"):
            report += "## Strengths\n\n"
            for strength in feedback["strong_areas"]:
                report += f"✅ {strength}\n"
            report += "\n"
        
        if feedback.get("improvement_areas"):
            report += "## Areas for Improvement\n\n"
            for improvement in feedback["improvement_areas"]:
                report += f"🎯 {improvement}\n"
            report += "\n"
        
        # Sample responses
        if responses:
            report += "## Sample Responses\n\n"
            for i, response in enumerate(responses[:3], 1):  # Show first 3
                q = response["question"]
                report += f"### Question {i} ({q['audience']})\n\n"
                report += f"**Q:** {q['question']}\n\n"
                report += f"**A:** {response['answer'][:200]}{'...' if len(response['answer']) > 200 else ''}\n\n"
                report += "---\n\n"
        
        return report


def main():
    """CLI entry point for Q&A system"""
    import argparse
    
    parser = argparse.ArgumentParser(description="Interactive Q&A System for Presentation Preparation")
    parser.add_argument("materials_path", help="Path to project materials")
    parser.add_argument("--audiences", nargs='+', 
                       choices=list(QASystemModes.__dict__.values()),
                       default=["expert", "layman"],
                       help="Audience types for questions")
    parser.add_argument("--questions-per-audience", type=int, default=3,
                       help="Number of questions per audience type")
    parser.add_argument("--mode", choices=["practice", "quiz", "preview"], 
                       default="practice", help="Q&A session mode")
    parser.add_argument("--output-dir", help="Output directory for results")
    
    args = parser.parse_args()
    
    qa_session = InteractiveQASession()
    results = qa_session.prepare_for_presentation(
        args.materials_path,
        args.audiences,
        args.questions_per_audience,
        args.mode
    )
    
    if args.mode != "preview":
        qa_session.save_session_results(results, args.output_dir)


if __name__ == "__main__":
    main() 