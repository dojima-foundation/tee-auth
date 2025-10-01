/** @type {import('next').NextConfig} */
const nextConfig = {
  // Enable standalone output for Docker deployment
  output: 'standalone',
  
  // Optimize for production
  experimental: {
    // Enable server components
    serverComponentsExternalPackages: [],
  },
  
  // Image optimization
  images: {
    unoptimized: true, // Disable for Docker deployment
  },
  
  // Environment variables - expose to client-side
  env: {
    CUSTOM_KEY: process.env.CUSTOM_KEY,
    NEXT_PUBLIC_GAUTH_API_URL: process.env.NEXT_PUBLIC_GAUTH_API_URL,
    NEXT_PUBLIC_API_URL: process.env.NEXT_PUBLIC_API_URL,
    NEXT_PUBLIC_GRPC_URL: process.env.NEXT_PUBLIC_GRPC_URL,
  },
};

module.exports = nextConfig;
