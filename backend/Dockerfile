# syntax=docker/dockerfile:1
# Build context is backend/. Two stages: compile static binaries, then copy them
# into a distroless image with no shell, no package manager and a non-root user.

# Base images are pinned by digest (tag kept for readability) so a moved tag cannot change the build.
FROM golang:1.26.11-alpine@sha256:ae52b5bc6ed045a9ed26ab03dc19327a802a9cc9aa8b27ed57f5c6b60579b0bb AS build
WORKDIR /src

# Module files first: this layer is rebuilt only when dependencies change, not on every code edit.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# CGO off gives a fully static binary that runs on distroless/static (no libc).
# -trimpath removes build paths; -s -w strips symbols and debug info to shrink the binary.
# healthcheck is a second tiny binary because distroless has no shell or curl for a HEALTHCHECK.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api \
 && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/healthcheck ./cmd/healthcheck

FROM gcr.io/distroless/static-debian12:nonroot@sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab
WORKDIR /app
COPY --from=build /out/api /out/healthcheck /app/
COPY config /app/config

# APP_ENV=prod layers config/prod.toml on default.toml. APP_CONFIG_DIR points the loader at the copied files.
ENV APP_ENV=prod \
    APP_CONFIG_DIR=/app/config

# Never run as root: a compromised process then has no write access outside its own files.
USER nonroot:nonroot
EXPOSE 8080

# No HEALTHCHECK here: docker-compose.yml owns it (it runs /app/healthcheck), so there is one definition.

ENTRYPOINT ["/app/api"]
