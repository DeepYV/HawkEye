# Next.js

Initialize in `useEffect` or a client component. Do not call `init` at module scope in a server component. `init` also returns without throwing when `window` is missing, so the package can be imported during SSR.

`setRoute` is how a Next.js app supplies the path template. Pages Router `router.pathname` uses `[id]`; HawkEye templates use `:id`.

```jsx
// pages/_app.js
import { useEffect } from 'react';
import { useRouter } from 'next/router';
import { init, teardown, setRoute } from '@hawkeye/observer-sdk';

function template(pathname) {
  return pathname.replace(/\[\.\.\.[^\]]+\]/g, ':id').replace(/\[[^\]]+\]/g, ':id');
}

function MyApp({ Component, pageProps }) {
  const router = useRouter();

  useEffect(() => {
    init({
      apiKey: process.env.NEXT_PUBLIC_HAWKEYE_API_KEY,
      ingestionUrl: process.env.NEXT_PUBLIC_HAWKEYE_INGESTION_URL,
      appVersion: process.env.NEXT_PUBLIC_APP_VERSION,
      events: ['signup_submitted'],
      importantFlows: [{ name: 'signup', steps: ['account'], important: true }],
    });
    return () => {
      teardown();
    };
  }, []);

  useEffect(() => {
    setRoute(template(router.pathname));
  }, [router.pathname]);

  return <Component {...pageProps} />;
}

export default MyApp;
```

App Router: put the same `useEffect` in a client component and render that component from the root layout. Do not call `init` in the server layout itself.

```jsx
'use client';

import { useEffect } from 'react';
import { usePathname } from 'next/navigation';
import { init, teardown, setRoute } from '@hawkeye/observer-sdk';

export function HawkEyeProvider({ children }) {
  const pathname = usePathname();

  useEffect(() => {
    init({
      apiKey: process.env.NEXT_PUBLIC_HAWKEYE_API_KEY,
      ingestionUrl: process.env.NEXT_PUBLIC_HAWKEYE_INGESTION_URL,
      events: ['signup_submitted'],
      importantFlows: [{ name: 'signup', steps: ['account'], important: true }],
    });
    return () => {
      teardown();
    };
  }, []);

  useEffect(() => {
    if (pathname) setRoute(pathname);
  }, [pathname]);

  return children;
}
```

On the App Router, pass a template you already rewrote (`/users/:id`), or rewrite dynamic segments before calling `setRoute`. A template that contains `?`, `#`, `@`, or `%` is ignored.

Host calls:

```javascript
import { track, formOutcome, flow } from '@hawkeye/observer-sdk';

const signup = flow('signup');
signup.start();
formOutcome({ formKey: 'signup', result: 'validation_error', fieldKeys: ['email_field'], errorCode: 'invalid' });
track('signup_submitted', { plan: 'free' });
signup.complete();
```

Use a field key such as `email_field`, not the address the person typed. `track` drops a prop named `email`.

Clicks, dead clicks, navigation, JavaScript errors, and failed HTTP calls are automatic. The SDK does not record the DOM, input values, or scroll position.

Environment variables:

```env
NEXT_PUBLIC_HAWKEYE_API_KEY=your-api-key-here
NEXT_PUBLIC_HAWKEYE_INGESTION_URL=http://localhost:8080
NEXT_PUBLIC_APP_VERSION=1.4.0
```
