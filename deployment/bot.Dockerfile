# =========================================================
# Stage 1: Builder (ساخت فایل اجرایی)
# =========================================================
FROM golang:1.25.1-alpine AS builder

# نصب ابزارهای پایه
RUN apk add --no-cache git

# تنظیم دایرکتوری کاری در کانتینر
WORKDIR /app

# 1. کپی کردن فایل‌های ورک‌اسپیس (برای مدیریت وابستگی‌های مشترک)
COPY go.work .
COPY go.work.sum .

# 2. کپی کردن فایل‌های go.mod و go.sum تمام سرویس‌ها
# (این کار ضروری است چون go.work به همه این‌ها ارجاع دارد)
COPY pkg/go.mod pkg/go.sum ./pkg/
COPY services/bot-service/go.mod services/bot-service/go.sum ./services/bot-service/
COPY services/core-service/go.mod services/core-service/go.sum ./services/core-service/
COPY services/api-gateway/go.mod ./services/api-gateway/

# 3. دانلود وابستگی‌ها (این لایه کش می‌شود)
RUN go mod download

# 4. کپی کردن سورس کد (فقط چیزهایی که بات نیاز دارد)
COPY pkg/ ./pkg/
COPY services/bot-service/ ./services/bot-service/

# 5. بیلد کردن باینری (فشرده و استاتیک)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bot-app services/bot-service/cmd/main.go

# =========================================================
# Stage 2: Runner (اجرای نهایی - بسیار سبک)
# =========================================================
FROM alpine:latest

# نصب گواهی‌های امنیتی (برای ارتباط HTTPS با تلگرام و Core Service) و منطقه زمانی
RUN apk add --no-cache ca-certificates tzdata

# تنظیم منطقه زمانی به تهران
ENV TZ=Asia/Tehran

WORKDIR /root/

# کپی کردن فایل اجرایی از مرحله قبل
COPY --from=builder /app/bot-app .

# کپی کردن ساختار پوشه کانفیگ (اختیاری - برای نظم)
RUN mkdir -p deployment
# کپی فایل .env از روت پروژه به مسیر /deployment/.env داخل کانتینر
COPY deployment/.env /deployment/.env

# دستور اجرا
CMD ["./bot-app"]