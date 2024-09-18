# wsproxy

Lightweight websocket->TCP proxy. Look at `main.go` for available configuration options.

Read more about how to deploy with Neon Serverless proxy: https://github.com/neondatabase/serverless/blob/main/DEPLOY.md

Run:

```bash
docker run -d -p 80:80 -p 2112:2112 --name wsproxy ghcr.io/neondatabase/wsproxy:latest
```
