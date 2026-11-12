# syntax=docker/dockerfile:1
# Build context is frontend/. Stage 1 builds the static bundle with Node; stage 2
# serves it with nginx and proxies /api to the backend service.

FROM node:24-alpine AS build
WORKDIR /app

# Lockfile first: dependencies are reinstalled only when package.json or the lockfile change.
COPY package.json package-lock.json ./
RUN npm ci

COPY . .
# VITE_API_BASE_URL is deliberately unset: the browser calls /api on its own origin and nginx forwards it.
RUN npm run build

# The unprivileged image runs as a non-root user and listens on 8080 (ports below 1024 need root).
FROM nginxinc/nginx-unprivileged:1.29-alpine
COPY nginx/default.conf /etc/nginx/conf.d/default.conf
COPY nginx/security-headers.conf /etc/nginx/snippets/security-headers.conf
COPY --from=build /app/dist /usr/share/nginx/html

EXPOSE 8080
# No HEALTHCHECK here: docker-compose.yml owns it, so there is one definition to keep correct.
