# syntax=docker/dockerfile:1
# خط بالا برای فعال‌سازی ویژگی‌های پیشرفته داکر الزامی است

# =========================================================
# Stage 1: Builder
# =========================================================
FROM golang:1.25.1-alpine AS builder

# نصب ابزارهای ضروری
RUN apk add --no-cache git

WORKDIR /app

# 1. کپی کردن فایل‌های وابستگی (فقط ماژول‌های مرتبط)
COPY pkg/go.mod pkg/go.sum ./pkg/
COPY services/core-service/go.mod services/core-service/go.sum ./services/core-service/

# 2. ایجاد ورک‌اسپیس موقت
RUN go work init ./services/core-service ./pkg

# 3. دانلود ماژول‌ها با کش هوشمند (Mount Cache)
# این دستور معجزه می‌کند:
# target=/go/pkg/mod: محل ذخیره پکیج‌های دانلود شده را به کش داکر متصل می‌کند
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# 4. کپی کردن سورس کد
COPY . .

# 5. بیلد کردن باینری با کش بیلد
# target=/root/.cache/go-build: کش کامپایلر گو را هم نگه می‌دارد تا بیلد سریع‌تر شود
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -trimpath -o core-app services/core-service/cmd/main.go

# =========================================================
# Stage 2: Runner
# =========================================================
FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Tehran

WORKDIR /root/

COPY --from=builder /app/core-app .
# کپی فایل‌های مایگریشن (مسیر دقیق را چک کنید که با ساختار پروژه همخوانی داشته باشد)
COPY --from=builder /app/services/core-service/db/migrations ./db/migrations

# ایجاد پوشه کانفیگ
RUN mkdir -p deployment

EXPOSE 8080

CMD ["./core-app"]