# Frustration Engine

A microservices-based system for detecting and tracking user frustration signals in web applications.

## Architecture

- **Event Ingestion API** (Port 8080) - Receives and stores user events
- **Session Manager** (Port 8081) - Manages user sessions
- **UFSE** (Port 8082) - User Frustration Signal Engine
- **Incident Store** (Port 8084) - Stores detected incidents
- **Ticket Exporter** (Port 8085) - Exports incidents to Jira/Linear

## Quick Start

```bash
# Start databases
./scripts/start_databases.sh

# Start services
export POSTGRES_PORT=5434 && export CLICKHOUSE_PORT=9001
./scripts/start_services_custom_port.sh
```

## Documentation

- [RUN_AND_INTEGRATE.md](./RUN_AND_INTEGRATE.md) - Complete setup guide
- [INTEGRATION_STEPS.md](./INTEGRATION_STEPS.md) - Integration guide
- [SIMPLE_COMMANDS.md](./SIMPLE_COMMANDS.md) - Quick command reference
- [TERMINAL_COMMANDS.md](./TERMINAL_COMMANDS.md) - All commands

## Tech Stack

- **Backend:** Go 1.21+
- **Databases:** PostgreSQL, ClickHouse
- **Frontend SDK:** JavaScript/TypeScript

## License

MIT
