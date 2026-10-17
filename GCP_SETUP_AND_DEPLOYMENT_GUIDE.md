# 🛠️ NeuroProxy: GCP Setup, Local Execution & Cloud Deployment Guide

> **Project:** NeuroProxy — Autonomous Self-Healing API Gateway  
> **Target Platform:** Google Cloud Run Edge, Google Gemini 2.5 Flash, Google Cloud Memorystore (Redis), Google Cloud Logging & Trace  
> **Team:** Zenith Systems | Google Cloud AI Builder Cup 2026

---

## 1. Quick Local Execution Guide

You can run NeuroProxy locally with or without Docker, with or without a live Gemini API key (a deterministic structural synthesizer is built in as an offline fallback).

### Prerequisites
- **Go 1.23+** (or Docker)
- **Node.js 18+ & npm** (only needed if modifying the React frontend; pre-compiled production assets are already bundled in `frontend/out/`)
- *(Optional)* **Gemini API Key** from [Google AI Studio](https://aistudio.google.com/apikey)
- *(Optional)* **Redis** (Local Docker container or Google Cloud Memorystore)

---

### Step 1.1: Local Environment Configuration
Copy the template and create your `.env` file:
```bash
cp .env.example .env
```
Edit `.env` with your preferred settings:
```ini
# Google Gemini API Key (from Google AI Studio or Vertex AI)
GEMINI_API_KEY=your-gemini-api-key-here

# HTTP Server Port
PORT=8080

# Google Cloud Project ID (used for Cloud Trace and Cloud Logging)
GOOGLE_CLOUD_PROJECT=neuroproxy-demo-2026

# Optional: Redis / Cloud Memorystore connection string
# Leave empty to use the built-in 5,000-entry in-memory LRU cache with 24h TTL
REDIS_URL=
```

---

### Step 1.2: Run Locally via Go
Run the canonical production binary directly:
```bash
go run ./cmd/neuroproxy
# or
go run .
```

You will see the startup banner in your terminal:
```text
⚡ NeuroProxy v2.0 Self-Healing API Gateway running on http://localhost:8080
   ├─ Mission Control UI : http://localhost:8080/
   ├─ Proxy Ingestion    : POST http://localhost:8080/api/proxy/{checkout|inventory|telemetry}
   ├─ Middleware Export  : GET  http://localhost:8080/api/export/middleware.go
   ├─ Cloud Project      : neuroproxy-demo-2026
   └─ Gemini API Key     : true
```
Open **`http://localhost:8080`** in your browser to interact with the **Mission Control Dashboard & Chaos Simulator**.

---

### Step 1.3: Run Locally via Docker
Build and run the sub-15MB container image:
```bash
# Build container
docker build -t neuroproxy:latest .

# Run container
docker run -p 8080:8080 \
  -e GEMINI_API_KEY="your-gemini-api-key" \
  -e PORT=8080 \
  neuroproxy:latest
```

---

### Step 1.4: (Optional) Local Redis Setup
To test distributed multi-instance cache sharing locally using Redis:
```bash
# Start local Redis in Docker
docker run -d --name neuroproxy-redis -p 6379:6379 redis:alpine

# Run NeuroProxy pointing to local Redis
REDIS_URL="redis://localhost:6379" go run ./cmd/neuroproxy
```
NeuroProxy will log:
```text
✅ Google Cloud Memorystore (Redis) connected at localhost:6379
```
*(If Redis is stopped or unreachable, NeuroProxy automatically falls back to its internal thread-safe LRU cache with zero downtime).*

---

## 2. Google Cloud Platform Setup (Cloud Run & Services)

Follow this step-by-step checklist to deploy NeuroProxy to Google Cloud Run.

```mermaid
flowchart TD
    A["1. Google Cloud Project Setup<br>(gcloud config set project)"] --> B["2. Enable GCP APIs<br>(Cloud Run, Artifact Registry, Build)"]
    B --> C["3. Store Secrets<br>(Secret Manager: GEMINI_API_KEY)"]
    C --> D{"4. Cache Selection"}
    D -->|"Option A (Standard)"| E["Google Cloud Memorystore (Redis)<br>+ Serverless VPC Access Connector"]
    D -->|"Option B (Serverless)"| F["Upstash / Redis Cloud URI<br>(Zero-VPC Setup)"]
    D -->|"Option C (Simplest)"| G["Built-In In-Memory LRU<br>(Zero Setup, <0.2ms, 5,000 Cap)"]
    E --> H["5. Build & Deploy to Cloud Run Edge<br>(cloudbuild.yaml or gcloud run deploy)"]
    F --> H
    G --> H
    H --> I["6. Verify Live HTTPS Deployment<br>(Cloud Logging & Cloud Trace active)"]
```

---

### Step 2.1: Authenticate and Configure Project
```bash
# Log in to Google Cloud
gcloud auth login

# Set your active GCP project ID
export PROJECT_ID="your-gcp-project-id"
export REGION="us-central1"

gcloud config set project $PROJECT_ID
gcloud config set run/region $REGION
```

---

### Step 2.2: Enable Required Google Cloud APIs
Enable all services needed for container hosting, build automation, secrets, and observability:
```bash
gcloud services enable \
  run.googleapis.com \
  artifactregistry.googleapis.com \
  cloudbuild.googleapis.com \
  secretmanager.googleapis.com \
  logging.googleapis.com \
  cloudtrace.googleapis.com \
  redis.googleapis.com \
  vpcaccess.googleapis.com
```

---

### Step 2.3: Store Gemini API Key in Secret Manager
Protect your Google AI Studio or Vertex AI Gemini API key:
```bash
# Create the secret
gcloud secrets create gemini-api-key --replication-policy="automatic"

# Add your key as a version
echo -n "AIzaSy..." | gcloud secrets versions add gemini-api-key --data-file=-

# Grant Cloud Run service account access to read the secret
PROJECT_NUMBER=$(gcloud projects describe $PROJECT_ID --format="value(projectNumber)")
gcloud secrets add-iam-policy-binding gemini-api-key \
  --member="serviceAccount:${PROJECT_NUMBER}-compute@developer.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor"
```

---

### Step 2.4: Setting Up the Resilient Cache (3 Options)

NeuroProxy features a **Two-Tier Resilient Cache Layer** (`internal/cache/cache.go` and `internal/cache/redis.go`). Choose the option that fits your deployment:

#### Option A: Built-In In-Memory LRU Cache (Simplest, Zero Extra Cost)
- **Configuration:** Leave `REDIS_URL` unset.
- **How it works:** Uses a thread-safe doubly linked list + hash map capped at 5,000 entries with 24-hour TTL expiration.
- **Performance:** **`<0.2ms`** cache lookup directly in Go memory.
- **Ideal for:** Rapid evaluation, hackathon judging, and single-replica Cloud Run setups.

#### Option B: Google Cloud Memorystore for Redis (Enterprise GCP Native)
If deploying across multiple horizontally auto-scaling Cloud Run replicas, use Google Cloud Memorystore:
```bash
# 1. Create a Serverless VPC Access Connector (required for Cloud Run to reach Memorystore)
gcloud compute networks vpc-access connectors create neuroproxy-vpc-connector \
  --region=$REGION \
  --range="10.8.0.0/28" \
  --network=default

# 2. Provision Google Cloud Memorystore (Redis instance)
gcloud redis instances create neuroproxy-cache \
  --size=1 \
  --region=$REGION \
  --tier=basic \
  --redis-version=redis_7_0

# 3. Retrieve the Redis Host IP
REDIS_IP=$(gcloud redis instances describe neuroproxy-cache --region=$REGION --format="value(host)")
echo "Memorystore IP: $REDIS_IP"
# Your REDIS_URL will be: redis://${REDIS_IP}:6379
```

#### Option C: Serverless Redis (Upstash / Redis Cloud)
If you prefer not to configure a VPC connector:
1. Create a free Redis instance on [Upstash](https://upstash.com) or Redis Cloud.
2. Set `REDIS_URL="rediss://default:password@your-endpoint.upstash.io:6379"`.

---

### Step 2.5: Create Google Artifact Registry Repository
```bash
gcloud artifacts repositories create neuroproxy-repo \
  --repository-format=docker \
  --location=$REGION \
  --description="Docker repository for NeuroProxy gateway"
```

---

## 3. Automated Deployment to Google Cloud Run

### Method 1: 1-Command Automated Cloud Build (Recommended)
Submit the deployment using the included [`cloudbuild.yaml`](file:///home/hitesh/Documents/Antigravity%20files/neuroproxy/cloudbuild.yaml):
```bash
gcloud builds submit --config=cloudbuild.yaml \
  --substitutions=_REGION=$REGION,_REPO_NAME=neuroproxy-repo
```
Cloud Build will execute:
1. Automated Go unit & integration test suite (`go test -v ./...`)
2. Production React bundle build (`npm run build`)
3. Multi-stage Docker image build (<15MB container)
4. Push to Google Artifact Registry
5. Deploy to Google Cloud Run with edge optimizations

---

### Method 2: Direct CLI Deployment via `gcloud run deploy`
If building and deploying directly from your terminal:
```bash
gcloud run deploy neuroproxy-gateway \
  --source=. \
  --region=$REGION \
  --platform=managed \
  --allow-unauthenticated \
  --cpu=1 \
  --memory=512Mi \
  --concurrency=80 \
  --timeout=30s \
  --min-instances=1 \
  --max-instances=10 \
  --set-secrets="GEMINI_API_KEY=gemini-api-key:latest" \
  --set-env-vars="GCP_PROJECT=$PROJECT_ID,PORT=8080"
```
*(Add `--vpc-connector=neuroproxy-vpc-connector --set-env-vars="REDIS_URL=redis://${REDIS_IP}:6379"` if using Option B).*

Once deployed, Google Cloud Run will output your live URL:
```text
Service [neuroproxy-gateway] revision [neuroproxy-gateway-00001-xyz] has been deployed and is serving 100 percent of traffic.
Service URL: https://neuroproxy-gateway-xyz-uc.a.run.app
```

---

## 4. Observability: Cloud Logging & Cloud Trace Setup

NeuroProxy emits **Google Cloud Structured Logging JSON** with automated trace correlation.

### Viewing Logs in Google Cloud Logs Explorer
1. Navigate to **Google Cloud Console $\rightarrow$ Logging $\rightarrow$ Logs Explorer**.
2. Query specifically for self-healing incidents:
   ```sql
   resource.type="cloud_run_revision"
   resource.labels.service_name="neuroproxy-gateway"
   jsonPayload.component="neuroproxy-engine"
   ```
3. Filter by severity:
   - `severity = "WARNING"`: Intercepted schema drifts that required Gemini agentic healing.
   - `severity = "CRITICAL"`: Blocked hallucination attempts flagged by the Value Conservation Guardrail (`SAFETY_VETO`).
   - `severity = "INFO"`: Sub-millisecond hot-path cache hits and valid passthrough requests.

### Viewing Distributed Traces in Google Cloud Trace
When requests pass through Cloud Load Balancing, the `X-Cloud-Trace-Context` header is extracted and correlated into `logging.googleapis.com/trace`.
1. Go to **Google Cloud Console $\rightarrow$ Trace $\rightarrow$ Trace Explorer**.
2. Select any trace span to see the exact time spent in:
   - Request Ingestion
   - LRU Cache Lookup
   - Upstream Microservice Execution
   - Gemini 2.5 Flash Autonomous Synthesis Loop
   - Value Conservation Verification

---

## 5. Verification & Smoke Test Checklist for Evaluators

Execute these quick tests against your live Cloud Run URL to verify all capabilities:

| Test # | Test Purpose | Command | Expected Result |
| :---: | :--- | :--- | :--- |
| **1** | **Health & Readiness** | `curl -i https://YOUR_URL/healthz` | `HTTP 200 OK` + `{"status":"ok","service":"neuroproxy-gateway"}` |
| **2** | **Mission Control UI** | Open `https://YOUR_URL/` in an **Incognito Browser Window** | Animated React 19 dashboard renders with live node-graph visualizer |
| **3** | **Autonomous Self-Healing** | `curl -i -X POST https://YOUR_URL/api/proxy/checkout -H "Content-Type: application/json" -d '{"numero_orden":"ORD-99","pago":{"monto_usd":149.50,"moneda":"USD"},"cliente":{"correo_electronico":"test@user.io"}}'` | `HTTP 200 OK`<br>`X-Healed-By: NeuroProxy-Gemini-Turn-1`<br>`X-Value-Conservation: 100%-Verified` |
| **4** | **Sub-ms Cache Hit** | Re-run Test #3 command immediately | `HTTP 200 OK`<br>`X-Healed-By: NeuroProxy-Cache`<br>Latency drops from ~90ms to `<0.2ms` |
| **5** | **Safety Veto Guardrail** | `curl -i -X POST https://YOUR_URL/api/proxy/checkout -H "Content-Type: application/json" -d '{"order_number":"ORD-99","currency":"USD"}'` | `HTTP 422 Unprocessable Entity`<br>`X-Healed-By: Blocked-Safety-Veto` |
| **6** | **Live External Reverse Proxy** | `curl -i -X POST "https://YOUR_URL/api/proxy/my-service?target=https://httpbin.org/post" -H "Content-Type: application/json" -d '{"status":"ok"}'` | Proxies live to `httpbin.org` and returns `200 OK` |
| **7** | **Raw Bypass Mode** | `curl -i -X POST "https://YOUR_URL/api/proxy/checkout?bypass=true" -H "Content-Type: application/json" -d '{"pago":{"monto_usd":10}}'` | `HTTP 400 Bad Request` (confirms self-healing was disabled by client flag) |
