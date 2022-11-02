# wsproxy

Lightweight websocket->TCP proxy. Look at `main.go` for available configuration options.

Build:

```bash
docker build -t wsproxy .
```

Run:

```bash
docker run -d -p 80:80 --name wsproxy ghcr.io/petuhovskiy/wsproxy:latest
```
