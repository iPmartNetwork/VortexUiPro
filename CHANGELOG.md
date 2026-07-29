# 📋 Changelog

> **VortexUiPro** — پروکسی پنل نسل جدید با ترکیب قدرت VortexUI + Heimdall

---

## [0.0.1] — 2025-07-29 🎉

### 🚀 فاز ۱ — Core Engine
- ساختار پروتکل‌های کامل: VMess, VLESS, Trojan, Shadowsocks, WireGuard, SOCKS, HTTP
- هسته دوگانه **Xray + Sing-Box** با قابلیت سوییچ خودکار
- **SQLite** repository با GORM + migration خودکار
- **PostgreSQL** migration پشتیبانی کامل
- **gRPC real implementation** برای ارتباط با Xray core
- معماری **Clean Architecture** (Domain → Repository → Service → Handler)

### 🛡️ فاز ۲ — امنیت و احراز هویت
- **JWT** توکن با refresh/access مجزا
- **RBAC** سه سطح: `super_admin` → `admin` → `reseller`
- **TOTP 2FA** دو مرحله‌ای
- **Security Headers** خودکار
- **Rate Limiting** هوشمند با تشخیص زون (auth, api, subscription, default)
- **Session Management** با device tracking
- **Audit Log** کامل

### 👑 فاز ۳ — مدیریت کاربران
- مدیریت کاربران با گروه‌بندی
- Bulk operations (فعال/غیرفعال, حذف, انتقال گروهی)
- **Client Template** برای تنظیمات پیش‌فرض
- **Client Activity Monitoring** لحظه‌ای
- **User Notification** از طریق Telegram
- وضعیت آنلاین کاربران + traffic لحظه‌ای

### 💰 فاز ۴ — مالی و ریسلر
- **Wallet System** با شارژ و برداشت
- **Payment Gateways**: NowPayments + Zarinpal
- **Reseller System** کامل با commission و فروش
- **Plans & Orders** با proof image
- **Billing** با قابلیت اشتراک (subscription)
- **Quota Notifications** خودکار

### 🎨 فاز ۵ — طراحی UI
- **Dark/Light Mode** با system detection
- **40+ صفحه** با React + TypeScript + Tailwind CSS
- **Responsive Design** برای موبایل و دسکتاپ
- **Animated Sidebar** با layout transitions
- **i18n** 10 زبان: EN, FA, ES, RU, ZH, AR, DE, FR, PT, TR
- **RTL Support** کامل برای فارسی و عربی

### 🌐 فاز ۶ — اشتراک و کانفیگ
- **Subscription System** با Clash, Sing-Box, JSON, Outline, Xray
- **Link Generation** (VMess, VLESS, Trojan, Shadowsocks, WireGuard)
- **QR Code** برای اشتراک گذاری
- **Config Versions** با تاریخچه تغییرات
- **Subscription Profiling** با اولویت‌بندی

### 🔄 فاز ۷ — Routing و Optimization
- **Routing System** پیشرفته با Rule Packs
- **Routing Templates** با مدیریت گروهی
- **Smart Config** با بهینه‌سازی خودکار
- **TLS Tricks Suite** (Fragment, Padding, TLS Hello)
- **WARP+ Outbound** یکپارچه
- **Anti-Censorship Suite** کامل: Domain Fronting, SNI Obfuscation, Split HTTP, PT

### 🌍 فاز ۸ — شبکه و پروتکل
- **WireGuard Mesh VPN** با full mesh
- **Multi Protocol Groups** با load balancing
- **Port Conflict Detection** هوشمند
- **Inbound Migration** بدون قطعی
- **WARP 2.0 Support**
- **MTProto** پروتکل تلگرام

### 📊 فاز ۹ — مانیتورینگ و تحلیل
- **Real-time Monitoring** (CPU, RAM, Disk, Traffic)
- **Metrics Dashboard** با Recharts
- **Geographical Analytics** با نقشه
- **Telegram Bot** کامل (Clients + Admin)
- **Web Terminal (SSH Console)**
- **Live Log Streaming** با رنگ‌بندی

### 🔧 فاز ۱۰ — P2P و ارتباطات
- **WebRTC Direct Connections** با STUN/TURN
- **P2P Mesh** برای نودها
- **NAT Traversal** خودکار
- **Plugin System** با معماری extensible
- **Network Topology Visualizer** با Canvas

### 💖 فاز ۱۱ — سلامت و بازیابی
- **Smart Health Check** با probing چندگانه
- **Auto-Recovery** هوشمند نودها
- **Health Dashboard** با history charts

### 🔐 فاز ۱۲ — امنیت پیشرفته
- Advanced Security Settings
- Fail2Ban Integration
- GeoIP Filtering + Clean IP Scan
- IP Limit Policies

### 🌍 فاز ۱۳ — زبان و بین‌المللی
- i18n با 10 زبان کامل
- Auto-translate pipeline
- RTL کامل برای عربی و فارسی

### 💾 فاز ۱۴ — Backup
- **Backup & Restore** خودکار و دستی
- **S3 / Google Drive** remote storage
- **Telegram** دانلود مستقیم backup
- **Encrypted Backups** با AES-256
- Schedule-based خودکار

### 🔬 فاز ۱۵ — ابزارهای پیشرفته
- **Xray gRPC Real Integration** (hot-reload)
- **Code Splitting** (1.2MB → 362KB bundle)
- **CORS محدود** (قابل تنظیم با CORS_ORIGIN)
- **Rate Limiting هوشمند** (۴ زون مجزا)

### ☁️ فاز ۱۶ — استقرار
- **Docker Compose** production-ready
- **Dockerfile** multi-stage
- **Helm Chart** کوبرنتیز (۱۷ فایل)
- **CI/CD Pipeline** GitHub Actions (build, test, lint, deploy)
- **Prometheus Metrics** + Grafana
- **Production Audit** کامل

---

## پیش از انتشار

| نسخه | تاریخ | وضعیت |
|:----:|:----:|:------|
| v0.0.1 | 2025-07-29 | ✅ توسعه کامل |
| v0.1.0-beta | — | 🔄 تست سرور واقعی |
| v1.0.0-stable | — | 🎯 انتشار پایدار |

---

> **توسعه‌دهنده:** iPmart Network
> **مخزن:** [github.com/iPmartNetwork/VortexUiPro](https://github.com/iPmartNetwork/VortexUiPro)
