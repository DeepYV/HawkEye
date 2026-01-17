# Integration Steps - Complete Guide

Step-by-step guide to integrate the Frustration Engine into your application.

---

## 📋 Overview

You have two options:
1. **Use as a service** - Run the backend services and integrate via API
2. **Embed in your project** - Add the code directly to your repository

**You don't need to push to Git first** - you can integrate locally and push later.

---

## 🚀 Option 1: Use as a Service (Recommended)

### Step 1: Keep Services Running

The Frustration Engine runs as separate microservices. Keep them running:

```bash
# Start databases (one time)
cd /Users/deepakyadav/ucfp
./scripts/start_databases.sh

# Start all services (in separate terminals or background)
export POSTGRES_PORT=5434
export CLICKHOUSE_PORT=9001
./scripts/start_services_custom_port.sh
```

**Services will run on:**
- Event Ingestion API: `http://localhost:8080`
- Session Manager: `http://localhost:8081`
- UFSE: `http://localhost:8082`
- Incident Store: `http://localhost:8084`
- Ticket Exporter: `http://localhost:8085`

---

### Step 2: Integrate Frontend SDK

#### For React/Next.js Applications

**Install SDK:**
```bash
npm install @frustration-engine/observer-sdk
# or
yarn add @frustration-engine/observer-sdk
```

**Initialize in your app:**

**Next.js (App Router):**
```typescript
// app/layout.tsx or app/page.tsx
'use client';

import { useEffect } from 'react';
import { initFrustrationObserver } from '@frustration-engine/observer-sdk';

export default function RootLayout({ children }) {
  useEffect(() => {
    initFrustrationObserver({
      apiKey: process.env.NEXT_PUBLIC_FRUSTRATION_API_KEY || 'test-api-key',
      ingestionUrl: process.env.NEXT_PUBLIC_FRUSTRATION_INGESTION_URL || 'http://localhost:8080/v1/events',
      enableDebug: process.env.NODE_ENV === 'development',
    });
  }, []);

  return <html>{children}</html>;
}
```

**Next.js (Pages Router):**
```typescript
// pages/_app.tsx
import { useEffect } from 'react';
import { initFrustrationObserver } from '@frustration-engine/observer-sdk';

export default function App({ Component, pageProps }) {
  useEffect(() => {
    initFrustrationObserver({
      apiKey: process.env.NEXT_PUBLIC_FRUSTRATION_API_KEY || 'test-api-key',
      ingestionUrl: process.env.NEXT_PUBLIC_FRUSTRATION_INGESTION_URL || 'http://localhost:8080/v1/events',
    });
  }, []);

  return <Component {...pageProps} />;
}
```

**React (Create React App):**
```typescript
// src/index.tsx
import { StrictMode } from 'react';
import { initFrustrationObserver } from '@frustration-engine/observer-sdk';

if (process.env.REACT_APP_FRUSTRATION_API_KEY) {
  initFrustrationObserver({
    apiKey: process.env.REACT_APP_FRUSTRATION_API_KEY,
    ingestionUrl: process.env.REACT_APP_FRUSTRATION_INGESTION_URL || 'http://localhost:8080/v1/events',
  });
}

ReactDOM.render(
  <StrictMode>
    <App />
  </StrictMode>,
  document.getElementById('root')
);
```

**Environment Variables (.env.local or .env):**
```bash
# For Next.js
NEXT_PUBLIC_FRUSTRATION_API_KEY=test-api-key
NEXT_PUBLIC_FRUSTRATION_INGESTION_URL=http://localhost:8080/v1/events

# For Create React App
REACT_APP_FRUSTRATION_API_KEY=test-api-key
REACT_APP_FRUSTRATION_INGESTION_URL=http://localhost:8080/v1/events
```

---

#### For Vanilla JavaScript/HTML

```html
<!DOCTYPE html>
<html>
<head>
  <title>My App</title>
</head>
<body>
  <!-- Your app content -->
  
  <!-- Load SDK -->
  <script src="https://cdn.yourdomain.com/frustration-observer.min.js"></script>
  <script>
    FrustrationObserver.init({
      apiKey: 'test-api-key',
      ingestionUrl: 'http://localhost:8080/v1/events',
    });
  </script>
</body>
</html>
```

---

### Step 3: Test Integration

```bash
# 1. Open your app in browser
# 2. Interact with your app (click buttons, fill forms, trigger errors)
# 3. Check Event Ingestion logs
tail -f /tmp/Event\ Ingestion.log

# 4. Send test event manually
curl -X POST http://localhost:8080/v1/events \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-api-key" \
  -d '{
    "project_id": "my-project",
    "events": [{
      "eventType": "click",
      "timestamp": "2024-01-16T10:00:00Z",
      "sessionId": "test-session-123",
      "route": "/test",
      "target": {"type": "button", "id": "test-btn"}
    }]
  }'
```

---

## 🔧 Option 2: Embed in Your Project

### Step 1: Add to Your Git Repository

**Option A: Git Submodule (Recommended)**
```bash
# In your project directory
cd /path/to/your/project
git submodule add https://github.com/your-org/frustration-engine.git frustration-engine
git commit -m "Add Frustration Engine as submodule"
```

**Option B: Copy Code**
```bash
# Copy the frustration-engine code into your project
cp -r /Users/deepakyadav/ucfp /path/to/your/project/frustration-engine
cd /path/to/your/project
git add frustration-engine
git commit -m "Add Frustration Engine"
```

**Option C: Git Subtree**
```bash
# In your project directory
cd /path/to/your/project
git subtree add --prefix=frustration-engine https://github.com/your-org/frustration-engine.git main --squash
```

---

### Step 2: Initialize Git (If Not Already Done)

```bash
cd /Users/deepakyadav/ucfp

# Check if git is initialized
git status

# If not initialized, initialize it
git init
git add .
git commit -m "Initial commit: Frustration Engine MVP"

# Create remote repository (GitHub, GitLab, etc.) and push
git remote add origin https://github.com/your-username/frustration-engine.git
git branch -M main
git push -u origin main
```

---

### Step 3: Use in Your Application

**If using Go modules:**
```bash
# In your project's go.mod
go get github.com/your-org/frustration-engine@latest

# Or use local path
go mod edit -replace github.com/your-org/frustration-engine=./frustration-engine
```

**Import in your Go code:**
```go
import (
    "github.com/your-org/frustration-engine/internal/server"
    "github.com/your-org/frustration-engine/internal/types"
)
```

---

## 📦 Deployment Options

### Option A: Deploy Services Separately

Deploy the Frustration Engine services to your infrastructure:

```bash
# Build Docker images
docker build -t frustration-engine-event-ingestion ./cmd/event-ingestion
docker build -t frustration-engine-session-manager ./cmd/session-manager
# ... etc

# Deploy to your cloud provider (AWS, GCP, Azure)
# Use Kubernetes, ECS, Cloud Run, etc.
```

**Update frontend SDK URL:**
```typescript
initFrustrationObserver({
  apiKey: 'your-production-api-key',
  ingestionUrl: 'https://api.yourdomain.com/v1/events', // Production URL
});
```

---

### Option B: Deploy as Part of Your Application

If embedding in your project, deploy together:

```bash
# Build your application with Frustration Engine included
go build -o myapp ./cmd/myapp

# Deploy as usual
```

---

## 🔐 Production Setup

### 1. Generate Production API Keys

```bash
# Create API keys through Incident Store API
curl -X POST http://localhost:8084/v1/admin/api-keys \
  -H "Content-Type: application/json" \
  -d '{
    "project_id": "production-project",
    "name": "Production API Key"
  }'
```

### 2. Update Environment Variables

**Backend (.env):**
```bash
# Production PostgreSQL
DATABASE_URL=postgres://user:password@prod-db:5432/frustration_engine?sslmode=require

# Production ClickHouse
CLICKHOUSE_DSN=clickhouse://prod-clickhouse:9000

# Production API Keys
FRUSTRATION_API_KEY=your-production-api-key
```

**Frontend (.env.production):**
```bash
NEXT_PUBLIC_FRUSTRATION_API_KEY=your-production-api-key
NEXT_PUBLIC_FRUSTRATION_INGESTION_URL=https://api.yourdomain.com/v1/events
```

### 3. Configure Ticket Export (Optional)

```bash
# Jira
export JIRA_URL="https://your-domain.atlassian.net"
export JIRA_EMAIL="your-email@example.com"
export JIRA_API_TOKEN="your-token"
export JIRA_PROJECT="PROJ"

# Linear
export LINEAR_URL="https://api.linear.app/graphql"
export LINEAR_KEY="your-key"
export LINEAR_TEAM="your-team-id"
```

---

## 📝 Integration Checklist

- [ ] Services are running and healthy
- [ ] Frontend SDK is installed
- [ ] SDK is initialized in your app
- [ ] Environment variables are set
- [ ] Test event is sent successfully
- [ ] Events appear in logs
- [ ] Incidents are created (after session completion)
- [ ] Production API keys are generated
- [ ] Production URLs are configured
- [ ] Monitoring is set up

---

## 🧪 Testing Your Integration

### 1. Verify SDK is Loaded

Open browser console:
```javascript
// Should see SDK initialization message
// Check network tab for requests to /v1/events
```

### 2. Send Test Event

```bash
curl -X POST http://localhost:8080/v1/events \
  -H "X-API-Key: test-api-key" \
  -H "Content-Type: application/json" \
  -d '{"project_id":"test","events":[]}'
```

### 3. Check Service Health

```bash
curl http://localhost:8080/health
curl http://localhost:8081/health
curl http://localhost:8082/health
curl http://localhost:8084/health
```

### 4. Query Incidents

```bash
# After session completes (~15 minutes)
curl http://localhost:8084/v1/incidents?project_id=test-project
```

---

## 📚 Next Steps

1. **Set up monitoring** - Configure Prometheus/Grafana
2. **Configure alerts** - Set up alerts for high-confidence incidents
3. **Tune thresholds** - Adjust frustration detection based on your data
4. **Set up CI/CD** - Automate deployment
5. **Documentation** - Document your integration for your team

---

## ❓ FAQ

**Q: Do I need to push to Git first?**  
A: No! You can integrate locally first, then push to Git when ready.

**Q: Can I use this without deploying the backend?**  
A: Yes, run services locally and point your frontend to `http://localhost:8080`.

**Q: How do I deploy to production?**  
A: Deploy services to your cloud provider (AWS, GCP, etc.) and update frontend URLs.

**Q: Can I customize the detection logic?**  
A: Yes, modify the UFSE module (`internal/ufse/`) to adjust signal detection.

**Q: How do I get production API keys?**  
A: Generate them through the Incident Store API or admin interface.

---

## 📄 Related Documentation

- **[RUN_AND_INTEGRATE.md](./RUN_AND_INTEGRATE.md)** - Complete run and integration guide
- **[TERMINAL_COMMANDS.md](./TERMINAL_COMMANDS.md)** - Quick command reference
- **[CREDENTIALS.md](./CREDENTIALS.md)** - All credentials reference
- **[DEPLOYMENT.md](./DEPLOYMENT.md)** - Production deployment guide

---

**Ready to integrate! Start with Option 1 (Use as Service) - it's the easiest.** 🚀
