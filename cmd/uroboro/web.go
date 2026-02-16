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
.event .project{font-weight:600;font-size:0.8rem;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;cursor:pointer}.event .project:hover{text-decoration:underline}
.event .type-icon{font-size:0.75rem;font-weight:500;text-align:center}
.event .content{color:var(--text-secondary);overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.type-git{color:var(--text-dim)}.type-milestone{color:var(--project-yellow)}.type-learning{color:var(--project-lightblue)}.type-decision{color:var(--project-pink)}.type-bugfix{color:var(--project-red)}.type-feature{color:var(--project-teal)}.type-capture{color:var(--text-dim)}.type-blocker{color:var(--project-red)}.type-question{color:var(--project-yellow)}
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
.h-timeline-event.type-commit{background:var(--text-dim)}.h-timeline-event.type-milestone{background:var(--project-yellow)}.h-timeline-event.type-learning{background:var(--project-lightblue)}.h-timeline-event.type-decision{background:var(--project-pink)}.h-timeline-event.type-bugfix{background:var(--project-red)}.h-timeline-event.type-feature{background:var(--project-teal)}.h-timeline-event.type-capture{background:var(--text-muted)}.h-timeline-event.type-blocker{background:var(--project-red)}.h-timeline-event.type-question{background:var(--project-yellow)}
.h-timeline-now{position:absolute;top:30px;bottom:0;width:2px;background:var(--accent);z-index:15}
.h-timeline-now::before{content:'now';position:absolute;top:-20px;left:50%;transform:translateX(-50%);font-family:var(--font-mono);font-size:0.65rem;color:var(--accent)}
.h-timeline-gap{position:absolute;top:0;height:100%;background:repeating-linear-gradient(90deg,transparent,transparent 4px,var(--border) 4px,var(--border) 8px);opacity:0.5}
.h-timeline-gap::before{content:attr(data-hours);position:absolute;top:50%;left:50%;transform:translate(-50%,-50%);font-family:var(--font-mono);font-size:0.6rem;color:var(--text-dim);background:var(--bg-secondary);padding:0 4px;border-radius:2px}
.h-timeline-gap-line{position:absolute;top:30px;bottom:0;background:repeating-linear-gradient(90deg,transparent,transparent 4px,var(--border) 4px,var(--border) 8px);opacity:0.3;z-index:0}
.modal-backdrop{position:fixed;inset:0;background:rgba(15,10,26,0.85);backdrop-filter:blur(4px);display:flex;align-items:center;justify-content:center;padding:var(--space-md);z-index:100}
.modal-content{background:var(--bg-secondary);border:1px solid var(--border);border-radius:var(--radius-lg);padding:var(--space-lg);max-width:600px;width:100%;max-height:80vh;overflow-y:auto;box-shadow:0 25px 50px -12px rgba(0,0,0,0.5)}
.modal-content h3{font-family:var(--font-mono);font-size:1rem;color:var(--text-primary);margin-bottom:var(--space-md);padding-bottom:var(--space-sm);border-bottom:1px solid var(--border)}
.modal-content dl{display:grid;grid-template-columns:80px 1fr;gap:var(--space-sm);font-family:var(--font-mono);font-size:0.875rem}
.modal-content dt{color:var(--text-dim)}
.modal-content dd{color:var(--text-primary);word-break:break-word}
.modal-content .content-block{margin-top:var(--space-md);padding:var(--space-md);background:var(--bg-tertiary);border-radius:var(--radius);white-space:pre-wrap;line-height:1.6}
.modal-content button{margin-top:var(--space-lg);padding:var(--space-sm) var(--space-md);background:var(--accent);border:none;border-radius:var(--radius);color:white;font-family:var(--font-mono);font-size:0.875rem;cursor:pointer;transition:opacity 0.15s ease}
.modal-content button:hover{opacity:0.9}
/* Vertical Timeline (Project Detail) */
.v-timeline{max-width:700px;margin:0 auto;padding:var(--space-md) 0}
.v-timeline-stats{font-family:var(--font-mono);font-size:0.8rem;color:var(--text-muted);display:flex;gap:var(--space-sm);margin-bottom:var(--space-lg);padding-bottom:var(--space-sm);border-bottom:1px dashed var(--border)}
.v-timeline-line{position:relative;padding-left:28px}
.v-timeline-line::before{content:'';position:absolute;left:8px;top:0;bottom:0;width:2px;background:var(--border)}
.v-timeline-node{position:relative;margin-bottom:var(--space-lg)}
.v-timeline-dot{position:absolute;left:-24px;top:4px;width:14px;height:14px;border-radius:50%;border:2px solid var(--bg-secondary);z-index:1}
.v-timeline-dot.type-commit{background:var(--text-dim)}
.v-timeline-dot.type-milestone{background:var(--project-yellow)}
.v-timeline-dot.type-learning{background:var(--project-lightblue)}
.v-timeline-dot.type-decision{background:var(--project-pink)}
.v-timeline-dot.type-bugfix{background:var(--project-red)}
.v-timeline-dot.type-feature{background:var(--project-teal)}
.v-timeline-dot.type-capture{background:var(--text-muted)}
.v-timeline-dot.type-blocker{background:var(--project-red)}
.v-timeline-dot.type-question{background:var(--project-yellow)}
.v-timeline-card{background:var(--bg-tertiary);border:1px solid var(--border);border-radius:var(--radius-lg);padding:var(--space-md);cursor:pointer;transition:all 0.15s ease}
.v-timeline-card:hover{border-color:var(--accent);background:var(--bg-primary)}
.v-timeline-meta{display:flex;justify-content:space-between;align-items:center;margin-bottom:var(--space-sm)}
.v-timeline-meta time{font-family:var(--font-mono);font-size:0.75rem;color:var(--text-dim)}
.v-timeline-type{font-family:var(--font-mono);font-size:0.7rem;font-weight:600;padding:1px 6px;border-radius:3px;background:var(--bg-secondary)}
.v-timeline-content{font-family:var(--font-mono);font-size:0.875rem;color:var(--text-secondary);line-height:1.6;white-space:pre-wrap;word-break:break-word}
.v-timeline-tags{display:flex;flex-wrap:wrap;gap:var(--space-xs);margin-top:var(--space-sm)}
.v-timeline-tag{font-family:var(--font-mono);font-size:0.7rem;color:var(--text-muted);background:var(--bg-secondary);padding:1px 6px;border-radius:3px;border:1px solid var(--border)}
.v-timeline-hash{font-family:var(--font-mono);font-size:0.7rem;color:var(--text-dim);margin-top:var(--space-xs)}
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
  <div class="container" :class="{ 'list-view': viewMode === 'list' || viewMode === 'project' }" x-data="timeline()" x-init="init()">
    <header>
      <h1>
        <span x-show="viewMode !== 'project'">uroboro</span>
        <span x-show="viewMode === 'project'" x-text="selectedProject"></span>
      </h1>
      <button x-show="viewMode === 'project'" @click="backToMain()" style="background:var(--bg-tertiary);border:1px solid var(--border);border-radius:var(--radius);color:var(--text-secondary);font-family:var(--font-mono);font-size:0.875rem;padding:var(--space-xs) var(--space-sm);cursor:pointer">&larr; Back</button>
      <div class="stats">
        <span x-text="filteredEvents.length + ' events'"></span>
        <span x-text="data.projects?.length + ' projects'"></span>
      </div>
    </header>
    <div class="filters" x-show="viewMode !== 'project'">
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
      <button x-show="viewMode === 'timeline'" @click="compactMode = !compactMode" :class="{ active: compactMode }">Compact</button>
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
                  <span class="project" :style="'color: ' + getProjectColor(event.project)" x-text="event.project || '-'" @click.stop="event.project && showProject(event.project)"></span>
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
              <span class="project" :style="'color: ' + getProjectColor(event.project)" x-text="event.project || '-'" @click.stop="event.project && showProject(event.project)"></span>
              <span class="type-icon" :class="'type-' + event.eventType" x-text="getTypeIcon(event.eventType)"></span>
              <span class="content" x-text="event.content"></span>
            </div>
          </template>
        </div>
      </template>
    </main>
    <!-- Horizontal Timeline View -->
    <div class="h-timeline" x-show="viewMode === 'timeline'" x-ref="htimeline">
      <div class="h-timeline-inner" :style="'width: ' + (compactMode ? compactTimelineWidth : timelineWidth) + 'px'">
        <div class="h-timeline-ruler">
          <template x-for="tick in (compactMode ? compactTicks : timelineTicks)" :key="tick.time">
            <div class="h-timeline-tick" :style="'left: ' + (compactMode ? tick.pos : tick.pos) + 'px'" x-text="tick.label"></div>
          </template>
          <template x-if="compactMode">
            <template x-for="gap in gapMarkers" :key="gap.pos">
              <div class="h-timeline-gap" :style="'left: ' + gap.pos + 'px; width: ' + gap.width + 'px'" :data-hours="gap.hours + 'h'"></div>
            </template>
          </template>
        </div>
        <div class="h-timeline-lanes">
          <template x-if="compactMode">
            <template x-for="gap in gapMarkers" :key="'lane-gap-' + gap.pos">
              <div class="h-timeline-gap-line" :style="'left: ' + (100 + gap.pos) + 'px; width: ' + gap.width + 'px'"></div>
            </template>
          </template>
          <template x-for="lane in projectLanes" :key="lane.name">
            <div class="h-timeline-lane">
              <div class="h-timeline-lane-label" :style="'color: ' + getProjectColor(lane.name)" x-text="lane.name || 'No Project'" @click="lane.name && showProject(lane.name)" style="cursor:pointer"></div>
              <div class="h-timeline-lane-events">
                <template x-for="event in getLaneEvents(lane.name)" :key="event.timestamp + event.content">
                  <div class="h-timeline-event"
                       :class="'type-' + event.eventType"
                       :style="'left: ' + (compactMode ? getCompactPosition(new Date(event.timestamp).getTime()) : getEventPosition(event)) + 'px; background-color: ' + getProjectColor(event.project)"
                       :data-time="formatTime(event.timestamp)"
                       @click="selectedEvent = event"></div>
                </template>
              </div>
            </div>
          </template>
        </div>
        <div class="h-timeline-now" :style="'left: ' + (compactMode ? getCompactPosition(Date.now()) : nowPosition) + 'px'" x-show="compactMode ? getCompactPosition(Date.now()) > 100 : nowPosition > 100"></div>
      </div>
    </div>
    <!-- Project Detail: Vertical Timeline -->
    <div class="v-timeline" x-show="viewMode === 'project'">
      <div class="v-timeline-stats" x-show="selectedProjectSummary">
        <span x-text="projectEvents.length + ' events'"></span>
        <span>&middot;</span>
        <span x-text="selectedProjectSummary ? 'since ' + formatDateTime(selectedProjectSummary.startDate) : ''"></span>
      </div>
      <div class="empty-state" x-show="projectEvents.length === 0">No events for this project.</div>
      <div class="v-timeline-line">
        <template x-for="(event, idx) in projectEvents" :key="event.timestamp + event.content">
          <div class="v-timeline-node" @click="selectedEvent = event">
            <div class="v-timeline-dot" :class="'type-' + event.eventType"></div>
            <div class="v-timeline-card">
              <div class="v-timeline-meta">
                <time x-text="formatDateTime(event.timestamp)"></time>
                <span class="v-timeline-type" :class="'type-' + event.eventType" x-text="getTypeIcon(event.eventType)"></span>
              </div>
              <div class="v-timeline-content" x-text="event.content"></div>
              <div class="v-timeline-tags" x-show="event.tags && event.tags.length">
                <template x-for="tag in (event.tags || [])" :key="tag">
                  <span class="v-timeline-tag" x-text="tag"></span>
                </template>
              </div>
              <div class="v-timeline-hash" x-show="event.gitHash" x-text="event.gitHash ? event.gitHash.slice(0, 8) : ''"></div>
            </div>
          </div>
        </template>
      </div>
    </div>
    <div class="modal-backdrop" x-show="selectedEvent" x-transition.opacity @click.self="selectedEvent = null" @keydown.escape.window="selectedEvent = null">
      <div class="modal-content" x-show="selectedEvent" x-transition.scale.90>
        <h3>Event Details</h3>
        <dl>
          <dt>Time</dt><dd x-text="selectedEvent ? formatDateTime(selectedEvent.timestamp) : ''"></dd>
          <dt>Project</dt><dd :style="'color: ' + getProjectColor(selectedEvent?.project)" x-text="selectedEvent?.project || '-'"></dd>
          <dt>Type</dt><dd x-text="selectedEvent?.eventType"></dd>
          <template x-if="selectedEvent?.branch"><dt>Branch</dt></template>
          <template x-if="selectedEvent?.branch"><dd x-text="selectedEvent?.branch"></dd></template>
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
        selectedProject: null,
        compactMode: false,
        eventTypes: ['capture', 'commit', 'milestone', 'learning', 'decision', 'bugfix', 'feature', 'blocker', 'question'],
        projectColors: {},
        colorPalette: ['#FF6B6B', '#4ECDC4', '#45B7D1', '#96CEB4', '#FECA57', '#FF9FF3', '#54A0FF', '#5F27CD'],
        timeRange: { start: null, end: null },
        pxPerHour: 60,
        restGapThreshold: 2 * 60 * 60 * 1000, // 2 hours = rest gap
        collapsedGapWidth: 40, // px for collapsed gap
        gaps: [], // detected rest gaps
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
          this.detectGaps();
        },
        detectGaps() {
          // Find gaps longer than threshold (rest periods)
          if (!this.data.events || this.data.events.length < 2) { this.gaps = []; return; }
          const sorted = [...this.data.events].sort((a, b) => new Date(a.timestamp) - new Date(b.timestamp));
          const gaps = [];
          for (let i = 1; i < sorted.length; i++) {
            const prev = new Date(sorted[i-1].timestamp).getTime();
            const curr = new Date(sorted[i].timestamp).getTime();
            const gap = curr - prev;
            if (gap > this.restGapThreshold) {
              gaps.push({ start: prev, end: curr, duration: gap });
            }
          }
          this.gaps = gaps;
        },
        getCompactPosition(timestamp) {
          // Calculate position with gaps collapsed
          const t = new Date(timestamp).getTime();
          let pos = 0;
          let lastEnd = this.timeRange.start;
          for (const gap of this.gaps) {
            if (t <= gap.start) {
              // Event is before this gap
              pos += ((t - lastEnd) / (1000 * 60 * 60)) * this.pxPerHour;
              return pos;
            }
            // Add time before gap
            pos += ((gap.start - lastEnd) / (1000 * 60 * 60)) * this.pxPerHour;
            // Add collapsed gap width
            pos += this.collapsedGapWidth;
            lastEnd = gap.end;
          }
          // Event is after all gaps
          pos += ((t - lastEnd) / (1000 * 60 * 60)) * this.pxPerHour;
          return pos;
        },
        get compactTimelineWidth() {
          if (!this.timeRange.start || !this.timeRange.end) return 1000;
          let totalMs = this.timeRange.end - this.timeRange.start;
          let collapsedMs = 0;
          for (const gap of this.gaps) { collapsedMs += gap.duration; }
          const activeMs = totalMs - collapsedMs;
          const activeHours = activeMs / (1000 * 60 * 60);
          const collapsedWidth = this.gaps.length * this.collapsedGapWidth;
          return Math.max(1000, 100 + activeHours * this.pxPerHour + collapsedWidth + 50);
        },
        get compactTicks() {
          // Generate ticks that skip over gaps
          if (!this.timeRange.start || !this.timeRange.end) return [];
          const ticks = [];
          const start = new Date(this.timeRange.start);
          start.setMinutes(0, 0, 0);
          start.setHours(start.getHours() < 12 ? 0 : 12);
          const end = new Date(this.timeRange.end);
          for (let d = new Date(start); d <= end; d.setHours(d.getHours() + 12)) {
            const t = d.getTime();
            // Skip ticks that fall inside gaps
            const inGap = this.gaps.some(g => t > g.start && t < g.end);
            if (inGap) continue;
            const pos = this.getCompactPosition(t);
            const label = d.getHours() === 0 ? d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' }) : '12:00';
            ticks.push({ time: t, pos, label });
          }
          return ticks;
        },
        get gapMarkers() {
          // Return positions for gap indicators
          return this.gaps.map(gap => {
            const pos = this.getCompactPosition(gap.start);
            const hours = Math.round(gap.duration / (1000 * 60 * 60));
            return { pos, hours, width: this.collapsedGapWidth };
          });
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
              const inBranch = (e.branch || '').toLowerCase().includes(q);
              const inTags = (e.tags || []).some(t => t.toLowerCase().includes(q));
              if (!inContent && !inProject && !inBranch && !inTags) return false;
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
        get projectEvents() {
          if (!this.selectedProject || !this.data.events) return [];
          return this.data.events
            .filter(e => e.project === this.selectedProject)
            .sort((a, b) => new Date(b.timestamp) - new Date(a.timestamp));
        },
        get selectedProjectSummary() {
          if (!this.selectedProject || !this.data.projects) return null;
          return this.data.projects.find(p => p.name === this.selectedProject) || null;
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
        getTypeIcon(type) { const icons = { commit: 'git', milestone: '***', learning: 'lrn', decision: 'dec', bugfix: 'fix', feature: 'fea', blocker: 'blk', question: '?' }; return icons[type] || 'cap'; },
        showProject(name) { this.selectedProject = name; this.viewMode = 'project'; },
        backToMain() { this.viewMode = 'timeline'; this.selectedProject = null; }
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

var graphCSS = `/* Uroboro Graph - Canvas scatter plot */
*{margin:0;padding:0;box-sizing:border-box}
:root{--bg-primary:#2a1f3d;--bg-secondary:#1e1b2e;--bg-tertiary:#3a2d4a;--text-primary:#e2d5f0;--text-muted:#a78bbd;--text-dim:#626262;--border:#4c3a5f;--accent:#7D56F4;--font-mono:"JetBrains Mono","SF Mono",monospace}
html,body{height:100%;overflow:hidden}
body{font-family:var(--font-mono);background:var(--bg-secondary);color:var(--text-primary)}
.container{display:flex;flex-direction:column;height:100vh;padding:0.75rem}
header{display:flex;justify-content:space-between;align-items:center;padding:0.5rem 0;margin-bottom:0.5rem;flex-shrink:0}
header h1{font-size:1.1rem;display:flex;align-items:center;gap:0.5rem}
header h1::before{content:"";width:8px;height:8px;background:var(--accent);border-radius:50%}
.stats{display:flex;gap:1rem;font-size:0.75rem;color:var(--text-muted)}
.filters{display:flex;gap:0.5rem;margin-bottom:0.5rem;flex-shrink:0}
.filters select{padding:0.2rem 0.4rem;background:var(--bg-tertiary);border:1px solid var(--border);border-radius:3px;color:var(--text-primary);font-family:var(--font-mono);font-size:0.7rem}
.graph-wrap{flex:1;position:relative;min-height:0}
canvas{display:block;width:100%;height:100%}
.tooltip{position:fixed;background:var(--bg-tertiary);border:1px solid var(--border);border-radius:4px;padding:0.4rem 0.6rem;font-size:0.7rem;pointer-events:none;z-index:50;max-width:300px;display:none}
.tooltip.show{display:block}
.tooltip .time{color:var(--text-dim);font-size:0.65rem}
.tooltip .content{margin-top:0.2rem;color:var(--text-primary)}`

var graphTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>uroboro graph</title>
  <style>{{.CSS}}</style>
</head>
<body>
  <div class="container">
    <header>
      <h1>uroboro</h1>
      <div class="stats" id="stats"></div>
    </header>
    <div class="filters">
      <select id="projectFilter"><option value="">All Projects</option></select>
      <select id="typeFilter">
        <option value="">All Types</option>
        <option value="capture">capture</option>
        <option value="commit">commit</option>
        <option value="decision">decision</option>
        <option value="milestone">milestone</option>
      </select>
    </div>
    <div class="graph-wrap">
      <canvas id="graph"></canvas>
    </div>
    <div class="tooltip" id="tooltip">
      <div class="time"></div>
      <div class="content"></div>
    </div>
  </div>
  <script>
    const DATA = {{.DataJSON}};
    const DAYS = {{.Days}};
    const COLORS = ['#FF6B6B','#4ECDC4','#45B7D1','#96CEB4','#FECA57','#FF9FF3','#54A0FF','#5F27CD'];

    let projectFilter = '', typeFilter = '';
    let projectColors = {}, projectList = [], events = [];
    let canvas, ctx, tooltip, dpr;
    let margin = { top: 30, right: 20, bottom: 30, left: 120 };
    let hoveredEvent = null;

    function init() {
      canvas = document.getElementById('graph');
      ctx = canvas.getContext('2d');
      tooltip = document.getElementById('tooltip');
      dpr = window.devicePixelRatio || 1;

      // Build project list and colors
      const projSet = new Set();
      (DATA.events || []).forEach(e => projSet.add(e.project || ''));
      projectList = Array.from(projSet).sort();
      projectList.forEach((p, i) => { projectColors[p] = COLORS[i % COLORS.length]; });

      // Populate project filter
      const pf = document.getElementById('projectFilter');
      projectList.forEach(p => {
        const opt = document.createElement('option');
        opt.value = p; opt.textContent = p || '(no project)';
        pf.appendChild(opt);
      });

      // Filter events to time range
      const now = Date.now();
      const start = now - DAYS * 24 * 60 * 60 * 1000;
      events = (DATA.events || []).map(e => ({
        ...e,
        ts: new Date(e.timestamp).getTime()
      })).filter(e => e.ts >= start && e.ts <= now);

      // Event listeners
      document.getElementById('projectFilter').onchange = e => { projectFilter = e.target.value; draw(); };
      document.getElementById('typeFilter').onchange = e => { typeFilter = e.target.value; draw(); };
      window.onresize = () => { resize(); draw(); };
      canvas.onmousemove = onMouseMove;
      canvas.onmouseleave = () => { tooltip.classList.remove('show'); hoveredEvent = null; };

      resize();
      draw();
    }

    function resize() {
      const rect = canvas.parentElement.getBoundingClientRect();
      canvas.width = rect.width * dpr;
      canvas.height = rect.height * dpr;
      canvas.style.width = rect.width + 'px';
      canvas.style.height = rect.height + 'px';
      ctx.scale(dpr, dpr);
    }

    function getFiltered() {
      return events.filter(e => {
        if (projectFilter && e.project !== projectFilter) return false;
        if (typeFilter && e.eventType !== typeFilter) return false;
        return true;
      });
    }

    function draw() {
      const w = canvas.width / dpr, h = canvas.height / dpr;
      ctx.clearRect(0, 0, w, h);

      const filtered = getFiltered();
      const projects = projectFilter ? [projectFilter] : projectList;

      // Update stats
      document.getElementById('stats').textContent = filtered.length + ' events · ' + projects.length + ' projects · ' + DAYS + ' days';

      const plotW = w - margin.left - margin.right;
      const plotH = h - margin.top - margin.bottom;
      const now = Date.now();
      const start = now - DAYS * 24 * 60 * 60 * 1000;

      // Y scale: projects
      const yStep = plotH / Math.max(projects.length, 1);
      const projectY = {};
      projects.forEach((p, i) => { projectY[p] = margin.top + yStep * (i + 0.5); });

      // Draw grid and labels
      ctx.strokeStyle = '#3a2d4a';
      ctx.fillStyle = '#626262';
      ctx.font = '11px "JetBrains Mono", monospace';
      ctx.textAlign = 'right';
      ctx.textBaseline = 'middle';

      projects.forEach(p => {
        const y = projectY[p];
        ctx.beginPath();
        ctx.moveTo(margin.left, y);
        ctx.lineTo(w - margin.right, y);
        ctx.stroke();
        ctx.fillStyle = projectColors[p] || '#626262';
        ctx.fillText(p || '(none)', margin.left - 8, y);
      });

      // X axis: time ticks
      ctx.textAlign = 'center';
      ctx.textBaseline = 'top';
      ctx.fillStyle = '#626262';
      const tickInterval = DAYS <= 30 ? 7 : DAYS <= 90 ? 14 : DAYS <= 365 ? 30 : 90;
      for (let d = 0; d <= DAYS; d += tickInterval) {
        const t = start + d * 24 * 60 * 60 * 1000;
        const x = margin.left + (d / DAYS) * plotW;
        ctx.beginPath();
        ctx.moveTo(x, margin.top);
        ctx.lineTo(x, h - margin.bottom);
        ctx.stroke();
        const date = new Date(t);
        ctx.fillText(date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' }), x, h - margin.bottom + 4);
      }

      // Draw points
      filtered.forEach(e => {
        const x = margin.left + ((e.ts - start) / (now - start)) * plotW;
        const y = projectY[e.project] || projectY[''] || margin.top + plotH / 2;
        ctx.beginPath();
        ctx.arc(x, y, 3, 0, Math.PI * 2);
        ctx.fillStyle = projectColors[e.project] || '#626262';
        ctx.fill();
      });

      // Now line
      const nowX = margin.left + plotW;
      ctx.strokeStyle = '#7D56F4';
      ctx.lineWidth = 2;
      ctx.beginPath();
      ctx.moveTo(nowX, margin.top);
      ctx.lineTo(nowX, h - margin.bottom);
      ctx.stroke();
      ctx.lineWidth = 1;
    }

    function onMouseMove(e) {
      const rect = canvas.getBoundingClientRect();
      const mx = e.clientX - rect.left, my = e.clientY - rect.top;
      const w = canvas.width / dpr, h = canvas.height / dpr;
      const plotW = w - margin.left - margin.right;
      const plotH = h - margin.top - margin.bottom;
      const now = Date.now();
      const start = now - DAYS * 24 * 60 * 60 * 1000;
      const filtered = getFiltered();
      const projects = projectFilter ? [projectFilter] : projectList;
      const yStep = plotH / Math.max(projects.length, 1);
      const projectY = {};
      projects.forEach((p, i) => { projectY[p] = margin.top + yStep * (i + 0.5); });

      // Find closest event
      let closest = null, minDist = 20;
      filtered.forEach(ev => {
        const x = margin.left + ((ev.ts - start) / (now - start)) * plotW;
        const y = projectY[ev.project] || projectY[''] || margin.top + plotH / 2;
        const dist = Math.sqrt((mx - x) ** 2 + (my - y) ** 2);
        if (dist < minDist) { minDist = dist; closest = ev; }
      });

      if (closest) {
        tooltip.querySelector('.time').textContent = new Date(closest.ts).toLocaleString();
        tooltip.querySelector('.content').textContent = closest.content;
        tooltip.style.left = (e.clientX + 10) + 'px';
        tooltip.style.top = (e.clientY + 10) + 'px';
        tooltip.classList.add('show');
      } else {
        tooltip.classList.remove('show');
      }
    }

    init();
  </script>
</body>
</html>`

func handleGraph(args []string) {
	fs := flag.NewFlagSet("graph", flag.ExitOnError)
	days := fs.Int("days", 30, "Days to show")
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

		tmpl, err := template.New("graph").Parse(graphTemplate)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.Execute(w, map[string]interface{}{
			"CSS":      template.CSS(graphCSS),
			"DataJSON": template.JS(string(jsonBytes)),
			"Days":     *days,
		})
	})

	fmt.Printf("Graph at http://localhost:%d\n", *port)
	fmt.Println("Press Ctrl+C to stop")

	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
