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


def load_config(root):
    try:
        data = json.loads(config_path(root).read_text(encoding="utf-8"))
        url, project, token = data["url"], data["project_id"], data["api_token"]
        if not all(isinstance(value, str) and value.strip() for value in (url, project, token)):
            raise ValueError()
        if any(c in token for c in "\r\n"):
            raise ValueError()
        validate_url(url)
    except (OSError, ValueError, KeyError, TypeError, AttributeError):
        raise RewindError("Project configuration is missing or invalid; run configure --project PROJECT_ID --url URL in a terminal")
    return url, project, token


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


def backup(url, project, token, path, key, description):
    if not path.is_file():
        raise RewindError("File does not exist or is not a regular file")
    boundary = "rewind-" + uuid.uuid4().hex
    def field(name, value):
        return (f'--{boundary}\r\nContent-Disposition: form-data; name="{name}"\r\n\r\n'.encode() + value.encode("utf-8") + b"\r\n")
    # Escape header parameters while retaining UTF-8 filenames.
    filename = path.name.replace("\\", "\\\\").replace('"', '\\"').replace("\r", "_").replace("\n", "_")
    prefix = (field("project_id", project) + field("key", key) + field("description", description) +
              f'--{boundary}\r\nContent-Disposition: form-data; name="file"; filename="{filename}"\r\nContent-Type: {mimetypes.guess_type(path.name)[0] or "application/octet-stream"}\r\n\r\n'.encode())
    suffix = f"\r\n--{boundary}--\r\n".encode()
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


def download(url, project, token, version, output, force):
    output = Path(output)
    if output.exists() and not force:
        raise RewindError("Destination exists; get user confirmation before using --force")
    conn, base = connection(url)
    temp = None
    try:
        endpoint = "/api/v1/backups/download/" + quote(project, safe="") + "/" + quote(version, safe="")
        conn.request("GET", base + endpoint, headers={"Authorization": "Bearer " + token})
        response = conn.getresponse()
        # Successful downloads always have attachment disposition, including JSON files.
        if not response.getheader("Content-Disposition", "").lower().startswith("attachment"):
            parse_response(response)
            raise RewindError("Server did not return a file")
        expected = response.getheader("Content-Length")
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
        conn.close()
        if temp is not None:
            temp.unlink(missing_ok=True)


def main():
    parser = argparse.ArgumentParser(description="Back up and restore files using Rewind")
    parser.add_argument("--project-root", default=".", help="Project directory containing .rewind/config.json (default: current directory)")
    sub = parser.add_subparsers(dest="command", required=True)
    setup = sub.add_parser("configure", help="Save project settings and securely prompt for API token")
    setup.add_argument("--url", required=True)
    setup.add_argument("--project", required=True)
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
    args = parser.parse_args()
    root = Path(args.project_root).resolve()
    if not root.is_dir():
        raise RewindError("Project root must be an existing directory")
    if args.command == "configure":
        validate_url(args.url)
        if not args.project.strip():
            raise RewindError("Project ID is required")
        if not sys.stdin.isatty():
            raise RewindError("Run configure in an interactive terminal to enter the API token privately")
        try:
            with warnings.catch_warnings():
                warnings.simplefilter("error", getpass.GetPassWarning)
                token = getpass.getpass("Rewind API token (input hidden): ")
        except getpass.GetPassWarning:
            raise RewindError("Cannot hide input; run configure in an interactive terminal")
        if not token.strip() or any(c in token for c in "\r\n"):
            raise RewindError("A non-empty API token is required")
        path = config_path(root)
        path.parent.mkdir(mode=0o700, parents=True, exist_ok=True)
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
                json.dump({"url": args.url.rstrip("/"), "project_id": args.project, "api_token": token}, f)
            os.replace(temp, path)
            temp = None
        finally:
            if temp is not None:
                temp.unlink(missing_ok=True)
        print(json.dumps({"configured": True, "path": str(path), "verified": False}))
        return 0
    url, project, token = load_config(root)
    if args.command == "backup":
        if args.key is not None and (not args.key or len(args.files) != 1):
            raise RewindError("--key requires a non-empty key and exactly one file")
        results = []
        for name in args.files:
            try:
                path = Path(name)
                key = args.key if args.key is not None else file_key(path, args.root or root)
                results.append({"file": name, "success": True, "data": backup(url, project, token, path, key, args.description)})
            except (RewindError, OSError, http.client.HTTPException, ValueError) as exc:
                results.append({"file": name, "success": False, "error": str(exc)})
        print(json.dumps(results, ensure_ascii=False))
        return 0 if all(x["success"] for x in results) else 1
    if args.command == "history":
        if not args.key or not 1 <= args.page or not 1 <= args.page_size <= 100:
            raise RewindError("Provide a non-empty key, positive page, and page size between 1 and 100")
        body = json.dumps({"project_id": project, "key": args.key, "page": args.page, "page_size": args.page_size}).encode()
        result = request(url, token, "POST", "/api/v1/backups/history", body, "application/json")
    else:
        result = download(url, project, token, args.version, args.output, args.force)
    print(json.dumps(result, ensure_ascii=False))
    return 0


if __name__ == "__main__":
    try:
        sys.exit(main())
    except (RewindError, OSError, ValueError, http.client.HTTPException) as exc:
        print(f"Rewind error: {exc}", file=sys.stderr)
        sys.exit(1)
