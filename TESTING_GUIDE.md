# End-to-End Testing Guide

**Date:** 2024-01-15  
**Prepared By:** John Smith (Solution Architect)

---

## Quick Start

All integrations are complete with logging. You can test the entire system end-to-end.

---

## Service Ports

| Service | Port | URL |
|---------|------|-----|
| Event Ingestion API | 8080 | http://localhost:8080 |
| Session Manager | 8081 | http://localhost:8081 |
| UFSE | 8082 | http://localhost:8082 |
| Ticket Exporter | 8083 | http://localhost:8083 |
| Incident Store | 8084 | http://localhost:8084 |

---

## Start All Services

### Terminal 1: Incident Store
```bash
cd module5_implementation
go run cmd/incident-store/main.go \
  --port=8084 \
  --dsn="log-only"
```

**Note:** Using `--dsn="log-only"` runs in logging mode without PostgreSQL. For production, use actual PostgreSQL DSN:
```bash
--dsn="postgres://user:password@localhost/incidents?sslmode=disable"
```

### Terminal 2: UFSE
```bash
cd module4_implementation
go run cmd/ufse/main.go \
  --port=8082 \
  --incident-store-url=http://localhost:8084
```

### Terminal 3: Session Manager
```bash
cd module3_implementation
go run cmd/session-manager/main.go \
  --port=8081 \
  --ufse-url=http://localhost:8082
```

### Terminal 4: Event Ingestion API
```bash
cd module2_implementation
go run cmd/api/main.go \
  --port=8080 \
  --session-manager=http://localhost:8081
```

### Terminal 5: Ticket Exporter
```bash
cd module7_implementation
go run cmd/ticket-exporter/main.go \
  --port=8083 \
  --incident-store-url=http://localhost:8084 \
  --export-interval=1m \
  --max-per-interval=10
```

---

## Test Data Flow

### 1. Send Test Events

```bash
curl -X POST http://localhost:8080/v1/events \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-key-123" \
  -d '{
    "events": [
      {
        "eventType": "click",
        "timestamp": "2024-01-15T10:00:00Z",
        "sessionId": "test-session-1",
        "route": "/checkout",
        "target": {"type": "button", "id": "submit-btn"},
        "metadata": {}
      },
      {
        "eventType": "click",
        "timestamp": "2024-01-15T10:00:01Z",
        "sessionId": "test-session-1",
        "route": "/checkout",
        "target": {"type": "button", "id": "submit-btn"},
        "metadata": {}
      },
      {
        "eventType": "click",
        "timestamp": "2024-01-15T10:00:02Z",
        "sessionId": "test-session-1",
        "route": "/checkout",
        "target": {"type": "button", "id": "submit-btn"},
        "metadata": {}
      },
      {
        "eventType": "error",
        "timestamp": "2024-01-15T10:00:03Z",
        "sessionId": "test-session-1",
        "route": "/checkout",
        "target": {"type": "javascript_error"},
        "metadata": {"error": "Validation failed"}
      }
    ]
  }'
```

### 2. Monitor Logs

Watch each service's console output:

**Event Ingestion API:**
```
[Event Ingestion] Successfully forwarded 4 events to Session Manager (session: test-session-1)
```

**Session Manager:**
```
[Session Manager] Successfully forwarded session test-session-1 to UFSE, 1 incidents detected
```

**UFSE:**
```
[UFSE] Successfully forwarded incident abc-123 to Incident Store
```

**Incident Store:**
```
[Incident Store] Stored incident abc-123
```

**Ticket Exporter:**
```
[Ticket Exporter] Retrieved 1 eligible incidents from Incident Store
```

---

## Expected Behavior

### With All Services Running:
1. Events sent to Ingestion API
2. Events forwarded to Session Manager
3. Session Manager groups events into session
4. Completed session forwarded to UFSE
5. UFSE detects incidents (if signals match)
6. Incidents forwarded to Incident Store
7. Ticket Exporter queries Incident Store
8. Tickets formatted (but not created if Jira/Linear not configured)

### If Services Not Running:
- System logs data instead of failing
- You can see what would be sent
- No data loss, just logging

---

## Verification Steps

### 1. Check Event Ingestion
```bash
# Should return success
curl -X POST http://localhost:8080/v1/events \
  -H "X-API-Key: test-key" \
  -H "Content-Type: application/json" \
  -d '{"events": [...]}'
```

### 2. Check Session Manager
```bash
# Check health
curl http://localhost:8081/health
```

### 3. Check UFSE
```bash
# Check health
curl http://localhost:8082/health
```

### 4. Check Incident Store
```bash
# Query incidents
curl "http://localhost:8084/v1/incidents?status=confirmed&limit=10"
```

### 5. Check Ticket Exporter
```bash
# Trigger export manually
curl -X POST http://localhost:8083/v1/export/trigger
```

---

## Troubleshooting

### Issue: Events not forwarded
**Check:**
- Session Manager URL configured in Event Ingestion API
- Session Manager is running
- Check logs for errors

### Issue: Sessions not forwarded
**Check:**
- UFSE URL configured in Session Manager
- UFSE is running
- Check logs for errors

### Issue: Incidents not stored
**Check:**
- Incident Store URL configured in UFSE
- Incident Store is running
- PostgreSQL connection (if using DB)
- Check logs for errors

### Issue: No tickets queried
**Check:**
- Incident Store URL configured in Ticket Exporter
- Incidents are confirmed (status=confirmed)
- Check logs for query results

---

## Logging Format

All services use consistent logging:
- `[Service Name]` prefix
- Action description
- Success/failure status
- Relevant IDs (session, incident, etc.)

Example:
```
[Event Ingestion] Successfully forwarded 4 events to Session Manager (session: test-session-1)
[Session Manager] Successfully forwarded session test-session-1 to UFSE, 1 incidents detected
[UFSE] Successfully forwarded incident abc-123 to Incident Store
```

---

## Success Criteria

✅ Events flow through all modules  
✅ Sessions are created and completed  
✅ Incidents are detected  
✅ Incidents are stored  
✅ Ticket Exporter queries work  
✅ All steps logged clearly  
✅ No silent failures  

---

**Status:** ✅ **READY FOR TESTING**

All integrations complete. Start services and test!