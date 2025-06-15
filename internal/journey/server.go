package journey

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/QRY91/uroboro/internal/database"
)

// Server handles HTTP requests for journey visualization
type Server struct {
	service *JourneyService
	port    int
}

// NewServer creates a new journey visualization server
func NewServer(db *database.DB, port int) *Server {
	return &Server{
		service: NewJourneyService(db),
		port:    port,
	}
}

// Start begins serving the journey visualization
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// API endpoints with CORS support
	mux.HandleFunc("/api/journey", s.withCORS(s.handleJourneyAPI))
	mux.HandleFunc("/api/health", s.withCORS(s.handleHealth))

	// Serve Svelte app assets
	mux.HandleFunc("/assets/", s.handleSvelteAssets)
	mux.HandleFunc("/src/", s.handleSvelteAssets)

	// Serve static files (manifest, icons, etc.)
	mux.HandleFunc("/manifest.json", s.handleStaticFile)
	mux.HandleFunc("/icons/", s.handleStaticFile)
	mux.HandleFunc("/screenshots/", s.handleStaticFile)

	// Main Svelte app (catch-all for SPA routing)
	mux.HandleFunc("/", s.handleSvelteApp)

	addr := fmt.Sprintf(":%d", s.port)
	fmt.Printf("🚀 Journey visualization server starting on http://localhost%s\n", addr)
	fmt.Printf("   📊 Timeline interface: http://localhost%s\n", addr)
	fmt.Printf("   🔗 API endpoint: http://localhost%s/api/journey\n", addr)

	return http.ListenAndServe(addr, mux)
}

// handleSvelteApp serves the main Svelte application
func (s *Server) handleSvelteApp(w http.ResponseWriter, r *http.Request) {
	// Try to serve built Svelte app first
	distPath := filepath.Join("web", "dist", "index.html")
	if _, err := os.Stat(distPath); err == nil {
		s.serveFile(w, r, distPath)
		return
	}

	// Fallback to development server notice if dist doesn't exist
	if r.URL.Path == "/" {
		s.handleDevFallback(w, r)
		return
	}

	// For SPA routing, always serve index.html for non-API routes
	if !strings.HasPrefix(r.URL.Path, "/api/") {
		s.handleDevFallback(w, r)
		return
	}

	http.NotFound(w, r)
}

// handleDevFallback serves a development notice when Svelte app isn't built
func (s *Server) handleDevFallback(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.New("dev").Parse(devFallbackTemplate))

	data := struct {
		Title string
		Port  int
	}{
		Title: "Uroboro Journey Timeline",
		Port:  s.port,
	}

	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleJourneyAPI serves journey data as JSON
func (s *Server) handleJourneyAPI(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	options := JourneyOptions{
		Days:     7, // default
		Export:   false,
		Live:     false,
		Port:     s.port,
		AutoOpen: false,
		Share:    false,
		Title:    "My Journey",
		Theme:    "default",
	}

	// Parse parameters
	if daysStr := r.URL.Query().Get("days"); daysStr != "" {
		if days, err := strconv.Atoi(daysStr); err == nil {
			options.Days = days
		}
	}

	if r.URL.Query().Get("live") == "true" {
		options.Live = true
	}

	if theme := r.URL.Query().Get("theme"); theme != "" {
		options.Theme = theme
	}

	if title := r.URL.Query().Get("title"); title != "" {
		options.Title = title
	}

	// Generate journey data
	journey, err := s.service.GenerateJourney(options)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if err := json.NewEncoder(w).Encode(journey); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleHealth provides a health check endpoint
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   "1.0.0",
	})
}

// withCORS adds CORS headers to responses
func (s *Server) withCORS(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		handler(w, r)
	}
}

// handleSvelteAssets serves Svelte development assets
func (s *Server) handleSvelteAssets(w http.ResponseWriter, r *http.Request) {
	// In development, proxy to Vite dev server on port 3000
	// In production, serve from web/dist directory
	distPath := filepath.Join("web", "dist", strings.TrimPrefix(r.URL.Path, "/"))

	if _, err := os.Stat(distPath); err == nil {
		s.serveFile(w, r, distPath)
		return
	}

	// Development mode notice
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, "Asset not found. Run 'npm run dev' in the web directory for development mode.")
}

// handleStaticFile serves static files like manifest.json, icons, etc.
func (s *Server) handleStaticFile(w http.ResponseWriter, r *http.Request) {
	// Remove leading slash and serve from web/public
	filePath := filepath.Join("web", "public", strings.TrimPrefix(r.URL.Path, "/"))
	s.serveFile(w, r, filePath)
}

// serveFile serves a file with appropriate content type
func (s *Server) serveFile(w http.ResponseWriter, r *http.Request, filePath string) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}

	// Set content type based on file extension
	ext := filepath.Ext(filePath)
	switch ext {
	case ".html":
		w.Header().Set("Content-Type", "text/html")
	case ".css":
		w.Header().Set("Content-Type", "text/css")
	case ".js":
		w.Header().Set("Content-Type", "application/javascript")
	case ".json":
		w.Header().Set("Content-Type", "application/json")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".svg":
		w.Header().Set("Content-Type", "image/svg+xml")
	case ".ico":
		w.Header().Set("Content-Type", "image/x-icon")
	default:
		w.Header().Set("Content-Type", "application/octet-stream")
	}

	// Serve the file
	http.ServeFile(w, r, filePath)
}

// HTML template for the main visualization page
const devFallbackTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}} - Development Mode</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            margin: 0;
            padding: 2rem;
            background: #0a0a0a;
            color: #ffffff;
        }
        .container {
            max-width: 800px;
            margin: 0 auto;
        }
        h1 {
            color: #4ecdc4;
            margin-bottom: 1rem;
        }
        h2 {
            color: #ff6b6b;
            margin-top: 2rem;
            margin-bottom: 1rem;
        }
        .code {
            background: #1a1a1a;
            padding: 1rem;
            border-radius: 4px;
            font-family: 'Monaco', 'Menlo', monospace;
            border: 1px solid #333;
            margin: 1rem 0;
        }
        .api-link {
            color: #4ecdc4;
            text-decoration: none;
        }
        .api-link:hover {
            text-decoration: underline;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🐍 {{.Title}}</h1>
        <p>The modern Svelte-based timeline interface is not yet built.</p>

        <h2>Development Setup</h2>
        <p>To run the new timeline interface:</p>
        <div class="code">
cd web<br>
npm install<br>
npm run dev
        </div>

        <p>The development server will run on <strong>http://localhost:3000</strong></p>

        <h2>API Access</h2>
        <p>Journey data is available at:
           <a href="/api/journey" class="api-link">http://localhost:{{.Port}}/api/journey</a>
        </p>

        <h2>Production Build</h2>
        <p>To build for production:</p>
        <div class="code">
cd web<br>
npm run build
        </div>

        <p>After building, refresh this page to see the timeline interface.</p>
    </div>
</body>
</html>`

// 🧛‍♂️⚰️ THE EMBEDDED TEMPLATE VAMPIRES HAVE BEEN SLAIN! ⚰️🧛‍♂️
// No more 1000+ lines of embedded CSS/JS horror!
// The Svelte app now reigns supreme! 👑
