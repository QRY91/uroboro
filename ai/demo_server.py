#!/usr/bin/env python3
"""
uroboro AI Semantic Search Demo Server

A lightweight web interface showcasing local-first AI semantic search
using ChromaDB + Ollama integration.

Usage:
    python demo_server.py

Then visit: http://localhost:5000
"""

import os
import sys
import json
from datetime import datetime
from flask import Flask, render_template_string, request, jsonify

# Add current directory to path to import our ChromaDB integration
sys.path.append(os.path.dirname(os.path.abspath(__file__)))

try:
    from chromadb_integration import UroboroChromaDBIntegration
except ImportError as e:
    print(f"❌ Failed to import ChromaDB integration: {e}")
    print("Make sure chromadb_integration.py is in the same directory")
    sys.exit(1)

app = Flask(__name__)

# Global integration instance
integration = None

# HTML template for the demo interface
HTML_TEMPLATE = """
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>uroboro AI Semantic Search Demo</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            color: #333;
        }

        .container {
            background: white;
            border-radius: 12px;
            padding: 30px;
            box-shadow: 0 10px 30px rgba(0,0,0,0.2);
        }

        .header {
            text-align: center;
            margin-bottom: 30px;
        }

        .header h1 {
            color: #2d3748;
            margin: 0;
            font-size: 2.5rem;
        }

        .header p {
            color: #718096;
            margin: 10px 0;
            font-size: 1.1rem;
        }

        .stats {
            display: flex;
            justify-content: center;
            gap: 30px;
            margin-bottom: 30px;
            flex-wrap: wrap;
        }

        .stat {
            text-align: center;
            padding: 15px;
            background: #f7fafc;
            border-radius: 8px;
            border: 1px solid #e2e8f0;
        }

        .stat-value {
            font-size: 1.8rem;
            font-weight: bold;
            color: #2b6cb0;
        }

        .stat-label {
            color: #718096;
            font-size: 0.9rem;
        }

        .search-form {
            margin-bottom: 30px;
        }

        .search-input {
            width: 100%;
            padding: 15px;
            border: 2px solid #e2e8f0;
            border-radius: 8px;
            font-size: 1.1rem;
            margin-bottom: 15px;
            transition: border-color 0.2s;
        }

        .search-input:focus {
            outline: none;
            border-color: #4299e1;
        }

        .search-btn {
            background: #4299e1;
            color: white;
            border: none;
            padding: 15px 30px;
            border-radius: 8px;
            font-size: 1.1rem;
            cursor: pointer;
            transition: background 0.2s;
        }

        .search-btn:hover {
            background: #3182ce;
        }

        .search-btn:disabled {
            background: #a0aec0;
            cursor: not-allowed;
        }

        .loading {
            text-align: center;
            color: #718096;
            display: none;
        }

        .results {
            margin-top: 30px;
        }

        .result {
            background: #f7fafc;
            border: 1px solid #e2e8f0;
            border-radius: 8px;
            padding: 20px;
            margin-bottom: 15px;
            transition: transform 0.2s;
        }

        .result:hover {
            transform: translateY(-2px);
            box-shadow: 0 4px 12px rgba(0,0,0,0.1);
        }

        .result-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 10px;
        }

        .result-id {
            font-weight: bold;
            color: #2d3748;
        }

        .similarity-score {
            background: #48bb78;
            color: white;
            padding: 4px 8px;
            border-radius: 4px;
            font-size: 0.9rem;
        }

        .result-meta {
            color: #718096;
            font-size: 0.9rem;
            margin-bottom: 10px;
        }

        .result-content {
            color: #2d3748;
            line-height: 1.6;
        }

        .tags {
            margin-top: 10px;
        }

        .tag {
            background: #bee3f8;
            color: #2c5282;
            padding: 2px 6px;
            border-radius: 4px;
            font-size: 0.8rem;
            margin-right: 5px;
        }

        .footer {
            text-align: center;
            margin-top: 40px;
            padding-top: 20px;
            border-top: 1px solid #e2e8f0;
            color: #718096;
        }

        .tech-stack {
            display: flex;
            justify-content: center;
            gap: 15px;
            margin-top: 10px;
            flex-wrap: wrap;
        }

        .tech-item {
            background: #edf2f7;
            padding: 5px 10px;
            border-radius: 4px;
            font-size: 0.9rem;
        }

        .example-queries {
            margin-bottom: 20px;
            text-align: center;
        }

        .example-query {
            display: inline-block;
            background: #edf2f7;
            color: #4a5568;
            padding: 5px 10px;
            margin: 2px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 0.9rem;
            transition: background 0.2s;
        }

        .example-query:hover {
            background: #e2e8f0;
        }

        @media (max-width: 768px) {
            .stats {
                flex-direction: column;
                gap: 15px;
            }

            .header h1 {
                font-size: 2rem;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🤖 uroboro AI Semantic Search</h1>
            <p>Local-first AI-powered search across your captures</p>
            <p><strong>ChromaDB + Ollama + nomic-embed-text</strong></p>
        </div>

        <div class="stats">
            <div class="stat">
                <div class="stat-value" id="total-captures">{{ stats.total_captures }}</div>
                <div class="stat-label">Total Captures</div>
            </div>
            <div class="stat">
                <div class="stat-value" id="embedded-captures">{{ stats.embedded_captures }}</div>
                <div class="stat-label">Embedded</div>
            </div>
            <div class="stat">
                <div class="stat-value" id="coverage">{{ "%.1f"|format(stats.coverage_percent) }}%</div>
                <div class="stat-label">Coverage</div>
            </div>
        </div>

        <div class="search-form">
            <div class="example-queries">
                <strong>Try these examples:</strong><br>
                <span class="example-query" onclick="setQuery('PostHog AI integration')">PostHog AI integration</span>
                <span class="example-query" onclick="setQuery('local-first AI cost savings')">local-first AI cost savings</span>
                <span class="example-query" onclick="setQuery('ESP32 hardware development')">ESP32 hardware development</span>
                <span class="example-query" onclick="setQuery('QRY methodology')">QRY methodology</span>
                <span class="example-query" onclick="setQuery('vector database semantic search')">vector database</span>
            </div>

            <form id="search-form">
                <input
                    type="text"
                    id="search-query"
                    class="search-input"
                    placeholder="Enter your semantic search query..."
                    value="{{ query or '' }}"
                >
                <button type="submit" class="search-btn" id="search-btn">
                    🔍 Search Semantically
                </button>
            </form>

            <div class="loading" id="loading">
                🔄 Searching through {{ stats.embedded_captures }} embedded captures...
            </div>
        </div>

        <div class="results" id="results">
            {% if results %}
                <h3>Found {{ results|length }} results{% if query %} for "{{ query }}"{% endif %}:</h3>
                {% for result in results %}
                <div class="result">
                    <div class="result-header">
                        <span class="result-id">Capture #{{ result.capture_id }}</span>
                        <span class="similarity-score">{{ "%.1f"|format(result.similarity * 100) }}% match</span>
                    </div>
                    <div class="result-meta">
                        Created: {{ result.metadata.created_at or 'N/A' }}
                        {% if result.metadata.project %} | Project: {{ result.metadata.project }}{% endif %}
                    </div>
                    <div class="result-content">{{ result.content }}</div>
                    {% if result.metadata.tags %}
                    <div class="tags">
                        {% for tag in result.metadata.tags.split(',') %}
                        <span class="tag">{{ tag.strip() }}</span>
                        {% endfor %}
                    </div>
                    {% endif %}
                </div>
                {% endfor %}
            {% endif %}
        </div>

        <div class="footer">
            <p><strong>Local-First AI Demo</strong> • No cloud dependencies • Zero API costs</p>
            <div class="tech-stack">
                <span class="tech-item">ChromaDB</span>
                <span class="tech-item">Ollama</span>
                <span class="tech-item">nomic-embed-text</span>
                <span class="tech-item">Python + Flask</span>
                <span class="tech-item">SQLite</span>
            </div>
            <p style="margin-top: 15px; font-size: 0.9rem;">
                Built in 24 hours from PostHog inspiration to working prototype
            </p>
        </div>
    </div>

    <script>
        function setQuery(query) {
            document.getElementById('search-query').value = query;
        }

        document.getElementById('search-form').addEventListener('submit', function(e) {
            e.preventDefault();

            const query = document.getElementById('search-query').value.trim();
            if (!query) return;

            const btn = document.getElementById('search-btn');
            const loading = document.getElementById('loading');

            btn.disabled = true;
            btn.textContent = '🔄 Searching...';
            loading.style.display = 'block';

            // Submit the form with the query
            window.location.href = '/?q=' + encodeURIComponent(query);
        });

        // Auto-focus on search input
        document.getElementById('search-query').focus();
    </script>
</body>
</html>
"""

@app.route('/')
def index():
    """Main demo page with search interface."""
    query = request.args.get('q', '').strip()
    results = []

    # Get statistics
    try:
        stats = integration.get_stats()
    except Exception as e:
        print(f"Failed to get stats: {e}")
        stats = {
            'total_captures': 0,
            'embedded_captures': 0,
            'coverage_percent': 0
        }

    # Perform search if query provided
    if query:
        try:
            results = integration.semantic_search(query, limit=10)
            print(f"Search for '{query}' returned {len(results)} results")
        except Exception as e:
            print(f"Search failed: {e}")
            results = []

    return render_template_string(HTML_TEMPLATE,
                                query=query,
                                results=results,
                                stats=stats)

@app.route('/api/search')
def api_search():
    """API endpoint for semantic search."""
    query = request.args.get('q', '').strip()
    limit = int(request.args.get('limit', 10))

    if not query:
        return jsonify({'error': 'Query parameter required'}), 400

    try:
        results = integration.semantic_search(query, limit=limit)
        return jsonify({
            'query': query,
            'results': [
                {
                    'capture_id': r.capture_id,
                    'content': r.content,
                    'similarity': r.similarity,
                    'metadata': r.metadata
                } for r in results
            ]
        })
    except Exception as e:
        return jsonify({'error': str(e)}), 500

@app.route('/api/stats')
def api_stats():
    """API endpoint for statistics."""
    try:
        stats = integration.get_stats()
        return jsonify(stats)
    except Exception as e:
        return jsonify({'error': str(e)}), 500

@app.route('/health')
def health():
    """Health check endpoint."""
    try:
        # Test all connections
        test_results = integration.test_connection()

        if all(test_results.values()):
            return jsonify({
                'status': 'healthy',
                'connections': test_results,
                'timestamp': datetime.now().isoformat()
            })
        else:
            return jsonify({
                'status': 'degraded',
                'connections': test_results,
                'timestamp': datetime.now().isoformat()
            }), 503

    except Exception as e:
        return jsonify({
            'status': 'unhealthy',
            'error': str(e),
            'timestamp': datetime.now().isoformat()
        }), 503

def main():
    """Initialize and run the demo server."""
    global integration

    print("🚀 Starting uroboro AI Semantic Search Demo")
    print("=" * 50)

    # Initialize ChromaDB integration
    try:
        print("🔄 Initializing ChromaDB integration...")
        integration = UroboroChromaDBIntegration()
        print("✅ ChromaDB integration ready")
    except Exception as e:
        print(f"❌ Failed to initialize ChromaDB integration: {e}")
        print("\nMake sure:")
        print("1. ChromaDB is installed: pip install chromadb")
        print("2. Ollama is running with nomic-embed-text model")
        print("3. uroboro database exists")
        sys.exit(1)

    # Test connections
    try:
        print("🔄 Testing connections...")
        test_results = integration.test_connection()

        status_map = {True: "✅", False: "❌"}
        print(f"   ChromaDB: {status_map[test_results['chromadb']]}")
        print(f"   Ollama: {status_map[test_results['ollama']]}")
        print(f"   Uroboro DB: {status_map[test_results['uroboro_db']]}")

        if not all(test_results.values()):
            print("\n⚠️  Some connections failed. Demo may not work properly.")
        else:
            print("\n🎉 All systems ready!")

    except Exception as e:
        print(f"⚠️  Connection test failed: {e}")

    # Get statistics
    try:
        stats = integration.get_stats()
        print(f"\n📊 Current Statistics:")
        print(f"   Total Captures: {stats['total_captures']}")
        print(f"   Embedded: {stats['embedded_captures']}")
        print(f"   Coverage: {stats['coverage_percent']:.1f}%")
    except Exception as e:
        print(f"⚠️  Failed to get stats: {e}")

    print(f"\n🌐 Demo server starting...")
    print(f"   URL: http://localhost:5000")
    print(f"   API: http://localhost:5000/api/search?q=your+query")
    print(f"   Health: http://localhost:5000/health")
    print(f"\n🔍 Try searching for:")
    print(f"   • PostHog AI integration")
    print(f"   • local-first AI cost savings")
    print(f"   • ESP32 hardware development")
    print(f"   • QRY methodology")
    print(f"\n📱 Use Ctrl+C to stop the server")
    print("=" * 50)

    # Run the Flask app
    app.run(host='0.0.0.0', port=5000, debug=False)

if __name__ == '__main__':
    main()
