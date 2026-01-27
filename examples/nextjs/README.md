# Next.js Integration Example

This example demonstrates how to integrate HawkEye Observer SDK into a Next.js application.

## Quick Setup

### 1. Install the SDK

```bash
npm install @hawkeye/observer-sdk
```

### 2. Configure Environment Variables

Create a `.env.local` file in your project root:

```env
NEXT_PUBLIC_HAWKEYE_API_KEY=your-api-key-here
NEXT_PUBLIC_HAWKEYE_INGESTION_URL=https://api.hawkeye.example.com
```

### 3. Initialize in `_app.js`

```jsx
// pages/_app.js
import { useEffect } from 'react';
import { initFrustrationObserver, teardown } from '@hawkeye/observer-sdk';

function MyApp({ Component, pageProps }) {
  useEffect(() => {
    if (typeof window !== 'undefined') {
      initFrustrationObserver({
        apiKey: process.env.NEXT_PUBLIC_HAWKEYE_API_KEY,
        ingestionUrl: process.env.NEXT_PUBLIC_HAWKEYE_INGESTION_URL,
        enableDebug: process.env.NODE_ENV === 'development',
        environment: process.env.NODE_ENV,
      });
    }

    return () => {
      teardown();
    };
  }, []);

  return <Component {...pageProps} />;
}

export default MyApp;
```

## Using the App Router (Next.js 13+)

For Next.js 13+ with App Router, create a client component:

```jsx
// app/providers.jsx
'use client';

import { useEffect } from 'react';
import { initFrustrationObserver, teardown } from '@hawkeye/observer-sdk';

export function HawkEyeProvider({ children }) {
  useEffect(() => {
    initFrustrationObserver({
      apiKey: process.env.NEXT_PUBLIC_HAWKEYE_API_KEY,
      ingestionUrl: process.env.NEXT_PUBLIC_HAWKEYE_INGESTION_URL,
      enableDebug: process.env.NODE_ENV === 'development',
    });

    return () => {
      teardown();
    };
  }, []);

  return <>{children}</>;
}
```

Then use it in your layout:

```jsx
// app/layout.jsx
import { HawkEyeProvider } from './providers';

export default function RootLayout({ children }) {
  return (
    <html>
      <body>
        <HawkEyeProvider>
          {children}
        </HawkEyeProvider>
      </body>
    </html>
  );
}
```

## Tracking Custom Events

```jsx
import { captureEvent } from '@hawkeye/observer-sdk';

export default function CheckoutPage() {
  const handlePurchase = () => {
    captureEvent(
      'purchase_completed',
      { type: 'button', id: 'checkout-button' },
      { orderId: '12345', amount: 99.99 }
    );
  };

  return <button onClick={handlePurchase}>Complete Purchase</button>;
}
```

## Server-Side Rendering (SSR) Considerations

HawkEye SDK is **client-side only**. Always ensure:

1. Initialize inside `useEffect` or client components
2. Check `typeof window !== 'undefined'` before initializing
3. Use `NEXT_PUBLIC_` prefix for environment variables
4. Dynamic imports for conditional loading (see `lib/hawkeye.js`)

## Best Practices

1. **Environment Variables**: Use `NEXT_PUBLIC_` prefix for client-side variables
2. **Client-Side Only**: Always initialize in `useEffect` or client components
3. **Single Initialization**: Initialize once in `_app.js` or root layout
4. **Dynamic Imports**: Use dynamic imports to prevent SSR issues
5. **Cleanup**: Call `teardown()` in cleanup function

## Automatic Route Tracking

HawkEye automatically tracks Next.js route changes through:
- Page Router: Automatically detects route changes
- App Router: Tracks navigation via browser history API

## What Gets Tracked?

- Click events on all interactive elements
- Form interactions (not values)
- Scroll behavior and patterns
- JavaScript errors
- Page navigation and route changes
- Custom events you define

## Privacy & Security

- **No sensitive data**: Form values are never captured
- **No passwords**: Password fields are completely ignored
- **Client-side only**: No server-side data collection
- **Environment aware**: Different behavior for dev/staging/prod
