# Credentials & Passwords Reference

All passwords and credentials needed to run the Frustration Engine.

---

## 🔐 Default Credentials (Development/Testing)

### PostgreSQL Database

**Username:** `postgres`  
**Password:** `postgres`  
**Database:** `frustration_engine`  
**Port:** `5432`  
**Connection String:**
```
postgres://postgres:postgres@localhost:5432/frustration_engine?sslmode=disable
```

**Docker Command:**
```bash
docker run -d \
  --name frustration-postgres \
  -p 5432:5432 \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=frustration_engine \
  postgres:15
```

---

### ClickHouse Database

**No password required** (default setup)  
**Port:** `9000` (native), `8123` (HTTP)  
**Connection String:**
```
clickhouse://localhost:9000
```

**Docker Command:**
```bash
docker run -d \
  --name frustration-clickhouse \
  -p 9000:9000 \
  -p 8123:8123 \
  clickhouse/clickhouse-server:latest
```

---

### API Keys

**Test API Key (Development):**
```
test-api-key
```

**Usage:**
```bash
# In HTTP requests
curl -H "X-API-Key: test-api-key" http://localhost:8080/v1/events

# In Frontend SDK
initFrustrationObserver({
  apiKey: 'test-api-key',
  ingestionUrl: 'http://localhost:8080/v1/events',
});
```

**Note:** For production, create your own API keys through the admin API.

---

## 🔒 Production Credentials

### PostgreSQL (Production)

**Change these in production!**

```bash
# Set via environment variable
export POSTGRES_USER=your_username
export POSTGRES_PASSWORD=your_secure_password
export POSTGRES_DB=frustration_engine

# Or in connection string
DATABASE_URL="postgres://your_username:your_secure_password@localhost:5432/frustration_engine?sslmode=disable"
```

---

### Jira Integration (Optional)

**Required for Ticket Exporter (Jira adapter):**

```bash
export JIRA_URL="https://your-domain.atlassian.net"
export JIRA_EMAIL="your-email@example.com"
export JIRA_API_TOKEN="your-jira-api-token"
export JIRA_PROJECT="PROJ"  # Your Jira project key
```

**How to get Jira API Token:**
1. Go to https://id.atlassian.com/manage-profile/security/api-tokens
2. Click "Create API token"
3. Copy the token and use it in `JIRA_API_TOKEN`

---

### Linear Integration (Optional)

**Required for Ticket Exporter (Linear adapter):**

```bash
export LINEAR_URL="https://api.linear.app/graphql"
export LINEAR_KEY="your-linear-api-key"
export LINEAR_TEAM="your-team-id"
```

**How to get Linear API Key:**
1. Go to Linear Settings → API
2. Create a new API key
3. Copy the key and use it in `LINEAR_KEY`

---

## 📝 Environment Variables Summary

### Development (.env file)

```bash
# PostgreSQL
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=frustration_engine
DATABASE_URL=postgres://postgres:postgres@localhost:5432/frustration_engine?sslmode=disable

# ClickHouse
CLICKHOUSE_DSN=clickhouse://localhost:9000

# API Key (for testing)
TEST_API_KEY=test-api-key

# Service URLs
SESSION_MANAGER_URL=http://localhost:8081
UFSE_URL=http://localhost:8082
INCIDENT_STORE_URL=http://localhost:8084

# Ticket Exporter (optional - use "noop" for testing)
ADAPTER=noop
```

### Production (.env file)

```bash
# PostgreSQL (CHANGE THESE!)
POSTGRES_USER=your_production_user
POSTGRES_PASSWORD=your_secure_password_here
DATABASE_URL=postgres://your_production_user:your_secure_password_here@your-db-host:5432/frustration_engine?sslmode=require

# ClickHouse
CLICKHOUSE_DSN=clickhouse://your-clickhouse-host:9000

# API Keys (generate your own)
FRUSTRATION_API_KEY=your-production-api-key-here

# Service URLs (use your domain)
SESSION_MANAGER_URL=https://session-manager.yourdomain.com
UFSE_URL=https://ufse.yourdomain.com
INCIDENT_STORE_URL=https://incident-store.yourdomain.com

# Ticket Exporter
ADAPTER=jira  # or "linear"
JIRA_URL=https://your-domain.atlassian.net
JIRA_EMAIL=your-email@example.com
JIRA_API_TOKEN=your-jira-token
JIRA_PROJECT=PROJ
```

---

## 🔑 Quick Reference

| Credential | Development | Production |
|------------|-------------|------------|
| **PostgreSQL User** | `postgres` | ⚠️ Change this |
| **PostgreSQL Password** | `postgres` | ⚠️ Change this |
| **PostgreSQL DB** | `frustration_engine` | `frustration_engine` |
| **ClickHouse** | No password | Configure as needed |
| **API Key** | `test-api-key` | ⚠️ Generate your own |
| **Jira Token** | Not needed | ⚠️ Required for Jira export |
| **Linear Key** | Not needed | ⚠️ Required for Linear export |

---

## ⚠️ Security Notes

1. **Never commit passwords to git** - Use `.env` files and add them to `.gitignore`
2. **Change default passwords** - The `postgres/postgres` credentials are for development only
3. **Use secrets manager** - In production, use AWS Secrets Manager, HashiCorp Vault, etc.
4. **Rotate credentials** - Regularly rotate API keys and database passwords
5. **Use SSL/TLS** - In production, always use `sslmode=require` for PostgreSQL

---

## 🧪 Testing Without Database

You can run the Incident Store in "log-only" mode (no database required):

```bash
go run ./cmd/incident-store/main.go \
  -port=8084 \
  -dsn="log-only"
```

This will log incidents to console instead of storing them in PostgreSQL.

---

## 📚 Related Documentation

- **[RUN_AND_INTEGRATE.md](./RUN_AND_INTEGRATE.md)** - How to run and integrate
- **[QUICK_COMMANDS.md](./QUICK_COMMANDS.md)** - Quick command reference
- **[DEPLOYMENT.md](./DEPLOYMENT.md)** - Production deployment guide

---

**Default Development Credentials:**
- PostgreSQL: `postgres` / `postgres`
- API Key: `test-api-key`
- ClickHouse: No password

**⚠️ Remember to change these for production!**
