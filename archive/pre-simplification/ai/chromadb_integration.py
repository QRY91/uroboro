#!/usr/bin/env python3
"""
ChromaDB Integration for uroboro AI Features

This module provides vector storage and semantic search capabilities using ChromaDB,
designed to work with the existing uroboro SQLite database and Ollama embeddings.

Usage:
    python chromadb_integration.py embed-all
    python chromadb_integration.py search "query text"
    python chromadb_integration.py stats
"""

import os
import sys
import json
import sqlite3
import argparse
import logging
from typing import List, Dict, Any, Optional, Tuple
from datetime import datetime
import requests

try:
    import chromadb
    from chromadb.config import Settings
except ImportError:
    print("❌ ChromaDB not installed. Install with: pip install chromadb")
    sys.exit(1)


class UroboroChromaDBIntegration:
    """ChromaDB integration for uroboro semantic search and AI features."""

    def __init__(self,
                 uroboro_db_path: str = None,
                 chroma_db_path: str = None,
                 ollama_url: str = "http://localhost:11434",
                 embed_model: str = "nomic-embed-text"):
        """Initialize the ChromaDB integration.

        Args:
            uroboro_db_path: Path to uroboro SQLite database
            chroma_db_path: Path to ChromaDB storage directory
            ollama_url: URL for Ollama API
            embed_model: Embedding model to use
        """
        # Set up paths
        home_dir = os.path.expanduser("~")
        self.uroboro_db_path = uroboro_db_path or os.path.join(
            home_dir, ".local/share/uroboro/uroboro.sqlite"
        )
        self.chroma_db_path = chroma_db_path or os.path.join(
            home_dir, ".local/share/uroboro/chromadb"
        )

        # Ollama configuration
        self.ollama_url = ollama_url
        self.embed_model = embed_model

        # Set up logging
        self.logger = self._setup_logging()

        # Initialize ChromaDB
        self.chroma_client = None
        self.collection = None
        self._init_chromadb()

    def _setup_logging(self) -> logging.Logger:
        """Set up logging configuration."""
        logger = logging.getLogger("uroboro_chromadb")
        handler = logging.StreamHandler()
        formatter = logging.Formatter(
            '%(asctime)s - %(name)s - %(levelname)s - %(message)s'
        )
        handler.setFormatter(formatter)
        logger.addHandler(handler)
        logger.setLevel(logging.INFO)
        return logger

    def _init_chromadb(self):
        """Initialize ChromaDB client and collection."""
        try:
            # Create ChromaDB directory if it doesn't exist
            os.makedirs(self.chroma_db_path, exist_ok=True)

            # Initialize ChromaDB client with persistent storage
            self.chroma_client = chromadb.PersistentClient(
                path=self.chroma_db_path,
                settings=Settings(
                    anonymized_telemetry=False,
                    allow_reset=True
                )
            )

            # Get or create collection for uroboro captures
            self.collection = self.chroma_client.get_or_create_collection(
                name="uroboro_captures",
                metadata={"description": "uroboro captures with semantic embeddings"}
            )

            self.logger.info(f"ChromaDB initialized at {self.chroma_db_path}")

        except Exception as e:
            self.logger.error(f"Failed to initialize ChromaDB: {e}")
            raise

    def get_ollama_embedding(self, text: str) -> List[float]:
        """Get embedding for text using Ollama.

        Args:
            text: Text to embed

        Returns:
            List of float values representing the embedding
        """
        try:
            response = requests.post(
                f"{self.ollama_url}/api/embed",
                json={
                    "model": self.embed_model,
                    "input": text
                },
                timeout=30
            )
            response.raise_for_status()

            data = response.json()
            if "embeddings" in data and len(data["embeddings"]) > 0:
                return data["embeddings"][0]
            else:
                raise ValueError("No embeddings returned from Ollama")

        except Exception as e:
            self.logger.error(f"Failed to get embedding from Ollama: {e}")
            raise

    def get_captures_from_uroboro(self) -> List[Dict[str, Any]]:
        """Get all captures from uroboro SQLite database.

        Returns:
            List of capture dictionaries
        """
        try:
            conn = sqlite3.connect(self.uroboro_db_path)
            conn.row_factory = sqlite3.Row
            cursor = conn.cursor()

            # Query all captures
            cursor.execute("""
                SELECT id, content, created_at, tags, project
                FROM captures
                ORDER BY created_at DESC
            """)

            captures = []
            for row in cursor.fetchall():
                captures.append({
                    "id": row["id"],
                    "content": row["content"],
                    "created_at": row["created_at"],
                    "tags": row["tags"] or "",
                    "project": row["project"] or ""
                })

            conn.close()
            self.logger.info(f"Retrieved {len(captures)} captures from uroboro database")
            return captures

        except Exception as e:
            self.logger.error(f"Failed to get captures from uroboro: {e}")
            raise

    def embed_capture(self, capture: Dict[str, Any]) -> bool:
        """Embed a single capture in ChromaDB.

        Args:
            capture: Capture dictionary with id, content, etc.

        Returns:
            True if successful, False otherwise
        """
        try:
            # Get embedding for the content
            embedding = self.get_ollama_embedding(capture["content"])

            # Prepare metadata
            metadata = {
                "created_at": capture["created_at"],
                "tags": capture["tags"],
                "project": capture["project"],
                "content_length": len(capture["content"])
            }

            # Add to ChromaDB collection
            self.collection.add(
                ids=[str(capture["id"])],
                embeddings=[embedding],
                documents=[capture["content"]],
                metadatas=[metadata]
            )

            return True

        except Exception as e:
            self.logger.error(f"Failed to embed capture {capture['id']}: {e}")
            return False

    def embed_all_captures(self, force_reembed: bool = False) -> Dict[str, int]:
        """Embed all captures from uroboro database.

        Args:
            force_reembed: If True, re-embed all captures even if they exist

        Returns:
            Dictionary with embedding statistics
        """
        captures = self.get_captures_from_uroboro()

        if not captures:
            self.logger.warning("No captures found in uroboro database")
            return {"total": 0, "embedded": 0, "skipped": 0, "failed": 0}

        stats = {"total": len(captures), "embedded": 0, "skipped": 0, "failed": 0}

        # Get existing embeddings if not force re-embedding
        existing_ids = set()
        if not force_reembed:
            try:
                result = self.collection.get()
                existing_ids = set(result["ids"])
                self.logger.info(f"Found {len(existing_ids)} existing embeddings")
            except Exception as e:
                self.logger.warning(f"Could not check existing embeddings: {e}")

        for i, capture in enumerate(captures):
            capture_id = str(capture["id"])

            # Skip if already embedded (unless force re-embedding)
            if not force_reembed and capture_id in existing_ids:
                stats["skipped"] += 1
                continue

            # Embed the capture
            if self.embed_capture(capture):
                stats["embedded"] += 1
                self.logger.info(f"Embedded capture {capture_id} ({i+1}/{len(captures)})")
            else:
                stats["failed"] += 1

            # Progress indicator
            if (i + 1) % 10 == 0:
                self.logger.info(f"Progress: {i+1}/{len(captures)} processed")

        self.logger.info(f"Embedding complete: {stats}")
        return stats

    def semantic_search(self, query: str, limit: int = 10) -> List[Dict[str, Any]]:
        """Perform semantic search across embedded captures.

        Args:
            query: Search query
            limit: Maximum number of results to return

        Returns:
            List of search results with similarity scores
        """
        try:
            # Get embedding for query
            query_embedding = self.get_ollama_embedding(query)

            # Search in ChromaDB
            results = self.collection.query(
                query_embeddings=[query_embedding],
                n_results=limit,
                include=["documents", "metadatas", "distances"]
            )

            # Format results
            search_results = []
            for i in range(len(results["ids"][0])):
                result = {
                    "capture_id": int(results["ids"][0][i]),
                    "content": results["documents"][0][i],
                    "similarity": 1 - results["distances"][0][i],  # Convert distance to similarity
                    "distance": results["distances"][0][i],
                    "metadata": results["metadatas"][0][i]
                }
                search_results.append(result)

            self.logger.info(f"Found {len(search_results)} results for query: '{query}'")
            return search_results

        except Exception as e:
            self.logger.error(f"Semantic search failed: {e}")
            raise

    def get_stats(self) -> Dict[str, Any]:
        """Get statistics about the ChromaDB collection.

        Returns:
            Dictionary with collection statistics
        """
        try:
            # Get collection info
            collection_info = self.collection.get()

            # Get uroboro database stats
            conn = sqlite3.connect(self.uroboro_db_path)
            cursor = conn.cursor()
            cursor.execute("SELECT COUNT(*) FROM captures")
            total_captures = cursor.fetchone()[0]
            conn.close()

            embedded_count = len(collection_info["ids"])
            coverage = (embedded_count / total_captures * 100) if total_captures > 0 else 0

            stats = {
                "total_captures": total_captures,
                "embedded_captures": embedded_count,
                "coverage_percent": round(coverage, 1),
                "chromadb_path": self.chroma_db_path,
                "uroboro_db_path": self.uroboro_db_path,
                "embed_model": self.embed_model,
                "collection_name": self.collection.name
            }

            return stats

        except Exception as e:
            self.logger.error(f"Failed to get stats: {e}")
            raise

    def test_connection(self) -> Dict[str, bool]:
        """Test connections to ChromaDB and Ollama.

        Returns:
            Dictionary with connection test results
        """
        results = {
            "chromadb": False,
            "ollama": False,
            "uroboro_db": False
        }

        # Test ChromaDB
        try:
            self.collection.count()
            results["chromadb"] = True
        except Exception as e:
            self.logger.error(f"ChromaDB connection failed: {e}")

        # Test Ollama
        try:
            response = requests.get(f"{self.ollama_url}/api/tags", timeout=5)
            results["ollama"] = response.status_code == 200
        except Exception as e:
            self.logger.error(f"Ollama connection failed: {e}")

        # Test uroboro database
        try:
            conn = sqlite3.connect(self.uroboro_db_path)
            cursor = conn.cursor()
            cursor.execute("SELECT COUNT(*) FROM captures")
            conn.close()
            results["uroboro_db"] = True
        except Exception as e:
            self.logger.error(f"Uroboro database connection failed: {e}")

        return results

    def reset_collection(self):
        """Reset (delete) the ChromaDB collection."""
        try:
            self.chroma_client.delete_collection("uroboro_captures")
            self.collection = self.chroma_client.create_collection(
                name="uroboro_captures",
                metadata={"description": "uroboro captures with semantic embeddings"}
            )
            self.logger.info("ChromaDB collection reset successfully")
        except Exception as e:
            self.logger.error(f"Failed to reset collection: {e}")
            raise


def main():
    """Command-line interface for ChromaDB integration."""
    parser = argparse.ArgumentParser(description="uroboro ChromaDB Integration")
    subparsers = parser.add_subparsers(dest="command", help="Available commands")

    # Embed command
    embed_parser = subparsers.add_parser("embed", help="Embed captures")
    embed_parser.add_argument("--force", action="store_true",
                             help="Force re-embedding of all captures")

    # Search command
    search_parser = subparsers.add_parser("search", help="Semantic search")
    search_parser.add_argument("query", help="Search query")
    search_parser.add_argument("--limit", type=int, default=10,
                              help="Maximum results to return")

    # Stats command
    subparsers.add_parser("stats", help="Show statistics")

    # Test command
    subparsers.add_parser("test", help="Test connections")

    # Reset command
    subparsers.add_parser("reset", help="Reset ChromaDB collection")

    args = parser.parse_args()

    if not args.command:
        parser.print_help()
        return

    # Initialize integration
    try:
        integration = UroboroChromaDBIntegration()
    except Exception as e:
        print(f"❌ Failed to initialize ChromaDB integration: {e}")
        return

    # Execute command
    try:
        if args.command == "embed":
            print("🔄 Embedding captures...")
            stats = integration.embed_all_captures(force_reembed=args.force)
            print(f"✅ Embedding complete!")
            print(f"   Total: {stats['total']}")
            print(f"   Embedded: {stats['embedded']}")
            print(f"   Skipped: {stats['skipped']}")
            print(f"   Failed: {stats['failed']}")

        elif args.command == "search":
            print(f"🔍 Searching for: '{args.query}'")
            results = integration.semantic_search(args.query, args.limit)

            if not results:
                print("No results found")
                return

            print(f"\nFound {len(results)} results:\n")
            for i, result in enumerate(results):
                similarity_percent = result["similarity"] * 100
                print(f"🎯 Result {i+1} ({similarity_percent:.1f}% similarity)")
                print(f"   ID: {result['capture_id']}")
                print(f"   Created: {result['metadata'].get('created_at', 'N/A')}")
                if result['metadata'].get('project'):
                    print(f"   Project: {result['metadata']['project']}")
                if result['metadata'].get('tags'):
                    print(f"   Tags: {result['metadata']['tags']}")

                # Truncate long content
                content = result["content"]
                if len(content) > 200:
                    content = content[:200] + "..."
                print(f"   {content}\n")

        elif args.command == "stats":
            print("📊 ChromaDB Statistics")
            stats = integration.get_stats()
            print(f"   Total Captures: {stats['total_captures']}")
            print(f"   Embedded: {stats['embedded_captures']}")
            print(f"   Coverage: {stats['coverage_percent']}%")
            print(f"   Model: {stats['embed_model']}")
            print(f"   ChromaDB Path: {stats['chromadb_path']}")

        elif args.command == "test":
            print("🔧 Testing connections...")
            results = integration.test_connection()

            status_map = {True: "✅ Connected", False: "❌ Failed"}
            print(f"   ChromaDB: {status_map[results['chromadb']]}")
            print(f"   Ollama: {status_map[results['ollama']]}")
            print(f"   Uroboro DB: {status_map[results['uroboro_db']]}")

            if all(results.values()):
                print("\n🎉 All connections working!")
            else:
                print("\n⚠️  Some connections failed. Check configuration.")

        elif args.command == "reset":
            confirm = input("⚠️  This will delete all embeddings. Continue? (y/N): ")
            if confirm.lower() == 'y':
                integration.reset_collection()
                print("✅ Collection reset successfully")
            else:
                print("Operation cancelled")

    except Exception as e:
        print(f"❌ Command failed: {e}")


if __name__ == "__main__":
    main()
