package journey

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
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

	// Main visualization page
	mux.HandleFunc("/", s.handleIndex)

	// API endpoints
	mux.HandleFunc("/api/journey", s.handleJourneyAPI)
	mux.HandleFunc("/api/health", s.handleHealth)

	// Static assets
	mux.HandleFunc("/static/", s.handleStatic)

	addr := fmt.Sprintf(":%d", s.port)
	fmt.Printf("🚀 Journey visualization server starting on http://localhost%s\n", addr)

	return http.ListenAndServe(addr, mux)
}

// handleIndex serves the main visualization page
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	tmpl := template.Must(template.New("index").Parse(indexTemplate))

	data := struct {
		Title string
		Port  int
	}{
		Title: "Journey Replay",
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
		Theme:    ThemeDefault,
	}

	// Parse parameters
	if daysStr := r.URL.Query().Get("days"); daysStr != "" {
		if days, err := strconv.Atoi(daysStr); err == nil {
			options.Days = days
		}
	}

	if projects := r.URL.Query().Get("projects"); projects != "" {
		options.Projects = strings.Split(projects, ",")
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
		http.Error(w, fmt.Sprintf("Failed to generate journey: %v", err), http.StatusInternalServerError)
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

// handleStatic serves static assets (CSS, JS, images)
func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[8:] // Remove "/static/" prefix

	switch {
	case strings.HasSuffix(path, ".css"):
		w.Header().Set("Content-Type", "text/css")
		fmt.Fprint(w, journeyCSS)
	case strings.HasSuffix(path, ".js"):
		w.Header().Set("Content-Type", "application/javascript")
		fmt.Fprint(w, journeyJS)
	default:
		http.NotFound(w, r)
	}
}

// HTML template for the main visualization page
const indexTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}} - Journey Replay</title>
    <link rel="stylesheet" href="/static/journey.css">
</head>
<body>
    <div id="app">
        <header class="header">
            <h1>{{.Title}}</h1>
            <div class="controls">
                <div class="playback-controls">
                    <button id="playBtn" class="btn-primary">▶ Play</button>
                    <button id="pauseBtn" class="btn-secondary">⏸ Pause</button>
                    <button id="restartBtn" class="btn-secondary">⏮ Restart</button>
                </div>
                <div class="timeline-progress">
                    <div class="progress-bar">
                        <div id="progressFill" class="progress-fill"></div>
                        <input type="range" id="progressSlider" class="progress-slider" min="0" max="100" value="0">
                    </div>
                    <div class="progress-time">
                        <span id="currentTime">0:00</span> / <span id="totalTime">0:00</span>
                    </div>
                </div>
                <div class="speed-control">
                    <label>Speed:</label>
                    <input type="range" id="speedSlider" min="0.5" max="3" step="0.5" value="1">
                    <span id="speedValue">1x</span>
                </div>
            </div>
        </header>

        <main class="main-content">
            <div class="timeline-container">
                <div class="timeline-header">
                    <h2 id="timelineTitle">Development Journey</h2>
                    <div class="timeline-info">
                        <span id="eventCounter">0 / 0 events</span>
                        <span id="currentDate">Select timeline to begin</span>
                    </div>
                </div>
                <div id="timeline" class="timeline">
                    <div class="loading">Loading journey data...</div>
                </div>
                <div class="timeline-legend">
                    <div class="legend-item">
                        <div class="legend-dot milestone"></div>
                        <span>Milestones</span>
                    </div>
                    <div class="legend-item">
                        <div class="legend-dot learning"></div>
                        <span>Learning</span>
                    </div>
                    <div class="legend-item">
                        <div class="legend-dot decision"></div>
                        <span>Decisions</span>
                    </div>
                    <div class="legend-item">
                        <div class="legend-dot commit"></div>
                        <span>Commits</span>
                    </div>
                    <div class="legend-item">
                        <div class="legend-dot capture"></div>
                        <span>Captures</span>
                    </div>
                </div>
            </div>

            <aside class="sidebar">
                <div class="filters">
                    <h3>Filters</h3>
                    <div class="filter-group">
                        <label>Days:</label>
                        <select id="daysFilter">
                            <option value="1">Last 1 day</option>
                            <option value="3">Last 3 days</option>
                            <option value="7" selected>Last 7 days</option>
                            <option value="14">Last 14 days</option>
                            <option value="30">Last 30 days</option>
                        </select>
                    </div>
                    <div class="filter-group">
                        <label>Theme:</label>
                        <select id="themeFilter">
                            <option value="default">Default</option>
                            <option value="dark">Dark</option>
                            <option value="light">Light</option>
                            <option value="matrix">Matrix</option>
                            <option value="neon">Neon</option>
                        </select>
                    </div>
                </div>

                <div class="stats">
                    <h3>Journey Stats</h3>
                    <div id="journeyStats" class="stats-container">
                        <div class="stat-item">
                            <span class="stat-label">Total Events:</span>
                            <span class="stat-value" id="totalEvents">-</span>
                        </div>
                        <div class="stat-item">
                            <span class="stat-label">Projects:</span>
                            <span class="stat-value" id="projectCount">-</span>
                        </div>
                        <div class="stat-item">
                            <span class="stat-label">Milestones:</span>
                            <span class="stat-value" id="milestoneCount">-</span>
                        </div>
                        <div class="stat-item">
                            <span class="stat-label">Learning Moments:</span>
                            <span class="stat-value" id="learningMoments">-</span>
                        </div>
                    </div>
                </div>

                <div class="projects">
                    <h3>Projects</h3>
                    <div id="projectList" class="project-list">
                        <!-- Projects will be populated by JavaScript -->
                    </div>
                </div>
            </aside>
        </main>

        <div id="eventDetail" class="event-detail hidden">
            <div class="event-detail-content">
                <button class="close-btn" id="closeEventDetail">&times;</button>
                <h3 id="eventTitle"></h3>
                <div class="event-meta">
                    <span id="eventTime"></span>
                    <span id="eventProject"></span>
                    <span id="eventType"></span>
                </div>
                <div id="eventDescription"></div>
                <div id="eventTags" class="event-tags"></div>
            </div>
        </div>
    </div>

    <script src="/static/journey.js"></script>
</body>
</html>`

// CSS styles for the journey visualization
const journeyCSS = `
* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}

body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    background: #0a0a0a;
    color: #ffffff;
    overflow-x: hidden;
}

.header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem 2rem;
    background: rgba(20, 20, 20, 0.9);
    backdrop-filter: blur(10px);
    border-bottom: 1px solid #333;
}

.header h1 {
    font-size: 1.5rem;
    font-weight: 600;
    background: linear-gradient(45deg, #ff6b6b, #4ecdc4);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
}

.controls {
    display: flex;
    align-items: center;
    gap: 2rem;
    flex-wrap: wrap;
}

.playback-controls {
    display: flex;
    align-items: center;
    gap: 0.5rem;
}

.timeline-progress {
    display: flex;
    align-items: center;
    gap: 1rem;
    flex: 1;
    min-width: 200px;
}

.progress-bar {
    position: relative;
    flex: 1;
    height: 6px;
    background: #333;
    border-radius: 3px;
    overflow: hidden;
}

.progress-fill {
    height: 100%;
    background: linear-gradient(90deg, #4ecdc4, #45b7d1);
    border-radius: 3px;
    width: 0%;
    transition: width 0.1s ease;
}

.progress-slider {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    opacity: 0;
    cursor: pointer;
}

.progress-time {
    font-size: 0.8rem;
    color: #ccc;
    white-space: nowrap;
}

.timeline-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
    padding: 0 2rem;
}

.timeline-header h2 {
    margin: 0;
    color: #4ecdc4;
    font-size: 1.3rem;
}

.timeline-info {
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    gap: 0.2rem;
    font-size: 0.9rem;
    color: #ccc;
}

.timeline-legend {
    display: flex;
    justify-content: center;
    gap: 2rem;
    margin-top: 1rem;
    padding: 1rem;
    background: rgba(20, 20, 20, 0.5);
    border-top: 1px solid #333;
}

.legend-item {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.8rem;
    color: #ccc;
}

.legend-dot {
    width: 12px;
    height: 12px;
    border-radius: 50%;
}

.legend-dot.milestone {
    border: 2px solid #ff6b6b;
    background: transparent;
}

.legend-dot.learning {
    background: #45b7d1;
}

.legend-dot.decision {
    background: #feca57;
    border-radius: 2px;
}

.legend-dot.commit {
    background: #96ceb4;
    border-radius: 1px;
}

.legend-dot.capture {
    background: #4ecdc4;
}

.btn-primary, .btn-secondary {
    padding: 0.5rem 1rem;
    border: none;
    border-radius: 6px;
    cursor: pointer;
    font-weight: 500;
    transition: all 0.2s ease;
}

.btn-primary {
    background: #4ecdc4;
    color: #0a0a0a;
}

.btn-primary:hover {
    background: #45b7d1;
    transform: translateY(-1px);
}

.btn-secondary {
    background: #333;
    color: #fff;
}

.btn-secondary:hover {
    background: #444;
}

.speed-control {
    display: flex;
    align-items: center;
    gap: 0.5rem;
}

.speed-control label {
    font-size: 0.9rem;
}

#speedSlider {
    width: 80px;
}

.main-content {
    display: flex;
    height: calc(100vh - 140px);
}

.timeline-container {
    flex: 1;
    position: relative;
    overflow: hidden;
    background: radial-gradient(circle at 50% 50%, #1a1a2e 0%, #0a0a0a 100%);
    display: flex;
    flex-direction: column;
}

.timeline {
    position: relative;
    width: 100%;
    flex: 1;
    padding: 2rem;
    overflow: hidden;
}

.sidebar {
    width: 300px;
    background: rgba(20, 20, 20, 0.9);
    border-left: 1px solid #333;
    padding: 1.5rem;
    overflow-y: auto;
}

.sidebar h3 {
    margin-bottom: 1rem;
    color: #4ecdc4;
    font-size: 1.1rem;
}

.filter-group {
    margin-bottom: 1rem;
}

.filter-group label {
    display: block;
    margin-bottom: 0.3rem;
    font-size: 0.9rem;
    color: #ccc;
}

.filter-group select {
    width: 100%;
    padding: 0.5rem;
    background: #333;
    color: #fff;
    border: 1px solid #444;
    border-radius: 4px;
}

.stats, .projects {
    margin-top: 2rem;
}

.stat-item {
    display: flex;
    justify-content: space-between;
    margin-bottom: 0.5rem;
    padding: 0.3rem 0;
    border-bottom: 1px solid #333;
}

.stat-label {
    color: #ccc;
    font-size: 0.9rem;
}

.stat-value {
    color: #4ecdc4;
    font-weight: 600;
}

.project-list {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
}

.project-item {
    padding: 0.5rem;
    background: #333;
    border-radius: 4px;
    border-left: 4px solid;
    cursor: pointer;
    transition: all 0.2s ease;
}

.project-item:hover {
    background: #444;
    transform: translateX(2px);
}

.project-name {
    font-weight: 500;
    margin-bottom: 0.2rem;
}

.project-count {
    font-size: 0.8rem;
    color: #ccc;
}

.event-detail {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background: rgba(0, 0, 0, 0.8);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 1000;
}

.event-detail.hidden {
    display: none;
}

.event-detail-content {
    background: #1a1a1a;
    border-radius: 8px;
    padding: 2rem;
    max-width: 600px;
    width: 90%;
    max-height: 80%;
    overflow-y: auto;
    position: relative;
}

.close-btn {
    position: absolute;
    top: 1rem;
    right: 1rem;
    background: none;
    border: none;
    color: #ccc;
    font-size: 1.5rem;
    cursor: pointer;
}

.close-btn:hover {
    color: #fff;
}

.event-meta {
    display: flex;
    gap: 1rem;
    margin: 1rem 0;
    font-size: 0.9rem;
    color: #ccc;
}

.event-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    margin-top: 1rem;
}

.tag {
    padding: 0.2rem 0.5rem;
    background: #333;
    border-radius: 12px;
    font-size: 0.8rem;
    color: #4ecdc4;
}

.timeline-event {
    position: absolute;
    width: 14px;
    height: 14px;
    border-radius: 50%;
    cursor: pointer;
    transition: all 0.3s ease;
    animation: pulse 2s infinite;
    opacity: 0.3;
    transform: scale(0.5);
    z-index: 1;
}

.timeline-event:hover {
    transform: scale(1.5) !important;
    box-shadow: 0 0 20px var(--project-color, rgba(78, 205, 196, 0.5));
    z-index: 1000 !important;
}

.timeline-event.revealed {
    opacity: 1;
    transform: scale(1);
}

.timeline-event.milestone {
    width: 18px;
    height: 18px;
    border: 3px solid #ff6b6b;
    background: transparent;
}

.timeline-event.learning {
    background: #45b7d1;
    border-radius: 50%;
}

.timeline-event.decision {
    background: #feca57;
    border-radius: 4px;
}

.timeline-event.integration {
    background: #ff9ff3;
    transform: rotate(45deg);
}

.timeline-event.commit {
    background: #96ceb4;
    border-radius: 2px;
}

.event-preview {
    position: absolute;
    bottom: 25px;
    left: 50%;
    transform: translateX(-50%);
    background: rgba(0, 0, 0, 0.9);
    color: white;
    padding: 4px 8px;
    border-radius: 4px;
    font-size: 0.7rem;
    white-space: nowrap;
    max-width: 200px;
    overflow: hidden;
    text-overflow: ellipsis;
    opacity: 0;
    display: none;
    transition: opacity 0.3s ease;
    z-index: 1001;
}

.timeline-axis {
    box-shadow: 0 0 10px rgba(78, 205, 196, 0.3);
}

/* Enhanced Animations */
@keyframes milestone-entrance {
    0% { transform: scale(0) rotate(0deg); opacity: 0; }
    50% { transform: scale(1.5) rotate(180deg); opacity: 0.8; }
    100% { transform: scale(1) rotate(360deg); opacity: 1; }
}

@keyframes learning-bounce {
    0% { transform: scale(0) translateY(20px); opacity: 0; }
    60% { transform: scale(1.2) translateY(-5px); opacity: 0.9; }
    100% { transform: scale(1) translateY(0); opacity: 1; }
}

@keyframes decision-flash {
    0% { transform: scale(0); opacity: 0; background: #fff; }
    50% { transform: scale(1.3); opacity: 0.8; background: #feca57; }
    100% { transform: scale(1); opacity: 1; background: #feca57; }
}

@keyframes default-entrance {
    0% { transform: scale(0); opacity: 0; }
    70% { transform: scale(1.1); opacity: 0.9; }
    100% { transform: scale(1); opacity: 1; }
}

@keyframes milestone-pulse {
    0%, 100% { box-shadow: 0 0 5px #ff6b6b; }
    50% { box-shadow: 0 0 20px #ff6b6b, 0 0 30px #ff6b6b40; }
}

.loading {
    display: flex;
    justify-content: center;
    align-items: center;
    height: 100%;
    font-size: 1.2rem;
    color: #4ecdc4;
}

@keyframes pulse {
    0% { opacity: 0.6; }
    50% { opacity: 1; }
    100% { opacity: 0.6; }
}

@keyframes slideIn {
    from {
        opacity: 0;
        transform: translateY(20px);
    }
    to {
        opacity: 1;
        transform: translateY(0);
    }
}

.timeline-event {
    animation: slideIn 0.5s ease-out;
}
`

// JavaScript for the journey visualization interface
const journeyJS = `
class JourneyVisualization {
    constructor() {
        this.journeyData = null;
        this.isPlaying = false;
        this.currentEventIndex = 0;
        this.playSpeed = 1;
        this.animationFrame = null;

        this.initializeEventListeners();
        this.loadJourneyData();
    }

    initializeEventListeners() {
        document.getElementById('playBtn').addEventListener('click', () => this.play());
        document.getElementById('pauseBtn').addEventListener('click', () => this.pause());
        document.getElementById('restartBtn').addEventListener('click', () => this.restart());

        const speedSlider = document.getElementById('speedSlider');
        speedSlider.addEventListener('input', (e) => {
            this.playSpeed = parseFloat(e.target.value);
            document.getElementById('speedValue').textContent = this.playSpeed + 'x';
        });

        document.getElementById('daysFilter').addEventListener('change', (e) => {
            this.loadJourneyData({ days: parseInt(e.target.value) });
        });

        document.getElementById('themeFilter').addEventListener('change', (e) => {
            this.loadJourneyData({ theme: e.target.value });
        });

        document.getElementById('closeEventDetail').addEventListener('click', () => {
            document.getElementById('eventDetail').classList.add('hidden');
        });

        // Progress slider for timeline scrubbing
        document.getElementById('progressSlider').addEventListener('input', (e) => {
            this.scrubToPosition(parseInt(e.target.value));
        });
    }

    async loadJourneyData(params = {}) {
        try {
            const urlParams = new URLSearchParams();
            if (params.days) urlParams.set('days', params.days);
            if (params.theme) urlParams.set('theme', params.theme);

            const response = await fetch('/api/journey?' + urlParams.toString());
            this.journeyData = await response.json();

            this.renderTimeline();
            this.updateStats();
            this.updateProjects();
            this.updateTimelineInfo();
        } catch (error) {
            console.error('Failed to load journey data:', error);
            document.querySelector('.loading').textContent = 'Failed to load journey data';
        }
    }

    renderTimeline() {
        const timeline = document.getElementById('timeline');
        timeline.innerHTML = '';

        if (!this.journeyData || !this.journeyData.events.length) {
            timeline.innerHTML = '<div class="loading">No events found</div>';
            return;
        }

        // Create main timeline container with improved layout
        const container = document.createElement('div');
        container.className = 'timeline-events';
        container.style.position = 'relative';
        container.style.width = '100%';
        container.style.height = '100%';
        container.style.overflow = 'hidden';

        // Create timeline axis
        this.createTimelineAxis(container);

        // Create events with staggered positioning
        this.journeyData.events.forEach((event, index) => {
            const eventEl = this.createEventElement(event, index);
            eventEl.style.opacity = '0.3'; // Start hidden for animation
            eventEl.style.transform = 'scale(0.5)';
            container.appendChild(eventEl);
        });

        timeline.appendChild(container);

        // Initialize timeline state
        this.currentEventIndex = 0;
        this.resetTimeline();
    }

    createEventElement(event, index) {
        const eventEl = document.createElement('div');
        eventEl.className = 'timeline-event ' + event.eventType;
        eventEl.setAttribute('data-index', index);
        eventEl.setAttribute('data-timestamp', event.timestamp);

        // Enhanced positioning algorithm
        const timePercent = this.getTimePercent(event.timestamp);
        const lane = this.calculateEventLane(event, index);
        const yPosition = this.calculateYPosition(event, lane);

        eventEl.style.left = timePercent + '%';
        eventEl.style.top = yPosition + 'px';

        // Enhanced visual styling based on event type
        this.styleEventElement(eventEl, event);

        // Set color based on project with enhanced visuals
        const project = this.journeyData.projects.find(p => p.name === event.project);
        if (project) {
            eventEl.style.setProperty('--project-color', project.color);
            eventEl.style.background = project.color;
            eventEl.style.boxShadow = '0 0 10px ' + project.color + '40';
        }

        // Add hover effects
        eventEl.addEventListener('mouseenter', () => this.onEventHover(eventEl, event));
        eventEl.addEventListener('mouseleave', () => this.onEventLeave(eventEl));
        eventEl.addEventListener('click', () => this.showEventDetail(event));

        // Add content preview
        const preview = document.createElement('div');
        preview.className = 'event-preview';
        preview.textContent = event.content.substring(0, 50) + '...';
        eventEl.appendChild(preview);

        return eventEl;
    }

    getTimePercent(timestamp) {
        const start = new Date(this.journeyData.dateRange.start);
        const end = new Date(this.journeyData.dateRange.end);
        const eventTime = new Date(timestamp);

        const totalDuration = end - start;
        const eventDuration = eventTime - start;

        return Math.max(0, Math.min(100, (eventDuration / totalDuration) * 100));
    }

    showEventDetail(event) {
        document.getElementById('eventTitle').textContent = event.content;
        document.getElementById('eventTime').textContent = new Date(event.timestamp).toLocaleString();
        document.getElementById('eventProject').textContent = event.project || 'No project';
        document.getElementById('eventType').textContent = event.eventType;

        const tagsContainer = document.getElementById('eventTags');
        tagsContainer.innerHTML = '';
        event.tags.forEach(tag => {
            const tagEl = document.createElement('span');
            tagEl.className = 'tag';
            tagEl.textContent = tag;
            tagsContainer.appendChild(tagEl);
        });

        document.getElementById('eventDetail').classList.remove('hidden');
    }

    updateStats() {
        if (!this.journeyData) return;

        document.getElementById('totalEvents').textContent = this.journeyData.stats.totalEvents;
        document.getElementById('projectCount').textContent = this.journeyData.stats.projectCount;
        document.getElementById('milestoneCount').textContent = this.journeyData.stats.milestoneCount;
        document.getElementById('learningMoments').textContent = this.journeyData.stats.learningMoments;
    }

    updateProjects() {
        if (!this.journeyData) return;

        const projectList = document.getElementById('projectList');
        projectList.innerHTML = '';

        this.journeyData.projects.forEach(project => {
            const projectEl = document.createElement('div');
            projectEl.className = 'project-item';
            projectEl.style.borderLeftColor = project.color;

            projectEl.innerHTML =
                '<div class="project-name">' + project.name + '</div>' +
                '<div class="project-count">' + project.eventCount + ' events</div>';

            projectList.appendChild(projectEl);
        });
    }

    play() {
        if (!this.journeyData || !this.journeyData.events.length) return;

        this.isPlaying = true;
        this.startTime = Date.now();
        this.animate();

        // Update button states
        document.getElementById('playBtn').style.opacity = '0.5';
        document.getElementById('pauseBtn').style.opacity = '1';
    }

    pause() {
        this.isPlaying = false;
        if (this.animationFrame) {
            cancelAnimationFrame(this.animationFrame);
        }

        // Update button states
        document.getElementById('playBtn').style.opacity = '1';
        document.getElementById('pauseBtn').style.opacity = '0.5';
    }

    restart() {
        this.pause();
        this.currentEventIndex = 0;
        this.resetTimeline();

        // Update button states
        document.getElementById('playBtn').style.opacity = '1';
        document.getElementById('pauseBtn').style.opacity = '0.5';
    }

    animate() {
        if (!this.isPlaying || !this.journeyData) return;

        const eventInterval = 1000 / this.playSpeed; // Base interval
        const currentTime = Date.now();

        if (!this.lastEventTime) {
            this.lastEventTime = currentTime;
        }

        if (currentTime - this.lastEventTime >= eventInterval) {
            this.revealNextEvent();
            this.lastEventTime = currentTime;
        }

        if (this.currentEventIndex < this.journeyData.events.length) {
            this.animationFrame = requestAnimationFrame(() => this.animate());
        } else {
            this.pause();
        }
    }

    // New enhanced methods for Phase 2
    createTimelineAxis(container) {
        const axis = document.createElement('div');
        axis.className = 'timeline-axis';
        axis.style.position = 'absolute';
        axis.style.bottom = '50px';
        axis.style.left = '0';
        axis.style.right = '0';
        axis.style.height = '2px';
        axis.style.background = 'linear-gradient(90deg, #4ecdc4, #45b7d1)';
        axis.style.opacity = '0.6';
        container.appendChild(axis);

        // Add time markers
        for (let i = 0; i <= 10; i++) {
            const marker = document.createElement('div');
            marker.className = 'time-marker';
            marker.style.position = 'absolute';
            marker.style.left = (i * 10) + '%';
            marker.style.bottom = '45px';
            marker.style.width = '1px';
            marker.style.height = '10px';
            marker.style.background = '#4ecdc4';
            marker.style.opacity = '0.4';
            container.appendChild(marker);
        }
    }

    calculateEventLane(event, index) {
        // Smart lane assignment to prevent overlapping
        const timePercent = this.getTimePercent(event.timestamp);
        const laneWidth = 10; // % of timeline per lane
        return Math.floor(timePercent / laneWidth) % 4;
    }

    calculateYPosition(event, lane) {
        const baseY = 150;
        const laneSpacing = 80;
        const importanceOffset = (event.importance - 1) * 20;
        const randomOffset = (Math.random() - 0.5) * 30;

        return baseY + (lane * laneSpacing) + importanceOffset + randomOffset;
    }

    styleEventElement(eventEl, event) {
        // Enhanced styling based on event type
        switch (event.eventType) {
            case 'milestone':
                eventEl.style.width = '20px';
                eventEl.style.height = '20px';
                eventEl.style.border = '3px solid #ff6b6b';
                eventEl.style.animation = 'milestone-pulse 2s infinite';
                break;
            case 'learning':
                eventEl.style.background = '#45b7d1';
                eventEl.style.borderRadius = '50%';
                break;
            case 'decision':
                eventEl.style.background = '#feca57';
                eventEl.style.borderRadius = '4px';
                break;
            case 'integration':
                eventEl.style.background = '#ff9ff3';
                eventEl.style.transform = 'rotate(45deg)';
                break;
            case 'commit':
                eventEl.style.background = '#96ceb4';
                eventEl.style.borderRadius = '2px';
                break;
            default:
                eventEl.style.borderRadius = '50%';
        }
    }

    onEventHover(eventEl, event) {
        eventEl.style.transform = 'scale(1.5)';
        eventEl.style.zIndex = '1000';

        // Show content preview
        const preview = eventEl.querySelector('.event-preview');
        if (preview) {
            preview.style.display = 'block';
            preview.style.opacity = '1';
        }
    }

    onEventLeave(eventEl) {
        eventEl.style.transform = 'scale(1)';
        eventEl.style.zIndex = 'auto';

        // Hide content preview
        const preview = eventEl.querySelector('.event-preview');
        if (preview) {
            preview.style.display = 'none';
            preview.style.opacity = '0';
        }
    }

    resetTimeline() {
        // Reset all events to initial state
        document.querySelectorAll('.timeline-event').forEach(eventEl => {
            eventEl.style.opacity = '0.3';
            eventEl.style.transform = 'scale(0.5)';
            eventEl.classList.remove('revealed');
        });
    }

    revealNextEvent() {
        if (this.currentEventIndex >= this.journeyData.events.length) return;

        const eventEl = document.querySelector('[data-index="' + this.currentEventIndex + '"]');
        if (eventEl) {
            eventEl.style.transition = 'all 0.5s ease-out';
            eventEl.style.opacity = '1';
            eventEl.style.transform = 'scale(1)';
            eventEl.classList.add('revealed');

            // Add entrance animation based on event type
            const event = this.journeyData.events[this.currentEventIndex];
            this.playEventEntrance(eventEl, event);
        }

        this.currentEventIndex++;
        this.updateProgress();
        this.updateEventCounter();
    }

    playEventEntrance(eventEl, event) {
        switch (event.eventType) {
            case 'milestone':
                eventEl.style.animation = 'milestone-entrance 0.8s ease-out';
                break;
            case 'learning':
                eventEl.style.animation = 'learning-bounce 0.6s ease-out';
                break;
            case 'decision':
                eventEl.style.animation = 'decision-flash 0.4s ease-out';
                break;
            default:
                eventEl.style.animation = 'default-entrance 0.5s ease-out';
        }

        // Clear animation after completion
        setTimeout(() => {
            eventEl.style.animation = '';
        }, 1000);
    }

    // Progress tracking and scrubbing methods
    updateProgress() {
        if (!this.journeyData || !this.journeyData.events.length) return;

        const progress = (this.currentEventIndex / this.journeyData.events.length) * 100;
        const progressFill = document.getElementById('progressFill');
        const progressSlider = document.getElementById('progressSlider');

        if (progressFill) {
            progressFill.style.width = progress + '%';
        }
        if (progressSlider) {
            progressSlider.value = progress;
        }
    }

    updateEventCounter() {
        const counter = document.getElementById('eventCounter');
        const currentDate = document.getElementById('currentDate');

        if (counter && this.journeyData) {
            counter.textContent = this.currentEventIndex + ' / ' + this.journeyData.events.length + ' events';
        }

        if (currentDate && this.journeyData && this.currentEventIndex > 0) {
            const currentEvent = this.journeyData.events[this.currentEventIndex - 1];
            if (currentEvent) {
                const date = new Date(currentEvent.timestamp);
                currentDate.textContent = date.toLocaleDateString() + ' ' + date.toLocaleTimeString();
            }
        }
    }

    updateTimelineInfo() {
        const totalTime = document.getElementById('totalTime');
        const currentTime = document.getElementById('currentTime');

        if (this.journeyData && this.journeyData.events.length > 0) {
            const duration = this.calculateJourneyDuration();
            if (totalTime) {
                totalTime.textContent = this.formatDuration(duration);
            }
            if (currentTime) {
                currentTime.textContent = '0:00';
            }

            // Update event counter
            this.updateEventCounter();
        }
    }

    calculateJourneyDuration() {
        if (!this.journeyData || this.journeyData.events.length < 2) return 0;

        const firstEvent = new Date(this.journeyData.events[0].timestamp);
        const lastEvent = new Date(this.journeyData.events[this.journeyData.events.length - 1].timestamp);
        return Math.floor((lastEvent - firstEvent) / (1000 * 60 * 60)); // hours
    }

    formatDuration(hours) {
        if (hours < 1) return '0:30'; // Minimum display
        if (hours < 24) return hours + ':00';
        const days = Math.floor(hours / 24);
        const remainingHours = hours % 24;
        return days + 'd ' + remainingHours + 'h';
    }

    scrubToPosition(percentage) {
        if (!this.journeyData || !this.journeyData.events.length) return;

        this.pause();
        const targetIndex = Math.floor((percentage / 100) * this.journeyData.events.length);

        // Reset timeline
        this.resetTimeline();

        // Reveal events up to target index instantly
        for (let i = 0; i < targetIndex; i++) {
            const eventEl = document.querySelector('[data-index="' + i + '"]');
            if (eventEl) {
                eventEl.style.transition = 'none';
                eventEl.style.opacity = '1';
                eventEl.style.transform = 'scale(1)';
                eventEl.classList.add('revealed');
            }
        }

        this.currentEventIndex = targetIndex;
        this.updateProgress();
        this.updateEventCounter();

        // Restore transitions
        setTimeout(() => {
            document.querySelectorAll('.timeline-event').forEach(el => {
                el.style.transition = 'all 0.5s ease-out';
            });
        }, 50);
    }
}

// Initialize the visualization when the page loads
document.addEventListener('DOMContentLoaded', () => {
    new JourneyVisualization();
});
`
