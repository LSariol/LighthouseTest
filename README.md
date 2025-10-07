# LighthouseTest

Small Go web service exposing port `2001` with simple versioning. Build and run via Docker Compose.

## Quick Start

1. Build and run the service:

   - `docker compose up --build`

2. Open in your browser:

   - http://localhost:2001/
   - http://localhost:2001/healthz
   - http://localhost:2001/version

3. Stop:

   - `docker compose down`

## Change the Version

You have three easy options. Priority: runtime env > build arg > embedded file.

- Runtime env (no rebuild, just restart):
  - `APP_VERSION=1.2.3 docker compose up --build` (or set in `docker-compose.yml`)

- Build-time arg (bake into the image):
  - Uncomment `VERSION_ARG` under `build.args` in `docker-compose.yml` and set a value.
  - Or: `docker build --build-arg VERSION_ARG=1.2.3 -t lighthouse-test .`

- Edit the embedded file and rebuild:
  - Change `VERSION` and run `docker compose up --build`.
  - Optional: mount the file so edits apply on restart without rebuild — uncomment the `volumes` entry in `docker-compose.yml`.

## Config

- `PORT` (env): listening port (default `2001`).
- `APP_VERSION` (env): overrides version at runtime.

## Project Layout

- `cmd/server/main.go` — HTTP server and handlers.
- `internal/version/version.go` — version resolution (env > embedded file > `dev`).
- `VERSION` — default version string.
- `Dockerfile` — multi-stage build.
- `docker-compose.yml` — local build + run.

## Local (without Docker)

```bash
go run ./cmd/server
# or
PORT=2001 APP_VERSION=dev go run ./cmd/server
```

