#!/bin/bash

# Fix Chunk Loading and Hydration Issues Script
# This script cleans the build cache and restarts the development server

echo "🔧 Fixing chunk loading and hydration issues..."

# Step 1: Kill any existing Next.js processes
echo "1️⃣ Stopping existing Next.js processes..."
pkill -f "next dev" || true
lsof -ti:3000 | xargs kill -9 2>/dev/null || true

# Step 2: Clear Next.js cache and build files
echo "2️⃣ Clearing build cache..."
rm -rf .next
rm -rf node_modules/.cache
rm -rf .swc

# Step 3: Clear browser service worker (instructions)
echo "3️⃣ Browser cleanup required:"
echo "   Please open your browser DevTools and:"
echo "   - Go to Application → Storage"
echo "   - Click 'Clear site data'"
echo "   - OR in Application → Service Workers → Unregister all"
echo ""
echo "   Press Enter when done..."
read -r

# Step 4: Rebuild the application
echo "4️⃣ Rebuilding the application..."
npm run build

# Step 5: Start the development server
echo "5️⃣ Starting development server..."
npm run dev

echo "✅ Fix applied! The application should now load without chunk errors."