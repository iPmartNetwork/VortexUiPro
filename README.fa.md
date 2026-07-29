<div align="center" dir="rtl">

<h1>VortexUiPro</h1>
<p><strong>پنل مدیریت حرفه‌ای پروکسی</strong> — <em>نسل بعدی • سطح سازمانی • چند هسته‌ای</em></p>

<p>
  <a href="https://github.com/iPmartNetwork/VortexUiPro/releases">
    <img src="https://img.shields.io/github/v/release/iPmartNetwork/VortexUiPro?style=for-the-badge&logo=github&color=blueviolet" alt="انتشار"/>
  </a>
  <a href="https://go.dev/">
    <img src="https://img.shields.io/badge/Go-1.23+-00ADD8?style=for-the-badge&logo=go" alt="Go"/>
  </a>
  <a href="https://react.dev/">
    <img src="https://img.shields.io/badge/React-18-61DAFB?style=for-the-badge&logo=react" alt="React"/>
  </a>
  <a href="LICENSE">
    <img src="https://img.shields.io/github/license/iPmartNetwork/VortexUiPro?style=for-the-badge&color=green" alt="مجوز"/>
  </a>
  <a href="https://github.com/iPmartNetwork/VortexUiPro/stargazers">
    <img src="https://img.shields.io/github/stars/iPmartNetwork/VortexUiPro?style=for-the-badge&logo=github&color=yellow" alt="ستاره‌ها"/>
  </a>
</p>

<p>
  <a href="README.md">English</a> •
  <strong>فارسی</strong>
</p>

<br/>

<p align="center">
  <b>⚡ ۴۰+ سرویس بک‌اند • ۳۹ صفحه فرانت‌اند • ۱۰ زبان • ۷ پروتکل • چند هسته‌ای ⚡</b>
</p>

</div>

---

<div dir="rtl">

## ✨ ویژگی‌ها

### 🎯 هسته مرکزی

<table>
  <tr>
    <th>ویژگی</th>
    <th>توضیحات</th>
  </tr>
  <tr>
    <td><b>موتور چند هسته‌ای</b></td>
    <td>مدیریت همزمان <b>Xray</b> + <b>Sing-box</b> با قابلیت hot-reload</td>
  </tr>
  <tr>
    <td><b>۴۰+ سرویس بک‌اند</b></td>
    <td>مدیریت کامل چرخه عمر پروکسی از ایجاد تا پایش</td>
  </tr>
  <tr>
    <td><b>۳۹ صفحه فرانت‌اند</b></td>
    <td>رابط کاربری مدرن با طراحی شیشه‌ای (Glassmorphism) و تم تاریک/روشن</td>
  </tr>
  <tr>
    <td><b>۷ پروتکل پروکسی</b></td>
    <td>VMess, VLESS, Trojan, Shadowsocks, Hysteria2, WireGuard, MTProto</td>
  </tr>
  <tr>
    <td><b>کلاستر چند نوده</b></td>
    <td>شبکه gRPC با انتخاب رهبر، رمزنگاری mTLS و نمایش توپولوژی زنده</td>
  </tr>
</table>

---

### 🛡️ مجموعه ضد سانسور (Anti-Censorship Suite)

<div align="center">
  
```
┌─────────────────────────────────────────────────┐
│                 Anti-Censorship Suite              │
├──────────┬──────────┬──────────┬─────────────────┤
│ TLS Tricks│ Domain    │ WARP+     │ MTProto Proxy    │
│ Fragment  │ Fronting  │ Outbound  │                  │
│ Padding   │ CDN Scan  │ Cloudflare│ Telegram Secret  │
│ Mixing    │ Config Gen│ Integration│ Auto-Generate    │
│ Anti-DPI  │           │           │                  │
├──────────┴──────────┴──────────┴─────────────────┤
│ Clean IP Scanner        •        Reality Scanner   │
│ Cloudflare IP Discovery  •      Fingerprint Detect │
└─────────────────────────────────────────────────┘
```
  
</div>

| تکنیک | توضیحات |
|:------|:---------|
| **Fragment** | تکه‌تکه کردن بسته‌های TLS برای عبور از Deep Packet Inspection |
| **Padding** | افزودن داده‌های تصادفی به بسته‌ها برای مخفی‌سازی الگوی ترافیک |
| **Mixing** | ترکیب چند تکنیک همزمان برای امنیت بیشتر |
| **Anti-DPI** | تکنیک‌های تخصصی برای مقابله با سیستم‌های تشخیص الگو |
| **CDN Fronting** | کشف خودکار CDN (Cloudflare, Fastly, Akamai) + تولید کانفیگ |
| **WARP+** | یکپارچه‌سازی با Cloudflare WARP برای اتصال رمزنگاری شده |
| **MTProto** | تولید پروکسی تلگرام با مدیریت کلید محرمانه |
| **Clean IP** | اسکنر خودکار IPهای سالم Cloudflare |
| **Reality Scan** | اسکن TLS Reality با انگشت‌نگاری پیشرفته |

---

### 🔐 امنیت و کنترل دسترسی

| لایه | فناوری | توضیحات |
|:----:|:------:|:---------|
| 👨‍💼 | **RBAC** | ابرمدیر، مدیر، ریسلر با مجوزهای قابل تنظیم |
| 🔑 | **API Tokens** | دسترسی محدود با انقضا برای اپلیکیشن‌های شخص ثالث |
| 🌍 | **Geo-Blocking** | مسدودسازی بر اساس کشور با لیست مجاز/غیرمجاز |
| 🔒 | **Password Policy** | قوانین قابل تنظیم برای پیچیدگی و مدت اعتبار رمز عبور |
| 🚫 | **IP Ban/Whitelist** | مسدودسازی هوشمند IP با قابلیت Auto-Ban |
| 📱 | **TOTP/2FA** | احراز هویت دو مرحله‌ای با Google Authenticator |
| 📝 | **Audit Logs** | ثبت کامل رویدادهای امنیتی با قابلیت خروجی |

---

### 🌐 کلاستر و فدراسیون (Cluster & Federation)

```
┌──────────────────────────────────────────────────┐
│                   شبکه کلاستر                       │
│                                                   │
│   ┌──────────┐     ┌──────────┐     ┌──────────┐ │
│   │  گره ۱   │◄───│  گره ۲   │◄───│  گره ۳   │ │
│   │ (رهبر)   │───►│(دنباله‌رو)│───►│(دنباله‌رو)│ │
│   └──────────┘     └──────────┘     └──────────┘ │
│        │               │               │          │
│   ┌────┴────┐     ┌────┴────┐     ┌────┴────┐    │
│   │ ۱۰,۰۰۰  │     │ ۱۵,۰۰۰  │     │ ۱۲,۰۰۰  │    │
│   │ کاربر   │     │ کاربر   │     │ کاربر   │    │
│   └─────────┘     └─────────┘     └─────────┘    │
│                                                   │
│   ویژگی‌ها: ضربان قلب • انتخاب رهبر • توپولوژی    │
│   فدراسیون: همگام‌سازی کاربران • پلن‌ها • ترافیک  │
└──────────────────────────────────────────────────┘
```

#### قابلیت‌های کلاستر

| ویژگی | توضیحات |
|:------|:---------|
| **انتخاب رهبر (Leader Election)** | الگوریتم رأی‌گیری با اولویت قابل تنظیم |
| **ضربان قلب (Heartbeat)** | تشخیص خودکار گره‌های مرده و جایگزینی |
| **رمزنگاری mTLS** | TLS دوطرفه برای ارتباط امن بین گره‌ها |
| **توپولوژی زنده** | نمایش گرافیکی شبکه گره‌ها با Canvas 2D |
| **فدراسیون** | همگام‌سازی کاربران، پلن‌ها و ترافیک بین پنل‌ها |

---

### 📊 تحلیل و پایش (Analytics & Monitoring)

| قابلیت | توضیحات |
|:-------|:---------|
| **داشبورد زنده** | نمایش لحظه‌ای ترافیک، کاربران آنلاین و آمار سیستم |
| **Prometheus Metrics** | خروجی معیارها برای اتصال به Grafana |
| **WebSocket Streaming** | انتقال داده با تأخیر زیر یک ثانیه |
| **Health Check هوشمند** | پروب‌های قابل تنظیم با بازیابی خودکار |
| **ردیاب کاربران آنلاین** | پایش لحظه‌ای کاربران با ردیابی IP |
| **نقشه توپولوژی** | نمایش بصری گراف کلاسترها و اتصالات |
| **تحلیل ترافیک** | آمار مصرف به تفکیک کاربر و inbound |

---

### 💳 پرداخت و صورتحساب

| درگاه | پشتیبانی | ویژگی خاص |
|:-----:|:--------:|:----------|
| **زرین‌پال** | ✅ کامل | پرداخت ریالی، بازگشت خودکار |
| **NOWPayments** | ✅ کامل | ۱۰۰+ ارز دیجیتال، IPN خودکار |
| **کیف پول داخلی** | ✅ کامل | واریز، برداشت، انتقال موجودی |
| **پلن‌ها** | ✅ کامل | پلن‌های اشتراک با محدودیت پهنای باند |

---

### 📦 پشتیبان‌گیری و بازیابی

| روش | رمزنگاری | زمان‌بندی | مقصد |
|:---:|:--------:|:---------:|:----:|
| 🔐 AES-256-GCM | ✅ خودکار | درخواستی | محلی |
| ☁️ S3 | ✅ | خودکار (هر ۲۴ ساعت) | MinIO, AWS S3 |
| 🗄️ Google Drive | ✅ | خودکار (هر ۲۴ ساعت) | Google Drive |
| 🤖 تلگرام | ✅ | درخواستی | ربات تلگرام |

---

### 🎨 پورتال کاربری

- **داشبورد کاربر**: پنل سلف‌سرویس برای کاربران نهایی
- **لینک اشتراک**: تولید خودکار لینک‌های اشتراک (Xray JSON, Clash YAML, Sing-box)
- **نمودار ترافیک**: نمودارهای مصرف به صورت بلادرنگ
- **تیکت پشتیبانی**: سیستم تیکت یکپارچه با پاسخ‌دهی
- **ربات تلگرام**: سلف‌سرویس کاربران از طریق تلگرام

---

### 🐳 زیرساخت و DevOps

```
نصب بومی (Native)            نصب با Docker
┌────────────────────┐      ┌────────────────────┐
│  systemd service    │      │  docker compose     │
│  باینری Go          │      │  ├─ گره۱:۸۰۸۰      │
│  SQLite/PostgreSQL   │      │  ├─ گره۲:۸۰۸۱      │
│  Xray + Sing-box    │      │  └─ گره۳:۸۰۸۲      │
│  Caddy reverse proxy│      │  Caddy + mTLS      │
└────────────────────┘      └────────────────────┘
```

---

## 🚀 شروع سریع

### 🔥 نصب یک کلیکی

```bash
bash <(curl -fsSL https://github.com/iPmartNetwork/VortexUiPro/raw/master/install.sh)
```

### 🎯 نصب پیشرفته

```bash
# با پورت سفارشی + SSL
bash <(curl -fsSL https://github.com/iPmartNetwork/VortexUiPro/raw/master/install.sh) \
  --port 9090 \
  --ssl-domain panel.example.com

# کلاستر با Docker
bash <(curl -fsSL https://github.com/iPmartNetwork/VortexUiPro/raw/master/install.sh) --docker

# غیرتعاملی + بدون SSL
bash <(curl -fsSL https://github.com/iPmartNetwork/VortexUiPro/raw/master/install.sh) \
  --port 8080 \
  --skip-ssl
```

### 🐳 Docker Compose

```bash
git clone https://github.com/iPmartNetwork/VortexUiPro.git
cd VortexUiPro
docker compose -f deploy/compose.yml up -d
```

| گره | آدرس | توضیحات |
|:---:|:----:|:--------|
| گره-۱ | http://localhost:8080 | کاندید رهبر (اولویت بالا) |
| گره-۲ | http://localhost:8081 | دنباله‌رو |
| گره-۳ | http://localhost:8082 | دنباله‌رو |
| **ورود** | `admin` / `admin123` | پیش‌فرض |

### 📦 کامپایل دستی

```bash
git clone https://github.com/iPmartNetwork/VortexUiPro.git
cd VortexUiPro

# کامپایل بک‌اند
go build -o vortexuipro -ldflags="-s -w" ./cmd/panel

# کامپایل فرانت‌اند
cd web && npm ci && npm run build && cd ..

# تنظیم و اجرا
export VORTEX_HTTP_ADDR=:8080
export VORTEX_DB_TYPE=sqlite
export VORTEX_DATABASE_URL=./data/vortex.db
export VORTEX_JWT_SECRET=$(openssl rand -base64 32)

./vortexuipro
```

---

## 🏗️ معماری

```
┌────────────────────────────────────────────────────────┐
│              Caddy / Nginx Reverse Proxy                │
│           TLS • فایل‌های ایستا • محدودیت نرخ             │
└────────────────────────┬───────────────────────────────┘
                         │
┌────────────────────────┴───────────────────────────────┐
│                    Gin HTTP Router                      │
│          REST API • WebSocket • اشتراک                  │
└────────────────────────┬───────────────────────────────┘
                         │
┌────────────────────────┴───────────────────────────────┐
│                  لایه سرویس (Service Layer)             │
│  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌────┐ │
│  │کاربر │ │ورودی │ │خروجی │ │اشتراک│ │تحلیل │ │پشتی│ │
│  └──┬───┘ └──┬───┘ └──┬───┘ └──┬───┘ └──┬───┘ └──┬─┘ │
│  ┌──┴───┐ ┌──┴───┐ ┌──┴───┐ ┌──┴───┐ ┌──┴───┐ ┌──┴─┐ │
│  │تیکت │ │ضدسان│ │سلامت│ │کلاستر│ │فدراسیون│ │پلاگین│ │
│  └──────┘ └──────┘ └──────┘ └──────┘ └──────┘ └────┘ │
└────────────────────────┬───────────────────────────────┘
                         │
┌────────────────────────┴───────────────────────────────┐
│                  لایه موتور (Core Engine)              │
│  ┌──────────────────┐      ┌──────────────────┐       │
│  │   Xray gRPC API  │      │  Sing-box Config  │       │
│  │  آمار + مسیریابی  │      │   تولید کانفیگ    │       │
│  └────────┬─────────┘      └────────┬─────────┘       │
│           │                         │                   │
│  ┌────────┴─────────────────────────┴─────────┐        │
│  │        Engine Manager (چند هسته‌ای)          │        │
│  │  Hot Reload • Diff Config • Failover         │        │
│  └────────────────────────────────────────────┘        │
└────────────────────────┬───────────────────────────────┘
                         │
┌────────────────────────┴───────────────────────────────┐
│                  لایه داده (Data Layer)                │
│  ┌──────────────┐  ┌──────────────┐  ┌────────────┐  │
│  │   SQLite /   │  │  Prometheus  │  │   Redis    │  │
│  │  PostgreSQL  │  │   Metrics    │  │ (اختیاری)  │  │
│  └──────────────┘  └──────────────┘  └────────────┘  │
└────────────────────────────────────────────────────────┘
```

### 📁 ساختار پروژه

```
VortexUiPro/
├── cmd/panel/main.go         # نقطه ورود
├── internal/
│   ├── api/handlers/         # ۳۰+ هندلر
│   ├── api/middleware/       # احراز هویت، CORS، Rate Limiter
│   ├── cluster/              # کلاستر چند گره
│   ├── config/               # تنظیمات (Environment-based)
│   ├── core/xray/            # سرویس‌دهنده gRPC Xray
│   ├── core/singbox/         # سازنده کانفیگ Sing-box
│   ├── database/             # مدل‌های GORM (۵۵+)
│   ├── events/               # اتوبوس رویداد
│   ├── metrics/              # Prometheus collector
│   ├── monitor/              # موتور بررسی سلامت
│   ├── rbac/                 # کنترل دسترسی مبتنی بر نقش
│   └── service/              # لایه منطق کسب و کار (۴۰+ سرویس)
├── web/src/
│   ├── pages/                # ۳۹ صفحه React
│   ├── components/           # کامپوننت‌های UI
│   └── locales/              # ۱۰ زبان
├── deploy/                   # Docker, Caddy, systemd
├── install.sh               # اسکریپت نصب خودکار
└── vortexui.sh              # اسکریپت مدیریت
```

---

## 🔧 مدیریت

پس از نصب، از دستور `vortexui` استفاده کنید:

```bash
# کنترل سرویس
vortexui start                     # شروع پنل
vortexui stop                      # توقف پنل
vortexui restart                   # راه‌اندازی مجدد
vortexui status                    # وضعیت سرویس
vortexui logs [-f]                 # مشاهده لاگ‌ها

# بروزرسانی و پشتیبان
vortexui update                    # بروزرسانی پنل
vortexui backup                    # پشتیبان‌گیری کامل
vortexui restore /path/to/file     # بازیابی از پشتیبان

# تنظیمات
vortexui password                  # تغییر رمز ادمین
vortexui port 9090                 # تغییر پورت
vortexui cert                      # تنظیم گواهی SSL
vortexui info                      # اطلاعات نصب
```

---

## 📖 API Reference

### نقاط پایانی عمومی

```http
GET  /api/v1/health              # بررسی سلامت
GET  /metrics                    # معیارهای Prometheus
GET  /ws                         # اتصال WebSocket
GET  /sub/:clientId              # کانفیگ اشتراک
GET  /sub/:clientId/info         # اطلاعات اشتراک
GET  /sub/:clientId/link         # لینک اشتراک
GET  /sub/:clientId/share-links  # لینک‌های اشتراک
POST /api/v1/login               # احراز هویت
POST /api/v1/auth/refresh        # تازه‌سازی توکن
```

### نقاط پایانی محافظت شده

```http
# ادمین
GET  /api/v1/admin/users              # لیست کاربران
POST /api/v1/admin/users              # ایجاد کاربر
POST /api/v1/admin/users/:id/clients  # افزودن کلاینت

# Inbound
GET  /api/v1/inbounds                 # لیست ورودی‌ها
POST /api/v1/inbounds                 # ایجاد ورودی

# پایش
GET  /api/v1/monitor/online           # کاربران آنلاین
GET  /api/v1/monitor/activity         # فعالیت‌های اخیر

# پشتیبان
GET  /api/v1/backups                  # لیست پشتیبان‌ها
POST /api/v1/backups                  # ایجاد پشتیبان
POST /api/v1/backups/:id/restore      # بازیابی
```

---

## ⚙️ متغیرهای محیطی

```bash
# === سرور ===
VORTEX_HTTP_ADDR=:8080                     # آدرس HTTP
VORTEX_GRPC_ADDR=:50051                    # آدرس gRPC
VORTEX_JWT_SECRET=<32-کاراکتر>              # کلید امضای JWT
VORTEX_LOG_LEVEL=info                      # debug | info | warn | error

# === دیتابیس ===
VORTEX_DB_TYPE=sqlite                      # sqlite | postgres
VORTEX_DATABASE_URL=/etc/vortexuipro/data/vortex.db

# === هسته ===
VORTEX_CORE_BIN=/usr/local/bin/xray        # مسیر باینری Xray
VORTEX_CORE_API_PORT=10085

# === کلاستر ===
# VORTEX_CLUSTER_ENABLED=true
# VORTEX_CLUSTER_NODE_NAME=node-1
# VORTEX_CLUSTER_PEERS=node-2:1337,node-3:1337
```

---

## 🗺️ نقشه راه

### ✅ انجام شده (نسخه ۰.۰.۱)

| فاز | ویژگی | وضعیت |
|:---:|:------|:-----:|
| ۱-۲ | معماری هسته + Inbound/Outbound/Routing | ✅ |
| ۳ | مدیریت + RBAC + اشتراک | ✅ |
| ۴ | کلاستر + فدراسیون | ✅ |
| ۵ | ضد سانسور + تیکت‌ها | ✅ |
| ۶ | تنظیمات امنیتی + ایمیل | ✅ |
| ۷ | گروه‌های کاربر + عملیات دسته‌جمعی | ✅ |
| ۸ | ترمینال وب + لاگ زنده + WARP + TLS + پلاگین | ✅ |
| ۹ | تحلیل + پرداخت + پلن + کیف پول + ربات تلگرام | ✅ |
| ۱۰ | WebRTC + شبکه P2P | ✅ |
| ۱۱ | نمایشگر توپولوژی شبکه | ✅ |
| ۱۲ | بررسی سلامت هوشمند + بازیابی خودکار | ✅ |
| ۱۳ | چند زبانه i18n (۱۰ زبان + RTL) | ✅ |
| ۱۴ | پشتیبان‌گیری پیشرفته (AES-256 + S3 + GDrive) | ✅ |
| ۱۵ | Domain Fronting + DNS هوشمند + Docker | ✅ |
| ۱۶ | یکپارچه‌سازی واقعی Xray gRPC | ✅ |
| ۱۷ | بهبود سیستم اشتراک (۷ پروتکل + Clash + Sing-box) | ✅ |

### 🔮 برنامه آینده

| فاز | ویژگی | اولویت |
|:---:|:------|:------:|
| ۱۸ | تست Coverage (unit + integration + e2e) | 🔴 بالا |
| ۱۹ | لاگینگ ساختاریافته (zerolog/zap) | 🟡 متوسط |
| ۲۰ | ACME / Let's Encrypt خودکار | 🟡 متوسط |
| ۲۱ | WireGuard Mesh VPN | 🟢 پایین |

---

## 🤝 مشارکت

از مشارکت شما استقبال می‌کنیم!

```bash
git clone https://github.com/iPmartNetwork/VortexUiPro.git
cd VortexUiPro
go mod download
cd web && npm ci && cd ..
go run ./cmd/panel &              # بک‌اند
cd web && npm run dev &           # فرانت‌اند
```

---

## 📄 مجوز

این پروژه تحت مجوز **GNU General Public License v3.0** منتشر شده است.

```
Copyright (C) 2026 iPmartNetwork

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.
```

---

## 🙏 قدردانی

| فناوری | کاربرد |
|:-------|:-------|
| [Xray-core](https://github.com/XTLS/Xray-core) | موتور اصلی پروکسی |
| [Sing-box](https://github.com/SagerNet/sing-box) | پلتفرم پروکسی جهانی |
| [Gin](https://github.com/gin-gonic/gin) | فریم‌ورک HTTP |
| [GORM](https://gorm.io) | کتابخانه ORM |
| [React](https://reactjs.org/) | فریم‌ورک فرانت‌اند |
| [Tailwind CSS](https://tailwindcss.com/) | فریم‌ورک CSS |

---

<div align="center">

**ساخته شده با ❤️ توسط تیم VortexUiPro**

<a href="https://github.com/iPmartNetwork/VortexUiPro/issues">گزارش باگ</a> •
<a href="https://github.com/iPmartNetwork/VortexUiPro/discussions">پیشنهاد ویژگی</a>

<br/>
<sub>اگر VortexUiPro برای شما مفید است، با یک ⭐ در گیت‌هاب از ما حمایت کنید!</sub>

</div>

</div>
