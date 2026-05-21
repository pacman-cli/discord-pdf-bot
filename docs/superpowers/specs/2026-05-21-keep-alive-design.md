# Keep-Alive Design

**Date:** 2026-05-21
**Status:** Approved

## Problem

Render free tier spins down after ~15 minutes of inactivity. The Discord bot maintains a WebSocket connection (gateway), which doesn't count as HTTP activity. Users experience 50+ second delays when the service spins back up.

## Solution

Add a lightweight HTTP server with a `/health` endpoint and a self-ping goroutine to keep Render alive.

## Design

### Components

**HTTP Server** (in `cmd/bot/main.go`)
- `GET /health` returns `200 OK` with `{"status":"ok"}`
- Listens on `PORT` env var (Render sets this automatically, fallback to `8080`)
- Starts before Discord bot connection

**Self-Ping Goroutine**
- Hits `http://localhost:{PORT}/health` every 10 minutes
- Logs errors but doesn't crash (non-critical)
- Starts after HTTP server is ready

### Data Flow

1. Bot starts
2. HTTP server starts on `PORT`
3. Discord bot connects
4. Self-ping goroutine runs every 10 min
5. Render sees HTTP traffic, keeps service awake

### Files

| File | Action |
|------|--------|
| `cmd/bot/main.go` | Modify — add HTTP server and self-ping |

### Dependencies

Only Go stdlib: `net/http`, `time`, `encoding/json`
