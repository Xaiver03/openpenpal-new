-- Insert test products
INSERT INTO products (name, description, category, product_type, price, original_price, discount, stock, sold, image_url, tags, is_featured)
VALUES 
('复古牛皮纸信封套装', '优质牛皮纸制作，复古文艺风格，适合各种主题书信', '信封', 'envelope', 28.00, 35.00, 20, 100, 1456, '/api/placeholder/300/300', '["热销", "复古"]', true),
('樱花主题信纸礼盒', '精美樱花图案信纸，配套信封，春日限定款', '信纸', 'stationery', 68.00, 88.00, 23, 50, 892, '/api/placeholder/300/300', '["限定", "精美"]', true),
('手绘校园风景邮票', '手绘各大高校标志性建筑，收藏与使用兼备', '邮票', 'stamp', 45.00, 45.00, 0, 200, 567, '/api/placeholder/300/300', '["手绘", "校园"]', false),
('OpenPenPal定制礼品套装', '品牌定制信纸、信封、封蜡、钢笔，完整书信体验', '礼品套装', 'gift', 288.00, 328.00, 12, 30, 234, '/api/placeholder/300/300', '["定制", "豪华"]', true),
('北大纪念明信片套装', '北大标志性建筑明信片，12张精选', '明信片', 'postcard', 36.00, 36.00, 0, 150, 789, '/api/placeholder/300/300', '["纪念", "北大"]', false);