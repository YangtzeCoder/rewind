# Rewind backend

Run the server from this directory:

```sh
go run ./cmd/rewind serve --listen 0.0.0.0:8080 --data-dir /data
```

Use `rewind --help` or `rewind serve --help` for command-line help.

| Flag | Default | Description |
| --- | --- | --- |
| `--listen` | `0.0.0.0:8080` | Server listen address |
| `--data-dir` | `/data` | Data directory for the SQLite metadata database |

On startup, the server creates `rewind.db` in the data directory and applies
the database schema. The SQLite driver does not require CGO.

The first startup creates the administrator `admin` with the initial password
`123456` if no administrator exists. Change this password after deployment.
Authentication uses `/api/v1/auth/login`, `/logout`, `/password`, and `/me`.
Login returns a bearer token and its initial expiry; authenticated requests
extend its idle expiry by 30 minutes. Logout revokes the current token. A
password change leaves existing login tokens valid and applies to future logins.

JSON responses use HTTP 200 and include `success`. Successful responses include
`data` when there is a result; array results are always arrays, including `[]`
for an empty result. Errors instead include an `error` object with string `code`
and `message` fields. File downloads return binary data on success and JSON on
failure. The `http/` directory contains request files for the VS Code REST Client.

`rewind version` prints the version. Set it at build time with
`go build -ldflags "-X main.Version=1.0.0" ./cmd/rewind`.
