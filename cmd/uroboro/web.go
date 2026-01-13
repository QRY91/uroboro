package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"

	"github.com/QRY91/uroboro/internal/database"
	"github.com/QRY91/uroboro/internal/journey"
)

var timelineCSS = `/* Uroboro Timeline */
*{margin:0;padding:0;box-sizing:border-box}
:root{--primary:#8b7aa8;--accent:#7D56F4;--bg-primary:#2a1f3d;--bg-secondary:#1e1b2e;--bg-tertiary:#3a2d4a;--bg-dark:#0f0a1a;--text-primary:#e2d5f0;--text-secondary:#c4b5d9;--text-muted:#a78bbd;--text-dim:#626262;--border:#4c3a5f;--project-red:#FF6B6B;--project-teal:#4ECDC4;--project-blue:#45B7D1;--project-green:#96CEB4;--project-yellow:#FECA57;--project-pink:#FF9FF3;--project-lightblue:#54A0FF;--project-purple:#5F27CD;--font-sans:"Inter",-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;--font-mono:"JetBrains Mono","SF Mono",monospace;--space-xs:0.25rem;--space-sm:0.5rem;--space-md:1rem;--space-lg:1.5rem;--space-xl:2rem;--radius:0.375rem;--radius-lg:0.5rem}
body{font-family:var(--font-sans);background:linear-gradient(135deg,#1e1b2e 0%,#2a1f3d 50%,#342951 100%);background-attachment:fixed;color:var(--text-primary);min-height:100vh;line-height:1.5;-webkit-font-smoothing:antialiased}
.container{max-width:100%;margin:0 auto;padding:var(--space-md)}
.container.list-view{max-width:900px}
header{display:flex;justify-content:space-between;align-items:center;padding:var(--space-md) 0;border-bottom:1px solid var(--border);margin-bottom:var(--space-lg)}
header h1{font-family:var(--font-mono);font-size:1.5rem;font-weight:600;color:var(--text-primary);display:flex;align-items:center;gap:var(--space-sm)}
header h1::before{content:"";width:12px;height:12px;background:var(--accent);border-radius:50%}
.stats{display:flex;gap:var(--space-md);font-family:var(--font-mono);font-size:0.875rem;color:var(--text-muted)}
.filters{display:flex;flex-wrap:wrap;gap:var(--space-sm);margin-bottom:var(--space-lg);padding:var(--space-md);background:var(--bg-tertiary);border-radius:var(--radius-lg);border:1px solid var(--border)}
.filters select,.filters input[type="search"]{padding:var(--space-sm) var(--space-md);background:var(--bg-secondary);border:1px solid var(--border);border-radius:var(--radius);color:var(--text-primary);font-family:var(--font-mono);font-size:0.875rem;min-width:140px}
.filters select:focus,.filters input[type="search"]:focus{outline:none;border-color:var(--accent);box-shadow:0 0 0 2px rgba(125,86,244,0.2)}
.filters input[type="search"]{flex:1;min-width:200px}
.filters input[type="search"]::placeholder{color:var(--text-dim)}
.filters button{padding:var(--space-sm) var(--space-md);background:var(--bg-secondary);border:1px solid var(--border);border-radius:var(--radius);color:var(--text-secondary);font-family:var(--font-mono);font-size:0.875rem;cursor:pointer;transition:all 0.15s ease}
.filters button:hover{background:var(--bg-primary);color:var(--text-primary)}
.filters button.active{background:var(--accent);border-color:var(--accent);color:white}
.view-toggle{display:flex;gap:2px;background:var(--bg-secondary);border-radius:var(--radius);padding:2px;margin-left:auto}
.view-toggle button{padding:var(--space-xs) var(--space-sm);background:transparent;border:none;border-radius:calc(var(--radius) - 2px);color:var(--text-muted);font-family:var(--font-mono);font-size:0.75rem;cursor:pointer}
.view-toggle button.active{background:var(--accent);color:white}
.timeline{display:flex;flex-direction:column;gap:var(--space-xs)}
.day-group{margin-bottom:var(--space-md)}
.day-header{font-family:var(--font-mono);font-size:0.75rem;font-weight:600;color:var(--project-yellow);padding:var(--space-sm) 0;margin-top:var(--space-md);border-bottom:1px dashed var(--border)}
.day-header::before{content:"── ";color:var(--border)}
.day-header::after{content:" ──";color:var(--border)}
.event{display:grid;grid-template-columns:50px 120px 40px 1fr;gap:var(--space-sm);align-items:center;padding:var(--space-sm) var(--space-md);border-left:3px solid var(--border);background:transparent;border-radius:0 var(--radius) var(--radius) 0;cursor:pointer;transition:all 0.15s ease;font-family:var(--font-mono);font-size:0.875rem}
.event:hover{background:var(--bg-tertiary)}
.event time{color:var(--text-dim);font-size:0.8rem}
.event .project{font-weight:600;font-size:0.8rem;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.event .type-icon{font-size:0.75rem;font-weight:500;text-align:center}
.event .content{color:var(--text-secondary);overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.type-git{color:var(--text-dim)}.type-milestone{color:var(--project-yellow)}.type-learning{color:var(--project-lightblue)}.type-decision{color:var(--project-pink)}.type-bugfix{color:var(--project-red)}.type-feature{color:var(--project-teal)}.type-capture{color:var(--text-dim)}
.empty-state{text-align:center;padding:var(--space-xl);color:var(--text-muted);font-family:var(--font-mono)}
/* Horizontal Timeline View */
.h-timeline{position:relative;overflow-x:auto;overflow-y:hidden;padding:var(--space-md) 0}
.h-timeline-inner{position:relative;min-width:100%;height:auto;padding-bottom:var(--space-md)}
.h-timeline-ruler{position:sticky;top:0;left:0;right:0;height:30px;background:var(--bg-secondary);border-bottom:1px solid var(--border);display:flex;font-family:var(--font-mono);font-size:0.7rem;color:var(--text-dim);z-index:10}
.h-timeline-tick{position:absolute;top:0;display:flex;flex-direction:column;align-items:center;padding-top:4px}
.h-timeline-tick::after{content:'';width:1px;height:10px;background:var(--border);margin-top:2px}
.h-timeline-lanes{position:relative;min-height:200px}
.h-timeline-lane{position:relative;height:50px;border-bottom:1px solid var(--border);display:flex;align-items:center}
.h-timeline-lane-label{position:sticky;left:0;width:100px;min-width:100px;padding:0 var(--space-sm);background:var(--bg-tertiary);font-family:var(--font-mono);font-size:0.75rem;font-weight:600;z-index:5;height:100%;display:flex;align-items:center}
.h-timeline-lane-events{position:relative;flex:1;height:100%}
.h-timeline-event{position:absolute;top:50%;transform:translateY(-50%);width:12px;height:12px;border-radius:50%;cursor:pointer;transition:all 0.15s ease;z-index:1}
.h-timeline-event:hover{transform:translateY(-50%) scale(1.5);z-index:20}
.h-timeline-event::before{content:attr(data-time);position:absolute;bottom:calc(100% + 4px);left:50%;transform:translateX(-50%);font-family:var(--font-mono);font-size:0.65rem;color:var(--text-dim);white-space:nowrap;opacity:0;transition:opacity 0.15s}
.h-timeline-event:hover::before{opacity:1}
.h-timeline-event.type-commit{background:var(--text-dim)}.h-timeline-event.type-milestone{background:var(--project-yellow)}.h-timeline-event.type-learning{background:var(--project-lightblue)}.h-timeline-event.type-decision{background:var(--project-pink)}.h-timeline-event.type-bugfix{background:var(--project-red)}.h-timeline-event.type-feature{background:var(--project-teal)}.h-timeline-event.type-capture{background:var(--text-muted)}
.h-timeline-now{position:absolute;top:30px;bottom:0;width:2px;background:var(--accent);z-index:15}
.h-timeline-now::before{content:'now';position:absolute;top:-20px;left:50%;transform:translateX(-50%);font-family:var(--font-mono);font-size:0.65rem;color:var(--accent)}
.modal-backdrop{position:fixed;inset:0;background:rgba(15,10,26,0.85);backdrop-filter:blur(4px);display:flex;align-items:center;justify-content:center;padding:var(--space-md);z-index:100}
.modal-content{background:var(--bg-secondary);border:1px solid var(--border);border-radius:var(--radius-lg);padding:var(--space-lg);max-width:600px;width:100%;max-height:80vh;overflow-y:auto;box-shadow:0 25px 50px -12px rgba(0,0,0,0.5)}
.modal-content h3{font-family:var(--font-mono);font-size:1rem;color:var(--text-primary);margin-bottom:var(--space-md);padding-bottom:var(--space-sm);border-bottom:1px solid var(--border)}
.modal-content dl{display:grid;grid-template-columns:80px 1fr;gap:var(--space-sm);font-family:var(--font-mono);font-size:0.875rem}
.modal-content dt{color:var(--text-dim)}
.modal-content dd{color:var(--text-primary);word-break:break-word}
.modal-content .content-block{margin-top:var(--space-md);padding:var(--space-md);background:var(--bg-tertiary);border-radius:var(--radius);white-space:pre-wrap;line-height:1.6}
.modal-content button{margin-top:var(--space-lg);padding:var(--space-sm) var(--space-md);background:var(--accent);border:none;border-radius:var(--radius);color:white;font-family:var(--font-mono);font-size:0.875rem;cursor:pointer;transition:opacity 0.15s ease}
.modal-content button:hover{opacity:0.9}
@media(max-width:640px){.event{grid-template-columns:45px 1fr;gap:var(--space-xs)}.event .project,.event .type-icon{display:none}.filters{flex-direction:column}.filters select,.filters input[type="search"]{width:100%}.view-toggle{display:none}}`

var htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>uroboro timeline</title>
  <script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
  <style>{{.CSS}}</style>
</head>
<body>
  <div class="container" :class="{ 'list-view': viewMode === 'list' }" x-data="timeline()" x-init="init()">
    <header>
      <h1>uroboro</h1>
      <div class="stats">
        <span x-text="filteredEvents.length + ' events'"></span>
        <span x-text="data.projects?.length + ' projects'"></span>
      </div>
    </header>
    <div class="filters">
      <select x-model="projectFilter">
        <option value="">All Projects</option>
        <template x-for="p in data.projects" :key="p.name">
          <option :value="p.name" x-text="p.name"></option>
        </template>
      </select>
      <select x-model="typeFilter">
        <option value="">All Types</option>
        <template x-for="t in eventTypes" :key="t">
          <option :value="t" x-text="t"></option>
        </template>
      </select>
      <input type="search" x-model="searchQuery" placeholder="Search content, project, tags..." @keydown.escape="searchQuery = ''">
      <button x-show="viewMode === 'list'" @click="groupByDay = !groupByDay" :class="{ active: groupByDay }">Group by Day</button>
      <div class="view-toggle">
        <button @click="viewMode = 'list'" :class="{ active: viewMode === 'list' }">List</button>
        <button @click="viewMode = 'timeline'" :class="{ active: viewMode === 'timeline' }">Timeline</button>
      </div>
    </div>
    <!-- List View -->
    <main class="timeline" x-show="viewMode === 'list'">
      <div class="empty-state" x-show="filteredEvents.length === 0">No events match filters.</div>
      <template x-if="groupByDay && viewMode === 'list'">
        <div>
          <template x-for="(group, date) in groupedEvents" :key="date">
            <div class="day-group">
              <h2 class="day-header" x-text="date"></h2>
              <template x-for="event in group" :key="event.timestamp + event.content">
                <div class="event" :style="'border-left-color: ' + getProjectColor(event.project)" @click="selectedEvent = event">
                  <time x-text="formatTime(event.timestamp)"></time>
                  <span class="project" :style="'color: ' + getProjectColor(event.project)" x-text="event.project || '-'"></span>
                  <span class="type-icon" :class="'type-' + event.eventType" x-text="getTypeIcon(event.eventType)"></span>
                  <span class="content" x-text="event.content"></span>
                </div>
              </template>
            </div>
          </template>
        </div>
      </template>
      <template x-if="!groupByDay && viewMode === 'list'">
        <div>
          <template x-for="event in filteredEvents" :key="event.timestamp + event.content">
            <div class="event" :style="'border-left-color: ' + getProjectColor(event.project)" @click="selectedEvent = event">
              <time x-text="formatTime(event.timestamp)"></time>
              <span class="project" :style="'color: ' + getProjectColor(event.project)" x-text="event.project || '-'"></span>
              <span class="type-icon" :class="'type-' + event.eventType" x-text="getTypeIcon(event.eventType)"></span>
              <span class="content" x-text="event.content"></span>
            </div>
          </template>
        </div>
      </template>
    </main>
    <!-- Horizontal Timeline View -->
    <div class="h-timeline" x-show="viewMode === 'timeline'" x-ref="htimeline">
      <div class="h-timeline-inner" :style="'width: ' + timelineWidth + 'px'">
        <div class="h-timeline-ruler">
          <template x-for="tick in timelineTicks" :key="tick.time">
            <div class="h-timeline-tick" :style="'left: ' + tick.pos + 'px'" x-text="tick.label"></div>
          </template>
        </div>
        <div class="h-timeline-lanes">
          <template x-for="lane in projectLanes" :key="lane.name">
            <div class="h-timeline-lane">
              <div class="h-timeline-lane-label" :style="'color: ' + getProjectColor(lane.name)" x-text="lane.name || 'No Project'"></div>
              <div class="h-timeline-lane-events">
                <template x-for="event in getLaneEvents(lane.name)" :key="event.timestamp + event.content">
                  <div class="h-timeline-event"
                       :class="'type-' + event.eventType"
                       :style="'left: ' + getEventPosition(event) + 'px; background-color: ' + getProjectColor(event.project)"
                       :data-time="formatTime(event.timestamp)"
                       @click="selectedEvent = event"></div>
                </template>
              </div>
            </div>
          </template>
        </div>
        <div class="h-timeline-now" :style="'left: ' + nowPosition + 'px'" x-show="nowPosition > 100"></div>
      </div>
    </div>
    <div class="modal-backdrop" x-show="selectedEvent" x-transition.opacity @click.self="selectedEvent = null" @keydown.escape.window="selectedEvent = null">
      <div class="modal-content" x-show="selectedEvent" x-transition.scale.90>
        <h3>Event Details</h3>
        <dl>
          <dt>Time</dt><dd x-text="selectedEvent ? formatDateTime(selectedEvent.timestamp) : ''"></dd>
          <dt>Project</dt><dd :style="'color: ' + getProjectColor(selectedEvent?.project)" x-text="selectedEvent?.project || '-'"></dd>
          <dt>Type</dt><dd x-text="selectedEvent?.eventType"></dd>
          <template x-if="selectedEvent?.tags?.length"><dt>Tags</dt></template>
          <template x-if="selectedEvent?.tags?.length"><dd x-text="selectedEvent?.tags?.join(', ')"></dd></template>
          <template x-if="selectedEvent?.gitHash"><dt>Commit</dt></template>
          <template x-if="selectedEvent?.gitHash"><dd x-text="selectedEvent?.gitHash?.slice(0, 8)"></dd></template>
        </dl>
        <div class="content-block" x-text="selectedEvent?.content"></div>
        <button @click="selectedEvent = null">Close</button>
      </div>
    </div>
  </div>
  <script>
    window.JOURNEY_DATA = {{.DataJSON}};
    function timeline() {
      return {
        data: window.JOURNEY_DATA || { events: [], projects: [], stats: {}, milestones: [] },
        projectFilter: '', typeFilter: '', searchQuery: '', groupByDay: true, selectedEvent: null,
        viewMode: 'timeline',
        eventTypes: ['capture', 'commit', 'milestone', 'learning', 'decision', 'bugfix', 'feature'],
        projectColors: {},
        colorPalette: ['#FF6B6B', '#4ECDC4', '#45B7D1', '#96CEB4', '#FECA57', '#FF9FF3', '#54A0FF', '#5F27CD'],
        timeRange: { start: null, end: null },
        pxPerHour: 60,
        init() {
          if (this.data.projects) {
            this.data.projects.forEach((p, i) => {
              this.projectColors[p.name] = p.color || this.colorPalette[i % this.colorPalette.length];
            });
          }
          this.computeTimeRange();
        },
        computeTimeRange() {
          if (!this.data.events || this.data.events.length === 0) return;
          const times = this.data.events.map(e => new Date(e.timestamp).getTime());
          this.timeRange.start = Math.min(...times);
          this.timeRange.end = Math.max(...times);
        },
        get filteredEvents() {
          if (!this.data.events) return [];
          return this.data.events.filter(e => {
            if (this.projectFilter && e.project !== this.projectFilter) return false;
            if (this.typeFilter && e.eventType !== this.typeFilter) return false;
            if (this.searchQuery) {
              const q = this.searchQuery.toLowerCase();
              const inContent = (e.content || '').toLowerCase().includes(q);
              const inProject = (e.project || '').toLowerCase().includes(q);
              const inTags = (e.tags || []).some(t => t.toLowerCase().includes(q));
              if (!inContent && !inProject && !inTags) return false;
            }
            return true;
          });
        },
        get groupedEvents() {
          const groups = {};
          this.filteredEvents.forEach(event => {
            const date = new Date(event.timestamp).toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric', year: 'numeric' });
            if (!groups[date]) groups[date] = [];
            groups[date].push(event);
          });
          return groups;
        },
        get projectLanes() {
          const projects = new Set();
          this.filteredEvents.forEach(e => projects.add(e.project || ''));
          return Array.from(projects).map(name => ({ name }));
        },
        get timelineWidth() {
          if (!this.timeRange.start || !this.timeRange.end) return 1000;
          const hours = (this.timeRange.end - this.timeRange.start) / (1000 * 60 * 60);
          return Math.max(1000, 100 + hours * this.pxPerHour + 50);
        },
        get timelineTicks() {
          if (!this.timeRange.start || !this.timeRange.end) return [];
          const ticks = [];
          const start = new Date(this.timeRange.start);
          // Align to midnight or noon
          start.setMinutes(0, 0, 0);
          start.setHours(start.getHours() < 12 ? 0 : 12);
          const end = new Date(this.timeRange.end);
          // Always 12-hour steps (midnight and noon)
          for (let d = new Date(start); d <= end; d.setHours(d.getHours() + 12)) {
            const pos = 100 + ((d.getTime() - this.timeRange.start) / (1000 * 60 * 60)) * this.pxPerHour;
            // Midnight: show date, Noon: show 12:00
            const label = d.getHours() === 0 ? d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' }) : '12:00';
            ticks.push({ time: d.getTime(), pos, label });
          }
          return ticks;
        },
        get nowPosition() {
          if (!this.timeRange.start) return 0;
          const now = Date.now();
          if (now < this.timeRange.start || now > this.timeRange.end + 3600000) return 0;
          return 100 + ((now - this.timeRange.start) / (1000 * 60 * 60)) * this.pxPerHour;
        },
        getLaneEvents(projectName) {
          return this.filteredEvents.filter(e => (e.project || '') === projectName);
        },
        getEventPosition(event) {
          if (!this.timeRange.start) return 0;
          const t = new Date(event.timestamp).getTime();
          return ((t - this.timeRange.start) / (1000 * 60 * 60)) * this.pxPerHour;
        },
        formatTime(ts) { return new Date(ts).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', hour12: false }); },
        formatDateTime(ts) { return new Date(ts).toLocaleString('en-US', { weekday: 'short', month: 'short', day: 'numeric', year: 'numeric', hour: '2-digit', minute: '2-digit', hour12: false }); },
        getProjectColor(project) { return this.projectColors[project] || '#626262'; },
        getTypeIcon(type) { const icons = { commit: 'git', milestone: '***', learning: 'lrn', decision: 'dec', bugfix: 'fix', feature: 'fea' }; return icons[type] || 'cap'; }
      };
    }
  </script>
</body>
</html>`

func handleWeb(args []string) {
	fs := flag.NewFlagSet("web", flag.ExitOnError)
	days := fs.Int("days", 7, "Days to show")
	port := fs.Int("port", 8080, "HTTP port")
	project := fs.String("project", "", "Filter by project")
	fs.Parse(args)

	db, err := database.NewDB(getDBPath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Database error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	var projects []string
	if *project != "" {
		projects = strings.Split(*project, ",")
	}

	svc := journey.NewService(db)

	// Serve timeline page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		data, err := svc.GenerateJourney(journey.Options{Days: *days, Projects: projects})
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		jsonBytes, _ := json.Marshal(data)

		tmpl, err := template.New("timeline").Parse(htmlTemplate)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.Execute(w, map[string]interface{}{
			"CSS":      template.CSS(timelineCSS),
			"DataJSON": template.JS(string(jsonBytes)),
		})
	})

	// API endpoint for live refresh
	http.HandleFunc("/api/journey", func(w http.ResponseWriter, r *http.Request) {
		data, err := svc.GenerateJourney(journey.Options{Days: *days, Projects: projects})
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	})

	// Serve static CSS
	http.HandleFunc("/timeline.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		w.Write([]byte(timelineCSS))
	})

	fmt.Printf("Timeline at http://localhost:%d\n", *port)
	fmt.Println("Press Ctrl+C to stop")

	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

// GenerateStandaloneHTML creates a self-contained HTML file with embedded data
func GenerateStandaloneHTML(data *journey.JourneyData) (string, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	tmpl, err := template.New("timeline").Parse(htmlTemplate)
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, map[string]interface{}{
		"CSS":      template.CSS(timelineCSS),
		"DataJSON": template.JS(string(jsonBytes)),
	})
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
