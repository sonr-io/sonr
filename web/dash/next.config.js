/** @type {import('next').NextConfig} */
const nextConfig = {
  // OpenNext handles the runtime configuration
  // Removed experimental.runtime and output configuration

  transpilePackages: ['@sonr.io/ui', '@sonr.io/shared', '@sonr.io/es', '@sonr.io/sdk'],

  // API proxy configuration for development
  async rewrites() {
    // Only apply in development
    if (process.env.NODE_ENV !== 'production') {
      return [
        {
          source: '/api/chain/:path*',
          destination: `${process.env.NEXT_PUBLIC_CHAIN_ENDPOINT || 'http://localhost:26657'}/:path*`,
        },
        {
          source: '/api/rpc/:path*',
          destination: `${process.env.NEXT_PUBLIC_RPC_ENDPOINT || 'http://localhost:1317'}/:path*`,
        },
        {
          source: '/api/grpc/:path*',
          destination: `${process.env.NEXT_PUBLIC_GRPC_ENDPOINT || 'http://localhost:9090'}/:path*`,
        },
      ];
    }
    return [];
  },

  // CORS headers for API routes
  async headers() {
    return [
      {
        source: '/api/:path*',
        headers: [
          { key: 'Access-Control-Allow-Credentials', value: 'true' },
          { key: 'Access-Control-Allow-Origin', value: '*' },
          { key: 'Access-Control-Allow-Methods', value: 'GET,DELETE,PATCH,POST,PUT' },
          {
            key: 'Access-Control-Allow-Headers',
            value:
              'X-CSRF-Token, X-Requested-With, Accept, Accept-Version, Content-Length, Content-MD5, Content-Type, Date, X-Api-Version',
          },
        ],
      },
    ];
  },

  // Environment variables passed to the client
  env: {
    NEXT_PUBLIC_CHAIN_ID: process.env.NEXT_PUBLIC_CHAIN_ID || 'sonrtest_1-1',
    NEXT_PUBLIC_AUTH_URL: process.env.NEXT_PUBLIC_AUTH_URL || 'http://localhost:3001',
    NEXT_PUBLIC_API_URL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:3000',
  },

  // Webpack configuration
  webpack: (config) => {
    // Handle WebAssembly modules
    config.experiments = {
      ...config.experiments,
      asyncWebAssembly: true,
      layers: true,
    };

    // Ignore optional dependencies warnings
    config.resolve.fallback = {
      ...config.resolve.fallback,
      fs: false,
      net: false,
      tls: false,
      crypto: false,
    };

    return config;
  },

  // TypeScript and ESLint configuration
  typescript: {
    // Allow production builds to succeed even if there are type errors
    ignoreBuildErrors: process.env.NODE_ENV === 'production',
  },

  eslint: {
    // Warning: This allows production builds to successfully complete even if
    // your project has ESLint errors.
    ignoreDuringBuilds: process.env.NODE_ENV === 'production',
  },
};

module.exports = nextConfig;
