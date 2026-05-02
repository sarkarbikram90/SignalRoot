# SignalRoot
# Incident Intelligence Platform

## Architecture

SignalRoot is composed of the following services:

| Service | Language | Port | Description |
|---------|----------|------|-------------|
| Gateway | Go | 8080 | Ingestion webhook receiver |
| API | Go | 8081 | REST API server |
| Worker | Go | — | Background job processor |
| ML | Python | 8082 | Embedding & similarity service |
| Web | React/TS | 3000 | Frontend dashboard |

## Quick Start

### Prerequisites
- Docker & Docker Compose
- Go 1.23+
- Node.js 20+
- Python 3.12+

### 1. Start infrastructure
```bash
docker compose up -d postgres redis redpanda qdrant
```

### 2. Run database migrations
```bash
# Install goose
go install github.com/pressly/goose/v3/cmd/goose@latest

# Run migrations
goose -dir migrations postgres "postgres://signalroot:signalroot@localhost:5432/signalroot?sslmode=disable" up
```

### 3. Start the backend
```bash
# Terminal 1: Gateway
go run ./cmd/gateway

# Terminal 2: API server
go run ./cmd/api

# Terminal 3: Worker
go run ./cmd/worker
```

### 4. Start the ML service
```bash
cd ml
pip install -r requirements.txt
uvicorn serve:app --host 0.0.0.0 --port 8082 --reload
```

### 5. Start the frontend
```bash
cd web
npm install
npm run dev
```

### 6. Open the dashboard
Navigate to http://localhost:3000

## Full Stack (Docker)
```bash
docker compose up --build
```

## Project Structure
```
signalroot/
├── cmd/                 # Go service binaries
│   ├── gateway/         # Ingestion gateway
│   ├── api/             # REST API server
│   └── worker/          # Background worker
├── internal/            # Go internal packages
│   ├── auth/            # JWT, API keys, RBAC
│   ├── config/          # Configuration
│   ├── correlation/     # Signal-to-incident correlation
│   ├── db/              # Database & repositories
│   └── incident/        # Domain models
├── ml/                  # Python ML service
├── web/                 # React frontend
├── migrations/          # SQL migrations (goose)
├── integrations/        # Source adapters
├── docker/              # Dockerfiles
└── docker-compose.yml   # Local dev stack
```

## MVP Features
- ✅ PagerDuty + Slack webhook ingestion
- ✅ Automatic signal-to-incident correlation
- ✅ Incident status machine (open → acknowledged → investigating → mitigated → resolved → closed)
- ✅ AI-generated incident summaries
- ✅ Incident similarity search via vector embeddings
- ✅ SOC2 compliance report generation
- ✅ Dashboard with MTTA/MTTR metrics
- ✅ Slack notifications for critical incidents
- ✅ API key authentication
- ✅ Multi-tenant data isolation

## Tech Stack
- **Backend**: Go (chi router, pgx, zap)
- **ML**: Python (FastAPI, sentence-transformers, Qdrant)
- **Frontend**: React + TypeScript + TailwindCSS v4
- **Database**: PostgreSQL 16
- **Vector DB**: Qdrant
- **Queue**: Redpanda (Kafka-compatible)
- **Cache**: Redis 7

## One-Line Pitch
> SignalRoot turns your incident history into institutional intelligence — so your team resolves issues faster, your system gets smarter with every outage, and your auditors stop asking questions.
