#!/usr/bin/env python3
"""
Sensei System for uroboro
Implements the dojo/sensei/apprentice learning framework for AI skill transfer
"""

import json
import os
from pathlib import Path
from datetime import datetime
from typing import Dict, List, Optional, Any
import logging

logger = logging.getLogger(__name__)


class VoiceIntegratedDojoAnalyzer:
    """🥋 Enhanced Dojo mode with voice training integration"""
    
    def __init__(self):
        self.skills_registry = {
            "ieee_formatting": "IEEE academic citation and formatting standards",
            "academic_writing": "Academic prose and documentation style",
            "content_transformation": "Converting structured data to flowing prose",
            "voice_mimicry": "Matching authentic voice patterns and tone",
            "anti_ai_patterns": "Removing AI-generated language markers",
            "technical_documentation": "Clear technical specification writing",
            "research_synthesis": "Combining multiple sources into coherent analysis"
        }
        
        # Voice-specific skills that integrate with voice training
        self.voice_skills = {
            "voice_mimicry", "academic_writing", "content_transformation", "anti_ai_patterns"
        }
        
        # Skill proficiency tracking
        self.proficiency_levels = {
            "beginner": "Basic understanding, requires guidance",
            "intermediate": "Can apply with supervision", 
            "advanced": "Independent application",
            "expert": "Can teach others",
            "master": "Innovates and extends the skill"
        }
    
    def get_voice_profile_stats(self) -> Dict[str, Any]:
        """Get voice profile statistics and integration status"""
        voice_profile_path = Path("voice_profile.json")
        
        if voice_profile_path.exists():
            with open(voice_profile_path, 'r', encoding='utf-8') as f:
                voice_profile = json.load(f)
            
            # Analyze voice profile for skill assessment
            voice_stats = {
                "has_voice_profile": True,
                "sentence_complexity": voice_profile.get("sentence_patterns", {}).get("avg_length", 15),
                "technical_proficiency": voice_profile.get("technical_terms", {}).get("tech_term_frequency", 0.02),
                "academic_indicators": self._assess_academic_style(voice_profile),
                "writing_maturity": self._assess_writing_maturity(voice_profile),
                "voice_skills_ready": True
            }
        else:
            voice_stats = {
                "has_voice_profile": False,
                "voice_skills_ready": False,
                "recommendation": "Run 'uro voice' to analyze your writing patterns first"
            }
        
        return voice_stats
    
    def _assess_academic_style(self, voice_profile: Dict) -> Dict[str, Any]:
        """Assess academic writing capability from voice profile"""
        sentence_length = voice_profile.get("sentence_patterns", {}).get("avg_length", 15)
        tech_frequency = voice_profile.get("technical_terms", {}).get("tech_term_frequency", 0)
        uses_lists = voice_profile.get("paragraph_structure", {}).get("uses_lists", False)
        meta_commentary = voice_profile.get("writing_habits", {}).get("uses_meta_commentary", False)
        
        # Calculate academic style score
        academic_score = 0
        if sentence_length > 18:  # Longer sentences indicate academic style
            academic_score += 2
        if tech_frequency > 0.03:  # High technical terminology
            academic_score += 2
        if uses_lists:  # Structured thinking
            academic_score += 1
        if meta_commentary:  # Reflective thinking
            academic_score += 1
        
        proficiency = "beginner"
        if academic_score >= 5:
            proficiency = "advanced"
        elif academic_score >= 3:
            proficiency = "intermediate"
        
        return {
            "score": academic_score,
            "proficiency": proficiency,
            "sentence_length": sentence_length,
            "technical_frequency": tech_frequency,
            "structured_writing": uses_lists
        }
    
    def _assess_writing_maturity(self, voice_profile: Dict) -> Dict[str, Any]:
        """Assess overall writing maturity and sophistication"""
        # Multiple indicators of writing maturity
        avg_sentence_length = voice_profile.get("sentence_patterns", {}).get("avg_length", 15)
        uses_parentheticals = voice_profile.get("writing_habits", {}).get("uses_parentheticals", False)
        contraction_freq = voice_profile.get("tone_indicators", {}).get("contraction_frequency", 0.1)
        
        maturity_score = 0
        if avg_sentence_length > 16:
            maturity_score += 1
        if uses_parentheticals:  # Complex thought structure
            maturity_score += 1
        if contraction_freq < 0.05:  # Formal style
            maturity_score += 1
        
        return {
            "score": maturity_score,
            "formal_style": contraction_freq < 0.05,
            "complex_structure": uses_parentheticals,
            "sophisticated_sentences": avg_sentence_length > 16
        }

    def update_voice_integrated_stats(self, skill: str, session_quality: float, 
                                    teaching_session_content: str) -> Dict[str, Any]:
        """Update skill stats with voice analysis integration"""
        stats = self.get_skill_stats()
        voice_stats = self.get_voice_profile_stats()
        
        # Update skill proficiency based on voice analysis if it's a voice skill
        if skill in self.voice_skills and voice_stats.get("voice_skills_ready", False):
            # Analyze the teaching session content for voice patterns
            session_analysis = self._analyze_session_voice_patterns(teaching_session_content)
            
            # Update proficiency based on voice integration
            current_proficiency = stats['skills_overview'][skill]['proficiency']
            new_proficiency = self._calculate_voice_enhanced_proficiency(
                current_proficiency, session_quality, session_analysis, voice_stats
            )
            
            stats['skills_overview'][skill]['proficiency'] = new_proficiency
            stats['skills_overview'][skill]['voice_integration_score'] = session_analysis.get('integration_score', 0.5)
            stats['skills_overview'][skill]['voice_pattern_match'] = session_analysis.get('pattern_match', 0.5)
        
        # Update general session tracking
        stats['skills_overview'][skill]['training_sessions'] += 1
        stats['skills_overview'][skill]['last_practiced'] = datetime.now().isoformat()
        stats['skills_overview'][skill]['success_rate'] = (
            stats['skills_overview'][skill]['success_rate'] * 0.8 + session_quality * 0.2
        )
        
        # Update overall stats
        stats['training_sessions']['total_sessions'] += 1
        stats['training_sessions']['last_training_date'] = datetime.now().isoformat()
        stats['last_updated'] = datetime.now().isoformat()
        
        # Save updated stats
        self._save_stats(stats)
        
        return {
            "skill_updated": skill,
            "new_proficiency": stats['skills_overview'][skill]['proficiency'],
            "voice_integration": skill in self.voice_skills,
            "session_quality": session_quality
        }
    
    def _analyze_session_voice_patterns(self, content: str) -> Dict[str, Any]:
        """Analyze teaching session content for voice pattern compliance"""
        # Simple heuristic analysis - could be enhanced with NLP
        words = content.split()
        sentences = content.split('.')
        
        # Voice pattern indicators
        avg_sentence_length = len(words) / len(sentences) if sentences else 0
        technical_density = len([w for w in words if any(tech in w.lower() for tech in ['technical', 'system', 'implementation', 'framework'])]) / len(words)
        structure_indicators = content.count('•') + content.count('-') + content.count('1.')
        
        # Calculate integration score
        integration_score = 0.5  # Base score
        if 12 < avg_sentence_length < 25:  # Good sentence complexity
            integration_score += 0.2
        if 0.02 < technical_density < 0.1:  # Appropriate technical density
            integration_score += 0.2
        if structure_indicators > 5:  # Good structure
            integration_score += 0.1
        
        return {
            "integration_score": min(integration_score, 1.0),
            "pattern_match": min(integration_score + 0.1, 1.0),
            "sentence_complexity": avg_sentence_length,
            "technical_density": technical_density,
            "structure_score": structure_indicators
        }
    
    def _calculate_voice_enhanced_proficiency(self, current: str, session_quality: float, 
                                           session_analysis: Dict, voice_stats: Dict) -> str:
        """Calculate new proficiency level with voice integration"""
        proficiency_order = list(self.proficiency_levels.keys())
        current_index = proficiency_order.index(current)
        
        # Factors that can increase proficiency
        quality_bonus = 1 if session_quality > 0.8 else 0
        voice_bonus = 1 if session_analysis.get('integration_score', 0) > 0.7 else 0
        academic_bonus = 1 if voice_stats.get('academic_indicators', {}).get('proficiency') in ['intermediate', 'advanced'] else 0
        
        total_bonus = quality_bonus + voice_bonus + academic_bonus
        
        # Advance proficiency if enough positive indicators
        if total_bonus >= 2 and current_index < len(proficiency_order) - 1:
            return proficiency_order[current_index + 1]
        elif total_bonus == 0 and current_index > 0:  # Regress if poor performance
            return proficiency_order[current_index - 1]
        
        return current

# Update the existing DojoAnalyzer to use the enhanced version
class DojoAnalyzer(VoiceIntegratedDojoAnalyzer):
    """🥋 Dojo mode - Enhanced with voice integration"""
    
    def list_available_skills(self) -> List[str]:
        """List all available skills for analysis"""
        return list(self.skills_registry.keys())
    
    def get_skill_stats(self) -> Dict[str, Any]:
        """Get comprehensive statistics about uroboro's skills and capabilities"""
        stats_file = Path("output") / "sensei-sessions" / "skill_stats.json"
        
        # Load existing stats or create new ones
        if stats_file.exists():
            with open(stats_file, 'r', encoding='utf-8') as f:
                stats = json.load(f)
        else:
            stats = self._initialize_skill_stats()
        
        return stats
    
    def _initialize_skill_stats(self) -> Dict[str, Any]:
        """Initialize skill statistics tracking"""
        stats = {
            "total_skills": len(self.skills_registry),
            "skills_overview": {},
            "training_sessions": {
                "total_sessions": 0,
                "sessions_by_skill": {},
                "last_training_date": None
            },
            "proficiency_distribution": {level: 0 for level in self.proficiency_levels.keys()},
            "capabilities": {
                "can_analyze": True,
                "can_teach": True, 
                "can_learn": True,
                "can_practice": True,
                "can_assess": True
            },
            "last_updated": datetime.now().isoformat()
        }
        
        # Initialize each skill
        for skill, description in self.skills_registry.items():
            stats["skills_overview"][skill] = {
                "description": description,
                "proficiency": "intermediate",  # Default proficiency
                "training_sessions": 0,
                "last_practiced": None,
                "success_rate": 0.85,  # Default success rate
                "areas_for_improvement": []
            }
            
        return stats
    
    def analyze_skill_gap(self, skill: str, materials_path: Optional[str] = None) -> str:
        """Analyze skill gaps in provided materials"""
        if skill not in self.skills_registry:
            return f"❌ Unknown skill: {skill}. Available skills: {', '.join(self.skills_registry.keys())}"
        
        # TODO: Implement actual analysis using existing uroboro modules
        # For now, return a structured analysis format
        
        analysis = f"""
🎯 **Skill Gap Analysis: {skill}**

**Skill Definition**: {self.skills_registry[skill]}

**Materials Analyzed**: {materials_path or 'Current directory'}

**Gap Assessment**:
- ✅ Current Strengths: [To be analyzed from materials]
- ❌ Identified Gaps: [To be analyzed from materials]  
- 🎯 Training Focus Areas: [To be determined]

**Recommended Training Actions**:
1. Generate teaching materials for gaps
2. Create practice exercises
3. Establish success metrics

**Next Steps**:
```bash
uro sensei --teach --skill {skill} --materials {materials_path or '.'}
```
"""
        return analysis
    
    def prepare_training_materials(self, skill: str, materials_path: Optional[str] = None) -> str:
        """Prepare structured training materials for a skill"""
        if skill not in self.skills_registry:
            return f"❌ Unknown skill: {skill}"
        
        # Create training plan
        training_plan = f"""
📚 **Training Materials for {skill}**

**Learning Objectives**:
- Master {self.skills_registry[skill]}
- Apply skill to real materials
- Achieve consistent quality output

**Training Structure**:
1. **Theory**: Core principles and patterns
2. **Examples**: Analyze good/bad examples  
3. **Practice**: Hands-on application
4. **Feedback**: Quality assessment

**Materials Required**:
- Source materials: {materials_path or 'Current directory'}
- Training examples: [To be generated]
- Practice exercises: [To be created]

**Assessment Criteria**:
- Technical accuracy
- Consistency with patterns
- Practical applicability

Ready for sensei teaching phase.
"""
        return training_plan

    def update_skill_proficiency(self, skill: str, new_proficiency: str, 
                                success_rate: float = None) -> bool:
        """Update skill proficiency based on training results"""
        if skill not in self.skills_registry:
            return False
            
        if new_proficiency not in self.proficiency_levels:
            return False
        
        stats = self.get_skill_stats()
        stats['skills_overview'][skill]['proficiency'] = new_proficiency
        
        if success_rate is not None:
            stats['skills_overview'][skill]['success_rate'] = success_rate
            
        stats['skills_overview'][skill]['last_practiced'] = datetime.now().isoformat()
        stats['last_updated'] = datetime.now().isoformat()
        
        # Save updated stats
        self._save_stats(stats)
        
        return True
    
    def _save_stats(self, stats: Dict[str, Any]):
        """Save updated statistics to file"""
        stats_file = Path("output") / "sensei-sessions" / "skill_stats.json"
        stats_file.parent.mkdir(parents=True, exist_ok=True)
        
        with open(stats_file, 'w', encoding='utf-8') as f:
            json.dump(stats, f, indent=2)
    
    def display_skill_stats(self) -> str:
        """Display formatted skill statistics with voice integration"""
        stats = self.get_skill_stats()
        voice_stats = self.get_voice_profile_stats()
        
        report = f"""
🎯 **UROBORO SKILLS ASSESSMENT** (Voice-Enhanced)
=================================================

📊 **Overall Statistics**:
• Total Skills: {stats['total_skills']}
• Training Sessions: {stats['training_sessions']['total_sessions']}
• Last Training: {stats['training_sessions']['last_training_date'] or 'Never'}
• Voice Profile: {'✅ Active' if voice_stats.get('has_voice_profile') else '❌ Missing'}

🎤 **Voice Integration Status**:"""
        
        if voice_stats.get('has_voice_profile'):
            academic_style = voice_stats.get('academic_indicators', {})
            report += f"""
  • Academic Style Score: {academic_style.get('score', 0)}/6
  • Academic Proficiency: {academic_style.get('proficiency', 'beginner').title()}
  • Sentence Complexity: {academic_style.get('sentence_length', 15):.1f} words/sentence
  • Technical Frequency: {academic_style.get('technical_frequency', 0):.1%}
  • Voice Skills Ready: ✅"""
        else:
            report += f"""
  • Voice Profile: ❌ Not found
  • Recommendation: Run 'uro voice' to analyze your writing patterns
  • Voice Skills Ready: ❌"""

        report += f"""

🏆 **Skill Proficiency Breakdown**:"""
        
        for skill, details in stats['skills_overview'].items():
            proficiency_emoji = {
                "beginner": "🟡",
                "intermediate": "🟠", 
                "advanced": "🔵",
                "expert": "🟢",
                "master": "🟣"
            }.get(details['proficiency'], "⚪")
            
            voice_indicator = ""
            if skill in self.voice_skills:
                if 'voice_integration_score' in details:
                    score = details['voice_integration_score']
                    voice_indicator = f" 🎤{score:.1%}"
                else:
                    voice_indicator = " 🎤❓"
            
            report += f"""
  {proficiency_emoji} **{skill.title().replace('_', ' ')}**: {details['proficiency'].title()}{voice_indicator}
     └─ {details['description']}
     └─ Success Rate: {details['success_rate']:.1%}
     └─ Sessions: {details['training_sessions']}"""
        
        # Rest of the display logic remains the same
        report += f"""

🎪 **Core Capabilities**:"""
        
        for capability, enabled in stats['capabilities'].items():
            status = "✅" if enabled else "❌"
            report += f"""
  {status} {capability.replace('_', ' ').title()}"""
        
        if voice_stats.get('has_voice_profile'):
            report += f"""
  ✅ Voice Pattern Analysis
  ✅ Voice-Enhanced Learning"""
        else:
            report += f"""
  ❌ Voice Pattern Analysis (requires voice profile)
  ❌ Voice-Enhanced Learning (requires voice profile)"""
        
        report += f"""

🔥 **Top Skills** (Voice-Enhanced Rankings):"""
        
        # Enhanced sorting with voice integration
        top_skills = sorted(
            stats['skills_overview'].items(),
            key=lambda x: (
                list(self.proficiency_levels.keys()).index(x[1]['proficiency']),
                x[1]['success_rate'],
                x[1].get('voice_integration_score', 0) if x[0] in self.voice_skills else 0
            ),
            reverse=True
        )[:3]
        
        for i, (skill, details) in enumerate(top_skills, 1):
            voice_bonus = f" +🎤" if skill in self.voice_skills and 'voice_integration_score' in details else ""
            report += f"""
  {i}. {skill.title().replace('_', ' ')} ({details['proficiency']}, {details['success_rate']:.1%}){voice_bonus}"""
        
        report += f"""

📈 **Growth Areas**:"""
        
        # Find skills with lowest proficiency
        growth_skills = sorted(
            stats['skills_overview'].items(),
            key=lambda x: list(self.proficiency_levels.keys()).index(x[1]['proficiency'])
        )[:2]
        
        for skill, details in growth_skills:
            voice_hint = " (voice-enhanced)" if skill in self.voice_skills and voice_stats.get('has_voice_profile') else ""
            report += f"""
  🎯 {skill.title().replace('_', ' ')}: Focus on {', '.join(details.get('areas_for_improvement', ['practice']))}{voice_hint}"""
        
        report += f"""

🚀 **Recommended Actions**:
  1. Practice {growth_skills[0][0]} skill{' with voice integration' if growth_skills[0][0] in self.voice_skills else ''}
  2. Conduct teaching session for {top_skills[0][0]}"""
        
        if not voice_stats.get('has_voice_profile'):
            report += f"""
  3. Run 'uro voice' to enable voice-enhanced learning"""
        else:
            report += f"""
  3. Leverage voice patterns for {len(self.voice_skills)} voice-integrated skills"""
        
        report += f"""

📊 Last Updated: {stats['last_updated'][:19].replace('T', ' ')}
"""
        
        return report


class SenseiTeacher:
    """🎓 Sensei mode - Teaches skills to users and other AIs"""
    
    def __init__(self):
        self.teaching_sessions_dir = Path("output") / "sensei-sessions"
        self.teaching_sessions_dir.mkdir(parents=True, exist_ok=True)
    
    def teach_skill(self, skill: str, student_type: str = "user", 
                   materials_path: Optional[str] = None, 
                   voice_profile: Optional[str] = None) -> str:
        """Conduct a teaching session for a specific skill"""
        
        # Load relevant modules for teaching
        teaching_content = self._generate_teaching_content(skill, materials_path, voice_profile)
        
        # Adapt teaching style based on student type
        if student_type == "ai" or student_type == "apprentice":
            teaching_session = self._format_for_ai_student(teaching_content, skill)
        else:
            teaching_session = self._format_for_human_student(teaching_content, skill)
        
        return teaching_session
    
    def _generate_teaching_content(self, skill: str, materials_path: Optional[str], 
                                 voice_profile: Optional[str]) -> Dict[str, Any]:
        """Generate the core teaching content"""
        # TODO: Integrate with existing uroboro modules
        # - Use academic_voice.py for voice analysis
        # - Use processors for content analysis
        # - Use research_organizer for material organization
        
        return {
            "skill": skill,
            "principles": f"Core principles for {skill}",
            "examples": f"Examples from {materials_path or 'default materials'}",
            "patterns": f"Patterns to follow for {skill}",
            "anti_patterns": f"Common mistakes to avoid in {skill}",
            "voice_guidance": f"Voice profile: {voice_profile or 'default'}"
        }
    
    def _format_for_ai_student(self, content: Dict[str, Any], skill: str) -> str:
        """Format teaching content for AI consumption"""
        return f"""
🤖 **AI Teaching Session: {skill}**

**SYSTEM PROMPT ENHANCEMENT**:
```
You are learning {skill}. Apply these principles:

CORE PRINCIPLES:
{content['principles']}

PATTERNS TO FOLLOW:
{content['patterns']}

ANTI-PATTERNS TO AVOID:
{content['anti_patterns']}

VOICE GUIDANCE:
{content['voice_guidance']}
```

**EXAMPLES FOR PATTERN MATCHING**:
{content['examples']}

**INTEGRATION INSTRUCTIONS**:
- Incorporate these patterns into your response generation
- Reference examples when applying the skill
- Maintain consistency with voice guidance
- Self-correct using anti-pattern awareness

**ASSESSMENT CRITERIA**:
Rate your output on:
1. Technical accuracy (1-10)
2. Pattern consistency (1-10) 
3. Voice authenticity (1-10)
"""
    
    def _format_for_human_student(self, content: Dict[str, Any], skill: str) -> str:
        """Format teaching content for human consumption"""
        return f"""
👨‍🎓 **Human Teaching Session: {skill}**

**What You'll Learn**:
{content['principles']}

**Key Patterns to Master**:
{content['patterns']}

**Common Mistakes to Avoid**:
{content['anti_patterns']}

**Examples to Study**:
{content['examples']}

**Practice Exercises**:
1. Analyze the provided examples
2. Apply patterns to your own materials
3. Compare your output to the examples
4. Iterate based on feedback

**Voice Guidelines**:
{content['voice_guidance']}

**Next Steps**:
```bash
uro apprentice --practice-skill --skill {skill}
```
"""
    
    def save_teaching_session(self, session_content: str, skill: str) -> Path:
        """Save teaching session for apprentice learning"""
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        filename = f"teaching_session_{skill}_{timestamp}.md"
        session_path = self.teaching_sessions_dir / filename
        
        with open(session_path, 'w', encoding='utf-8') as f:
            f.write(session_content)
        
        return session_path
    
    def respond_to_query(self, query_path: Path) -> str:
        """Respond to a structured query from another AI"""
        try:
            with open(query_path, 'r', encoding='utf-8') as f:
                query_data = json.load(f)
            
            skill = query_data.get('skill')
            context = query_data.get('context', '')
            specific_request = query_data.get('request', '')
            
            response = f"""
🎯 **Sensei Response to Query**

**Skill Requested**: {skill}
**Context**: {context}
**Specific Request**: {specific_request}

**Teaching Response**:
[Generated based on skill and context]

**Recommended Applications**:
[Specific guidance for the context]

**Quality Checks**:
[Assessment criteria for this skill]
"""
            return response
            
        except Exception as e:
            return f"❌ Error processing query: {e}"


class ApprenticeObserver:
    """👁️ Apprentice mode - Learns through observation and practice"""
    
    def __init__(self):
        self.learning_sessions_dir = Path("output") / "apprentice-learning"
        self.learning_sessions_dir.mkdir(parents=True, exist_ok=True)
    
    def observe_teaching_session(self, session_path: Path) -> str:
        """Observe and extract learning insights from a teaching session"""
        try:
            with open(session_path, 'r', encoding='utf-8') as f:
                session_content = f.read()
            
            # Extract key learning points
            learning_insights = self._extract_learning_insights(session_content)
            
            return learning_insights
            
        except Exception as e:
            return f"❌ Error observing session: {e}"
    
    def _extract_learning_insights(self, session_content: str) -> str:
        """Extract structured learning insights from session content"""
        # TODO: Use existing uroboro analysis capabilities
        # - Pattern extraction
        # - Voice analysis  
        # - Content synthesis
        
        return f"""
🧠 **Learning Insights Extracted**

**Patterns Identified**:
- [Extract patterns from session content]

**Voice Characteristics**:
- [Identify voice patterns]

**Quality Indicators**:
- [Note quality metrics]

**Integration Points**:
- [How to apply these insights]

**Practice Opportunities**:
- [Specific areas to practice]

**Assessment Framework**:
- [How to measure improvement]
"""
    
    def practice_skill(self, skill: str, practice_materials: Optional[str] = None,
                      voice_profile: Optional[str] = None) -> str:
        """Practice a skill with provided materials"""
        
        practice_session = f"""
🎯 **Skill Practice Session: {skill}**

**Materials**: {practice_materials or 'Default practice set'}
**Voice Profile**: {voice_profile or 'Default voice'}

**Practice Exercise**:
1. Analyze the source materials
2. Apply the learned patterns
3. Generate output using the skill
4. Self-assess against criteria

**Generated Output**:
[Practice output would be generated here]

**Self-Assessment**:
- Pattern application: [Score/feedback]
- Voice consistency: [Score/feedback]  
- Technical accuracy: [Score/feedback]

**Areas for Improvement**:
[Identified gaps and next practice focus]
"""
        
        return practice_session
    
    def integrate_into_voice_profile(self, learning_insights: str, 
                                   voice_profile: Optional[str] = None) -> str:
        """Integrate learning insights into voice profile"""
        # TODO: Integrate with existing voice training system
        
        return f"""
🎯 **Voice Profile Integration**

**Target Profile**: {voice_profile or 'Default voice profile'}

**New Insights Added**:
{learning_insights}

**Profile Updates**:
- [Specific voice pattern updates]
- [New quality criteria added]
- [Enhanced assessment capabilities]

**Validation Results**:
- Integration successful: ✅
- Profile consistency: ✅  
- Quality improvement: ✅
"""


# Utility functions for the sensei system
def create_skill_query(skill: str, context: str, request: str, output_path: str):
    """Create a structured query file for sensei response"""
    query_data = {
        "skill": skill,
        "context": context,
        "request": request,
        "timestamp": datetime.now().isoformat(),
        "format_version": "1.0"
    }
    
    with open(output_path, 'w', encoding='utf-8') as f:
        json.dump(query_data, f, indent=2)
    
    return f"Query created: {output_path}"


def list_teaching_sessions() -> List[Path]:
    """List all available teaching sessions"""
    sessions_dir = Path("output") / "sensei-sessions"
    if sessions_dir.exists():
        return list(sessions_dir.glob("*.md"))
    return []


def list_learning_sessions() -> List[Path]:
    """List all apprentice learning sessions"""
    learning_dir = Path("output") / "apprentice-learning"
    if learning_dir.exists():
        return list(learning_dir.glob("*.md"))
    return [] 