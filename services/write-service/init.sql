-- OpenPenPal Write Service 数据库初始化脚本

-- 创建letters表
CREATE TABLE IF NOT EXISTS letters (
    id VARCHAR(20) PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    sender_id VARCHAR(50) NOT NULL,
    sender_nickname VARCHAR(100),
    receiver_hint VARCHAR(200),
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    priority VARCHAR(20) NOT NULL DEFAULT 'normal',
    anonymous BOOLEAN NOT NULL DEFAULT FALSE,
    delivery_instructions TEXT,
    read_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建letters表索引（性能优化）
CREATE INDEX IF NOT EXISTS idx_letters_sender_id ON letters(sender_id);
CREATE INDEX IF NOT EXISTS idx_letters_status ON letters(status);
CREATE INDEX IF NOT EXISTS idx_letters_created_at ON letters(created_at);

-- 复合索引：用户+状态查询优化
CREATE INDEX IF NOT EXISTS idx_letters_sender_status ON letters(sender_id, status);

-- 复合索引：状态+创建时间查询优化（管理员仪表板）
CREATE INDEX IF NOT EXISTS idx_letters_status_created ON letters(status, created_at DESC);

-- 复合索引：用户+创建时间查询优化（用户信件列表）
CREATE INDEX IF NOT EXISTS idx_letters_sender_created ON letters(sender_id, created_at DESC);

-- 部分索引：只为热点状态创建索引
CREATE INDEX IF NOT EXISTS idx_letters_active_status ON letters(id, sender_id) 
WHERE status IN ('draft', 'generated', 'collected', 'in_transit');

-- 文本搜索索引（如果需要支持标题搜索）
CREATE INDEX IF NOT EXISTS idx_letters_title_search ON letters USING gin(to_tsvector('simple', title));

-- 创建read_logs表
CREATE TABLE IF NOT EXISTS read_logs (
    id SERIAL PRIMARY KEY,
    letter_id VARCHAR(20) NOT NULL REFERENCES letters(id) ON DELETE CASCADE,
    reader_ip VARCHAR(45),
    reader_user_agent TEXT,
    reader_location VARCHAR(200),
    read_duration INTEGER,
    is_complete_read BOOLEAN DEFAULT TRUE,
    referer VARCHAR(500),
    device_info TEXT,
    read_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建read_logs表索引（性能优化）
CREATE INDEX IF NOT EXISTS idx_read_logs_letter_id ON read_logs(letter_id);
CREATE INDEX IF NOT EXISTS idx_read_logs_read_at ON read_logs(read_at DESC);
CREATE INDEX IF NOT EXISTS idx_read_logs_ip ON read_logs(reader_ip);

-- 复合索引：信件+阅读时间查询优化
CREATE INDEX IF NOT EXISTS idx_read_logs_letter_time ON read_logs(letter_id, read_at DESC);

-- 复合索引：IP+时间查询优化（防刷统计）
CREATE INDEX IF NOT EXISTS idx_read_logs_ip_time ON read_logs(reader_ip, read_at DESC);

-- 部分索引：只为完整阅读创建索引
CREATE INDEX IF NOT EXISTS idx_read_logs_complete ON read_logs(letter_id, read_at) 
WHERE is_complete_read = true;

-- 创建更新时间触发器
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_letters_updated_at 
    BEFORE UPDATE ON letters 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- 插入测试数据
INSERT INTO letters (
    id, title, content, sender_id, sender_nickname, 
    receiver_hint, status, priority, anonymous
) VALUES 
(
    'OP1234567890', 
    '测试信件', 
    '这是一封测试信件的内容。', 
    'user123', 
    '测试用户', 
    '北京大学宿舍楼',
    'draft',
    'normal',
    false
) ON CONFLICT (id) DO NOTHING;

-- 创建广场帖子表
CREATE TABLE IF NOT EXISTS plaza_posts (
    id VARCHAR(20) PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    excerpt VARCHAR(500),
    author_id VARCHAR(50) NOT NULL,
    author_nickname VARCHAR(100),
    category VARCHAR(20) NOT NULL DEFAULT 'others',
    tags VARCHAR(200),
    status VARCHAR(20) NOT NULL DEFAULT 'published',
    allow_comments BOOLEAN NOT NULL DEFAULT TRUE,
    anonymous BOOLEAN NOT NULL DEFAULT FALSE,
    view_count INTEGER NOT NULL DEFAULT 0,
    like_count INTEGER NOT NULL DEFAULT 0,
    comment_count INTEGER NOT NULL DEFAULT 0,
    favorite_count INTEGER NOT NULL DEFAULT 0,
    letter_id VARCHAR(20) REFERENCES letters(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    published_at TIMESTAMP WITH TIME ZONE
);

-- 创建广场帖子索引
CREATE INDEX IF NOT EXISTS idx_plaza_posts_author_id ON plaza_posts(author_id);
CREATE INDEX IF NOT EXISTS idx_plaza_posts_category ON plaza_posts(category);
CREATE INDEX IF NOT EXISTS idx_plaza_posts_status ON plaza_posts(status);
CREATE INDEX IF NOT EXISTS idx_plaza_posts_created_at ON plaza_posts(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_plaza_posts_published_at ON plaza_posts(published_at DESC);

-- 复合索引：状态+分类查询优化
CREATE INDEX IF NOT EXISTS idx_plaza_posts_status_category ON plaza_posts(status, category);

-- 复合索引：作者+状态查询优化
CREATE INDEX IF NOT EXISTS idx_plaza_posts_author_status ON plaza_posts(author_id, status);

-- 热度排序索引
CREATE INDEX IF NOT EXISTS idx_plaza_posts_hot ON plaza_posts(
    (like_count * 3 + comment_count * 2 + view_count) DESC, 
    created_at DESC
) WHERE status = 'published';

-- 文本搜索索引
CREATE INDEX IF NOT EXISTS idx_plaza_posts_title_search ON plaza_posts USING gin(to_tsvector('simple', title));
CREATE INDEX IF NOT EXISTS idx_plaza_posts_content_search ON plaza_posts USING gin(to_tsvector('simple', content));
CREATE INDEX IF NOT EXISTS idx_plaza_posts_tags_search ON plaza_posts USING gin(to_tsvector('simple', tags));

-- 创建广场点赞表
CREATE TABLE IF NOT EXISTS plaza_likes (
    post_id VARCHAR(20) NOT NULL REFERENCES plaza_posts(id) ON DELETE CASCADE,
    user_id VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (post_id, user_id)
);

-- 创建广场点赞索引
CREATE INDEX IF NOT EXISTS idx_plaza_likes_post_id ON plaza_likes(post_id);
CREATE INDEX IF NOT EXISTS idx_plaza_likes_user_id ON plaza_likes(user_id);

-- 创建广场评论表
CREATE TABLE IF NOT EXISTS plaza_comments (
    id VARCHAR(20) PRIMARY KEY,
    post_id VARCHAR(20) NOT NULL REFERENCES plaza_posts(id) ON DELETE CASCADE,
    user_id VARCHAR(50) NOT NULL,
    user_nickname VARCHAR(100),
    content TEXT NOT NULL,
    parent_id VARCHAR(20) REFERENCES plaza_comments(id) ON DELETE CASCADE,
    reply_to_user VARCHAR(100),
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    like_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建广场评论索引
CREATE INDEX IF NOT EXISTS idx_plaza_comments_post_id ON plaza_comments(post_id);
CREATE INDEX IF NOT EXISTS idx_plaza_comments_user_id ON plaza_comments(user_id);
CREATE INDEX IF NOT EXISTS idx_plaza_comments_parent_id ON plaza_comments(parent_id);
CREATE INDEX IF NOT EXISTS idx_plaza_comments_created_at ON plaza_comments(created_at DESC);

-- 复合索引：帖子+时间查询优化
CREATE INDEX IF NOT EXISTS idx_plaza_comments_post_time ON plaza_comments(post_id, created_at DESC);

-- 创建广场分类表
CREATE TABLE IF NOT EXISTS plaza_categories (
    id VARCHAR(20) PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    description VARCHAR(200),
    icon VARCHAR(50),
    color VARCHAR(20),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    post_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建广场分类索引
CREATE INDEX IF NOT EXISTS idx_plaza_categories_is_active ON plaza_categories(is_active);
CREATE INDEX IF NOT EXISTS idx_plaza_categories_sort_order ON plaza_categories(sort_order);

-- 创建更新时间触发器（广场相关表）
CREATE TRIGGER update_plaza_posts_updated_at 
    BEFORE UPDATE ON plaza_posts 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_plaza_comments_updated_at 
    BEFORE UPDATE ON plaza_comments 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_plaza_categories_updated_at 
    BEFORE UPDATE ON plaza_categories 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- 插入默认分类数据
INSERT INTO plaza_categories (id, name, description, icon, color, sort_order) VALUES 
('letters', '信件作品', '基于真实信件创作的内容', '📮', '#3B82F6', 1),
('poetry', '诗歌', '诗词歌赋相关创作', '🌸', '#EC4899', 2),
('prose', '散文', '散文随笔类作品', '📝', '#10B981', 3),
('stories', '故事', '小说故事类创作', '📚', '#F59E0B', 4),
('thoughts', '感想', '心情感悟分享', '💭', '#8B5CF6', 5),
('others', '其他', '其他类型的创作', '🎨', '#6B7280', 6)
ON CONFLICT (id) DO NOTHING;

-- 创建博物馆信件表
CREATE TABLE IF NOT EXISTS museum_letters (
    id VARCHAR(20) PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    summary VARCHAR(500),
    original_author VARCHAR(100),
    original_recipient VARCHAR(100),
    historical_date TIMESTAMP WITH TIME ZONE,
    era VARCHAR(20) NOT NULL DEFAULT 'present',
    location VARCHAR(200),
    category VARCHAR(50) NOT NULL,
    tags VARCHAR(300),
    language VARCHAR(10) DEFAULT 'zh',
    source_type VARCHAR(20) NOT NULL,
    source_description TEXT,
    contributor_id VARCHAR(50),
    contributor_name VARCHAR(100),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    reviewer_id VARCHAR(50),
    review_note TEXT,
    reviewed_at TIMESTAMP WITH TIME ZONE,
    is_featured BOOLEAN NOT NULL DEFAULT FALSE,
    display_order INTEGER NOT NULL DEFAULT 0,
    featured_until TIMESTAMP WITH TIME ZONE,
    view_count INTEGER NOT NULL DEFAULT 0,
    favorite_count INTEGER NOT NULL DEFAULT 0,
    share_count INTEGER NOT NULL DEFAULT 0,
    rating_avg FLOAT NOT NULL DEFAULT 0.0,
    rating_count INTEGER NOT NULL DEFAULT 0,
    letter_id VARCHAR(20) REFERENCES letters(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建博物馆信件索引
CREATE INDEX IF NOT EXISTS idx_museum_letters_era ON museum_letters(era);
CREATE INDEX IF NOT EXISTS idx_museum_letters_category ON museum_letters(category);
CREATE INDEX IF NOT EXISTS idx_museum_letters_status ON museum_letters(status);
CREATE INDEX IF NOT EXISTS idx_museum_letters_contributor ON museum_letters(contributor_id);
CREATE INDEX IF NOT EXISTS idx_museum_letters_historical_date ON museum_letters(historical_date);
CREATE INDEX IF NOT EXISTS idx_museum_letters_created_at ON museum_letters(created_at DESC);

-- 复合索引：状态+时期查询优化
CREATE INDEX IF NOT EXISTS idx_museum_letters_status_era ON museum_letters(status, era);

-- 复合索引：精选+时期查询优化
CREATE INDEX IF NOT EXISTS idx_museum_letters_featured_era ON museum_letters(is_featured, era) WHERE is_featured = true;

-- 热度排序索引
CREATE INDEX IF NOT EXISTS idx_museum_letters_popularity ON museum_letters(
    (view_count + favorite_count * 2 + rating_avg * 10) DESC,
    created_at DESC
) WHERE status IN ('approved', 'featured');

-- 文本搜索索引
CREATE INDEX IF NOT EXISTS idx_museum_letters_title_search ON museum_letters USING gin(to_tsvector('simple', title));
CREATE INDEX IF NOT EXISTS idx_museum_letters_content_search ON museum_letters USING gin(to_tsvector('simple', content));
CREATE INDEX IF NOT EXISTS idx_museum_letters_author_search ON museum_letters USING gin(to_tsvector('simple', original_author));

-- 创建博物馆收藏表
CREATE TABLE IF NOT EXISTS museum_favorites (
    museum_letter_id VARCHAR(20) NOT NULL REFERENCES museum_letters(id) ON DELETE CASCADE,
    user_id VARCHAR(50) NOT NULL,
    note VARCHAR(200),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (museum_letter_id, user_id)
);

-- 创建博物馆收藏索引
CREATE INDEX IF NOT EXISTS idx_museum_favorites_museum_letter ON museum_favorites(museum_letter_id);
CREATE INDEX IF NOT EXISTS idx_museum_favorites_user ON museum_favorites(user_id);

-- 创建博物馆评分表
CREATE TABLE IF NOT EXISTS museum_ratings (
    museum_letter_id VARCHAR(20) NOT NULL REFERENCES museum_letters(id) ON DELETE CASCADE,
    user_id VARCHAR(50) NOT NULL,
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    comment VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (museum_letter_id, user_id)
);

-- 创建博物馆评分索引
CREATE INDEX IF NOT EXISTS idx_museum_ratings_museum_letter ON museum_ratings(museum_letter_id);
CREATE INDEX IF NOT EXISTS idx_museum_ratings_user ON museum_ratings(user_id);
CREATE INDEX IF NOT EXISTS idx_museum_ratings_rating ON museum_ratings(rating);

-- 创建时间线事件表
CREATE TABLE IF NOT EXISTS timeline_events (
    id VARCHAR(20) PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    event_date TIMESTAMP WITH TIME ZONE NOT NULL,
    era VARCHAR(20) NOT NULL,
    location VARCHAR(200),
    event_type VARCHAR(30) NOT NULL,
    category VARCHAR(50),
    importance INTEGER NOT NULL DEFAULT 1 CHECK (importance >= 1 AND importance <= 5),
    museum_letter_id VARCHAR(20) REFERENCES museum_letters(id) ON DELETE SET NULL,
    is_featured BOOLEAN NOT NULL DEFAULT FALSE,
    display_order INTEGER NOT NULL DEFAULT 0,
    image_url VARCHAR(500),
    audio_url VARCHAR(500),
    video_url VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建时间线事件索引
CREATE INDEX IF NOT EXISTS idx_timeline_events_event_date ON timeline_events(event_date DESC);
CREATE INDEX IF NOT EXISTS idx_timeline_events_era ON timeline_events(era);
CREATE INDEX IF NOT EXISTS idx_timeline_events_event_type ON timeline_events(event_type);
CREATE INDEX IF NOT EXISTS idx_timeline_events_importance ON timeline_events(importance DESC);
CREATE INDEX IF NOT EXISTS idx_timeline_events_museum_letter ON timeline_events(museum_letter_id);

-- 复合索引：时期+日期查询优化
CREATE INDEX IF NOT EXISTS idx_timeline_events_era_date ON timeline_events(era, event_date DESC);

-- 复合索引：重要性+精选查询优化
CREATE INDEX IF NOT EXISTS idx_timeline_events_featured ON timeline_events(is_featured, importance DESC, event_date DESC);

-- 创建博物馆收藏集表
CREATE TABLE IF NOT EXISTS museum_collections (
    id VARCHAR(20) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    theme VARCHAR(50),
    creator_id VARCHAR(50) NOT NULL,
    creator_name VARCHAR(100),
    is_public BOOLEAN NOT NULL DEFAULT TRUE,
    is_featured BOOLEAN NOT NULL DEFAULT FALSE,
    letter_count INTEGER NOT NULL DEFAULT 0,
    view_count INTEGER NOT NULL DEFAULT 0,
    follow_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建博物馆收藏集索引
CREATE INDEX IF NOT EXISTS idx_museum_collections_creator ON museum_collections(creator_id);
CREATE INDEX IF NOT EXISTS idx_museum_collections_theme ON museum_collections(theme);
CREATE INDEX IF NOT EXISTS idx_museum_collections_is_public ON museum_collections(is_public);
CREATE INDEX IF NOT EXISTS idx_museum_collections_is_featured ON museum_collections(is_featured);

-- 创建收藏集信件关联表
CREATE TABLE IF NOT EXISTS collection_letters (
    collection_id VARCHAR(20) NOT NULL REFERENCES museum_collections(id) ON DELETE CASCADE,
    museum_letter_id VARCHAR(20) NOT NULL REFERENCES museum_letters(id) ON DELETE CASCADE,
    added_by VARCHAR(50),
    note VARCHAR(200),
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (collection_id, museum_letter_id)
);

-- 创建收藏集信件关联索引
CREATE INDEX IF NOT EXISTS idx_collection_letters_collection ON collection_letters(collection_id);
CREATE INDEX IF NOT EXISTS idx_collection_letters_museum_letter ON collection_letters(museum_letter_id);
CREATE INDEX IF NOT EXISTS idx_collection_letters_sort_order ON collection_letters(collection_id, sort_order);

-- 创建更新时间触发器（博物馆相关表）
CREATE TRIGGER update_museum_letters_updated_at 
    BEFORE UPDATE ON museum_letters 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_museum_ratings_updated_at 
    BEFORE UPDATE ON museum_ratings 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_timeline_events_updated_at 
    BEFORE UPDATE ON timeline_events 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_museum_collections_updated_at 
    BEFORE UPDATE ON museum_collections 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- 插入默认博物馆信件数据
INSERT INTO museum_letters (
    id, title, content, summary, original_author, era, location, category, 
    source_type, contributor_id, contributor_name, status, historical_date
) VALUES 
(
    'MS1234567890',
    '家书一封',
    '亲爱的家人：见字如面，甚念。近日在外一切安好，勿念。工作繁忙，但身体健康。望家中老小皆安，盼早日团聚。',
    '一封温馨的家书，表达对家人的思念之情',
    '张三',
    'contemporary',
    '上海',
    '家书',
    'contributed',
    'system',
    '系统管理员',
    'approved',
    '1945-08-15 10:00:00+08'
),
(
    'MS0987654321',
    '战地书信',
    '同志们：革命尚未成功，同志仍需努力。我们要坚定信念，为了人民的解放事业，不怕牺牲，勇往直前。',
    '激励人心的革命书信，体现了革命先烈的坚定信念',
    '李四',
    'contemporary', 
    '延安',
    '革命书信',
    'digitized',
    'system',
    '系统管理员',
    'featured',
    '1940-12-25 15:30:00+08'
) ON CONFLICT (id) DO NOTHING;

-- 插入默认时间线事件
INSERT INTO timeline_events (
    id, title, description, event_date, era, location, event_type, 
    category, importance, is_featured
) VALUES 
(
    'TL1234567890',
    '中华人民共和国成立',
    '1949年10月1日，中华人民共和国在北京宣告成立，中国人民从此站起来了。',
    '1949-10-01 10:00:00+08',
    'contemporary',
    '北京',
    'historical',
    '政治',
    5,
    true
),
(
    'TL0987654321', 
    '第一封电子邮件发送',
    '1971年，雷·汤姆林森发送了第一封电子邮件，开启了数字通讯的新时代。',
    '1971-10-01 10:00:00+08',
    'present',
    '美国',
    'cultural',
    '科技',
    4,
    true
) ON CONFLICT (id) DO NOTHING;

-- 插入默认收藏集
INSERT INTO museum_collections (
    id, name, description, theme, creator_id, creator_name, is_featured
) VALUES 
(
    'CL1234567890',
    '革命年代的书信',
    '收录了革命战争年代的珍贵书信，展现了革命先烈的崇高理想和坚定信念。',
    '革命历史',
    'system',
    '系统管理员',
    true
),
(
    'CL0987654321',
    '家书里的温情',
    '精选温馨感人的家书，展现中华民族深厚的家庭情感和传统文化。',
    '家庭情感',
    'system', 
    '系统管理员',
    true
) ON CONFLICT (id) DO NOTHING;

-- 创建商品表
CREATE TABLE IF NOT EXISTS shop_products (
    id VARCHAR(20) PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    short_description VARCHAR(500),
    category VARCHAR(50) NOT NULL,
    product_type VARCHAR(20) NOT NULL,
    tags VARCHAR(300),
    brand VARCHAR(100),
    price FLOAT NOT NULL,
    original_price FLOAT,
    cost_price FLOAT,
    currency VARCHAR(3) DEFAULT 'CNY',
    stock_quantity INTEGER DEFAULT 0,
    min_stock INTEGER DEFAULT 0,
    max_quantity_per_order INTEGER DEFAULT 999,
    status VARCHAR(20) DEFAULT 'draft' NOT NULL,
    is_featured BOOLEAN DEFAULT FALSE,
    is_digital BOOLEAN DEFAULT FALSE,
    weight FLOAT,
    dimensions VARCHAR(100),
    color VARCHAR(50),
    material VARCHAR(100),
    main_image VARCHAR(500),
    gallery_images TEXT,
    video_url VARCHAR(500),
    seo_title VARCHAR(200),
    seo_description VARCHAR(500),
    seo_keywords VARCHAR(300),
    view_count INTEGER DEFAULT 0,
    sales_count INTEGER DEFAULT 0,
    rating_avg FLOAT DEFAULT 0.0,
    rating_count INTEGER DEFAULT 0,
    favorite_count INTEGER DEFAULT 0,
    creator_id VARCHAR(50),
    creator_name VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    published_at TIMESTAMP WITH TIME ZONE
);

-- 创建商品表索引
CREATE INDEX IF NOT EXISTS idx_shop_products_category ON shop_products(category);
CREATE INDEX IF NOT EXISTS idx_shop_products_product_type ON shop_products(product_type);
CREATE INDEX IF NOT EXISTS idx_shop_products_status ON shop_products(status);
CREATE INDEX IF NOT EXISTS idx_shop_products_brand ON shop_products(brand);
CREATE INDEX IF NOT EXISTS idx_shop_products_price ON shop_products(price);
CREATE INDEX IF NOT EXISTS idx_shop_products_stock ON shop_products(stock_quantity);
CREATE INDEX IF NOT EXISTS idx_shop_products_created_at ON shop_products(created_at DESC);

-- 复合索引：状态+分类查询优化
CREATE INDEX IF NOT EXISTS idx_shop_products_status_category ON shop_products(status, category);

-- 复合索引：精选商品查询优化
CREATE INDEX IF NOT EXISTS idx_shop_products_featured ON shop_products(is_featured, status) WHERE is_featured = true;

-- 热度排序索引
CREATE INDEX IF NOT EXISTS idx_shop_products_popularity ON shop_products(
    (sales_count * 0.5 + view_count * 0.3 + rating_avg * 0.2) DESC,
    created_at DESC
) WHERE status = 'active';

-- 文本搜索索引
CREATE INDEX IF NOT EXISTS idx_shop_products_name_search ON shop_products USING gin(to_tsvector('simple', name));
CREATE INDEX IF NOT EXISTS idx_shop_products_description_search ON shop_products USING gin(to_tsvector('simple', description));
CREATE INDEX IF NOT EXISTS idx_shop_products_tags_search ON shop_products USING gin(to_tsvector('simple', tags));

-- 创建商品分类表
CREATE TABLE IF NOT EXISTS shop_categories (
    id VARCHAR(20) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description VARCHAR(500),
    parent_id VARCHAR(20) REFERENCES shop_categories(id) ON DELETE SET NULL,
    icon VARCHAR(100),
    banner_image VARCHAR(500),
    is_active BOOLEAN DEFAULT TRUE,
    sort_order INTEGER DEFAULT 0,
    product_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建商品分类索引
CREATE INDEX IF NOT EXISTS idx_shop_categories_parent_id ON shop_categories(parent_id);
CREATE INDEX IF NOT EXISTS idx_shop_categories_is_active ON shop_categories(is_active);
CREATE INDEX IF NOT EXISTS idx_shop_categories_sort_order ON shop_categories(sort_order);

-- 创建订单表
CREATE TABLE IF NOT EXISTS shop_orders (
    id VARCHAR(20) PRIMARY KEY,
    user_id VARCHAR(50) NOT NULL,
    user_name VARCHAR(100),
    user_email VARCHAR(200),
    user_phone VARCHAR(20),
    status VARCHAR(20) DEFAULT 'pending' NOT NULL,
    payment_status VARCHAR(20) DEFAULT 'pending' NOT NULL,
    subtotal FLOAT NOT NULL,
    shipping_fee FLOAT DEFAULT 0.0,
    tax_fee FLOAT DEFAULT 0.0,
    discount_amount FLOAT DEFAULT 0.0,
    total_amount FLOAT NOT NULL,
    currency VARCHAR(3) DEFAULT 'CNY',
    shipping_name VARCHAR(100),
    shipping_phone VARCHAR(20),
    shipping_address TEXT,
    shipping_city VARCHAR(100),
    shipping_province VARCHAR(100),
    shipping_postal_code VARCHAR(20),
    shipping_method VARCHAR(50),
    user_note VARCHAR(500),
    admin_note VARCHAR(500),
    coupon_code VARCHAR(50),
    coupon_discount FLOAT DEFAULT 0.0,
    payment_method VARCHAR(50),
    payment_transaction_id VARCHAR(100),
    paid_at TIMESTAMP WITH TIME ZONE,
    tracking_number VARCHAR(100),
    shipping_company VARCHAR(100),
    shipped_at TIMESTAMP WITH TIME ZONE,
    delivered_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建订单表索引
CREATE INDEX IF NOT EXISTS idx_shop_orders_user_id ON shop_orders(user_id);
CREATE INDEX IF NOT EXISTS idx_shop_orders_status ON shop_orders(status);
CREATE INDEX IF NOT EXISTS idx_shop_orders_payment_status ON shop_orders(payment_status);
CREATE INDEX IF NOT EXISTS idx_shop_orders_created_at ON shop_orders(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_shop_orders_tracking_number ON shop_orders(tracking_number);

-- 复合索引：用户+状态查询优化
CREATE INDEX IF NOT EXISTS idx_shop_orders_user_status ON shop_orders(user_id, status);

-- 复合索引：状态+创建时间查询优化
CREATE INDEX IF NOT EXISTS idx_shop_orders_status_created ON shop_orders(status, created_at DESC);

-- 创建订单商品项表
CREATE TABLE IF NOT EXISTS shop_order_items (
    id VARCHAR(20) PRIMARY KEY,
    order_id VARCHAR(20) NOT NULL REFERENCES shop_orders(id) ON DELETE CASCADE,
    product_id VARCHAR(20) NOT NULL REFERENCES shop_products(id),
    product_name VARCHAR(200) NOT NULL,
    product_image VARCHAR(500),
    product_sku VARCHAR(100),
    unit_price FLOAT NOT NULL,
    quantity INTEGER NOT NULL,
    total_price FLOAT NOT NULL,
    product_attributes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建订单商品项索引
CREATE INDEX IF NOT EXISTS idx_shop_order_items_order_id ON shop_order_items(order_id);
CREATE INDEX IF NOT EXISTS idx_shop_order_items_product_id ON shop_order_items(product_id);

-- 创建购物车表
CREATE TABLE IF NOT EXISTS shop_carts (
    id VARCHAR(20) PRIMARY KEY,
    user_id VARCHAR(50) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建购物车索引
CREATE INDEX IF NOT EXISTS idx_shop_carts_user_id ON shop_carts(user_id);

-- 创建购物车商品项表
CREATE TABLE IF NOT EXISTS shop_cart_items (
    id VARCHAR(20) PRIMARY KEY,
    cart_id VARCHAR(20) NOT NULL REFERENCES shop_carts(id) ON DELETE CASCADE,
    product_id VARCHAR(20) NOT NULL REFERENCES shop_products(id),
    quantity INTEGER NOT NULL,
    product_attributes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建购物车商品项索引
CREATE INDEX IF NOT EXISTS idx_shop_cart_items_cart_id ON shop_cart_items(cart_id);
CREATE INDEX IF NOT EXISTS idx_shop_cart_items_product_id ON shop_cart_items(product_id);

-- 购物车项唯一约束（一个购物车中同一商品只能有一条记录）
CREATE UNIQUE INDEX IF NOT EXISTS idx_shop_cart_items_unique ON shop_cart_items(cart_id, product_id);

-- 创建商品评价表
CREATE TABLE IF NOT EXISTS shop_product_reviews (
    id VARCHAR(20) PRIMARY KEY,
    product_id VARCHAR(20) NOT NULL REFERENCES shop_products(id) ON DELETE CASCADE,
    user_id VARCHAR(50) NOT NULL,
    user_name VARCHAR(100),
    order_id VARCHAR(20) REFERENCES shop_orders(id) ON DELETE SET NULL,
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    title VARCHAR(200),
    content TEXT,
    images TEXT,
    reply_content TEXT,
    reply_at TIMESTAMP WITH TIME ZONE,
    is_anonymous BOOLEAN DEFAULT FALSE,
    is_verified BOOLEAN DEFAULT FALSE,
    is_hidden BOOLEAN DEFAULT FALSE,
    helpful_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建商品评价索引
CREATE INDEX IF NOT EXISTS idx_shop_product_reviews_product_id ON shop_product_reviews(product_id);
CREATE INDEX IF NOT EXISTS idx_shop_product_reviews_user_id ON shop_product_reviews(user_id);
CREATE INDEX IF NOT EXISTS idx_shop_product_reviews_order_id ON shop_product_reviews(order_id);
CREATE INDEX IF NOT EXISTS idx_shop_product_reviews_rating ON shop_product_reviews(rating);
CREATE INDEX IF NOT EXISTS idx_shop_product_reviews_created_at ON shop_product_reviews(created_at DESC);

-- 评价唯一约束（同一用户对同一商品只能评价一次）
CREATE UNIQUE INDEX IF NOT EXISTS idx_shop_product_reviews_unique ON shop_product_reviews(product_id, user_id);

-- 创建商品收藏表
CREATE TABLE IF NOT EXISTS shop_product_favorites (
    product_id VARCHAR(20) NOT NULL REFERENCES shop_products(id) ON DELETE CASCADE,
    user_id VARCHAR(50) NOT NULL,
    note VARCHAR(200),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (product_id, user_id)
);

-- 创建商品收藏索引
CREATE INDEX IF NOT EXISTS idx_shop_product_favorites_product_id ON shop_product_favorites(product_id);
CREATE INDEX IF NOT EXISTS idx_shop_product_favorites_user_id ON shop_product_favorites(user_id);

-- 创建更新时间触发器（商店相关表）
CREATE TRIGGER update_shop_products_updated_at 
    BEFORE UPDATE ON shop_products 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_shop_categories_updated_at 
    BEFORE UPDATE ON shop_categories 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_shop_orders_updated_at 
    BEFORE UPDATE ON shop_orders 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_shop_carts_updated_at 
    BEFORE UPDATE ON shop_carts 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_shop_cart_items_updated_at 
    BEFORE UPDATE ON shop_cart_items 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_shop_product_reviews_updated_at 
    BEFORE UPDATE ON shop_product_reviews 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- 插入默认商品分类数据
INSERT INTO shop_categories (id, name, description, icon, sort_order) VALUES 
('envelopes', '信封', '各种样式和材质的信封', '✉️', 1),
('stationery', '文具', '钢笔、墨水、纸张等写作文具', '🖊️', 2),
('stamps', '邮票', '纪念邮票和通用邮票', '📮', 3),
('postcards', '明信片', '风景明信片和艺术明信片', '🏞️', 4),
('gifts', '礼品', '书信相关的精美礼品', '🎁', 5),
('digital', '数字商品', '电子模板和数字内容', '💻', 6)
ON CONFLICT (id) DO NOTHING;

-- 插入示例商品数据
INSERT INTO shop_products (
    id, name, description, short_description, category, product_type, 
    price, stock_quantity, status, is_featured, main_image
) VALUES 
(
    'PD1234567890',
    '复古牛皮纸信封',
    '采用优质牛皮纸制作，给您的信件增添复古韵味。每包10枚，规格：220mm x 110mm。',
    '优质牛皮纸信封，复古质感，10枚装',
    'envelopes',
    'envelope',
    15.8,
    100,
    'active',
    true,
    '/images/products/envelope_vintage.jpg'
),
(
    'PD0987654321',
    '经典钢笔墨水套装',
    '包含三色墨水：蓝黑、纯蓝、黑色，每瓶30ml，适用于各种钢笔。',
    '三色钢笔墨水套装，每瓶30ml',
    'stationery', 
    'stationery',
    45.0,
    50,
    'active',
    true,
    '/images/products/ink_set.jpg'
) ON CONFLICT (id) DO NOTHING;

-- ================================================================
-- 草稿管理表
-- ================================================================

-- 信件草稿表
CREATE TABLE IF NOT EXISTS letter_drafts (
    id VARCHAR(20) PRIMARY KEY,
    user_id VARCHAR(20) NOT NULL,
    
    -- 草稿内容
    title VARCHAR(200),
    content TEXT,
    
    -- 收件人信息
    recipient_id VARCHAR(20),
    recipient_type VARCHAR(20),  -- friend/stranger/group
    
    -- 样式配置
    paper_style VARCHAR(50) DEFAULT 'classic',
    envelope_style VARCHAR(50) DEFAULT 'simple',
    
    -- 草稿元数据
    draft_type VARCHAR(20) DEFAULT 'letter',  -- letter/reply
    parent_letter_id VARCHAR(20),
    
    -- 版本控制
    version INTEGER DEFAULT 1,
    word_count INTEGER DEFAULT 0,
    character_count INTEGER DEFAULT 0,
    
    -- 自动保存配置
    auto_save_enabled BOOLEAN DEFAULT TRUE,
    last_edit_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- 状态
    is_active BOOLEAN DEFAULT TRUE,
    is_discarded BOOLEAN DEFAULT FALSE,
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 草稿历史记录表
CREATE TABLE IF NOT EXISTS draft_history (
    id VARCHAR(20) PRIMARY KEY,
    draft_id VARCHAR(20) NOT NULL,
    user_id VARCHAR(20) NOT NULL,
    
    -- 历史版本内容
    title VARCHAR(200),
    content TEXT,
    version INTEGER NOT NULL,
    
    -- 变更信息
    change_summary VARCHAR(500),
    change_type VARCHAR(20) DEFAULT 'auto_save',  -- auto_save/manual_save/version_backup
    
    -- 统计信息
    word_count INTEGER DEFAULT 0,
    character_count INTEGER DEFAULT 0,
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- 外键约束
    FOREIGN KEY (draft_id) REFERENCES letter_drafts(id) ON DELETE CASCADE
);

-- ================================================================
-- 草稿相关索引
-- ================================================================

-- 草稿表索引
CREATE INDEX IF NOT EXISTS idx_letter_drafts_user_id ON letter_drafts(user_id);
CREATE INDEX IF NOT EXISTS idx_letter_drafts_user_active ON letter_drafts(user_id, is_active);
CREATE INDEX IF NOT EXISTS idx_letter_drafts_user_edit_time ON letter_drafts(user_id, last_edit_time DESC);
CREATE INDEX IF NOT EXISTS idx_letter_drafts_type ON letter_drafts(draft_type);
CREATE INDEX IF NOT EXISTS idx_letter_drafts_recipient ON letter_drafts(recipient_id, recipient_type);
CREATE INDEX IF NOT EXISTS idx_letter_drafts_parent ON letter_drafts(parent_letter_id);

-- 草稿历史索引
CREATE INDEX IF NOT EXISTS idx_draft_history_draft_id ON draft_history(draft_id);
CREATE INDEX IF NOT EXISTS idx_draft_history_user_id ON draft_history(user_id);
CREATE INDEX IF NOT EXISTS idx_draft_history_version ON draft_history(draft_id, version DESC);
CREATE INDEX IF NOT EXISTS idx_draft_history_type ON draft_history(change_type);
CREATE INDEX IF NOT EXISTS idx_draft_history_created ON draft_history(created_at DESC);

-- 创建更新时间触发器（草稿相关表）
CREATE TRIGGER update_letter_drafts_updated_at 
    BEFORE UPDATE ON letter_drafts 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- 授予权限（如果需要）
-- GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO write_service_user;