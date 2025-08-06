const { BundleAnalyzerPlugin } = require('webpack-bundle-analyzer')

/** @type {import('next').NextConfig} */
const nextConfig = {
  webpack: (config, { isServer, dev }) => {
    // 只在客户端构建和非开发环境下启用分析器
    if (!isServer && !dev && process.env.ANALYZE === 'true') {
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

    // 服务端分析
    if (isServer && !dev && process.env.ANALYZE === 'true') {
      config.plugins.push(
        new BundleAnalyzerPlugin({
          analyzerMode: 'static',
          reportFilename: './analyze/server.html',
          openAnalyzer: false,
          generateStatsFile: true,
          statsFilename: './analyze/server.json',
        })
      )
    }

    return config
  },
}

module.exports = nextConfig