-- Fix Foreign Key Data Type Mismatch Issues
-- Following CLAUDE.md principles: Think before action, cautious approach

-- 1. First, analyze the current state
\echo '=== Analyzing current table structures ==='

-- Check products table ID type
SELECT column_name, data_type, udt_name, character_maximum_length
FROM information_schema.columns
WHERE table_name = 'products' AND column_name = 'id';

-- Check cart_items product_id type
SELECT column_name, data_type, udt_name, character_maximum_length
FROM information_schema.columns
WHERE table_name = 'cart_items' AND column_name = 'product_id';

-- Check orders table ID type
SELECT column_name, data_type, udt_name, character_maximum_length
FROM information_schema.columns
WHERE table_name = 'orders' AND column_name = 'id';

-- Check order_items order_id type
SELECT column_name, data_type, udt_name, character_maximum_length
FROM information_schema.columns
WHERE table_name = 'order_items' AND column_name = 'order_id';

\echo '=== Starting foreign key fixes ==='

-- 2. Drop existing foreign key constraints that are causing issues
\echo 'Dropping problematic foreign key constraints...'

ALTER TABLE cart_items DROP CONSTRAINT IF EXISTS cart_items_product_id_fkey;
ALTER TABLE order_items DROP CONSTRAINT IF EXISTS order_items_order_id_fkey;
ALTER TABLE order_items DROP CONSTRAINT IF EXISTS order_items_product_id_fkey;
ALTER TABLE product_reviews DROP CONSTRAINT IF EXISTS fk_product_reviews_product;
ALTER TABLE product_reviews DROP CONSTRAINT IF EXISTS fk_product_reviews_order;
ALTER TABLE product_favorites DROP CONSTRAINT IF EXISTS fk_product_favorites_product;

-- 3. Fix cart_items table if needed
-- The issue is that GORM is trying to alter product_id to varchar(36) but it's already uuid
-- This happens because the Go model might define it as string instead of uuid.UUID
\echo 'Checking if cart_items.product_id needs type adjustment...'

DO $$
BEGIN
    -- Only alter if the column is uuid and needs to be varchar(36) for consistency
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'cart_items' 
        AND column_name = 'product_id' 
        AND udt_name = 'uuid'
    ) AND EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'products' 
        AND column_name = 'id' 
        AND udt_name = 'uuid'
    ) THEN
        -- Both are UUID, so we can recreate the constraint
        RAISE NOTICE 'Both cart_items.product_id and products.id are UUID type - compatible';
    END IF;
END $$;

-- 4. Fix order_items table
DO $$
BEGIN
    -- Check if order_items exists
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'order_items') THEN
        IF EXISTS (
            SELECT 1 FROM information_schema.columns 
            WHERE table_name = 'order_items' 
            AND column_name = 'order_id' 
            AND udt_name = 'uuid'
        ) AND EXISTS (
            SELECT 1 FROM information_schema.columns 
            WHERE table_name = 'orders' 
            AND column_name = 'id' 
            AND udt_name = 'uuid'
        ) THEN
            RAISE NOTICE 'Both order_items.order_id and orders.id are UUID type - compatible';
        END IF;
    END IF;
END $$;

-- 5. Recreate foreign key constraints with proper types
\echo 'Recreating foreign key constraints...'

-- Cart items to products
ALTER TABLE cart_items 
ADD CONSTRAINT cart_items_product_id_fkey 
FOREIGN KEY (product_id) REFERENCES products(id);

-- Order items to orders (if table exists)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'order_items') THEN
        EXECUTE 'ALTER TABLE order_items ADD CONSTRAINT order_items_order_id_fkey FOREIGN KEY (order_id) REFERENCES orders(id)';
        EXECUTE 'ALTER TABLE order_items ADD CONSTRAINT order_items_product_id_fkey FOREIGN KEY (product_id) REFERENCES products(id)';
    END IF;
END $$;

-- Product reviews (if table exists)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'product_reviews') THEN
        -- First ensure the table has the right structure
        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'product_reviews' AND column_name = 'id') THEN
            EXECUTE 'ALTER TABLE product_reviews ADD COLUMN id UUID DEFAULT gen_random_uuid() PRIMARY KEY';
        END IF;
        
        EXECUTE 'ALTER TABLE product_reviews ADD CONSTRAINT fk_product_reviews_product FOREIGN KEY (product_id) REFERENCES products(id)';
        EXECUTE 'ALTER TABLE product_reviews ADD CONSTRAINT fk_product_reviews_order FOREIGN KEY (order_id) REFERENCES orders(id)';
        EXECUTE 'ALTER TABLE product_reviews ADD CONSTRAINT fk_product_reviews_user FOREIGN KEY (user_id) REFERENCES users(id)';
    END IF;
END $$;

-- Product favorites (if table exists)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'product_favorites') THEN
        EXECUTE 'ALTER TABLE product_favorites ADD CONSTRAINT fk_product_favorites_product FOREIGN KEY (product_id) REFERENCES products(id)';
        EXECUTE 'ALTER TABLE product_favorites ADD CONSTRAINT fk_product_favorites_user FOREIGN KEY (user_id) REFERENCES users(id)';
    END IF;
END $$;

-- 6. Verify the fixes
\echo '=== Verifying foreign key constraints ==='

SELECT
    tc.table_name, 
    kcu.column_name, 
    ccu.table_name AS foreign_table_name,
    ccu.column_name AS foreign_column_name 
FROM 
    information_schema.table_constraints AS tc 
    JOIN information_schema.key_column_usage AS kcu
      ON tc.constraint_name = kcu.constraint_name
    JOIN information_schema.constraint_column_usage AS ccu
      ON ccu.constraint_name = tc.constraint_name
WHERE constraint_type = 'FOREIGN KEY' 
    AND tc.table_name IN ('cart_items', 'order_items', 'product_reviews', 'product_favorites')
ORDER BY tc.table_name;

\echo '=== Foreign key fix script completed ===';