<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import { viewport, controls, journeyData, timelineActions } from '@stores/timeline';
  import { TIME_SCALES } from '@types/timeline';
  import type { TimeScaleName } from '@types/timeline';

  const dispatch = createEventDispatcher();

  // Component state
  let scrubberTrack: HTMLDivElement;
  let viewportIndicator: HTMLDivElement;
  let isDragging = false;
  let dragStartX = 0;
  let dragStartPosition = 0;
  let trackWidth = 0;
  let indicatorWidth = 0;

  // Reactive statements
  $: currentScale = $controls.currentScale;
  $: viewportPosition = $controls.viewportPosition;
  $: hasJourneyData = !!$journeyData;

  // Calculate viewport indicator size and position
  $: {
    if (hasJourneyData && scrubberTrack) {
      updateIndicatorSize();
      updateIndicatorPosition();
    }
  }

  onMount(() => {
    setupEventListeners();
    updateTrackWidth();
  });

  function setupEventListeners() {
    if (!scrubberTrack) return;

    // Mouse events
    scrubberTrack.addEventListener('mousedown', handleMouseDown);
    scrubberTrack.addEventListener('click', handleTrackClick);

    // Touch events
    scrubberTrack.addEventListener('touchstart', handleTouchStart, { passive: false });

    // Global events for dragging
    window.addEventListener('mousemove', handleGlobalMouseMove);
    window.addEventListener('mouseup', handleGlobalMouseUp);
    window.addEventListener('touchmove', handleGlobalTouchMove, { passive: false });
    window.addEventListener('touchend', handleGlobalTouchEnd);

    // Resize handling
    window.addEventListener('resize', updateTrackWidth);
  }

  function updateTrackWidth() {
    if (scrubberTrack) {
      trackWidth = scrubberTrack.clientWidth;
    }
  }

  function updateIndicatorSize() {
    if (!$journeyData || currentScale === 'full') {
      indicatorWidth = trackWidth;
      return;
    }

    const timeScale = TIME_SCALES[currentScale];
    const fullDuration =
      new Date($journeyData.timeline.endTime).getTime() - new Date($journeyData.timeline.startTime).getTime();
    const viewportDuration = timeScale.duration!;
    const ratio = Math.min(1, viewportDuration / fullDuration);

    indicatorWidth = Math.max(20, trackWidth * ratio); // Minimum 20px width
  }

  function updateIndicatorPosition() {
    if (!viewportIndicator) return;

    const maxPosition = trackWidth - indicatorWidth;
    const position = viewportPosition * maxPosition;

    viewportIndicator.style.left = `${position}px`;
    viewportIndicator.style.width = `${indicatorWidth}px`;
  }

  function handleMouseDown(e: MouseEvent) {
    if (e.target === viewportIndicator) {
      // Dragging the viewport indicator
      startDrag(e.clientX);
      e.preventDefault();
    }
  }

  function handleTouchStart(e: TouchEvent) {
    if (e.target === viewportIndicator && e.touches.length === 1) {
      // Dragging the viewport indicator
      startDrag(e.touches[0].clientX);
      e.preventDefault();
    }
  }

  function startDrag(clientX: number) {
    isDragging = true;
    dragStartX = clientX;
    dragStartPosition = viewportPosition;

    // Visual feedback
    if (viewportIndicator) {
      viewportIndicator.classList.add('dragging');
    }

    dispatch('dragStart');
  }

  function handleGlobalMouseMove(e: MouseEvent) {
    if (isDragging) {
      updateDragPosition(e.clientX);
    }
  }

  function handleGlobalTouchMove(e: TouchEvent) {
    if (isDragging && e.touches.length === 1) {
      updateDragPosition(e.touches[0].clientX);
      e.preventDefault();
    }
  }

  function updateDragPosition(clientX: number) {
    const deltaX = clientX - dragStartX;
    const maxPosition = trackWidth - indicatorWidth;
    const deltaPosition = deltaX / maxPosition;

    const newPosition = Math.max(0, Math.min(1, dragStartPosition + deltaPosition));
    timelineActions.setViewportPosition(newPosition);

    dispatch('drag', { position: newPosition });
  }

  function handleGlobalMouseUp() {
    endDrag();
  }

  function handleGlobalTouchEnd() {
    endDrag();
  }

  function endDrag() {
    if (!isDragging) return;

    isDragging = false;

    if (viewportIndicator) {
      viewportIndicator.classList.remove('dragging');
    }

    dispatch('dragEnd');
  }

  function handleTrackClick(e: MouseEvent) {
    if (e.target === viewportIndicator || isDragging) return;

    const rect = scrubberTrack.getBoundingClientRect();
    const clickX = e.clientX - rect.left;
    const clickPosition = clickX / trackWidth;

    // Adjust for indicator width to center it on click
    const maxPosition = trackWidth - indicatorWidth;
    const targetIndicatorLeft = clickX - indicatorWidth / 2;
    const targetPosition = Math.max(0, Math.min(1, targetIndicatorLeft / maxPosition));

    timelineActions.setViewportPosition(targetPosition);
    dispatch('jump', { position: targetPosition });
  }

  // Activity density visualization
  function getActivityDensity(): number[] {
    if (!$journeyData) return [];

    const segments = 100; // Divide timeline into 100 segments
    const density = new Array(segments).fill(0);

    const startTime = new Date($journeyData.timeline.startTime).getTime();
    const endTime = new Date($journeyData.timeline.endTime).getTime();
    const duration = endTime - startTime;
    const segmentDuration = duration / segments;

    $journeyData.events.forEach(event => {
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

  $: activityDensity = getActivityDensity();

  // Format time for tooltips
  function formatTimeForTooltip(position: number): string {
    if (!$journeyData) return '';

    const startTime = new Date($journeyData.timeline.startTime).getTime();
    const endTime = new Date($journeyData.timeline.endTime).getTime();
    const duration = endTime - startTime;
    const targetTime = new Date(startTime + duration * position);

    return targetTime.toLocaleString();
  }
</script>

<div class="viewport-scrubber">
  <div class="scrubber-track" bind:this={scrubberTrack}>
    <!-- Activity density background -->
    <div class="activity-density">
      {#each activityDensity as density, index}
        <div
          class="density-segment"
          style="height: {Math.max(2, density * 100)}%; left: {(index / activityDensity.length) * 100}%">
        </div>
      {/each}
    </div>

    <!-- Viewport indicator -->
    <div
      class="scrubber-viewport"
      bind:this={viewportIndicator}
      title="Drag to scrub timeline - {formatTimeForTooltip(viewportPosition)}">
      <div class="viewport-handle left"></div>
      <div class="viewport-content">
        <span class="viewport-label">{TIME_SCALES[currentScale].label}</span>
      </div>
      <div class="viewport-handle right"></div>
    </div>

    <!-- Scale markers -->
    <div class="scale-markers">
      {#if $journeyData}
        {#each Array(10) as _, i}
          <div class="scale-marker" style="left: {(i / 9) * 100}%" title={formatTimeForTooltip(i / 9)}></div>
        {/each}
      {/if}
    </div>
  </div>
</div>

<style>
  .viewport-scrubber {
    position: relative;
    width: 100%;
    height: 32px;
  }

  .scrubber-track {
    position: relative;
    width: 100%;
    height: 20px;
    background: linear-gradient(135deg, #444444 0%, #333333 100%);
    border-radius: 10px;
    overflow: hidden;
    cursor: pointer;
    box-shadow: inset 0 2px 4px rgba(0, 0, 0, 0.5);
    border: 1px solid #666666;
  }

  .activity-density {
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
    height: 100%;
    display: flex;
    align-items: flex-end;
    pointer-events: none;
  }

  .density-segment {
    position: absolute;
    bottom: 0;
    width: 1%;
    background: linear-gradient(to top, rgba(0, 255, 255, 0.8) 0%, rgba(0, 255, 255, 0.4) 100%);
    transition: all 0.2s ease;
  }

  .scrubber-viewport {
    position: absolute;
    top: 0;
    height: 100%;
    background: linear-gradient(135deg, rgba(0, 255, 255, 0.9) 0%, rgba(0, 230, 230, 1) 100%);
    border: 2px solid #00ffff;
    border-radius: 8px;
    min-width: 20px;
    cursor: grab;
    transition: all 0.2s ease;
    display: flex;
    align-items: center;
    box-shadow:
      0 2px 8px rgba(0, 255, 255, 0.6),
      inset 0 1px 2px rgba(255, 255, 255, 0.3);
  }

  .scrubber-viewport:hover {
    transform: translateY(-1px);
    box-shadow:
      0 4px 12px rgba(0, 255, 255, 0.8),
      inset 0 1px 2px rgba(255, 255, 255, 0.4);
  }

  .scrubber-viewport.dragging {
    cursor: grabbing;
    transform: translateY(-2px);
    box-shadow:
      0 6px 16px rgba(0, 255, 255, 0.9),
      inset 0 1px 2px rgba(255, 255, 255, 0.5);
    border-color: #00e6e6;
  }

  .viewport-handle {
    width: 4px;
    height: 60%;
    background: rgba(255, 255, 255, 0.9);
    border-radius: 2px;
    margin: 0 2px;
  }

  .viewport-content {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    min-width: 0;
  }

  .viewport-label {
    font-size: 0.7rem;
    font-weight: 600;
    color: #1a1a1a;
    text-shadow: 0 1px 2px rgba(255, 255, 255, 0.3);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .scale-markers {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 100%;
    pointer-events: none;
  }

  .scale-marker {
    position: absolute;
    top: 0;
    width: 1px;
    height: 100%;
    background: rgba(0, 255, 255, 0.4);
  }

  .scale-marker:nth-child(1),
  .scale-marker:nth-child(10) {
    background: rgba(0, 255, 255, 0.8);
    width: 2px;
  }

  .scale-marker:nth-child(5) {
    background: rgba(0, 255, 255, 0.6);
    width: 1.5px;
  }

  /* Responsive design */
  @media (max-width: 768px) {
    .viewport-scrubber {
      height: 40px;
    }

    .scrubber-track {
      height: 28px;
      border-radius: 14px;
    }

    .scrubber-viewport {
      min-width: 28px;
    }

    .viewport-label {
      font-size: 0.6rem;
    }

    .viewport-handle {
      width: 3px;
      height: 50%;
    }
  }

  @media (max-width: 480px) {
    .viewport-label {
      display: none;
    }

    .scrubber-viewport {
      min-width: 24px;
    }
  }

  /* Accessibility */
  @media (prefers-reduced-motion: reduce) {
    .scrubber-viewport,
    .density-segment {
      transition: none;
    }
  }

  /* High contrast mode */
  @media (prefers-contrast: high) {
    .scrubber-track {
      background: #000000;
      border: 1px solid #ffffff;
    }

    .scrubber-viewport {
      background: #ffffff;
      border-color: #000000;
    }

    .viewport-label {
      color: #000000;
    }
  }
</style>
