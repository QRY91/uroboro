# Journey Replay Timeline Scrubbing Implementation

## 🎯 **Feature Overview**

Transform the Journey Replay visualization into a video editor style timeline with:
- **Time scale selection** (60min, 2hr, 6hr, 24hr, 7day windows)
- **Viewport panning** through development journey 
- **Zoom controls** for detailed vs overview perspectives
- **Scrubbing timeline** like video editing software
- **Rolling display** of events within time window

**Inspiration**: Solar system scale visualization (https://joshworth.com/dev/pixelspace/pixelspace_solarsystem.html)

## 🏗️ **Technical Architecture**

### **Core Components**

1. **TimelineViewport** - Manages visible time window
2. **TimeScaleController** - Handles zoom levels and scale selection  
3. **ViewportRenderer** - Renders events within current viewport
4. **ScrubController** - Handles timeline scrubbing interactions
5. **TimeNavigator** - Manages panning and navigation

### **Data Structures**

```javascript
class TimelineViewport {
    constructor() {
        this.startTime = null;
        this.endTime = null;
        this.scale = '24h';        // Current time scale
        this.position = 0;         // Position along full timeline (0-1)
        this.eventsInView = [];    // Events currently visible
    }
}

class TimeScale {
    constructor(name, duration, tickInterval, label) {
        this.name = name;           // '60m', '6h', '24h', '7d'
        this.duration = duration;   // Duration in milliseconds
        this.tickInterval = tickInterval; // Major tick interval
        this.label = label;         // Display label
    }
}
```

### **Time Scales Configuration**

```javascript
const TIME_SCALES = {
    '15m': new TimeScale('15m', 15 * 60 * 1000, 5 * 60 * 1000, '15 Minutes'),
    '1h': new TimeScale('1h', 60 * 60 * 1000, 10 * 60 * 1000, '1 Hour'),
    '6h': new TimeScale('6h', 6 * 60 * 60 * 1000, 60 * 60 * 1000, '6 Hours'),
    '24h': new TimeScale('24h', 24 * 60 * 60 * 1000, 4 * 60 * 60 * 1000, '24 Hours'),
    '7d': new TimeScale('7d', 7 * 24 * 60 * 60 * 1000, 24 * 60 * 60 * 1000, '7 Days'),
    'full': new TimeScale('full', null, null, 'Full Journey')
};
```

## 🎨 **UI Components**

### **Timeline Controls Panel**

Add to existing header controls:

```html
<div class="timeline-viewport-controls">
    <div class="scale-selector">
        <label>Time Scale:</label>
        <select id="timeScaleSelect">
            <option value="15m">15 Minutes</option>
            <option value="1h">1 Hour</option>
            <option value="6h">6 Hours</option>
            <option value="24h" selected>24 Hours</option>
            <option value="7d">7 Days</option>
            <option value="full">Full Journey</option>
        </select>
    </div>
    
    <div class="viewport-scrubber">
        <div class="scrubber-track">
            <div class="scrubber-viewport" id="viewportIndicator"></div>
            <input type="range" id="viewportSlider" min="0" max="100" value="0" class="viewport-slider">
        </div>
        <div class="time-labels">
            <span id="viewportStartTime">--:--</span>
            <span id="viewportEndTime">--:--</span>
        </div>
    </div>
    
    <div class="zoom-controls">
        <button id="zoomOut" class="btn-secondary">-</button>
        <span id="zoomLevel">100%</span>
        <button id="zoomIn" class="btn-secondary">+</button>
    </div>
</div>
```

### **Timeline Ruler Enhancement**

Replace basic axis with detailed ruler:

```html
<div class="timeline-ruler">
    <div class="major-ticks" id="majorTicks"></div>
    <div class="minor-ticks" id="minorTicks"></div>
    <div class="time-labels" id="timeLabels"></div>
    <div class="viewport-bounds">
        <div class="viewport-start-line"></div>
        <div class="viewport-end-line"></div>
    </div>
</div>
```

## 🔧 **Implementation Steps**

### **Phase 1: Viewport Infrastructure**

1. **Create TimelineViewport class**
   - Initialize with journey start/end times
   - Calculate viewport boundaries based on scale
   - Track current position along timeline

2. **Implement TimeScaleController**
   - Handle scale selection changes
   - Recalculate viewport when scale changes
   - Maintain position relative to full timeline

3. **Add viewport CSS and basic controls**
   - Style timeline scrubber
   - Add scale selector dropdown
   - Position controls in header

### **Phase 2: Event Filtering & Rendering**

1. **Viewport Event Filtering**
   ```javascript
   getEventsInViewport(viewport) {
       return this.journeyData.events.filter(event => {
           const eventTime = new Date(event.timestamp);
           return eventTime >= viewport.startTime && eventTime <= viewport.endTime;
       });
   }
   ```

2. **Enhanced Event Positioning**
   - Position events relative to viewport, not full journey
   - Increase density of events in smaller time scales
   - Add time-based clustering for overlapping events

3. **Dynamic Detail Levels**
   - 15min scale: Show individual capture content
   - 1hr scale: Show event summaries
   - 24hr scale: Show event types only
   - 7day scale: Show milestone markers

### **Phase 3: Scrubbing Interactions**

1. **Viewport Slider Implementation**
   ```javascript
   onViewportSliderChange(position) {
       const fullDuration = this.journeyEnd - this.journeyStart;
       const viewportDuration = this.currentScale.duration;
       const maxPosition = fullDuration - viewportDuration;
       
       this.viewport.position = position / 100;
       this.viewport.startTime = new Date(this.journeyStart.getTime() + (maxPosition * this.viewport.position));
       this.viewport.endTime = new Date(this.viewport.startTime.getTime() + viewportDuration);
       
       this.renderViewport();
   }
   ```

2. **Mouse/Touch Interactions**
   - Drag viewport indicator to pan
   - Mouse wheel to zoom in/out
   - Click timeline to jump to position
   - Touch gestures for mobile

3. **Keyboard Shortcuts**
   - Arrow keys: Pan left/right
   - +/- keys: Zoom in/out
   - Space: Play/pause within viewport
   - Home/End: Jump to journey start/end

### **Phase 4: Advanced Visualizations**

1. **Viewport-Aware Connections**
   - Only draw connections between events in viewport
   - Fade connections near viewport edges
   - Show "more events" indicators at boundaries

2. **Rolling Content Display**
   - Embed capture content in timeline at appropriate scales
   - Show git commit messages inline
   - Display project context switches visually

3. **Time-Density Visualization**
   - Color-code timeline based on activity density
   - Show "productivity heat map" in scrubber track
   - Highlight periods of intense development

## 🎨 **CSS Enhancements**

### **Viewport Controls Styling**

```css
.timeline-viewport-controls {
    display: flex;
    align-items: center;
    gap: 2rem;
    padding: 1rem;
    background: rgba(20, 20, 20, 0.9);
    border-bottom: 1px solid #333;
}

.viewport-scrubber {
    flex: 1;
    min-width: 300px;
}

.scrubber-track {
    position: relative;
    height: 20px;
    background: #333;
    border-radius: 10px;
    overflow: hidden;
}

.scrubber-viewport {
    position: absolute;
    height: 100%;
    background: rgba(78, 205, 196, 0.3);
    border: 2px solid #4ecdc4;
    border-radius: 8px;
    min-width: 20px;
}

.viewport-slider {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    opacity: 0;
    cursor: grab;
}

.timeline-ruler {
    position: absolute;
    bottom: 40px;
    left: 0;
    right: 0;
    height: 30px;
    pointer-events: none;
}

.major-ticks, .minor-ticks {
    position: absolute;
    width: 100%;
    height: 100%;
}

.tick {
    position: absolute;
    background: #4ecdc4;
    opacity: 0.6;
}

.major-tick {
    height: 20px;
    width: 2px;
    bottom: 0;
}

.minor-tick {
    height: 10px;
    width: 1px;
    bottom: 0;
    opacity: 0.3;
}

.time-label {
    position: absolute;
    font-size: 0.75rem;
    color: #4ecdc4;
    bottom: -18px;
    transform: translateX(-50%);
}

.viewport-bounds {
    position: absolute;
    top: 0;
    bottom: 0;
    pointer-events: none;
}

.viewport-start-line, .viewport-end-line {
    position: absolute;
    width: 2px;
    height: 100%;
    background: #ff6b6b;
    opacity: 0.8;
}
```

## 🎮 **Enhanced Event Rendering**

### **Scale-Aware Event Display**

```javascript
class ScaleAwareEventRenderer {
    renderEvent(event, scale, viewportWidth) {
        const baseSize = this.getEventSize(scale);
        const eventEl = this.createEventElement(event, baseSize);
        
        // Add scale-specific content
        switch(scale.name) {
            case '15m':
                this.addDetailedContent(eventEl, event);
                break;
            case '1h':
                this.addSummaryContent(eventEl, event);
                break;
            case '24h':
                this.addIconOnly(eventEl, event);
                break;
            case '7d':
                this.addMilestoneOnly(eventEl, event);
                break;
        }
        
        return eventEl;
    }
    
    addDetailedContent(element, event) {
        // Show full capture content for very zoomed in view
        const contentEl = document.createElement('div');
        contentEl.className = 'event-detailed-content';
        contentEl.textContent = event.content.substring(0, 100);
        element.appendChild(contentEl);
    }
}
```

## 📱 **Mobile Optimizations**

### **Touch Interactions**

```javascript
class TouchViewportController {
    constructor(viewport) {
        this.viewport = viewport;
        this.gestureStart = null;
        this.initialViewport = null;
    }
    
    onTouchStart(e) {
        if (e.touches.length === 1) {
            // Single finger - pan
            this.gestureStart = { x: e.touches[0].clientX, type: 'pan' };
        } else if (e.touches.length === 2) {
            // Two fingers - zoom
            const distance = this.getTouchDistance(e.touches);
            this.gestureStart = { distance, type: 'zoom' };
        }
        this.initialViewport = { ...this.viewport };
    }
    
    onTouchMove(e) {
        if (this.gestureStart.type === 'pan') {
            this.handlePan(e);
        } else if (this.gestureStart.type === 'zoom') {
            this.handleZoom(e);
        }
    }
}
```

## 🚀 **Performance Optimizations**

### **Viewport Culling**

```javascript
class PerformantViewportRenderer {
    constructor() {
        this.eventPool = []; // Reuse DOM elements
        this.visibleEvents = new Set();
        this.renderThrottled = this.throttle(this.render.bind(this), 16); // 60fps
    }
    
    updateViewport(viewport) {
        // Only re-render if viewport actually changed
        if (this.viewportEquals(viewport, this.lastViewport)) return;
        
        this.renderThrottled(viewport);
        this.lastViewport = { ...viewport };
    }
    
    cullEvents(allEvents, viewport) {
        // Only process events that could be visible
        return allEvents.filter(event => {
            const eventTime = new Date(event.timestamp);
            const buffer = (viewport.endTime - viewport.startTime) * 0.1; // 10% buffer
            return eventTime >= (viewport.startTime - buffer) && 
                   eventTime <= (viewport.endTime + buffer);
        });
    }
}
```

## 🧪 **Testing Strategy**

### **Viewport Logic Tests**

```javascript
describe('TimelineViewport', () => {
    test('should correctly calculate events in 1-hour viewport', () => {
        const viewport = new TimelineViewport();
        viewport.setScale('1h');
        viewport.setPosition(0.5); // Middle of journey
        
        const eventsInView = viewport.getEventsInView();
        expect(eventsInView.length).toBeGreaterThan(0);
        
        eventsInView.forEach(event => {
            const eventTime = new Date(event.timestamp);
            expect(eventTime).toBeGreaterThanOrEqual(viewport.startTime);
            expect(eventTime).toBeLessThanOrEqual(viewport.endTime);
        });
    });
    
    test('should maintain relative position when changing scales', () => {
        const viewport = new TimelineViewport();
        viewport.setScale('24h');
        viewport.setPosition(0.3);
        
        const originalMiddleTime = viewport.getMiddleTime();
        
        viewport.setScale('1h');
        const newMiddleTime = viewport.getMiddleTime();
        
        expect(Math.abs(originalMiddleTime - newMiddleTime)).toBeLessThan(1000); // Within 1 second
    });
});
```

## 📝 **Implementation Priority**

### **Phase 1 (Core Infrastructure) - Week 1**
- [ ] TimelineViewport class implementation
- [ ] Basic scale selector UI
- [ ] Event filtering by viewport
- [ ] Simple viewport slider

### **Phase 2 (Enhanced UI) - Week 2**  
- [ ] Timeline ruler with ticks
- [ ] Improved scrubbing interactions
- [ ] Zoom controls
- [ ] Keyboard shortcuts

### **Phase 3 (Advanced Features) - Week 3**
- [ ] Scale-aware event rendering
- [ ] Touch/mobile optimizations
- [ ] Performance optimizations
- [ ] Rolling content display

### **Phase 4 (Polish) - Week 4**
- [ ] Smooth animations and transitions
- [ ] Advanced viewport visualizations
- [ ] Testing and bug fixes
- [ ] Documentation and examples

## 🎬 **User Experience Flow**

1. **Default View**: Opens with 24-hour viewport showing current activity
2. **Scale Selection**: User selects different time scale from dropdown
3. **Viewport Panning**: User drags slider or uses mouse wheel to pan through journey
4. **Zoom Interaction**: User zooms in/out to see more/less detail
5. **Play Controls**: Timeline plays within current viewport window
6. **Jump Navigation**: Click events in console to jump viewport to that time

This implementation will transform the Journey Replay into a powerful, video-editor-style timeline that allows deep exploration of development patterns at any time scale!