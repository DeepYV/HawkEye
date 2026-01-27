# React Integration Example

This example demonstrates how to integrate HawkEye Observer SDK into a React application.

## Quick Setup

### 1. Install the SDK

```bash
npm install @hawkeye/observer-sdk
```

### 2. Initialize in Your App Component

```jsx
import React, { useEffect } from 'react';
import { initFrustrationObserver, teardown } from '@hawkeye/observer-sdk';

function App() {
  useEffect(() => {
    initFrustrationObserver({
      apiKey: 'your-api-key-here',
      ingestionUrl: 'https://api.hawkeye.example.com',
      enableDebug: process.env.NODE_ENV === 'development',
      environment: process.env.NODE_ENV,
    });

    return () => {
      teardown();
    };
  }, []);

  return <div>Your App</div>;
}
```

## Using the Custom Hook

For cleaner code, use the provided `useHawkEye` hook:

```jsx
import { useHawkEye } from './useHawkEye';

function App() {
  useHawkEye({
    apiKey: 'your-api-key-here',
    ingestionUrl: 'https://api.hawkeye.example.com',
    enableDebug: process.env.NODE_ENV === 'development',
  });

  return <div>Your App</div>;
}
```

## Tracking Custom Events

```jsx
import { captureEvent } from '@hawkeye/observer-sdk';

function PremiumButton() {
  const handleUpgrade = () => {
    captureEvent(
      'upgrade_clicked',
      { type: 'button', id: 'premium-upgrade' },
      { plan: 'premium', price: 29.99 }
    );
  };

  return <button onClick={handleUpgrade}>Upgrade to Premium</button>;
}
```

## Best Practices

1. **Initialize Early**: Place initialization in your root App component
2. **Environment Detection**: Use `process.env.NODE_ENV` for automatic environment detection
3. **Debug Mode**: Enable debug mode in development to see SDK logs
4. **Cleanup**: Always call `teardown()` in the useEffect cleanup function
5. **Single Instance**: Only initialize once in your application

## What Gets Tracked?

HawkEye automatically tracks:
- Click events
- Form inputs (content not captured, only interactions)
- Scroll behavior
- JavaScript errors
- Page navigation
- Route changes (React Router compatible)

## Privacy & Security

- **No sensitive data**: Input values are never captured
- **No passwords**: Password fields are ignored
- **No PII**: Personal information is not collected
- **GDPR compliant**: Only behavioral signals are tracked
