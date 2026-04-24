# SRE Workshop Capstone

**English** · [繁體中文](README.zh-TW.md)

End-to-end capstone for the SDC SRE Workshop (4/25): ship a tiny Go service through CI/CD to the SDC machine, scrape it with Prometheus, and fire a Discord alert when `/crash` is hit.

## Architecture

```
  push to main
       │
       ▼
 GitHub Actions (CI)  ──►  build image  ──►  ghcr.io/Ocean1029/sre-workshop-capstone
       │
       ▼
 GitHub Actions (CD)  ──►  self-hosted runner (SDC machine)  ──►  docker run app
                                                                    │
                                               Prometheus scrapes  ─┘
                                                     │
                                                alert fires
                                                     ▼
                                          Alertmanager ──► Discord
```

## Service

- `GET /` → `ok`
- `GET /healthz` → `ok`
- `GET /crash` → 500, increments `app_crash_total` counter
- `GET /metrics` → Prometheus exporter

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

- Forked this repo to your own GitHub account.
- Opened your fork → **Settings → Secrets and variables → Actions → Variables** and added a variable named `STUDENT_ID` set to the two-digit number your instructor gave you (e.g. `07`). Your container deploys to port `80<ID>` on the SDC machine so up to 50 people can coexist.

Then:

1. Open [`.github/workflows/cd.yml`](.github/workflows/cd.yml). Fill in the `TODO` steps following CI/CD workshop ch04 (self-hosted runner + `docker pull`/`docker run`). The hints already mention `vars.STUDENT_ID` for container name and port.
2. Commit and push:
   ```bash
   git add .github/workflows/cd.yml
   git commit -m "Fill in CD deploy steps"
   git push
   ```
3. Watch the GitHub Actions run. CI must go green before CD kicks in.
4. Once CD is green, hit your own deployed `/crash` (replace `<ID>`):
   ```bash
   curl <sdc-host>:80<ID>/crash
   curl <sdc-host>:80<ID>/crash
   ```
5. Within ~1 minute, check the Discord alerts channel. You should see `AppCrashing` with an `instance` label pointing at your port.

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
