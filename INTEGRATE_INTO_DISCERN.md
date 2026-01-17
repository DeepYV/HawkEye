# How to Integrate Frustration Engine into Your "discern" Project

Complete guide for integrating this project into your existing codebase.

---

## 🎯 Recommended Approach: Keep Separate (Microservices)

**Best for:** Production, scalability, independent deployment

Keep Frustration Engine as a **separate service** and integrate via API/SDK.

### Why This is Best:
- ✅ Clean separation of concerns
- ✅ Independent deployment
- ✅ Easy to update/scale
- ✅ No code conflicts
- ✅ Works with any tech stack

### How It Works:
1. **Frustration Engine** runs as separate services (already set up)
2. **Your "discern" app** uses the Frontend SDK to send events
3. Services communicate via HTTP API

---

## 📦 Option 1: Keep Separate + Use SDK (Recommended)

### Step 1: Keep Frustration Engine Separate

**Don't move it into discern directory!** Keep it where it is:
```
/Users/deepakyadav/ucfp/          ← Frustration Engine (keep here)
/Users/deepakyadav/discern/       ← Your app (keep separate)
```

### Step 2: Start Frustration Engine Services

```bash
cd /Users/deepakyadav/ucfp

# Start databases
./scripts/start_databases.sh

# Start services
export POSTGRES_PORT=5434
export CLICKHOUSE_PORT=9001
./scripts/start_services_custom_port.sh
```

### Step 3: Integrate SDK in Your "discern" App

```bash
cd /Users/deepakyadav/discern

# Install SDK (if published to npm)
npm install @frustration-engine/observer-sdk

# Or copy SDK code if not published yet
```

**In your discern app (React/Next.js):**

```typescript
// app/layout.tsx or pages/_app.tsx
'use client';

import { useEffect } from 'react';
import { initFrustrationObserver } from '@frustration-engine/observer-sdk';

export default function RootLayout({ children }) {
  useEffect(() => {
    initFrustrationObserver({
      apiKey: process.env.NEXT_PUBLIC_FRUSTRATION_API_KEY || 'test-api-key',
      ingestionUrl: process.env.NEXT_PUBLIC_FRUSTRATION_INGESTION_URL || 'http://localhost:8080/v1/events',
    });
  }, []);

  return <html>{children}</html>;
}
```

**Environment variables (.env.local):**
```bash
NEXT_PUBLIC_FRUSTRATION_API_KEY=test-api-key
NEXT_PUBLIC_FRUSTRATION_INGESTION_URL=http://localhost:8080/v1/events
```

### Step 4: Git Setup (Separate Repos)

```bash
# Frustration Engine repo
cd /Users/deepakyadav/ucfp
git init
git add .
git commit -m "Initial commit: Frustration Engine"
git remote add origin https://github.com/your-username/frustration-engine.git
git push -u origin main

# Your discern repo (no changes needed)
cd /Users/deepakyadav/discern
# Just add the SDK integration code
```

**Result:** Two separate repos, integrated via API.

---

## 📁 Option 2: Add as Git Submodule (Keep Code Separate)

If you want Frustration Engine code in your discern repo but keep it separate:

### Step 1: Initialize Frustration Engine Git

```bash
cd /Users/deepakyadav/ucfp
git init
git add .
git commit -m "Initial commit"
git remote add origin https://github.com/your-username/frustration-engine.git
git push -u origin main
```

### Step 2: Add as Submodule in discern

```bash
cd /Users/deepakyadav/discern
git submodule add https://github.com/your-username/frustration-engine.git frustration-engine
git commit -m "Add Frustration Engine as submodule"
```

**Result:**
```
discern/
  ├── src/
  ├── frustration-engine/    ← Submodule (separate repo)
  └── ...
```

**To update:**
```bash
cd /Users/deepakyadav/discern/frustration-engine
git pull origin main
cd ..
git add frustration-engine
git commit -m "Update Frustration Engine"
```

---

## 📋 Option 3: Copy Code into discern (Monorepo)

If you want everything in one repo:

### Step 1: Copy Frustration Engine into discern

```bash
cd /Users/deepakyadav/discern

# Create a services directory
mkdir -p services

# Copy Frustration Engine
cp -r /Users/deepakyadav/ucfp services/frustration-engine

# Or use rsync to exclude .git
rsync -av --exclude='.git' /Users/deepakyadav/ucfp/ services/frustration-engine/
```

**Result:**
```
discern/
  ├── src/                    ← Your app code
  ├── services/
  │   └── frustration-engine/ ← Frustration Engine code
  └── ...
```

### Step 2: Update Paths in Frustration Engine

If you copied it, you might need to update import paths if they're absolute.

### Step 3: Add to Git

```bash
cd /Users/deepakyadav/discern
git add services/frustration-engine
git commit -m "Add Frustration Engine service"
```

**Note:** This makes your repo larger, but everything is in one place.

---

## 🔄 Option 4: Git Subtree (Merge History)

If you want to merge Frustration Engine into discern with history:

```bash
cd /Users/deepakyadav/discern

# First, push Frustration Engine to its own repo
cd /Users/deepakyadav/ucfp
git init
git add .
git commit -m "Initial commit"
git remote add origin https://github.com/your-username/frustration-engine.git
git push -u origin main

# Then add as subtree in discern
cd /Users/deepakyadav/discern
git subtree add --prefix=services/frustration-engine https://github.com/your-username/frustration-engine.git main --squash
```

**To update later:**
```bash
git subtree pull --prefix=services/frustration-engine https://github.com/your-username/frustration-engine.git main --squash
```

---

## 🚀 Quick Start: Recommended Setup

### For Development:

```bash
# Terminal 1: Start Frustration Engine
cd /Users/deepakyadav/ucfp
./scripts/start_databases.sh
export POSTGRES_PORT=5434 && export CLICKHOUSE_PORT=9001
./scripts/start_services_custom_port.sh

# Terminal 2: Run your discern app
cd /Users/deepakyadav/discern
npm run dev  # or your start command
```

### For Production:

1. **Deploy Frustration Engine** to your cloud (AWS, GCP, etc.)
2. **Update SDK URL** in your discern app:
   ```typescript
   ingestionUrl: 'https://api.yourdomain.com/v1/events'
   ```
3. **Deploy discern app** as usual

---

## 📝 Integration Checklist

- [ ] Decide on integration approach (recommended: Option 1)
- [ ] Start Frustration Engine services
- [ ] Install/copy Frontend SDK into discern
- [ ] Initialize SDK in your app
- [ ] Set environment variables
- [ ] Test event sending
- [ ] Push to Git (if using separate repos)
- [ ] Deploy to production

---

## 🎯 My Recommendation

**Use Option 1 (Keep Separate + SDK):**

1. ✅ Clean architecture
2. ✅ Easy to maintain
3. ✅ Independent deployment
4. ✅ No code conflicts
5. ✅ Works with any framework

**Steps:**
1. Keep `/Users/deepakyadav/ucfp` as separate project
2. Push it to its own Git repo
3. In your `discern` app, just use the SDK
4. Services communicate via HTTP API

**No need to move code or merge repos!**

---

## 📚 Next Steps

1. **Choose your approach** (I recommend Option 1)
2. **Set up Git** for Frustration Engine (if keeping separate)
3. **Integrate SDK** in your discern app
4. **Test integration**
5. **Deploy**

---

## ❓ FAQ

**Q: Do I need to move this project into discern?**  
A: No! Keep it separate and use the SDK.

**Q: Should I push this to Git?**  
A: Yes, push Frustration Engine to its own repo. Your discern repo stays separate.

**Q: Can I use this without Git?**  
A: Yes, but Git is recommended for version control.

**Q: What if I want everything in one repo?**  
A: Use Option 3 (Copy Code) or Option 4 (Git Subtree).

---

**Ready to integrate? Start with Option 1 - it's the cleanest!** 🚀
