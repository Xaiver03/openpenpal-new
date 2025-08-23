#!/bin/bash

# 生成安全的加密密钥
# 使用: ./scripts/generate-encryption-key.sh

echo "🔐 Generating secure encryption key for OpenPenPal..."

# 生成256位（32字节）的随机密钥
if command -v openssl &> /dev/null; then
    # 使用OpenSSL生成
    ENCRYPTION_KEY=$(openssl rand -hex 32)
elif command -v python3 &> /dev/null; then
    # 使用Python生成
    ENCRYPTION_KEY=$(python3 -c "import secrets; print(secrets.token_hex(32))")
else
    echo "❌ Error: Neither openssl nor python3 found. Please install one of them."
    exit 1
fi

echo "✅ Generated encryption key:"
echo "ENCRYPTION_KEY=$ENCRYPTION_KEY"
echo ""
echo "📝 Please add this to your environment variables:"
echo "   1. Add to .env file: echo 'ENCRYPTION_KEY=$ENCRYPTION_KEY' >> .env"
echo "   2. Or export in shell: export ENCRYPTION_KEY=$ENCRYPTION_KEY"
echo ""
echo "⚠️  IMPORTANT SECURITY NOTES:"
echo "   - Keep this key secure and NEVER commit it to version control"
echo "   - Use different keys for development and production"
echo "   - Store production keys in secure key management systems"
echo "   - If this key is lost, encrypted data cannot be recovered"
echo ""
echo "🔄 To apply the new key:"
echo "   1. Set the environment variable"
echo "   2. Restart the application"
echo "   3. Run data migration if needed: go run cmd/migrate/encrypt_data.go"