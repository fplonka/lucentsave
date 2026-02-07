# Deployment

lucentsave runs on a Hetzner VPS (Debian 13) as Docker containers behind Caddy.

## What's running

Two containers managed by Docker Compose:

- **app** — Go binary + Node.js subprocess (article extraction). Listens on `127.0.0.1:8080`.
- **db** — Postgres 16 with pgvector. Data lives in a Docker volume called `lucentsave_pgdata`.

Caddy runs on the host (not in Docker) as a reverse proxy. It handles TLS certificates automatically via Let's Encrypt.

## Paths on the VPS

- `/opt/lucentsave` — the repo
- `/opt/lucentsave/.env` — secrets (DB_PASSWORD, JWT_SECRET, LS2_OPENAI_KEY)
- `/etc/caddy/Caddyfile` — Caddy config

## How deploys work

Push to `main` triggers GitHub Actions which SSHes into the VPS and runs:

```
cd /opt/lucentsave && git pull && docker compose up -d --build
```

This rebuilds the app image and restarts the container. The DB container is unaffected unless its config in `docker-compose.yml` changed. Downtime is a few seconds while the image builds.

You can also deploy manually by SSH-ing in and running the same command.

## Common commands

All of these assume you've SSH-ed into the VPS.

```bash
# status
docker compose -f /opt/lucentsave/docker-compose.yml ps

# logs (follow)
docker compose -f /opt/lucentsave/docker-compose.yml logs -f app

# restart just the app
docker compose -f /opt/lucentsave/docker-compose.yml restart app

# psql shell
docker compose -f /opt/lucentsave/docker-compose.yml exec db psql -U postgres -d lucentsave

# full rebuild
cd /opt/lucentsave && docker compose up -d --build
```

## Database

Postgres is initialized automatically on first start using `init_db_docker.sql`. This creates the `users` and `posts` tables, full-text search indexes, the pgvector extension, and the HNSW index on the embedding column.

The DB data persists in the `lucentsave_pgdata` Docker volume. `docker compose down` preserves it. `docker compose down -v` deletes it (fresh start).

There's no migration system — schema changes need to be applied manually via `psql`.

## Secrets

The `.env` file is not in git. It contains:

- `DB_PASSWORD` — Postgres password (used in the connection string)
- `JWT_SECRET` — signs auth tokens
- `LS2_OPENAI_KEY` — OpenAI API key for embeddings/search

## Adding another app behind Caddy

Edit `/etc/caddy/Caddyfile` and add a block:

```
whatever.fplonka.dev {
    reverse_proxy localhost:<port>
}
```

Then `systemctl reload caddy`. Caddy provisions a TLS cert automatically.
