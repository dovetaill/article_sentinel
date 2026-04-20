# Docker Compose Env Config Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace the hard-coded local dependency setup with a `.env`-driven Docker Compose configuration for MySQL 5.7 and Redis, including remote MySQL access, remote root access, and a dedicated application user with full privileges on the configured database.

**Architecture:** Keep a single root `docker-compose.yml` and introduce a root `.env.example`. Use Compose variable substitution for ports and credentials, rely on the official MySQL initialization environment variables to provision the database and app user, and update the README so local startup instructions match the new workflow.

**Tech Stack:** Docker Compose, MySQL 5.7, Redis 7.2, dotenv-style environment files.

---

### Task 1: Parameterize the compose services

**Files:**
- Modify: `docker-compose.yml`

**Step 1: Write the expected rendered config contract**

Document the intended compose behavior inside the file changes:

- MySQL image is `mysql:5.7`
- Redis image remains `redis:7.2`
- MySQL host port comes from `MYSQL_PORT`
- Redis host port comes from `REDIS_PORT`
- MySQL root password, database, app user, and app password come from `.env`
- Redis password comes from `.env`
- MySQL enables remote root access with `MYSQL_ROOT_HOST=%`

**Step 2: Implement the minimal compose changes**

Update `docker-compose.yml` to:

- replace hard-coded values with `${...}` substitutions
- add `MYSQL_ROOT_HOST: "%"`
- add `MYSQL_USER` / `MYSQL_PASSWORD`
- add password-aware Redis command and healthcheck

**Step 3: Verify the compose file renders**

Run: `docker compose config`
Expected: the compose file parses successfully with all variables resolved.

### Task 2: Add an environment template

**Files:**
- Create: `.env.example`

**Step 1: Add the example variables**

Create `.env.example` with example values for:

- `MYSQL_PORT`
- `MYSQL_ROOT_PASSWORD`
- `MYSQL_DATABASE`
- `MYSQL_APP_USER`
- `MYSQL_APP_PASSWORD`
- `REDIS_PORT`
- `REDIS_PASSWORD`

**Step 2: Keep secrets out of version control**

Confirm the repository ignore rules already exclude `.env` and related local env files.

**Step 3: Re-verify compose expectations**

Run: `cp .env.example .env && docker compose config`
Expected: compose renders successfully using the example values.

### Task 3: Update operator documentation

**Files:**
- Modify: `README.md`

**Step 1: Update local startup docs**

Adjust the local run section so it tells operators to:

- copy `.env.example` to `.env`
- start dependencies with `make up`
- read ports and credentials from `.env`

**Step 2: Update connection examples**

Replace hard-coded examples with `.env`-aligned examples for:

- MySQL root access
- MySQL app user access
- Redis password-authenticated access
- demo seed import using the configured root credentials

**Step 3: Final verification**

Run: `docker compose --env-file .env.example config`
Expected: exit code 0 and rendered services for MySQL and Redis.
