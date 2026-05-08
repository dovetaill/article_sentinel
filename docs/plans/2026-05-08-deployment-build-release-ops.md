# Deployment / Build / Release / Ops Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a production-ready artifact build, release packaging, Nginx/systemd templates, deploy/rollback scripts, and operator documentation for `article-sentinel`, while preserving local development workflows.

**Architecture:** Keep the existing Go multi-entrypoint layout and the standalone Vite admin app. Add a release pipeline that builds four Go binaries plus a static admin bundle, assembles a versioned release tree, and deploys it to `/srv/article-sentinel/releases/<version>` with `/srv/article-sentinel/current` controlled by scripts and systemd. Run the externally reachable surface through same-domain Nginx and manage long-running processes with dedicated systemd units and an aggregate target.

**Tech Stack:** Go, React, Vite, GNU Make, Bash, Nginx, systemd, Jenkins Pipeline, journald.

---

### Task 1: Add version and artifact directory conventions

**Files:**
- Modify: `Makefile`

**Step 1: Add the release metadata variables**

Define stable Makefile variables near the top of `Makefile` for:

- `VERSION ?=` resolved from a tag or short commit fallback
- `COMMIT ?=` resolved from `git rev-parse --short HEAD`
- `BUILD_TIME ?=` resolved once in ISO-8601 format
- `TARGET_OS ?= linux`
- `TARGET_ARCH ?= amd64`
- `BUILD_DIR := build`
- `RELEASE_DIR := release`
- `PACKAGE_ROOT := $(BUILD_DIR)/package/article-sentinel_$(VERSION)_$(TARGET_OS)_$(TARGET_ARCH)`

**Step 2: Add the metadata inspection target**

Add `print-version` to echo the resolved values.

Run: `make print-version`
Expected: prints non-empty `VERSION`, `COMMIT`, `BUILD_TIME`, `TARGET_OS`, `TARGET_ARCH`.

**Step 3: Commit**

```bash
git add Makefile
git commit -m "build: add release metadata conventions"
```

### Task 2: Build the four Go binaries into a stable directory

**Files:**
- Modify: `Makefile`

**Step 1: Add per-binary build targets**

Add:

- `build-server`
- `build-worker`
- `build-scheduler`
- `build-migrate`

Each target should emit into `build/bin/linux-amd64/` with names:

- `article-sentinel-server`
- `article-sentinel-worker`
- `article-sentinel-scheduler`
- `article-sentinel-migrate`

Use `GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH) CGO_ENABLED=0 go build -trimpath` and inject metadata with `-ldflags` if a build info package is introduced.

**Step 2: Add the aggregate Go build target**

Add `build-go` depending on the four targets.

Run: `make build-go`
Expected: four executable files appear under `build/bin/linux-amd64/`.

**Step 3: Verify the binaries exist**

Run: `ls -1 build/bin/linux-amd64`
Expected:

- `article-sentinel-server`
- `article-sentinel-worker`
- `article-sentinel-scheduler`
- `article-sentinel-migrate`

**Step 4: Commit**

```bash
git add Makefile
git commit -m "build: add Go release binaries"
```

### Task 3: Build the admin static bundle into a stable directory

**Files:**
- Modify: `Makefile`
- Verify: `web/admin/package.json`

**Step 1: Add the admin build target**

Add `build-admin` to:

- run `npm ci` in `web/admin`
- run `npm run build`
- copy the final `web/admin/dist` into `build/admin-dist`

**Step 2: Add the aggregate build target**

Add `build` depending on `build-go` and `build-admin`.

Run: `make build-admin`
Expected: `build/admin-dist/index.html` exists and `build/admin-dist/assets/` is populated.

**Step 3: Verify the aggregate build**

Run: `make build`
Expected: Go binaries and admin bundle both exist under `build/`.

**Step 4: Commit**

```bash
git add Makefile
git commit -m "build: add admin release bundle target"
```

### Task 4: Assemble the release tree and tarball

**Files:**
- Modify: `Makefile`
- Create: `scripts/package_release.sh`
- Create: `scripts/write_release_manifest.sh`

**Step 1: Write the packaging helper**

Create `scripts/package_release.sh` to assemble:

```text
build/package/article-sentinel_<version>_linux_amd64/
  bin/
  admin/
  migrations/
  configs/
  deploy/
```

The script should:

- remove any prior package root
- copy `build/bin/linux-amd64/*` into `bin/`
- copy `build/admin-dist/*` into `admin/`
- copy `migrations/*.sql` into `migrations/`
- copy `configs/config.example.yaml` into `configs/`
- copy `deploy/` assets into `deploy/`

**Step 2: Write the manifest helper**

Create `scripts/write_release_manifest.sh` to emit:

- `manifest.json`
- `manifest.sha256`

Include at least:

- `app`
- `version`
- `git_sha`
- `build_time`
- `target_os`
- `target_arch`

**Step 3: Add Makefile wrappers**

Add:

- `package`: depends on `build`, deploy assets, and the two helper scripts
- `release`: depends on `package`, then creates `release/article-sentinel_<version>_linux_amd64.tar.gz`

Run: `make release`
Expected:

- the packaged directory exists under `build/package/`
- `release/article-sentinel_<version>_linux_amd64.tar.gz` exists
- `manifest.json` and `manifest.sha256` exist inside the package root

**Step 4: Commit**

```bash
git add Makefile scripts/package_release.sh scripts/write_release_manifest.sh
git commit -m "build: add release package and manifest generation"
```

### Task 5: Add deploy templates for Nginx and systemd

**Files:**
- Create: `deploy/nginx/article-sentinel.conf`
- Create: `deploy/systemd/article-sentinel.target`
- Create: `deploy/systemd/article-sentinel-server.service`
- Create: `deploy/systemd/article-sentinel-worker.service`
- Create: `deploy/systemd/article-sentinel-scheduler.service`
- Create: `deploy/systemd/article-sentinel-migrate@.service`

**Step 1: Add the Nginx template**

Create `deploy/nginx/article-sentinel.conf` with:

- `root /srv/article-sentinel/current/admin`
- `/api/` proxy to `127.0.0.1:8080`
- `/auth/` proxy to `127.0.0.1:8080`
- `/healthz` and `/readyz` proxy pass
- `try_files $uri $uri/ /index.html`
- long cache for `/assets/`
- no-store cache headers for `index.html`

**Step 2: Add the systemd units**

Use:

- `Type=exec` for server/worker/scheduler
- `Type=oneshot` for migrate
- `WorkingDirectory=/srv/article-sentinel/current` for long-running services
- `WorkingDirectory=/srv/article-sentinel/releases/%i` for migrate
- `EnvironmentFile=-/etc/article-sentinel/article-sentinel.env`
- `StandardOutput=journal`
- a compatibility-safe hardening baseline

**Step 3: Verify the templates package correctly**

Run: `make package`
Expected: `deploy/nginx/` and `deploy/systemd/` are copied into the release tree.

**Step 4: Commit**

```bash
git add deploy/nginx/article-sentinel.conf deploy/systemd
git commit -m "deploy: add nginx and systemd templates"
```

### Task 6: Add deploy, activate, and rollback scripts

**Files:**
- Create: `deploy/scripts/install-release.sh`
- Create: `deploy/scripts/activate-release.sh`
- Create: `deploy/scripts/rollback-release.sh`
- Create: `deploy/README.md`

**Step 1: Add the install script**

Create `install-release.sh` to:

- accept a tarball path and version
- verify checksum
- extract into `/srv/article-sentinel/releases/<version>`
- update `/srv/article-sentinel/previous` bookkeeping without switching `current`

**Step 2: Add the activate script**

Create `activate-release.sh` to:

- run `systemctl start article-sentinel-migrate@<version>.service`
- repoint `/srv/article-sentinel/current`
- restart `article-sentinel-server`, `article-sentinel-worker`, `article-sentinel-scheduler`
- run smoke checks against `/healthz`, `/readyz`, and `/`

**Step 3: Add the rollback script**

Create `rollback-release.sh` to:

- detect the rollback target or accept one explicitly
- repoint `current`
- restart the long-running services
- run smoke checks again

**Step 4: Verify the scripts are self-documented**

Run: `bash deploy/scripts/install-release.sh --help`
Expected: usage text prints without side effects.

**Step 5: Commit**

```bash
git add deploy/scripts deploy/README.md
git commit -m "deploy: add release activation and rollback scripts"
```

### Task 7: Update production-facing config and deployment docs

**Files:**
- Modify: `README.md`
- Create: `docs/ops/deploy.md`
- Create: `docs/ops/runtime.md`
- Create: `docs/ops/secrets.md`
- Create: `docs/ops/nginx.md`

**Step 1: Update the repository README**

Add sections covering:

- the new build/package/release targets
- the `/srv/article-sentinel` release layout
- the `/etc/article-sentinel/config.yaml` convention
- the fact that shared environments and prod are artifact-only

**Step 2: Add operator docs**

Document at least:

- deploy order
- migration order
- rollback order
- common `systemctl` / `journalctl` commands
- Nginx reload steps
- checksum verification

**Step 3: Verify documentation commands**

Run:

- `make print-version`
- `make build`
- `make release`

Expected: every command referenced in the docs exists and works.

**Step 4: Commit**

```bash
git add README.md docs/ops
git commit -m "docs: add release and operator runbooks"
```

### Task 8: Harden readiness, secrets, and production config handling

**Files:**
- Modify: `internal/api/handlers/ready.go`
- Modify: `internal/api/handlers/health_test.go` or relevant ready tests
- Modify: `pkg/config/config.go`
- Modify: `pkg/config/load.go`
- Modify: `pkg/database/mysql.go`
- Modify: `pkg/database/redis.go`
- Modify: `configs/config.example.yaml`

**Step 1: Strengthen `/readyz`**

Change `/readyz` so it can perform live DB and Redis probes with short timeouts instead of only reporting “configured.”

Run: `go test ./internal/api/handlers ./internal/api/register`
Expected: readiness tests pass with the new semantics.

**Step 2: Add file-backed secret support**

Introduce config fields or helper logic for file-based secrets such as:

- session secret file
- legacy session secret file
- MySQL password file
- Redis password file

Use the file content to override in-memory secret fields when present.

Run: `go test ./pkg/config ./pkg/database`
Expected: tests cover secret-file precedence and malformed-file failures.

**Step 3: Set a production-safe config example**

Add a production example or adjust docs so they explicitly recommend:

- `app.host: 127.0.0.1`
- `auth.session.secure_cookie: true`
- `log.output: stdout`
- docs disabled or intentionally scoped

**Step 4: Commit**

```bash
git add internal/api/handlers/ready.go pkg/config pkg/database configs/config.example.yaml
git commit -m "runtime: harden readiness and secret loading"
```

### Task 9: Add Jenkins pipeline assets and deployment policy docs

**Files:**
- Create: `Jenkinsfile.build`
- Create: `Jenkinsfile.deploy`
- Create: `docs/ops/jenkins.md`

**Step 1: Add the build pipeline**

Create `Jenkinsfile.build` that:

- checks out the requested ref
- runs backend and frontend tests
- runs `make release`
- archives and fingerprints the artifact outputs
- publishes them to the chosen artifact destination

**Step 2: Add the deploy pipeline**

Create `Jenkinsfile.deploy` that:

- accepts `VERSION` and target environment inputs
- uses an approval gate for shared environments and prod
- transfers or downloads the selected artifact
- calls `install-release.sh` and `activate-release.sh`
- calls `rollback-release.sh` when smoke checks fail

**Step 3: Document the governance rules**

Add `docs/ops/jenkins.md` describing:

- build once, deploy many
- production forbids source-tree deployment
- shared environments must use artifact promotion
- artifact metadata and fingerprint expectations

**Step 4: Commit**

```bash
git add Jenkinsfile.build Jenkinsfile.deploy docs/ops/jenkins.md
git commit -m "ci: add build and deploy pipeline templates"
```

### Task 10: Address the known auth and proxy trust risks

**Files:**
- Modify: `internal/api/handlers/auth.go`
- Modify: `internal/middleware/auth.go`
- Modify: `README.md`
- Modify: `docs/ops/nginx.md`

**Step 1: Replace or deprecate query-string JWT login**

Implement one of these minimal safe follow-ups:

- deprecate the direct `jwt` query bridge and document a replacement code exchange flow, or
- add an intermediate one-time code exchange endpoint and reduce exposure of bearer material in URLs

**Step 2: Add trusted-proxy handling**

Make forwarded IP handling configurable so the backend does not blindly trust arbitrary `X-Forwarded-For` input.

Run: `go test ./internal/api/handlers ./internal/middleware`
Expected: login and proxy-trust tests pass.

**Step 3: Commit**

```bash
git add internal/api/handlers/auth.go internal/middleware/auth.go README.md docs/ops/nginx.md
git commit -m "security: harden login bridge and proxy trust"
```

### Task 11: Final cross-check and release rehearsal

**Files:**
- Verify only

**Step 1: Run the repo verification suite**

Run: `make test`
Expected: Go tests pass.

**Step 2: Run the admin tests**

Run: `cd web/admin && npm test`
Expected: all Vitest suites pass.

**Step 3: Run the full release rehearsal**

Run:

```bash
make clean
make release
```

Expected:

- clean build directories are recreated
- release tarball is produced
- release tree contains `bin/`, `admin/`, `migrations/`, `deploy/`, `manifest.json`, `manifest.sha256`

**Step 4: Smoke-check the generated artifact**

Run:

```bash
tar -tf release/article-sentinel_<version>_linux_amd64.tar.gz | head -n 50
```

Expected: top-level artifact tree matches the documented layout.

**Step 5: Commit**

```bash
git add .
git commit -m "chore: finish deployment and release tooling"
```
