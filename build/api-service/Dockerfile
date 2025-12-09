# ===========================
# 1. Builder Stage
# ===========================
FROM node:20-alpine AS builder

WORKDIR /app

# Only copy package files first for better caching
COPY package*.json ./

# Install only production dependencies first (if lock exists)
RUN npm install

# Copy source code AFTER dependencies to leverage caching
COPY . .

# There is no build step in our case
#RUN npm run build

# ===========================
# 2. Runner Stage
# ===========================
FROM node:20-alpine AS runner

WORKDIR /app

# Copy only needed artifacts from builder
COPY --from=builder /app ./

# Security: create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

ENV NODE_ENV=production
EXPOSE 3000

CMD ["node", "app.js"]