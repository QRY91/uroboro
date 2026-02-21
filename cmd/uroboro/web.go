package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
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
@media(max-width:640px){.event{grid-template-columns:45px 1fr;gap:var(--space-xs)}.event .project,.event .type-icon{display:none}.filters{flex-direction:column}.filters select,.filters input[type="search"]{width:100%}.view-toggle{display:none}}
/* Summary Pane (slide-out) */
.summary-pane{position:fixed;top:0;right:-400px;width:380px;height:100vh;background:var(--bg-secondary);border-left:1px solid var(--border);padding:var(--space-lg);overflow-y:auto;z-index:90;transition:right 0.25s ease;box-shadow:-4px 0 20px rgba(0,0,0,0.3)}
.summary-pane.open{right:0}
.summary-pane h2{font-family:var(--font-mono);font-size:0.9rem;font-weight:600;color:var(--text-primary);margin-bottom:var(--space-md);padding-bottom:var(--space-sm);border-bottom:1px solid var(--border)}
.summary-pane h3{font-family:var(--font-mono);font-size:0.75rem;font-weight:600;color:var(--text-muted);margin-top:var(--space-lg);margin-bottom:var(--space-sm);text-transform:uppercase;letter-spacing:0.05em}
.summary-pane .close-btn{position:absolute;top:var(--space-md);right:var(--space-md);background:none;border:none;color:var(--text-muted);font-size:1.2rem;cursor:pointer;font-family:var(--font-mono)}
.summary-pane .close-btn:hover{color:var(--text-primary)}
.type-breakdown{display:flex;flex-direction:column;gap:var(--space-xs)}
.type-row{display:flex;align-items:center;gap:var(--space-sm);font-family:var(--font-mono);font-size:0.8rem}
.type-row .type-dot{width:8px;height:8px;border-radius:50%;flex-shrink:0}
.type-row .type-name{color:var(--text-secondary);flex:1}
.type-row .type-count{color:var(--text-primary);font-weight:600}
.project-list{display:flex;flex-direction:column;gap:var(--space-xs)}
.project-row{display:flex;align-items:center;gap:var(--space-sm);font-family:var(--font-mono);font-size:0.8rem;cursor:pointer;padding:2px 0;border-radius:var(--radius)}
.project-row:hover{background:var(--bg-tertiary)}
.project-row .project-dot{width:8px;height:8px;border-radius:50%;flex-shrink:0}
.project-row .project-name{color:var(--text-secondary);flex:1}
.project-row .project-count{color:var(--text-muted)}
.milestone-list-pane{display:flex;flex-direction:column;gap:var(--space-sm)}
.milestone-item{font-family:var(--font-mono);font-size:0.8rem;color:var(--text-secondary);padding:var(--space-sm);background:var(--bg-tertiary);border-radius:var(--radius);border-left:3px solid var(--project-yellow);cursor:pointer}
.milestone-item:hover{background:var(--bg-primary)}
.milestone-item time{display:block;font-size:0.7rem;color:var(--text-dim);margin-bottom:2px}
/* Activity heatmap */
.heatmap{display:flex;gap:2px;flex-wrap:wrap;margin-top:var(--space-sm)}
.heatmap-cell{width:14px;height:14px;border-radius:2px;background:var(--bg-tertiary)}
.heatmap-cell.l1{background:rgba(125,86,244,0.25)}
.heatmap-cell.l2{background:rgba(125,86,244,0.45)}
.heatmap-cell.l3{background:rgba(125,86,244,0.65)}
.heatmap-cell.l4{background:rgba(125,86,244,0.85)}
.heatmap-labels{display:flex;justify-content:space-between;font-family:var(--font-mono);font-size:0.65rem;color:var(--text-dim);margin-top:var(--space-xs)}
/* Present mode summary pane */
.present-mode .summary-pane{width:420px}
.present-mode .summary-pane h2{font-size:1.1rem}
.present-mode .summary-pane h3{font-size:0.85rem}
.present-mode .type-row,.present-mode .project-row,.present-mode .milestone-item{font-size:0.95rem}
/* Narrative Pane */
.narrative-pane{display:none;margin-bottom:var(--space-lg);padding:var(--space-lg);background:var(--bg-tertiary);border:1px solid var(--border);border-radius:var(--radius-lg);position:relative}
.narrative-pane.open{display:block}
.narrative-pane .narrative-header{display:flex;justify-content:space-between;align-items:center;margin-bottom:var(--space-md)}
.narrative-pane h3{font-family:var(--font-mono);font-size:0.85rem;font-weight:600;color:var(--text-primary);margin:0}
.narrative-pane .narrative-actions{display:flex;gap:var(--space-xs)}
.narrative-pane .narrative-actions button{padding:var(--space-xs) var(--space-sm);background:var(--bg-secondary);border:1px solid var(--border);border-radius:var(--radius);color:var(--text-muted);font-family:var(--font-mono);font-size:0.75rem;cursor:pointer}
.narrative-pane .narrative-actions button:hover{color:var(--text-primary);border-color:var(--text-muted)}
.narrative-pane .narrative-actions button.copied{color:var(--project-teal);border-color:var(--project-teal)}
.narrative-text{font-family:var(--font-mono);font-size:0.85rem;color:var(--text-secondary);line-height:1.8;white-space:pre-wrap}
.narrative-text .narrative-section{margin-top:var(--space-md)}
.narrative-text .narrative-section-title{color:var(--text-muted);font-weight:600;font-size:0.8rem}
.narrative-text .narrative-decisions{margin-top:var(--space-sm);padding-left:var(--space-md)}
.narrative-text .narrative-decisions li{margin-bottom:var(--space-xs);color:var(--text-secondary)}
.present-mode .narrative-pane{padding:var(--space-xl)}
.present-mode .narrative-pane h3{font-size:1rem}
.present-mode .narrative-text{font-size:1rem;line-height:1.9}
/* Diff mode */
.diff-mode .event.before-since{opacity:0.35}
.diff-mode .event.before-since:hover{opacity:0.6}
.diff-mode .event.after-since{border-left-width:4px}
.diff-mode .day-header.has-new::after{content:" (new)";color:var(--project-teal);font-weight:400}
.diff-marker{display:none;padding:var(--space-sm) var(--space-md);margin-bottom:var(--space-md);background:rgba(78,205,196,0.1);border:1px solid var(--project-teal);border-radius:var(--radius);font-family:var(--font-mono);font-size:0.8rem;color:var(--project-teal);text-align:center}
.diff-mode .diff-marker{display:flex;justify-content:space-between;align-items:center}
.diff-marker .diff-stats{color:var(--text-muted)}
.present-mode .diff-marker{font-size:0.95rem;padding:var(--space-md) var(--space-lg)}
/* Annotations */
.event.annotated{position:relative}
.event.annotated::after{content:"*";position:absolute;right:var(--space-sm);top:50%;transform:translateY(-50%);color:var(--project-teal);font-weight:700;font-size:1rem}
.present-mode .event.annotated::after{font-size:1.2rem}
.annotation-input{width:100%;margin-top:var(--space-md);padding:var(--space-sm);background:var(--bg-tertiary);border:1px solid var(--border);border-radius:var(--radius);color:var(--text-primary);font-family:var(--font-mono);font-size:0.85rem;resize:vertical;min-height:60px}
.annotation-input:focus{outline:none;border-color:var(--accent)}
.annotation-input::placeholder{color:var(--text-dim)}
.annotation-saved{font-family:var(--font-mono);font-size:0.75rem;color:var(--project-teal);margin-top:var(--space-xs)}
/* Export button */
.export-btn{padding:2px 10px;background:var(--bg-tertiary);border:1px solid var(--border);border-radius:var(--radius);color:var(--text-muted);font-family:var(--font-mono);font-size:0.75rem;cursor:pointer;transition:all 0.15s}
.export-btn:hover{color:var(--text-primary);border-color:var(--text-muted)}
.present-mode .export-btn{font-size:0.85rem;padding:4px 12px}
/* Loading overlay */
.loading-overlay{position:fixed;top:0;left:0;right:0;height:3px;background:transparent;z-index:200;pointer-events:none}
.loading-overlay.active{background:linear-gradient(90deg,transparent,var(--accent),transparent);animation:loading-slide 1s ease-in-out infinite}
@keyframes loading-slide{0%{background-position:-200% 0}100%{background-position:200% 0}}
/* Keyboard hints */
.kbd-hints{position:fixed;bottom:0;left:0;right:0;display:flex;justify-content:center;gap:var(--space-md);padding:var(--space-sm) var(--space-md);background:var(--bg-dark);border-top:1px solid var(--border);font-family:var(--font-mono);font-size:0.7rem;color:var(--text-dim);z-index:80;opacity:0;transition:opacity 0.2s}
.present-mode .kbd-hints{opacity:1}
.kbd-hints kbd{padding:1px 5px;background:var(--bg-tertiary);border:1px solid var(--border);border-radius:3px;color:var(--text-muted);font-size:0.65rem}
/* Print styles */
@media print{body{background:white!important;color:#222!important}.filters,.preset-bar,.kbd-hints,.summary-pane,.modal-backdrop{display:none!important}header{border-bottom:1px solid #ccc}.summary-bar{display:flex!important;background:#f5f5f5!important;border:1px solid #ccc!important;color:#222!important}.summary-bar .summary-value{color:#000!important}.event{border-left-color:#ccc!important;color:#222!important}.event .content{color:#444!important}.day-header{color:#666!important}.narrative-pane.open{display:block!important;background:#f5f5f5!important;border:1px solid #ccc!important;color:#222!important}}
/* Preset Bar */
.preset-bar{display:flex;gap:var(--space-xs);margin-bottom:var(--space-md);align-items:center;flex-wrap:wrap}
.preset-btn{padding:var(--space-xs) var(--space-md);background:var(--bg-secondary);border:1px solid var(--border);border-radius:var(--radius);color:var(--text-muted);font-family:var(--font-mono);font-size:0.8rem;cursor:pointer;transition:all 0.15s ease}
.preset-btn:hover{color:var(--text-primary);border-color:var(--text-muted)}
.preset-btn.active{background:var(--accent);border-color:var(--accent);color:white}
.preset-range{font-family:var(--font-mono);font-size:0.75rem;color:var(--text-dim);margin-left:var(--space-sm)}
.present-mode .preset-bar{margin-bottom:var(--space-lg)}
.present-mode .preset-btn{font-size:0.95rem;padding:var(--space-sm) var(--space-lg)}
.present-mode .preset-range{font-size:0.85rem}
/* Present Mode */
.present-mode{--text-primary:#f0e8fc;--text-secondary:#ddd0ee;--text-muted:#c4b5d9}
.present-mode .container{max-width:1100px}
.present-mode .container.list-view{max-width:1100px}
.present-mode header h1{font-size:2rem}
.present-mode .stats{font-size:1.1rem}
.present-mode .filters{display:none}
.present-mode .filters.show{display:flex}
.present-mode .event{font-size:1.05rem;padding:var(--space-md) var(--space-lg);grid-template-columns:60px 140px 45px 1fr}
.present-mode .event time{font-size:0.95rem}
.present-mode .event .project{font-size:0.95rem}
.present-mode .event .type-icon{font-size:0.9rem}
.present-mode .day-header{font-size:0.95rem;margin-top:var(--space-lg);padding:var(--space-sm) 0}
.present-mode .modal-content{max-width:750px;font-size:1rem}
.present-mode .modal-content dl{font-size:1rem}
.present-mode .empty-state{font-size:1.1rem}
/* Present mode: highlight milestones and decisions */
.present-mode .event.event-milestone{background:rgba(254,202,87,0.08);border-left-color:var(--project-yellow)!important}
.present-mode .event.event-decision{background:rgba(255,159,243,0.08);border-left-color:var(--project-pink)!important}
.present-mode .event.event-milestone .content,.present-mode .event.event-decision .content{color:var(--text-primary);font-weight:500}
/* Summary bar */
.summary-bar{display:none;gap:var(--space-lg);padding:var(--space-md) var(--space-lg);margin-bottom:var(--space-lg);background:var(--bg-tertiary);border:1px solid var(--border);border-radius:var(--radius-lg);font-family:var(--font-mono);font-size:0.85rem;color:var(--text-secondary);flex-wrap:wrap;align-items:center}
.present-mode .summary-bar{display:flex;font-size:1rem}
.summary-bar .summary-item{display:flex;align-items:center;gap:var(--space-xs)}
.summary-bar .summary-value{color:var(--text-primary);font-weight:600}
.summary-bar .summary-label{color:var(--text-muted)}
.summary-bar .summary-divider{width:1px;height:1.2em;background:var(--border)}
.summary-bar .milestone-list{display:flex;flex-wrap:wrap;gap:var(--space-xs);margin-left:var(--space-sm)}
.summary-bar .milestone-chip{padding:1px 8px;background:rgba(254,202,87,0.15);color:var(--project-yellow);border-radius:3px;font-size:0.8rem;white-space:nowrap;max-width:200px;overflow:hidden;text-overflow:ellipsis}
.present-mode .summary-bar .milestone-chip{font-size:0.9rem}
/* Keyboard nav: focused event */
.event.focused{background:var(--bg-tertiary);outline:2px solid var(--accent);outline-offset:-2px}
.present-mode .event.focused{outline-width:3px}
/* Present mode: vertical timeline */
.present-mode .v-timeline-content{font-size:1rem}
.present-mode .v-timeline-card{padding:var(--space-lg)}
.present-mode .v-timeline-meta time{font-size:0.85rem}
.present-mode .v-timeline-type{font-size:0.8rem}
/* Present mode: horizontal timeline */
.present-mode .h-timeline-lane-label{font-size:0.85rem;width:120px;min-width:120px}
.present-mode .h-timeline-event{width:16px;height:16px}
/* Replay View */
.replay-controls{position:sticky;top:0;z-index:50;display:flex;align-items:center;gap:var(--space-md);padding:var(--space-md);background:var(--bg-tertiary);border:1px solid var(--border);border-radius:var(--radius-lg);margin-bottom:var(--space-lg);font-family:var(--font-mono);font-size:0.85rem;flex-wrap:wrap;backdrop-filter:blur(8px);background:rgba(58,45,74,0.95)}
.replay-controls button{padding:var(--space-xs) var(--space-md);background:var(--bg-secondary);border:1px solid var(--border);border-radius:var(--radius);color:var(--text-secondary);font-family:var(--font-mono);font-size:0.85rem;cursor:pointer;transition:all 0.15s;min-width:36px}
.replay-controls button:hover{color:var(--text-primary);border-color:var(--text-muted)}
.replay-controls button.playing{background:var(--accent);border-color:var(--accent);color:white}
.replay-controls .replay-progress{flex:1;min-width:120px}
.replay-controls input[type="range"]{width:100%;accent-color:var(--accent);cursor:pointer}
.replay-controls .replay-step-label{color:var(--text-muted);font-size:0.8rem;white-space:nowrap}
.replay-controls select{padding:var(--space-xs) var(--space-sm);background:var(--bg-secondary);border:1px solid var(--border);border-radius:var(--radius);color:var(--text-primary);font-family:var(--font-mono);font-size:0.8rem}
.replay-stage{max-width:750px;margin:0 auto;padding:var(--space-md) 0}
.replay-scene{margin-bottom:var(--space-xl)}
.replay-scene-header{font-family:var(--font-mono);font-size:1.3rem;font-weight:600;color:var(--project-yellow);padding:var(--space-lg) 0 var(--space-md) 0;border-bottom:1px dashed var(--border);margin-bottom:var(--space-lg);opacity:0;transform:translateY(8px);transition:all 0.5s ease}
.replay-scene-header.visible{opacity:1;transform:translateY(0)}
.replay-scene-narration{font-family:var(--font-mono);font-style:italic;font-size:0.9rem;color:var(--text-muted);padding:var(--space-sm) 0 var(--space-md) var(--space-md);border-left:2px solid var(--border);opacity:0;transition:opacity 0.6s ease}
.replay-scene-narration.visible{opacity:1}
.replay-transition{font-family:var(--font-mono);font-style:italic;font-size:0.85rem;color:var(--text-dim);padding:var(--space-sm) 0 var(--space-sm) var(--space-lg);opacity:0;transition:opacity 0.5s ease}
.replay-transition.visible{opacity:1}
.replay-event-card{position:relative;margin-bottom:var(--space-md);padding:var(--space-md);padding-left:calc(var(--space-lg) + 16px);background:var(--bg-tertiary);border:1px solid var(--border);border-radius:var(--radius-lg);opacity:0;transform:translateY(12px);transition:all 0.5s ease}
.replay-event-card.visible{opacity:1;transform:translateY(0)}
.replay-event-card.past{opacity:0.5}
.replay-event-card.current{border-color:var(--accent);box-shadow:0 0 12px rgba(125,86,244,0.15)}
.replay-event-dot{position:absolute;left:var(--space-md);top:var(--space-md);width:12px;height:12px;border-radius:50%;border:2px solid var(--bg-secondary)}
.replay-event-meta{display:flex;justify-content:space-between;align-items:center;margin-bottom:var(--space-xs)}
.replay-event-meta time{font-family:var(--font-mono);font-size:0.75rem;color:var(--text-dim)}
.replay-event-type{font-family:var(--font-mono);font-size:0.7rem;font-weight:600;padding:1px 6px;border-radius:3px;background:var(--bg-secondary)}
.replay-event-content{font-family:var(--font-mono);font-size:0.875rem;color:var(--text-secondary);line-height:1.6;white-space:pre-wrap;word-break:break-word}
.replay-event-footer{display:flex;gap:var(--space-sm);margin-top:var(--space-sm);flex-wrap:wrap}
.replay-event-tag{font-family:var(--font-mono);font-size:0.7rem;color:var(--text-muted);background:var(--bg-secondary);padding:1px 6px;border-radius:3px;border:1px solid var(--border)}
.replay-event-hash{font-family:var(--font-mono);font-size:0.7rem;color:var(--text-dim)}
.replay-end{text-align:center;padding:var(--space-xl);font-family:var(--font-mono);color:var(--text-muted);font-style:italic;opacity:0;transition:opacity 0.8s ease}
.replay-end.visible{opacity:1}
.present-mode .replay-scene-header{font-size:1.6rem}
.present-mode .replay-scene-narration{font-size:1.05rem}
.present-mode .replay-transition{font-size:1rem}
.present-mode .replay-event-card{padding:var(--space-lg);padding-left:calc(var(--space-xl) + 16px)}
.present-mode .replay-event-content{font-size:1rem}
.present-mode .replay-controls{font-size:1rem}
.present-mode .replay-controls button{font-size:1rem;padding:var(--space-sm) var(--space-lg)}`

var htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>uroboro timeline</title>
  <script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
  <style>{{.CSS}}</style>
</head>
<body :class="{ 'present-mode': presentMode }" x-data="timeline()" x-init="init()" @keydown.window="handleKey($event)">
  <div class="loading-overlay" :class="{ active: loading }"></div>
  <div class="kbd-hints">
    <span><kbd>j</kbd><kbd>k</kbd> navigate</span>
    <span><kbd>1</kbd><kbd>2</kbd> views</span>
    <span><kbd>p</kbd> present</span>
    <span><kbd>f</kbd> filters</span>
    <span><kbd>n</kbd> summary</span>
    <span><kbd>N</kbd> narrative</span>
    <span><kbd>g</kbd> group</span>
    <span><kbd>/</kbd> search</span>
    <span><kbd>r</kbd> replay</span>
  </div>
  <div class="container" :class="{ 'list-view': viewMode === 'list' || viewMode === 'project', 'diff-mode': diffMode }">
    <header>
      <h1>
        <span x-show="viewMode !== 'project'">uroboro</span>
        <span x-show="viewMode === 'project'" x-text="selectedProject"></span>
      </h1>
      <button x-show="viewMode === 'project'" @click="backToMain()" style="background:var(--bg-tertiary);border:1px solid var(--border);border-radius:var(--radius);color:var(--text-secondary);font-family:var(--font-mono);font-size:0.875rem;padding:var(--space-xs) var(--space-sm);cursor:pointer">&larr; Back</button>
      <div class="stats">
        <span x-text="filteredEvents.length + ' events'"></span>
        <span x-text="data.projects?.length + ' projects'"></span>
        <button @click="downloadHTML()" class="export-btn" title="Export as HTML">Export</button>
      </div>
    </header>
    <!-- Summary Bar -->
    <div class="summary-bar">
      <div class="summary-item"><span class="summary-value" x-text="data.stats?.totalEvents || 0"></span><span class="summary-label">events</span></div>
      <div class="summary-divider"></div>
      <div class="summary-item"><span class="summary-value" x-text="data.projects?.length || 0"></span><span class="summary-label">projects</span></div>
      <div class="summary-divider"></div>
      <div class="summary-item"><span class="summary-value" x-text="summary.commitCount"></span><span class="summary-label">commits</span></div>
      <div class="summary-divider"></div>
      <div class="summary-item"><span class="summary-value" x-text="summary.decisionCount"></span><span class="summary-label">decisions</span></div>
      <div class="summary-divider"></div>
      <div class="summary-item"><span class="summary-value" x-text="data.milestones?.length || 0"></span><span class="summary-label">milestones</span></div>
      <template x-if="data.milestones?.length > 0">
        <div class="milestone-list">
          <template x-for="m in data.milestones.slice(0, 3)" :key="m.timestamp + m.content">
            <span class="milestone-chip" x-text="m.content"></span>
          </template>
        </div>
      </template>
    </div>
    <!-- Date Presets -->
    <div class="preset-bar" x-show="viewMode !== 'project'">
      <template x-for="preset in datePresets" :key="preset.label">
        <button class="preset-btn" :class="{ active: activePreset === preset.label }" @click="loadPreset(preset)" x-text="preset.label"></button>
      </template>
      <span class="preset-range" x-show="activePreset" x-text="presetRangeLabel"></span>
    </div>
    <!-- Diff Marker -->
    <div class="diff-marker">
      <span>Showing changes since <strong x-text="formatDateTime(new Date(sinceDate).toISOString())"></strong></span>
      <span class="diff-stats" x-text="filteredEvents.filter(e => new Date(e.timestamp).getTime() >= sinceDate).length + ' new events'"></span>
    </div>
    <!-- Narrative -->
    <div class="narrative-pane" :class="{ open: showNarrative }" x-show="viewMode !== 'project'">
      <div class="narrative-header">
        <h3>Narrative</h3>
        <div class="narrative-actions">
          <button @click="copyNarrative()" :class="{ copied: narrativeCopied }" x-text="narrativeCopied ? 'Copied' : 'Copy Markdown'"></button>
          <button @click="showNarrative = false">&times;</button>
        </div>
      </div>
      <div class="narrative-text" x-html="narrativeHTML"></div>
    </div>
    <div class="filters" :class="{ show: showFilters }" x-show="viewMode !== 'project' && (!presentMode || showFilters)">
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
              <template x-for="(event, idx) in group" :key="event.timestamp + event.content">
                <div class="event" :class="eventClasses(event, getGlobalIdx(event))" :style="'border-left-color: ' + getProjectColor(event.project)" @click="selectedEvent = event" :data-idx="getGlobalIdx(event)">
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
          <template x-for="(event, idx) in filteredEvents" :key="event.timestamp + event.content">
            <div class="event" :class="eventClasses(event, idx)" :style="'border-left-color: ' + getProjectColor(event.project)" @click="selectedEvent = event" :data-idx="idx">
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
    <!-- Replay View -->
    <div x-show="viewMode === 'replay'">
      <div class="replay-controls">
        <button @click="replayStepBack()" title="Step back">&larr;</button>
        <button @click="replayToggle()" :class="{ playing: replayPlaying }" x-text="replayPlaying ? '||' : '&#9654;'" :title="replayPlaying ? 'Pause' : 'Play'"></button>
        <button @click="replayStepForward()" title="Step forward">&rarr;</button>
        <span class="replay-step-label" x-text="replayStep + ' / ' + replayTotalSteps"></span>
        <div class="replay-progress">
          <input type="range" min="0" :max="replayTotalSteps" x-model.number="replayStep" @input="replayPause()">
        </div>
        <select x-model.number="replaySpeed" @change="if (replayPlaying) { replayPause(); replayPlay(); }">
          <option value="3000">0.5x</option>
          <option value="1500">1x</option>
          <option value="800">2x</option>
          <option value="400">4x</option>
        </select>
        <button @click="replayReset()">Reset</button>
      </div>
      <div class="replay-stage">
        <template x-for="(scene, si) in replayScenes" :key="'scene-' + si">
          <div class="replay-scene" x-show="scene.stepStart <= replayStep">
            <div class="replay-scene-header" :class="{ visible: scene.stepStart <= replayStep }" x-text="scene.date"></div>
            <div class="replay-scene-narration" :class="{ visible: scene.stepStart <= replayStep }" x-show="scene.narration" x-text="scene.narration"></div>
            <template x-for="(ev, ei) in scene.events" :key="'ev-' + si + '-' + ei">
              <div>
                <div class="replay-transition" :class="{ visible: ev.stepIdx <= replayStep }" x-show="ev.transition" x-text="ev.transition"></div>
                <div class="replay-event-card" :class="{ visible: ev.stepIdx <= replayStep, past: ev.stepIdx < replayStep - 1, current: ev.stepIdx === replayStep }" @click="selectedEvent = ev.event">
                  <div class="replay-event-dot" :class="'type-' + ev.event.eventType" :style="'background:' + getProjectColor(ev.event.project)"></div>
                  <div class="replay-event-meta">
                    <time x-text="formatDateTime(ev.event.timestamp)"></time>
                    <span class="replay-event-type" :class="'type-' + ev.event.eventType" x-text="getTypeIcon(ev.event.eventType)"></span>
                  </div>
                  <div class="replay-event-content" x-text="ev.event.content"></div>
                  <div class="replay-event-footer">
                    <span class="replay-event-tag" x-show="ev.event.project" :style="'color:' + getProjectColor(ev.event.project)" x-text="ev.event.project"></span>
                    <template x-for="tag in (ev.event.tags || []).filter(t => t !== 'git' && t !== 'commit')" :key="tag">
                      <span class="replay-event-tag" x-text="tag"></span>
                    </template>
                    <span class="replay-event-hash" x-show="ev.event.gitHash" x-text="ev.event.gitHash ? ev.event.gitHash.slice(0, 8) : ''"></span>
                  </div>
                </div>
              </div>
            </template>
          </div>
        </template>
        <div class="replay-end" :class="{ visible: replayStep >= replayTotalSteps }" x-show="replayStep >= replayTotalSteps">fin.</div>
      </div>
    </div>
    <!-- Summary Pane -->
    <div class="summary-pane" :class="{ open: showSummaryPane }">
      <button class="close-btn" @click="showSummaryPane = false">&times;</button>
      <h2>Summary</h2>
      <h3>Events by Type</h3>
      <div class="type-breakdown">
        <template x-for="t in typeCounts" :key="t.type">
          <div class="type-row" x-show="t.count > 0">
            <span class="type-dot" :style="'background:' + t.color"></span>
            <span class="type-name" x-text="t.type"></span>
            <span class="type-count" x-text="t.count"></span>
          </div>
        </template>
      </div>
      <h3>Projects</h3>
      <div class="project-list">
        <template x-for="p in data.projects || []" :key="p.name">
          <div class="project-row" @click="showProject(p.name); showSummaryPane = false">
            <span class="project-dot" :style="'background:' + getProjectColor(p.name)"></span>
            <span class="project-name" x-text="p.name"></span>
            <span class="project-count" x-text="p.eventCount"></span>
          </div>
        </template>
      </div>
      <h3>Milestones</h3>
      <div class="milestone-list-pane">
        <template x-for="m in (data.milestones || [])" :key="m.timestamp + m.content">
          <div class="milestone-item" @click="selectedEvent = m; showSummaryPane = false">
            <time x-text="formatDateTime(m.timestamp)"></time>
            <span x-text="m.content"></span>
          </div>
        </template>
        <div x-show="!data.milestones?.length" style="font-family:var(--font-mono);font-size:0.8rem;color:var(--text-dim)">No milestones in this period.</div>
      </div>
      <h3>Activity</h3>
      <div class="heatmap">
        <template x-for="cell in heatmapCells" :key="cell.date">
          <div class="heatmap-cell" :class="cell.level" :title="cell.date + ': ' + cell.count + ' events'"></div>
        </template>
      </div>
      <div class="heatmap-labels">
        <span x-text="heatmapCells.length > 0 ? heatmapCells[0].date : ''"></span>
        <span x-text="heatmapCells.length > 0 ? heatmapCells[heatmapCells.length - 1].date : ''"></span>
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
        <textarea class="annotation-input" placeholder="Add a meeting note..." :value="selectedEvent ? (annotations[eventKey(selectedEvent)] || '') : ''" @input="if (selectedEvent) { if ($event.target.value) annotations[eventKey(selectedEvent)] = $event.target.value; else delete annotations[eventKey(selectedEvent)]; }"></textarea>
        <div class="annotation-saved" x-show="selectedEvent && annotations[eventKey(selectedEvent)]">Annotation saved (session only)</div>
        <button @click="selectedEvent = null">Close</button>
      </div>
    </div>
  </div>
  <script>
    window.JOURNEY_DATA = {{.DataJSON}};
    window.PRESENT_MODE = {{.PresentMode}};
    window.SINCE_DATE = '{{.SinceDate}}';
    window.REPLAY_MODE = {{.ReplayMode}};
    function timeline() {
      return {
        data: window.JOURNEY_DATA || { events: [], projects: [], stats: {}, milestones: [] },
        presentMode: window.PRESENT_MODE || false,
        showFilters: false,
        focusedIdx: -1,
        activePreset: null,
        loading: false,
        showSummaryPane: false,
        showNarrative: false,
        narrativeCopied: false,
        diffMode: !!window.SINCE_DATE,
        sinceDate: window.SINCE_DATE ? new Date(window.SINCE_DATE).getTime() : 0,
        annotations: {},
        datePresets: [
          { label: 'Today', days: 1 },
          { label: 'This Week', days: 7 },
          { label: '2 Weeks', days: 14 },
          { label: '30 Days', days: 30 },
        ],
        replayStep: 0,
        replayPlaying: false,
        replaySpeed: 1500,
        replayTimer: null,
        projectFilter: '', typeFilter: '', searchQuery: '', groupByDay: true, selectedEvent: null,
        viewMode: window.REPLAY_MODE ? 'replay' : 'list',
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
        get summary() {
          const events = this.data.events || [];
          return {
            commitCount: events.filter(e => e.eventType === 'commit').length,
            decisionCount: events.filter(e => e.eventType === 'decision').length,
            featureCount: events.filter(e => e.eventType === 'feature').length,
            bugfixCount: events.filter(e => e.eventType === 'bugfix').length,
          };
        },
        get typeCounts() {
          const events = this.data.events || [];
          const typeColors = {
            commit: '#626262', capture: '#a78bbd', decision: '#FF9FF3',
            milestone: '#FECA57', feature: '#4ECDC4', bugfix: '#FF6B6B',
            learning: '#54A0FF', blocker: '#FF6B6B', question: '#FECA57'
          };
          return this.eventTypes.map(t => ({
            type: t,
            count: events.filter(e => e.eventType === t).length,
            color: typeColors[t] || '#626262'
          }));
        },
        get heatmapCells() {
          const events = this.data.events || [];
          if (events.length === 0) return [];
          const dayCounts = {};
          events.forEach(e => {
            const d = new Date(e.timestamp).toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
            dayCounts[d] = (dayCounts[d] || 0) + 1;
          });
          const start = this.data.timeline?.startTime ? new Date(this.data.timeline.startTime) : new Date();
          const end = this.data.timeline?.endTime ? new Date(this.data.timeline.endTime) : new Date();
          const cells = [];
          for (let d = new Date(start); d <= end; d.setDate(d.getDate() + 1)) {
            const label = d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
            const count = dayCounts[label] || 0;
            let level = '';
            if (count >= 20) level = 'l4';
            else if (count >= 10) level = 'l3';
            else if (count >= 4) level = 'l2';
            else if (count >= 1) level = 'l1';
            cells.push({ date: label, count, level });
          }
          return cells;
        },
        get narrativeMarkdown() {
          const events = this.data.events || [];
          const projects = this.data.projects || [];
          const milestones = this.data.milestones || [];
          if (events.length === 0) return 'No events in this period.';

          const days = this.activePreset || 'the selected period';
          const topProjects = projects.slice(0, 3).map(p => p.name);
          const decisions = events.filter(e => e.eventType === 'decision');
          const commits = events.filter(e => e.eventType === 'commit');
          const features = events.filter(e => e.eventType === 'feature');
          const bugfixes = events.filter(e => e.eventType === 'bugfix');

          let md = '';
          // Opening line
          if (topProjects.length === 1) {
            md += 'Focused on **' + topProjects[0] + '**. ';
          } else if (topProjects.length > 1) {
            md += 'Worked across **' + topProjects.join('**, **') + '**. ';
          }
          md += events.length + ' events';
          if (commits.length > 0) md += ', ' + commits.length + ' commits';
          if (projects.length > 0) md += ' across ' + projects.length + ' projects';
          md += '.';

          // Features and fixes
          if (features.length > 0 || bugfixes.length > 0) {
            md += '\n\n';
            if (features.length > 0) md += features.length + ' feature' + (features.length > 1 ? 's' : '');
            if (features.length > 0 && bugfixes.length > 0) md += ', ';
            if (bugfixes.length > 0) md += bugfixes.length + ' fix' + (bugfixes.length > 1 ? 'es' : '');
            md += '.';
          }

          // Milestones
          if (milestones.length > 0) {
            md += '\n\n**Milestones:**\n';
            milestones.slice(0, 5).forEach(m => {
              md += '- ' + m.content + '\n';
            });
          }

          // Key decisions
          if (decisions.length > 0) {
            md += '\n**Key decisions:**\n';
            decisions.slice(0, 5).forEach(d => {
              md += '- ' + d.content + '\n';
            });
          }

          // Annotations
          const annotationKeys = Object.keys(this.annotations);
          if (annotationKeys.length > 0) {
            md += '\n**Meeting notes:**\n';
            annotationKeys.forEach(key => {
              const note = this.annotations[key];
              const eventContent = key.split('|')[1] || 'event';
              md += '- ' + eventContent + ': *' + note + '*\n';
            });
          }

          return md.trim();
        },
        get narrativeHTML() {
          // Simple markdown-to-html: bold, newlines, list items
          return this.narrativeMarkdown
            .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
            .replace(/^- (.+)$/gm, '<li>$1</li>')
            .replace(/(<li>.*<\/li>\n?)+/g, '<ul>$&</ul>')
            .replace(/\n\n/g, '<br><br>')
            .replace(/\n/g, '<br>');
        },
        downloadHTML() {
          const css = document.querySelector('style').textContent;
          const dataJSON = JSON.stringify(this.data);
          const scriptContent = document.querySelector('script:last-of-type').textContent
            .replace(/window\.JOURNEY_DATA\s*=\s*[^;]+;/, 'window.JOURNEY_DATA = ' + dataJSON + ';')
            .replace(/window\.PRESENT_MODE\s*=\s*[^;]+;/, 'window.PRESENT_MODE = ' + this.presentMode + ';');
          const html = '<!DOCTYPE html><html lang="en"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0"><title>uroboro timeline</title><script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"><\/script><style>' + css + '</style></head><body :class="{ \'present-mode\': presentMode }" x-data="timeline()" x-init="init()" @keydown.window="handleKey($event)"><div class="loading-overlay" :class="{ active: loading }"></div><div class="kbd-hints"><span><kbd>j</kbd><kbd>k</kbd> navigate</span><span><kbd>1</kbd><kbd>2</kbd> views</span><span><kbd>p</kbd> present</span><span><kbd>f</kbd> filters</span><span><kbd>n</kbd> summary</span><span><kbd>N</kbd> narrative</span><span><kbd>g</kbd> group</span><span><kbd>/</kbd> search</span></div>' + document.querySelector('.container').outerHTML.replace(/@click="downloadHTML\(\)"/g, 'style="display:none"') + '<script>' + scriptContent + '<\/script></body></html>';
          const blob = new Blob([html], { type: 'text/html' });
          const url = URL.createObjectURL(blob);
          const a = document.createElement('a');
          a.href = url;
          a.download = 'uroboro-timeline-' + new Date().toISOString().slice(0, 10) + '.html';
          a.click();
          URL.revokeObjectURL(url);
        },
        async copyNarrative() {
          try {
            await navigator.clipboard.writeText(this.narrativeMarkdown);
            this.narrativeCopied = true;
            setTimeout(() => { this.narrativeCopied = false; }, 2000);
          } catch (e) {
            console.error('Copy failed:', e);
          }
        },
        eventKey(event) {
          return (event.timestamp || '') + '|' + (event.content || '').slice(0, 50);
        },
        eventClasses(event, idx) {
          let cls = 'event-' + event.eventType;
          if (idx >= 0 && this.focusedIdx === idx) cls += ' focused';
          if (this.diffMode && this.sinceDate) {
            cls += new Date(event.timestamp).getTime() >= this.sinceDate ? ' after-since' : ' before-since';
          }
          if (this.annotations[this.eventKey(event)]) cls += ' annotated';
          return cls;
        },
        getGlobalIdx(event) {
          return this.filteredEvents.indexOf(event);
        },
        handleKey(e) {
          if (this.selectedEvent) return; // modal open, let Esc handle it
          if (e.key === 'Escape' && this.viewMode === 'replay') { this.replayPause(); this.viewMode = 'list'; return; }
          if (e.key === 'Escape' && this.showSummaryPane) { this.showSummaryPane = false; return; }
          if (e.target.tagName === 'INPUT' || e.target.tagName === 'SELECT') return;
          if (this.viewMode === 'replay') {
            switch(e.key) {
              case ' ': e.preventDefault(); this.replayToggle(); break;
              case 'ArrowRight': case 'l': e.preventDefault(); this.replayStepForward(); break;
              case 'ArrowLeft': case 'h': e.preventDefault(); this.replayStepBack(); break;
              case 'r': this.replayPause(); this.viewMode = 'list'; break;
            }
            return;
          }
          switch(e.key) {
            case 'j': case 'ArrowDown':
              e.preventDefault();
              if (this.viewMode === 'list') {
                this.focusedIdx = Math.min(this.focusedIdx + 1, this.filteredEvents.length - 1);
                this.scrollToFocused();
              }
              break;
            case 'k': case 'ArrowUp':
              e.preventDefault();
              if (this.viewMode === 'list') {
                this.focusedIdx = Math.max(this.focusedIdx - 1, 0);
                this.scrollToFocused();
              }
              break;
            case 'Enter':
              if (this.focusedIdx >= 0 && this.focusedIdx < this.filteredEvents.length) {
                this.selectedEvent = this.filteredEvents[this.focusedIdx];
              }
              break;
            case '1': this.viewMode = 'list'; break;
            case '2': this.viewMode = 'timeline'; break;
            case '3': if (this.selectedProject) this.viewMode = 'project'; break;
            case 'p': this.presentMode = !this.presentMode; break;
            case 'f': this.showFilters = !this.showFilters; break;
            case '/':
              e.preventDefault();
              this.showFilters = true;
              this.$nextTick(() => {
                const input = document.querySelector('.filters input[type="search"]');
                if (input) input.focus();
              });
              break;
            case 'g': this.groupByDay = !this.groupByDay; break;
            case 'n': this.showSummaryPane = !this.showSummaryPane; break;
            case 'N': this.showNarrative = !this.showNarrative; break;
            case 'd': if (this.sinceDate) this.diffMode = !this.diffMode; break;
            case 'r': this.viewMode = 'replay'; this.replayStep = 0; break;
          }
        },
        get presetRangeLabel() {
          if (!this.data.timeline?.startTime || !this.data.timeline?.endTime) return '';
          const s = new Date(this.data.timeline.startTime);
          const e = new Date(this.data.timeline.endTime);
          const fmt = d => d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
          return fmt(s) + ' — ' + fmt(e);
        },
        async loadPreset(preset) {
          if (this.loading) return;
          this.activePreset = preset.label;
          this.loading = true;
          try {
            const resp = await fetch('/api/journey?days=' + preset.days);
            const newData = await resp.json();
            this.data = newData;
            this.focusedIdx = -1;
            this.computeTimeRange();
          } catch (e) {
            console.error('Failed to load preset:', e);
          }
          this.loading = false;
        },
        // Replay: computed properties
        get replayEvents() {
          if (!this.data.events) return [];
          return [...this.data.events].sort((a, b) => new Date(a.timestamp) - new Date(b.timestamp));
        },
        get replayScenes() {
          const events = this.replayEvents;
          if (events.length === 0) return [];
          const GAP = 4 * 60 * 60 * 1000; // 4 hours = new scene
          const scenes = [];
          let stepCounter = 0;
          let prevSceneEvent = null;

          for (let i = 0; i < events.length; i++) {
            const ev = events[i];
            const ts = new Date(ev.timestamp).getTime();
            const lastScene = scenes[scenes.length - 1];
            const needsNew = !lastScene || (ts - lastScene.lastTs > GAP);

            if (needsNew) {
              const narration = this._sceneNarration(ev, prevSceneEvent, lastScene);
              const date = new Date(ev.timestamp).toLocaleDateString('en-US', { weekday: 'long', month: 'long', day: 'numeric', year: 'numeric' });
              scenes.push({ date, narration, events: [], stepStart: stepCounter, lastTs: ts });
              stepCounter++; // scene header is a step
            }

            const scene = scenes[scenes.length - 1];
            const prevEvent = scene.events.length > 0 ? scene.events[scene.events.length - 1].event : (scenes.length > 1 ? scenes[scenes.length - 2].events.slice(-1)[0]?.event : null);
            const transition = this._eventTransition(prevEvent, ev);
            scene.events.push({ event: ev, transition, stepIdx: stepCounter });
            scene.lastTs = ts;
            stepCounter++;
            prevSceneEvent = ev;
          }
          return scenes;
        },
        get replayTotalSteps() {
          const scenes = this.replayScenes;
          if (scenes.length === 0) return 0;
          const last = scenes[scenes.length - 1];
          if (last.events.length === 0) return last.stepStart;
          return last.events[last.events.length - 1].stepIdx;
        },
        _sceneNarration(firstEvent, prevSceneLastEvent, prevScene) {
          if (!prevScene) {
            const typeNames = { decision: 'a decision', question: 'a question', blocker: 'an obstacle', commit: 'code', feature: 'a feature', bugfix: 'a fix', milestone: 'a milestone', capture: 'a thought', learning: 'a discovery' };
            return 'It started with ' + (typeNames[firstEvent.eventType] || 'a thought') + '.';
          }
          if (prevSceneLastEvent) {
            const gap = new Date(firstEvent.timestamp).getTime() - new Date(prevSceneLastEvent.timestamp).getTime();
            const dur = this._humanDuration(gap);
            if (dur) return 'After ' + dur + ' of silence\u2026';
          }
          return 'A new chapter.';
        },
        _eventTransition(prev, curr) {
          if (!prev) return null;
          const gap = new Date(curr.timestamp).getTime() - new Date(prev.timestamp).getTime();
          const pt = prev.eventType, ct = curr.eventType;

          // Type-based transitions
          if (pt === 'question' && ct === 'decision') return 'An answer took shape.';
          if (pt === 'decision' && ct === 'commit') {
            const d = this._humanDuration(gap);
            return d ? 'Code followed ' + d + ' later.' : 'Code followed.';
          }
          if (ct === 'blocker') return 'Then things hit a wall.';
          if (pt === 'blocker' && ct !== 'blocker') return 'Back in motion.';
          if (ct === 'milestone') return 'A milestone.';
          if (pt === 'question' && ct === 'commit') return 'Straight to code.';
          if (pt === 'commit' && ct === 'decision') return 'A pause to reconsider.';
          if (pt === 'decision' && ct === 'decision') return 'Another choice.';
          if (pt === 'commit' && ct === 'blocker') return 'Progress met resistance.';

          // Gap-based transitions
          if (gap > 2 * 60 * 60 * 1000) {
            const d = this._humanDuration(gap);
            return d + ' passed.';
          }

          // Consecutive commits — no transition needed
          if (pt === 'commit' && ct === 'commit' && gap < 30 * 60 * 1000) return null;

          return null;
        },
        _humanDuration(ms) {
          const mins = Math.floor(ms / 60000);
          if (mins < 2) return null;
          if (mins < 60) return mins + ' minutes';
          const hours = Math.floor(mins / 60);
          if (hours < 24) return hours + ' hour' + (hours > 1 ? 's' : '');
          const days = Math.floor(hours / 24);
          return days + ' day' + (days > 1 ? 's' : '');
        },
        // Replay: controls
        replayToggle() {
          if (this.replayPlaying) { this.replayPause(); } else { this.replayPlay(); }
        },
        replayPlay() {
          if (this.replayStep >= this.replayTotalSteps) this.replayStep = 0;
          this.replayPlaying = true;
          this.replayTimer = setInterval(() => {
            if (this.replayStep >= this.replayTotalSteps) {
              this.replayPause();
              return;
            }
            this.replayStep++;
            this._replayScrollToCurrent();
          }, this.replaySpeed);
        },
        replayPause() {
          this.replayPlaying = false;
          if (this.replayTimer) { clearInterval(this.replayTimer); this.replayTimer = null; }
        },
        replayStepForward() {
          this.replayPause();
          if (this.replayStep < this.replayTotalSteps) {
            this.replayStep++;
            this._replayScrollToCurrent();
          }
        },
        replayStepBack() {
          this.replayPause();
          if (this.replayStep > 0) {
            this.replayStep--;
            this._replayScrollToCurrent();
          }
        },
        replayReset() {
          this.replayPause();
          this.replayStep = 0;
          window.scrollTo({ top: 0, behavior: 'smooth' });
        },
        _replayScrollToCurrent() {
          this.$nextTick(() => {
            const el = document.querySelector('.replay-event-card.current');
            if (!el) return;
            const rect = el.getBoundingClientRect();
            const inView = rect.top >= 0 && rect.bottom <= window.innerHeight;
            if (!inView) el.scrollIntoView({ block: 'nearest', behavior: 'smooth' });
          });
        },
        scrollToFocused() {
          this.$nextTick(() => {
            const el = document.querySelector('.event.focused');
            if (el) el.scrollIntoView({ block: 'nearest', behavior: 'smooth' });
          });
        },
        showProject(name) { this.selectedProject = name; this.viewMode = 'project'; },
        backToMain() { this.viewMode = 'list'; this.selectedProject = null; }
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
	present := fs.Bool("present", false, "Present mode (large text, minimal chrome)")
	export := fs.String("export", "", "Export static HTML to file and exit")
	since := fs.String("since", "", "Diff mode: dim events before this date (YYYY-MM-DD or YYYY-MM-DDTHH:MM)")
	repos := fs.String("repos", "", "Additional git repos (comma-separated paths, optional :name suffix)")
	replay := fs.Bool("replay", false, "Start in replay mode")
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

	var repoConfigs []journey.RepoConfig
	if *repos != "" {
		for _, r := range strings.Split(*repos, ",") {
			r = strings.TrimSpace(r)
			if r == "" {
				continue
			}
			parts := strings.SplitN(r, ":", 2)
			rc := journey.RepoConfig{Path: parts[0]}
			if len(parts) == 2 {
				rc.Name = parts[1]
			}
			repoConfigs = append(repoConfigs, rc)
		}
	}

	svc := journey.NewService(db)

	// Static HTML export mode
	if *export != "" {
		data, err := svc.GenerateJourney(journey.Options{Days: *days, Projects: projects, Repos: repoConfigs})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Journey error: %v\n", err)
			os.Exit(1)
		}
		html, err := GenerateStandaloneHTML(data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Template error: %v\n", err)
			os.Exit(1)
		}
		if err := os.WriteFile(*export, []byte(html), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Write error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Exported to %s\n", *export)
		return
	}

	// Serve timeline page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		data, err := svc.GenerateJourney(journey.Options{Days: *days, Projects: projects, Repos: repoConfigs})
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
			"CSS":         template.CSS(timelineCSS),
			"DataJSON":    template.JS(string(jsonBytes)),
			"PresentMode": *present,
			"SinceDate":   *since,
			"ReplayMode":  *replay,
		})
	})

	// API endpoint for live refresh (accepts ?days=N override)
	http.HandleFunc("/api/journey", func(w http.ResponseWriter, r *http.Request) {
		reqDays := *days
		if d := r.URL.Query().Get("days"); d != "" {
			if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 {
				reqDays = parsed
			}
		}
		data, err := svc.GenerateJourney(journey.Options{Days: reqDays, Projects: projects, Repos: repoConfigs})
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
		"CSS":         template.CSS(timelineCSS),
		"DataJSON":    template.JS(string(jsonBytes)),
		"PresentMode": false,
		"SinceDate":   "",
		"ReplayMode":  false,
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
