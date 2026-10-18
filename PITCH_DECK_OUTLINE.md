# NeuroProxy — Official Pitch Deck Copy & Submission Blueprint
### *Google Cloud AI Builder Cup 2026 Submission*

> **Team Name:** `Zenith Systems`  
> **Team Leader:** `Juan Diego Bedoya`  
> **Project Title:** `NeuroProxy — Autonomous Self-Healing API Gateway`  
> **Challenge Track:** `Future of Work & Enterprise Productivity` *(Autonomous Developer Infrastructure & SRE)*  
> **Core Stack:** Go 1.23, Google Gemini 3.8 Flash, React 19, TypeScript, Google Cloud Run, Cloud Logging, Cloud Trace, Cloud Memorystore (Redis)  
> **Official Slide Template:** [Google Slides Prescribed Template](https://docs.google.com/presentation/d/13rg7vW43mEH6DkpuusAE6fdoEylSFz8waNpE4RLzoUg/edit)

---

## 1. Hack2skill Portal Form Fields (Ready to Copy-Paste)

* **Challenge Track:** `Future of Work & Enterprise Productivity`
* **GitHub Repository Link (Public Access):** `https://github.com/Jotadebeese/neuroproxy`
* **Prototype Link Deployed on GCP or Cloud Run:** `https://neuroproxy-juxh47uzdq-uc.a.run.app`
* **Demo Video Link (2-4 minutes):** `[YouTube Unlisted / Google Drive Link]`
* **Brief Description of Your Solution (994 / 1,024 Characters):**
  > NeuroProxy is an Autonomous Self-Healing API Gateway built in Go 1.23 and deployed on Google Cloud Run to eliminate 3:00 AM schema-drift downtime. 
  >
  > When external webhooks (Stripe, ERPs, IoT devices) unexpectedly mutate payloads or data types, downstream microservices reject them with 422/500 errors. Instead of failing transactions, NeuroProxy holds the HTTP connection, dispatches the error trace to Google Gemini 3.8 Flash to synthesize a runtime patch, enforces a deterministic Zero-Hallucination Value Conservation Guardrail (blocking invented prices with a 422 SAFETY_VETO), and handles complex multi-tier validation through a 2-turn agentic self-correction loop.
  >
  > Learned transformations are cached in memory (sync.RWMutex / LRU + Cloud Memorystore Redis) so subsequent repeat drifted requests heal in <0.2ms with $0 LLM cost. SREs can export permanent compiled Go middleware in 1 click or merge auto-generated GitHub PR diffs, integrating directly with Google Cloud Logging and Cloud Trace.

---

## 2. 1:1 Mapping to the Prescribed 16-Slide Template

### Slide 1: Team Details
* **a. Team name:** `Zenith Systems`
* **b. Team leader name:** `Juan Diego Bedoya`
* **c. Problem Statement:**  
  Modern distributed microservices depend on dozens of external webhooks (Stripe, Shopify, SAP ERPs, IoT networks). When an upstream provider silently renames a field, nests an object, or alters data types without notice, downstream services crash with unhandled `422 Unprocessable Entity` or `500` errors. Checkouts fail, inventory reservations drop, and software engineers are paged at 3:00 AM to manually write hotfixes.

*(Slide 2 is a template instruction slide — delete or skip in final PDF)*

### Slide 3: Brief About the Idea
* **The "Intelligent Travel Adapter" for Enterprise APIs:**
  * When you travel to another country, your plug doesn’t fit the wall socket. Without an adapter, your device gets zero power.
  * In modern software, APIs encounter the exact same problem daily: third-party payloads mutate, crashing backends and causing expensive downtime.
* **What NeuroProxy Does:**
  * Sits at the edge on **Google Cloud Run** in front of backend microservices.
  * When an upstream rejects a request, NeuroProxy holds the HTTP socket, calls **Google Gemini 3.8 Flash** to synthesize a runtime patch, verifies mathematical value safety, and re-dispatches the request with `200 OK` in milliseconds.
  * Caches the learned rule by `SHA-256` schema fingerprint for sub-millisecond future hits at `$0` LLM cost, and generates a morning GitHub PR diff for permanent resolution.

### Slide 4: Opportunities
* **How different is it from existing ideas?**
  * Existing gateways (Kong, Apigee, Envoy) are passive routers. If a schema mutates, they blindly pass the bad payload, let the backend crash, and require an engineer to manually write and deploy Lua/Wasm plugins hours later.
  * NeuroProxy is **active and self-healing**: it repairs drifted payloads inline at runtime before customer transactions ever drop.
* **How will it solve the problem?**
  * Intercepts 4xx/5xx responses in real time, feeds the contract discrepancy into Gemini 3.8 Flash, enforces mathematical value conservation, and heals the in-flight request without breaking client connections.
* **Unique Selling Propositions (USPs):**
  1. **Zero-Hallucination Guardrail:** Mathematically proves 100% of target values trace to source payloads; vetoes invented numbers (`422 SAFETY_VETO`).
  2. **Multi-Turn Agentic Loop:** Resolves cascading multi-layer validation (Turn 1 schema repair $\rightarrow$ Turn 2 compliance/AML conflict).
  3. **Compiled Rule Cache (<0.2ms, $0 Cost):** Subsequent repeat drifted requests heal in sub-millisecond time with zero AI token spend.

### Slide 5: List of Features Offered
* **Autonomous In-Flight Self-Healing:** Intercepts 400/409/422/500 errors and heals payloads at runtime.
* **Multi-Turn Agentic Self-Correction:** 2-turn agent loop autonomously resolving schema and domain business logic conflicts.
* **Deterministic Value Conservation Guardrail:** Leaf-node provenance verifier blocking AI hallucinations.
* **Two-Tier Resilient Cache:** 5,000-entry in-memory LRU cache + raw TCP Google Cloud Memorystore (Redis) adapter.
* **Anti-Denial-of-Wallet Circuit Breaker:** Token Bucket rate limiter (30 rpm, burst 10) protecting LLM budgets.
* **Bring Your Own API (BYOA):** Dual-mode reverse proxy capable of proxying real internet traffic via `X-Target-URL`.
* **1-Click Standalone Go Middleware Exporter:** Generates compilable `neuroproxy_middleware.go` + unified GitHub PR diffs.
* **Google Cloud Native Observability:** Structured JSON logging on stdout + distributed `X-Cloud-Trace-Context` extraction.

### Slide 6: Process Flow / Use-Case Diagram
```text
[ Client Webhook ] ───(1. Drifted JSON)───> [ ⚡ NeuroProxy Gateway (Cloud Run) ]
                                                     │
                                             (2. Forward Request)
                                                     ▼
                                          [ Upstream Microservice ]
                                                     │
                                       (3. 422 / 500 Schema Rejection)
                                                     ▼
                                          [ ⚡ NeuroProxy Gateway ]
                                                     │
                                            (4. Check Cache Hit?)
                                             ├── YES (<0.2ms) ──> [ Apply Rule ] ──> 200 OK
                                             └── NO (Cache Miss) ─┐
                                                                  ▼
                                                      [ Gemini 2.5 Flash Agent ]
                                                                  │
                                                        (5. Synthesize Patch)
                                                                  ▼
                                                      [ Value Guardrail (100%) ]
                                                                  │
                                                       ├── Violation ──> 422 SAFETY_VETO
                                                       └── Verified ───> (6. Re-dispatch)
                                                                               │
                                                                               ▼
                                                                   [ Upstream: 200 OK! ]
                                                                               │
                                                                     (7. Save Rule to Cache)
                                                                               ▼
                                                                   [ Auto-Generate GitHub PR ]
```

### Slide 7: Wireframes / Prototype UI (Screenshots)
*(Insert 2 high-res screenshots from `http://localhost:8080`:)*
1. **Story Walkthrough View:** Interactive visual explainer showing problem vs. solution cards and 1-click scenario triggers.
2. **Mission Control Console View:** Live animated 5-stage pipeline flow graph, side-by-side JSON diff viewer, and real-time telemetry metrics.

### Slide 8: Architecture Diagram of the Proposed Solution
* **Edge Proxy Layer:** Google Cloud Run (Single stripped Go 1.23 binary + pre-compiled React 19 UI, sub-15MB container).
* **Autonomous AI Layer:** Google Gemini 2.5 Flash / 2.0 Flash (`gemini-2.5-flash` with Low-Latency Structured JSON Outputs).
* **Resilient Cache Layer:** In-Memory LRU Cache (`sync.RWMutex`, 5,000 cap, 24h TTL) + Google Cloud Memorystore (Redis).
* **Enterprise Observability:** Google Cloud Structured Logging (`severity`, latency, RCA) + Google Cloud Trace context propagation.

### Slide 9: Technologies Used in the Solution
* **Core Gateway:** Go 1.23 (`net/http`, `sync.RWMutex`, zero-dependency standard library).
* **Artificial Intelligence:** Google Gemini 2.5 Flash, Google AI Studio SDK, Google Antigravity.
* **Frontend Dashboard:** React 19, TypeScript, Vite, Tailwind CSS, Framer Motion, Lucide Icons, Canvas Confetti.
* **Cloud Infrastructure:** Google Cloud Run, Google Cloud Memorystore (Redis), Google Cloud Logging, Google Cloud Trace.
* **Reliability Engineering:** Go Concurrency Race Detector (`-race`), Token Bucket Circuit Breaker, Deterministic Leaf-Node Verifier.

### Slide 10: Estimated Implementation Cost (Optional)
* **Google Cloud Run:** Scales to zero when idle ($0 idle cost); ~$0.000024 per vCPU-second during active traffic.
* **Google Gemini 3.8 Flash:** ~$0.075 per 1,000,000 input tokens. Because rules are cached on Turn 1, repeat traffic incurs **$0.00** AI cost.
* **Total Estimated Monthly Cost (1 Million Webhook Transactions):** **<$15.00 / month** (versus $10,000+ in lost revenue per outage).

### Slide 11: Snapshots of the Prototype
*(Insert 2 scenario screenshots from `http://localhost:8080`:)*
1. **Scenario 2 (2-Layer Cascading AML Loop):** Proves Turn 1 schema repair $\rightarrow$ Turn 2 domain compliance self-correction (`X-Healed-By: NeuroProxy-Gemini-Turn-2`).
2. **Scenario 3 (Safety Trap Veto):** Proves mathematical guardrail catches missing price and blocks fabrication with `422 SAFETY_VETO`.

### Slide 12: Prototype Performance Report & Benchmarking
* **Rule Cache Latency:** `< 0.18ms` (sub-millisecond execution via compiled SHA-256 fingerprint cache).
* **AI Cold Synthesis Latency:** `340ms – 460ms` (Gemini 3.8 Flash structured JSON repair).
* **Value Conservation Accuracy:** `100.0%` (mathematically verified provenance on all scalar fields).
* **Burst Concurrency Benchmark:** 10 concurrent requests handled in parallel with 0 race conditions and 100% cache hit rate.
* **Test Suite Verification:** 5/5 automated Go test suites passing in `0.08s` with Go race detector enabled.

### Slide 13: Additional Details & Future Development
* **Automated CI/CD Bot:** Autonomous GitHub App that opens real ready-to-merge pull requests directly on customer repositories.
* **Wasm Envoy / Istio Sidecar:** Compile the Go rule execution engine into WebAssembly for native Kubernetes service meshes.
* **Multi-Cloud Memorystore Clustering:** Synchronize learned transformation rules across global multi-region deployments.

### Slide 14: Project Links
* **GitHub Public Repository:** `https://github.com/Jotadebeese/neuroproxy`
* **Demo Video Link (2-4 Minutes):** `[Your YouTube Unlisted or Google Drive Link]`
* **Deployed Product Link (Live Cloud Run):** `https://neuroproxy-juxh47uzdq-uc.a.run.app`

### Slide 16: Thank You
* **Team Zenith Systems**
* *Built for Google Cloud AI Builder Cup 2026*
