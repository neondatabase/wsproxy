# wsproxy

Lightweight websocket->TCP proxy. Look at `main.go` for available configuration options.

Run:

```bash
docker run -d -p 80:80 -p 2112:2112 --name wsproxy ghcr.io/neondatabase/wsproxy:latest
```
