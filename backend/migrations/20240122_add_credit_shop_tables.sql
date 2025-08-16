-- Migration: Add Credit Shop System Tables
-- Created: 2024-01-22
-- Purpose: Add tables for credit-based redemption system (Phase 2.1)

-- 积分商城分类表
CREATE TABLE IF NOT EXISTS credit_shop_categories (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    icon_url TEXT,
    parent_id UUID REFERENCES credit_shop_categories(id) ON DELETE CASCADE,
    sort_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 添加分类索引
CREATE INDEX IF NOT EXISTS idx_credit_shop_categories_parent_id ON credit_shop_categories(parent_id);
CREATE INDEX IF NOT EXISTS idx_credit_shop_categories_active_sort ON credit_shop_categories(is_active, sort_order);

-- 积分商城商品表
CREATE TABLE IF NOT EXISTS credit_shop_products (
    id UUID PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    short_desc VARCHAR(500),
    category VARCHAR(100),
    product_type VARCHAR(50) NOT NULL CHECK (product_type IN ('physical', 'virtual', 'service', 'voucher')),
    credit_price INTEGER NOT NULL,
    original_price DECIMAL(10,2),
    stock INTEGER DEFAULT 0,
    total_stock INTEGER DEFAULT 0,
    redeem_count INTEGER DEFAULT 0,
    image_url TEXT,
    images JSONB,
    tags JSONB,
    specifications JSONB,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('draft', 'active', 'inactive', 'sold_out', 'deleted')),
    is_featured BOOLEAN DEFAULT false,
    is_limited BOOLEAN DEFAULT false,
    limit_per_user INTEGER DEFAULT 0,
    priority INTEGER DEFAULT 0,
    valid_from TIMESTAMP,
    valid_to TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 添加商品索引
CREATE INDEX IF NOT EXISTS idx_credit_shop_products_status ON credit_shop_products(status);
CREATE INDEX IF NOT EXISTS idx_credit_shop_products_category ON credit_shop_products(category);
CREATE INDEX IF NOT EXISTS idx_credit_shop_products_type ON credit_shop_products(product_type);
CREATE INDEX IF NOT EXISTS idx_credit_shop_products_featured ON credit_shop_products(is_featured) WHERE is_featured = true;
CREATE INDEX IF NOT EXISTS idx_credit_shop_products_price ON credit_shop_products(credit_price);
CREATE INDEX IF NOT EXISTS idx_credit_shop_products_stock ON credit_shop_products(stock) WHERE stock > 0;
CREATE INDEX IF NOT EXISTS idx_credit_shop_products_priority ON credit_shop_products(priority DESC, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_credit_shop_products_valid_period ON credit_shop_products(valid_from, valid_to);
CREATE INDEX IF NOT EXISTS idx_credit_shop_products_deleted ON credit_shop_products(deleted_at) WHERE deleted_at IS NOT NULL;

-- 积分购物车表
CREATE TABLE IF NOT EXISTS credit_carts (
    id UUID PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    total_items INTEGER DEFAULT 0,
    total_credits INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 添加购物车索引
CREATE INDEX IF NOT EXISTS idx_credit_carts_user_id ON credit_carts(user_id);

-- 积分购物车项目表
CREATE TABLE IF NOT EXISTS credit_cart_items (
    id UUID PRIMARY KEY,
    cart_id UUID NOT NULL REFERENCES credit_carts(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES credit_shop_products(id) ON DELETE CASCADE,
    quantity INTEGER NOT NULL DEFAULT 1,
    credit_price INTEGER NOT NULL,
    subtotal INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 添加购物车项目索引
CREATE INDEX IF NOT EXISTS idx_credit_cart_items_cart_id ON credit_cart_items(cart_id);
CREATE INDEX IF NOT EXISTS idx_credit_cart_items_product_id ON credit_cart_items(product_id);

-- 积分兑换订单表
CREATE TABLE IF NOT EXISTS credit_redemptions (
    id UUID PRIMARY KEY,
    redemption_no VARCHAR(50) UNIQUE NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    product_id UUID NOT NULL REFERENCES credit_shop_products(id) ON DELETE RESTRICT,
    quantity INTEGER NOT NULL DEFAULT 1,
    credit_price INTEGER NOT NULL,
    total_credits INTEGER NOT NULL,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'confirmed', 'processing', 'shipped', 'delivered', 'completed', 'cancelled', 'refunded')),
    delivery_info JSONB,
    redemption_code VARCHAR(100),
    tracking_number VARCHAR(100),
    notes TEXT,
    processed_at TIMESTAMP,
    shipped_at TIMESTAMP,
    delivered_at TIMESTAMP,
    completed_at TIMESTAMP,
    cancelled_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 添加兑换订单索引
CREATE INDEX IF NOT EXISTS idx_credit_redemptions_user_id ON credit_redemptions(user_id);
CREATE INDEX IF NOT EXISTS idx_credit_redemptions_product_id ON credit_redemptions(product_id);
CREATE INDEX IF NOT EXISTS idx_credit_redemptions_status ON credit_redemptions(status);
CREATE INDEX IF NOT EXISTS idx_credit_redemptions_redemption_no ON credit_redemptions(redemption_no);
CREATE INDEX IF NOT EXISTS idx_credit_redemptions_created_at ON credit_redemptions(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_credit_redemptions_user_status ON credit_redemptions(user_id, status);

-- 用户兑换历史统计表
CREATE TABLE IF NOT EXISTS user_redemption_histories (
    id UUID PRIMARY KEY,
    user_id VARCHAR(36) UNIQUE NOT NULL,
    total_redemptions INTEGER DEFAULT 0,
    total_credits_used INTEGER DEFAULT 0,
    last_redemption_at TIMESTAMP,
    favorite_category VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 添加兑换历史索引
CREATE INDEX IF NOT EXISTS idx_user_redemption_histories_user_id ON user_redemption_histories(user_id);
CREATE INDEX IF NOT EXISTS idx_user_redemption_histories_total_redemptions ON user_redemption_histories(total_redemptions DESC);
CREATE INDEX IF NOT EXISTS idx_user_redemption_histories_last_redemption ON user_redemption_histories(last_redemption_at DESC);

-- 积分商城配置表
CREATE TABLE IF NOT EXISTS credit_shop_configs (
    id UUID PRIMARY KEY,
    key VARCHAR(100) UNIQUE NOT NULL,
    value TEXT NOT NULL,
    description TEXT,
    category VARCHAR(50),
    is_editable BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 添加配置索引
CREATE INDEX IF NOT EXISTS idx_credit_shop_configs_key ON credit_shop_configs(key);
CREATE INDEX IF NOT EXISTS idx_credit_shop_configs_category ON credit_shop_configs(category);

-- 添加外键约束（如果用户表存在）
DO $$
BEGIN
    -- 检查users表是否存在
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'users') THEN
        -- 为credit_carts添加外键
        IF NOT EXISTS (
            SELECT constraint_name 
            FROM information_schema.table_constraints 
            WHERE table_name = 'credit_carts' 
            AND constraint_name = 'fk_credit_carts_user_id'
        ) THEN
            ALTER TABLE credit_carts 
            ADD CONSTRAINT fk_credit_carts_user_id 
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
        END IF;
        
        -- 为credit_redemptions添加外键
        IF NOT EXISTS (
            SELECT constraint_name 
            FROM information_schema.table_constraints 
            WHERE table_name = 'credit_redemptions' 
            AND constraint_name = 'fk_credit_redemptions_user_id'
        ) THEN
            ALTER TABLE credit_redemptions 
            ADD CONSTRAINT fk_credit_redemptions_user_id 
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
        END IF;
        
        -- 为user_redemption_histories添加外键
        IF NOT EXISTS (
            SELECT constraint_name 
            FROM information_schema.table_constraints 
            WHERE table_name = 'user_redemption_histories' 
            AND constraint_name = 'fk_user_redemption_histories_user_id'
        ) THEN
            ALTER TABLE user_redemption_histories 
            ADD CONSTRAINT fk_user_redemption_histories_user_id 
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
        END IF;
    END IF;
END $$;

-- 创建触发器自动更新 updated_at
CREATE OR REPLACE FUNCTION update_credit_shop_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为各表添加触发器
DROP TRIGGER IF EXISTS update_credit_shop_categories_updated_at ON credit_shop_categories;
CREATE TRIGGER update_credit_shop_categories_updated_at
    BEFORE UPDATE ON credit_shop_categories
    FOR EACH ROW
    EXECUTE FUNCTION update_credit_shop_updated_at_column();

DROP TRIGGER IF EXISTS update_credit_shop_products_updated_at ON credit_shop_products;
CREATE TRIGGER update_credit_shop_products_updated_at
    BEFORE UPDATE ON credit_shop_products
    FOR EACH ROW
    EXECUTE FUNCTION update_credit_shop_updated_at_column();

DROP TRIGGER IF EXISTS update_credit_carts_updated_at ON credit_carts;
CREATE TRIGGER update_credit_carts_updated_at
    BEFORE UPDATE ON credit_carts
    FOR EACH ROW
    EXECUTE FUNCTION update_credit_shop_updated_at_column();

DROP TRIGGER IF EXISTS update_credit_cart_items_updated_at ON credit_cart_items;
CREATE TRIGGER update_credit_cart_items_updated_at
    BEFORE UPDATE ON credit_cart_items
    FOR EACH ROW
    EXECUTE FUNCTION update_credit_shop_updated_at_column();

DROP TRIGGER IF EXISTS update_credit_redemptions_updated_at ON credit_redemptions;
CREATE TRIGGER update_credit_redemptions_updated_at
    BEFORE UPDATE ON credit_redemptions
    FOR EACH ROW
    EXECUTE FUNCTION update_credit_shop_updated_at_column();

DROP TRIGGER IF EXISTS update_user_redemption_histories_updated_at ON user_redemption_histories;
CREATE TRIGGER update_user_redemption_histories_updated_at
    BEFORE UPDATE ON user_redemption_histories
    FOR EACH ROW
    EXECUTE FUNCTION update_credit_shop_updated_at_column();

DROP TRIGGER IF EXISTS update_credit_shop_configs_updated_at ON credit_shop_configs;
CREATE TRIGGER update_credit_shop_configs_updated_at
    BEFORE UPDATE ON credit_shop_configs
    FOR EACH ROW
    EXECUTE FUNCTION update_credit_shop_updated_at_column();

-- 插入默认积分商城分类
INSERT INTO credit_shop_categories (id, name, description, icon_url, sort_order) VALUES
(gen_random_uuid(), '实物商品', '可配送的实体商品', '/icons/physical.svg', 1),
(gen_random_uuid(), '虚拟商品', '数字化商品和服务', '/icons/virtual.svg', 2),
(gen_random_uuid(), '优惠券', '各种优惠券和折扣码', '/icons/voucher.svg', 3),
(gen_random_uuid(), '服务类', '各种服务项目', '/icons/service.svg', 4);

-- 插入默认积分商城配置
INSERT INTO credit_shop_configs (id, key, value, description, category) VALUES
(gen_random_uuid(), 'shop_enabled', 'true', '是否启用积分商城', 'system'),
(gen_random_uuid(), 'min_redemption_credits', '10', '最低兑换积分要求', 'redemption'),
(gen_random_uuid(), 'max_cart_items', '20', '购物车最大商品数量', 'cart'),
(gen_random_uuid(), 'default_shipping_fee', '0', '默认配送费用（积分）', 'shipping'),
(gen_random_uuid(), 'auto_confirm_virtual', 'true', '虚拟商品是否自动确认', 'processing'),
(gen_random_uuid(), 'refund_window_days', '7', '退款申请窗口期（天）', 'refund');

-- 插入示例积分商城商品
INSERT INTO credit_shop_products (
    id, name, description, short_desc, category, product_type, 
    credit_price, original_price, stock, total_stock, 
    image_url, status, is_featured, priority
) VALUES
-- 实物商品
(gen_random_uuid(), 'OpenPenPal定制笔记本', '高质量定制笔记本，印有OpenPenPal logo', '精美定制笔记本', '文具用品', 'physical', 200, 25.00, 50, 100, '/images/notebook.jpg', 'active', true, 100),
(gen_random_uuid(), 'OpenPenPal钢笔', '经典钢笔，适合书写信件', '经典书写钢笔', '文具用品', 'physical', 500, 80.00, 20, 50, '/images/pen.jpg', 'active', true, 90),
(gen_random_uuid(), 'OpenPenPal信纸套装', '精美信纸套装，包含信封', '精美信纸套装', '文具用品', 'physical', 150, 20.00, 100, 200, '/images/letter-set.jpg', 'active', false, 80),

-- 虚拟商品  
(gen_random_uuid(), '专属头像框', '独特的个人资料头像框', '个性化头像装饰', '装饰道具', 'virtual', 100, 0.00, 999, 999, '/images/avatar-frame.png', 'active', true, 70),
(gen_random_uuid(), 'VIP会员1个月', '享受VIP特权服务1个月', '1个月VIP权限', '会员服务', 'virtual', 300, 15.00, 999, 999, '/images/vip-1month.png', 'active', true, 85),
(gen_random_uuid(), 'AI写信助手高级版', '解锁AI写信助手的高级功能', '高级AI写信功能', '工具服务', 'virtual', 250, 12.00, 999, 999, '/images/ai-premium.png', 'active', false, 60),

-- 优惠券
(gen_random_uuid(), '商城9折优惠券', '传统商城商品9折优惠', '商城购物9折优惠', '优惠券', 'voucher', 50, 5.00, 200, 500, '/images/coupon-10off.png', 'active', false, 50),
(gen_random_uuid(), '免费邮寄优惠券', '一次免费邮寄服务', '免邮寄费用券', '优惠券', 'voucher', 80, 8.00, 150, 300, '/images/free-shipping.png', 'active', false, 55),

-- 服务类
(gen_random_uuid(), '专属信件定制服务', '专业团队帮您定制特别的信件', '个性化信件定制', '定制服务', 'service', 800, 100.00, 10, 20, '/images/custom-letter.jpg', 'active', true, 95),
(gen_random_uuid(), '书法代写服务', '专业书法师代写您的信件', '专业书法代写', '写作服务', 'service', 400, 50.00, 30, 50, '/images/calligraphy.jpg', 'active', false, 65);

-- 添加商品规格和标签示例
UPDATE credit_shop_products 
SET 
    specifications = '{"size": "A5", "pages": 200, "material": "高质量纸张"}',
    tags = '["文具", "定制", "笔记本", "学习"]'
WHERE name = 'OpenPenPal定制笔记本';

UPDATE credit_shop_products 
SET 
    specifications = '{"type": "钢笔", "color": "黑色", "brand": "OpenPenPal"}',
    tags = '["钢笔", "书写", "文具", "礼品"]'
WHERE name = 'OpenPenPal钢笔';

UPDATE credit_shop_products 
SET 
    specifications = '{"duration": "30天", "features": ["AI优先", "专属客服", "高级模板"]}',
    tags = '["VIP", "会员", "特权", "服务"]'
WHERE name = 'VIP会员1个月';