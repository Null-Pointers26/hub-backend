# ── Build stage ──────────────────────────────────────────────────────────────
FROM node:20-alpine AS builder

# Enable corepack to get pnpm
RUN corepack enable && corepack prepare pnpm@latest --activate

WORKDIR /app

COPY package.json pnpm-lock.yaml* ./
RUN pnpm install --frozen-lockfile || pnpm install

COPY tsconfig.json build-inline.js ./
COPY src ./src

# Compile TypeScript → JS bundle
RUN pnpm run build

# Inline CSS and JS into a single self-contained index.html
RUN node build-inline.js

# ── Production stage ──────────────────────────────────────────────────────────
FROM nginx:alpine

# curl is needed for the hub registration script
RUN apk add --no-cache curl

COPY --from=builder /app/dist/index.html /usr/share/nginx/html/index.html
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY register-and-run.sh /register-and-run.sh
RUN chmod +x /register-and-run.sh

EXPOSE 80

CMD ["/register-and-run.sh"]
