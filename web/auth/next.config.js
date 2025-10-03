/** @type {import('next').NextConfig} */
const nextConfig = {
  // Environment variables
  env: {
    NEXT_PUBLIC_API_URL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080',
    NEXT_PUBLIC_WEBAUTHN_RP_ID: process.env.NEXT_PUBLIC_WEBAUTHN_RP_ID || 'localhost',
    NEXT_PUBLIC_WEBAUTHN_RP_NAME: process.env.NEXT_PUBLIC_WEBAUTHN_RP_NAME || 'Sonr Auth',
    NEXT_PUBLIC_CHAIN_ID: process.env.NEXT_PUBLIC_CHAIN_ID || 'sonrtest_1-1',
  },

  // OpenNext handles the output configuration
  // output is managed by @opennextjs/cloudflare

  // Enable trailing slashes for better static hosting
  trailingSlash: true,

  // Image optimization settings for static export
  images: {
    unoptimized: true,
  },

  // Webpack configuration for edge runtime compatibility
  webpack: (config, { isServer, webpack }) => {
    // Handle node: protocol imports
    config.plugins.push(
      new webpack.NormalModuleReplacementPlugin(/^node:/, (resource) => {
        resource.request = resource.request.replace(/^node:/, '');
      })
    );

    // Add fallbacks for node modules in browser/edge
    if (!isServer) {
      config.resolve.fallback = {
        ...config.resolve.fallback,
        fs: false,
        path: false,
        stream: false,
        crypto: false,
        buffer: false,
        util: false,
        process: false,
      };
    }

    // Ignore node-specific modules
    config.resolve.alias = {
      ...config.resolve.alias,
      'node:fs': false,
      'node:path': false,
      'node:stream': false,
      'node:buffer': false,
      'node:util': false,
      'node:process': false,
    };

    return config;
  },

  // TypeScript and ESLint configuration
  typescript: {
    ignoreBuildErrors: process.env.NODE_ENV === 'production',
  },

  eslint: {
    ignoreDuringBuilds: process.env.NODE_ENV === 'production',
  },
};

module.exports = nextConfig;
