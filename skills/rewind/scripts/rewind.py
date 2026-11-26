#!/usr/bin/env python3
"""Small stdlib client for the Rewind agent API."""
import argparse
import http.client
import getpass
import json
import mimetypes
import os
from pathlib import Path
import sys
import tempfile
from urllib.parse import quote, urlsplit
import uuid
import warnings


class RewindError(Exception):
    pass


def config_path(root):
    return Path(root) / ".rewind" / "config.json"


def global_config_path():
    return Path.home() / ".rewind" / "config.json"


def write_config(path, data, add_gitignore=False):
    path.parent.mkdir(mode=0o700, parents=True, exist_ok=True)
    if add_gitignore:
        ignore = path.parent / ".gitignore"
        content = ignore.read_text(encoding="utf-8") if ignore.exists() else ""
        # Append rules last so existing negations cannot expose credentials or temporary files.
        if content.splitlines()[-2:] != ["config.json", ".config-*"]:
            with ignore.open("a", encoding="utf-8") as f:
                f.write(("\n" if content and not content.endswith("\n") else "") + "config.json\n.config-*\n")
    temp = None
    try:
        with tempfile.NamedTemporaryFile(mode="w", encoding="utf-8", dir=path.parent, prefix=".config-", delete=False) as f:
            temp = Path(f.name)
            json.dump(data, f)
        os.replace(temp, path)
        temp = None
    finally:
        if temp is not None:
            temp.unlink(missing_ok=True)


def read_config(path, label):
    try:
        data = json.loads(path.read_text(encoding="utf-8"))
    except FileNotFoundError:
        return None
    except (OSError, ValueError, UnicodeDecodeError):
        raise RewindError(f"{label} configuration is unreadable or invalid")
    if not isinstance(data, dict):
        raise RewindError(f"{label} configuration is invalid")
    for key in ("url", "project_id", "api_token"):
        if key not in data:
            continue
        value = data[key]
        if not isinstance(value, str) or not value.strip():
            raise RewindError(f"{label} configuration has an invalid {key}")
        if key == "url":
            try:
                validate_url(value)
            except ValueError as exc:
                raise RewindError(f"{label} configuration has an invalid URL: {exc}")
        if key == "api_token" and any(c in value for c in "\r\n"):
            raise RewindError(f"{label} configuration has an invalid API token")
    return data


def load_config(root):
    local = read_config(config_path(root), "Project")
    if local is None:
        local = {}
    missing = [key for key in ("url", "project_id", "api_token") if key not in local]
    global_config = read_config(global_config_path(), "Global") if missing else {}
    if global_config is None:
        global_config = {}
    config = dict(global_config)
    config.update(local)
    missing = [key for key in ("url", "project_id", "api_token") if key not in config]
    if missing:
        fields = ", ".join(missing)
        raise RewindError(
            f"Rewind configuration is missing {fields}; ask the Agent to configure it"
        )
    return config["url"], config["project_id"], config["api_token"]


def read_api_token(token_stdin):
    if token_stdin:
        if sys.stdin.isatty():
            raise RewindError("Standard input must be a secure non-interactive stream with --token-stdin")
        token = sys.stdin.read(4097)
        if len(token) > 4096:
            raise RewindError("API token input is too long")
        token = token.removesuffix("\n").removesuffix("\r")
    else:
        if not sys.stdin.isatty():
            raise RewindError("An interactive terminal or a secure standard input stream with --token-stdin is required")
        try:
            with warnings.catch_warnings():
                warnings.simplefilter("error", getpass.GetPassWarning)
                token = getpass.getpass("Rewind API token (input hidden): ")
        except getpass.GetPassWarning:
            raise RewindError("Cannot hide API token input")
    if not token.strip() or any(c in token for c in "\r\n"):
        raise RewindError("A non-empty API token is required")
    return token


def validate_url(url):
    parts = urlsplit(url)
    if parts.scheme not in ("http", "https") or not parts.hostname or parts.username or parts.password or parts.query or parts.fragment:
        raise ValueError("URL must be an HTTP(S) service URL without credentials, query, or fragment")
    try:
        _ = parts.port
    except ValueError:
        raise ValueError("Invalid URL port")
    return parts


def connection(url):
    parts = validate_url(url)
    cls = http.client.HTTPSConnection if parts.scheme == "https" else http.client.HTTPConnection
    return cls(parts.hostname, parts.port, timeout=60), parts.path.rstrip("/")


def parse_response(response):
    try:
        payload = json.loads(response.read())
        if not isinstance(payload, dict) or not isinstance(payload.get("success"), bool):
            raise ValueError()
    except (ValueError, UnicodeDecodeError):
        raise RewindError("Server returned an invalid JSON response")
    if not payload["success"]:
        error = payload.get("error") or {}
        raise RewindError(f"{error.get('code', 'ERROR')}: {error.get('message', 'Request failed')}")
    return payload.get("data")


def request(url, token, method, endpoint, body=None, content_type=None):
    conn, base = connection(url)
    try:
        headers = {"Authorization": "Bearer " + token}
        if content_type:
            headers["Content-Type"] = content_type
        conn.request(method, base + endpoint, body=body, headers=headers)
        response = conn.getresponse()
        return parse_response(response)
    finally:
        conn.close()


def file_key(path, root):
    try:
        relative = path.resolve().relative_to(Path(root).resolve())
    except ValueError:
        raise RewindError("File is outside the project root; specify --key explicitly")
    if not relative.parts:
        raise RewindError("File must be inside the project root")
    return relative.as_posix()


def multipart_field(boundary, name, value):
    return (
        f'--{boundary}\r\nContent-Disposition: form-data; name="{name}"\r\n\r\n'.encode()
        + value.encode("utf-8")
        + b"\r\n"
    )


def build_backup_multipart(path, project, key, description):
    boundary = "rewind-" + uuid.uuid4().hex
    fields = (
        multipart_field(boundary, "project_id", project)
        + multipart_field(boundary, "key", key)
        + multipart_field(boundary, "description", description)
    )
    # Escape header parameters while retaining UTF-8 filenames.
    filename = path.name.replace("\\", "\\\\").replace('"', '\\"').replace("\r", "_").replace("\n", "_")
    file_header = (
        f'--{boundary}\r\nContent-Disposition: form-data; name="file"; filename="{filename}"\r\n'
        f'Content-Type: {mimetypes.guess_type(path.name)[0] or "application/octet-stream"}\r\n\r\n'
    ).encode()
    return boundary, fields + file_header, f"\r\n--{boundary}--\r\n".encode()


def backup(url, project, token, path, key, description):
    if not path.is_file():
        raise RewindError("File does not exist or is not a regular file")
    boundary, prefix, suffix = build_backup_multipart(path, project, key, description)
    conn, base = connection(url)
    try:
        with path.open("rb") as source:
            size = os.fstat(source.fileno()).st_size
            conn.putrequest("POST", base + "/api/v1/backups")
            conn.putheader("Authorization", "Bearer " + token)
            conn.putheader("Content-Type", "multipart/form-data; boundary=" + boundary)
            conn.putheader("Content-Length", str(len(prefix) + size + len(suffix)))
            conn.endheaders()
            conn.send(prefix)
            remaining = size
            while remaining:
                chunk = source.read(min(1024 * 1024, remaining))
                if not chunk:
                    raise RewindError("File changed during upload")
                conn.send(chunk)
                remaining -= len(chunk)
            conn.send(suffix)
            return parse_response(conn.getresponse())
    finally:
        conn.close()


def save_download_response(response, output, version, force):
    if not response.getheader("Content-Disposition", "").lower().startswith("attachment"):
        parse_response(response)
        raise RewindError("Server did not return a file")
    expected = response.getheader("Content-Length")
    temp = None
    try:
        with tempfile.NamedTemporaryFile(dir=output.parent, prefix=".rewind-", delete=False) as target:
            temp = Path(target.name)
            count = 0
            while True:
                chunk = response.read(1024 * 1024)
                if not chunk:
                    break
                target.write(chunk)
                count += len(chunk)
        if expected is not None and count != int(expected):
            raise RewindError("Incomplete download")
        if force:
            os.replace(temp, output)
        else:
            os.link(temp, output)  # Atomic no-overwrite, even if the destination appeared meanwhile.
            temp.unlink()
        temp = None
        return {"version_id": version, "output": str(output), "size": count}
    finally:
        if temp is not None:
            temp.unlink(missing_ok=True)


def download(url, project, token, version, output, force):
    output = Path(output)
    if output.exists() and not force:
        raise RewindError("Destination exists; get user confirmation before using --force")
    conn, base = connection(url)
    try:
        endpoint = "/api/v1/backups/download/" + quote(project, safe="") + "/" + quote(version, safe="")
        conn.request("GET", base + endpoint, headers={"Authorization": "Bearer " + token})
        response = conn.getresponse()
        # Successful downloads always have attachment disposition, including JSON files.
        return save_download_response(response, output, version, force)
    finally:
        conn.close()


def build_parser():
    parser = argparse.ArgumentParser(description="Back up and restore files using Rewind")
    parser.add_argument("--project-root", default=".", help="Project directory containing .rewind/config.json (default: current directory)")
    sub = parser.add_subparsers(dest="command", required=True)
    setup = sub.add_parser("configure", help="Save Rewind settings and read the API token securely")
    scope = setup.add_mutually_exclusive_group()
    scope.add_argument("--scope", choices=("shared", "local", "global"), help="Save credentials and project settings using the selected scope")
    scope.add_argument("--global", dest="scope", action="store_const", const="global", help="Save all settings in the user-wide ~/.rewind/config.json")
    setup.add_argument("--url", help="Rewind service URL")
    setup.add_argument("--project", required=True, help="Project ID")
    setup.add_argument("--token-stdin", action="store_true", help="Read the API token from standard input instead of an interactive prompt")
    up = sub.add_parser("backup", help="Back up one or more files")
    group = up.add_mutually_exclusive_group()
    group.add_argument("--root", help="Project root, for relative-path keys")
    group.add_argument("--key", help="Explicit key (one file only)")
    up.add_argument("--description", default="")
    up.add_argument("files", nargs="+")
    hist = sub.add_parser("history", help="List file versions")
    hist.add_argument("--key", required=True)
    hist.add_argument("--page", type=int, default=1)
    hist.add_argument("--page-size", type=int, default=20)
    down = sub.add_parser("download", help="Download a version")
    down.add_argument("--version", required=True)
    down.add_argument("--output", required=True)
    down.add_argument("--force", action="store_true", help="Overwrite an existing destination after user confirmation")
    return parser


def configure_mode(args):
    mode = args.scope
    if mode is None:
        mode = "local" if args.url is not None or args.token_stdin else "project"
    needs_credentials = mode != "project"
    if needs_credentials and args.url is None:
        raise RewindError(f"--scope {mode} requires --url")
    if needs_credentials and not args.token_stdin and not sys.stdin.isatty():
        raise RewindError("An interactive terminal or a secure standard input stream with --token-stdin is required")
    if mode == "project" and (args.url is not None or args.token_stdin):
        raise RewindError("Project-only configuration does not accept credentials")
    return mode


def project_settings_with_fallback(project_path, project_id):
    settings = {"project_id": project_id}
    existing = read_config(project_path, "Project")
    if existing is not None:
        settings = dict(existing)
        settings["project_id"] = project_id
    needs_fallback = any(key not in settings for key in ("url", "api_token"))
    fallback = read_config(global_config_path(), "Global") if needs_fallback else {}
    resolved = dict(fallback or {})
    resolved.update(settings)
    if any(key not in resolved for key in ("url", "api_token")):
        raise RewindError(
            "Project-only configuration requires existing URL and API token settings"
        )
    return settings


def save_config_for_mode(mode, root, project_id, credentials):
    project_path = config_path(root)
    project_settings = {"project_id": project_id}
    if mode == "project":
        project_settings = project_settings_with_fallback(project_path, project_id)
        write_config(project_path, project_settings, add_gitignore=True)
        return {"path": str(project_path), "global": False}
    if mode == "local":
        write_config(project_path, dict(credentials, project_id=project_id), add_gitignore=True)
        return {"path": str(project_path), "global": False}
    if mode == "global":
        path = global_config_path()
        write_config(path, dict(credentials, project_id=project_id))
        return {"path": str(path), "global": True}

    global_path = global_config_path()
    write_config(global_path, credentials)
    write_config(project_path, project_settings, add_gitignore=True)
    return {"path": str(project_path), "global_path": str(global_path), "global": False}


def run_configure(args, root):
    if not args.project.strip():
        raise RewindError("Project ID is required")
    if args.url is not None:
        validate_url(args.url)
    mode = configure_mode(args)
    credentials = {}
    if mode != "project":
        credentials = {
            "url": args.url.rstrip("/"),
            "api_token": read_api_token(args.token_stdin),
        }
    result = save_config_for_mode(mode, root, args.project.strip(), credentials)
    print(json.dumps(dict({"configured": True, "verified": False}, **result)))
    return 0


def run_backup(args, root, url, project, token):
    if args.key is not None and (not args.key or len(args.files) != 1):
        raise RewindError("--key requires a non-empty key and exactly one file")
    results = []
    for name in args.files:
        try:
            path = Path(name)
            key = args.key if args.key is not None else file_key(path, args.root or root)
            data = backup(url, project, token, path, key, args.description)
            results.append({"file": name, "success": True, "data": data})
        except (RewindError, OSError, http.client.HTTPException, ValueError) as exc:
            results.append({"file": name, "success": False, "error": str(exc)})
    print(json.dumps(results, ensure_ascii=False))
    return 0 if all(result["success"] for result in results) else 1


def run_history(args, url, project, token):
    if not args.key or not 1 <= args.page or not 1 <= args.page_size <= 100:
        raise RewindError("Provide a non-empty key, positive page, and page size between 1 and 100")
    body = json.dumps({
        "project_id": project,
        "key": args.key,
        "page": args.page,
        "page_size": args.page_size,
    }).encode()
    result = request(url, token, "POST", "/api/v1/backups/history", body, "application/json")
    print(json.dumps(result, ensure_ascii=False))
    return 0


def run_download(args, url, project, token):
    result = download(url, project, token, args.version, args.output, args.force)
    print(json.dumps(result, ensure_ascii=False))
    return 0


def run_command(args, root):
    if args.command == "configure":
        return run_configure(args, root)
    url, project, token = load_config(root)
    if args.command == "backup":
        return run_backup(args, root, url, project, token)
    if args.command == "history":
        return run_history(args, url, project, token)
    return run_download(args, url, project, token)


def main(argv=None):
    args = build_parser().parse_args(argv)
    root = Path(args.project_root).resolve()
    if not root.is_dir():
        raise RewindError("Project root must be an existing directory")
    return run_command(args, root)


if __name__ == "__main__":
    try:
        sys.exit(main())
    except (RewindError, OSError, ValueError, http.client.HTTPException) as exc:
        print(f"Rewind error: {exc}", file=sys.stderr)
        sys.exit(1)
