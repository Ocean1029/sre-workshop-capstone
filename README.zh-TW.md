# SRE Workshop Capstone

[English](README.md) · **繁體中文**

SDC SRE Workshop 的綜合練習。這個 repo 包含了一個 Go 服務，一個未完成 CI/CD pipeline 以及 Prometheus 設定，這個練習會用 Docker 把服務打包起來，透過 CI/CD 流程把 Image 上傳到 ghcr registry 上，讓伺服器上的 Prometheus 可以讀取到該服務的 Metrics，達成監控與告警功能。

## 架構

```
  push 到 main
       │
       ▼
 GitHub Actions (CI)  ──►  build image  ──►  ghcr.io/Ocean1029/sre-workshop-capstone
       │
       ▼
 GitHub Actions (CD)  ──►   self-hosted runner  ──►  docker run app
                                                                │
                                               Prometheus scrape┘
                                                     │
                                                告警觸發
                                                     ▼
                                          Alertmanager ──► Discord
```

## 服務

- `GET /` → `ok`
- `GET /healthz` → `ok`
- `GET /crash` → 500，增加 `app_crash_total` counter，當有數字變化時就會發通知。
- `GET /metrics` → Prometheus Metrics

### Part 1 — 在本地跑起來

我們先在自己電腦上把整套服務跑一次，看 Prometheus 怎麼抓到 app 的 metric、`/crash` 怎麼讓 counter 動起來。Part 2 把這條路搬到社辦機器之後也是同樣模式，差別只在「誰」啟動 container、scrape 從哪裡來。

先把 repo clone 下來（Part 2 也會在同一份 working copy 開新 branch）：

```bash
git clone https://github.com/Ocean1029/sre-workshop-capstone.git
cd sre-workshop-capstone
```

接著啟動服務。`docker-compose.yml` 把 Go app 跟 Prometheus 兩個 container 一起拉起來：

```bash
docker compose up --build
```

第一次跑會 build image 加 pull Prometheus，大概一兩分鐘。看到 `app | ... listening on :8080` 跟 `prometheus | ... msg="Server is ready to receive web requests."` 兩行就代表都起來了。

驗證 Prometheus 真的有抓到 app 的 metric：

1. 打開 [http://localhost:9090](http://localhost:9090)，這是 Prometheus 的 UI。
2. 點 **Status → Targets**，找 `capstone-app` 這個 job，state 應該是綠色的 **UP**。如果是紅色 DOWN，多半是 compose network 還沒準備好，回去看 `docker compose` 的 log。
3. 點上方 **Graph**，輸入 `app_crash_total` 按 Execute。圖是 0，因為還沒人 crash。

現在我們刻意打 `/crash` 讓 counter 動起來：

```bash
curl localhost:8080/crash
curl localhost:8080/crash
```

每次回 500，counter +1。回 Prometheus Graph 重新 Execute `app_crash_total`，值會跳到 2。

### Part 2 — 補齊 CD，觸發真正的告警

Part 1 的 app 跑在你電腦，scrape 也是你電腦上的 Prometheus。Part 2 把這條路搬到社辦機器：CI 把 image build 完 push 上 GHCR，社辦機 self-hosted runner 拉下來跑成 `capstone-app-<ID>` container，由社辦機那一份 Prometheus scrape，crash 發生時 Alertmanager 會發 Discord。整條 pipeline 已經接好，你要做的只是 push 一個 student branch、看它跑。

工作坊開始前你應該已經：

- 收到 repo collaborator 邀請（去 GitHub 收件匣接受）
- 拿到一個兩位數的 `ID`（例如 `07`）。你的部署會在共用 SDC 機器上跑 `capstone-app-<ID>` container，對外 port `80<ID>`，例如編號 07 會獲得 8007 port。

流程：

1. 建立你的 student branch（Part 1 已經 clone 過了）。命名一定要是 `student-<ID>`，因為 cd.yml 第一個 step 會從 `GITHUB_REF_NAME` 把 `<ID>` 切出來，整條 pipeline 後面都靠這個 branch name 串：
   ```bash
   git checkout -b student-<ID>           # 例如 student-07
   ```

2. 翻一下 [`.github/workflows/cd.yml`](.github/workflows/cd.yml)，看一遍 deploy job 在社辦機上做了什麼：登入 GHCR → `docker pull` 新 image → `docker rm -f` 舊 container → `docker run -d` 起新 container 並對映到 port `80<ID>`。`STUDENT_ID` 跟 `IMAGE` 由第一個 step 從 branch name 推導好寫進 `$GITHUB_ENV`，後面的 step 直接引用就好。內容都已經寫好，你不用改。

3. 把 branch push 上去就會觸發 pipeline：
   ```bash
   git push -u origin student-<ID>
   ```

4. 到 GitHub Actions 看 workflow。`student-*` branch 一 push 就會觸發 CI 跑 lint + test + build，build 完的 image 會用 `:student-<ID>` 這個 tag push 到 GHCR；CI 綠了之後 deploy job 會接著在社辦機 runner 上跑 cd.yml。整條跑下來大約 1–2 分鐘。

5. CI/CD 都綠了之後，打自己那台的 `/crash`（把 `<ID>` 換成自己的兩位數編號）：
   ```bash
   curl https://workshop-<ID>.ocean1029.com/crash
   curl https://workshop-<ID>.ocean1029.com/crash
   ```

6. 約一分鐘內到 Discord 告警頻道確認。社辦機那份 Prometheus 每 15 秒 scrape 一次你的 port，看到 `app_crash_total` 變動就會觸發 `AppCrashing` rule，Alertmanager 收到後 push 到 Discord，你會看到 `服務 localhost:80<ID> crashed`，敘述帶著你的編號。如果一分鐘還沒看到，回社辦機那份 Prometheus 的 **Status → Targets** 確認你的 port 是 UP。

## Repo 結構

```
sre-workshop-capstone/
├── main.go / main_test.go       # HTTP server + 測試
├── Dockerfile                   # 多階段、non-root
├── docker-compose.yml           # Part 1 的本地環境
├── .github/workflows/
│   ├── ci.yml                   # 完整：lint / test / build & push 到 GHCR
│   └── cd.yml                   # Deploy job：pull image + restart container 在社辦機 runner 上
└── prometheus/
    ├── prometheus.yml           # 本地 compose 用的 scrape 設定
    └── rules/alerts.yml         # AppCrashing 告警規則
```

## 相關連結

- [SRE Workshop 主 repo](https://github.com/Ocean1029/sre-workshop) — Docker、CI/CD、Prometheus 教材
