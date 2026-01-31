/**
 * HawkEye Vanilla JS Integration Example
 * Demonstrates initialization and custom event tracking
 */

// Import SDK (adjust path based on your setup)
import { initFrustrationObserver, getSessionId, captureEvent, teardown } from '@hawkeye/observer-sdk';

// Configuration
// For local development with the dev server (./scripts/start_dev.sh),
// use http://localhost:8080 as the single ingestionUrl.
const config = {
  apiKey: 'your-api-key-here',             // Use 'dev-api-key' with the dev server
  ingestionUrl: 'https://api.hawkeye.example.com', // Use 'http://localhost:8080' for local dev
  enableDebug: true,
  environment: 'production',
  appId: 'my-vanilla-app',
  batchSize: 10,
  batchInterval: 5000,
};

// Initialize on page load
document.addEventListener('DOMContentLoaded', () => {
  initFrustrationObserver(config);

  console.log('HawkEye initialized with session:', getSessionId());
});

// Cleanup on page unload
window.addEventListener('beforeunload', () => {
  teardown();
});

// Example: Track custom events
function trackCustomEvent(eventType, metadata = {}) {
  captureEvent(
    eventType,
    { type: 'custom', id: 'app' },
    metadata
  );
}

// Example: Track specific button clicks
document.querySelectorAll('[data-track]').forEach((element) => {
  element.addEventListener('click', (e) => {
    const eventName = e.target.getAttribute('data-track');
    trackCustomEvent(eventName, {
      elementId: e.target.id,
      elementText: e.target.textContent,
    });
  });
});

// Example: Track page visibility
document.addEventListener('visibilitychange', () => {
  if (document.hidden) {
    captureEvent('page_hidden', { type: 'page' }, { timestamp: Date.now() });
  } else {
    captureEvent('page_visible', { type: 'page' }, { timestamp: Date.now() });
  }
});
