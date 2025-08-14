#!/bin/bash

# Check PostgreSQL database for shop tables
echo "=== Checking Shop Tables in Database ==="

# Database connection details
DB_NAME="openpenpal"
DB_USER=$(whoami)

# Check if tables exist
echo -e "\n1. Checking if shop tables exist:"
psql -U $DB_USER -d $DB_NAME -c "\dt *shop*" 2>/dev/null || echo "Error connecting to database"

echo -e "\n2. Checking products table structure:"
psql -U $DB_USER -d $DB_NAME -c "\d products" 2>/dev/null || echo "Products table not found"

echo -e "\n3. Checking carts table structure:"
psql -U $DB_USER -d $DB_NAME -c "\d carts" 2>/dev/null || echo "Carts table not found"

echo -e "\n4. Checking orders table structure:"
psql -U $DB_USER -d $DB_NAME -c "\d orders" 2>/dev/null || echo "Orders table not found"

echo -e "\n5. Counting records in shop tables:"
psql -U $DB_USER -d $DB_NAME -c "
SELECT 'products' as table_name, COUNT(*) as count FROM products
UNION ALL
SELECT 'carts', COUNT(*) FROM carts
UNION ALL
SELECT 'orders', COUNT(*) FROM orders
UNION ALL
SELECT 'product_reviews', COUNT(*) FROM product_reviews
UNION ALL
SELECT 'product_favorites', COUNT(*) FROM product_favorites;
" 2>/dev/null || echo "Error counting records"