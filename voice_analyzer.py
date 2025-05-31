#!/usr/bin/env python3
"""
Voice Analyzer - Extract writing patterns from personal notes
Part of the uroboro content pipeline
"""

import os
import re
import json
from pathlib import Path
from collections import Counter, defaultdict
from typing import Dict, List, Tuple

class VoiceAnalyzer:
    def __init__(self, notes_path: str):
        self.notes_path = Path(notes_path).expanduser()
        self.voice_profile = {
            "sentence_patterns": {},
            "common_phrases": {},
            "technical_terms": {},
            "paragraph_structure": {},
            "tone_indicators": {},
            "writing_habits": {},
        }
    
    def analyze_notes(self) -> Dict:
        """Analyze all markdown and text files in notes directory"""
        print(f"🔍 Analyzing voice patterns in {self.notes_path}...")
        
        # Collect all text content
        all_content = []
        file_count = 0
        
        for file_path in self.notes_path.glob("*.md"):
            try:
                with open(file_path, 'r', encoding='utf-8') as f:
                    content = f.read()
                    all_content.append(content)
                    file_count += 1
            except:
                continue
        
        for file_path in self.notes_path.glob("*.txt"):
            try:
                with open(file_path, 'r', encoding='utf-8') as f:
                    content = f.read()
                    all_content.append(content)
                    file_count += 1
            except:
                continue
        
        print(f"📄 Found {file_count} files to analyze")
        
        # Combine all content
        full_text = "\n\n".join(all_content)
        
        # Run analysis
        self._analyze_sentence_patterns(full_text)
        self._analyze_common_phrases(full_text)
        self._analyze_technical_style(full_text)
        self._analyze_paragraph_structure(full_text)
        self._analyze_tone_indicators(full_text)
        self._analyze_writing_habits(full_text)
        
        return self.voice_profile
    
    def _analyze_sentence_patterns(self, text: str):
        """Extract sentence length and structure preferences"""
        sentences = re.split(r'[.!?]+', text)
        sentences = [s.strip() for s in sentences if s.strip() and len(s.strip()) > 10]
        
        lengths = [len(s.split()) for s in sentences]
        
        self.voice_profile["sentence_patterns"] = {
            "avg_length": sum(lengths) / len(lengths) if lengths else 0,
            "short_sentences": len([l for l in lengths if l < 10]) / len(lengths) if lengths else 0,
            "long_sentences": len([l for l in lengths if l > 20]) / len(lengths) if lengths else 0,
            "sentence_starters": self._extract_sentence_starters(sentences)
        }
    
    def _extract_sentence_starters(self, sentences: List[str]) -> Dict:
        """Find common ways sentences begin"""
        starters = []
        for sentence in sentences:
            words = sentence.strip().split()
            if len(words) >= 2:
                starter = " ".join(words[:2]).lower()
                starters.append(starter)
        
        return dict(Counter(starters).most_common(10))
    
    def _analyze_common_phrases(self, text: str):
        """Extract frequently used phrases and expressions"""
        # Remove code blocks and technical snippets
        clean_text = re.sub(r'```.*?```', '', text, flags=re.DOTALL)
        clean_text = re.sub(r'`[^`]+`', '', clean_text)
        
        # Extract 2-4 word phrases
        words = re.findall(r'\b\w+\b', clean_text.lower())
        
        bigrams = [f"{words[i]} {words[i+1]}" for i in range(len(words)-1)]
        trigrams = [f"{words[i]} {words[i+1]} {words[i+2]}" for i in range(len(words)-2)]
        
        self.voice_profile["common_phrases"] = {
            "bigrams": dict(Counter(bigrams).most_common(15)),
            "trigrams": dict(Counter(trigrams).most_common(10))
        }
    
    def _analyze_technical_style(self, text: str):
        """Analyze technical depth and terminology usage"""
        # Technical indicators
        code_blocks = len(re.findall(r'```', text))
        inline_code = len(re.findall(r'`[^`]+`', text))
        
        # Technical terms (basic heuristic)
        tech_terms = re.findall(r'\b(?:API|CSS|HTML|JS|JSON|HTTP|SQL|Git|React|Python|Node|Docker|AWS|TCP|UDP|REST|GraphQL|WebSocket|OAuth|JWT|CRUD|MVC|SPA|PWA|CDN|DNS|SSL|TLS|HTTPS|SSH|CLI|GUI|IDE|SDK|NPM|Yarn|Webpack|Babel|TypeScript|MongoDB|PostgreSQL|Redis|Nginx|Apache|Linux|Ubuntu|Debian|CentOS|MacOS|Windows|VSCode|Vim|Emacs|IntelliJ|Eclipse|Sublime|Atom)\b', text, re.IGNORECASE)
        
        self.voice_profile["technical_terms"] = {
            "code_blocks": code_blocks,
            "inline_code": inline_code,
            "tech_term_frequency": len(tech_terms) / len(text.split()) if text.split() else 0,
            "most_used_terms": dict(Counter([t.lower() for t in tech_terms]).most_common(10))
        }
    
    def _analyze_paragraph_structure(self, text: str):
        """Analyze paragraph and list usage patterns"""
        paragraphs = [p.strip() for p in text.split('\n\n') if p.strip()]
        
        # Count different structural elements
        lists = len(re.findall(r'^[-*+]\s', text, re.MULTILINE))
        numbered_lists = len(re.findall(r'^\d+\.\s', text, re.MULTILINE))
        headers = len(re.findall(r'^#+\s', text, re.MULTILINE))
        
        self.voice_profile["paragraph_structure"] = {
            "avg_paragraph_length": sum(len(p.split()) for p in paragraphs) / len(paragraphs) if paragraphs else 0,
            "uses_lists": lists > 0,
            "uses_numbered_lists": numbered_lists > 0,
            "uses_headers": headers > 0,
            "list_to_paragraph_ratio": lists / len(paragraphs) if paragraphs else 0
        }
    
    def _analyze_tone_indicators(self, text: str):
        """Extract tone and style indicators"""
        # Punctuation patterns
        exclamations = len(re.findall(r'!', text))
        questions = len(re.findall(r'\?', text))
        ellipses = len(re.findall(r'\.\.\.', text))
        
        # Casual vs formal indicators
        contractions = len(re.findall(r"\b\w+'[a-z]+\b", text, re.IGNORECASE))
        first_person = len(re.findall(r'\b[Ii]\s', text))
        
        self.voice_profile["tone_indicators"] = {
            "exclamation_frequency": exclamations / len(text.split()) if text.split() else 0,
            "question_frequency": questions / len(text.split()) if text.split() else 0,
            "uses_ellipses": ellipses > 0,
            "contraction_frequency": contractions / len(text.split()) if text.split() else 0,
            "first_person_frequency": first_person / len(text.split()) if text.split() else 0
        }
    
    def _analyze_writing_habits(self, text: str):
        """Extract writing habits and preferences"""
        # Meta-commentary (thinking about thinking)
        meta_phrases = re.findall(r'\b(?:I think|I believe|seems like|appears to|probably|maybe|might|could be|not sure|interesting|worth noting|side note|quick note|update|edit|note to self)\b', text, re.IGNORECASE)
        
        # Parenthetical asides
        parentheticals = len(re.findall(r'\([^)]+\)', text))
        
        # Links and references
        links = len(re.findall(r'\[.*?\]\(.*?\)', text))
        
        self.voice_profile["writing_habits"] = {
            "uses_meta_commentary": len(meta_phrases) > 0,
            "meta_frequency": len(meta_phrases) / len(text.split()) if text.split() else 0,
            "uses_parentheticals": parentheticals > 0,
            "parenthetical_frequency": parentheticals / len(text.split()) if text.split() else 0,
            "links_references": links,
            "common_meta_phrases": dict(Counter([p.lower() for p in meta_phrases]).most_common(5))
        }
    
    def generate_voice_prompt(self) -> str:
        """Generate a custom voice prompt based on analysis"""
        profile = self.voice_profile
        
        # Build prompt components
        components = []
        
        # Sentence structure
        if profile["sentence_patterns"]["avg_length"] < 15:
            components.append("Use concise sentences")
        elif profile["sentence_patterns"]["avg_length"] > 20:
            components.append("Use detailed, longer sentences")
        
        # Technical style
        if profile["technical_terms"]["tech_term_frequency"] > 0.05:
            components.append("Include specific technical terminology")
        
        # Structure preferences
        if profile["paragraph_structure"]["uses_lists"]:
            components.append("Use bullet points and lists when appropriate")
        
        # Tone
        if profile["tone_indicators"]["first_person_frequency"] > 0.02:
            components.append("Write in first person")
        
        if profile["tone_indicators"]["contraction_frequency"] > 0.01:
            components.append("Use casual contractions")
        
        # Meta habits
        if profile["writing_habits"]["uses_meta_commentary"]:
            components.append("Include brief meta-commentary and asides")
        
        if profile["writing_habits"]["uses_parentheticals"]:
            components.append("Use parenthetical clarifications")
        
        # Common phrases
        if profile["common_phrases"]["trigrams"]:
            top_phrases = list(profile["common_phrases"]["trigrams"].keys())[:3]
            components.append(f"Consider using phrases like: {', '.join(top_phrases)}")
        
        return ". ".join(components) + "."
    
    def save_profile(self, output_path: str = "voice_profile.json"):
        """Save the voice profile to JSON"""
        with open(output_path, 'w') as f:
            json.dump(self.voice_profile, f, indent=2)
        print(f"💾 Voice profile saved to {output_path}")

def main():
    """Run voice analysis on notes directory"""
    import sys
    
    if len(sys.argv) > 1:
        notes_path = sys.argv[1]
    else:
        # Load from config
        with open('config/settings.json', 'r') as f:
            config = json.load(f)
        notes_path = config['projects']['notes']['path']
    
    analyzer = VoiceAnalyzer(notes_path)
    profile = analyzer.analyze_notes()
    
    print("\n📊 VOICE ANALYSIS RESULTS:")
    print("=" * 40)
    
    # Display key findings
    patterns = profile["sentence_patterns"]
    print(f"📝 Average sentence length: {patterns['avg_length']:.1f} words")
    print(f"📏 Short sentences: {patterns['short_sentences']:.1%}")
    
    tech = profile["technical_terms"]
    print(f"🔧 Technical term frequency: {tech['tech_term_frequency']:.2%}")
    print(f"💻 Code blocks found: {tech['code_blocks']}")
    
    tone = profile["tone_indicators"]
    print(f"❗ Exclamation frequency: {tone['exclamation_frequency']:.3f}")
    print(f"👤 First person usage: {tone['first_person_frequency']:.3f}")
    
    habits = profile["writing_habits"]
    print(f"🧠 Uses meta-commentary: {habits['uses_meta_commentary']}")
    print(f"📎 Uses parentheticals: {habits['uses_parentheticals']}")
    
    print(f"\n🎯 GENERATED VOICE PROMPT:")
    print(analyzer.generate_voice_prompt())
    
    # Save profile
    analyzer.save_profile()

if __name__ == "__main__":
    main() 