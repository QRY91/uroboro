<script lang="ts">
  import { createEventDispatcher, onMount, onDestroy } from 'svelte';
  import { viewport } from '../stores/timeline';
  import { DEFAULT_TIMELINE_CONFIG } from '../types/timeline';
  import type { JourneyEvent } from '../types/timeline';

  const dispatch = createEventDispatcher();

  // Canvas references
  let canvasContainer: HTMLDivElement;
  let timelineCanvas: HTMLCanvasElement;

  let ctx: CanvasRenderingContext2D;

  // Props
  export let events: JourneyEvent[] = [];
  export let projectLanes: any[] = [];
  export let mode: 'explorer' | 'playback' = 'explorer';
  export let interactive: boolean = true;
  export let currentScale: string = '1h';
  export let isPlaying: boolean = false;
  export let playbackSpeed: number = 1;

  // Component state
  let canvasWidth = 0;
  let canvasHeight = 0;
  let interactionZones: InteractionZone[] = [];
  let hoveredEvent: JourneyEvent | null = null;
  let tooltipPosition = { x: 0, y: 0, visible: false };
  let animationFrameId: number;
  let lastRenderTime = 0;

  interface InteractionZone {
    x: number;
    y: number;
    width: number;
    height: number;
    event: JourneyEvent;
  }

  // Initialize canvas
  onMount(() => {
    if (timelineCanvas) {
      ctx = timelineCanvas.getContext('2d')!;
      setupCanvas();
      setupEventListeners();
      startRenderLoop();
    }
  });

  onDestroy(() => {
    if (animationFrameId) {
      cancelAnimationFrame(animationFrameId);
    }
  });

  function setupCanvas() {
    const container = canvasContainer;
    if (!container) return;

    const rect = container.getBoundingClientRect();
    canvasWidth = rect.width;
    canvasHeight = rect.height;

    // Set canvas size with device pixel ratio for crisp rendering
    const dpr = window.devicePixelRatio || 1;
    timelineCanvas.width = canvasWidth * dpr;
    timelineCanvas.height = canvasHeight * dpr;
    timelineCanvas.style.width = `${canvasWidth}px`;
    timelineCanvas.style.height = `${canvasHeight}px`;

    ctx.scale(dpr, dpr);
  }

  function setupEventListeners() {
    window.addEventListener('resize', handleResize);
  }

  function handleResize() {
    setupCanvas();
    generateInteractionZones();
  }

  function startRenderLoop() {
    function render(timestamp: number) {
      if (timestamp - lastRenderTime > 16) { // ~60fps
        renderTimeline();
        lastRenderTime = timestamp;
      }
      animationFrameId = requestAnimationFrame(render);
    }
    animationFrameId = requestAnimationFrame(render);
  }

  function renderTimeline() {
    if (!ctx || !canvasWidth || !canvasHeight) return;

    // Clear canvas
    ctx.clearRect(0, 0, canvasWidth, canvasHeight);

    // Draw project lanes background
    drawProjectLanes();

    // Draw timeline events
    drawTimelineEvents();

    // Draw connections (if any)
    drawEventConnections();

    // Update interaction zones
    generateInteractionZones();
  }

  function drawProjectLanes() {
    const laneHeight = Math.min(80, Math.max(40, canvasHeight / Math.max(projectLanes.length, 4)));

    projectLanes.forEach((lane, index) => {
      const y = 60 + (index * laneHeight);
      const projectColor = getProjectColor(lane.project);

      // Lane background
      ctx.fillStyle = projectColor + '15'; // 15% opacity
      ctx.fillRect(0, y, canvasWidth, laneHeight);

      // Lane border
      ctx.strokeStyle = projectColor + '40';
      ctx.lineWidth = 1;
      ctx.beginPath();
      ctx.moveTo(0, y);
      ctx.lineTo(canvasWidth, y);
      ctx.stroke();

      // Lane label background (left side)
      ctx.fillStyle = projectColor + '25';
      ctx.fillRect(0, y, 150, laneHeight);

      // Lane text
      ctx.fillStyle = projectColor;
      ctx.font = '500 12px Inter, sans-serif';
      ctx.textAlign = 'left';
      ctx.textBaseline = 'middle';
      ctx.fillText(lane.project, 12, y + laneHeight / 2);
    });
  }

  function drawTimelineEvents() {
    if (!$viewport.startTime || !$viewport.endTime) return;

    const viewportStart = $viewport.startTime.getTime();
    const viewportEnd = $viewport.endTime.getTime();
    const viewportDuration = viewportEnd - viewportStart;
    const drawableWidth = canvasWidth - 160; // Account for lane labels

    events.forEach(event => {
      const eventTime = new Date(event.timestamp).getTime();

      // Skip events outside viewport
      if (eventTime < viewportStart || eventTime > viewportEnd) return;

      // Calculate position
      const x = 160 + ((eventTime - viewportStart) / viewportDuration) * drawableWidth;
      const lane = projectLanes.find(l => l.project === event.project);
      const laneHeight = Math.min(80, Math.max(40, canvasHeight / Math.max(projectLanes.length, 4)));
      const y = 60 + (lane?.laneIndex || 0) * laneHeight + laneHeight / 2;

      // Draw event
      drawEvent(x, y, event);
    });
  }

  function drawEvent(x: number, y: number, event: JourneyEvent) {
    const eventColor = getEventColor(event.type);
    const projectColor = getProjectColor(event.project);
    const isHovered = hoveredEvent?.id === event.id;

    // Mode-aware sizing
    const baseRadius = mode === 'playback' ? 7 : 6;
    const radius = isHovered ? baseRadius + 2 : baseRadius;

    // Event dot background with mode-aware opacity
    const bgOpacity = mode === 'playback' ? '30' : '40';
    ctx.fillStyle = projectColor + bgOpacity;
    ctx.beginPath();
    ctx.arc(x, y, radius + 2, 0, 2 * Math.PI);
    ctx.fill();

    // Event dot
    ctx.fillStyle = projectColor;
    ctx.beginPath();
    ctx.arc(x, y, radius, 0, 2 * Math.PI);
    ctx.fill();

    // Event border with mode-aware styling
    ctx.strokeStyle = isHovered ? '#ffffff' : projectColor;
    ctx.lineWidth = isHovered ? 2 : (mode === 'playback' ? 1.5 : 1);
    ctx.stroke();

    // Context switch indicator
    if (event.type === 'context_switch') {
      ctx.strokeStyle = '#ff6b35';
      ctx.lineWidth = mode === 'playback' ? 3 : 2;
      ctx.setLineDash([4, 4]);
      ctx.beginPath();
      ctx.arc(x, y, radius + 6, 0, 2 * Math.PI);
      ctx.stroke();
      ctx.setLineDash([]);
    }
  }

  function drawEventConnections() {
    // Draw connections between related events
    events.forEach(event => {
      if (event.metadata?.connections) {
        event.metadata.connections.forEach((connectionId: string) => {
          const connectedEvent = events.find(e => e.id === connectionId);
          if (connectedEvent) {
            drawConnection(event, connectedEvent);
          }
        });
      }
    });
  }

  function drawConnection(fromEvent: JourneyEvent, toEvent: JourneyEvent) {
    const fromPos = getEventPosition(fromEvent);
    const toPos = getEventPosition(toEvent);

    if (!fromPos || !toPos) return;

    ctx.strokeStyle = '#64748b40';
    ctx.lineWidth = 2;
    ctx.setLineDash([5, 5]);
    ctx.beginPath();
    ctx.moveTo(fromPos.x, fromPos.y);
    ctx.lineTo(toPos.x, toPos.y);
    ctx.stroke();
    ctx.setLineDash([]);
  }

  function getEventPosition(event: JourneyEvent): {x: number, y: number} | null {
    if (!$viewport.startTime || !$viewport.endTime) return null;

    const viewportStart = $viewport.startTime.getTime();
    const viewportEnd = $viewport.endTime.getTime();
    const viewportDuration = viewportEnd - viewportStart;
    const drawableWidth = canvasWidth - 160;
    const eventTime = new Date(event.timestamp).getTime();

    if (eventTime < viewportStart || eventTime > viewportEnd) return null;

    const x = 160 + ((eventTime - viewportStart) / viewportDuration) * drawableWidth;
    const lane = projectLanes.find(l => l.project === event.project);
    const laneHeight = Math.min(80, Math.max(40, canvasHeight / Math.max(projectLanes.length, 4)));
    const y = 60 + (lane?.laneIndex || 0) * laneHeight + laneHeight / 2;

    return { x, y };
  }

  function generateInteractionZones() {
    if (!$viewport.startTime || !$viewport.endTime || !interactive) {
      interactionZones = [];
      return;
    }

    // For explorer mode during scrolling, reduce interaction zones for performance
    const zoneSize = mode === 'explorer' && !interactive ? 16 : 24;
    const zoneOffset = zoneSize / 2;

    interactionZones = events
      .map(event => {
        const pos = getEventPosition(event);
        if (!pos) return null;

        return {
          x: pos.x - zoneOffset,
          y: pos.y - zoneOffset,
          width: zoneSize,
          height: zoneSize,
          event
        };
      })
      .filter(zone => zone !== null) as InteractionZone[];
  }

  function handleZoneMouseEnter(zone: InteractionZone, mouseEvent: MouseEvent) {
    hoveredEvent = zone.event;
    tooltipPosition = {
      x: mouseEvent.clientX + 10,
      y: mouseEvent.clientY - 10,
      visible: true
    };
    dispatch('eventHover', { event: zone.event, position: tooltipPosition });
  }

  function handleZoneMouseLeave() {
    hoveredEvent = null;
    tooltipPosition.visible = false;
    dispatch('eventHover', { event: null });
  }

  function handleZoneClick(zone: InteractionZone) {
    dispatch('eventClick', { event: zone.event });
  }

  function getProjectColor(project: string): string {
    const hash = project.split('').reduce((a, b) => ((a << 5) - a + b.charCodeAt(0)) & 0xffffffff, 0);
    const colors = DEFAULT_TIMELINE_CONFIG.colorScheme.projectColors;
    return colors[Math.abs(hash) % colors.length];
  }

  function getEventColor(eventType: string): string {
    return DEFAULT_TIMELINE_CONFIG.colorScheme.eventColors[eventType] || DEFAULT_TIMELINE_CONFIG.colorScheme.eventColors.milestone;
  }

  // Reactive updates
  $: if (events || projectLanes || $viewport || interactive || mode) {
    generateInteractionZones();
  }
</script>

<div class="canvas-timeline-container" bind:this={canvasContainer}>
  <!-- Canvas layer for timeline visualization -->
  <canvas
    bind:this={timelineCanvas}
    class="timeline-canvas"
  ></canvas>

  <!-- DOM overlay for interactions -->
  <div class="interaction-overlay">
    {#each interactionZones as zone}
      <div
        class="interaction-zone"
        style="left: {zone.x}px; top: {zone.y}px; width: {zone.width}px; height: {zone.height}px;"
        on:mouseenter={(e) => handleZoneMouseEnter(zone, e)}
        on:mouseleave={handleZoneMouseLeave}
        on:click={() => handleZoneClick(zone)}
        role="button"
        tabindex="0"
        aria-label="Timeline event: {zone.event.content} in {zone.event.project}"
      ></div>
    {/each}
  </div>

  <!-- Tooltip -->
  {#if tooltipPosition.visible && hoveredEvent}
    <div
      class="canvas-tooltip"
      style="left: {tooltipPosition.x}px; top: {tooltipPosition.y}px;"
    >
      <div class="tooltip-header">
        <span class="tooltip-project" style="color: {getProjectColor(hoveredEvent.project)}">
          {hoveredEvent.project}
        </span>
        <span class="tooltip-time">
          {new Date(hoveredEvent.timestamp).toLocaleTimeString()}
        </span>
      </div>
      <div class="tooltip-content">
        <div class="tooltip-action">{hoveredEvent.content}</div>
        {#if hoveredEvent.description}
          <div class="tooltip-details">{hoveredEvent.description}</div>
        {/if}
        {#if hoveredEvent.tags && hoveredEvent.tags.length > 0}
          <div class="tooltip-tags">
            {#each hoveredEvent.tags.slice(0, 3) as tag}
              <span class="tooltip-tag">{tag}</span>
            {/each}
          </div>
        {/if}
      </div>
    </div>
  {/if}
</div>

<style>
  .canvas-timeline-container {
    position: relative;
    width: 100%;
    height: 100%;
    overflow: hidden;
  }

  .timeline-canvas {
    display: block;
    width: 100%;
    height: 100%;
    cursor: crosshair;
  }

  .interaction-overlay {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    pointer-events: none;
  }

  .interaction-zone {
    position: absolute;
    pointer-events: all;
    cursor: pointer;
    border-radius: 50%;
    transition: background-color 0.2s ease;
  }

  .interaction-zone:hover {
    background-color: rgba(255, 255, 255, 0.1);
  }

  .interaction-zone:focus {
    outline: 2px solid #3b82f6;
    outline-offset: 2px;
  }

  .canvas-tooltip {
    position: fixed;
    z-index: 1000;
    background: rgba(0, 0, 0, 0.9);
    border: 1px solid rgba(255, 255, 255, 0.2);
    border-radius: 8px;
    padding: 12px;
    max-width: 300px;
    box-shadow: 0 10px 25px rgba(0, 0, 0, 0.5);
    backdrop-filter: blur(10px);
    font-size: 12px;
    line-height: 1.4;
    pointer-events: none;
  }

  .tooltip-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 8px;
    padding-bottom: 6px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  }

  .tooltip-project {
    font-weight: 600;
    font-size: 11px;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .tooltip-time {
    color: #9ca3af;
    font-size: 10px;
    font-family: 'JetBrains Mono', monospace;
  }

  .tooltip-action {
    color: #ffffff;
    font-weight: 500;
    margin-bottom: 4px;
  }

  .tooltip-details {
    color: #d1d5db;
    font-size: 11px;
    margin-bottom: 6px;
  }

  .tooltip-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
  }

  .tooltip-tag {
    background: rgba(59, 130, 246, 0.2);
    color: #93c5fd;
    padding: 2px 6px;
    border-radius: 4px;
    font-size: 10px;
    font-weight: 500;
  }

  /* High contrast mode support */
  @media (prefers-contrast: high) {
    .canvas-tooltip {
      background: #000000;
      border: 2px solid #ffffff;
    }

    .interaction-zone:focus {
      outline: 3px solid #ffff00;
    }
  }

  /* Reduced motion support */
  @media (prefers-reduced-motion: reduce) {
    .interaction-zone {
      transition: none;
    }
  }
</style>
