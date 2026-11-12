# syntax=docker/dockerfile:1
# Build context is frontend/. Stage 1 builds the static bundle with Node; stage 2
# serves it with nginx and proxies /api to the backend service.

# Base images are pinned by digest (tag kept for readability) so a moved tag cannot change the build.
# Keep the Node major in step with frontend/.nvmrc.
FROM node:24-alpine@sha256:116b800c3754b440ee0aa3ed2f20f5bf3d99a554bfd9adeea09ea8017f6e85f5 AS build
WORKDIR /app

# Lockfile first: dependencies are reinstalled only when package.json or the lockfile change.
COPY package.json package-lock.json ./
RUN npm ci

COPY . .
# VITE_API_BASE_URL is deliberately unset: the browser calls /api on its own origin and nginx forwards it.
RUN npm run build

# The unprivileged image runs as a non-root user and listens on 8080 (ports below 1024 need root).
FROM nginxinc/nginx-unprivileged:1.29-alpine@sha256:0c79d56aee561a1d81c63f00eee5fb5fe29279560cdc55e91425133104c7fbe6
COPY nginx/default.conf /etc/nginx/conf.d/default.conf
COPY nginx/security-headers.conf /etc/nginx/snippets/security-headers.conf
COPY --from=build /app/dist /usr/share/nginx/html

EXPOSE 8080
# No HEALTHCHECK here: docker-compose.yml owns it, so there is one definition to keep correct.
