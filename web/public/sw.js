// Uroboro Timeline Service Worker
// Provides offline support and caching for the timeline web application

const CACHE_NAME = 'uroboro-timeline-v1.0.0';
const STATIC_CACHE_NAME = 'uroboro-static-v1.0.0';
const API_CACHE_NAME = 'uroboro-api-v1.0.0';

// Assets to cache immediately
const STATIC_ASSETS = [
  '/',
  '/index.html',
  '/manifest.json',
  '/src/main.ts',
  '/src/app.css',
  '/src/App.svelte'
];

// API endpoints to cache
const API_ENDPOINTS = [
  '/api/health',
  '/api/journey'
];

// Install event - cache static assets
self.addEventListener('install', event => {
  console.log('Service Worker: Installing...');

  event.waitUntil(
    Promise.all([
      // Cache static assets
      caches.open(STATIC_CACHE_NAME).then(cache => {
        console.log('Service Worker: Caching static assets');
        return cache.addAll(STATIC_ASSETS);
      }),

      // Cache API health check
      caches.open(API_CACHE_NAME).then(cache => {
        console.log('Service Worker: Caching API endpoints');
        return cache.add('/api/health');
      })
    ]).then(() => {
      console.log('Service Worker: Installation complete');
      // Force activation of new service worker
      return self.skipWaiting();
    }).catch(error => {
      console.error('Service Worker: Installation failed', error);
    })
  );
});

// Activate event - clean up old caches
self.addEventListener('activate', event => {
  console.log('Service Worker: Activating...');

  event.waitUntil(
    caches.keys().then(cacheNames => {
      return Promise.all(
        cacheNames.map(cacheName => {
          // Delete old caches
          if (cacheName !== STATIC_CACHE_NAME &&
              cacheName !== API_CACHE_NAME &&
              cacheName !== CACHE_NAME) {
            console.log('Service Worker: Deleting old cache', cacheName);
            return caches.delete(cacheName);
          }
        })
      );
    }).then(() => {
      console.log('Service Worker: Activation complete');
      // Take control of all clients
      return self.clients.claim();
    })
  );
});

// Fetch event - handle network requests
self.addEventListener('fetch', event => {
  const { request } = event;
  const url = new URL(request.url);

  // Skip non-GET requests
  if (request.method !== 'GET') {
    return;
  }

  // Handle different types of requests
  if (url.pathname.startsWith('/api/')) {
    // API requests - network first, cache fallback
    event.respondWith(handleApiRequest(request));
  } else if (url.pathname.startsWith('/src/') ||
             url.pathname.startsWith('/assets/') ||
             url.pathname.endsWith('.js') ||
             url.pathname.endsWith('.css') ||
             url.pathname.endsWith('.json')) {
    // Static assets - cache first, network fallback
    event.respondWith(handleStaticAsset(request));
  } else {
    // HTML pages - network first, cache fallback, offline page
    event.respondWith(handlePageRequest(request));
  }
});

// Handle API requests with network-first strategy
async function handleApiRequest(request) {
  try {
    // Try network first
    const networkResponse = await fetch(request);

    if (networkResponse.ok) {
      // Cache successful responses
      const cache = await caches.open(API_CACHE_NAME);
      cache.put(request, networkResponse.clone());
      return networkResponse;
    }

    throw new Error('Network response not ok');
  } catch (error) {
    console.log('Service Worker: Network failed, trying cache', error);

    // Fallback to cache
    const cachedResponse = await caches.match(request);
    if (cachedResponse) {
      return cachedResponse;
    }

    // Return offline response for API
    return new Response(JSON.stringify({
      error: 'Offline',
      message: 'No network connection and no cached data available',
      offline: true
    }), {
      status: 503,
      statusText: 'Service Unavailable',
      headers: {
        'Content-Type': 'application/json'
      }
    });
  }
}

// Handle static assets with cache-first strategy
async function handleStaticAsset(request) {
  try {
    // Try cache first
    const cachedResponse = await caches.match(request);
    if (cachedResponse) {
      return cachedResponse;
    }

    // Fallback to network
    const networkResponse = await fetch(request);
    if (networkResponse.ok) {
      // Cache the response
      const cache = await caches.open(STATIC_CACHE_NAME);
      cache.put(request, networkResponse.clone());
    }

    return networkResponse;
  } catch (error) {
    console.error('Service Worker: Failed to fetch static asset', error);

    // For development assets that might not be cacheable
    if (request.url.includes('localhost:3000')) {
      return fetch(request);
    }

    throw error;
  }
}

// Handle page requests with network-first strategy
async function handlePageRequest(request) {
  try {
    // Try network first
    const networkResponse = await fetch(request);

    if (networkResponse.ok) {
      // Cache successful page responses
      const cache = await caches.open(STATIC_CACHE_NAME);
      cache.put(request, networkResponse.clone());
      return networkResponse;
    }

    throw new Error('Network response not ok');
  } catch (error) {
    console.log('Service Worker: Network failed for page, trying cache');

    // Fallback to cache
    const cachedResponse = await caches.match(request);
    if (cachedResponse) {
      return cachedResponse;
    }

    // Fallback to index.html for SPA routing
    const indexResponse = await caches.match('/index.html');
    if (indexResponse) {
      return indexResponse;
    }

    // Final fallback - offline page
    return new Response(`
      <!DOCTYPE html>
      <html>
      <head>
        <title>Uroboro Timeline - Offline</title>
        <meta charset="utf-8">
        <meta name="viewport" content="width=device-width, initial-scale=1">
        <style>
          body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            margin: 0;
            padding: 2rem;
            background: #1a1a1a;
            color: #ffffff;
            display: flex;
            align-items: center;
            justify-content: center;
            min-height: 100vh;
            text-align: center;
          }
          .offline-content {
            max-width: 400px;
          }
          h1 { color: #4ecdc4; margin-bottom: 1rem; }
          p { color: #cccccc; line-height: 1.6; }
          .retry-btn {
            background: #4ecdc4;
            color: #1a1a1a;
            border: none;
            padding: 0.75rem 1.5rem;
            border-radius: 6px;
            cursor: pointer;
            margin-top: 1rem;
          }
        </style>
      </head>
      <body>
        <div class="offline-content">
          <h1>🐍 Uroboro Timeline</h1>
          <p>You're currently offline. The timeline visualization requires an internet connection to load journey data.</p>
          <button class="retry-btn" onclick="window.location.reload()">
            Try Again
          </button>
        </div>
      </body>
      </html>
    `, {
      status: 200,
      statusText: 'OK',
      headers: {
        'Content-Type': 'text/html'
      }
    });
  }
}

// Handle background sync for offline actions
self.addEventListener('sync', event => {
  console.log('Service Worker: Background sync triggered', event.tag);

  if (event.tag === 'timeline-sync') {
    event.waitUntil(syncTimelineData());
  }
});

// Sync timeline data when back online
async function syncTimelineData() {
  try {
    console.log('Service Worker: Syncing timeline data...');

    // Attempt to fetch fresh data
    const response = await fetch('/api/journey');
    if (response.ok) {
      const cache = await caches.open(API_CACHE_NAME);
      cache.put('/api/journey', response.clone());
      console.log('Service Worker: Timeline data synced successfully');

      // Notify all clients about the update
      const clients = await self.clients.matchAll();
      clients.forEach(client => {
        client.postMessage({
          type: 'TIMELINE_SYNC_SUCCESS',
          timestamp: new Date().toISOString()
        });
      });
    }
  } catch (error) {
    console.error('Service Worker: Timeline sync failed', error);
  }
}

// Handle push notifications (future feature)
self.addEventListener('push', event => {
  if (event.data) {
    const data = event.data.json();
    console.log('Service Worker: Push notification received', data);

    const options = {
      body: data.body || 'New journey events available',
      icon: '/icons/icon-192x192.png',
      badge: '/icons/icon-72x72.png',
      tag: 'timeline-update',
      data: {
        url: data.url || '/'
      }
    };

    event.waitUntil(
      self.registration.showNotification(
        data.title || 'Uroboro Timeline',
        options
      )
    );
  }
});

// Handle notification clicks
self.addEventListener('notificationclick', event => {
  console.log('Service Worker: Notification clicked', event);

  event.notification.close();

  const urlToOpen = event.notification.data?.url || '/';

  event.waitUntil(
    self.clients.matchAll({ type: 'window' }).then(clients => {
      // Check if there's already a window/tab open with the target URL
      for (const client of clients) {
        if (client.url === urlToOpen && 'focus' in client) {
          return client.focus();
        }
      }

      // If not, open a new window/tab
      if (self.clients.openWindow) {
        return self.clients.openWindow(urlToOpen);
      }
    })
  );
});

// Log service worker errors
self.addEventListener('error', event => {
  console.error('Service Worker: Error occurred', event.error);
});

// Log unhandled promise rejections
self.addEventListener('unhandledrejection', event => {
  console.error('Service Worker: Unhandled promise rejection', event.reason);
});

console.log('Service Worker: Script loaded successfully');
