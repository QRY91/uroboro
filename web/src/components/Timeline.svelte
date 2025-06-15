<script lang="ts">
  import { onMount, onDestroy, createEventDispatcher } from 'svelte';
  import { timelineState, timelineActions, eventsInCurrentViewport, controls, viewport } from '@stores/timeline';
  import type { TimeScaleName, JourneyEvent, TouchGesture } from '@types/timeline';
  import { TIME_SCALES } from '@types/timeline';
  import TimelineEvent from './TimelineEvent.svelte';
  import TimelineRuler from './TimelineRuler.svelte';
  import ViewportScrubber from './ViewportScrubber.svelte';
  import anime from 'animejs';

  const dispatch = createEventDispatcher();

  // Component state
  let timelineContainer: HTMLDivElement;
  let timelineCanvas: HTMLCanvasElement;
  let connectionCanvas: HTMLCanvasElement;
  let ctx: CanvasRenderingContext2D | null = null;

  // Interaction state
  let isDragging = false;
  let dragStartX = 0;
  let dragStartPosition = 0;
  let isZooming = false;
  let touchStartDistance = 0;
  let lastTouchTime = 0;

  // Performance state
  let animationFrameId: number;
  let lastRenderTime = 0;
  let renderThrottleMs = 16; // 60fps
  let mounted = false;
  let initializationComplete = false;

  // Reactive statements with timeline readiness check
  $: timelineReady = mounted && $viewport?.startTime && $viewport?.endTime && $controls?.currentScale && !$isLoading;
  $: eventsToRender = (timelineReady && initializationComplete) ? $eventsInCurrentViewport : [];

  // Add delay after timeline becomes ready to prevent ghost events
  $: if (timelineReady && !initializationComplete) {
    setTimeout(() => {
      initializationComplete = true;
    }, 100); // Small delay to ensure viewport is fully set up
  }
  $: currentScale = $controls.currentScale;
  $: isPlaying = $controls.isPlaying;
  $: viewportPosition = $controls.viewportPosition;

  // Lifecycle
  onMount(() => {
    initializeCanvas();
    setupEventListeners();
    setupKeyboardShortcuts();

    // Start render loop
    startRenderLoop();
  });

  onDestroy(() => {
    cleanup();
  });

  // Canvas initialization
  function initializeCanvas() {
    if (!connectionCanvas) return;

    ctx = connectionCanvas.getContext('2d');
    if (!ctx) return;

    // Set up high DPI support
    const devicePixelRatio = window.devicePixelRatio || 1;
    const rect = timelineContainer.getBoundingClientRect();

    connectionCanvas.width = rect.width * devicePixelRatio;
    connectionCanvas.height = rect.height * devicePixelRatio;
    connectionCanvas.style.width = rect.width + 'px';
    connectionCanvas.style.height = rect.height + 'px';

    ctx.scale(devicePixelRatio, devicePixelRatio);

    // Update render context
    timelineActions.updateRenderContext({
      bounds: {
        left: rect.left,
        top: rect.top,
        right: rect.right,
        bottom: rect.bottom,
        width: rect.width,
        height: rect.height,
      },
      devicePixelRatio,
    });
  }

  // Event listeners setup
  function setupEventListeners() {
    if (!timelineContainer) return;

    // Mouse events
    timelineContainer.addEventListener('mousedown', handleMouseDown);
    timelineContainer.addEventListener('mousemove', handleMouseMove);
    timelineContainer.addEventListener('mouseup', handleMouseUp);
    timelineContainer.addEventListener('wheel', handleWheel, { passive: false });

    // Touch events
    timelineContainer.addEventListener('touchstart', handleTouchStart, { passive: false });
    timelineContainer.addEventListener('touchmove', handleTouchMove, { passive: false });
    timelineContainer.addEventListener('touchend', handleTouchEnd);

    // Window events
    window.addEventListener('resize', handleResize);
    window.addEventListener('mousemove', handleGlobalMouseMove);
    window.addEventListener('mouseup', handleGlobalMouseUp);
  }

  function setupKeyboardShortcuts() {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Only handle shortcuts when timeline is focused or no input is focused
      const activeElement = document.activeElement;
      const isInputFocused =
        activeElement &&
        (activeElement.tagName === 'INPUT' ||
          activeElement.tagName === 'TEXTAREA' ||
          activeElement.getAttribute('contenteditable') === 'true');

      if (isInputFocused) return;

      e.preventDefault();
      timelineActions.handleKeyboardShortcut(e.key, e.ctrlKey, e.shiftKey);
    };

    window.addEventListener('keydown', handleKeyDown);

    return () => window.removeEventListener('keydown', handleKeyDown);
  }

  // Mouse interaction handlers
  function handleMouseDown(e: MouseEvent) {
    if (e.button !== 0) return; // Only left mouse button

    isDragging = true;
    dragStartX = e.clientX;
    dragStartPosition = $controls.viewportPosition;

    timelineContainer.style.cursor = 'grabbing';

    // Prevent text selection
    e.preventDefault();
  }

  function handleMouseMove(e: MouseEvent) {
    if (!isDragging) return;

    const deltaX = e.clientX - dragStartX;
    const containerWidth = timelineContainer.clientWidth;
    const deltaPosition = deltaX / containerWidth;

    timelineActions.setViewportPosition(dragStartPosition - deltaPosition);
  }

  function handleMouseUp() {
    if (!isDragging) return;

    isDragging = false;
    timelineContainer.style.cursor = 'grab';
  }

  function handleGlobalMouseMove(e: MouseEvent) {
    if (isDragging) {
      handleMouseMove(e);
    }
  }

  function handleGlobalMouseUp() {
    handleMouseUp();
  }

  // Wheel/scroll handling for zoom
  function handleWheel(e: WheelEvent) {
    e.preventDefault();

    const zoomSensitivity = 0.001;
    const delta = e.deltaY * zoomSensitivity;

    if (e.ctrlKey || e.metaKey) {
      // Zoom in/out by changing time scale
      const scaleOrder: TimeScaleName[] = ['15m', '1h', '6h', '24h', '7d', 'full'];
      const currentIndex = scaleOrder.indexOf(currentScale);

      if (delta > 0 && currentIndex < scaleOrder.length - 1) {
        // Zoom out
        timelineActions.setViewportScale(scaleOrder[currentIndex + 1]);
      } else if (delta < 0 && currentIndex > 0) {
        // Zoom in
        timelineActions.setViewportScale(scaleOrder[currentIndex - 1]);
      }
    } else {
      // Pan horizontally
      const panSensitivity = 0.05;
      timelineActions.panViewport(delta * panSensitivity);
    }
  }

  // Touch interaction handlers
  function handleTouchStart(e: TouchEvent) {
    e.preventDefault();

    if (e.touches.length === 1) {
      // Single touch - pan
      const touch = e.touches[0];
      isDragging = true;
      dragStartX = touch.clientX;
      dragStartPosition = $controls.viewportPosition;
      lastTouchTime = Date.now();
    } else if (e.touches.length === 2) {
      // Two touches - zoom
      isZooming = true;
      touchStartDistance = getTouchDistance(e.touches);
    }
  }

  function handleTouchMove(e: TouchEvent) {
    e.preventDefault();

    if (e.touches.length === 1 && isDragging && !isZooming) {
      // Pan with single touch
      const touch = e.touches[0];
      const deltaX = touch.clientX - dragStartX;
      const containerWidth = timelineContainer.clientWidth;
      const deltaPosition = deltaX / containerWidth;

      timelineActions.setViewportPosition(dragStartPosition - deltaPosition);
    } else if (e.touches.length === 2 && isZooming) {
      // Zoom with two touches
      const currentDistance = getTouchDistance(e.touches);
      const scaleChange = currentDistance / touchStartDistance;

      if (scaleChange > 1.2) {
        // Zoom in
        zoomIn();
        touchStartDistance = currentDistance;
      } else if (scaleChange < 0.8) {
        // Zoom out
        zoomOut();
        touchStartDistance = currentDistance;
      }
    }
  }

  function handleTouchEnd(e: TouchEvent) {
    const touchEndTime = Date.now();
    const touchDuration = touchEndTime - lastTouchTime;

    if (e.touches.length === 0) {
      // Check for double tap
      if (touchDuration < 300 && !isDragging) {
        // Double tap to zoom in
        zoomIn();
      }

      isDragging = false;
      isZooming = false;
    }
  }

  // Utility functions
  function getTouchDistance(touches: TouchList): number {
    const dx = touches[0].clientX - touches[1].clientX;
    const dy = touches[0].clientY - touches[1].clientY;
    return Math.sqrt(dx * dx + dy * dy);
  }

  function zoomIn() {
    const scaleOrder: TimeScaleName[] = ['7d', '24h', '6h', '1h', '15m'];
    const currentIndex = scaleOrder.indexOf(currentScale);
    if (currentIndex < scaleOrder.length - 1) {
      timelineActions.setViewportScale(scaleOrder[currentIndex + 1]);
    }
  }

  function zoomOut() {
    const scaleOrder: TimeScaleName[] = ['15m', '1h', '6h', '24h', '7d', 'full'];
    const currentIndex = scaleOrder.indexOf(currentScale);
    if (currentIndex < scaleOrder.length - 1) {
      timelineActions.setViewportScale(scaleOrder[currentIndex + 1]);
    }
  }

  function handleResize() {
    // Debounce resize events
    clearTimeout(animationFrameId);
    animationFrameId = window.setTimeout(() => {
      initializeCanvas();
    }, 100);
  }

  // Scale selector handlers
  function handleScaleChange(e: Event) {
    const select = e.target as HTMLSelectElement;
    timelineActions.setViewportScale(select.value as TimeScaleName);
  }

  // Playback control handlers
  function togglePlayback() {
    if (isPlaying) {
      timelineActions.pause();
    } else {
      timelineActions.play();
    }
  }

  function handleSpeedChange(e: Event) {
    const slider = e.target as HTMLInputElement;
    timelineActions.setPlaySpeed(parseFloat(slider.value));
  }

  // Rendering
  function startRenderLoop() {
    const render = (timestamp: number) => {
      if (timestamp - lastRenderTime >= renderThrottleMs) {
        renderConnections();
        lastRenderTime = timestamp;
      }

      animationFrameId = requestAnimationFrame(render);
    };

    animationFrameId = requestAnimationFrame(render);
  }

  function renderConnections() {
    if (!ctx || !eventsToRender.length) return;

    ctx.clearRect(0, 0, connectionCanvas.width, connectionCanvas.height);

    // Draw connections between related events
    ctx.strokeStyle = 'rgba(78, 205, 196, 0.3)';
    ctx.lineWidth = 1;

    eventsToRender.forEach((event, index) => {
      if (event.metadata?.connections) {
        event.metadata.connections.forEach(connectionId => {
          const connectedEvent = eventsToRender.find(e => e.id === connectionId);
          if (connectedEvent) {
            drawConnection(event, connectedEvent);
          }
        });
      }
    });
  }

  function drawConnection(event1: JourneyEvent, event2: JourneyEvent) {
    if (!ctx) return;

    // Calculate positions (simplified - would need actual event element positions)
    const containerRect = timelineContainer.getBoundingClientRect();
    const event1Time = new Date(event1.timestamp).getTime();
    const event2Time = new Date(event2.timestamp).getTime();

    if (!$viewport.startTime || !$viewport.endTime) return;

    const viewportStart = $viewport.startTime.getTime();
    const viewportEnd = $viewport.endTime.getTime();
    const viewportDuration = viewportEnd - viewportStart;

    const x1 = ((event1Time - viewportStart) / viewportDuration) * containerRect.width;
    const x2 = ((event2Time - viewportStart) / viewportDuration) * containerRect.width;

    const y1 = containerRect.height * 0.4; // Simplified positioning
    const y2 = containerRect.height * 0.6;

    ctx.beginPath();
    ctx.moveTo(x1, y1);
    ctx.bezierCurveTo(x1, y1 + 50, x2, y2 - 50, x2, y2);
    ctx.stroke();
  }

  function cleanup() {
    if (animationFrameId) {
      cancelAnimationFrame(animationFrameId);
    }

    window.removeEventListener('resize', handleResize);
    window.removeEventListener('mousemove', handleGlobalMouseMove);
    window.removeEventListener('mouseup', handleGlobalMouseUp);
  }

  // Format time helpers
  function formatTime(date: Date | null): string {
    if (!date) return '--:--';
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  }

  function formatDuration(ms: number): string {
    const seconds = Math.floor(ms / 1000);
    const minutes = Math.floor(seconds / 60);
    const hours = Math.floor(minutes / 60);
    const days = Math.floor(hours / 24);

    if (days > 0) {
      return `${days}d ${hours % 24}h`;
    } else if (hours > 0) {
      return `${hours}h ${minutes % 60}m`;
    } else if (minutes > 0) {
      return `${minutes}m`;
    } else {
      return `${seconds}s`;
    }
  }
</script>

<div class="timeline-container" bind:this={timelineContainer}>
  <!-- Timeline Controls Panel -->
  <div class="timeline-viewport-controls">
    <div class="scale-selector">
      <label for="timeScaleSelect">Time Scale:</label>
      <select id="timeScaleSelect" value={currentScale} on:change={handleScaleChange}>
        {#each Object.entries(TIME_SCALES) as [value, scale]}
          <option {value}>{scale.label}</option>
        {/each}
      </select>
    </div>

    <div class="viewport-scrubber">
      <ViewportScrubber />
      <div class="time-labels">
        <span id="viewportStartTime">{formatTime($viewport.startTime)}</span>
        <span id="viewportEndTime">{formatTime($viewport.endTime)}</span>
      </div>
    </div>

    <div class="playback-controls">
      <button on:click={togglePlayback} class="btn-primary" title="Play/Pause">
        {isPlaying ? '⏸' : '▶'}
      </button>
      <button on:click={() => timelineActions.restart()} class="btn-secondary" title="Restart"> ⏮ </button>
      <div class="speed-control">
        <label for="speedSlider">Speed:</label>
        <input
          id="speedSlider"
          type="range"
          min="0.5"
          max="3"
          step="0.5"
          value={$controls.playSpeed}
          on:input={handleSpeedChange} />
        <span>{$controls.playSpeed}x</span>
      </div>
    </div>

    <div class="zoom-controls">
      <button on:click={zoomOut} class="btn-secondary" title="Zoom Out">-</button>
      <span>100%</span>
      <button on:click={zoomIn} class="btn-secondary" title="Zoom In">+</button>
    </div>
  </div>

  <!-- Timeline Visualization Area -->
  <div class="timeline-main">
    <!-- Timeline Ruler -->
    <TimelineRuler />

    <!-- Connection Canvas -->
    <canvas bind:this={connectionCanvas} class="connection-canvas" style="pointer-events: none;"></canvas>

    <!-- Timeline Events -->
    <div class="timeline-events">
      {#each eventsToRender as event (event.id)}
        <TimelineEvent
          {event}
          scale={currentScale}
          viewport={$viewport}
          on:click={e => dispatch('eventClick', e.detail)}
          on:hover={e => dispatch('eventHover', e.detail)} />
      {/each}
    </div>

    <!-- Bottom Time Axis -->
    <div class="bottom-time-axis">
      {#if $viewport.startTime && $viewport.endTime}
        {@const startTime = $viewport.startTime}
        {@const endTime = $viewport.endTime}
        {@const duration = endTime.getTime() - startTime.getTime()}
        <div class="time-axis-line"></div>
        <div class="time-axis-labels">
          <span class="time-start">
            {startTime.toLocaleDateString([], { month: 'short', day: 'numeric' })}
            {startTime.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', hour12: false })}
          </span>
          <span class="time-duration">
            {formatDuration(duration)}
          </span>
          <span class="time-end">
            {endTime.toLocaleDateString([], { month: 'short', day: 'numeric' })}
            {endTime.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', hour12: false })}
          </span>
        </div>
      {/if}
    </div>

    <!-- Loading State -->
    {#if $timelineState.loading}
      <div class="timeline-loading">
        <div class="loading-spinner"></div>
        <span>Loading timeline...</span>
      </div>
    {/if}

    <!-- Error State -->
    {#if $timelineState.error}
      <div class="timeline-error">
        <span>Error: {$timelineState.error}</span>
      </div>
    {/if}
  </div>
</div>

<style>
  .timeline-container {
    width: 100%;
    height: 100vh;
    background: var(--bg-primary, #1a1a1a);
    color: var(--text-primary, #ffffff);
    display: flex;
    flex-direction: column;
    position: relative;
    cursor: grab;
    user-select: none;
  }

  .timeline-container:active {
    cursor: grabbing;
  }

  .timeline-viewport-controls {
    display: flex;
    align-items: center;
    gap: 2rem;
    padding: 1rem;
    background: rgba(20, 20, 20, 0.95);
    border-bottom: 1px solid var(--border-color, #333);
    backdrop-filter: blur(10px);
    z-index: 10;
  }

  .scale-selector {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .scale-selector label {
    font-size: 0.875rem;
    color: var(--text-secondary, #cccccc);
  }

  .scale-selector select {
    background: var(--bg-secondary, #2a2a2a);
    border: 1px solid var(--border-color, #333);
    color: var(--text-primary, #ffffff);
    padding: 0.5rem;
    border-radius: 4px;
    font-size: 0.875rem;
  }

  .viewport-scrubber {
    flex: 1;
    min-width: 300px;
  }

  .time-labels {
    display: flex;
    justify-content: space-between;
    font-size: 0.75rem;
    color: var(--accent-color, #4ecdc4);
    margin-top: 0.25rem;
  }

  .playback-controls {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .speed-control {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin-left: 1rem;
  }

  .speed-control label {
    font-size: 0.875rem;
    color: var(--text-secondary, #cccccc);
  }

  .speed-control input[type='range'] {
    width: 80px;
  }

  .speed-control span {
    font-size: 0.875rem;
    color: var(--accent-color, #4ecdc4);
    min-width: 30px;
  }

  .zoom-controls {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .btn-primary,
  .btn-secondary {
    padding: 0.5rem 1rem;
    border-radius: 4px;
    border: none;
    cursor: pointer;
    font-size: 0.875rem;
    transition: all 0.2s ease;
  }

  .btn-primary {
    background: var(--accent-color, #4ecdc4);
    color: var(--bg-primary, #1a1a1a);
  }

  .btn-primary:hover {
    background: var(--accent-hover, #45b7b8);
  }

  .btn-secondary {
    background: var(--bg-secondary, #2a2a2a);
    color: var(--text-primary, #ffffff);
    border: 1px solid var(--border-color, #333);
  }

  .btn-secondary:hover {
    background: var(--bg-tertiary, #3a3a3a);
  }

  .timeline-main {
    flex: 1;
    position: relative;
    overflow: hidden;
  }

  .connection-canvas {
    position: absolute;
    top: 60px;
    left: 0;
    width: 100%;
    height: calc(100% - 110px);
    pointer-events: none;
    z-index: 1;
  }

  .timeline-events {
    position: absolute;
    top: 60px;
    left: 0;
    width: 100%;
    height: calc(100% - 110px);
    z-index: 2;
    overflow: visible;
  }

  .bottom-time-axis {
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
    height: 50px;
    background: linear-gradient(
      to top,
      rgba(26, 26, 26, 0.98) 0%,
      rgba(26, 26, 26, 0.9) 70%,
      rgba(26, 26, 26, 0.6) 100%
    );
    border-top: 3px solid var(--accent-color, #00ffff);
    box-shadow: 0 -2px 10px rgba(0, 255, 255, 0.4);
    z-index: 6;
    display: flex;
    flex-direction: column;
    justify-content: center;
    padding: 0 1rem;
  }

  .time-axis-line {
    height: 3px;
    background: var(--accent-color, #00ffff);
    margin-bottom: 6px;
    box-shadow: 0 0 8px rgba(0, 255, 255, 1);
    border-radius: 1px;
  }

  .time-axis-labels {
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 0.8rem;
    color: var(--accent-color, #00ffff);
    font-weight: 700;
    text-shadow: 0 1px 3px rgba(0, 0, 0, 0.8);
  }

  .time-start,
  .time-end {
    background: rgba(26, 26, 26, 0.95);
    padding: 4px 8px;
    border-radius: 4px;
    border: 2px solid var(--accent-color, #00ffff);
    box-shadow: 0 2px 8px rgba(0, 255, 255, 0.6);
    backdrop-filter: blur(4px);
  }

  .time-duration {
    background: rgba(0, 255, 255, 0.2);
    padding: 4px 12px;
    border-radius: 6px;
    border: 2px solid var(--accent-color, #00ffff);
    font-weight: 800;
    text-transform: uppercase;
    letter-spacing: 0.8px;
    box-shadow: 0 0 12px rgba(0, 255, 255, 0.5);
    backdrop-filter: blur(4px);
  }

  .timeline-loading,
  .timeline-error {
    position: absolute;
    top: calc(50% + 30px);
    left: 50%;
    transform: translate(-50%, -50%);
    display: flex;
    align-items: center;
    gap: 1rem;
    font-size: 1.1rem;
    z-index: 5;
  }

  .timeline-error {
    color: var(--error-color, #ff6b6b);
  }

  .loading-spinner {
    width: 20px;
    height: 20px;
    border: 2px solid transparent;
    border-top: 2px solid var(--accent-color, #4ecdc4);
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    0% {
      transform: rotate(0deg);
    }
    100% {
      transform: rotate(360deg);
    }
  }

  /* Mobile Optimizations */
  @media (max-width: 768px) {
    .timeline-viewport-controls {
      flex-wrap: wrap;
      gap: 1rem;
      padding: 0.75rem;
    }

    .viewport-scrubber {
      order: -1;
      width: 100%;
      min-width: 0;
    }

    .playback-controls {
      gap: 0.25rem;
    }

    .speed-control {
      margin-left: 0.5rem;
    }

    .zoom-controls span {
      display: none;
    }
  }

  @media (max-width: 480px) {
    .timeline-viewport-controls {
      padding: 0.5rem;
    }

    .scale-selector label,
    .speed-control label {
      display: none;
    }

    .btn-primary,
    .btn-secondary {
      padding: 0.5rem;
      font-size: 1rem;
    }
  }

  /* CSS Custom Properties for Theming */
  :root {
    --bg-primary: #1a1a1a;
    --bg-secondary: #2a2a2a;
    --bg-tertiary: #3a3a3a;
    --text-primary: #ffffff;
    --text-secondary: #cccccc;
    --accent-color: #4ecdc4;
    --accent-hover: #45b7b8;
    --border-color: #333333;
    --error-color: #ff6b6b;
  }
</style>
