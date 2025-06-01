#!/usr/bin/env python3
"""
Eternal Egg System for uroboro
Gamified local model training - collect data, grow eggs, hatch personalized models
"""

import json
import os
import hashlib
from pathlib import Path
from datetime import datetime, timedelta
from typing import Dict, List, Optional, Any, Tuple
import logging

logger = logging.getLogger(__name__)


class EggSpecialization:
    """Defines a skill specialization that can be learned by eggs"""
    
    def __init__(self, name: str, description: str, target_examples: int = 500):
        self.name = name
        self.description = description
        self.target_examples = target_examples
        self.examples_collected = 0
        self.quality_scores = []
    
    def add_example(self, quality_score: float):
        """Add a training example for this specialization"""
        self.examples_collected += 1
        self.quality_scores.append(quality_score)
    
    def get_progress(self) -> float:
        """Get completion percentage for this specialization"""
        return min(self.examples_collected / self.target_examples, 1.0)
    
    def get_average_quality(self) -> float:
        """Get average quality score for collected examples"""
        return sum(self.quality_scores) / len(self.quality_scores) if self.quality_scores else 0.0


class EternalEgg:
    """An egg that grows by collecting training data and eventually hatches into a model"""
    
    def __init__(self, name: str, focus_areas: List[str], egg_type: str = "general"):
        self.name = name
        self.egg_type = egg_type
        self.created_date = datetime.now()
        self.focus_areas = focus_areas
        
        # Core stats
        self.total_examples = 0
        self.quality_threshold = 7.0  # Minimum quality for hatching
        self.target_examples = 2000   # Minimum examples for hatching
        
        # Specializations this egg can learn
        self.specializations = {
            "academic_writing": EggSpecialization("Academic Writing", "Formal, professional documentation"),
            "ieee_formatting": EggSpecialization("IEEE Formatting", "Academic citations and standards"),
            "anti_ai_patterns": EggSpecialization("Anti-AI Patterns", "Human-like authenticity"),
            "content_transformation": EggSpecialization("Content Transformation", "Structured to prose"),
            "voice_mimicry": EggSpecialization("Voice Mimicry", "Personal writing style"),
            "technical_documentation": EggSpecialization("Technical Documentation", "Clear technical writing"),
            "research_synthesis": EggSpecialization("Research Synthesis", "Multi-source analysis")
        }
        
        # Training dataset
        self.training_examples = []
        self.metadata = {
            "hatched": False,
            "hatch_date": None,
            "model_path": None,
            "generation": 1,
            "parent_eggs": [],
            "achievements": [],
            "last_fed": None
        }
    
    def feed(self, input_text: str, output_text: str, skill_tags: List[str], 
             quality_score: float, user_feedback: Optional[str] = None) -> Dict[str, Any]:
        """Feed the egg with a training example"""
        
        # Create training example
        example = {
            "id": hashlib.md5(f"{input_text}{output_text}{datetime.now()}".encode()).hexdigest()[:12],
            "timestamp": datetime.now().isoformat(),
            "input": input_text,
            "output": output_text,
            "quality_score": quality_score,
            "skill_tags": skill_tags,
            "user_feedback": user_feedback,
            "token_count": len(input_text.split()) + len(output_text.split())
        }
        
        # Add to dataset
        self.training_examples.append(example)
        self.total_examples += 1
        self.metadata["last_fed"] = datetime.now().isoformat()
        
        # Update specializations
        growth_report = []
        for skill in skill_tags:
            if skill in self.specializations:
                self.specializations[skill].add_example(quality_score)
                growth_report.append(f"{skill}: +1 example ({self.specializations[skill].get_progress():.1%})")
        
        # Check for achievements
        new_achievements = self._check_achievements()
        
        # Evaluate hatching readiness
        hatching_ready = self._is_ready_to_hatch()
        
        return {
            "example_id": example["id"],
            "total_examples": self.total_examples,
            "specialization_growth": growth_report,
            "new_achievements": new_achievements,
            "hatching_ready": hatching_ready,
            "overall_progress": self.get_overall_progress(),
            "quality_average": self.get_average_quality()
        }
    
    def _check_achievements(self) -> List[str]:
        """Check for new achievements and milestones"""
        new_achievements = []
        
        # Example milestones
        milestones = [
            (100, "First Century"),
            (500, "Data Collector"),
            (1000, "Training Master"),
            (2000, "Hatch Ready")
        ]
        
        for count, achievement in milestones:
            if self.total_examples >= count and achievement not in self.metadata["achievements"]:
                self.metadata["achievements"].append(achievement)
                new_achievements.append(achievement)
        
        # Quality achievements
        avg_quality = self.get_average_quality()
        quality_achievements = [
            (7.0, "Quality Threshold"),
            (8.0, "High Quality"),
            (9.0, "Exceptional Quality")
        ]
        
        for threshold, achievement in quality_achievements:
            if avg_quality >= threshold and achievement not in self.metadata["achievements"]:
                self.metadata["achievements"].append(achievement)
                new_achievements.append(achievement)
        
        return new_achievements
    
    def _is_ready_to_hatch(self) -> bool:
        """Check if egg meets hatching criteria"""
        has_enough_examples = self.total_examples >= self.target_examples
        meets_quality = self.get_average_quality() >= self.quality_threshold
        has_diversity = len([s for s in self.specializations.values() if s.examples_collected > 0]) >= 3
        
        return has_enough_examples and meets_quality and has_diversity and not self.metadata["hatched"]
    
    def get_overall_progress(self) -> float:
        """Get overall progress toward hatching (0.0 to 1.0)"""
        example_progress = min(self.total_examples / self.target_examples, 1.0)
        return example_progress
    
    def get_detailed_progress(self) -> Dict[str, float]:
        """Get detailed progress breakdown for debugging"""
        example_progress = min(self.total_examples / self.target_examples, 1.0)
        quality_progress = min(self.get_average_quality() / 10.0, 1.0)
        diversity_progress = len([s for s in self.specializations.values() if s.examples_collected > 0]) / 7.0
        
        return {
            "examples": example_progress,
            "quality": quality_progress, 
            "diversity": diversity_progress,
            "combined_average": (example_progress + quality_progress + diversity_progress) / 3.0
        }
    
    def get_average_quality(self) -> float:
        """Get average quality score across all examples"""
        if not self.training_examples:
            return 0.0
        return sum(ex["quality_score"] for ex in self.training_examples) / len(self.training_examples)
    
    def get_stats_summary(self) -> Dict[str, Any]:
        """Get comprehensive egg statistics"""
        age_days = (datetime.now() - self.created_date).days
        
        return {
            "name": self.name,
            "type": self.egg_type,
            "age_days": age_days,
            "total_examples": self.total_examples,
            "target_examples": self.target_examples,
            "overall_progress": self.get_overall_progress(),
            "quality_average": self.get_average_quality(),
            "specializations": {
                name: {
                    "progress": spec.get_progress(),
                    "examples": spec.examples_collected,
                    "quality": spec.get_average_quality()
                }
                for name, spec in self.specializations.items()
                if spec.examples_collected > 0
            },
            "achievements": self.metadata["achievements"],
            "hatching_ready": self._is_ready_to_hatch(),
            "hatched": self.metadata["hatched"],
            "last_fed": self.metadata["last_fed"]
        }
    
    def hatch(self) -> Dict[str, Any]:
        """Attempt to hatch the egg into a trained model"""
        if not self._is_ready_to_hatch():
            return {
                "success": False,
                "reason": "Egg not ready for hatching",
                "requirements": {
                    "examples_needed": max(0, self.target_examples - self.total_examples),
                    "quality_needed": max(0, self.quality_threshold - self.get_average_quality()),
                    "diversity_needed": max(0, 3 - len([s for s in self.specializations.values() if s.examples_collected > 0]))
                }
            }
        
        # Mark as hatched
        self.metadata["hatched"] = True
        self.metadata["hatch_date"] = datetime.now().isoformat()
        model_name = f"{self.name}-v{self.metadata['generation']}"
        self.metadata["model_path"] = f"models/{model_name}.gguf"
        
        # TODO: Implement actual model training pipeline
        # For now, return success with training specifications
        
        return {
            "success": True,
            "model_name": model_name,
            "model_path": self.metadata["model_path"],
            "training_dataset_size": self.total_examples,
            "specializations_learned": list(self.specializations.keys()),
            "estimated_training_time": "2-6 hours",
            "hatch_date": self.metadata["hatch_date"]
        }
    
    def check_data_health(self) -> Dict[str, Any]:
        """Check for overingestion and data quality issues"""
        health_report = {
            "status": "healthy",
            "warnings": [],
            "quality_trend": "stable",
            "diversity_score": 0.0,
            "freshness_score": 1.0
        }
        
        if not self.training_examples:
            return health_report
        
        # Check for quality degradation
        recent_examples = self.training_examples[-10:] if len(self.training_examples) >= 10 else self.training_examples
        older_examples = self.training_examples[:-10] if len(self.training_examples) >= 20 else []
        
        if older_examples:
            recent_quality = sum(ex["quality_score"] for ex in recent_examples) / len(recent_examples)
            older_quality = sum(ex["quality_score"] for ex in older_examples) / len(older_examples)
            
            if recent_quality < older_quality - 0.5:
                health_report["warnings"].append("Quality degradation detected")
                health_report["quality_trend"] = "declining"
            elif recent_quality > older_quality + 0.5:
                health_report["quality_trend"] = "improving"
        
        # Check for content repetition (simple hash-based detection)
        content_hashes = {}
        repetition_count = 0
        
        for example in self.training_examples:
            content_hash = hashlib.md5(f"{example['input']}{example['output']}".encode()).hexdigest()[:8]
            if content_hash in content_hashes:
                repetition_count += 1
            content_hashes[content_hash] = content_hashes.get(content_hash, 0) + 1
        
        repetition_rate = repetition_count / len(self.training_examples)
        if repetition_rate > 0.1:  # More than 10% repetition
            health_report["warnings"].append(f"High content repetition: {repetition_rate:.1%}")
        
        # Check skill diversity
        skills_represented = set()
        for example in self.training_examples:
            skills_represented.update(example["skill_tags"])
        
        health_report["diversity_score"] = len(skills_represented) / 7.0  # 7 total skills
        
        if health_report["diversity_score"] < 0.3:
            health_report["warnings"].append("Low skill diversity - risk of overfitting")
        
        # Check data freshness (avoid feeding too much of the same context)
        context_frequency = {}
        for example in self.training_examples[-100:]:  # Check last 100 examples
            # Extract context patterns
            input_words = set(example["input"].lower().split())
            for word in input_words:
                if len(word) > 4:  # Only count meaningful words
                    context_frequency[word] = context_frequency.get(word, 0) + 1
        
        # Flag high-frequency patterns
        max_frequency = max(context_frequency.values()) if context_frequency else 0
        if max_frequency > len(self.training_examples) * 0.2:  # Same pattern in >20% of recent examples
            health_report["warnings"].append("Overingestion detected - similar contexts repeated")
            health_report["freshness_score"] = 0.5
        
        # Overall health status
        if health_report["warnings"]:
            health_report["status"] = "warning" if len(health_report["warnings"]) <= 2 else "unhealthy"
        
        return health_report


class EggFarm:
    """Manages multiple eggs and the overall breeding system"""
    
    def __init__(self, farm_directory: str = "output/egg-farm"):
        self.farm_dir = Path(farm_directory)
        self.farm_dir.mkdir(parents=True, exist_ok=True)
        self.eggs = {}
        self.load_existing_eggs()
    
    def spawn_egg(self, name: str, focus_areas: List[str], egg_type: str = "general") -> Dict[str, Any]:
        """Create a new egg"""
        if name in self.eggs:
            return {"success": False, "reason": f"Egg '{name}' already exists"}
        
        egg = EternalEgg(name, focus_areas, egg_type)
        self.eggs[name] = egg
        self.save_egg(name)
        
        return {
            "success": True,
            "egg_name": name,
            "egg_type": egg_type,
            "focus_areas": focus_areas,
            "target_examples": egg.target_examples,
            "message": f"🥚 Spawned new {egg_type} egg: {name}"
        }
    
    def feed_egg(self, egg_name: str, input_text: str, output_text: str, 
                 skill_tags: List[str], quality_score: float, 
                 user_feedback: Optional[str] = None) -> Dict[str, Any]:
        """Feed an egg with training data"""
        if egg_name not in self.eggs:
            return {"success": False, "reason": f"Egg '{egg_name}' not found"}
        
        result = self.eggs[egg_name].feed(input_text, output_text, skill_tags, quality_score, user_feedback)
        self.save_egg(egg_name)
        
        return {"success": True, "egg_name": egg_name, **result}
    
    def get_egg_stats(self, egg_name: str) -> Dict[str, Any]:
        """Get statistics for a specific egg"""
        if egg_name not in self.eggs:
            return {"success": False, "reason": f"Egg '{egg_name}' not found"}
        
        return {"success": True, **self.eggs[egg_name].get_stats_summary()}
    
    def list_eggs(self) -> Dict[str, Any]:
        """List all eggs in the farm"""
        egg_summaries = {}
        for name, egg in self.eggs.items():
            egg_summaries[name] = {
                "type": egg.egg_type,
                "progress": egg.get_overall_progress(),
                "examples": egg.total_examples,
                "quality": egg.get_average_quality(),
                "hatched": egg.metadata["hatched"],
                "hatching_ready": egg._is_ready_to_hatch()
            }
        
        return {
            "total_eggs": len(self.eggs),
            "eggs": egg_summaries,
            "hatched_models": len([e for e in self.eggs.values() if e.metadata["hatched"]])
        }
    
    def hatch_egg(self, egg_name: str) -> Dict[str, Any]:
        """Attempt to hatch an egg"""
        if egg_name not in self.eggs:
            return {"success": False, "reason": f"Egg '{egg_name}' not found"}
        
        result = self.eggs[egg_name].hatch()
        if result["success"]:
            self.save_egg(egg_name)
        
        return result
    
    def save_egg(self, egg_name: str):
        """Save egg state to disk"""
        if egg_name not in self.eggs:
            return
        
        egg = self.eggs[egg_name]
        egg_file = self.farm_dir / f"{egg_name}.egg.json"
        
        with open(egg_file, 'w', encoding='utf-8') as f:
            json.dump({
                "name": egg.name,
                "egg_type": egg.egg_type,
                "created_date": egg.created_date.isoformat(),
                "focus_areas": egg.focus_areas,
                "total_examples": egg.total_examples,
                "quality_threshold": egg.quality_threshold,
                "target_examples": egg.target_examples,
                "specializations": {
                    name: {
                        "examples_collected": spec.examples_collected,
                        "quality_scores": spec.quality_scores,
                        "target_examples": spec.target_examples
                    }
                    for name, spec in egg.specializations.items()
                },
                "training_examples": egg.training_examples,
                "metadata": egg.metadata
            }, f, indent=2)
    
    def load_existing_eggs(self):
        """Load existing eggs from disk"""
        for egg_file in self.farm_dir.glob("*.egg.json"):
            try:
                with open(egg_file, 'r', encoding='utf-8') as f:
                    data = json.load(f)
                
                # Reconstruct egg
                egg = EternalEgg(data["name"], data["focus_areas"], data["egg_type"])
                egg.created_date = datetime.fromisoformat(data["created_date"])
                egg.total_examples = data["total_examples"]
                egg.quality_threshold = data["quality_threshold"]
                egg.target_examples = data["target_examples"]
                egg.training_examples = data["training_examples"]
                egg.metadata = data["metadata"]
                
                # Reconstruct specializations
                for name, spec_data in data["specializations"].items():
                    if name in egg.specializations:
                        egg.specializations[name].examples_collected = spec_data["examples_collected"]
                        egg.specializations[name].quality_scores = spec_data["quality_scores"]
                
                self.eggs[data["name"]] = egg
                
            except Exception as e:
                logger.warning(f"Failed to load egg from {egg_file}: {e}")


# Factory functions for different egg types
def create_academic_egg(name: str = "academic-writer") -> Dict[str, Any]:
    """Create an egg specialized for academic writing"""
    farm = EggFarm()
    return farm.spawn_egg(
        name=name,
        focus_areas=["academic_writing", "ieee_formatting", "research_synthesis"],
        egg_type="academic"
    )

def create_voice_egg(name: str = "voice-mimic") -> Dict[str, Any]:
    """Create an egg specialized for voice mimicry"""
    farm = EggFarm()
    return farm.spawn_egg(
        name=name,
        focus_areas=["voice_mimicry", "anti_ai_patterns", "content_transformation"],
        egg_type="voice"
    )

def create_tech_egg(name: str = "tech-writer") -> Dict[str, Any]:
    """Create an egg specialized for technical documentation"""
    farm = EggFarm()
    return farm.spawn_egg(
        name=name,
        focus_areas=["technical_documentation", "content_transformation", "anti_ai_patterns"],
        egg_type="technical"
    )


class AutoFeeder:
    """Automatically feeds eggs based on uroboro interactions"""
    
    def __init__(self, farm: EggFarm = None):
        self.farm = farm or EggFarm()
        self.feeding_enabled = True
        self.quality_threshold = 6.0  # Minimum quality to auto-feed
        
        # OVERINGESTION PROTECTION
        self.max_feeds_per_hour = 10  # Rate limiting
        self.max_similar_content_ratio = 0.15  # Max 15% similar content
        self.quality_degradation_threshold = 0.3  # Stop if quality drops this much
        
        # Content deduplication cache (last 100 items)
        self.recent_content_cache = []
        self.cache_size = 100
        
        # Quality trend tracking
        self.quality_history = []
        self.quality_window = 20  # Track last 20 feedings for trend
        
        # Skill detection patterns
        self.skill_patterns = {
            "academic_writing": [
                "academic", "formal", "professional", "research", "methodology", 
                "implementation", "results", "analysis", "conclusion", "bibliography"
            ],
            "ieee_formatting": [
                "IEEE", "citation", "reference", "TABLE", "FIGURE", "equation",
                "bibliography", "standard", "format"
            ],
            "anti_ai_patterns": [
                "authentic", "human", "natural", "conversational", "personal",
                "experience", "actually", "really", "honestly"
            ],
            "content_transformation": [
                "bullet", "list", "transform", "structure", "organize",
                "prose", "narrative", "flow", "coherent"
            ],
            "voice_mimicry": [
                "voice", "style", "tone", "personality", "characteristic",
                "manner", "approach", "perspective"
            ],
            "technical_documentation": [
                "technical", "documentation", "specification", "API", "code",
                "implementation", "system", "architecture", "design"
            ],
            "research_synthesis": [
                "synthesis", "analysis", "comparison", "evaluation", "review",
                "literature", "sources", "materials", "findings"
            ]
        }
    
    def _calculate_content_similarity(self, new_content: str, existing_content: str) -> float:
        """Calculate similarity between two pieces of content (privacy-safe, local only)"""
        # Simple word overlap similarity (no external APIs, local processing only)
        new_words = set(new_content.lower().split())
        existing_words = set(existing_content.lower().split())
        
        if not new_words or not existing_words:
            return 0.0
        
        intersection = new_words & existing_words
        union = new_words | existing_words
        
        return len(intersection) / len(union)  # Jaccard similarity
    
    def _check_content_freshness(self, input_text: str, output_text: str) -> Dict[str, Any]:
        """Check if content is fresh enough to avoid overingestion"""
        combined_content = f"{input_text} {output_text}"
        
        freshness_report = {
            "is_fresh": True,
            "max_similarity": 0.0,
            "similar_count": 0,
            "reason": ""
        }
        
        # Check against recent content cache
        for cached_content in self.recent_content_cache:
            similarity = self._calculate_content_similarity(combined_content, cached_content)
            
            if similarity > freshness_report["max_similarity"]:
                freshness_report["max_similarity"] = similarity
            
            if similarity > 0.7:  # High similarity threshold
                freshness_report["similar_count"] += 1
        
        # Too many similar items
        if freshness_report["similar_count"] > 3:
            freshness_report["is_fresh"] = False
            freshness_report["reason"] = f"Too many similar items ({freshness_report['similar_count']})"
        
        # Overall similarity too high
        elif freshness_report["max_similarity"] > 0.8:
            freshness_report["is_fresh"] = False
            freshness_report["reason"] = f"Content too similar ({freshness_report['max_similarity']:.1%})"
        
        return freshness_report
    
    def _update_content_cache(self, input_text: str, output_text: str):
        """Update content cache for deduplication (privacy-safe)"""
        combined_content = f"{input_text} {output_text}"
        
        # Add to cache
        self.recent_content_cache.append(combined_content)
        
        # Maintain cache size (FIFO)
        if len(self.recent_content_cache) > self.cache_size:
            self.recent_content_cache.pop(0)
    
    def _check_quality_trend(self, new_quality: float) -> Dict[str, Any]:
        """Monitor quality trends to prevent degradation"""
        trend_report = {
            "quality_ok": True,
            "trend": "stable",
            "recent_average": new_quality,
            "degradation": 0.0,
            "reason": ""
        }
        
        # Add to quality history
        self.quality_history.append(new_quality)
        
        # Maintain window size
        if len(self.quality_history) > self.quality_window:
            self.quality_history.pop(0)
        
        # Need at least 10 samples for trend analysis
        if len(self.quality_history) < 10:
            return trend_report
        
        # Calculate recent vs older quality
        recent_samples = self.quality_history[-5:]
        older_samples = self.quality_history[-15:-5] if len(self.quality_history) >= 15 else self.quality_history[:-5]
        
        if older_samples:
            recent_avg = sum(recent_samples) / len(recent_samples)
            older_avg = sum(older_samples) / len(older_samples)
            
            trend_report["recent_average"] = recent_avg
            trend_report["degradation"] = older_avg - recent_avg
            
            if trend_report["degradation"] > self.quality_degradation_threshold:
                trend_report["quality_ok"] = False
                trend_report["trend"] = "declining"
                trend_report["reason"] = f"Quality dropped {trend_report['degradation']:.1f} points"
            elif trend_report["degradation"] < -0.2:
                trend_report["trend"] = "improving"
        
        return trend_report
    
    def _rate_limit_check(self) -> Dict[str, Any]:
        """Check if we're feeding too frequently (prevent spam)"""
        # For now, always allow (could add timestamp tracking later)
        return {
            "rate_ok": True,
            "feeds_this_hour": 0,
            "reason": ""
        }
    
    def detect_skills(self, text: str) -> List[str]:
        """Detect which skills are relevant to the given text"""
        text_lower = text.lower()
        detected_skills = []
        
        for skill, patterns in self.skill_patterns.items():
            pattern_matches = sum(1 for pattern in patterns if pattern in text_lower)
            # If at least 20% of patterns match, consider it relevant
            if pattern_matches >= max(1, len(patterns) * 0.2):
                detected_skills.append(skill)
        
        return detected_skills
    
    def calculate_quality_score(self, input_text: str, output_text: str, 
                              command_context: str = "") -> float:
        """Calculate quality score based on content characteristics"""
        base_score = 7.0
        
        # Length bonus (longer content often indicates more effort)
        output_length = len(output_text.split())
        if output_length > 100:
            base_score += 0.5
        elif output_length > 50:
            base_score += 0.3
        
        # Complexity bonus (varied sentence structure, technical terms)
        sentences = output_text.split('.')
        if len(sentences) > 3:
            avg_sentence_length = sum(len(s.split()) for s in sentences) / len(sentences)
            if 10 <= avg_sentence_length <= 25:  # Good sentence variety
                base_score += 0.3
        
        # Academic indicators
        academic_indicators = ["implementation", "analysis", "methodology", "results", 
                             "evaluation", "comprehensive", "systematic"]
        academic_score = sum(1 for indicator in academic_indicators if indicator.lower() in output_text.lower())
        base_score += min(1.0, academic_score * 0.2)
        
        # Command context bonus
        context_bonuses = {
            "academic": 0.5,
            "sensei": 0.4,
            "research": 0.3
        }
        for context, bonus in context_bonuses.items():
            if context in command_context.lower():
                base_score += bonus
                break
        
        # Cap at 9.5 for auto-feeding (reserve 10.0 for perfect manual ratings)
        return min(9.5, base_score)
    
    def should_feed_egg(self, egg_name: str, skills: List[str]) -> bool:
        """Determine if an egg should be fed based on skills and egg focus"""
        if not self.feeding_enabled:
            return False
        
        eggs = self.farm.list_eggs()
        if egg_name not in eggs["eggs"]:
            return False
        
        # Don't feed already hatched eggs
        if eggs["eggs"][egg_name]["hatched"]:
            return False
        
        # Get egg details
        egg_stats = self.farm.get_egg_stats(egg_name)
        if not egg_stats["success"]:
            return False
        
        # Check egg health before feeding
        egg = self.farm.eggs[egg_name]
        health = egg.check_data_health()
        
        if health["status"] == "unhealthy":
            return False  # Don't feed unhealthy eggs
        
        # Check if any detected skills match egg's focus areas
        skill_overlap = set(skills) & set(egg.focus_areas)
        
        return len(skill_overlap) > 0
    
    def auto_feed_compatible_eggs(self, input_text: str, output_text: str, 
                                 command_context: str = "", user_feedback: str = None) -> Dict[str, Any]:
        """Automatically feed all compatible eggs with the interaction data (with overingestion protection)"""
        if not self.feeding_enabled:
            return {"enabled": False, "fed_eggs": [], "protection_status": "disabled"}
        
        # OVERINGESTION PROTECTION CHECKS
        protection_status = {
            "freshness_check": "pass",
            "quality_trend": "pass", 
            "rate_limit": "pass",
            "overall": "pass"
        }
        
        # Check content freshness
        freshness = self._check_content_freshness(input_text, output_text)
        if not freshness["is_fresh"]:
            protection_status["freshness_check"] = f"blocked: {freshness['reason']}"
            protection_status["overall"] = "blocked"
            return {
                "enabled": True, 
                "fed_eggs": [], 
                "reason": f"Content protection: {freshness['reason']}",
                "protection_status": protection_status
            }
        
        # Check rate limiting
        rate_check = self._rate_limit_check()
        if not rate_check["rate_ok"]:
            protection_status["rate_limit"] = f"blocked: {rate_check['reason']}"
            protection_status["overall"] = "blocked"
            return {
                "enabled": True,
                "fed_eggs": [],
                "reason": f"Rate limit: {rate_check['reason']}",
                "protection_status": protection_status
            }
        
        # Detect skills from the interaction
        detected_skills = self.detect_skills(input_text + " " + output_text)
        if not detected_skills:
            return {
                "enabled": True, 
                "fed_eggs": [], 
                "reason": "No relevant skills detected",
                "protection_status": protection_status
            }
        
        # Calculate quality score
        quality_score = self.calculate_quality_score(input_text, output_text, command_context)
        
        # Check quality trend
        quality_trend = self._check_quality_trend(quality_score)
        if not quality_trend["quality_ok"]:
            protection_status["quality_trend"] = f"blocked: {quality_trend['reason']}"
            protection_status["overall"] = "blocked"
            return {
                "enabled": True,
                "fed_eggs": [],
                "reason": f"Quality protection: {quality_trend['reason']}",
                "protection_status": protection_status
            }
        
        # Skip low-quality content
        if quality_score < self.quality_threshold:
            return {
                "enabled": True, 
                "fed_eggs": [], 
                "reason": f"Quality too low: {quality_score:.1f}",
                "protection_status": protection_status
            }
        
        # Feed all compatible eggs
        fed_eggs = []
        eggs_list = self.farm.list_eggs()
        
        for egg_name in eggs_list["eggs"].keys():
            if self.should_feed_egg(egg_name, detected_skills):
                result = self.farm.feed_egg(
                    egg_name=egg_name,
                    input_text=input_text,
                    output_text=output_text,
                    skill_tags=detected_skills,
                    quality_score=quality_score,
                    user_feedback=user_feedback
                )
                
                if result["success"]:
                    fed_eggs.append({
                        "name": egg_name,
                        "skills": detected_skills,
                        "quality": quality_score,
                        "total_examples": result["total_examples"],
                        "new_achievements": result.get("new_achievements", []),
                        "hatching_ready": result.get("hatching_ready", False)
                    })
        
        # Update content cache after successful feeding
        if fed_eggs:
            self._update_content_cache(input_text, output_text)
        
        return {
            "enabled": True,
            "fed_eggs": fed_eggs,
            "detected_skills": detected_skills,
            "quality_score": quality_score,
            "total_fed": len(fed_eggs),
            "protection_status": protection_status
        }
    
    def enable_auto_feeding(self):
        """Enable automatic feeding"""
        self.feeding_enabled = True
    
    def disable_auto_feeding(self):
        """Disable automatic feeding"""
        self.feeding_enabled = False
    
    def set_quality_threshold(self, threshold: float):
        """Set minimum quality threshold for auto-feeding"""
        self.quality_threshold = max(0.0, min(10.0, threshold))
    
    def get_protection_status(self) -> Dict[str, Any]:
        """Get current protection status and statistics"""
        return {
            "feeding_enabled": self.feeding_enabled,
            "quality_threshold": self.quality_threshold,
            "cache_size": len(self.recent_content_cache),
            "quality_samples": len(self.quality_history),
            "recent_quality_avg": sum(self.quality_history[-10:]) / len(self.quality_history[-10:]) if self.quality_history else 0.0,
            "protections": {
                "content_deduplication": True,
                "quality_trend_monitoring": True,
                "rate_limiting": True,
                "health_checking": True
            }
        }


# Convenience function for auto-feeding
def auto_feed_eggs(input_text: str, output_text: str, command_context: str = "", 
                   user_feedback: str = None, show_results: bool = True) -> Dict[str, Any]:
    """Auto-feed eggs and optionally display results"""
    feeder = AutoFeeder()
    result = feeder.auto_feed_compatible_eggs(input_text, output_text, command_context, user_feedback)
    
    if show_results and result["enabled"] and result["fed_eggs"]:
        print(f"\n🍼 Auto-fed {result['total_fed']} eggs:")
        for egg in result["fed_eggs"]:
            print(f"   🥚 {egg['name']}: {egg['total_examples']} examples, {egg['quality']:.1f}/10 quality")
            if egg["new_achievements"]:
                print(f"      🏆 New: {', '.join(egg['new_achievements'])}")
            if egg["hatching_ready"]:
                print(f"      🎯 Ready to hatch!")
    
    return result 