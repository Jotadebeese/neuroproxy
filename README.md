# ⚡ NeuroProxy — Autonomous Self-Healing API Gateway

> **Google Cloud AI Builder Cup 2026 Submission** — Team **Zenith Systems** (Juan Diego Bedoya & Kevin Juan Carlos)  
> **Track:** Future of Work & Enterprise Productivity *(Autonomous Developer Infrastructure & SRE)*  
> **Stack:** Go 1.23 (`net/http`, `sync.RWMutex`), Google Gemini (`gemini-2.5-flash`), Next.js 15 / React 19 / Tailwind CSS v4, Google Cloud Run

> [!IMPORTANT]
> **👋 Team Zenith Systems Onboarding (Start Here):**
> * **Master Execution Plan & Timeline:** See [`PLAN.md`](./PLAN.md)
> * **Official Pitch Deck Slide Copy & Portal Text:** See [`PITCH_DECK_OUTLINE.md`](./PITCH_DECK_OUTLINE.md)
> * **Run Locally in 10 Seconds:** Copy `.env.example` to `.env.local`, add your `GEMINI_API_KEY` from [Google AI Studio](https://aistudio.google.com/apikey), run `go run .`, and open **`http://localhost:8080`**.

---

## 1. The Problem: Schema Drift & 3:00 AM Outages

Modern distributed architectures rely on dozens of external webhooks (Stripe, Shopify, logistics ERPs) and client applications. When an upstream provider renames a JSON field (`amount_cents` $\rightarrow$ `payment.amount_dollars`), nests an object (`customer_email` $\rightarrow$ `buyer.contact.email`), or changes a type/enum without warning:
1. Downstream microservices throw unhandled **`422 Unprocessable Entity` / `500 Internal Server Error`** exceptions.
2. Customer checkouts and critical industrial telemetry drop on the floor.
3. On-call SREs are paged at 3:00 AM to manually inspect logs, write hotfix adapters, and redeploy containers.

---

## 2. The Solution: Autonomous Runtime Self-Healing

**NeuroProxy** is an ultra-light reverse proxy deployed on **Google Cloud Run** that sits in front of backend microservices and eliminates schema-drift downtime autonomously:

```mermaid
flowchart LR
    Client["External Webhook / Client"] -->|"1. Drifted JSON Payload"| Proxy["⚡ NeuroProxy Gateway (Go)"]
    Proxy -->|"2. Forward Request"| Upstream["Target Microservice"]
    Upstream -->|"3. 422 / 500 Schema Error"| Proxy
    Proxy -->|"4. Check SHA-256 Fingerprint"| Cache[("In-Memory RuleCache\n(sync.RWMutex)")]
    Cache -->|"Cache Hit (<1ms)"| Healed["Apply Learned Adapter"]
    Cache -->|"Cache Miss"| Gemini["Google Gemini 2.5 Flash"]
    Gemini -->|"5. Synthesize Patch + Rule + GitHub PR"| Cache
    Gemini --> Healed
    Healed -->|"6. Re-Dispatch Repaired JSON"| Upstream
    Upstream -->|"7. 200 OK (X-Healed-By: NeuroProxy)"| Client
```

### Key Engineering Capabilities
* **Real-Time Interception:** Holds the client HTTP connection open when a downstream microservice rejects a payload (`400`/`422`/`500`).
* **Gemini Flash Runtime Synthesis:** Passes the broken payload, strict upstream contract, and error trace to **Google Gemini 2.5 Flash** (`responseMimeType: "application/json"`), synthesizing both the repaired JSON payload and a deterministic field-mapping rule.
* **Sub-Millisecond In-Memory Rule Cache (`0ms` LLM Cost on Repeat Traffic):** Computes a structural `SHA-256` fingerprint of the incoming JSON keys and types. Once Gemini heals a drifted schema once, the learned transformation rule is stored in a thread-safe `sync.RWMutex` cache—healing 10,000+ subsequent requests in **`<1ms` without calling the LLM**.
* **Autonomous GitHub Pull Request Generator:** Generates a ready-to-merge unified Git diff (`--- a/internal/handlers/...`) so developers can permanently commit the schema adapter at their convenience.

---

## 3. Repository Structure

```text
neuroproxy/
├── main.go                        # Unified Go HTTP server (Reverse Proxy API + Static UI host)
├── go.mod                         # Zero-bloat Go standard library module
├── Dockerfile                     # Multi-stage Google Cloud Run container build
├── PLAN.md                        # AI Builder Cup 2026 Execution Plan, Demo Script & Slide Outline
├── internal/
│   └── proxy/
│       ├── types.go               # Domain contracts, telemetry metrics & incident models
│       ├── upstreams.go           # Strict target microservices (Checkout, ERP Inventory, IoT Telemetry)
│       ├── cache.go               # Thread-safe sync.RWMutex RuleCache & SHA-256 schema fingerprinter
│       ├── healer.go              # Google Gemini 2.5 Flash integration + deterministic fallback synthesizer
│       ├── engine.go              # Self-healing proxy pipeline coordinator
│       └── engine_test.go         # End-to-end healing & concurrent race-detector tests
└── frontend/                      # Next.js 15 + React 19 + TypeScript + Tailwind v4 Mission Control UI
    └── out/                       # Pre-compiled static bundle served directly by the Go binary at `/`
```

---

## 4. Quickstart (Run Locally in 5 Seconds)

### Prerequisites
* **Go 1.22+**
* *(Optional)* `GEMINI_API_KEY` environment variable for live Google AI Studio calls (includes a deterministic structural synthesis fallback if unset).

### 1. Run Unit & Concurrency Tests
```powershell
go test -v -race ./...
```

### 2. Start the NeuroProxy Gateway & Mission Control UI
```powershell
$env:GEMINI_API_KEY="your-gemini-api-key"
go run .
```
Open **`http://localhost:8080`** in your browser to launch the interactive **NeuroProxy Mission Control & Chaos Simulator**.

### 3. Test via Terminal (`curl`)
```bash
# 1. Send a drifted Stripe/Shopify webhook with Self-Healing ON (returns 200 OK + X-Healed-By header)
curl -i -X POST http://localhost:8080/api/proxy/checkout \
  -H "Content-Type: application/json" \
  -d '{"id":"ORD-991","payment":{"amount_dollars":"149.50","iso_currency":"usd"},"buyer":{"contact":{"email":"sre@googlecloud.com"}}}'

# 2. Send the same drifted webhook in Bypass Mode (returns 422 crash)
curl -i -X POST "http://localhost:8080/api/proxy/checkout?bypass=true" \
  -H "Content-Type: application/json" \
  -d '{"id":"ORD-991","payment":{"amount_dollars":"149.50","iso_currency":"usd"},"buyer":{"contact":{"email":"sre@googlecloud.com"}}}'
```

---

## 5. Deploy to Google Cloud Run (1 Command)

```bash
gcloud run deploy neuroproxy \
  --source . \
  --region asia-southeast1 \
  --allow-unauthenticated \
  --set-env-vars="GEMINI_API_KEY=YOUR_GEMINI_API_KEY"
```
