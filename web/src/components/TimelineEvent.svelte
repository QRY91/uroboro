<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { JourneyEvent, TimelineViewport, TimeScaleName } from '@types/timeline';
  import { DEFAULT_TIMELINE_CONFIG } from '@types/timeline';
  import anime from 'animejs';

  const dispatch = createEventDispatcher();

  // Props
  export let event: JourneyEvent;
  export let scale: TimeScaleName;
  export let viewport: TimelineViewport;
  export let selected = false;
  export let hovered = false;

  // Component state
  let eventElement: HTMLDivElement;
  let eventPosition = { x: 0, y: 0 };
  let eventSize = 16;

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

    // Calculate horizontal position
    const relativePosition = (eventTime - viewportStart) / viewportDuration;
    const newX = relativePosition * 100; // Percentage

    // Calculate vertical position (simple clustering by project)
    const projectHash = hashString(event.project);
    const newY = 80 + (projectHash % 4) * 50; // Distribute across 4 lanes, staying within viewport

    // Update position and trigger reactivity
    eventPosition = { x: newX, y: newY };
  }

  function updateEventSize() {
    eventSize = DEFAULT_TIMELINE_CONFIG.eventSizes[scale] || 16;
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

  function handleClick() {
    dispatch('click', { event, element: eventElement });

    // Animate click feedback
    if (eventElement) {
      anime({
        targets: eventElement,
        scale: [1, 1.2, 1],
        duration: 200,
        easing: 'easeOutBack',
      });
    }
  }

  function handleMouseEnter() {
    dispatch('hover', { event, hovered: true });

    // Animate hover effect
    if (eventElement) {
      anime({
        targets: eventElement,
        scale: 1.15,
        duration: 150,
        easing: 'easeOutQuart',
      });
    }
  }

  function handleMouseLeave() {
    dispatch('hover', { event, hovered: false });

    // Animate hover out
    if (eventElement) {
      anime({
        targets: eventElement,
        scale: 1.1,
        duration: 150,
        easing: 'easeOutQuart',
      });
    }
  }

  function getScaleSpecificContent(): { showContent: boolean; showIcon: boolean; showDetails: boolean } {
    switch (scale) {
      case '15m':
        return { showContent: true, showIcon: true, showDetails: true };
      case '1h':
        return { showContent: true, showIcon: true, showDetails: false };
      case '6h':
        return { showContent: false, showIcon: true, showDetails: false };
      case '24h':
      case '7d':
      case 'full':
        return { showContent: false, showIcon: true, showDetails: false };
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
</script>

<!-- Always show events passed to this component - filtering happens at parent level -->
  <div
    bind:this={eventElement}
    class="timeline-event"
    class:selected
    class:hovered
    class:large-scale={scale === '15m' || scale === '1h'}
    style="
      left: {eventPosition.x}%;
      top: {eventPosition.y}px;
      --event-color: {eventColor};
      --project-color: {projectColor};
      --event-size: {eventSize}px;
    "
    on:click={handleClick}
    on:mouseenter={handleMouseEnter}
    on:mouseleave={handleMouseLeave}
    on:keydown={e => e.key === 'Enter' && handleClick()}
    role="button"
    tabindex="0"
    title="{event.content} - {formatEventTime(event.timestamp)}">
    <!-- Event core/icon -->
    <div class="event-core">
      {#if scaleContent.showIcon}
        <span class="event-icon">{getEventIcon(event.type)}</span>
      {/if}

      <!-- Event type indicator -->
      <div class="event-type-indicator" style="background-color: {eventColor}"></div>
    </div>

    <!-- Content bubble (shown at larger scales) -->
    {#if scaleContent.showContent}
      <div class="event-content">
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
    {/if}

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
    transition: all 0.2s ease;
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
    transition: all 0.2s ease;
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
    transform: translateX(-50%);
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

  /* Reduced motion */
  @media (prefers-reduced-motion: reduce) {
    .timeline-event,
    .event-core,
    .selection-ring,
    .hover-glow {
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
