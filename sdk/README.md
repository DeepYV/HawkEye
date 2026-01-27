# HawkEye Observer SDK

The HawkEye Observer SDK is a lightweight, privacy-focused JavaScript library for tracking user frustration signals in web applications.

## Quick Start

Get up and running in under 5 minutes.

### 1. Install

```bash
npm install @hawkeye/observer-sdk
```

### 2. Initialize

```javascript
import { initFrustrationObserver } from '@hawkeye/observer-sdk';

initFrustrationObserver({
  apiKey: 'your-api-key-here',
  ingestionUrl: 'https://api.hawkeye.example.com',
});
```

### 3. Done!

The SDK automatically tracks user interactions. No additional code needed.

## Mental Model: How HawkEye Works

### What Happens After You Initialize?

```
┌─────────────────────────────────────────────────────────────┐
│ 1. User interacts with your app                            │
│    (clicks, scrolls, types, navigates, encounters errors)   │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│ 2. SDK captures behavioral signals (NOT content)            │
│    ✓ What: Button clicked, form focused, error occurred     │
│    ✗ NOT: Input values, passwords, personal data            │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│ 3. Events are batched locally (default: 10 events)          │
│    Reduces network overhead, improves performance            │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│ 4. Batch sent to your ingestion endpoint                    │
│    Includes session ID, timestamps, interaction patterns     │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│ 5. HawkEye Frustration Engine analyzes patterns             │
│    Detects rage clicks, error cascades, navigation loops    │
└─────────────────────────────────────────────────────────────┘
```

### Core Concepts

- **Session**: A unique identifier for each user's visit
- **Event**: A captured user interaction (click, scroll, error, etc.)
- **Batch**: A group of events sent together to reduce network calls
- **Observer**: A component that watches for specific interaction types
- **Ingestion**: The process of receiving and storing events on your backend

## Installation & Setup

### React

```jsx
import { useEffect } from 'react';
import { initFrustrationObserver, teardown } from '@hawkeye/observer-sdk';

function App() {
  useEffect(() => {
    initFrustrationObserver({
      apiKey: 'your-api-key',
      ingestionUrl: 'https://api.hawkeye.example.com',
      environment: process.env.NODE_ENV,
    });

    return () => teardown();
  }, []);

  return <div>Your App</div>;
}
```

[See full React example →](../examples/react/README.md)

### Next.js

```jsx
// pages/_app.js
import { useEffect } from 'react';
import { initFrustrationObserver } from '@hawkeye/observer-sdk';

function MyApp({ Component, pageProps }) {
  useEffect(() => {
    if (typeof window !== 'undefined') {
      initFrustrationObserver({
        apiKey: process.env.NEXT_PUBLIC_HAWKEYE_API_KEY,
        ingestionUrl: process.env.NEXT_PUBLIC_HAWKEYE_INGESTION_URL,
      });
    }
  }, []);

  return <Component {...pageProps} />;
}
```

[See full Next.js example →](../examples/nextjs/README.md)

### Vanilla JavaScript

```html
<script type="module">
  import { initFrustrationObserver } from 'https://unpkg.com/@hawkeye/observer-sdk';

  initFrustrationObserver({
    apiKey: 'your-api-key',
    ingestionUrl: 'https://api.hawkeye.example.com',
  });
</script>
```

[See full Vanilla JS example →](../examples/vanilla-js/README.md)

## API Reference

### `initFrustrationObserver(config)`

Initialize the SDK and start tracking.

```javascript
initFrustrationObserver({
  // Required
  apiKey: string,              // Your HawkEye API key
  ingestionUrl: string,        // Your ingestion endpoint

  // Optional
  enableDebug?: boolean,       // Enable console logs (default: false)
  environment?: string,        // 'production' | 'staging' | 'development' (default: auto-detect)
  appId?: string,              // Application identifier
  projectId?: string,          // Project identifier
  batchSize?: number,          // Events per batch (default: 10)
  batchInterval?: number,      // Batch interval in ms (default: 5000)
});
```

### `captureEvent(eventType, target, metadata?)`

Manually track a custom event.

```javascript
import { captureEvent } from '@hawkeye/observer-sdk';

captureEvent(
  'purchase_completed',           // Event type (string)
  { type: 'button', id: 'buy' },  // Target element info
  { amount: 99.99, items: 3 }     // Optional metadata
);
```

### `getSessionId()`

Get the current session ID.

```javascript
import { getSessionId } from '@hawkeye/observer-sdk';

const sessionId = getSessionId();
console.log('Current session:', sessionId);
```

### `flushEvents()`

Immediately send all pending events (useful before page unload).

```javascript
import { flushEvents } from '@hawkeye/observer-sdk';

window.addEventListener('beforeunload', () => {
  flushEvents();
});
```

### `teardown()`

Stop tracking and cleanup resources.

```javascript
import { teardown } from '@hawkeye/observer-sdk';

// Call when unmounting in SPA
teardown();
```

## Configuration

### Sensible Defaults

HawkEye comes with production-ready defaults:

| Option | Default | Purpose |
|--------|---------|---------|
| `batchSize` | `10` | Balance between latency and network overhead |
| `batchInterval` | `5000ms` | Send events every 5 seconds |
| `enableDebug` | `false` | No console noise in production |
| `environment` | auto-detect | Detects dev/staging/prod from URL |

### Environment Detection

The SDK automatically detects your environment:

- `localhost` or `127.0.0.1` → `development`
- URLs with `staging`, `stage`, or `test` → `staging`
- Everything else → `production`

Override with: `environment: 'production'`

### Safe-by-Default Behavior

- **No false alarms**: Events are deduplicated using idempotency keys
- **No memory leaks**: Automatic cleanup on page unload
- **No blocking**: All operations are non-blocking and async
- **No crashes**: Errors are caught and logged (only in debug mode)

## What Gets Tracked?

### Automatic Tracking

The SDK automatically captures:

| Event Type | What's Captured | What's NOT Captured |
|------------|-----------------|---------------------|
| **Clicks** | Element type, ID, selector, timestamp | Text content, user data |
| **Form Inputs** | Focus, blur, submission events | Input values, passwords |
| **Scrolling** | Scroll behavior, patterns | Page content |
| **Errors** | Error message, stack trace | Sensitive data in messages |
| **Navigation** | Route changes, page loads | Query parameters, URL fragments |

### Privacy Guarantees

✅ **What IS captured:**
- Interaction patterns (click sequences, scroll behavior)
- Element identifiers (IDs, selectors, tag names)
- Timestamps and session data
- Error types and messages

❌ **What is NOT captured:**
- Input field values
- Password fields (completely ignored)
- Personal information (PII)
- Page content or text
- Cookies or local storage
- User keystrokes (only focus/blur events)

## Common Use Cases

### Track Custom Events

```javascript
import { captureEvent } from '@hawkeye/observer-sdk';

// Track feature usage
function enableDarkMode() {
  // ... your code ...
  captureEvent('feature_toggled', { type: 'toggle' }, { feature: 'dark_mode', enabled: true });
}

// Track checkout steps
function proceedToPayment() {
  // ... your code ...
  captureEvent('checkout_step', { type: 'form' }, { step: 'payment', cartValue: 299.99 });
}
```

### Track Page Views (SPA)

```javascript
// React Router
import { useLocation } from 'react-router-dom';
import { captureEvent } from '@hawkeye/observer-sdk';

function RouteTracker() {
  const location = useLocation();

  useEffect(() => {
    captureEvent('page_view', { type: 'navigation' }, { path: location.pathname });
  }, [location]);

  return null;
}
```

### Track Errors

```javascript
// Global error handler
window.addEventListener('error', (event) => {
  captureEvent('app_error', { type: 'error' }, {
    message: event.message,
    filename: event.filename,
    line: event.lineno,
  });
});
```

### Debug Mode

```javascript
initFrustrationObserver({
  apiKey: 'your-key',
  ingestionUrl: 'https://api.example.com',
  enableDebug: true, // See SDK logs in console
});
```

## Common Mistakes

### ❌ Initializing Multiple Times

```javascript
// DON'T: Initialize in every component
function Header() {
  useEffect(() => {
    initFrustrationObserver({ ... }); // ❌
  }, []);
}

function Footer() {
  useEffect(() => {
    initFrustrationObserver({ ... }); // ❌
  }, []);
}
```

```javascript
// DO: Initialize once in root component
function App() {
  useEffect(() => {
    initFrustrationObserver({ ... }); // ✅
  }, []);

  return <><Header /><Content /><Footer /></>;
}
```

### ❌ Forgetting Server-Side Rendering (Next.js)

```javascript
// DON'T: Initialize without window check
initFrustrationObserver({ ... }); // ❌ Crashes on SSR
```

```javascript
// DO: Check for window or use useEffect
useEffect(() => {
  if (typeof window !== 'undefined') {
    initFrustrationObserver({ ... }); // ✅
  }
}, []);
```

### ❌ Tracking Sensitive Data

```javascript
// DON'T: Send PII in custom events
captureEvent('form_submit', target, {
  email: 'user@example.com',  // ❌
  password: '123456',          // ❌
  ssn: '123-45-6789',          // ❌
});
```

```javascript
// DO: Track behavior without sensitive data
captureEvent('form_submit', target, {
  formType: 'signup',          // ✅
  step: 'credentials',         // ✅
  hasErrors: false,            // ✅
});
```

### ❌ Blocking the Main Thread

```javascript
// DON'T: Flush synchronously on every click
button.addEventListener('click', () => {
  flushEvents(); // ❌ Unnecessary, SDK batches automatically
});
```

```javascript
// DO: Let SDK batch automatically, only flush when needed
window.addEventListener('beforeunload', () => {
  flushEvents(); // ✅ Flush before page unload
});
```

## FAQ

### Performance

**Q: Will this slow down my app?**

No. The SDK is designed for zero-impact performance:
- Event capture is non-blocking
- Events are batched (default: 10 events or 5 seconds)
- Observers use passive event listeners
- Total bundle size: ~15KB gzipped

**Q: How much network traffic does it generate?**

Minimal. With default settings:
- ~10 events per request
- Average event size: ~200 bytes
- ~2KB per request, every 5 seconds (active users only)

### Privacy

**Q: Is this GDPR compliant?**

Yes. The SDK:
- Does not capture personal information
- Does not track users across sites
- Does not use cookies or fingerprinting
- Only captures behavioral patterns

Consult your legal team for your specific use case.

**Q: What about password fields?**

Password fields are completely ignored. No events are captured for password input fields.

**Q: Can I disable tracking for specific users?**

Yes. Simply don't initialize the SDK, or call `teardown()` to stop tracking.

```javascript
if (user.hasOptedOut) {
  teardown();
}
```

### Debugging

**Q: How do I know if it's working?**

Enable debug mode:

```javascript
initFrustrationObserver({
  apiKey: 'your-key',
  ingestionUrl: 'https://api.example.com',
  enableDebug: true,
});
```

Check your browser's console and network tab.

**Q: Events aren't being sent?**

Common issues:
1. Check API key and ingestion URL
2. Verify CORS settings on your endpoint
3. Check network tab for failed requests
4. Enable debug mode to see console logs
5. Verify endpoint expects `POST /v1/events`

**Q: Can I test without a backend?**

Yes. Set a dummy ingestion URL and enable debug mode. Check the console to see captured events.

### Integration

**Q: Does this work with [framework]?**

Yes. The SDK is framework-agnostic and works with:
- React, Vue, Angular, Svelte
- Next.js, Nuxt, SvelteKit
- Vanilla JavaScript
- Any JavaScript environment with `fetch` and DOM APIs

**Q: What about TypeScript?**

Full TypeScript support included. Types are automatically available:

```typescript
import { initFrustrationObserver, SDKConfig } from '@hawkeye/observer-sdk';

const config: SDKConfig = {
  apiKey: 'your-key',
  ingestionUrl: 'https://api.example.com',
};

initFrustrationObserver(config);
```

**Q: Can I use this with Google Analytics / Mixpanel / etc?**

Yes. HawkEye is complementary to traditional analytics. It focuses on frustration signals, not general analytics.

## Troubleshooting

### Initialization Errors

```
Error: HawkEye SDK: apiKey is required
```

**Solution**: Provide a valid API key.

```javascript
initFrustrationObserver({
  apiKey: 'your-actual-api-key', // Don't forget this!
  ingestionUrl: 'https://api.example.com',
});
```

### Network Errors

```
Error: HTTP 401: Unauthorized
```

**Solution**: Check your API key is correct and has permissions.

```
Error: HTTP 404: Not Found
```

**Solution**: Verify your ingestion URL. The SDK expects `POST /v1/events`.

### SSR Errors (Next.js)

```
ReferenceError: window is not defined
```

**Solution**: Initialize in `useEffect` or check for `window`:

```javascript
useEffect(() => {
  if (typeof window !== 'undefined') {
    initFrustrationObserver({ ... });
  }
}, []);
```

## Support

- **Documentation**: See [examples](../examples/) for integration guides
- **Issues**: Open an issue on GitHub
- **Security**: Report security issues to security@hawkeye.example.com

## License

MIT License - see LICENSE file for details.

---

Made with ❤️ by the HawkEye Team
