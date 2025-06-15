// Timeline Utility Functions
// Comprehensive utilities for timeline operations, calculations, and optimizations

import type {
  JourneyEvent,
  TimelineViewport,
  TimeScaleName,
  EventCluster,
  TouchGesture,
  TimelinePosition
} from '@types/timeline';
import { TIME_SCALES } from '@types/timeline';

// ================================================
// Time Calculations and Formatting
// ================================================

/**
 * Format a duration in milliseconds to human-readable string
 */
export function formatDuration(ms: number): string {
  const seconds = Math.floor(ms / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);
  const days = Math.floor(hours / 24);

  if (days > 0) {
    return `${days}d ${hours % 24}h`;
  } else if (hours > 0) {
    return `${hours}h ${minutes % 60}m`;
  } else if (minutes > 0) {
    return `${minutes}m ${seconds % 60}s`;
  } else {
    return `${seconds}s`;
  }
}

/**
 * Format a timestamp relative to now (e.g., "2 hours ago")
 */
export function formatRelativeTime(timestamp: string | Date): string {
  const date = typeof timestamp === 'string' ? new Date(timestamp) : timestamp;
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();

  const diffSeconds = Math.floor(diffMs / 1000);
  const diffMinutes = Math.floor(diffSeconds / 60);
  const diffHours = Math.floor(diffMinutes / 60);
  const diffDays = Math.floor(diffHours / 24);
  const diffWeeks = Math.floor(diffDays / 7);
  const diffMonths = Math.floor(diffDays / 30);
  const diffYears = Math.floor(diffDays / 365);

  if (diffYears > 0) {
    return `${diffYears} year${diffYears === 1 ? '' : 's'} ago`;
  } else if (diffMonths > 0) {
    return `${diffMonths} month${diffMonths === 1 ? '' : 's'} ago`;
  } else if (diffWeeks > 0) {
    return `${diffWeeks} week${diffWeeks === 1 ? '' : 's'} ago`;
  } else if (diffDays > 0) {
    return `${diffDays} day${diffDays === 1 ? '' : 's'} ago`;
  } else if (diffHours > 0) {
    return `${diffHours} hour${diffHours === 1 ? '' : 's'} ago`;
  } else if (diffMinutes > 0) {
    return `${diffMinutes} minute${diffMinutes === 1 ? '' : 's'} ago`;
  } else if (diffSeconds > 10) {
    return `${diffSeconds} seconds ago`;
  } else {
    return 'Just now';
  }
}

/**
 * Format time for timeline scale labels
 */
export function formatTimeForScale(date: Date, scale: TimeScaleName): string {
  const now = new Date();
  const isToday = date.toDateString() === now.toDateString();
  const isThisYear = date.getFullYear() === now.getFullYear();

  switch (scale) {
    case '15m':
    case '1h':
      return date.toLocaleTimeString([], {
        hour: '2-digit',
        minute: '2-digit',
        hour12: false
      });

    case '6h':
      return date.toLocaleTimeString([], {
        hour: '2-digit',
        minute: '2-digit',
        hour12: false
      });

    case '24h':
      if (isToday) {
        return date.toLocaleTimeString([], {
          hour: '2-digit',
          minute: '2-digit',
          hour12: false
        });
      } else {
        return date.toLocaleDateString([], {
          month: 'short',
          day: 'numeric'
        });
      }

    case '7d':
      return date.toLocaleDateString([], {
        month: 'short',
        day: 'numeric',
        weekday: 'short'
      });

    case 'full':
      if (isThisYear) {
        return date.toLocaleDateString([], {
          month: 'short',
          day: 'numeric'
        });
      } else {
        return date.toLocaleDateString([], {
          year: '2-digit',
          month: 'short'
        });
      }

    default:
      return date.toLocaleString();
  }
}

/**
 * Calculate the optimal time scale for a given duration
 */
export function getOptimalTimeScale(durationMs: number): TimeScaleName {
  const hours = durationMs / (1000 * 60 * 60);

  if (hours <= 0.5) return '15m';
  if (hours <= 2) return '1h';
  if (hours <= 12) return '6h';
  if (hours <= 48) return '24h';
  if (hours <= 168) return '7d'; // 1 week
  return 'full';
}

// ================================================
// Event Positioning and Clustering
// ================================================

/**
 * Calculate position of an event within a viewport
 */
export function calculateEventPosition(
  event: JourneyEvent,
  viewport: TimelineViewport,
  containerWidth: number,
  containerHeight: number
): TimelinePosition | null {
  if (!viewport.startTime || !viewport.endTime) return null;

  const eventTime = new Date(event.timestamp).getTime();
  const viewportStart = viewport.startTime.getTime();
  const viewportEnd = viewport.endTime.getTime();
  const viewportDuration = viewportEnd - viewportStart;

  // Check if event is within viewport bounds (with small buffer)
  const buffer = viewportDuration * 0.05; // 5% buffer
  const isVisible = eventTime >= (viewportStart - buffer) &&
                   eventTime <= (viewportEnd + buffer);

  if (!isVisible) {
    return { x: 0, y: 0, visible: false, scale: 1 };
  }

  // Calculate horizontal position (0-1 range)
  const relativePosition = (eventTime - viewportStart) / viewportDuration;
  const x = relativePosition * containerWidth;

  // Calculate vertical position with simple clustering by project
  const projectHash = hashString(event.project);
  const laneCount = Math.min(8, Math.max(3, Math.floor(containerHeight / 80))); // Adaptive lanes
  const y = 60 + (projectHash % laneCount) * (containerHeight - 120) / laneCount;

  return {
    x,
    y,
    visible: true,
    scale: 1 // Can be adjusted based on importance or scale
  };
}

/**
 * Cluster events that are close together to prevent overlap
 */
export function clusterEvents(
  events: JourneyEvent[],
  viewport: TimelineViewport,
  containerWidth: number,
  minDistance: number = 20
): EventCluster[] {
  if (!viewport.startTime || !viewport.endTime) return [];

  const sortedEvents = [...events].sort((a, b) =>
    new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime()
  );

  const clusters: EventCluster[] = [];
  let currentCluster: JourneyEvent[] = [];

  for (let i = 0; i < sortedEvents.length; i++) {
    const event = sortedEvents[i];
    const position = calculateEventPosition(event, viewport, containerWidth, 400);

    if (!position?.visible) continue;

    if (currentCluster.length === 0) {
      currentCluster = [event];
    } else {
      const lastEvent = currentCluster[currentCluster.length - 1];
      const lastPosition = calculateEventPosition(lastEvent, viewport, containerWidth, 400);

      if (lastPosition && Math.abs(position.x - lastPosition.x) < minDistance) {
        // Events are close together, add to current cluster
        currentCluster.push(event);
      } else {
        // Events are far apart, finalize current cluster and start new one
        if (currentCluster.length > 0) {
          clusters.push(createCluster(currentCluster, viewport, containerWidth));
        }
        currentCluster = [event];
      }
    }
  }

  // Don't forget the last cluster
  if (currentCluster.length > 0) {
    clusters.push(createCluster(currentCluster, viewport, containerWidth));
  }

  return clusters;
}

/**
 * Create a cluster from a group of events
 */
function createCluster(
  events: JourneyEvent[],
  viewport: TimelineViewport,
  containerWidth: number
): EventCluster {
  // Use the most recent event as representative
  const representative = events[events.length - 1];
  const position = calculateEventPosition(representative, viewport, containerWidth, 400);

  // Calculate density (events per minute)
  const timeSpan = new Date(events[events.length - 1].timestamp).getTime() -
                  new Date(events[0].timestamp).getTime();
  const density = events.length / Math.max(1, timeSpan / (1000 * 60));

  return {
    id: `cluster-${events.map(e => e.id).join('-')}`,
    events,
    position: position || { x: 0, y: 0, visible: false, scale: 1 },
    density,
    representative
  };
}

// ================================================
// Performance and Device Detection
// ================================================

/**
 * Detect if device is low-spec based on various factors
 */
export function detectLowSpecDevice(): boolean {
  // Check for mobile device
  const isMobile = /Android|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent);

  // Check memory (if available)
  const memory = (navigator as any).deviceMemory;
  const isLowMemory = memory && memory <= 4; // 4GB or less

  // Check CPU cores
  const cores = navigator.hardwareConcurrency || 1;
  const isLowCores = cores <= 2;

  // Check connection speed
  const connection = (navigator as any).connection;
  const isSlowConnection = connection &&
    (connection.effectiveType === 'slow-2g' || connection.effectiveType === '2g');

  return isMobile || isLowMemory || isLowCores || isSlowConnection;
}

/**
 * Get performance level recommendation
 */
export function getPerformanceLevel(): 'high' | 'medium' | 'low' {
  if (detectLowSpecDevice()) return 'low';

  const memory = (navigator as any).deviceMemory;
  const cores = navigator.hardwareConcurrency || 1;

  if (memory >= 8 && cores >= 4) return 'high';
  return 'medium';
}

/**
 * Throttle function calls for performance
 */
export function throttle<T extends (...args: any[]) => any>(
  func: T,
  delay: number
): (...args: Parameters<T>) => void {
  let timeoutId: number | null = null;
  let lastExecTime = 0;

  return (...args: Parameters<T>) => {
    const currentTime = Date.now();

    if (currentTime - lastExecTime > delay) {
      func(...args);
      lastExecTime = currentTime;
    } else {
      if (timeoutId) clearTimeout(timeoutId);
      timeoutId = window.setTimeout(() => {
        func(...args);
        lastExecTime = Date.now();
      }, delay - (currentTime - lastExecTime));
    }
  };
}

/**
 * Debounce function calls
 */
export function debounce<T extends (...args: any[]) => any>(
  func: T,
  delay: number
): (...args: Parameters<T>) => void {
  let timeoutId: number | null = null;

  return (...args: Parameters<T>) => {
    if (timeoutId) clearTimeout(timeoutId);
    timeoutId = window.setTimeout(() => func(...args), delay);
  };
}

// ================================================
// Touch and Gesture Utilities
// ================================================

/**
 * Calculate distance between two touch points
 */
export function getTouchDistance(touches: TouchList): number {
  if (touches.length < 2) return 0;

  const dx = touches[0].clientX - touches[1].clientX;
  const dy = touches[0].clientY - touches[1].clientY;
  return Math.sqrt(dx * dx + dy * dy);
}

/**
 * Detect gesture type from touch events
 */
export function detectGesture(
  startTouches: TouchList,
  currentTouches: TouchList,
  startTime: number,
  currentTime: number
): TouchGesture | null {
  if (startTouches.length !== currentTouches.length) return null;

  const timeDiff = currentTime - startTime;
  const minGestureTime = 50; // Minimum time for gesture recognition

  if (timeDiff < minGestureTime) return null;

  if (startTouches.length === 1) {
    // Single finger gesture
    const startX = startTouches[0].clientX;
    const startY = startTouches[0].clientY;
    const currentX = currentTouches[0].clientX;
    const currentY = currentTouches[0].clientY;

    const deltaX = currentX - startX;
    const deltaY = currentY - startY;
    const distance = Math.sqrt(deltaX * deltaX + deltaY * deltaY);

    if (distance < 10) {
      return {
        type: 'tap',
        startPosition: { x: startX, y: startY },
        currentPosition: { x: currentX, y: currentY }
      };
    } else {
      const velocity = {
        x: deltaX / timeDiff,
        y: deltaY / timeDiff
      };

      return {
        type: 'pan',
        startPosition: { x: startX, y: startY },
        currentPosition: { x: currentX, y: currentY },
        velocity
      };
    }
  } else if (startTouches.length === 2) {
    // Two finger gesture (zoom)
    const startDistance = getTouchDistance(startTouches);
    const currentDistance = getTouchDistance(currentTouches);
    const scale = currentDistance / startDistance;

    return {
      type: 'zoom',
      startPosition: {
        x: (startTouches[0].clientX + startTouches[1].clientX) / 2,
        y: (startTouches[0].clientY + startTouches[1].clientY) / 2
      },
      currentPosition: {
        x: (currentTouches[0].clientX + currentTouches[1].clientX) / 2,
        y: (currentTouches[0].clientY + currentTouches[1].clientY) / 2
      },
      scale
    };
  }

  return null;
}

// ================================================
// Data Processing Utilities
// ================================================

/**
 * Simple hash function for consistent project colors
 */
export function hashString(str: string): number {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    const char = str.charCodeAt(i);
    hash = ((hash << 5) - hash) + char;
    hash = hash & hash; // Convert to 32-bit integer
  }
  return Math.abs(hash);
}

/**
 * Filter events by search query
 */
export function filterEventsByQuery(events: JourneyEvent[], query: string): JourneyEvent[] {
  if (!query.trim()) return events;

  const searchTerms = query.toLowerCase().split(/\s+/);

  return events.filter(event => {
    const searchableText = [
      event.content,
      event.description || '',
      event.project,
      event.type,
      ...event.tags
    ].join(' ').toLowerCase();

    return searchTerms.every(term => searchableText.includes(term));
  });
}

/**
 * Calculate activity density over time
 */
export function calculateActivityDensity(
  events: JourneyEvent[],
  segments: number = 100
): number[] {
  if (events.length === 0) return new Array(segments).fill(0);

  const sortedEvents = [...events].sort((a, b) =>
    new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime()
  );

  const startTime = new Date(sortedEvents[0].timestamp).getTime();
  const endTime = new Date(sortedEvents[sortedEvents.length - 1].timestamp).getTime();
  const duration = endTime - startTime;
  const segmentDuration = duration / segments;

  const density = new Array(segments).fill(0);

  sortedEvents.forEach(event => {
    const eventTime = new Date(event.timestamp).getTime();
    const segmentIndex = Math.floor((eventTime - startTime) / segmentDuration);
    if (segmentIndex >= 0 && segmentIndex < segments) {
      density[segmentIndex]++;
    }
  });

  // Normalize to 0-1 range
  const maxDensity = Math.max(...density);
  return maxDensity > 0 ? density.map(d => d / maxDensity) : density;
}

/**
 * Group events by project
 */
export function groupEventsByProject(events: JourneyEvent[]): Map<string, JourneyEvent[]> {
  const groups = new Map<string, JourneyEvent[]>();

  events.forEach(event => {
    const existing = groups.get(event.project) || [];
    existing.push(event);
    groups.set(event.project, existing);
  });

  return groups;
}

/**
 * Calculate project statistics
 */
export function calculateProjectStats(events: JourneyEvent[]) {
  const projectGroups = groupEventsByProject(events);
  const stats = new Map();

  projectGroups.forEach((projectEvents, projectName) => {
    const sortedEvents = [...projectEvents].sort((a, b) =>
      new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime()
    );

    const firstEvent = sortedEvents[0];
    const lastEvent = sortedEvents[sortedEvents.length - 1];
    const duration = new Date(lastEvent.timestamp).getTime() -
                    new Date(firstEvent.timestamp).getTime();

    const eventTypes = new Map<string, number>();
    projectEvents.forEach(event => {
      eventTypes.set(event.type, (eventTypes.get(event.type) || 0) + 1);
    });

    stats.set(projectName, {
      totalEvents: projectEvents.length,
      duration,
      firstActivity: firstEvent.timestamp,
      lastActivity: lastEvent.timestamp,
      eventTypes: Object.fromEntries(eventTypes),
      averageEventsPerDay: projectEvents.length / Math.max(1, duration / (1000 * 60 * 60 * 24))
    });
  });

  return stats;
}

// ================================================
// Animation and Easing Utilities
// ================================================

/**
 * Custom easing functions for smooth animations
 */
export const easingFunctions = {
  easeInOut: (t: number): number => t < 0.5 ? 2 * t * t : -1 + (4 - 2 * t) * t,
  easeOut: (t: number): number => 1 - Math.pow(1 - t, 3),
  easeIn: (t: number): number => t * t * t,
  easeOutBack: (t: number): number => {
    const c1 = 1.70158;
    const c3 = c1 + 1;
    return 1 + c3 * Math.pow(t - 1, 3) + c1 * Math.pow(t - 1, 2);
  }
};

/**
 * Animate a value over time
 */
export function animateValue(
  from: number,
  to: number,
  duration: number,
  callback: (value: number) => void,
  easing: keyof typeof easingFunctions = 'easeInOut'
): () => void {
  const startTime = performance.now();
  const easingFn = easingFunctions[easing];
  let rafId: number;

  const animate = (currentTime: number) => {
    const elapsed = currentTime - startTime;
    const progress = Math.min(elapsed / duration, 1);

    const easedProgress = easingFn(progress);
    const currentValue = from + (to - from) * easedProgress;

    callback(currentValue);

    if (progress < 1) {
      rafId = requestAnimationFrame(animate);
    }
  };

  rafId = requestAnimationFrame(animate);

  // Return cancel function
  return () => cancelAnimationFrame(rafId);
}

// ================================================
// Viewport and Scale Utilities
// ================================================

/**
 * Calculate viewport bounds for a given scale and position
 */
export function calculateViewportBounds(
  scale: TimeScaleName,
  position: number,
  journeyStart: Date,
  journeyEnd: Date
): { startTime: Date; endTime: Date } | null {
  const timeScale = TIME_SCALES[scale];
  const fullDuration = journeyEnd.getTime() - journeyStart.getTime();

  if (scale === 'full') {
    return { startTime: journeyStart, endTime: journeyEnd };
  }

  const viewportDuration = timeScale.duration!;
  const maxPosition = fullDuration - viewportDuration;
  const clampedPosition = Math.max(0, Math.min(1, position));
  const currentOffset = maxPosition * clampedPosition;

  const startTime = new Date(journeyStart.getTime() + currentOffset);
  let endTime = new Date(startTime.getTime() + viewportDuration);

  // Ensure we don't go beyond journey bounds
  if (endTime > journeyEnd) {
    endTime = journeyEnd;
    const adjustedStartTime = new Date(endTime.getTime() - viewportDuration);
    if (adjustedStartTime >= journeyStart) {
      return { startTime: adjustedStartTime, endTime };
    }
  }

  return { startTime, endTime };
}

/**
 * Get next/previous time scale for zooming
 */
export function getAdjacentTimeScale(
  currentScale: TimeScaleName,
  direction: 'in' | 'out'
): TimeScaleName | null {
  const scaleOrder: TimeScaleName[] = ['15m', '1h', '6h', '24h', '7d', 'full'];
  const currentIndex = scaleOrder.indexOf(currentScale);

  if (direction === 'in' && currentIndex > 0) {
    return scaleOrder[currentIndex - 1];
  } else if (direction === 'out' && currentIndex < scaleOrder.length - 1) {
    return scaleOrder[currentIndex + 1];
  }

  return null;
}

// ================================================
// Export all utilities
// ================================================

export default {
  // Time utilities
  formatDuration,
  formatRelativeTime,
  formatTimeForScale,
  getOptimalTimeScale,

  // Event positioning
  calculateEventPosition,
  clusterEvents,

  // Performance
  detectLowSpecDevice,
  getPerformanceLevel,
  throttle,
  debounce,

  // Touch/gesture
  getTouchDistance,
  detectGesture,

  // Data processing
  hashString,
  filterEventsByQuery,
  calculateActivityDensity,
  groupEventsByProject,
  calculateProjectStats,

  // Animation
  easingFunctions,
  animateValue,

  // Viewport
  calculateViewportBounds,
  getAdjacentTimeScale
};
