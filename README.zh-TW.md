# SRE Workshop Capstone

[English](README.md) · **繁體中文**

SDC SRE Workshop（4/25）的綜合練習。把一個小小的 Go 服務透過 CI/CD 送到社辦機器，讓 Prometheus 抓 metric，在 `/crash` 被打到時觸發 Discord 告警。

## 架構

```
  push 到 main
       │
       ▼
 GitHub Actions (CI)  ──►  build image  ──►  ghcr.io/Ocean1029/sre-workshop-capstone
       │
       ▼
 GitHub Actions (CD)  ──►  社辦機器 self-hosted runner  ──►  docker run app
                                                                │
                                               Prometheus scrape ┘
                                                     │
                                                告警觸發
                                                     ▼
                                          Alertmanager ──► Discord
```

## 服務

- `GET /` → `ok`
- `GET /healthz` → `ok`
- `GET /crash` → 500，遞增 `app_crash_total` counter
- `GET /metrics` → Prometheus 格式

## 練習流程

### Part 0 — 介紹練習（5 分鐘）

講師走一次上面的架構，說明待會要改什麼。

### Part 1 — 在本地跑起來（10 分鐘）

```bash
docker compose up --build
```

接著：

1. 打開 [http://localhost:9090](http://localhost:9090) 的 Prometheus UI。
2. 進 **Status → Targets**，確認 `capstone-app` 是 **UP**。
3. 到 **Graph** 輸入 `app_crash_total`，目前是 0。
4. 打幾次 `/crash`，回去看 counter 變化：
   ```bash
   curl localhost:8080/crash
   curl localhost:8080/crash
   ```

### Part 2 — 補齊 CD，觸發真正的告警（10 分鐘）

工作坊開始前你應該已經做好：

- Fork 這個 repo 到自己的 GitHub 帳號
- 打開自己 fork 的 **Settings → Secrets and variables → Actions → Variables**，新增 `STUDENT_ID`，值是講師發給你的兩位數編號（例如 `07`）。你的 container 會部署到社辦機器的 `80<ID>` port，這樣最多 50 人可以共存

接著：

1. 打開 [`.github/workflows/cd.yml`](.github/workflows/cd.yml)，參考 CI/CD 工作坊 ch04 的做法，把 `TODO` 幾個 step 補完（self-hosted runner + `docker pull` / `docker run`）。hint 裡面已經有用到 `vars.STUDENT_ID`，直接照抄即可。
2. Commit 並 push：
   ```bash
   git add .github/workflows/cd.yml
   git commit -m "Fill in CD deploy steps"
   git push
   ```
3. 到 GitHub Actions 看 workflow，CI 綠了才會觸發 CD。
4. CD 也綠了之後，打自己那台的 `/crash`（把 `<ID>` 換掉）：
   ```bash
   curl <社辦機器>:80<ID>/crash
   curl <社辦機器>:80<ID>/crash
   ```
5. 約一分鐘內到 Discord 告警頻道確認，會看到 `AppCrashing`，`instance` label 會指到你的 port。

## Repo 結構

```
sre-workshop-capstone/
├── main.go / main_test.go       # HTTP server + 測試
├── Dockerfile                   # 多階段、non-root
├── docker-compose.yml           # Part 1 的本地環境
├── .github/workflows/
│   ├── ci.yml                   # 完整：lint / test / build & push 到 GHCR
│   └── cd.yml                   # 骨架：Part 2 要補完
└── prometheus/
    ├── prometheus.yml           # 本地 compose 用的 scrape 設定
    └── rules/alerts.yml         # AppCrashing 告警規則
```

## 相關連結

- [SRE Workshop 主 repo](https://github.com/Ocean1029/sre-workshop) — Docker、CI/CD、Prometheus 教材
