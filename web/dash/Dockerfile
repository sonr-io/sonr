# Use Node.js Alpine as base image
FROM node:20-alpine AS base

# Install dependencies only when needed
FROM base AS deps
RUN apk add --no-cache libc6-compat
WORKDIR /app

# Install pnpm
RUN corepack enable && corepack prepare pnpm@latest --activate

# Copy root package files and workspace configuration
COPY ../../package.json ../../pnpm-lock.yaml ../../pnpm-workspace.yaml ./
COPY ../../turbo.json ./

# Copy package.json files for all workspace dependencies
COPY ../../packages/es/package.json ./packages/es/
COPY ../../packages/sdk/package.json ./packages/sdk/
COPY ../../packages/com/package.json ./packages/com/
COPY ../../packages/ui/package.json ./packages/ui/
COPY ../../packages/pkl/package.json ./packages/pkl/
COPY ../../packages/cli/package.json ./packages/cli/
COPY ./package.json ./web/dash/

# Install dependencies
RUN pnpm install --frozen-lockfile

# Rebuild the source code only when needed
FROM base AS builder
WORKDIR /app

# Install pnpm
RUN corepack enable && corepack prepare pnpm@latest --activate

# Copy installed dependencies
COPY --from=deps /app/node_modules ./node_modules
COPY --from=deps /app/.pnpm ./.pnpm

# Copy source code
COPY ../.. .

# Build the application
WORKDIR /app/web/dash
RUN pnpm build

# Production image, copy all the files and run next
FROM base AS runner
WORKDIR /app

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1

# Create non-root user
RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs

# Copy built application
COPY --from=builder /app/web/dash/public ./public
COPY --from=builder --chown=nextjs:nodejs /app/web/dash/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/web/dash/.next/static ./.next/static

USER nextjs

# Expose port
EXPOSE 3000

ENV PORT=3000
ENV HOSTNAME="0.0.0.0"

# Start the application
CMD ["node", "server.js"]