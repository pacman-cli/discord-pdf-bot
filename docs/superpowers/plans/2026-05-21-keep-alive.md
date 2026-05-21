# Keep-Alive Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Prevent Render free tier from spinning down by adding an HTTP health endpoint and self-ping goroutine.

**Architecture:** Add `net/http` server with `/health` endpoint to `cmd/bot/main.go`. A goroutine pings `localhost:{PORT}/health` every 10 minutes. PORT defaults to `8080` if not set.

**Tech Stack:** Go stdlib (`net/http`, `time`, `encoding/json`)

---

## File Structure

| File | Action | Purpose |
|------|--------|---------|
| `cmd/bot/main.go` | Modify | Add HTTP server and self-ping goroutine |

---

### Task 1: Add HTTP health server and self-ping

**Files:**
- Modify: `cmd/bot/main.go`

- [ ] **Step 1: Add keep-alive HTTP server and self-ping to main.go**

Add these imports to the import block:

```go
"encoding/json"
"net/http"
"time"
```

Add this function after `main()`:

```go
func startKeepAlive(port string) {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	go func() {
		slog.Info("Health server starting", "port", port)
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			slog.Error("Health server failed", "error", err)
		}
	}()

	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			resp, err := http.Get("http://localhost:" + port + "/health")
			if err != nil {
				slog.Error("Keep-alive ping failed", "error", err)
				continue
			}
			resp.Body.Close()
			slog.Debug("Keep-alive ping sent")
		}
	}()
}
```

Then add this call in `main()`, right after the `adminRole` setup (after line 37):

```go
// Keep-alive for Render free tier
port := os.Getenv("PORT")
if port == "" {
	port = "8080"
}
startKeepAlive(port)
```

- [ ] **Step 2: Build to verify**

Run: `go build ./...`
Expected: no output (success)

- [ ] **Step 3: Commit**

```bash
git add cmd/bot/main.go
git commit -m "feat: add keep-alive health endpoint for Render free tier"
```

---

### Task 2: Verify and push

- [ ] **Step 1: Run all tests**

Run: `go test ./... -v`
Expected: all 12 tests pass

- [ ] **Step 2: Push**

```bash
git push
```
