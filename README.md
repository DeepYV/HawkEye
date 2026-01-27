# HawkEye - Frustration Detection Engine

HawkEye is a real-time user frustration detection system designed to help product teams identify and fix UX issues before they impact user retention.

## What is HawkEye?

HawkEye analyzes user behavior patterns to detect frustration signals like:
- **Rage clicks**: Rapid, repeated clicks indicating confusion
- **Error cascades**: Multiple errors in quick succession
- **Navigation loops**: Users going back and forth between pages
- **Dead ends**: Users getting stuck without clear next steps
- **Scroll thrashing**: Excessive scrolling suggesting users can't find what they need

## Quick Start

### 1. Install the SDK

```bash
npm install @hawkeye/observer-sdk
```

### 2. Initialize in Your App

```javascript
import { initFrustrationObserver } from '@hawkeye/observer-sdk';

initFrustrationObserver({
  apiKey: 'your-api-key-here',
  ingestionUrl: 'https://your-hawkeye-instance.com',
});
```

### 3. Deploy the Backend

```bash
# Clone the repository
git clone https://github.com/DeepYV/HawkEye.git
cd HawkEye

# Build and run
go build -o hawkeye ./cmd/server
./hawkeye
```

## Components

### 🎯 Observer SDK (JavaScript)

The frontend SDK that captures user interaction signals.

- **Lightweight**: ~15KB gzipped
- **Privacy-focused**: No PII, no content capture
- **Framework-agnostic**: Works with React, Vue, Next.js, vanilla JS
- **Zero-config**: Sensible defaults out of the box

[Read SDK Documentation →](./sdk/README.md)

### 🔧 Frustration Engine (Go)

The backend engine that processes signals and detects frustration patterns.

- **Real-time processing**: Instant frustration detection
- **Scalable**: Handles high-volume event streams
- **Extensible**: Plugin-based architecture
- **Observable**: Built-in metrics and tracing

[Read Backend Documentation →](./docs/backend.md) _(coming soon)_

## Integration Examples

### React

```jsx
import { useEffect } from 'react';
import { initFrustrationObserver } from '@hawkeye/observer-sdk';

function App() {
  useEffect(() => {
    initFrustrationObserver({
      apiKey: 'your-api-key',
      ingestionUrl: 'https://api.hawkeye.example.com',
    });
  }, []);

  return <div>Your App</div>;
}
```

[Full React Example →](./examples/react/README.md)

### Next.js

```jsx
// pages/_app.js
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

[Full Next.js Example →](./examples/nextjs/README.md)

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

[Full Vanilla JS Example →](./examples/vanilla-js/README.md)

## Architecture

```
┌──────────────────┐
│   Your Web App   │
│  (React/Next.js) │
└────────┬─────────┘
         │
         │ (Observer SDK captures signals)
         │
         ▼
┌──────────────────┐
│ Ingestion API    │
│ /v1/events       │
└────────┬─────────┘
         │
         │ (Events batched and queued)
         │
         ▼
┌──────────────────┐
│ Frustration      │
│ Engine (UFSE)    │
│ - Pattern Match  │
│ - Scoring        │
│ - Detection      │
└────────┬─────────┘
         │
         │ (Alerts and insights)
         │
         ▼
┌──────────────────┐
│ Webhooks / Alerts│
│ Dashboard / API  │
└──────────────────┘
```

## Features

### Privacy-First Design

- **No PII collection**: Only behavioral patterns
- **No content capture**: Input values never recorded
- **GDPR compliant**: Transparent data handling
- **User control**: Easy opt-out mechanism

### Production-Ready

- **Safe defaults**: No false alarms out of the box
- **Graceful degradation**: Fails silently without breaking your app
- **Low overhead**: Minimal performance impact
- **Battle-tested**: Used in production environments

### Developer-Friendly

- **10-minute setup**: Get started quickly
- **Framework examples**: React, Next.js, Vanilla JS
- **TypeScript support**: Full type definitions included
- **Comprehensive docs**: Clear guides and API reference

## Use Cases

### Product Teams

Identify where users struggle and prioritize fixes based on real frustration data.

### Engineering Teams

Detect errors and UX issues before they escalate into support tickets.

### Design Teams

Validate designs with real user behavior data, not just assumptions.

## Documentation

- [SDK Documentation](./sdk/README.md) - Complete SDK API reference
- [React Integration](./examples/react/README.md) - React setup guide
- [Next.js Integration](./examples/nextjs/README.md) - Next.js with SSR
- [Vanilla JS Integration](./examples/vanilla-js/README.md) - Plain JavaScript

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines.

## License

MIT License - see [LICENSE](./LICENSE) for details.

## Support

- **GitHub Issues**: For bug reports and feature requests
- **Discussions**: For questions and community support
- **Security**: Report security issues to security@hawkeye.example.com

---

**Built with ❤️ for better user experiences**
