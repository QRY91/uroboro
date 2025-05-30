#!/usr/bin/env python3
"""Test LLM integration"""

from src.processors.content_generator import ContentGenerator
from src.aggregator import ContentAggregator

def main():
    print("Testing LLM integration...")
    
    # Test content generator initialization
    generator = ContentGenerator()
    print(f"Using model: {generator.llm_model}")
    
    # Test simple LLM call
    print("Testing simple LLM call...")
    result = generator._call_ollama("Say hello and tell me what model you are.")
    print(f"LLM Response: {result}")
    
    # Test with real activity data
    print("\nTesting with activity data...")
    aggregator = ContentAggregator()
    activity = aggregator.collect_recent_activity(days=1)
    
    print(f"Activity summary: {len(activity.get('projects', {}))} projects, {len(activity.get('daily_notes', []))} daily notes")
    
    if activity.get('projects') or activity.get('daily_notes'):
        print("Generating devlog summary...")
        summary = generator.generate_devlog_summary(activity)
        print("--- DEVLOG SUMMARY ---")
        print(summary)
        print("--- END SUMMARY ---")
    else:
        print("No activity found to summarize")

if __name__ == "__main__":
    main() 