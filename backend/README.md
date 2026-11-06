# Rewind backend

Run the server from this directory:

```sh
go run ./cmd/rewind serve --listen 0.0.0.0:8080 --data-dir /data
```

Use `rewind --help` or `rewind serve --help` for command-line help.

| Flag | Default | Description |
| --- | --- | --- |
| `--listen` | `0.0.0.0:8080` | Server listen address |
| `--data-dir` | `/data` | Data directory, reserved for the planned storage feature |

`rewind version` prints the version. Set it at build time with
`go build -ldflags "-X main.Version=1.0.0" ./cmd/rewind`.
