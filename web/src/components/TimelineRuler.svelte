<script lang="ts">
  import { onMount } from 'svelte';
  import { viewport, journeyData } from '@stores/timeline';
  import { TIME_SCALES } from '@types/timeline';
  import type { TickMark, TimeLabel } from '@types/timeline';

  // Component state
  let rulerContainer: HTMLDivElement;
  let containerWidth = 0;

  // Reactive tick generation
  $: ticks = generateTicks($viewport);
  $: labels = generateLabels($viewport);

  onMount(() => {
    updateContainerWidth();
    window.addEventListener('resize', updateContainerWidth);

    return () => {
      window.removeEventListener('resize', updateContainerWidth);
    };
  });

  function updateContainerWidth() {
    if (rulerContainer) {
      containerWidth = rulerContainer.clientWidth;
    }
  }

  function generateTicks(currentViewport: typeof $viewport): { major: TickMark[]; minor: TickMark[] } {
    if (!currentViewport.startTime || !currentViewport.endTime || !$journeyData) {
      return { major: [], minor: [] };
    }

    const startTime = currentViewport.startTime.getTime();
    const endTime = currentViewport.endTime.getTime();
    const duration = endTime - startTime;
    const scale = TIME_SCALES[currentViewport.scale];

    const majorTicks: TickMark[] = [];
    const minorTicks: TickMark[] = [];

    // Determine tick intervals based on scale
    let majorInterval: number;
    let minorInterval: number;

    switch (currentViewport.scale) {
      case '15m':
        majorInterval = 5 * 60 * 1000; // 5 minutes
        minorInterval = 1 * 60 * 1000; // 1 minute
        break;
      case '1h':
        majorInterval = 15 * 60 * 1000; // 15 minutes
        minorInterval = 5 * 60 * 1000; // 5 minutes
        break;
      case '6h':
        majorInterval = 60 * 60 * 1000; // 1 hour
        minorInterval = 15 * 60 * 1000; // 15 minutes
        break;
      case '24h':
        majorInterval = 4 * 60 * 60 * 1000; // 4 hours
        minorInterval = 60 * 60 * 1000; // 1 hour
        break;
      case '7d':
        majorInterval = 24 * 60 * 60 * 1000; // 1 day
        minorInterval = 6 * 60 * 60 * 1000; // 6 hours
        break;
      case 'full':
        // For full timeline, calculate appropriate intervals
        const days = duration / (24 * 60 * 60 * 1000);
        if (days > 365) {
          majorInterval = 30 * 24 * 60 * 60 * 1000; // 1 month
          minorInterval = 7 * 24 * 60 * 60 * 1000; // 1 week
        } else if (days > 30) {
          majorInterval = 7 * 24 * 60 * 60 * 1000; // 1 week
          minorInterval = 24 * 60 * 60 * 1000; // 1 day
        } else {
          majorInterval = 24 * 60 * 60 * 1000; // 1 day
          minorInterval = 6 * 60 * 60 * 1000; // 6 hours
        }
        break;
      default:
        majorInterval = scale.tickInterval || 60 * 60 * 1000;
        minorInterval = majorInterval / 4;
    }

    // Find the first major tick aligned to interval
    const firstMajorTick = Math.ceil(startTime / majorInterval) * majorInterval;
    const firstMinorTick = Math.ceil(startTime / minorInterval) * minorInterval;

    // Generate major ticks
    for (let time = firstMajorTick; time <= endTime; time += majorInterval) {
      if (time >= startTime) {
        const position = (time - startTime) / duration;
        majorTicks.push({
          position,
          timestamp: new Date(time),
          type: 'major',
        });
      }
    }

    // Generate minor ticks (excluding positions where major ticks exist)
    for (let time = firstMinorTick; time <= endTime; time += minorInterval) {
      if (time >= startTime) {
        const position = (time - startTime) / duration;
        const isMajorTick = majorTicks.some(tick => Math.abs(tick.position - position) < 0.001);

        if (!isMajorTick) {
          minorTicks.push({
            position,
            timestamp: new Date(time),
            type: 'minor',
          });
        }
      }
    }

    return { major: majorTicks, minor: minorTicks };
  }

  function generateLabels(currentViewport: typeof $viewport): TimeLabel[] {
    if (!currentViewport.startTime || !currentViewport.endTime) {
      return [];
    }

    const majorTicks = ticks.major;
    const labels: TimeLabel[] = [];

    // Only show labels for major ticks, but space them out to avoid overlap
    const minLabelSpacing = 100; // Minimum pixels between labels for better readability
    const labelSpacing = minLabelSpacing / containerWidth; // As fraction of timeline

    let lastLabelPosition = -1;

    majorTicks.forEach(tick => {
      if (tick.position - lastLabelPosition >= labelSpacing) {
        labels.push({
          position: tick.position,
          text: formatTickLabel(tick.timestamp, currentViewport.scale),
          timestamp: tick.timestamp,
        });
        lastLabelPosition = tick.position;
      }
    });

    return labels;
  }

  function formatTickLabel(date: Date, scale: typeof $viewport.scale): string {
    const now = new Date();
    const isToday = date.toDateString() === now.toDateString();
    const isThisYear = date.getFullYear() === now.getFullYear();

    switch (scale) {
      case '15m':
      case '1h':
        return date.toLocaleTimeString([], {
          hour: '2-digit',
          minute: '2-digit',
          hour12: false,
        });

      case '6h':
        return date.toLocaleTimeString([], {
          hour: '2-digit',
          minute: '2-digit',
          hour12: false,
        });

      case '24h':
        if (isToday) {
          return (
            'Today\n' +
            date.toLocaleTimeString([], {
              hour: '2-digit',
              minute: '2-digit',
              hour12: false,
            })
          );
        } else {
          return (
            date.toLocaleDateString([], {
              month: 'short',
              day: 'numeric',
            }) +
            '\n' +
            date.toLocaleTimeString([], {
              hour: '2-digit',
              minute: '2-digit',
              hour12: false,
            })
          );
        }

      case '7d':
        const daysDiff = Math.floor((now.getTime() - date.getTime()) / (1000 * 60 * 60 * 24));
        if (daysDiff === 0) {
          return 'Today\n' + date.toLocaleDateString([], { month: 'short', day: 'numeric' });
        } else if (daysDiff === 1) {
          return 'Yesterday\n' + date.toLocaleDateString([], { month: 'short', day: 'numeric' });
        } else {
          return (
            date.toLocaleDateString([], {
              weekday: 'short',
            }) +
            '\n' +
            date.toLocaleDateString([], {
              month: 'short',
              day: 'numeric',
            })
          );
        }

      case 'full':
        const monthsDiff = (now.getFullYear() - date.getFullYear()) * 12 + (now.getMonth() - date.getMonth());
        if (monthsDiff === 0) {
          return 'This Month\n' + date.toLocaleDateString([], { month: 'short', day: 'numeric' });
        } else if (isThisYear) {
          return (
            date.toLocaleDateString([], {
              month: 'short',
            }) +
            '\n' +
            date.toLocaleDateString([], {
              day: 'numeric',
            })
          );
        } else {
          return date.toLocaleDateString([], {
            year: '2-digit',
            month: 'short',
          });
        }

      default:
        return date.toLocaleString();
    }
  }

  function getTickHeight(type: 'major' | 'minor'): number {
    return type === 'major' ? 20 : 10;
  }

  function getTickOpacity(type: 'major' | 'minor'): number {
    return type === 'major' ? 0.8 : 0.4;
  }
</script>

<div class="timeline-ruler" bind:this={rulerContainer}>
  <!-- Minor ticks -->
  <div class="tick-layer minor-ticks">
    {#each ticks.minor as tick}
      <div
        class="tick minor-tick"
        style="left: {tick.position * 100}%; height: {getTickHeight('minor')}px; opacity: {getTickOpacity('minor')}"
        title={tick.timestamp.toLocaleString()}>
      </div>
    {/each}
  </div>

  <!-- Major ticks -->
  <div class="tick-layer major-ticks">
    {#each ticks.major as tick}
      <div
        class="tick major-tick"
        style="left: {tick.position * 100}%; height: {getTickHeight('major')}px; opacity: {getTickOpacity('major')}"
        title={tick.timestamp.toLocaleString()}>
      </div>
    {/each}
  </div>

  <!-- Time labels -->
  <div class="label-layer">
    {#each labels as label}
      <div class="time-label" style="left: {label.position * 100}%" title={label.timestamp.toLocaleString()}>
        {label.text}
      </div>
    {/each}
  </div>

  <!-- Viewport bounds indicators -->
  {#if $viewport.scale !== 'full'}
    <div class="viewport-bounds">
      <div class="viewport-start-line"></div>
      <div class="viewport-end-line"></div>
    </div>
  {/if}

  <!-- Current time indicator (if within viewport) -->
  {#if $viewport.startTime && $viewport.endTime}
    {@const now = new Date()}
    {@const nowTime = now.getTime()}
    {@const startTime = $viewport.startTime.getTime()}
    {@const endTime = $viewport.endTime.getTime()}
    {#if nowTime >= startTime && nowTime <= endTime}
      {@const nowPosition = (nowTime - startTime) / (endTime - startTime)}
      <div class="current-time-indicator" style="left: {nowPosition * 100}%">
        <div class="current-time-line"></div>
        <div class="current-time-label">NOW</div>
      </div>
    {/if}
  {/if}
</div>

<style>
  .timeline-ruler {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 60px;
    pointer-events: none;
    background: linear-gradient(
      to bottom,
      rgba(26, 26, 26, 0.95) 0%,
      rgba(26, 26, 26, 0.8) 50%,
      rgba(26, 26, 26, 0.4) 100%
    );
    border-bottom: 2px solid var(--accent-color, #00ffff);
    z-index: 5;
  }

  .tick-layer {
    position: absolute;
    width: 100%;
    height: 100%;
    top: 0;
    left: 0;
  }

  .tick {
    position: absolute;
    top: 35px;
    width: 1px;
    background: var(--accent-color, #00ffff);
    transform: translateX(-50%);
    opacity: 0.8;
  }

  .major-tick {
    width: 2px;
    height: 25px;
    background: var(--accent-color, #00ffff);
    box-shadow: 0 0 4px rgba(0, 255, 255, 0.8);
    opacity: 1;
  }

  .minor-tick {
    width: 1px;
    height: 15px;
    background: var(--accent-color, #00ffff);
    opacity: 0.6;
  }

  .label-layer {
    position: absolute;
    width: 100%;
    height: 100%;
    top: 0;
    left: 0;
  }

  .time-label {
    position: absolute;
    top: 5px;
    font-size: 0.75rem;
    color: var(--accent-color, #00ffff);
    font-weight: 600;
    transform: translateX(-50%);
    white-space: pre-line;
    text-align: center;
    line-height: 1.3;
    text-shadow: 0 1px 3px rgba(0, 0, 0, 0.9);
    background: rgba(26, 26, 26, 0.95);
    padding: 3px 6px;
    border-radius: 4px;
    border: 1px solid var(--accent-color, #00ffff);
    backdrop-filter: blur(6px);
    box-shadow: 0 2px 6px rgba(0, 255, 255, 0.3);
    min-width: 60px;
  }

  .viewport-bounds {
    position: absolute;
    top: 0;
    bottom: 0;
    width: 100%;
    pointer-events: none;
  }

  .viewport-start-line,
  .viewport-end-line {
    position: absolute;
    width: 2px;
    height: 100%;
    background: var(--error-color, #ff6b6b);
    opacity: 0.6;
  }

  .viewport-start-line {
    left: 0;
  }

  .viewport-end-line {
    right: 0;
  }

  .current-time-indicator {
    position: absolute;
    top: 0;
    bottom: 0;
    transform: translateX(-50%);
    z-index: 10;
  }

  .current-time-line {
    width: 2px;
    height: 100%;
    background: #f9ca24;
    box-shadow: 0 0 4px rgba(249, 202, 36, 0.8);
    animation: pulse 2s ease-in-out infinite;
  }

  .current-time-label {
    position: absolute;
    top: -22px;
    left: 50%;
    transform: translateX(-50%);
    font-size: 0.6rem;
    font-weight: 700;
    color: #f9ca24;
    background: rgba(26, 26, 26, 0.9);
    padding: 2px 6px;
    border-radius: 3px;
    border: 1px solid #f9ca24;
    text-shadow: none;
  }

  @keyframes pulse {
    0%,
    100% {
      opacity: 1;
    }
    50% {
      opacity: 0.6;
    }
  }

  /* Responsive adjustments */
  @media (max-width: 768px) {
    .timeline-ruler {
      height: 55px;
    }

    .time-label {
      font-size: 0.65rem;
      padding: 2px 4px;
      min-width: 50px;
    }

    .current-time-label {
      font-size: 0.55rem;
      top: -20px;
    }
  }

  @media (max-width: 480px) {
    .timeline-ruler {
      height: 50px;
    }

    .time-label {
      font-size: 0.6rem;
      padding: 2px 3px;
      min-width: 45px;
    }

    .tick {
      top: 30px;
    }

    .current-time-label {
      font-size: 0.5rem;
      top: -18px;
      padding: 1px 4px;
    }
  }

  /* High contrast mode */
  @media (prefers-contrast: high) {
    .tick {
      background: #ffffff;
    }

    .time-label {
      color: #ffffff;
      background: #000000;
      border: 1px solid #ffffff;
    }
  }

  /* Reduced motion */
  @media (prefers-reduced-motion: reduce) {
    .current-time-line {
      animation: none;
    }
  }

  /* Dark theme specific adjustments */
  @media (prefers-color-scheme: dark) {
    .timeline-ruler {
      background: linear-gradient(to top, rgba(13, 13, 13, 0.95) 0%, rgba(13, 13, 13, 0.4) 70%, transparent 100%);
    }

    .time-label {
      background: rgba(13, 13, 13, 0.9);
    }
  }
</style>
