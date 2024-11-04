# with-unix-socket

Run Postgres instance with Unix socket and connect to it using Neon's WS Proxy and Drizzle ORM.

This project was created using `bun init` in bun v1.1.27. [Bun](https://bun.sh) is a fast all-in-one JavaScript runtime.

To install dependencies:

```bash
bun install
```

To run:

```bash
docker compose up -d
NEON_WS_PROXY_HOST=$(docker compose port neon 80) bun run index.ts
```

To clean up:

```bash
docker compose down
```
