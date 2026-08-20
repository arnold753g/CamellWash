FROM node:22-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM node:22-alpine
RUN apk add --no-cache dumb-init wget
WORKDIR /app
COPY --from=builder --chown=node:node /app/.output /app/.output
USER node
ENV NODE_ENV=production NUXT_HOST=0.0.0.0 NUXT_PORT=3000
EXPOSE 3000
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 CMD wget --quiet --tries=1 --spider http://localhost:3000/ || exit 1
ENTRYPOINT ["dumb-init", "--"]
CMD ["node", ".output/server/index.mjs"]

