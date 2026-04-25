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

We'll first run the whole stack on your laptop so you can see how Prometheus scrapes the app's metrics and how `/crash` moves a counter. Part 2 is the same pipeline pushed onto the SDC machine — only the runner and the Prometheus instance change.

First, clone the repo (Part 2 will branch off the same working copy):

```bash
git clone https://github.com/Ocean1029/sre-workshop-capstone.git
cd sre-workshop-capstone
```

Now bring up the stack. `docker-compose.yml` runs the Go app and Prometheus together, so you don't have to wire two `docker run` commands by hand:

```bash
docker compose up --build
```

First run takes a minute or two (Go build + Prometheus pull). When you see `app | ... listening on :8080` and `prometheus | ... msg="Server is ready to receive web requests."`, both containers are up.

Verify that Prometheus is actually scraping the app:

1. Open [http://localhost:9090](http://localhost:9090) — this is the Prometheus UI.
2. Click **Status → Targets**, find the `capstone-app` job; its state should be **UP** (green). If it's red/DOWN, the compose network probably isn't ready yet — check `docker compose` logs.
3. Click **Graph** at the top, type `app_crash_total`, and Execute. The chart shows 0 because nothing has crashed yet.

Now hit `/crash` to drive the counter:

```bash
curl localhost:8080/crash
curl localhost:8080/crash
```

Each call returns 500 and bumps the counter by 1. Re-execute `app_crash_total` in the Graph — the value jumps to 2. The pipeline **app exposes metric → Prometheus scrapes → graph reflects it** is working. Part 2 swaps the app for 50 student containers but keeps the same mechanism.

### Part 2 — Wire up CD and trigger a real alert (10 min)

Part 1 ran on your laptop, with your own Prometheus scraping localhost. In Part 2 we move the app onto the SDC machine: CI builds the image and pushes it to GHCR, the SDC self-hosted runner pulls it and runs it as `capstone-app-<ID>`, the SDC's Prometheus scrapes that container, and Alertmanager fires Discord on a crash. The pipeline is fully wired — your job is to push a student branch and watch it run.

Before the workshop, you should already have:

- Been added as a collaborator on this repo (check your GitHub inbox).
- Been assigned a two-digit `ID` (e.g. `07`). Your deploy runs as `capstone-app-<ID>` on port `80<ID>` of the shared SDC machine — for instance, ID `07` gets port `8007`.

Flow:

1. Create your student branch (you already cloned the repo in Part 1). The name must be `student-<ID>` — the first step in cd.yml extracts `<ID>` from `GITHUB_REF_NAME`, and the rest of the pipeline keys off of it:
   ```bash
   git checkout -b student-<ID>           # e.g. student-07
   ```
2. Skim [`.github/workflows/cd.yml`](.github/workflows/cd.yml) so you know what the deploy job does on the SDC runner: log in to GHCR → `docker pull` the new image → `docker rm -f` the old container → `docker run -d` the new one with port `80<ID>`. `STUDENT_ID` and `IMAGE` are derived from the branch name in the first step and exported into `$GITHUB_ENV`, so the rest of the steps just reference them. Everything is already written — you don't need to edit it.
3. Push your branch to trigger the pipeline:
   ```bash
   git push -u origin student-<ID>
   ```
4. Watch GitHub Actions. Pushing a `student-*` branch triggers CI to lint + test + build, tags the image as `:student-<ID>` and pushes to GHCR; once CI is green, the deploy job runs cd.yml on the SDC runner. End-to-end takes 1–2 minutes.
5. Once CI/CD is green, hit your deployed `/crash` (replace `<ID>` with your two-digit ID):
   ```bash
   curl https://workshop-<ID>.ocean1029.com/crash
   curl https://workshop-<ID>.ocean1029.com/crash
   ```
   Both calls should return 500. The SDC machine has no public IP; the subdomain is fronted by a Cloudflare Tunnel that the workshop host runs, so you don't need to set up any tunnel yourself.
6. Within ~1 minute, check the Discord alerts channel. The SDC's Prometheus scrapes your port every 15 s; once it sees `app_crash_total` move, the `AppCrashing` rule fires and Alertmanager pushes to Discord. You'll see `服務 https://workshop-<ID>.ocean1029.com crashed` with your student number in the description. If nothing shows up after a minute, check **Status → Targets** on the SDC's Prometheus to confirm your port is UP.

## Repository Layout

```
sre-workshop-capstone/
├── main.go / main_test.go       # HTTP server + tests
├── Dockerfile                   # Multi-stage, non-root image
├── docker-compose.yml           # Part 1 local stack
├── .github/workflows/
│   ├── ci.yml                   # Complete: lint / test / build & push to GHCR
│   └── cd.yml                   # Deploy job: pull image + restart container on SDC runner
└── prometheus/
    ├── prometheus.yml           # Scrape config (used by local compose)
    └── rules/alerts.yml         # AppCrashing alert rule
```

## Related

- [SRE Workshop main repo](https://github.com/Ocean1029/sre-workshop) — Docker, CI/CD, and Prometheus material
