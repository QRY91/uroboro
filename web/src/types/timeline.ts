// Core Timeline Types
export interface TimelineViewport {
  startTime: Date | null;
  endTime: Date | null;
  scale: TimeScaleName;
  position: number; // Position along full timeline (0-1)
  eventsInView: JourneyEvent[];
}

export interface TimeScale {
  name: TimeScaleName;
  duration: number | null; // Duration in milliseconds, null for 'full'
  tickInterval: number | null; // Major tick interval in milliseconds
  label: string;
}

export type TimeScaleName = '15m' | '1h' | '6h' | '24h' | '7d' | 'full';

// Journey Event Types
export interface JourneyEvent {
  id: string;
  timestamp: string; // ISO date string
  type: EventType;
  project: string;
  content: string;
  description?: string;
  tags: string[];
  metadata?: EventMetadata;
  position?: TimelinePosition; // Calculated position for rendering
}

export type EventType =
  | 'milestone'
  | 'learning'
  | 'decision'
  | 'commit'
  | 'capture'
  | 'error'
  | 'success'
  | 'context_switch';

export interface EventMetadata {
  duration?: number; // For events with duration
  impact?: 'low' | 'medium' | 'high';
  confidence?: number; // 0-1 for AI-generated content
  connections?: string[]; // IDs of related events
  files?: string[]; // Associated file paths
  gitHash?: string; // For commit events
  branchName?: string; // For git-related events
  [key: string]: unknown; // Allow additional metadata
}

export interface TimelinePosition {
  x: number; // Horizontal position in viewport
  y: number; // Vertical position (for clustering)
  visible: boolean; // Whether event is in current viewport
  scale: number; // Size multiplier based on current scale
}

// Journey Data Structure
export interface JourneyData {
  events: JourneyEvent[];
  projects: ProjectInfo[];
  timeline: {
    startTime: string;
    endTime: string;
    totalDuration: number;
  };
  stats: JourneyStats;
}

export interface ProjectInfo {
  name: string;
  color: string;
  eventCount: number;
  lastActivity: string;
  tags: string[];
}

export interface JourneyStats {
  totalEvents: number;
  projectCount: number;
  milestoneCount: number;
  learningMoments: number;
  avgEventsPerDay: number;
  mostActiveProject: string;
  longestSession: number; // in minutes
}

// Timeline UI Component Types
export interface TimelineControls {
  isPlaying: boolean;
  playSpeed: number; // 0.5 - 3.0
  currentScale: TimeScaleName;
  viewportPosition: number; // 0-1
  zoomLevel: number; // 50-300 (percentage)
}

export interface ViewportSliderState {
  isDragging: boolean;
  startPosition: number;
  currentPosition: number;
  dragOffset: number;
}

export interface TimelineRuler {
  majorTicks: TickMark[];
  minorTicks: TickMark[];
  labels: TimeLabel[];
}

export interface TickMark {
  position: number; // 0-1 relative position
  timestamp: Date;
  type: 'major' | 'minor';
}

export interface TimeLabel {
  position: number; // 0-1 relative position
  text: string;
  timestamp: Date;
}

// Interaction Types
export interface TimelineInteraction {
  type: InteractionType;
  startTime: Date;
  endTime?: Date;
  position: { x: number; y: number };
  target?: string; // Event ID or UI element
}

export type InteractionType = 'pan' | 'zoom' | 'scrub' | 'click' | 'hover' | 'keyboard';

export interface TouchGesture {
  type: 'pan' | 'zoom' | 'tap';
  startPosition: { x: number; y: number };
  currentPosition: { x: number; y: number };
  scale?: number; // For zoom gestures
  velocity?: { x: number; y: number };
}

// Animation Types
export interface TimelineAnimation {
  id: string;
  type: AnimationType;
  duration: number; // milliseconds
  easing: string; // anime.js easing function
  target: string | HTMLElement;
  properties: Record<string, unknown>;
  onComplete?: () => void;
}

export type AnimationType =
  | 'viewport_change'
  | 'scale_transition'
  | 'event_highlight'
  | 'connection_draw'
  | 'scrub_animation';

// Rendering and Performance Types
export interface ViewportBounds {
  left: number;
  right: number;
  top: number;
  bottom: number;
  width: number;
  height: number;
}

export interface RenderContext {
  viewport: TimelineViewport;
  bounds: ViewportBounds;
  devicePixelRatio: number;
  isLowSpec: boolean; // Detected low-spec device
  performanceLevel: 'high' | 'medium' | 'low';
}

export interface EventCluster {
  id: string;
  events: JourneyEvent[];
  position: TimelinePosition;
  density: number; // Events per unit time
  representative: JourneyEvent; // Main event to show
}

// Search and Filter Types
export interface TimelineFilter {
  projects: string[];
  eventTypes: EventType[];
  tags: string[];
  dateRange: {
    start: Date | null;
    end: Date | null;
  };
  searchQuery: string;
}

export interface SearchResult {
  event: JourneyEvent;
  relevance: number; // 0-1 score
  matchedFields: string[]; // Which fields matched the query
  highlightRanges: TextRange[]; // For highlighting matches
}

export interface TextRange {
  start: number;
  end: number;
  field: string;
}

// State Management Types
export interface TimelineState {
  // Data
  journeyData: JourneyData | null;
  loading: boolean;
  error: string | null;

  // Viewport
  viewport: TimelineViewport;
  controls: TimelineControls;

  // UI State
  selectedEvent: string | null;
  hoveredEvent: string | null;
  searchResults: SearchResult[];
  activeFilters: TimelineFilter;

  // Performance
  renderContext: RenderContext;
  eventClusters: EventCluster[];

  // Interaction
  currentInteraction: TimelineInteraction | null;
  lastUpdate: number; // Timestamp for throttling
}

// Configuration Types
export interface TimelineConfig {
  // Performance settings
  maxEventsInView: number;
  clusterThreshold: number; // Events per pixel before clustering
  animationDuration: number;
  throttleMs: number;

  // Visual settings
  eventSizes: Record<TimeScaleName, number>;
  colorScheme: ColorScheme;
  fontSizes: Record<TimeScaleName, string>;

  // Interaction settings
  panSensitivity: number;
  zoomSensitivity: number;
  doubleClickThreshold: number; // milliseconds

  // Mobile optimizations
  touchThresholds: {
    tapDistance: number; // pixels
    panDistance: number; // pixels
    zoomDistance: number; // pixels
  };
}

export interface ColorScheme {
  background: string;
  primary: string;
  secondary: string;
  accent: string;
  text: string;
  textSecondary: string;
  eventColors: Record<EventType, string>;
  projectColors: string[]; // Cycling colors for projects
}

// API Types
export interface JourneyAPIResponse {
  success: boolean;
  data?: JourneyData;
  error?: string;
  metadata?: {
    queryTime: number;
    cacheHit: boolean;
    version: string;
  };
}

export interface JourneyAPIRequest {
  days?: number;
  projects?: string[];
  eventTypes?: EventType[];
  startTime?: string;
  endTime?: string;
  limit?: number;
  offset?: number;
}

// Utility Types
export type DeepPartial<T> = {
  [P in keyof T]?: T[P] extends object ? DeepPartial<T[P]> : T[P];
};

export type TimelineEventHandler<T = Event> = (event: T) => void;

export type Unsubscribe = () => void;

// Constants and Defaults
export const TIME_SCALES: Record<TimeScaleName, TimeScale> = {
  '15m': {
    name: '15m',
    duration: 15 * 60 * 1000,
    tickInterval: 5 * 60 * 1000,
    label: '15 Minutes',
  },
  '1h': {
    name: '1h',
    duration: 60 * 60 * 1000,
    tickInterval: 10 * 60 * 1000,
    label: '1 Hour',
  },
  '6h': {
    name: '6h',
    duration: 6 * 60 * 60 * 1000,
    tickInterval: 60 * 60 * 1000,
    label: '6 Hours',
  },
  '24h': {
    name: '24h',
    duration: 24 * 60 * 60 * 1000,
    tickInterval: 4 * 60 * 60 * 1000,
    label: '24 Hours',
  },
  '7d': {
    name: '7d',
    duration: 7 * 24 * 60 * 60 * 1000,
    tickInterval: 24 * 60 * 60 * 1000,
    label: '7 Days',
  },
  full: {
    name: 'full',
    duration: null,
    tickInterval: null,
    label: 'Full Journey',
  },
};

export const DEFAULT_TIMELINE_CONFIG: TimelineConfig = {
  maxEventsInView: 1000,
  clusterThreshold: 5,
  animationDuration: 300,
  throttleMs: 16, // 60fps

  eventSizes: {
    '15m': 24,
    '1h': 20,
    '6h': 16,
    '24h': 12,
    '7d': 8,
    full: 6,
  },

  colorScheme: {
    background: '#1a1a1a',
    primary: '#00ffff',
    secondary: '#00e6e6',
    accent: '#ffff00',
    text: '#ffffff',
    textSecondary: '#dddddd',
    eventColors: {
      milestone: '#ff0080',
      learning: '#00ff80',
      decision: '#ffff00',
      commit: '#0080ff',
      capture: '#ff8000',
      error: '#ff0000',
      success: '#00ff00',
      context_switch: '#8000ff',
    },
    projectColors: [
      '#ff0080',
      '#00ff80',
      '#0080ff',
      '#ffff00',
      '#ff8000',
      '#8000ff',
      '#ff4080',
      '#80ff40',
      '#4080ff',
      '#ff8040',
    ],
  },

  fontSizes: {
    '15m': '14px',
    '1h': '12px',
    '6h': '11px',
    '24h': '10px',
    '7d': '9px',
    full: '8px',
  },

  panSensitivity: 1.0,
  zoomSensitivity: 0.1,
  doubleClickThreshold: 300,

  touchThresholds: {
    tapDistance: 10,
    panDistance: 5,
    zoomDistance: 20,
  },
};
