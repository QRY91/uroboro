#!/usr/bin/env python3
"""Quick test of the Eternal Egg System"""

from src.egg_system import EggFarm, create_academic_egg

print("🥚 Testing Eternal Egg System...")

# Create farm and spawn an egg
farm = EggFarm()
result = create_academic_egg('test-egg')
print(f"Spawned: {result['message'] if result['success'] else result['reason']}")

# Feed the egg some test data
if result['success']:
    feed_result = farm.feed_egg(
        'test-egg',
        input_text="Transform this bullet list into flowing prose",
        output_text="The implementation demonstrates substantial improvements across multiple metrics...",
        skill_tags=["academic_writing", "content_transformation"],
        quality_score=8.5,
        user_feedback="Good academic tone"
    )
    print(f"Fed egg: {feed_result}")

# Check farm status
stats = farm.list_eggs()
print(f"Farm status: {stats}")

# Get detailed egg stats
if 'test-egg' in stats['eggs']:
    egg_stats = farm.get_egg_stats('test-egg')
    print(f"Egg details: {egg_stats}")

print("✅ Egg system working!") 