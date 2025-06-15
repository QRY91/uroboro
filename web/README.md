# Uroboro Journey Timeline - Web Interface

A modern, interactive timeline visualization for development journey replay with video-editor-style scrubbing controls. Built with Svelte + TypeScript + Vite for optimal performance and developer experience.

## 🚀 Features

- **Video-Editor Timeline**: Professional scrubbing controls with multiple time scales (15min, 1hr, 6hr, 24hr, 7d, full journey)
- **Interactive Viewport**: Pan, zoom, and scrub through development history
- **Real-time Event Rendering**: Smooth animations with performant event clustering
- **Multi-touch Support**: Touch gestures for mobile devices
- **Theme System**: Dark, light, matrix, and neon themes
- **PWA Ready**: Offline support and native app-like experience
- **Responsive Design**: Optimized for desktop, tablet, and mobile
- **Accessibility**: Full keyboard navigation and screen reader support

## 🛠️ Quick Start

### Prerequisites

- Node.js 18+ 
- npm or yarn
- Go backend running on port 8080 (for API)

### Development Setup

1. **Install dependencies**
   ```bash
   cd web
   npm install
   ```

2. **Start development server**
   ```bash
   npm run dev
   ```
   
   The app will be available at `http://localhost:3000`

3. **Ensure Go backend is running**
   ```bash
   # From project root
   go run . journey --port 8080
   ```

### Production Build

```bash
npm run build
npm run preview  # Test production build locally
```

## 📁 Project Structure

```
web/
├── src/
│   ├── components/          # Svelte components
│   │   ├── Timeline.svelte      # Main timeline component
│   │   ├── TimelineEvent.svelte # Individual event rendering
│   │   ├── TimelineRuler.svelte # Time scale ruler
│   │   └── ViewportScrubber.svelte # Scrubbing control
│   ├── stores/             # Svelte stores for state management
│   │   └── timeline.ts         # Main timeline state
│   ├── types/              # TypeScript type definitions
│   │   └── timeline.ts         # Comprehensive type system
│   ├── utils/              # Utility functions
│   │   └── timeline.ts         # Timeline calculations & helpers
│   ├── App.svelte          # Root application component
│   ├── main.ts             # Application entry point
│   └── app.css             # Global styles with CSS custom properties
├── public/                 # Static assets
│   ├── manifest.json           # PWA manifest
│   └── icons/                  # App icons
├── dist/                   # Built files (generated)
├── index.html              # HTML entry point
├── vite.config.ts          # Vite configuration
├── tsconfig.json           # TypeScript configuration
└── package.json            # Dependencies and scripts
```

## 🎨 Architecture Overview

### Component Hierarchy

```
App.svelte
└── Timeline.svelte
    ├── ViewportScrubber.svelte
    ├── TimelineRuler.svelte
    └── TimelineEvent.svelte (multiple instances)
```

### State Management

The application uses Svelte stores for reactive state management:

- **`timelineState`**: Central state store containing journey data, viewport, and UI state
- **`timelineActions`**: Action creators for state mutations
- **Derived stores**: Computed values like `eventsInCurrentViewport`, `filteredEvents`

### Key Features Implementation

#### Time Scale System
- **15 minute scale**: Shows detailed event content and descriptions
- **1 hour scale**: Event summaries with reduced detail
- **6+ hour scales**: Icon-only representation with clustering
- **Full journey**: Overview with milestone markers

#### Viewport Management
- Calculates visible events based on current time window
- Smooth transitions between scales
- Position preservation when changing scales
- Performance optimizations for large datasets

#### Event Positioning
- Smart clustering prevents overlap
- Project-based vertical distribution
- Smooth animations using anime.js
- Responsive sizing based on current scale

## 🔧 Development Workflow

### Available Scripts

```bash
npm run dev          # Start development server
npm run build        # Build for production
npm run preview      # Preview production build
npm run check        # Type checking
npm run check:watch  # Watch mode type checking
npm run lint         # ESLint linting
npm run format       # Prettier formatting
```

### Code Quality

- **TypeScript**: Strict type checking with comprehensive type definitions
- **ESLint**: Code linting with Svelte-specific rules
- **Prettier**: Consistent code formatting
- **Accessibility**: Built-in a11y compliance

### Performance Considerations

- **Event Clustering**: Automatically groups nearby events to prevent overlap
- **Viewport Culling**: Only renders events within current viewport
- **Throttled Rendering**: 60fps render loop with intelligent throttling
- **Device Detection**: Automatic performance level adjustment for low-spec devices
- **Lazy Loading**: Components and assets loaded on demand

## 🌐 API Integration

The timeline interfaces with the Go backend through REST endpoints:

### Primary Endpoint
```
GET /api/journey?days=7&projects=project1,project2
```

**Response Format:**
```typescript
interface JourneyData {
  events: JourneyEvent[];
  projects: ProjectInfo[];
  timeline: {
    startTime: string;
    endTime: string;
    totalDuration: number;
  };
  stats: JourneyStats;
}
```

### Health Check
```
GET /api/health
```

## 📱 Mobile & Touch Support

- **Touch Gestures**: Pan to scroll, pinch to zoom, tap to select
- **Responsive Layout**: Adaptive UI for different screen sizes
- **Performance Optimization**: Reduced detail levels on mobile devices
- **PWA Features**: Install as native app, offline support

## 🎨 Theming System

Themes are implemented using CSS custom properties:

```typescript
const themes = {
  dark: { /* Default dark theme */ },
  light: { /* Light theme for daylight use */ },
  matrix: { /* Green-on-black Matrix style */ },
  neon: { /* Cyberpunk neon aesthetics */ }
};
```

Theme switching is instant and persists in localStorage.

## 🚀 Deployment

### Production Build

1. **Build the application**
   ```bash
   npm run build
   ```

2. **Serve with Go backend**
   The Go server automatically serves built files from `web/dist/`

### PWA Deployment

The application includes:
- Service worker for offline caching
- Web app manifest for installation
- Icon sets for various devices
- Offline functionality for cached data

## 🔍 Debugging

### Development Tools

1. **Svelte DevTools**: Browser extension for component inspection
2. **TypeScript**: Compile-time error checking
3. **Source Maps**: Full debugging support in development
4. **Console Logging**: Structured logging with performance metrics

### Performance Monitoring

The app includes built-in performance monitoring:
- FPS tracking
- Render time measurement  
- Event count tracking
- Memory usage alerts

## 🧪 Testing Strategy

### Planned Testing Approach

- **Unit Tests**: Utility functions and state management
- **Component Tests**: Individual Svelte component behavior
- **Integration Tests**: API interaction and data flow
- **E2E Tests**: Full user workflows with Playwright
- **Visual Regression**: Screenshot comparison testing

## 🔮 Future Enhancements

### Planned Features

- **Real-time Updates**: WebSocket connection for live journey updates
- **Export Functionality**: Export timeline views as images/videos
- **Advanced Filtering**: Complex queries and saved filter sets
- **Collaboration**: Share timeline views with team members
- **Plugin System**: Extensible event type renderers
- **Analytics Dashboard**: Detailed productivity insights

### Performance Improvements

- **Virtual Scrolling**: Handle extremely large datasets
- **Web Workers**: Background processing for complex calculations
- **IndexedDB**: Local caching of journey data
- **Lazy Hydration**: Progressive enhancement for faster initial loads

## 📄 License

Part of the Uroboro project. See main project LICENSE for details.

## 🤝 Contributing

1. Follow the existing code style (Prettier + ESLint)
2. Add TypeScript types for all new functionality
3. Test on multiple devices and browsers
4. Consider accessibility in all UI changes
5. Update documentation for new features

## 📚 Additional Resources

- [Svelte Documentation](https://svelte.dev/docs)
- [TypeScript Handbook](https://www.typescriptlang.org/docs/)
- [Vite Guide](https://vitejs.dev/guide/)
- [Anime.js Documentation](https://animejs.com/documentation/)
- [Web APIs for Touch](https://developer.mozilla.org/en-US/docs/Web/API/Touch_events)

---

**Built with ❤️ using modern web technologies for the Uroboro development journey visualization system.**