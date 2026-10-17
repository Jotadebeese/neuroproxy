# NeuroProxy — AI Builder Cup 2026 Master Plan & Execution Blueprint

> **Team Name:** `Zenith Systems`  
> **Competition:** Google Cloud AI Builder Cup 2026 (JAPAC / Hack2skill)  
> **Challenge Track:** Future of Work & Enterprise Productivity *(Autonomous Developer Infrastructure & SRE)*  
> **Hard Submission Deadline:** Sunday, October 18, 2026 @ 11:59 PM IST (**Monday, Oct 19 @ 1:29 AM ICT Thailand Time**)  
> **Internal Safe Target:** **Sunday, October 18, 2026 @ 7:30 PM ICT (6:00 PM IST)**

---

## 1. Executive Review: Why `NeuroProxy` Wins

### What Evaluators & Google Cloud Judges Look For
From the official Hack2skill & Google Cloud *Ask the Expert* guidelines, every valid submission is judged on **5 non-negotiable pillars**:
1. **Clear Problem & Real-World Impact:** Does it solve an expensive, painful problem?
2. **Google AI Integration:** Does it meaningfully use **Google Gemini** (`gemini-2.5-flash` / `gemini-2.0-flash`) or an agentic workflow (built with **Antigravity**)?
3. **Working Prototype (No Slideware):** Can judges open the link in an incognito window and see it actually working?
4. **Google Cloud Platform Deployment:** Is it deployed live on **Google Cloud Run** (or Firebase)?
5. **Complete 5-Part Submission Package:** Live URL + Demo Video + Prescribed Pitch Deck + Public GitHub Repo + Brief Description.

### Our Strategic Edge
95% of hackathon teams build generic chat wrappers ("chat with a PDF"). **NeuroProxy** speaks directly to Google Cloud's core DNA (**SRE, Cloud Run resilience, and zero-downtime distributed systems**):
* **The Problem:** Third-party webhook schema drift (Stripe, Shopify, logistics partners changing payload structures) and legacy client payload mismatches cause unhandled `400`/`422`/`500` runtime failures, broken checkouts, and 3:00 AM PagerDuty alerts.
* **The NeuroProxy Solution:** An ultra-light **Go self-healing reverse proxy** sitting in front of microservices on **Google Cloud Run**:
  1. **Intercepts Failures:** Catches failing upstream `4xx`/`5xx` schema/validation errors in real time without dropping the client connection.
  2. **Gemini Flash Runtime Synthesis:** Sends the broken payload + upstream target contract + error trace to **Gemini Flash**, synthesizing a repaired JSON payload and a deterministic translation rule in milliseconds.
  3. **Zero-Latency In-Memory Rule Cache:** Fingerprints the broken payload's structural schema (`SHA-256`) and caches the transformation rule in a thread-safe Go `sync.RWMutex` store. **All subsequent identical requests heal in `<1ms` at `$0` LLM cost.**
  4. **Autonomous GitHub PR Generation:** Generates a ready-to-merge code diff / adapter patch for the engineering team so they can permanently fix the schema drift at their convenience.

---

## 2. System Architecture (Single-Container Cloud Run Strategy)

To guarantee **zero CORS issues, zero multi-service networking bugs, and 100% uptime** when judges test the link in an incognito browser window, NeuroProxy is architected as a **single unified Google Cloud Run container**:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                   GOOGLE CLOUD RUN CONTAINER (PORT 8080)                    │
│                                                                             │
│  ┌───────────────────────────────────────────────────────────────────────┐  │
│  │ 1. Mission Control UI (Next.js 15 Static Export served at `/`)        │  │
│  │    - Interactive Webhook & Chaos Simulator (Bypass vs Self-Heal mode) │  │
│  │    - Real-Time Interception & Telemetry Pipeline Visualizer           │  │
│  │    - Side-by-Side JSON Diff, Rule Cache & Autonomous GitHub PR Viewer │  │
│  └───────────────────────────────────┬───────────────────────────────────┘  │
│                                      │ HTTP / JSON                          │
│                                      ▼                                      │
│  ┌───────────────────────────────────────────────────────────────────────┐  │
│  │ 2. NeuroProxy Core Engine (Go 1.22+ `net/http`)                       │  │
│  │    - `/api/proxy/{service}` : Reverse Proxy & Error Interceptor       │  │
│  │    - `RuleCache` (`sync.RWMutex`): Sub-millisecond SHA-256 schema map │  │
│  │    - `GeminiHealer`: Google Gemini Flash structured JSON synthesis    │  │
│  └─────────────────┬───────────────────────────────────┬─────────────────┘  │
│                    │ 1. Try Upstream                   │ 2. On 4xx/5xx Miss │
│                    ▼                                   ▼                    │
│  ┌───────────────────────────────────┐   ┌───────────────────────────────┐  │
│  │ 3. Target Microservices           │   │ Google Gemini Flash API       │  │
│  │    - `/internal/upstream/checkout`│   │ (`gemini-2.5-flash`)          │  │
│  │    - `/internal/upstream/inventory│   │ - Synthesizes healed payload  │  │
│  │    - `/internal/upstream/iot`     │   │ - Extracts reusable rule      │  │
│  └───────────────────────────────────┘   │ - Generates GitHub PR diff    │  │
│                                          └───────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 3. Step-by-Step Weekend Execution Plan

### Phase 1: Core Engine & UI Build (Saturday Afternoon — Today)
- [x] **Step 1.1:** Confirm team on Hack2skill (`Zenith Systems`).
- [ ] **Step 1.2:** Build the **Go Self-Healing Reverse Proxy Engine** (`main.go`, `proxy.go`, `healer.go`, `cache.go`, `upstreams.go`):
  - Strict simulated downstream microservices (`checkout`, `inventory`, `telemetry`).
  - Structural schema fingerprinting (`SHA-256` of key paths & value types).
  - Thread-safe `sync.RWMutex` rule cache for `<1ms` repeat healing.
  - Google Gemini (`gemini-2.5-flash` / `gemini-2.0-flash`) REST integration with deterministic structural fallback so demos never fail even under API rate limits.
- [ ] **Step 1.3:** Build the **Mission Control Dashboard** (`frontend/` — Next.js 15, React 19, TypeScript, Tailwind CSS v4) with signature brand styling (`#eeeeee` / `#181411` / `#ff9900` Warm Orange, `DM Sans` + `Roboto` + `JetBrains Mono`).
- [ ] **Step 1.4:** Verify end-to-end local flow (`Bypass Mode -> 422 Crash` vs `NeuroProxy Mode -> Gemini Heal -> 200 OK` vs `Repeat Request -> <1ms Cache Hit`).

### Phase 2: Containerization & Google Cloud Run Deployment (Saturday Evening)
- [ ] **Step 2.1:** Build multi-stage `Dockerfile` (builds Next.js static assets + compiles stripped Go binary into a single lightweight container).
- [ ] **Step 2.2:** Push to **Google Cloud Run** (`gcloud run deploy neuroproxy --source . --allow-unauthenticated --set-env-vars GEMINI_API_KEY=...`).
- [ ] **Step 2.3:** Open the live `.a.run.app` URL in an **Incognito Window** and verify all scenarios work publicly.

### Phase 3: Demo Video, Pitch Deck & GitHub Repo (Sunday Morning / Afternoon)
- [ ] **Step 3.1:** Push clean code + `README.md` + `PLAN.md` + `Dockerfile` to a **Public GitHub Repository**.
- [ ] **Step 3.2:** Fill out the **Official Hack2skill Pitch Deck Template**.
- [ ] **Step 3.3:** Record the **2-Minute Screen Demo Video** (script below).
- [ ] **Step 3.4:** Submit all 5 items on Hack2skill by **Sunday, Oct 18 @ 7:30 PM ICT** (6 hours ahead of the final deadline).

---

## 4. Workstream Breakdown

| Workstream | Deliverable |
|---|---|
| **Go Proxy Engine, Gemini Integration & Cloud Run Deployment** | Live Cloud Run URL (`https://neuroproxy-...a.run.app`) & Public GitHub Repo |
| **Mission Control Dashboard UI & Demo Recording** | Next.js 15 Dashboard + 2-Minute Demo Video |
| **Pitch Deck Slides, Architecture Polish & QA Testing** | Prescribed Hack2skill Pitch Deck PDF/PPTX + Incognito Stress Testing |

---

## 5. The Winning 2-Minute Demo Video Script

* **0:00 – 0:20 | The Hook & The Problem:**
  * *"Every engineering team has been woken up at 3 AM because Stripe, Shopify, or a mobile client changed a webhook JSON schema and crashed a production microservice with a 422 or 500 error. Today we’re showing **NeuroProxy**—an autonomous self-healing API gateway running on Google Cloud Run and powered by Gemini Flash."*
* **0:20 – 0:45 | The Crash (Bypass Mode):**
  * Show the **NeuroProxy Mission Control** dashboard live on Google Cloud Run.
  * Select **Scenario 1: Stripe/E-Commerce Checkout Webhook Drift** (where `amount_cents` drifted to `total.amount` and `customer_email` moved to `buyer.email`).
  * Click **"Send Direct (Bypass Proxy)"**.
  * Point out the red `422 Unprocessable Entity` crash returned by the strict Checkout microservice.
* **0:45 – 1:20 | The Magic (Gemini Self-Healing):**
  * Toggle **"NeuroProxy Self-Healing: ACTIVE"** and click **"Dispatch Webhook"**.
  * Watch the live pipeline trace light up:
    1. NeuroProxy catches the `422` failure from upstream.
    2. Passes the broken payload + expected schema contract to **Gemini Flash**.
    3. Gemini synthesizes the repaired payload and field-mapping rule in ~140ms.
    4. NeuroProxy re-dispatches the healed payload and returns a clean **`200 OK`** (`X-Healed-By: NeuroProxy-Gemini`)—**the customer checkout succeeds with zero downtime!**
* **1:20 – 1:45 | Zero-Latency In-Memory Caching (`<1ms`):**
  * Click **"Dispatch Webhook"** a second time with the same drifted schema.
  * Show that NeuroProxy's in-memory `SHA-256` schema fingerprint cache intercepts and heals the request in **`<1ms` without calling the LLM** (`X-Healed-By: NeuroProxy-Cache`).
* **1:45 – 2:00 | Autonomous GitHub PR & Close:**
  * Click the **Incident Inspector** tab on the right. Show the side-by-side JSON diff and the **auto-generated GitHub Pull Request diff** that NeuroProxy prepares for the engineering team to permanently patch their codebase.
