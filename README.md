# SRE Workshop Capstone

**English** · [繁體中文](README.zh-TW.md)

Capstone exercise for the SDC SRE Workshop. This repo contains a Go service, an unfinished CI/CD pipeline, and a Prometheus setup. You'll use Docker to package the service, push the image to the GHCR registry via CI/CD, and let Prometheus on the SDC machine scrape its metrics to drive monitoring and alerts.

## Architecture

```
  push to main
       │
       ▼
 GitHub Actions (CI)  ──►  build image  ──►  ghcr.io/Ocean1029/sre-workshop-capstone
       │
       ▼
 GitHub Actions (CD)  ──►   self-hosted runner  ──►  docker run app
                                                                │
                                               Prometheus scrape┘
                                                     │
                                                alert fires
                                                     ▼
                                          Alertmanager ──► Discord
```

## Service

- `GET /` → `ok`
- `GET /healthz` → `ok`
- `GET /crash` → 500, increments `app_crash_total` counter; a change in the counter fires an alert.
- `GET /metrics` → Prometheus metrics

## Flow

### Part 0 — Walk through the exercise (5 min)

Instructor explains the architecture above and what you'll change.

### Part 1 — Run it locally (10 min)

```bash
docker compose up --build
```

Then:

1. Open Prometheus at [http://localhost:9090](http://localhost:9090).
2. Under **Status → Targets**, confirm `capstone-app` is **UP**.
3. In **Graph**, query `app_crash_total`. Start at 0.
4. Hit the `/crash` endpoint a few times and watch the counter climb:
   ```bash
   curl localhost:8080/crash
   curl localhost:8080/crash
   ```

### Part 2 — Wire up CD and trigger a real alert (10 min)

Before the workshop, you should already have:

- Been added as a collaborator on this repo (check your GitHub inbox).
- Been assigned a two-digit `ID` (e.g. `07`). Your deploy runs as `capstone-app-<ID>` on port `80<ID>` of the shared SDC machine; 50 students share the 8001–8050 range.

Flow:

1. Clone the repo and create your student branch:
   ```bash
   git clone https://github.com/Ocean1029/sre-workshop-capstone.git
   cd sre-workshop-capstone
   git checkout -b student-<ID>           # e.g. student-07
   ```
2. Open [`.github/workflows/cd.yml`](.github/workflows/cd.yml). Fill in the `TODO` blocks following CI/CD workshop ch04 (self-hosted runner + `docker pull`/`docker run`). `STUDENT_ID` and `IMAGE` are already exported as env vars, so the TODO steps just need plain docker commands that reference `$STUDENT_ID` and `$IMAGE`.
3. Commit and push your branch:
   ```bash
   git add .github/workflows/cd.yml
   git commit -m "Fill in CD deploy steps"
   git push -u origin student-<ID>
   ```
4. Watch GitHub Actions: CI runs lint + test + build (pushing `:student-<ID>`), then the deploy job runs your filled-in CD on the SDC runner.
5. Once CD is green, hit your deployed `/crash` (replace `<ID>`):
   ```bash
   curl <sdc-host>:80<ID>/crash
   curl <sdc-host>:80<ID>/crash
   ```
6. Within ~1 minute, check the Discord alerts channel. You should see `服務 localhost:80<ID> crashed` with your student number in the description.

## Repository Layout

```
sre-workshop-capstone/
├── main.go / main_test.go       # HTTP server + tests
├── Dockerfile                   # Multi-stage, non-root image
├── docker-compose.yml           # Part 1 local stack
├── .github/workflows/
│   ├── ci.yml                   # Complete: lint / test / build & push to GHCR
│   └── cd.yml                   # Skeleton: fill in for Part 2
└── prometheus/
    ├── prometheus.yml           # Scrape config (used by local compose)
    └── rules/alerts.yml         # AppCrashing alert rule
```

## Related

- [SRE Workshop main repo](https://github.com/Ocean1029/sre-workshop) — Docker, CI/CD, and Prometheus material
