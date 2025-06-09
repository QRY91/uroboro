#!/bin/bash
# Note: Not using 'set -e' because we need to handle expected failures

# Integration Test Suite for uroboro
# Validates end-to-end Trinity functionality
# Based on systematic manual testing validation

echo "🧪 uroboro Integration Test Suite"
echo "================================="

# Test configuration
TEST_DB="/tmp/uroboro_test_$(date +%s).sqlite"
TEST_PROJECT="integration-test"
BINARY="$(pwd)/uroboro"

# Cleanup function
cleanup() {
    echo "🧹 Cleaning up test artifacts..."
    rm -f "$TEST_DB"
    rm -rf /tmp/test_integration_project
}
trap cleanup EXIT

# Color output functions
green() { echo -e "\033[32m$1\033[0m"; }
red() { echo -e "\033[31m$1\033[0m"; }
yellow() { echo -e "\033[33m$1\033[0m"; }

# Test result tracking
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_SKIPPED=0

test_result() {
    local exit_code=$1
    local test_name="$2"
    if [ $exit_code -eq 0 ]; then
        green "✅ $test_name"
        ((TESTS_PASSED++))
    else
        red "❌ $test_name"
        ((TESTS_FAILED++))
        return 1
    fi
}

echo
echo "📋 Test Plan:"
echo "  1. Build Verification"
echo "  2. Basic Functionality (capture/publish/status)"
echo "  3. Trinity Intelligence (auto-detection/tagging)" 
echo "  4. Project Filtering"
echo "  5. Database Integration"
echo "  6. Ripcord Functionality"
echo "  7. Error Handling"
echo "  8. Edge Cases"
echo

# Test 1: Build Verification
echo "🔨 Test 1: Build Verification"
echo "=============================="

go build -o "$BINARY" ./cmd/uroboro
BUILD_EXIT_CODE=$?
test_result $BUILD_EXIT_CODE "Binary compilation"

if [ ! -f "$BINARY" ]; then
    red "❌ Binary not created"
    exit 1
fi

# Test help output (uroboro exits 1 for unknown --help but shows usage)
HELP_OUTPUT=$($BINARY --help 2>&1)
HELP_EXIT_CODE=$?
if [ $HELP_EXIT_CODE -eq 1 ] && echo "$HELP_OUTPUT" | grep -q "uroboro - The Unified Development Assistant"; then
    test_result 0 "Binary execution and help output"
else
    red "❌ Help output test failed (exit code: $HELP_EXIT_CODE)"
    ((TESTS_FAILED++))
fi

echo

# Test 2: Basic Functionality
echo "📝 Test 2: Basic Functionality"
echo "==============================="

# Test capture
$BINARY capture "Integration test capture - basic functionality validation" --db "$TEST_DB" > /dev/null
CAPTURE_EXIT_CODE=$?
test_result $CAPTURE_EXIT_CODE "Basic capture command"

# Test status
STATUS_OUTPUT=$($BINARY status --db "$TEST_DB" 2>/dev/null)
STATUS_EXIT_CODE=$?
if [ $STATUS_EXIT_CODE -eq 0 ] && echo "$STATUS_OUTPUT" | grep -q "Integration test capture"; then
    test_result 0 "Status command shows captured content"
else
    red "❌ Status command doesn't show captured content (exit code: $STATUS_EXIT_CODE)"
    ((TESTS_FAILED++))
fi

# Test publish (requires ollama, skip if not available)
if command -v ollama >/dev/null 2>&1; then
    $BINARY publish --devlog --days 1 --db "$TEST_DB" --preview > /dev/null 2>&1
    PUBLISH_EXIT_CODE=$?
    test_result $PUBLISH_EXIT_CODE "Publish command with AI generation"
else
    yellow "⚠️  Skipping publish test (ollama not available)"
    ((TESTS_SKIPPED++))
fi

echo

# Test 3: Trinity Intelligence
echo "🧠 Test 3: Trinity Intelligence"
echo "================================"

# Create test project directory
mkdir -p /tmp/test_integration_project
cd /tmp/test_integration_project

# Test project auto-detection
CAPTURE_OUTPUT=$($BINARY capture "Testing auto-detection in test project" --db "$TEST_DB" 2>&1)
CAPTURE_EXIT_CODE=$?
if [ $CAPTURE_EXIT_CODE -eq 0 ] && echo "$CAPTURE_OUTPUT" | grep -q "Auto-detected project: test_integration_project"; then
    test_result 0 "Smart project detection"
else
    red "❌ Project auto-detection failed (exit code: $CAPTURE_EXIT_CODE)"
    echo "Output: $CAPTURE_OUTPUT"
    ((TESTS_FAILED++))
fi

# Test auto-tagging
if echo "$CAPTURE_OUTPUT" | grep -q "Auto-detected tags:"; then
    test_result 0 "Auto-tagging functionality"
else
    red "❌ Auto-tagging not working"
    ((TESTS_FAILED++))
fi

cd - > /dev/null

echo

# Test 4: Project Filtering  
echo "🎯 Test 4: Project Filtering"
echo "============================"

# Add captures for different projects
$BINARY capture "Project A content" --db "$TEST_DB" > /dev/null 2>&1
$BINARY capture "Project B content" --db "$TEST_DB" > /dev/null 2>&1

# Test status with project filter
FILTERED_OUTPUT=$($BINARY status --project test_integration_project --db "$TEST_DB" 2>/dev/null)
FILTER_EXIT_CODE=$?
if [ $FILTER_EXIT_CODE -eq 0 ] && echo "$FILTERED_OUTPUT" | grep -q "Testing auto-detection in test project"; then
    test_result 0 "Status project filtering"
else
    red "❌ Status project filtering not working (exit code: $FILTER_EXIT_CODE)"
    echo "Expected to find 'Testing auto-detection in test project' in filtered output"
    ((TESTS_FAILED++))
fi

# Test publish with project filter
if command -v ollama >/dev/null 2>&1; then
    $BINARY publish --devlog --project test_integration_project --days 1 --db "$TEST_DB" --preview > /dev/null 2>&1
    PUBLISH_FILTER_EXIT_CODE=$?
    test_result $PUBLISH_FILTER_EXIT_CODE "Publish project filtering"
else
    yellow "⚠️  Skipping publish project filter test (ollama not available)"
    ((TESTS_SKIPPED++))
fi

echo

# Test 5: Database Integration
echo "🗄️  Test 5: Database Integration" 
echo "================================"

# Test database creation
if [ -f "$TEST_DB" ]; then
    test_result 0 "Database file creation"
else
    red "❌ Database file not created"
    ((TESTS_FAILED++))
fi

# Test database content
CAPTURE_COUNT=$(sqlite3 "$TEST_DB" "SELECT COUNT(*) FROM captures;" 2>/dev/null || echo "0")
if [ "$CAPTURE_COUNT" -gt 0 ]; then
    test_result 0 "Database content storage (found $CAPTURE_COUNT captures)"
else
    red "❌ No captures found in database"
    ((TESTS_FAILED++))
fi

echo

# Test 6: Ripcord Functionality
echo "📋 Test 6: Ripcord Functionality"
echo "================================="

# Test ripcord (clipboard functionality - just verify it doesn't crash)
$BINARY status --ripcord --db "$TEST_DB" > /dev/null 2>&1
RIPCORD_EXIT_CODE=$?
test_result $RIPCORD_EXIT_CODE "Ripcord command execution"

echo

# Test 7: Error Handling
echo "🚨 Test 7: Error Handling"
echo "=========================="

# Test with non-existent database
$BINARY status --db "/nonexistent/path.sqlite" > /dev/null 2>&1
NONEXISTENT_EXIT_CODE=$?
if [ $NONEXISTENT_EXIT_CODE -ne 0 ]; then
    test_result 0 "Error handling for non-existent database"
else
    red "❌ Should fail with non-existent database"
    ((TESTS_FAILED++))
fi

# Test with invalid project
EMPTY_OUTPUT=$($BINARY publish --devlog --project "nonexistent-project-12345" --days 1 --db "$TEST_DB" 2>&1)
EMPTY_EXIT_CODE=$?
if echo "$EMPTY_OUTPUT" | grep -q "No recent activity found"; then
    test_result 0 "Error handling for empty project filter"
else
    red "❌ Should show 'No recent activity' message (exit code: $EMPTY_EXIT_CODE)"
    echo "Output: $EMPTY_OUTPUT"
    ((TESTS_FAILED++))
fi

echo

# Test 8: Edge Cases
echo "🎲 Test 8: Edge Cases"
echo "====================="

# Test empty capture (should fail due to no content provided)
$BINARY capture > /dev/null 2>&1
EMPTY_CAPTURE_EXIT_CODE=$?
if [ $EMPTY_CAPTURE_EXIT_CODE -ne 0 ]; then
    test_result 0 "Rejects empty capture content"
else
    red "❌ Should reject empty capture (exit code: $EMPTY_CAPTURE_EXIT_CODE)"
    ((TESTS_FAILED++))
fi

# Test with very long content
LONG_CONTENT=$(printf 'A%.0s' {1..1000})
$BINARY capture "$LONG_CONTENT" --db "$TEST_DB" > /dev/null 2>&1
LONG_EXIT_CODE=$?
test_result $LONG_EXIT_CODE "Handles long capture content"

# Test config command
CONFIG_OUTPUT=$($BINARY config 2>/dev/null)
CONFIG_EXIT_CODE=$?
if [ $CONFIG_EXIT_CODE -eq 0 ] && echo "$CONFIG_OUTPUT" | grep -q "Default database:"; then
    test_result 0 "Config command shows database path"
else
    red "❌ Config command not working (exit code: $CONFIG_EXIT_CODE)"
    ((TESTS_FAILED++))
fi

echo

# Test Summary
echo "📊 Test Summary"
echo "==============="
echo "Tests Passed: $TESTS_PASSED"
echo "Tests Failed: $TESTS_FAILED"
echo "Tests Skipped: $TESTS_SKIPPED"
echo "Total Tests:  $((TESTS_PASSED + TESTS_FAILED + TESTS_SKIPPED))"

if [ $TESTS_FAILED -eq 0 ]; then
    if [ $TESTS_SKIPPED -gt 0 ]; then
        green "✅ All available tests passed! ($TESTS_SKIPPED tests skipped due to missing dependencies)"
        yellow "   Note: Full AI integration requires ollama for complete validation"
    else
        green "🎉 All tests passed! Trinity integration fully verified."
    fi
    exit 0
else
    red "💥 $TESTS_FAILED test(s) failed!"
    if [ $TESTS_SKIPPED -gt 0 ]; then
        yellow "   ($TESTS_SKIPPED tests were skipped due to missing dependencies)"
    fi
    echo
    echo "This indicates issues with Trinity functionality that need to be addressed"
    echo "before the code should be considered ready for production."
    exit 1
fi