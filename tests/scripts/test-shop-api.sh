#!/bin/bash

echo "=== Testing Shop API ==="

# 1. Login and get token
echo -e "\n1. Getting auth token..."
TOKEN=$(curl -s -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username": "alice", "password": "secret"}' | jq -r '.data.token')

if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
    echo "Failed to get token"
    exit 1
fi

echo "Token obtained: ${TOKEN:0:50}..."

# 2. Test products API
echo -e "\n2. Testing GET /api/v1/shop/products..."
curl -X GET "http://localhost:8080/api/v1/shop/products" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Accept: application/json" | jq .

# 3. Test cart API
echo -e "\n3. Testing GET /api/v1/shop/cart..."
curl -X GET "http://localhost:8080/api/v1/shop/cart" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Accept: application/json" | jq .

# 4. Add item to cart
echo -e "\n4. Testing POST /api/v1/shop/cart/items..."
PRODUCT_ID=$(psql -U $(whoami) -d openpenpal -t -c "SELECT id FROM products LIMIT 1;" | xargs)
echo "Using product ID: $PRODUCT_ID"

curl -X POST "http://localhost:8080/api/v1/shop/cart/items" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"product_id\": \"$PRODUCT_ID\", \"quantity\": 2}" | jq .

# 5. Test orders
echo -e "\n5. Testing GET /api/v1/shop/orders..."
curl -X GET "http://localhost:8080/api/v1/shop/orders" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Accept: application/json" | jq .