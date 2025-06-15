import { writable, readable, derived, get } from 'svelte/store';
import type {
  TimelineState,
  TimelineViewport,
  TimelineControls,
  JourneyData,
  JourneyEvent,
  TimeScaleName,
  TimelineFilter,
  SearchResult,
  EventCluster,
  RenderContext,
  TimelineConfig,
} from '@types/timeline';
import { DEFAULT_TIMELINE_CONFIG, TIME_SCALES } from '@types/timeline';

// Core Timeline State
export const timelineState = writable<TimelineState>({
  // Data
  journeyData: null,
  loading: false,
  error: null,

  // Viewport
  viewport: {
    startTime: null,
    endTime: null,
    scale: '24h',
    position: 0,
    eventsInView: [],
  },
  controls: {
    isPlaying: false,
    playSpeed: 1,
    currentScale: '24h',
    viewportPosition: 0,
    zoomLevel: 100,
  },

  // UI State
  selectedEvent: null,
  hoveredEvent: null,
  searchResults: [],
  activeFilters: {
    projects: [],
    eventTypes: [],
    tags: [],
    dateRange: { start: null, end: null },
    searchQuery: '',
  },

  // Performance
  renderContext: {
    viewport: {
      startTime: null,
      endTime: null,
      scale: '24h',
      position: 0,
      eventsInView: [],
    },
    bounds: { left: 0, right: 0, top: 0, bottom: 0, width: 0, height: 0 },
    devicePixelRatio: 1,
    isLowSpec: false,
    performanceLevel: 'high',
  },
  eventClusters: [],

  // Interaction
  currentInteraction: null,
  lastUpdate: Date.now(),
});

// Configuration
export const timelineConfig = writable<TimelineConfig>(DEFAULT_TIMELINE_CONFIG);

// Derived stores for commonly accessed data with safe defaults
export const journeyData = derived(timelineState, $state => $state.journeyData || null);
export const viewport = derived(timelineState, $state => $state.viewport || { startTime: null, endTime: null, scale: 'full' });
export const controls = derived(timelineState, $state => $state.controls || {
  currentScale: 'full',
  isPlaying: false,
  playSpeed: 1,
  position: 0
});
export const selectedEvent = derived(timelineState, $state => $state.selectedEvent || null);
export const hoveredEvent = derived(timelineState, $state => $state.hoveredEvent || null);
export const isLoading = derived(timelineState, $state => $state.loading || false);
export const error = derived(timelineState, $state => $state.error || null);

// Derived computed values with safe null checks
export const eventsInCurrentViewport = derived([timelineState], ([$state]) => {
  if (!$state || !$state.journeyData || !$state.journeyData.events ||
      !$state.viewport || !$state.viewport.startTime || !$state.viewport.endTime) {
    return [];
  }

  const startTime = $state.viewport.startTime.getTime();
  const endTime = $state.viewport.endTime.getTime();

  return $state.journeyData.events.filter(event => {
    if (!event || !event.timestamp) return false;
    const eventTime = new Date(event.timestamp).getTime();
    return eventTime >= startTime && eventTime <= endTime;
  });
});

export const timelineStatistics = derived([journeyData], ([$data]) => {
  if (!$data || !$data.events || !Array.isArray($data.events)) {
    return {
      totalEvents: 0,
      projectCount: 0,
      milestoneCount: 0,
      projectCounts: {},
      eventTypeCounts: {}
    };
  }

  const projectCounts = new Map<string, number>();
  const eventTypeCounts = new Map<string, number>();
  let milestoneCount = 0;

  $data.events.forEach(event => {
    if (!event) return;
    projectCounts.set(event.project || 'unknown', (projectCounts.get(event.project || 'unknown') || 0) + 1);
    eventTypeCounts.set(event.eventType || 'unknown', (eventTypeCounts.get(event.eventType || 'unknown') || 0) + 1);
    if (event.eventType === 'milestone') milestoneCount++;
  });

  return {
    totalEvents: $data.events.length,
    projectCount: projectCounts.size,
    milestoneCount,
    projectCounts: Object.fromEntries(projectCounts),
    eventTypeCounts: Object.fromEntries(eventTypeCounts),
    timespan: $data.timeline.totalDuration,
    avgEventsPerDay: $data.stats.avgEventsPerDay,
  };
});

export const filteredEvents = derived([timelineState], ([$state]) => {
  if (!$state.journeyData) return [];

  let events = $state.journeyData.events;
  const filters = $state.activeFilters;

  // Apply project filter
  if (filters.projects.length > 0) {
    events = events.filter(event => event && event.project && filters.projects.includes(event.project));
  }

  // Apply event type filter
  if (filters.eventTypes.length > 0) {
    events = events.filter(event => event && event.eventType && filters.eventTypes.includes(event.eventType));
  }

  // Apply tag filter
  if (filters.tags.length > 0) {
    events = events.filter(event => event.tags && filters.tags.some(tag => event.tags.includes(tag)));
  }

  // Apply date range filter
  if (filters.dateRange.start || filters.dateRange.end) {
    events = events.filter(event => {
      if (!event || !event.timestamp) return false;
      const eventTime = new Date(event.timestamp);
      if (filters.dateRange.start && eventTime < filters.dateRange.start) return false;
      if (filters.dateRange.end && eventTime > filters.dateRange.end) return false;
      return true;
    });
  }

  // Apply search query
  if (filters.searchQuery.trim()) {
    const query = filters.searchQuery.toLowerCase();
    events = events.filter(
      event =>
        event &&
        ((event.content && event.content.toLowerCase().includes(query)) ||
         (event.project && event.project.toLowerCase().includes(query)) ||
         (event.tags && event.tags.some(tag => tag && tag.toLowerCase().includes(query))))
    );
  }

  return events;
});

// Actions for updating timeline state
export const timelineActions = {
  // Data management
  async loadJourneyData(params?: { days?: number; projects?: string[] }) {
    timelineState.update(state => ({ ...state, loading: true, error: null }));

    try {
      const queryParams = new URLSearchParams();
      if (params?.days) queryParams.set('days', params.days.toString());
      if (params?.projects) queryParams.set('projects', params.projects.join(','));

      const response = await fetch(`/api/journey?${queryParams}`);
      if (!response.ok) throw new Error(`HTTP ${response.status}: ${response.statusText}`);

      const data: JourneyData = await response.json();

      timelineState.update(state => ({
        ...state,
        journeyData: data,
        loading: false,
        // Don't set viewport times here - let setViewportScale handle it
      }));

      // Always initialize viewport scale to avoid race condition
      timelineActions.setViewportScale('24h');
    } catch (err) {
      timelineState.update(state => ({
        ...state,
        loading: false,
        error: err instanceof Error ? err.message : 'Failed to load journey data',
      }));
    }
  },

  // Viewport management
  setViewportScale(scale: TimeScaleName) {
    const state = get(timelineState);
    if (!state.journeyData) return;

    const timeScale = TIME_SCALES[scale];
    const journeyStart = new Date(state.journeyData.timeline.startTime);
    const journeyEnd = new Date(state.journeyData.timeline.endTime);
    const fullDuration = journeyEnd.getTime() - journeyStart.getTime();

    let startTime: Date;
    let endTime: Date;

    if (scale === 'full') {
      startTime = journeyStart;
      endTime = journeyEnd;
    } else {
      const viewportDuration = timeScale.duration!;
      const maxPosition = fullDuration - viewportDuration;
      const currentOffset = maxPosition * state.viewport.position;

      startTime = new Date(journeyStart.getTime() + currentOffset);
      endTime = new Date(startTime.getTime() + viewportDuration);

      // Ensure we don't go beyond journey bounds
      if (endTime > journeyEnd) {
        endTime = journeyEnd;
        startTime = new Date(endTime.getTime() - viewportDuration);
      }
      if (startTime < journeyStart) {
        startTime = journeyStart;
        endTime = new Date(startTime.getTime() + viewportDuration);
      }
    }

    timelineState.update(state => ({
      ...state,
      viewport: {
        ...state.viewport,
        scale,
        startTime,
        endTime,
      },
      controls: {
        ...state.controls,
        currentScale: scale,
      },
    }));
  },

  setViewportPosition(position: number) {
    const state = get(timelineState);
    if (!state.journeyData) return;

    const clampedPosition = Math.max(0, Math.min(1, position));
    const journeyStart = new Date(state.journeyData.timeline.startTime);
    const journeyEnd = new Date(state.journeyData.timeline.endTime);
    const fullDuration = journeyEnd.getTime() - journeyStart.getTime();

    if (state.viewport.scale === 'full') {
      // Full timeline - position doesn't change viewport bounds
      return;
    }

    const timeScale = TIME_SCALES[state.viewport.scale];
    const viewportDuration = timeScale.duration!;
    const maxPosition = fullDuration - viewportDuration;
    const currentOffset = maxPosition * clampedPosition;

    const startTime = new Date(journeyStart.getTime() + currentOffset);
    const endTime = new Date(startTime.getTime() + viewportDuration);

    timelineState.update(state => ({
      ...state,
      viewport: {
        ...state.viewport,
        position: clampedPosition,
        startTime,
        endTime,
      },
      controls: {
        ...state.controls,
        viewportPosition: clampedPosition,
      },
    }));
  },

  panViewport(delta: number) {
    const state = get(timelineState);
    if (!state.journeyData || state.viewport.scale === 'full') return;

    const currentPosition = state.viewport.position;
    const newPosition = currentPosition + delta;
    timelineActions.setViewportPosition(newPosition);
  },

  // Playback controls
  play() {
    timelineState.update(state => ({
      ...state,
      controls: { ...state.controls, isPlaying: true },
    }));
    timelineActions.startPlaybackAnimation();
  },

  pause() {
    timelineState.update(state => ({
      ...state,
      controls: { ...state.controls, isPlaying: false },
    }));
    timelineActions.stopPlaybackAnimation();
  },

  setPlaySpeed(speed: number) {
    const clampedSpeed = Math.max(0.5, Math.min(3, speed));
    timelineState.update(state => ({
      ...state,
      controls: { ...state.controls, playSpeed: clampedSpeed },
    }));
  },

  restart() {
    timelineActions.setViewportPosition(0);
    timelineState.update(state => ({
      ...state,
      controls: { ...state.controls, isPlaying: false },
    }));
    timelineActions.stopPlaybackAnimation();
  },

  // Playback animation
  playbackAnimationId: null as number | null,

  startPlaybackAnimation() {
    if (timelineActions.playbackAnimationId) return; // Already running

    const animate = () => {
      const state = get(timelineState);
      if (!state.controls.isPlaying) return;

      const currentPosition = state.controls.viewportPosition;
      const speed = state.controls.playSpeed;
      const increment = 0.002 * speed; // Adjust speed as needed
      const newPosition = currentPosition + increment;

      if (newPosition >= 1) {
        // Reached the end - pause playback
        timelineActions.pause();
        return;
      }

      timelineActions.setViewportPosition(newPosition);
      timelineActions.playbackAnimationId = requestAnimationFrame(animate);
    };

    timelineActions.playbackAnimationId = requestAnimationFrame(animate);
  },

  stopPlaybackAnimation() {
    if (timelineActions.playbackAnimationId) {
      cancelAnimationFrame(timelineActions.playbackAnimationId);
      timelineActions.playbackAnimationId = null;
    }
  },

  // Event selection
  selectEvent(eventId: string | null) {
    timelineState.update(state => ({
      ...state,
      selectedEvent: eventId,
    }));
  },

  hoverEvent(eventId: string | null) {
    timelineState.update(state => ({
      ...state,
      hoveredEvent: eventId,
    }));
  },

  // Search and filtering
  setSearchQuery(query: string) {
    timelineState.update(state => ({
      ...state,
      activeFilters: { ...state.activeFilters, searchQuery: query },
    }));
  },

  setProjectFilter(projects: string[]) {
    timelineState.update(state => ({
      ...state,
      activeFilters: { ...state.activeFilters, projects },
    }));
  },

  setEventTypeFilter(eventTypes: string[]) {
    timelineState.update(state => ({
      ...state,
      activeFilters: { ...state.activeFilters, eventTypes: eventTypes as any[] },
    }));
  },

  setTagFilter(tags: string[]) {
    timelineState.update(state => ({
      ...state,
      activeFilters: { ...state.activeFilters, tags },
    }));
  },

  setDateRangeFilter(start: Date | null, end: Date | null) {
    timelineState.update(state => ({
      ...state,
      activeFilters: {
        ...state.activeFilters,
        dateRange: { start, end },
      },
    }));
  },

  clearFilters() {
    timelineState.update(state => ({
      ...state,
      activeFilters: {
        projects: [],
        eventTypes: [],
        tags: [],
        dateRange: { start: null, end: null },
        searchQuery: '',
      },
    }));
  },

  // Performance and rendering
  updateRenderContext(context: Partial<RenderContext>) {
    timelineState.update(state => ({
      ...state,
      renderContext: { ...state.renderContext, ...context },
    }));
  },

  // Utility actions
  jumpToEvent(eventId: string) {
    const state = get(timelineState);
    if (!state.journeyData) return;

    const event = state.journeyData.events.find(e => e.id === eventId);
    if (!event) return;

    const eventTime = new Date(event.timestamp);
    const journeyStart = new Date(state.journeyData.timeline.startTime);
    const journeyEnd = new Date(state.journeyData.timeline.endTime);
    const fullDuration = journeyEnd.getTime() - journeyStart.getTime();

    // Calculate position along timeline
    const eventOffset = eventTime.getTime() - journeyStart.getTime();
    let targetPosition = eventOffset / fullDuration;

    // Adjust for current viewport scale
    if (state.viewport.scale !== 'full') {
      const timeScale = TIME_SCALES[state.viewport.scale];
      const viewportDuration = timeScale.duration!;
      const maxPosition = fullDuration - viewportDuration;

      // Center the event in the viewport
      const halfViewport = viewportDuration / 2;
      const centeredOffset = eventOffset - halfViewport;
      targetPosition = Math.max(0, Math.min(1, centeredOffset / maxPosition));
    }

    timelineActions.setViewportPosition(targetPosition);
    timelineActions.selectEvent(eventId);
  },

  // Keyboard shortcuts
  handleKeyboardShortcut(key: string, ctrlKey: boolean, shiftKey: boolean) {
    switch (key) {
      case 'Space':
        const state = get(timelineState);
        if (state.controls.isPlaying) {
          timelineActions.pause();
        } else {
          timelineActions.play();
        }
        break;
      case 'ArrowLeft':
        timelineActions.panViewport(shiftKey ? -0.1 : -0.02);
        break;
      case 'ArrowRight':
        timelineActions.panViewport(shiftKey ? 0.1 : 0.02);
        break;
      case 'Home':
        timelineActions.setViewportPosition(0);
        break;
      case 'End':
        timelineActions.setViewportPosition(1);
        break;
      case 'Escape':
        timelineActions.selectEvent(null);
        break;
      case '+':
      case '=':
        if (ctrlKey) {
          // Zoom in by switching to smaller time scale
          const state = get(timelineState);
          const scaleOrder: TimeScaleName[] = ['7d', '24h', '6h', '1h', '15m'];
          const currentIndex = scaleOrder.indexOf(state.viewport.scale);
          if (currentIndex < scaleOrder.length - 1) {
            timelineActions.setViewportScale(scaleOrder[currentIndex + 1]);
          }
        }
        break;
      case '-':
        if (ctrlKey) {
          // Zoom out by switching to larger time scale
          const state = get(timelineState);
          const scaleOrder: TimeScaleName[] = ['15m', '1h', '6h', '24h', '7d', 'full'];
          const currentIndex = scaleOrder.indexOf(state.viewport.scale);
          if (currentIndex < scaleOrder.length - 1) {
            timelineActions.setViewportScale(scaleOrder[currentIndex + 1]);
          }
        }
        break;
    }
  },
};

// Performance monitoring
export const performanceStore = readable({ fps: 60, renderTime: 0, eventCount: 0 }, set => {
  let frameCount = 0;
  let lastTime = performance.now();
  let renderTimes: number[] = [];

  const updatePerformance = () => {
    const currentTime = performance.now();
    const deltaTime = currentTime - lastTime;
    frameCount++;

    if (frameCount >= 60) {
      const fps = Math.round(1000 / (deltaTime / frameCount));
      const avgRenderTime = renderTimes.reduce((a, b) => a + b, 0) / renderTimes.length;
      const eventCount = get(eventsInCurrentViewport).length;

      set({ fps, renderTime: avgRenderTime, eventCount });

      frameCount = 0;
      renderTimes = [];
    }

    lastTime = currentTime;
    requestAnimationFrame(updatePerformance);
  };

  const unsubscribe = requestAnimationFrame(updatePerformance);
  return () => cancelAnimationFrame(unsubscribe);
});

// Auto-save viewport position
export const autoSaveViewport = () => {
  const unsubscribe = viewport.subscribe($viewport => {
    if (typeof window !== 'undefined') {
      localStorage.setItem(
        'uroboro-timeline-viewport',
        JSON.stringify({
          scale: $viewport.scale,
          position: $viewport.position,
        })
      );
    }
  });

  return unsubscribe;
};

// Load saved viewport position
export const loadSavedViewport = () => {
  if (typeof window !== 'undefined') {
    const saved = localStorage.getItem('uroboro-timeline-viewport');
    if (saved) {
      try {
        const { scale, position } = JSON.parse(saved);
        timelineActions.setViewportScale(scale);
        timelineActions.setViewportPosition(position);
      } catch (e) {
        console.warn('Failed to load saved viewport position:', e);
      }
    }
  }
};
