-- ایجاد جدول کوپن‌ها
CREATE TABLE IF NOT EXISTS coupons (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE,
    percent DECIMAL(5,2) NOT NULL DEFAULT 0,
    max_discount DECIMAL(15,0) DEFAULT 0,
    usage_limit INT DEFAULT 0,
    used_count INT DEFAULT 0,
    expires_at TIMESTAMP WITH TIME ZONE,
    assignee_id INT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- اضافه کردن ستون‌های تخفیف به جدول سفارشات
ALTER TABLE orders ADD COLUMN IF NOT EXISTS coupon_code VARCHAR(50);
ALTER TABLE orders ADD COLUMN IF NOT EXISTS discount_amount DECIMAL(15,0) DEFAULT 0;