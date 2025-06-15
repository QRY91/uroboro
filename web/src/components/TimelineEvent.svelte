<script lang="ts">
  import { createEventDispatcher, onDestroy, onMount } from 'svelte';
  import type { JourneyEvent, TimelineViewport, TimeScaleName } from '@types/timeline';
  import { DEFAULT_TIMELINE_CONFIG } from '@types/timeline';
  import anime from 'animejs';

  const dispatch = createEventDispatcher();

  // Props
  export let event: JourneyEvent;
  export let index: number;
  export let totalEvents: number;
  export let projectLanes: any[];
  export let isContextSwitch: boolean = false;
  export let scale: TimeScaleName;
  export let viewport: TimelineViewport;
  export let mode: 'playback' | 'explorer' = 'playback';
  export let selected = false;
  export let hovered = false;

  // Component state
  let eventElement: HTMLDivElement;
  let eventPosition = { x: 0, y: 0 };
  let eventSize = 16;
  let projectLane = 0;
  let isHovering = false;
  let isDestroyed = false;
  let viewportStartTime = viewport?.startTime?.getTime();

  // Reactive calculations - simplified to trust parent filtering
  $: if (viewport?.startTime && viewport?.endTime && scale && event) {
    updateEventPosition();
    updateEventSize();
  }

  function updateEventPosition() {
    const eventTime = new Date(event.timestamp).getTime();
    const viewportStart = viewport.startTime!.getTime();
    const viewportEnd = viewport.endTime!.getTime();
    const viewportDuration = viewportEnd - viewportStart;

    // Calculate horizontal position with better spacing
    const relativePosition = (eventTime - viewportStart) / viewportDuration;
    let baseX = relativePosition * 100; // Percentage

    // Add jitter based on event content hash to prevent exact overlaps
    const eventHash = hashString(event.id + event.content);
    const jitterAmount = 0.5; // Max 0.5% jitter
    const jitter = ((eventHash % 100) - 50) * jitterAmount / 50; // -0.5% to +0.5%

    // Apply jitter and ensure we stay within bounds
    const newX = Math.max(0.5, Math.min(99.5, baseX + jitter));

    // Calculate vertical position with project clustering
    calculateProjectLane();

    // Use more vertical space - spread across available height with dynamic lanes
    const baseY = 60; // Start higher to use more space
    const laneHeight = Math.min(80, Math.max(40, 400 / Math.max(projectLanes.length, 4))); // Adaptive lane height based on dynamic lanes

    // Add small vertical jitter within lane to prevent exact overlaps
    const verticalJitter = ((eventHash % 10) - 5) * 2; // -10px to +10px
    const newY = baseY + (projectLane * laneHeight) + verticalJitter;

    // Update position and trigger reactivity
    eventPosition = { x: newX, y: newY };
  }

  // Simple viewport change detection
  $: if (viewport?.startTime && viewportStartTime && viewport.startTime.getTime() !== viewportStartTime) {
    // Clear hover state on viewport changes to prevent stuck tooltips
    isHovering = false;
    viewportStartTime = viewport.startTime.getTime();
  }

  function calculateProjectLane() {
    // Use dynamic lane system - find this event's project in the lanes
    const lane = projectLanes.find(l => l.project === event.project);
    if (lane) {
      projectLane = lane.laneIndex;
    } else {
      // Fallback if project not found (shouldn't happen)
      projectLane = 0;
    }
  }

  function updateEventSize() {
    // Make event size responsive to timescale for better visibility
    const scaleSizes = {
      '5m': 24,    // Largest for most detailed view
      '15m': 20,
      '30m': 18,
      '1h': 16,    // Default size
      '2h': 14,
      '6h': 12,
      '12h': 10,
      '24h': 8,
      '7d': 6,     // Smallest for overview
      'full': 6
    };
    eventSize = scaleSizes[scale] || 16;
  }



  function hashString(str: string): number {
    let hash = 0;
    for (let i = 0; i < str.length; i++) {
      const char = str.charCodeAt(i);
      hash = (hash << 5) - hash + char;
      hash = hash & hash; // Convert to 32-bit integer
    }
    return Math.abs(hash);
  }

  function getEventIcon(eventType: string): string {
    const icons = {
      milestone: '🎯',
      learning: '💡',
      decision: '🎲',
      commit: '📝',
      capture: '📸',
      error: '⚠️',
      success: '✅',
      context_switch: '🔄',
    };
    return icons[eventType as keyof typeof icons] || '📌';
  }

  function getEventColor(eventType: string): string {
    const colors = DEFAULT_TIMELINE_CONFIG.colorScheme.eventColors;
    return colors[eventType as keyof typeof colors] || '#00ffff';
  }

  function getProjectColor(projectName: string): string {
    const colors = DEFAULT_TIMELINE_CONFIG.colorScheme.projectColors;
    const index = hashString(projectName) % colors.length;
    return colors[index];
  }

  function getProjectLaneColor(projectName: string): string {
    // Softer background color for the project lane
    const colors = DEFAULT_TIMELINE_CONFIG.colorScheme.projectColors;
    const index = hashString(projectName) % colors.length;
    const baseColor = colors[index];
    // Convert to rgba with low opacity for lane background
    return baseColor.replace('rgb', 'rgba').replace(')', ', 0.15)');
  }

  function handleClick() {
    if (isDestroyed) return;

    dispatch('click', { event, element: eventElement });

    // Simple click feedback without tracking animations
    if (eventElement && !isDestroyed) {
      anime({
        targets: eventElement,
        scale: [1, 1.2, 1],
        duration: 200,
        easing: 'easeOutBack',
      });
    }
  }

  function handleMouseEnter() {
    if (isDestroyed) return;

    isHovering = true;
    dispatch('hover', { event, hovered: true });

    // Mode-aware hover effect
    if (eventElement && !isDestroyed) {
      const duration = mode === 'playback' ? 200 : 100;
      const scale = mode === 'playback' ? 1.15 : 1.1;

      anime({
        targets: eventElement,
        scale: scale,
        duration: duration,
        easing: 'easeOutQuart',
      });
    }
  }

  function handleMouseLeave() {
    if (isDestroyed) return;

    isHovering = false;
    dispatch('hover', { event, hovered: false });

    // Mode-aware hover out
    if (eventElement && !isDestroyed) {
      const duration = mode === 'playback' ? 200 : 100;

      anime({
        targets: eventElement,
        scale: 1.1,
        duration: duration,
        easing: 'easeOutQuart',
      });
    }
  }

  function getScaleSpecificContent(): { showContent: boolean; showIcon: boolean; showDetails: boolean } {
    switch (scale) {
      case '5m':
        return { showContent: true, showIcon: true, showDetails: true };
      case '15m':
        return { showContent: true, showIcon: true, showDetails: true };
      case '30m':
        return { showContent: true, showIcon: true, showDetails: false };
      case '1h':
        return { showContent: true, showIcon: true, showDetails: false };
      case '2h':
        return { showContent: false, showIcon: true, showDetails: false };
      case '6h':
      case '12h':
        return { showContent: false, showIcon: true, showDetails: false };
      case '24h':
      case '7d':
      case 'full':
        return { showContent: false, showIcon: false, showDetails: false };
      default:
        return { showContent: false, showIcon: true, showDetails: false };
    }
  }

  function formatEventTime(timestamp: string): string {
    const date = new Date(timestamp);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / (1000 * 60));
    const diffHours = Math.floor(diffMins / 60);
    const diffDays = Math.floor(diffHours / 24);

    if (diffDays > 0) {
      return `${diffDays}d ago`;
    } else if (diffHours > 0) {
      return `${diffHours}h ago`;
    } else if (diffMins > 0) {
      return `${diffMins}m ago`;
    } else {
      return 'Just now';
    }
  }

  function truncateText(text: string, maxLength: number): string {
    if (text.length <= maxLength) return text;
    return text.substring(0, maxLength - 3) + '...';
  }

  $: scaleContent = getScaleSpecificContent();
  $: eventColor = getEventColor(event.type);
  $: projectColor = getProjectColor(event.project);
  $: projectLaneColor = getProjectLaneColor(event.project);
  $: shouldShowContent = isHovering && !isDestroyed && (mode === 'playback' || scale === '5m' || scale === '15m');

  // Lifecycle management
  onMount(() => {
    isDestroyed = false;
  });

  onDestroy(() => {
    isDestroyed = true;
    isHovering = false;

    // Simple cleanup - just mark as destroyed
    if (eventElement) {
      eventElement.classList.add('destroyed');
    }
  });
</script>

<!-- Project-clustered positioning with context-switch visualization -->
  <div
    bind:this={eventElement}
    class="timeline-event"
    class:selected
    class:hovered
    class:large-scale={scale === '5m' || scale === '15m' || scale === '30m' || scale === '1h'}
    class:context-switch={isContextSwitch}
    class:destroyed={isDestroyed}
    class:playback-mode={mode === 'playback'}
    class:explorer-mode={mode === 'explorer'}
    style="
      left: {eventPosition.x}%;
      top: {eventPosition.y}px;
      --event-color: {eventColor};
      --project-color: {projectColor};
      --project-lane-color: {projectLaneColor};
      --event-size: {eventSize}px;
    "
    on:click={handleClick}
    on:mouseenter={handleMouseEnter}
    on:mouseleave={handleMouseLeave}
    on:keydown={e => e.key === 'Enter' && handleClick()}
    role="button"
    tabindex="0"
    title="{event.project}: {event.content} - {formatEventTime(event.timestamp)}">

    <!-- Project lane background -->
    <div class="project-lane-background"></div>
    <!-- Event core/icon -->
    <div class="event-core">
      {#if scaleContent.showIcon}
        <span class="event-icon">{getEventIcon(event.type)}</span>
      {/if}

      <!-- Event type indicator -->
      <div class="event-type-indicator" style="background-color: {eventColor}"></div>
    </div>

    <!-- Content bubble (shown on hover only) -->
    <div class="event-content" class:content-visible={shouldShowContent}>
        <div class="event-header">
          <span class="event-project" style="color: {projectColor}">
            {event.project}
          </span>
          <span class="event-time">
            {formatEventTime(event.timestamp)}
          </span>
        </div>

        <div class="event-text">
          {truncateText(event.content, scale === '15m' ? 150 : 80)}
        </div>

        {#if scaleContent.showDetails && event.tags && event.tags.length > 0}
          <div class="event-tags">
            {#each event.tags.slice(0, 3) as tag}
              <span class="event-tag">{tag}</span>
            {/each}
            {#if event.tags.length > 3}
              <span class="event-tag-more">+{event.tags.length - 3}</span>
            {/if}
          </div>
        {/if}
    </div>

    <!-- Connection points -->
    {#if event.metadata?.connections && event.metadata.connections.length > 0}
      <div class="connection-indicators">
        <div class="connection-dot in"></div>
        <div class="connection-dot out"></div>
      </div>
    {/if}

    <!-- Selection highlight -->
    {#if selected}
      <div class="selection-ring"></div>
    {/if}

    <!-- Hover glow -->
    {#if hovered}
      <div class="hover-glow"></div>
    {/if}
  </div>

<style>
  .timeline-event {
    position: absolute;
    cursor: pointer;
    z-index: 4;
    transform: translateX(-50%);
    transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1),
                opacity 0.2s ease,
                visibility 0.2s ease;
  }

  .timeline-event.explorer-mode {
    transition: transform 0.15s ease,
                opacity 0.15s ease,
                visibility 0.15s ease;
  }

  .timeline-event.playback-mode {
    transition: transform 0.4s cubic-bezier(0.4, 0, 0.2, 1),
                opacity 0.3s ease,
                visibility 0.3s ease;
  }

  .timeline-event.destroyed {
    pointer-events: none;
    opacity: 0;
    visibility: hidden;
  }

  .project-lane-background {
    position: absolute;
    left: -20px;
    right: -20px;
    top: -8px;
    bottom: -8px;
    background: var(--project-lane-color);
    border-radius: 4px;
    z-index: -1;
    opacity: 0;
    transition: opacity 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  }

  .timeline-event:hover .project-lane-background {
    opacity: 0.8;
    transition-delay: 0.15s;
  }

  .timeline-event.context-switch {
    transform: translateX(-50%) scale(1.1);
  }

  .timeline-event.context-switch .event-core {
    box-shadow:
      0 4px 12px rgba(0, 0, 0, 0.6),
      0 0 0 3px rgba(255, 255, 255, 0.8),
      0 0 20px var(--event-color),
      0 0 40px rgba(255, 255, 255, 0.3);
  }

  .timeline-event:hover {
    z-index: 10;
  }

  .timeline-event.selected {
    z-index: 15;
  }

  .event-core {
    position: relative;
    width: var(--event-size);
    height: var(--event-size);
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--event-color);
    border-radius: 50%;
    box-shadow:
      0 4px 12px rgba(0, 0, 0, 0.6),
      0 0 0 3px rgba(255, 255, 255, 0.6),
      0 0 16px var(--event-color);
    transition: transform 0.2s cubic-bezier(0.4, 0, 0.2, 1),
                box-shadow 0.2s cubic-bezier(0.4, 0, 0.2, 1);
    transform: scale(1.1);
  }

  .timeline-event.large-scale .event-core {
    border-radius: 8px;
  }

  .event-icon {
    font-size: calc(var(--event-size) * 0.6);
    line-height: 1;
    filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.8));
  }

  .event-type-indicator {
    position: absolute;
    top: -3px;
    right: -3px;
    width: 10px;
    height: 10px;
    border-radius: 50%;
    border: 2px solid var(--bg-primary, #1a1a1a);
    box-shadow: 0 0 6px var(--event-color);
  }

  .event-content {
    position: absolute;
    top: calc(var(--event-size) + 8px);
    left: 50%;
    transform: translateX(-50%) translateY(-8px) scale(0.95);
    background: rgba(26, 26, 26, 0.98);
    border: 2px solid var(--event-color);
    border-radius: 8px;
    padding: 8px 12px;
    max-width: 250px;
    min-width: 180px;
    box-shadow:
      0 4px 16px rgba(0, 0, 0, 0.6),
      0 0 0 1px rgba(255, 255, 255, 0.2),
      0 0 12px var(--event-color);
    backdrop-filter: blur(10px);
    z-index: 1;
    opacity: 0;
    visibility: hidden;
    transform: translateX(-50%) translateY(-8px) scale(0.95);
    transition:
      opacity 0.25s cubic-bezier(0.4, 0, 0.2, 1),
      visibility 0.25s cubic-bezier(0.4, 0, 0.2, 1),
      transform 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  }

  .timeline-event.large-scale .event-content {
    max-width: 320px;
    padding: 12px 16px;
  }

  .event-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 6px;
    font-size: 0.75rem;
  }

  .event-project {
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .event-time {
    color: var(--text-secondary, #cccccc);
    font-size: 0.7rem;
  }

  .event-text {
    color: var(--text-primary, #ffffff);
    font-size: 0.85rem;
    line-height: 1.4;
    margin-bottom: 8px;
  }

  .timeline-event.large-scale .event-text {
    font-size: 0.9rem;
    line-height: 1.5;
  }

  .event-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
    margin-top: 8px;
  }

  .event-tag {
    background: rgba(0, 255, 255, 0.2);
    color: #00ffff;
    font-size: 0.65rem;
    padding: 2px 6px;
    border-radius: 12px;
    border: 1px solid rgba(0, 255, 255, 0.5);
  }

  .event-tag-more {
    background: rgba(255, 255, 255, 0.1);
    color: var(--text-secondary, #cccccc);
    font-size: 0.65rem;
    padding: 2px 6px;
    border-radius: 12px;
  }

  .connection-indicators {
    position: absolute;
    top: 50%;
    transform: translateY(-50%);
  }

  .connection-dot {
    position: absolute;
    width: 4px;
    height: 4px;
    border-radius: 50%;
    background: var(--accent-color, #00ffff);
    box-shadow: 0 0 6px rgba(0, 255, 255, 1);
  }

  .connection-dot.in {
    left: calc(var(--event-size) / -2 - 8px);
  }

  .connection-dot.out {
    right: calc(var(--event-size) / -2 - 8px);
  }

  .selection-ring {
    position: absolute;
    top: -4px;
    left: -4px;
    right: -4px;
    bottom: -4px;
    border: 2px solid var(--accent-color, #00ffff);
    border-radius: 50%;
    animation: pulse-selection 1.5s ease-in-out infinite;
    box-shadow: 0 0 8px var(--accent-color, #00ffff);
  }

  .timeline-event.large-scale .selection-ring {
    border-radius: 12px;
  }

  .hover-glow {
    position: absolute;
    top: -6px;
    left: -6px;
    right: -6px;
    bottom: -6px;
    background: radial-gradient(circle, rgba(0, 255, 255, 0.4) 0%, rgba(0, 255, 255, 0.2) 50%, transparent 100%);
    border-radius: 50%;
    animation: pulse-glow 1s ease-in-out infinite;
  }

  .timeline-event.large-scale .hover-glow {
    border-radius: 16px;
  }

  @keyframes pulse-selection {
    0%,
    100% {
      transform: scale(1);
      opacity: 1;
    }
    50% {
      transform: scale(1.1);
      opacity: 0.7;
    }
  }

  @keyframes pulse-glow {
    0%,
    100% {
      opacity: 0.6;
      transform: scale(1);
    }
    50% {
      opacity: 0.8;
      transform: scale(1.05);
    }
  }

  /* Responsive adjustments */
  @media (max-width: 768px) {
    .event-content {
      max-width: 200px;
      min-width: 150px;
      padding: 6px 10px;
      font-size: 0.8rem;
    }

    .event-text {
      font-size: 0.8rem;
    }

    .event-header {
      font-size: 0.7rem;
    }

    .event-tag {
      font-size: 0.6rem;
      padding: 1px 4px;
    }
  }

  @media (max-width: 480px) {
    .event-content {
      max-width: 180px;
      min-width: 120px;
      padding: 4px 8px;
    }

    .event-text {
      font-size: 0.75rem;
      line-height: 1.3;
    }

    .timeline-event:not(.large-scale) .event-content {
      display: none;
    }
  }

  /* High contrast mode */
  @media (prefers-contrast: high) {
    .event-core {
      border: 2px solid #ffffff;
    }

    .event-content {
      background: #000000;
      border: 2px solid #ffffff;
    }

    .event-tag {
      background: #ffffff;
      color: #000000;
      border: 1px solid #000000;
    }
  }

  .event-content.content-visible {
    opacity: 1;
    visibility: visible;
    transform: translateX(-50%) translateY(0) scale(1);
    transition-delay: 0.1s;
  }

  /* Reduced motion */
  @media (prefers-reduced-motion: reduce) {
    .timeline-event,
    .event-core,
    .selection-ring,
    .hover-glow,
    .event-content {
      transition: none;
      animation: none;
    }
  }

  /* Focus styles for accessibility */
  .timeline-event:focus {
    outline: 2px solid var(--accent-color, #4ecdc4);
    outline-offset: 2px;
  }

  .timeline-event:focus .selection-ring {
    display: block;
  }
</style>
