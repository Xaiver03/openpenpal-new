#!/bin/bash

echo "=== Testing Complete Order Flow ==="

# 1. Login
echo -e "\n1. Getting auth token..."
TOKEN=$(curl -s -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username": "alice", "password": "secret"}' | jq -r '.data.token')

echo "Token obtained"

# 2. Check cart
echo -e "\n2. Checking current cart..."
CART=$(curl -s -X GET "http://localhost:8080/api/v1/shop/cart" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Accept: application/json")
echo "$CART" | jq .

# 3. Create order from cart
echo -e "\n3. Creating order from cart..."
ORDER=$(curl -s -X POST "http://localhost:8080/api/v1/shop/orders" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "payment_method": "alipay",
    "shipping_address": {
      "name": "Alice Smith",
      "phone": "13800138000",
      "address": "北京大学理科5号楼303室",
      "city": "北京市",
      "province": "北京市",
      "postal_code": "100871"
    },
    "notes": "请小心轻放",
    "coupon_code": ""
  }')
echo "$ORDER" | jq .

# Extract order ID if successful
ORDER_ID=$(echo "$ORDER" | jq -r '.data.id // empty')

if [ -n "$ORDER_ID" ]; then
    # 4. Pay for order
    echo -e "\n4. Paying for order $ORDER_ID..."
    PAYMENT=$(curl -s -X POST "http://localhost:8080/api/v1/shop/orders/$ORDER_ID/pay" \
      -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json" \
      -d "{\"payment_id\": \"PAY_$(date +%s)\"}")
    echo "$PAYMENT" | jq .
    
    # 5. Check order status
    echo -e "\n5. Checking order status..."
    curl -s -X GET "http://localhost:8080/api/v1/shop/orders/$ORDER_ID" \
      -H "Authorization: Bearer $TOKEN" \
      -H "Accept: application/json" | jq .
fi

# 6. List all orders
echo -e "\n6. Listing all orders..."
curl -s -X GET "http://localhost:8080/api/v1/shop/orders" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Accept: application/json" | jq .