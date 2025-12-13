Local Development - Docker

This repository includes a `Makefile` with convenient Docker-based targets for running Postgres and Redis locally during development.

Usage:

- Create both services: `make create` (runs `create-db` and `create-redis`)
- Create only Postgres: `make create-db`
- Create only Redis: `make create-redis`
- Start services: `make start`
- Stop services: `make stop`
- Delete containers and volumes: `make delete`
- Attach to logs: `make logs`
- PSQL shell: `make psql`
- Redis shell: `make redis-cli`

You can override defaults with command-line variable overrides. Example:

```
make create PG_PASSWORD=secret PG_PORT=15432
```

Defaults:

- Postgres container: `fjord-pg` (port 5432)
- Redis container: `fjord-redis` (port 6379)
- Data volumes: `fjord-pg-data`, `fjord-redis-data`
