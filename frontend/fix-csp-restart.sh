#!/bin/bash

echo "🔧 Fixing CSP issues and restarting frontend..."

# Kill any running Next.js processes
echo "Stopping frontend..."
pkill -f "next dev" || true

# Clear Next.js cache
echo "Clearing Next.js cache..."
rm -rf .next
rm -rf node_modules/.cache

# Clear browser service worker (instructions)
echo ""
echo "⚠️  IMPORTANT: Clear your browser cache for localhost:3000"
echo "   Chrome/Edge: Open DevTools → Application → Storage → Clear site data"
echo "   Firefox: Open DevTools → Storage → Clear All"
echo ""

# Restart frontend
echo "Starting frontend with fixed CSP configuration..."
cd frontend && npm run dev