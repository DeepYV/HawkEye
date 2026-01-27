# Vanilla JavaScript Integration Example

This example demonstrates how to integrate HawkEye Observer SDK into a vanilla JavaScript application.

## Quick Setup

### Option 1: Using ES Modules (Modern)

```html
<!DOCTYPE html>
<html>
<head>
    <title>My App</title>
</head>
<body>
    <h1>My Application</h1>

    <script type="module">
        import { initFrustrationObserver } from 'https://unpkg.com/@hawkeye/observer-sdk';

        initFrustrationObserver({
            apiKey: 'your-api-key-here',
            ingestionUrl: 'https://api.hawkeye.example.com',
            enableDebug: true,
        });
    </script>
</body>
</html>
```

### Option 2: Using NPM and Bundler

```bash
npm install @hawkeye/observer-sdk
```

```javascript
import { initFrustrationObserver } from '@hawkeye/observer-sdk';

initFrustrationObserver({
    apiKey: 'your-api-key-here',
    ingestionUrl: 'https://api.hawkeye.example.com',
    enableDebug: true,
});
```

### Option 3: UMD Build (Legacy)

```html
<script src="https://unpkg.com/@hawkeye/observer-sdk/dist/umd/hawkeye.min.js"></script>
<script>
    window.HawkEye.initFrustrationObserver({
        apiKey: 'your-api-key-here',
        ingestionUrl: 'https://api.hawkeye.example.com',
    });
</script>
```

## Complete Example

```javascript
import {
    initFrustrationObserver,
    getSessionId,
    captureEvent,
    teardown
} from '@hawkeye/observer-sdk';

// Initialize on page load
document.addEventListener('DOMContentLoaded', () => {
    initFrustrationObserver({
        apiKey: 'your-api-key-here',
        ingestionUrl: 'https://api.hawkeye.example.com',
        enableDebug: true,
        environment: 'production',
        appId: 'my-app',
    });

    console.log('Session ID:', getSessionId());
});

// Cleanup on unload
window.addEventListener('beforeunload', () => {
    teardown();
});
```

## Tracking Custom Events

### Basic Custom Event

```javascript
import { captureEvent } from '@hawkeye/observer-sdk';

document.getElementById('checkout-button').addEventListener('click', () => {
    captureEvent(
        'checkout_initiated',
        { type: 'button', id: 'checkout-button' },
        { cartValue: 299.99, itemCount: 3 }
    );
});
```

### Track Data Attributes

Add `data-track` attributes to elements:

```html
<button data-track="signup_clicked">Sign Up</button>
<button data-track="demo_requested">Request Demo</button>
```

Then track them automatically:

```javascript
document.querySelectorAll('[data-track]').forEach((element) => {
    element.addEventListener('click', (e) => {
        const eventName = e.target.getAttribute('data-track');
        captureEvent(eventName, {
            type: e.target.tagName.toLowerCase(),
            id: e.target.id,
        });
    });
});
```

## Advanced Usage

### Track Page Visibility

```javascript
document.addEventListener('visibilitychange', () => {
    if (document.hidden) {
        captureEvent('page_hidden', { type: 'page' });
    } else {
        captureEvent('page_visible', { type: 'page' });
    }
});
```

### Track Scroll Depth

```javascript
let maxScroll = 0;

window.addEventListener('scroll', () => {
    const scrollPercent = (window.scrollY / (document.body.scrollHeight - window.innerHeight)) * 100;

    if (scrollPercent > maxScroll + 25) {
        maxScroll = Math.floor(scrollPercent / 25) * 25;
        captureEvent('scroll_depth', { type: 'scroll' }, { depth: maxScroll });
    }
});
```

### Track Feature Usage

```javascript
function trackFeatureUse(featureName, metadata = {}) {
    captureEvent('feature_used', { type: 'feature' }, {
        feature: featureName,
        ...metadata,
    });
}

// Usage
trackFeatureUse('dark_mode', { enabled: true });
trackFeatureUse('export', { format: 'pdf' });
```

## Configuration Options

```javascript
initFrustrationObserver({
    // Required
    apiKey: 'your-api-key',                    // Your HawkEye API key
    ingestionUrl: 'https://api.example.com',   // Ingestion endpoint

    // Optional
    enableDebug: false,                        // Enable console logs
    environment: 'production',                 // Environment name
    appId: 'my-app',                           // Application identifier
    batchSize: 10,                             // Events per batch
    batchInterval: 5000,                       // Batch interval (ms)
});
```

## Best Practices

1. **Initialize Early**: Call `initFrustrationObserver` as soon as possible
2. **Single Instance**: Only initialize once per page
3. **Cleanup**: Call `teardown()` before page unload for SPA navigation
4. **Custom Events**: Use meaningful event names (e.g., 'checkout_completed', not 'button_click')
5. **Metadata**: Include relevant context in metadata, but avoid PII

## What Gets Tracked Automatically?

- **Clicks**: All click events on interactive elements
- **Forms**: Form focus, blur, and submission (not values)
- **Scrolling**: Scroll behavior and patterns
- **Errors**: JavaScript errors and exceptions
- **Navigation**: Page loads and route changes

## Privacy & Security

- **No Content**: Input values and text content are never captured
- **No Passwords**: Password fields are completely ignored
- **No PII**: Personal information is not collected
- **Behavioral Only**: Only interaction patterns are tracked

## Troubleshooting

### SDK Not Working

1. Check browser console for initialization errors
2. Enable debug mode: `enableDebug: true`
3. Verify API key and ingestion URL are correct
4. Check network tab for API requests

### Events Not Sending

1. Check batch settings (size and interval)
2. Manually flush events: `flushEvents()`
3. Check for network connectivity
4. Verify CORS settings on ingestion endpoint
