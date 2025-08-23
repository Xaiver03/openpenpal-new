let BundleAnalyzerPlugin;
try {
  ({ BundleAnalyzerPlugin } = require('webpack-bundle-analyzer'));
} catch (e) {
  // webpack-bundle-analyzer is optional
}

/** @type {import('next').NextConfig} */
const nextConfig = {
  // 环境变量配置
  env: {
    NEXT_PUBLIC_API_URL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8000/api/v1',
    NEXT_PUBLIC_BACKEND_URL: process.env.NEXT_PUBLIC_BACKEND_URL || 'http://localhost:8080/api/v1',
    NEXT_PUBLIC_WS_URL: process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080/api/v1/ws/connect',
    NEXT_PUBLIC_APP_NAME: process.env.NEXT_PUBLIC_APP_NAME || 'OpenPenPal',
    NEXT_PUBLIC_ENVIRONMENT: process.env.NEXT_PUBLIC_ENVIRONMENT || 'development',
  },
  
  // 图片优化配置
  images: {
    domains: ['localhost', 'api.openpenpal.com', 'your-backend-api.com'],
    formats: ['image/webp', 'image/avif'],
    minimumCacheTTL: 60,
  },
  
  // 禁用Google字体优化以避免超时
  optimizeFonts: false,
  
  // 实验性功能和性能优化
  experimental: {
    optimizeCss: true,
    scrollRestoration: true,
    optimizePackageImports: ['lucide-react', '@radix-ui/react-icons'],
    turbo: {
      rules: {
        '*.svg': {
          loaders: ['@svgr/webpack'],
          as: '*.js',
        },
      },
    },
  },
  
  // 编译优化
  compiler: {
    removeConsole: process.env.NODE_ENV === 'production',
  },
  
  // 压缩配置
  compress: true,
  
  // PWA和Service Worker禁用
  generateEtags: false,
  
  // TypeScript严格检查已启用
  typescript: {
    ignoreBuildErrors: false,
  },
  
  // ESLint configuration
  eslint: {
    // Warning: This allows production builds to successfully complete even if
    // your project has ESLint errors.
    ignoreDuringBuilds: true,
  },
  
  // Webpack优化
  webpack: (config, { dev, isServer }) => {
    // 增加 chunk 加载超时时间
    if (!isServer) {
      config.output.chunkLoadTimeout = 120000; // 120 秒超时
    }
    // Bundle 分析器
    if (!isServer && !dev && process.env.ANALYZE === 'true' && BundleAnalyzerPlugin) {
      config.plugins.push(
        new BundleAnalyzerPlugin({
          analyzerMode: 'static',
          reportFilename: './analyze/client.html',
          openAnalyzer: false,
          generateStatsFile: true,
          statsFilename: './analyze/client.json',
        })
      )
    }

    // 解决canvas问题
    config.resolve.alias.canvas = false;
    
    // Fix webpack chunk loading issues
    if (!isServer) {
      config.output.publicPath = '/_next/';
      config.output.chunkLoadingGlobal = 'webpackChunkOpenPenPal';
      config.output.hotUpdateGlobal = 'webpackHotUpdateOpenPenPal';
      
      // 添加重试逻辑
      config.output.chunkLoadTimeout = 120000;
      config.output.webassemblyModuleFilename = 'static/wasm/[modulehash].wasm';
      config.output.enabledChunkLoadingTypes = ['jsonp', 'import-scripts'];
    }
    
    // 生产环境优化
    if (!dev && !isServer) {
      // 代码分割优化 - 减少碎片化
      config.optimization.splitChunks = {
        chunks: 'all',
        maxInitialRequests: 8, // 降低初始请求数
        maxAsyncRequests: 10,  // 降低异步请求数
        minSize: 20000,        // 最小块大小
        maxSize: 500000,       // 增大最大块大小
        cacheGroups: {
          // Framework chunk - 包含React和Next.js核心
          framework: {
            test: /[\\/]node_modules[\\/](react|react-dom|next)[\\/]/,
            name: 'framework',
            chunks: 'all',
            priority: 30,
            enforce: true,
          },
          // UI库和图标 - 合并为单个chunk
          ui: {
            test: /[\\/]node_modules[\\/](lucide-react|@radix-ui|@headlessui)[\\/]|[\\/]src[\\/]components[\\/]ui[\\/]/,
            name: 'ui-bundle',
            chunks: 'all',
            priority: 20,
            enforce: true,
          },
          // 所有node_modules - 合并为vendor chunk
          vendor: {
            test: /[\\/]node_modules[\\/]/,
            name: 'vendor',
            chunks: 'all',
            priority: 10,
            enforce: true,
          },
          // 应用代码 - 合并为app chunk
          app: {
            test: /[\\/]src[\\/]/,
            name: 'app',
            chunks: 'all',
            priority: 5,
            minChunks: 1,
            enforce: true,
          },
        },
      };

      // 资源优化
      config.optimization.usedExports = true;
      config.optimization.sideEffects = false;
    }
    
    return config;
  },
  
  // API代理配置 - 将API请求转发到后端
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: 'http://localhost:8080/api/:path*',
      },
    ];
  },
  
  // Headers优化 - Security headers are now handled by middleware
  async headers() {
    return [
      {
        source: '/static/(.*)',
        headers: [
          {
            key: 'Cache-Control',
            value: 'public, max-age=31536000, immutable',
          },
        ],
      },
    ];
  },
}

module.exports = nextConfig