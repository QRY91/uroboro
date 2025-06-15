<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import Timeline from './components/Timeline.svelte';
  import {
    timelineState,
    timelineActions,
    journeyData,
    isLoading,
    error,
    autoSaveViewport,
    loadSavedViewport,
    eventsInCurrentViewport,
    controls,
    selectedEvent,
    timelineStatistics,
    viewport,
  } from './stores/timeline';
  import type { JourneyEvent } from './types/timeline';

  // App state
  let appContainer: HTMLDivElement;
  let selectedEventDetail: JourneyEvent | null = null;
  let showEventDetail = false;
  let unsubscribeAutoSave: (() => void) | null = null;

  // 3-Pane Layout State
  let consolePanelCollapsed = false;
  let controlsPanelCollapsed = false;
  let selectedProjects: Record<string, boolean> = {};
  let selectedEventTypes: Record<string, boolean> = {};
  let eventTypes = ['capture', 'commit', 'milestone', 'learning', 'decision', 'error', 'success'];

  // Theme system
  let currentTheme = 'dark';
  let themes = [
    { value: 'dark', label: 'Dark' },
    { value: 'light', label: 'Light' },
    { value: 'matrix', label: 'Matrix' },
    { value: 'neon', label: 'Neon' },
  ];

  // Error handling
  let lastError: string | null = null;
  let showErrorDetail = false;

  onMount(async () => {
    // Initialize theme from localStorage or system preference
    initializeTheme();

    // Set up auto-save for viewport position
    unsubscribeAutoSave = autoSaveViewport();

    // Load saved viewport position
    loadSavedViewport();

    // Load initial journey data
    await timelineActions.loadJourneyData({ days: 7 });

    // Set up global error handling
    window.addEventListener('error', handleGlobalError);
    window.addEventListener('unhandledrejection', handleUnhandledRejection);

    // Set up visibility change handling for performance
    document.addEventListener('visibilitychange', handleVisibilityChange);
  });

  onDestroy(() => {
    if (unsubscribeAutoSave) {
      unsubscribeAutoSave();
    }

    window.removeEventListener('error', handleGlobalError);
    window.removeEventListener('unhandledrejection', handleUnhandledRejection);
    document.removeEventListener('visibilitychange', handleVisibilityChange);
  });

  function initializeTheme() {
    const savedTheme = localStorage.getItem('uroboro-timeline-theme');
    const systemPrefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;

    currentTheme = savedTheme || (systemPrefersDark ? 'dark' : 'light');
    applyTheme(currentTheme);

    // Listen for system theme changes
    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', e => {
      if (!savedTheme) {
        currentTheme = e.matches ? 'dark' : 'light';
        applyTheme(currentTheme);
      }
    });
  }

  function applyTheme(theme: string) {
    document.documentElement.setAttribute('data-theme', theme);
    localStorage.setItem('uroboro-timeline-theme', theme);
  }

  function handleThemeChange(e: Event) {
    const select = e.target as HTMLSelectElement;
    currentTheme = select.value;
    applyTheme(currentTheme);
  }

  function handleGlobalError(event: ErrorEvent) {
    console.error('Global error:', event.error);
    lastError = event.error?.message || 'An unexpected error occurred';
    showErrorDetail = true;
  }

  function handleUnhandledRejection(event: PromiseRejectionEvent) {
    console.error('Unhandled promise rejection:', event.reason);
    lastError = event.reason?.message || 'An unexpected promise rejection occurred';
    showErrorDetail = true;
  }

  function handleVisibilityChange() {
    if (document.hidden) {
      // Page is hidden - pause any animations or intensive operations
      timelineActions.pause();
    }
  }

  function handleEventClick(event: CustomEvent) {
    const { event: journeyEvent } = event.detail;
    selectedEventDetail = journeyEvent;
    showEventDetail = true;
    timelineActions.selectEvent(journeyEvent.id);
  }

  function handleEventHover(event: CustomEvent) {
    const { event: journeyEvent, hovered } = event.detail;
    if (hovered) {
      timelineActions.hoverEvent(journeyEvent.id);
    } else {
      timelineActions.hoverEvent(null);
    }
  }

  function closeEventDetail() {
    showEventDetail = false;
    selectedEventDetail = null;
    timelineActions.selectEvent(null);
  }

  function closeError() {
    showErrorDetail = false;
    lastError = null;
  }

  function handleRetry() {
    closeError();
    timelineActions.loadJourneyData({ days: 7 });
  }

  function formatEventTime(timestamp: string): string {
    return new Date(timestamp).toLocaleString();
  }

  function getEventTypeLabel(type: string): string {
    const labels = {
      milestone: 'Milestone',
      learning: 'Learning',
      decision: 'Decision',
      commit: 'Commit',
      capture: 'Capture',
      error: 'Error',
      success: 'Success',
      context_switch: 'Context Switch',
    };
    return labels[type as keyof typeof labels] || type;
  }

  // 3-Pane Layout Functions
  function applyFilters() {
    // Apply project and event type filters
    timelineActions.setProjectFilter(Object.keys(selectedProjects).filter(p => selectedProjects[p]));
    timelineActions.setEventTypeFilter(Object.keys(selectedEventTypes).filter(t => selectedEventTypes[t]));
  }

  function getEventTypeColor(type: string): string {
    const colors = {
      capture: '#4ecdc4',
      commit: '#ff6b6b',
      milestone: '#ffd93d',
      learning: '#6bcf7f',
      decision: '#a8e6cf',
      error: '#ff8a80',
      success: '#69f0ae',
      context_switch: '#ba68c8',
    };
    return colors[type as keyof typeof colors] || '#00ffff';
  }

  function getEventTypeIcon(type: string): string {
    const icons = {
      capture: '📸',
      commit: '📝',
      milestone: '🎯',
      learning: '💡',
      decision: '🎲',
      error: '⚠️',
      success: '✅',
      context_switch: '🔄',
    };
    return icons[type as keyof typeof icons] || '📌';
  }

  function getProjectColor(project: string): string {
    const colors = ['#ff6b6b', '#4ecdc4', '#45b7d1', '#96ceb4', '#ffeaa7', '#dda0dd', '#98d8c8'];
    let hash = 0;
    for (let i = 0; i < project.length; i++) {
      hash = project.charCodeAt(i) + ((hash << 5) - hash);
    }
    return colors[Math.abs(hash) % colors.length];
  }

  function togglePlayback() {
    if ($controls?.isPlaying) {
      timelineActions.pause();
    } else {
      timelineActions.play();
    }
  }

  function handleScaleChange() {
    if ($controls?.currentScale) {
      timelineActions.setViewportScale($controls.currentScale);
    }
  }

  function handleSpeedChange() {
    if ($controls?.playSpeed) {
      timelineActions.setPlaySpeed($controls.playSpeed);
    }
  }

  // Reactive statements with loading protection
  $: hasData = !$isLoading && !!$journeyData && !!$journeyData.events;
  $: currentError = $error || lastError;
</script>

<svelte:head>
  <title>Uroboro Journey Timeline - Development Replay</title>
  <meta
    name="description"
    content="Interactive timeline visualization for development journey replay with video-editor-style scrubbing controls" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0, user-scalable=no" />

  <!-- Preload critical resources -->
  <link rel="preload" href="/fonts/inter.woff2" as="font" type="font/woff2" crossorigin />

  <!-- Theme color for mobile browsers -->
  <meta name="theme-color" content="#4ecdc4" />

  <!-- Prevent zoom on iOS -->
  <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no" />
</svelte:head>

<div class="app" bind:this={appContainer} data-theme={currentTheme}>
  <!-- App Header -->
  <header class="app-header">
    <div class="app-title">
      <h1>🐍 Uroboro Journey Timeline</h1>
      <p>Development journey replay with timeline scrubbing</p>
    </div>

    <div class="app-controls">
      <div class="theme-selector">
        <label for="themeSelect">Theme:</label>
        <select id="themeSelect" value={currentTheme} on:change={handleThemeChange}>
          {#each themes as theme}
            <option value={theme.value}>{theme.label}</option>
          {/each}
        </select>
      </div>

      {#if hasData}
        <div class="data-info">
          <span class="data-stats">
            {$journeyData?.events?.length || 0} events • {$journeyData?.projects?.length || 0} projects
          </span>
        </div>
      {/if}
    </div>
  </header>

  <!-- Main Timeline Area -->
  <main class="app-main">
    {#if $isLoading}
      <div class="loading-screen">
        <div class="loading-animation">
          <div class="loading-spinner"></div>
          <h2>Loading Journey Timeline...</h2>
          <p>Gathering your development story...</p>
        </div>
      </div>
    {:else if currentError}
      <div class="error-screen">
        <div class="error-content">
          <h2>⚠️ Timeline Error</h2>
          <p>{currentError}</p>
          <div class="error-actions">
            <button class="btn-primary" on:click={handleRetry}>Retry</button>
            <button class="btn-secondary" on:click={closeError}>Dismiss</button>
          </div>
        </div>
      </div>
    {:else if hasData}
      <!-- 3-Pane Layout: Timeline + Event Console + Controls -->
      <div class="three-pane-layout">
        <!-- Main Timeline Panel -->
        <div class="timeline-panel">
          <Timeline on:eventClick={handleEventClick} on:eventHover={handleEventHover} />
        </div>

        <!-- Event Console Panel -->
        <div class="console-panel">
          <div class="console-header">
            <h3>📋 Event Console</h3>
            <button class="console-toggle" on:click={() => (consolePanelCollapsed = !consolePanelCollapsed)}>
              {consolePanelCollapsed ? '▶' : '◀'}
            </button>
          </div>

          {#if !consolePanelCollapsed}
            <div class="console-content">
              <!-- Project Filters -->
              <div class="filter-section">
                <h4>🎯 Project Filters</h4>
                <div class="project-filters">
                  {#each $journeyData?.projects || [] as project}
                    <label class="filter-checkbox">
                      <input
                        type="checkbox"
                        bind:checked={selectedProjects[project.name]}
                        on:change={() => applyFilters()} />
                      <span class="project-badge" style="--project-color: {project.color}">
                        {project.name}
                      </span>
                      <span class="event-count">({project.eventCount})</span>
                    </label>
                  {/each}
                </div>
              </div>

              <!-- Event Type Filters -->
              <div class="filter-section">
                <h4>🏷️ Event Types</h4>
                <div class="type-filters">
                  {#each eventTypes as type}
                    <label class="filter-checkbox">
                      <input type="checkbox" bind:checked={selectedEventTypes[type]} on:change={() => applyFilters()} />
                      <span class="type-badge" style="--type-color: {getEventTypeColor(type)}">
                        {getEventTypeIcon(type)}
                        {type}
                      </span>
                    </label>
                  {/each}
                </div>
              </div>

              <!-- Event List -->
              <div class="event-list-section">
                <h4>📅 Events ({$eventsInCurrentViewport?.length || 0})</h4>
                <div class="event-list">
                  {#each $eventsInCurrentViewport || [] as event}
                    <div
                      class="event-item"
                      class:selected={$selectedEvent?.id === event.id}
                      on:click={() => handleEventClick({ detail: { event } })}
                      on:keydown={e => e.key === 'Enter' && handleEventClick({ detail: { event } })}
                      role="button"
                      tabindex="0">
                      <div class="event-item-header">
                        <span class="event-icon">{getEventTypeIcon(event.type)}</span>
                        <span class="event-project" style="color: {getProjectColor(event.project)}">
                          {event.project}
                        </span>
                        <span class="event-time">{formatEventTime(event.timestamp)}</span>
                      </div>
                      <div class="event-content-preview">
                        {event.content ? event.content.substring(0, 100) : ''}{event.content && event.content.length > 100 ? '...' : ''}
                      </div>
                      {#if event.tags && event.tags.length > 0}
                        <div class="event-tags-preview">
                          {#each event.tags.slice(0, 3) as tag}
                            <span class="tag-mini">{tag}</span>
                          {/each}
                          {#if event.tags.length > 3}
                            <span class="tag-more">+{event.tags.length - 3}</span>
                          {/if}
                        </div>
                      {/if}
                    </div>
                  {/each}
                </div>
              </div>
            </div>
          {/if}
        </div>

        <!-- Controls Panel -->
        <div class="controls-panel">
          <div class="controls-header">
            <h3>⚙️ Timeline Controls</h3>
            <button class="controls-toggle" on:click={() => (controlsPanelCollapsed = !controlsPanelCollapsed)}>
              {controlsPanelCollapsed ? '▼' : '▲'}
            </button>
          </div>

          {#if !controlsPanelCollapsed}
            <div class="controls-content">
              <!-- Time Scale Controls -->
              <div class="control-group">
                <label>📏 Time Scale</label>
                <select bind:value={$controls.currentScale} on:change={handleScaleChange}>
                  <option value="15m">15 Minutes</option>
                  <option value="1h">1 Hour</option>
                  <option value="6h">6 Hours</option>
                  <option value="24h">24 Hours</option>
                  <option value="7d">7 Days</option>
                  <option value="full">Full Timeline</option>
                </select>
              </div>

              <!-- Playback Controls -->
              <div class="control-group">
                <label>▶️ Playback</label>
                <div class="playback-controls">
                  <button class="btn-primary" on:click={togglePlayback}>
                    {$controls?.isPlaying ? '⏸️' : '▶️'}
                  </button>
                  <button class="btn-secondary" on:click={() => timelineActions.restart()}> ⏮️ </button>
                </div>
              </div>

              <!-- Speed Control -->
              <div class="control-group">
                <label>🚀 Speed: {$controls?.playSpeed || 1}x</label>
                <input
                  type="range"
                  min="0.25"
                  max="4"
                  step="0.25"
                  bind:value={$controls.playSpeed}
                  on:input={handleSpeedChange} />
              </div>

              <!-- Statistics -->
              <div class="control-group">
                <label>📊 Statistics</label>
                <div class="stats-grid">
                  <div class="stat-item">
                    <span class="stat-label">Total Events</span>
                    <span class="stat-value">{$timelineStatistics?.totalEvents || 0}</span>
                  </div>
                  <div class="stat-item">
                    <span class="stat-label">Projects</span>
                    <span class="stat-value">{$timelineStatistics?.projectCount || 0}</span>
                  </div>
                  <div class="stat-item">
                    <span class="stat-label">Milestones</span>
                    <span class="stat-value">{$timelineStatistics?.milestoneCount || 0}</span>
                  </div>
                </div>
              </div>
            </div>
          {/if}
        </div>
      </div>
    {:else}
      <div class="empty-state">
        <div class="empty-content">
          <h2>No Journey Data</h2>
          <p>No timeline events found. Start capturing your development journey!</p>
          <button class="btn-primary" on:click={() => timelineActions.loadJourneyData()}>Load Journey</button>
        </div>
      </div>
    {/if}
  </main>

  <!-- Event Detail Modal -->
  {#if showEventDetail && selectedEventDetail}
    <div
      class="modal-overlay"
      on:click={closeEventDetail}
      role="button"
      tabindex="0"
      on:keydown={e => e.key === 'Escape' && closeEventDetail()}>
      <div class="event-detail-modal" on:click|stopPropagation role="dialog" aria-modal="true">
        <header class="modal-header">
          <h3>{getEventTypeLabel(selectedEventDetail.type)}</h3>
          <button class="close-btn" on:click={closeEventDetail} aria-label="Close"> × </button>
        </header>

        <div class="modal-content">
          <div class="event-meta">
            <div class="meta-item">
              <span class="meta-label">Time:</span>
              <span class="meta-value">{formatEventTime(selectedEventDetail.timestamp)}</span>
            </div>
            <div class="meta-item">
              <span class="meta-label">Project:</span>
              <span class="meta-value">{selectedEventDetail.project}</span>
            </div>
            <div class="meta-item">
              <span class="meta-label">Type:</span>
              <span class="meta-value">{getEventTypeLabel(selectedEventDetail.type)}</span>
            </div>
          </div>

          <div class="event-content-text">
            <h4>Content</h4>
            <p>{selectedEventDetail.content}</p>

            {#if selectedEventDetail.description}
              <h4>Description</h4>
              <p>{selectedEventDetail.description}</p>
            {/if}
          </div>

          {#if selectedEventDetail.tags && selectedEventDetail.tags.length > 0}
            <div class="event-tags-section">
              <h4>Tags</h4>
              <div class="tags-list">
                {#each selectedEventDetail.tags as tag}
                  <span class="tag">{tag}</span>
                {/each}
              </div>
            </div>
          {/if}

          {#if selectedEventDetail.metadata}
            <details class="event-metadata">
              <summary>Metadata</summary>
              <pre>{JSON.stringify(selectedEventDetail.metadata, null, 2)}</pre>
            </details>
          {/if}
        </div>
      </div>
    </div>
  {/if}

  <!-- Error Detail Modal -->
  {#if showErrorDetail && lastError}
    <div class="modal-overlay" on:click={closeError} role="button" tabindex="0">
      <div class="error-detail-modal" on:click|stopPropagation role="dialog" aria-modal="true">
        <header class="modal-header">
          <h3>⚠️ Error Details</h3>
          <button class="close-btn" on:click={closeError} aria-label="Close"> × </button>
        </header>

        <div class="modal-content">
          <p class="error-message">{lastError}</p>
          <div class="error-actions">
            <button class="btn-primary" on:click={handleRetry}> Retry </button>
            <button class="btn-secondary" on:click={closeError}> Close </button>
          </div>
        </div>
      </div>
    </div>
  {/if}
</div>

<style>
  :global(html) {
    height: 100%;
    font-family:
      'Inter',
      -apple-system,
      BlinkMacSystemFont,
      'Segoe UI',
      Roboto,
      sans-serif;
  }

  :global(body) {
    margin: 0;
    padding: 0;
    height: 100%;
    overflow: hidden;
  }

  :global(*, *::before, *::after) {
    box-sizing: border-box;
  }

  .app {
    height: 100vh;
    display: flex;
    flex-direction: column;
    background: var(--bg-primary);
    color: var(--text-primary);
    overflow: hidden;
  }

  .app-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem 1.5rem;
    background: var(--bg-secondary);
    border-bottom: 1px solid var(--border-color);
    backdrop-filter: blur(10px);
    z-index: 100;
  }

  .app-title h1 {
    margin: 0;
    font-size: 1.5rem;
    font-weight: 700;
    color: var(--accent-color);
  }

  .app-title p {
    margin: 0.25rem 0 0 0;
    font-size: 0.875rem;
    color: var(--text-secondary);
    opacity: 0.8;
  }

  .app-controls {
    display: flex;
    align-items: center;
    gap: 1.5rem;
  }

  .theme-selector {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .theme-selector label {
    font-size: 0.875rem;
    color: var(--text-secondary);
  }

  .theme-selector select {
    background: var(--bg-tertiary);
    border: 1px solid var(--border-color);
    color: var(--text-primary);
    padding: 0.5rem;
    border-radius: 4px;
    font-size: 0.875rem;
  }

  .data-info {
    padding: 0.5rem 1rem;
    background: var(--bg-tertiary);
    border-radius: 6px;
    border: 1px solid var(--border-color);
  }

  .data-stats {
    font-size: 0.875rem;
    color: var(--text-secondary);
  }

  .app-main {
    flex: 1;
    position: relative;
    overflow: hidden;
  }

  .loading-screen,
  .error-screen,
  .empty-state {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--bg-primary);
  }

  .loading-animation,
  .error-content,
  .empty-content {
    text-align: center;
    max-width: 400px;
    padding: 2rem;
  }

  .loading-spinner {
    width: 48px;
    height: 48px;
    border: 4px solid var(--bg-tertiary);
    border-top: 4px solid var(--accent-color);
    border-radius: 50%;
    animation: spin 1s linear infinite;
    margin: 0 auto 1.5rem;
  }

  .loading-animation h2,
  .error-content h2,
  .empty-content h2 {
    margin: 0 0 1rem 0;
    font-size: 1.5rem;
    color: var(--text-primary);
  }

  .loading-animation p,
  .error-content p,
  .empty-content p {
    margin: 0 0 1.5rem 0;
    color: var(--text-secondary);
    line-height: 1.5;
  }

  .error-actions {
    display: flex;
    gap: 1rem;
    justify-content: center;
  }

  .btn-primary,
  .btn-secondary {
    padding: 0.75rem 1.5rem;
    border-radius: 6px;
    border: none;
    cursor: pointer;
    font-size: 0.875rem;
    font-weight: 500;
    transition: all 0.2s ease;
  }

  .btn-primary {
    background: var(--accent-color);
    color: var(--bg-primary);
  }

  .btn-primary:hover {
    background: var(--accent-hover);
    transform: translateY(-1px);
  }

  .btn-secondary {
    background: var(--bg-tertiary);
    color: var(--text-primary);
    border: 1px solid var(--border-color);
  }

  .btn-secondary:hover {
    background: var(--bg-quaternary);
  }

  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.8);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    backdrop-filter: blur(4px);
  }

  .event-detail-modal,
  .error-detail-modal {
    background: var(--bg-secondary);
    border: 1px solid var(--border-color);
    border-radius: 12px;
    max-width: 600px;
    max-height: 80vh;
    width: 90vw;
    overflow: hidden;
    box-shadow: 0 20px 40px rgba(0, 0, 0, 0.3);
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1.5rem;
    border-bottom: 1px solid var(--border-color);
    background: var(--bg-tertiary);
  }

  .modal-header h3 {
    margin: 0;
    font-size: 1.25rem;
    color: var(--text-primary);
  }

  .close-btn {
    background: none;
    border: none;
    font-size: 1.5rem;
    color: var(--text-secondary);
    cursor: pointer;
    padding: 0.25rem;
    line-height: 1;
    transition: color 0.2s ease;
  }

  .close-btn:hover {
    color: var(--text-primary);
  }

  .modal-content {
    padding: 1.5rem;
    overflow-y: auto;
    max-height: calc(80vh - 120px);
  }

  .event-meta {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 1rem;
    margin-bottom: 1.5rem;
    padding: 1rem;
    background: var(--bg-tertiary);
    border-radius: 8px;
  }

  .meta-item {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .meta-label {
    font-size: 0.75rem;
    font-weight: 600;
    color: var(--text-secondary);
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .meta-value {
    font-size: 0.875rem;
    color: var(--text-primary);
  }

  .event-content-text h4 {
    margin: 0 0 0.5rem 0;
    font-size: 1rem;
    color: var(--accent-color);
  }

  .event-content-text p {
    margin: 0 0 1.5rem 0;
    line-height: 1.6;
    color: var(--text-primary);
  }

  .event-tags-section h4 {
    margin: 0 0 0.75rem 0;
    font-size: 1rem;
    color: var(--accent-color);
  }

  .tags-list {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
  }

  .tag {
    background: var(--bg-tertiary);
    color: var(--accent-color);
    padding: 0.25rem 0.75rem;
    border-radius: 16px;
    font-size: 0.75rem;
    border: 1px solid var(--accent-color);
  }

  .event-metadata {
    margin-top: 1.5rem;
  }

  .event-metadata summary {
    cursor: pointer;
    font-weight: 600;
    color: var(--accent-color);
    margin-bottom: 0.5rem;
  }

  .event-metadata pre {
    background: var(--bg-primary);
    padding: 1rem;
    border-radius: 6px;
    overflow-x: auto;
    font-size: 0.8rem;
    color: var(--text-secondary);
    border: 1px solid var(--border-color);
  }

  .error-message {
    background: var(--bg-tertiary);
    padding: 1rem;
    border-radius: 6px;
    border-left: 4px solid var(--error-color);
    margin-bottom: 1.5rem;
    font-family: monospace;
    font-size: 0.875rem;
  }

  @keyframes spin {
    0% {
      transform: rotate(0deg);
    }
    100% {
      transform: rotate(360deg);
    }
  }

  /* Theme definitions */
  :global([data-theme='dark']) {
    --bg-primary: #1a1a1a;
    --bg-secondary: #2a2a2a;
    --bg-tertiary: #3a3a3a;
    --bg-quaternary: #4a4a4a;
    --text-primary: #ffffff;
    --text-secondary: #cccccc;
    --accent-color: #4ecdc4;
    --accent-hover: #45b7b8;
    --border-color: #333333;
    --error-color: #ff6b6b;
  }

  :global([data-theme='light']) {
    --bg-primary: #ffffff;
    --bg-secondary: #f8f9fa;
    --bg-tertiary: #e9ecef;
    --bg-quaternary: #dee2e6;
    --text-primary: #212529;
    --text-secondary: #6c757d;
    --accent-color: #20c997;
    --accent-hover: #1aa179;
    --border-color: #dee2e6;
    --error-color: #dc3545;
  }

  :global([data-theme='matrix']) {
    --bg-primary: #000000;
    --bg-secondary: #001100;
    --bg-tertiary: #002200;
    --bg-quaternary: #003300;
    --text-primary: #00ff00;
    --text-secondary: #008800;
    --accent-color: #00ff41;
    --accent-hover: #00cc33;
    --border-color: #004400;
    --error-color: #ff0040;
  }

  :global([data-theme='neon']) {
    --bg-primary: #0a0a0a;
    --bg-secondary: #1a0a1a;
    --bg-tertiary: #2a1a2a;
    --bg-quaternary: #3a2a3a;
    --text-primary: #ff00ff;
    --text-secondary: #cc00cc;
    --accent-color: #00ffff;
    --accent-hover: #00cccc;
    --border-color: #440044;
    --error-color: #ff0080;
  }

  /* 3-Pane Layout Styles */
  .three-pane-layout {
    display: grid;
    grid-template-columns: 1fr 350px;
    grid-template-rows: 1fr auto;
    grid-template-areas:
      'timeline console'
      'timeline controls';
    height: 100%;
    gap: 1rem;
    padding: 1rem;
    background: var(--bg-primary);
  }

  .timeline-panel {
    grid-area: timeline;
    background: var(--bg-secondary);
    border: 1px solid var(--border-color);
    border-radius: 12px;
    overflow: hidden;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  }

  .console-panel {
    grid-area: console;
    background: var(--bg-secondary);
    border: 1px solid var(--border-color);
    border-radius: 12px;
    display: flex;
    flex-direction: column;
    max-height: 60vh;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
    transition: all 0.3s ease;
  }

  .controls-panel {
    grid-area: controls;
    background: var(--bg-secondary);
    border: 1px solid var(--border-color);
    border-radius: 12px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
    transition: all 0.3s ease;
  }

  .console-header,
  .controls-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem 1.5rem;
    background: var(--bg-tertiary);
    border-bottom: 1px solid var(--border-color);
    border-radius: 12px 12px 0 0;
  }

  .console-header h3,
  .controls-header h3 {
    margin: 0;
    font-size: 1.1rem;
    color: var(--text-primary);
    font-weight: 600;
  }

  .console-toggle,
  .controls-toggle {
    background: none;
    border: none;
    color: var(--accent-color);
    font-size: 1.2rem;
    cursor: pointer;
    padding: 0.25rem;
    border-radius: 4px;
    transition: all 0.2s ease;
  }

  .console-toggle:hover,
  .controls-toggle:hover {
    background: var(--bg-quaternary);
    transform: scale(1.1);
  }

  .console-content,
  .controls-content {
    flex: 1;
    overflow-y: auto;
    padding: 1rem;
  }

  /* Filter Sections */
  .filter-section {
    margin-bottom: 1.5rem;
  }

  .filter-section h4 {
    margin: 0 0 0.75rem 0;
    font-size: 0.9rem;
    color: var(--accent-color);
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .project-filters,
  .type-filters {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .filter-checkbox {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    cursor: pointer;
    padding: 0.5rem;
    border-radius: 6px;
    transition: all 0.2s ease;
  }

  .filter-checkbox:hover {
    background: var(--bg-tertiary);
    transform: translateX(2px);
  }

  .filter-checkbox input {
    accent-color: var(--accent-color);
  }

  .project-badge,
  .type-badge {
    padding: 0.25rem 0.75rem;
    border-radius: 16px;
    font-size: 0.8rem;
    font-weight: 500;
    background: var(--project-color, var(--type-color));
    color: #000;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
  }

  .event-count {
    font-size: 0.75rem;
    color: var(--text-secondary);
    opacity: 0.8;
  }

  /* Event List */
  .event-list-section h4 {
    margin: 0 0 1rem 0;
    font-size: 0.9rem;
    color: var(--accent-color);
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .event-list {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    max-height: 300px;
    overflow-y: auto;
  }

  .event-item {
    padding: 0.75rem;
    background: var(--bg-tertiary);
    border: 1px solid var(--border-color);
    border-radius: 8px;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .event-item:hover {
    background: var(--bg-quaternary);
    border-color: var(--accent-color);
    transform: translateY(-1px);
    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2);
  }

  .event-item.selected {
    border-color: var(--accent-color);
    background: var(--bg-quaternary);
    box-shadow: 0 0 0 2px rgba(78, 205, 196, 0.3);
  }

  .event-item-header {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin-bottom: 0.5rem;
  }

  .event-icon {
    font-size: 1rem;
  }

  .event-project {
    font-weight: 500;
    font-size: 0.8rem;
  }

  .event-time {
    margin-left: auto;
    font-size: 0.7rem;
    color: var(--text-secondary);
  }

  .event-content-preview {
    font-size: 0.8rem;
    color: var(--text-primary);
    line-height: 1.4;
    margin-bottom: 0.5rem;
  }

  .event-tags-preview {
    display: flex;
    gap: 0.25rem;
    flex-wrap: wrap;
  }

  .tag-mini {
    padding: 0.125rem 0.5rem;
    background: rgba(78, 205, 196, 0.2);
    color: var(--accent-color);
    border-radius: 10px;
    font-size: 0.65rem;
    border: 1px solid rgba(78, 205, 196, 0.3);
  }

  .tag-more {
    padding: 0.125rem 0.5rem;
    background: var(--bg-quaternary);
    color: var(--text-secondary);
    border-radius: 10px;
    font-size: 0.65rem;
  }

  /* Controls Panel */
  .control-group {
    margin-bottom: 1.5rem;
  }

  .control-group label {
    display: block;
    margin-bottom: 0.5rem;
    font-size: 0.9rem;
    font-weight: 600;
    color: var(--accent-color);
  }

  .control-group select,
  .control-group input[type='range'] {
    width: 100%;
    padding: 0.5rem;
    background: var(--bg-tertiary);
    border: 1px solid var(--border-color);
    border-radius: 6px;
    color: var(--text-primary);
    font-size: 0.9rem;
  }

  .control-group select:focus,
  .control-group input:focus {
    outline: none;
    border-color: var(--accent-color);
    box-shadow: 0 0 0 2px rgba(78, 205, 196, 0.2);
  }

  .playback-controls {
    display: flex;
    gap: 0.5rem;
  }

  .stats-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 0.75rem;
  }

  .stat-item {
    background: var(--bg-tertiary);
    padding: 0.75rem;
    border-radius: 6px;
    border: 1px solid var(--border-color);
  }

  .stat-label {
    display: block;
    font-size: 0.7rem;
    color: var(--text-secondary);
    text-transform: uppercase;
    letter-spacing: 0.5px;
    margin-bottom: 0.25rem;
  }

  .stat-value {
    display: block;
    font-size: 1.2rem;
    font-weight: 600;
    color: var(--accent-color);
  }

  /* Responsive 3-Pane Layout */
  @media (max-width: 1200px) {
    .three-pane-layout {
      grid-template-columns: 1fr 300px;
    }
  }

  @media (max-width: 1024px) {
    .three-pane-layout {
      grid-template-columns: 1fr;
      grid-template-rows: 1fr auto auto;
      grid-template-areas:
        'timeline'
        'console'
        'controls';
    }

    .console-panel,
    .controls-panel {
      max-height: 40vh;
    }
  }

  @media (max-width: 768px) {
    .three-pane-layout {
      padding: 0.5rem;
      gap: 0.5rem;
    }

    .console-header,
    .controls-header {
      padding: 0.75rem 1rem;
    }

    .console-content,
    .controls-content {
      padding: 0.75rem;
    }

    .stats-grid {
      grid-template-columns: 1fr;
    }
  }

  /* Mobile responsive */
  @media (max-width: 768px) {
    .app-header {
      flex-direction: column;
      gap: 1rem;
      padding: 1rem;
    }

    .app-controls {
      flex-direction: column;
      gap: 1rem;
      width: 100%;
    }

    .event-detail-modal,
    .error-detail-modal {
      margin: 1rem;
      width: calc(100vw - 2rem);
      max-height: calc(100vh - 2rem);
    }

    .modal-content {
      max-height: calc(100vh - 200px);
    }

    .event-meta {
      grid-template-columns: 1fr;
    }
  }
</style>
