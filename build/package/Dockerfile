# --- frontend: build the embedded admin-panel SPA -------------------------
FROM node:22-bookworm AS frontend

WORKDIR /app/web

# corepack ships with node; it activates the pnpm version pinned in
# package.json's "packageManager" field.
RUN corepack enable

# Install dependencies first so this layer is cached across source changes.
COPY web/package.json web/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

COPY web/ ./
RUN pnpm run build

# --- builder: compile the Go binary with the SPA embedded -----------------
FROM golang:1.26-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Overwrite the committed web/dist with the freshly built assets so the
# embedded UI always matches the source in this build.
COPY --from=frontend /app/web/dist ./web/dist

WORKDIR /app/cmd/koochooloo
RUN go build -o /koochooloo

# --- release --------------------------------------------------------------
FROM gcr.io/distroless/base-debian12:nonroot AS release

LABEL maintainer="Parham Alvani <parham.alvani@gmail.com>"

WORKDIR /app/

COPY --from=builder /koochooloo .

EXPOSE 1378

ENTRYPOINT ["./koochooloo"]

CMD ["server"]
