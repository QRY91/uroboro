#!/usr/bin/env python3
"""
Auto-feeding demonstration for the uroboro eternal egg system
"""

from src.egg_system import auto_feed_eggs, EggFarm

def demo_auto_feeding():
    print("🥚 UROBORO AUTO-FEEDING DEMONSTRATION")
    print("=" * 50)
    
    # Show current farm status
    farm = EggFarm()
    eggs = farm.list_eggs()
    print(f"📊 Current farm: {eggs['total_eggs']} eggs")
    
    # Demo cases
    demo_cases = [
        {
            "title": "Academic Writing",
            "input": "Generate methodology section for research paper",
            "output": "The implementation methodology employed systematic evaluation protocols to assess performance improvements. Our comprehensive analysis framework incorporated multiple measurement dimensions including response time optimization, system reliability assessment, and user experience evaluation. The systematic approach ensured reproducible results through rigorous testing procedures and statistical validation methods.",
            "context": "academic_research"
        },
        {
            "title": "Technical Documentation", 
            "input": "Document API endpoint specifications",
            "output": "The authentication endpoint accepts POST requests with JSON payload containing username and password credentials. The API returns JWT tokens with configurable expiration times. Error responses follow HTTP status conventions with structured error messages including error codes and detailed descriptions for debugging purposes.",
            "context": "technical_api_docs"
        },
        {
            "title": "Content Transformation",
            "input": "Transform bullet points into flowing prose",
            "output": "The user interface redesign introduced several significant improvements to the overall user experience. Navigation elements were streamlined for intuitive interaction patterns, while visual hierarchy was enhanced through consistent typography and spacing. The responsive design ensures optimal functionality across desktop and mobile platforms.",
            "context": "content_prose_transform"
        }
    ]
    
    print(f"\n🍼 Running {len(demo_cases)} auto-feeding tests...\n")
    
    for i, case in enumerate(demo_cases, 1):
        print(f"🔬 Test {i}: {case['title']}")
        print(f"   Input: {case['input'][:50]}...")
        
        result = auto_feed_eggs(
            case["input"], 
            case["output"], 
            case["context"],
            show_results=False  # We'll show our own summary
        )
        
        if result.get("fed_eggs"):
            print(f"   ✅ Fed {len(result['fed_eggs'])} eggs:")
            for egg in result["fed_eggs"]:
                achievements = f" +{len(egg['new_achievements'])} achievements" if egg['new_achievements'] else ""
                ready = " 🎯 READY TO HATCH!" if egg['hatching_ready'] else ""
                print(f"      🥚 {egg['name']}: {egg['total_examples']} examples, {egg['quality']:.1f}/10{achievements}{ready}")
        else:
            reason = result.get("reason", "No compatible eggs")
            print(f"   ⚪ No eggs fed: {reason}")
        
        print(f"   🧬 Detected skills: {', '.join(result.get('detected_skills', []))}")
        print()
    
    # Final farm status
    print("📊 FINAL FARM STATUS:")
    print("-" * 30)
    eggs = farm.list_eggs()
    for name, info in eggs["eggs"].items():
        status = "🐣 HATCHED" if info["hatched"] else ("🟢 READY" if info["hatching_ready"] else "🥚 Growing")
        print(f"{status} {name}: {info['examples']} examples, {info['quality']:.1f}/10 quality")
    
    print(f"\n🎉 Auto-feeding demonstration complete!")
    print(f"💡 Every uroboro command now feeds compatible eggs automatically!")

if __name__ == "__main__":
    demo_auto_feeding() 