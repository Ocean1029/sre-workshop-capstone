# SDC Machine Setup

Instructor-only reference for preparing the social office machine before the 4/25 workshop. These files are **not** what students clone; they're the server-side config that makes Part 2 work.

## Files

- `prometheus.yml` — scrape config for student containers on ports 8001–8030 plus the instructor dry-run on 8099. Drop this at `/etc/prometheus/prometheus.yml` on the SDC machine (or mount as a volume).

## Checklist

- [ ] Docker + docker compose installed on the SDC machine
- [ ] GitHub self-hosted runner registered to **each student's fork** (same runner serves all; runners match via the `sdc-runner` label in `cd.yml`)
- [ ] Prometheus container running with `prometheus.yml` from this folder
- [ ] Alertmanager container running with a Discord receiver (webhook for `#sre-workshop-alerts`)
- [ ] Alert rule `prometheus/rules/alerts.yml` (from the capstone repo root) mounted into Prometheus
- [ ] If the SDC host is Linux and Prometheus runs in Docker, add `--add-host=host.docker.internal:host-gateway` to make the scrape targets resolvable

## Port map

| Role | Host port |
|------|-----------|
| Student capstone containers | `8001`–`8030` |
| Instructor dry-run container | `8099` |
| Prometheus UI | `9090` |
| Alertmanager UI | `9093` |

Labels applied per target: `student_id="01"..."30"` and `student_id="99"` for the instructor. Alerts therefore include `student_id` in the payload, so Discord messages tell you which student's container fired.
