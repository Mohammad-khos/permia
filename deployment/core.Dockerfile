# =========================================================
# Stage 1: Builder (کامپایل کردن برنامه)
# =========================================================
FROM golang:1.25-alpine AS builder

# نصب ابزارهای مورد نیاز (مثل git)
RUN apk add --no-cache git

# تنظیم دایرکتوری کاری
WORKDIR /app

# کپی کردن فایل‌های وابستگی برای کش شدن لایه‌ها
# چون از go workspace استفاده می‌کنیم، باید کل ریشه را داشته باشیم
COPY go.work .
COPY go.work.sum .
COPY go.work .
COPY go.work.sum .

# پکیج pkg و Core کامل کپی شوند (چون go.sum دارند)
COPY pkg/go.mod pkg/go.sum ./pkg/
COPY services/core-service/go.mod services/core-service/go.sum ./services/core-service/

# ⚠️ اصلاحیه: برای بات و گیت‌وی فعلاً فقط go.mod را کپی کن (چون هنوز پکیجی ندارند و go.sum ندارند)
COPY services/bot-service/go.mod ./services/bot-service/
COPY services/api-gateway/go.mod ./services/api-gateway/

# دانلود ماژول‌ها (این لایه کش می‌شود تا بیلدهای بعدی سریع باشد)
RUN go mod download

# کپی کردن کل سورس کد
COPY . .

# بیلد کردن باینری (استاتیک و فشرده)
# -ldflags="-s -w" حجم فایل را کم می‌کند (حذف اطلاعات دیباگ)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o core-app services/core-service/cmd/main.go

# =========================================================
# Stage 2: Runner (اجرای برنامه)
# =========================================================
FROM alpine:latest

# نصب گواهی‌های امنیتی (برای HTTPS) و تنظیمات منطقه زمانی
RUN apk add --no-cache ca-certificates tzdata

# تنظیم منطقه زمانی به تهران (اختیاری ولی توصیه شده)
ENV TZ=Asia/Tehran

WORKDIR /root/

# کپی کردن باینری از مرحله قبل
COPY --from=builder /app/core-app .

# کپی کردن فایل‌های مایگریشن (چون برنامه فایل‌ها را می‌خواند)
# ساختار پوشه باید دقیقاً مثل انتظار برنامه باشد
COPY --from=builder /app/services/core-service/db/migrations ./db/migrations

# کپی کردن فایل .env (اختیاری - اگر بخواهید فایل را بخواند)
# اما در داکر معمولاً متغیرها را اینجکت می‌کنیم
# ما اینجا یک پوشه deployment فیک می‌سازیم تا منطق config.go راضی باشد
RUN mkdir deployment

# باز کردن پورت
EXPOSE 8080

# دستور اجرا
CMD ["./core-app"]