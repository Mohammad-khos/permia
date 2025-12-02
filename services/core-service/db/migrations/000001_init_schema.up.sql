-- جدول کاربران
CREATE TABLE "users" (
  "id" bigserial PRIMARY KEY,
  "telegram_id" bigint UNIQUE NOT NULL,
  "username" varchar(100),
  "first_name" varchar(100),
  "last_name" varchar(100),
  "wallet_balance" decimal(15, 0) DEFAULT 0,
  "total_spent" decimal(15, 0) DEFAULT 0,
  "referral_code" varchar(20) UNIQUE,
  "referred_by" bigint,
  "total_referrals" int DEFAULT 0,
  "is_banned" boolean DEFAULT false,
  "created_at" timestamptz DEFAULT (now()),
  "updated_at" timestamptz DEFAULT (now())
);

-- جدول محصولات (کاتالوگ)
CREATE TABLE "products" (
  "id" bigserial PRIMARY KEY,
  "sku" varchar(50) UNIQUE NOT NULL,
  "category" varchar(50) NOT NULL,
  "title" varchar(200) NOT NULL,
  "description" text,
  "price" decimal(15, 0) NOT NULL,
  "type" varchar(50) NOT NULL,
  "capacity" int DEFAULT 1,
  "is_active" boolean DEFAULT true,
  "display_order" int DEFAULT 0,
  "created_at" timestamptz DEFAULT (now()),
  "updated_at" timestamptz DEFAULT (now())
);

-- جدول انبار اکانت‌ها
CREATE TABLE "account_inventories" (
  "id" bigserial PRIMARY KEY,
  "product_sku" varchar(50) NOT NULL,
  "email" varchar(200),
  "password" text,
  "additional" text,
  "max_users" int DEFAULT 1,
  "current_users" int DEFAULT 0,
  "status" varchar(20) DEFAULT 'AVAILABLE',
  "purchased_at" timestamptz,
  "created_at" timestamptz DEFAULT (now()),
  "updated_at" timestamptz DEFAULT (now())
);

-- جدول سفارشات
CREATE TABLE "orders" (
  "id" bigserial PRIMARY KEY,
  "order_number" varchar(50) UNIQUE NOT NULL,
  "user_id" bigint NOT NULL,
  "product_id" bigint NOT NULL,
  "account_id" bigint,
  "amount" decimal(15, 0) NOT NULL,
  "status" varchar(20) DEFAULT 'PENDING',
  "payment_method" varchar(50),
  "delivered_data" text,
  "created_at" timestamptz DEFAULT (now()),
  "delivered_at" timestamptz
);

-- ایجاد روابط (Foreign Keys)
ALTER TABLE "orders" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");
ALTER TABLE "orders" ADD FOREIGN KEY ("product_id") REFERENCES "products" ("id");
ALTER TABLE "orders" ADD FOREIGN KEY ("account_id") REFERENCES "account_inventories" ("id");
ALTER TABLE "users" ADD FOREIGN KEY ("referred_by") REFERENCES "users" ("id");

-- ایندکس‌گذاری برای سرعت
CREATE INDEX ON "users" ("telegram_id");
CREATE INDEX ON "account_inventories" ("product_sku", "status");
CREATE INDEX ON "orders" ("user_id");