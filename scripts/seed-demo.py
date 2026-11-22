#!/usr/bin/env python3
"""Add sample projects and randomized backup histories to a running Rewind server."""

import argparse
import json
import os
import secrets
import sys
import urllib.error
import urllib.request
import uuid


PROJECTS = (
    ("Rewind Demo - Atlas", "Sample project with randomized file backups."),
    ("Rewind Demo - Beacon", "Second sample project with randomized file backups."),
)
FILE_KEYS = (
    "docs/overview.txt",
    "config/settings.txt",
    "notes/changes.txt",
    "data/payload.txt",
)
VERSIONS_PER_FILE = 4
REQUEST_TIMEOUT = 30


def request_json(api_base, path, token=None, method="GET", payload=None):
    headers = {"Accept": "application/json"}
    body = None
    if payload is not None:
        headers["Content-Type"] = "application/json"
        body = json.dumps(payload).encode("utf-8")
    if token:
        headers["Authorization"] = f"Bearer {token}"
    request = urllib.request.Request(
        api_base + path, data=body, headers=headers, method=method
    )
    try:
        with urllib.request.urlopen(request, timeout=REQUEST_TIMEOUT) as response:
            result = json.loads(response.read())
    except urllib.error.HTTPError as error:
        details = error.read().decode("utf-8", errors="replace")
        raise RuntimeError(f"{method} {path} failed with HTTP {error.code}: {details}") from error
    except urllib.error.URLError as error:
        raise RuntimeError(f"{method} {path} failed: {error.reason}") from error
    if not isinstance(result, dict) or result.get("success") is not True:
        failure = result.get("error", {}) if isinstance(result, dict) else {}
        code = failure.get("code", "UNKNOWN")
        message = failure.get("message", "invalid API response")
        raise RuntimeError(f"{method} {path} failed ({code}): {message}")
    return result.get("data")


def list_all_projects(api_base, login_token):
    projects = []
    page = 1
    while True:
        batch = request_json(
            api_base,
            f"/projects?page={page}&page_size=100",
            token=login_token,
        )
        if not isinstance(batch, list):
            raise RuntimeError("GET /projects returned an invalid project list")
        projects.extend(batch)
        if len(batch) < 100:
            return projects
        page += 1


def multipart_body(fields, filename, content):
    boundary = "RewindDemo" + uuid.uuid4().hex
    body = bytearray()
    for name, value in fields:
        body.extend(f"--{boundary}\r\n".encode("ascii"))
        body.extend(
            f'Content-Disposition: form-data; name="{name}"\r\n\r\n'.encode("ascii")
        )
        body.extend(value.encode("utf-8"))
        body.extend(b"\r\n")
    body.extend(f"--{boundary}\r\n".encode("ascii"))
    body.extend(
        (
            f'Content-Disposition: form-data; name="file"; filename="{filename}"\r\n'
            "Content-Type: text/plain; charset=utf-8\r\n\r\n"
        ).encode("ascii")
    )
    body.extend(content)
    body.extend(b"\r\n")
    body.extend(f"--{boundary}--\r\n".encode("ascii"))
    return boundary, bytes(body)


def upload_version(api_base, api_token, project_id, key, version_number):
    content = (
        f"Randomized Rewind demo content\n"
        f"Project: {project_id}\n"
        f"File: {key}\n"
        f"Version: {version_number}\n"
        f"Payload: {secrets.token_hex(256)}\n"
    ).encode("utf-8")
    boundary, body = multipart_body(
        [
            ("project_id", project_id),
            ("key", key),
            ("description", f"Demo version {version_number}"),
        ],
        os.path.basename(key),
        content,
    )
    request = urllib.request.Request(
        api_base + "/backups",
        data=body,
        headers={
            "Accept": "application/json",
            "Authorization": f"Bearer {api_token}",
            "Content-Type": f"multipart/form-data; boundary={boundary}",
        },
        method="POST",
    )
    try:
        with urllib.request.urlopen(request, timeout=REQUEST_TIMEOUT) as response:
            result = json.loads(response.read())
    except urllib.error.HTTPError as error:
        details = error.read().decode("utf-8", errors="replace")
        raise RuntimeError(
            f"POST /backups failed with HTTP {error.code}: {details}"
        ) from error
    except urllib.error.URLError as error:
        raise RuntimeError(f"POST /backups failed: {error.reason}") from error
    if not isinstance(result, dict) or result.get("success") is not True:
        failure = result.get("error", {}) if isinstance(result, dict) else {}
        code = failure.get("code", "UNKNOWN")
        message = failure.get("message", "invalid API response")
        raise RuntimeError(f"POST /backups failed ({code}): {message}")


def delete_api_token(api_base, login_token, token_id):
    request_json(
        api_base,
        f"/tokens/{token_id}",
        token=login_token,
        method="DELETE",
    )


def seed(api_base, password):
    login = request_json(
        api_base,
        "/auth/login",
        method="POST",
        payload={"username": "admin", "password": password},
    )
    login_token = login.get("token") if isinstance(login, dict) else None
    if not login_token:
        raise RuntimeError("POST /auth/login returned no login token")

    existing_names = {
        project.get("name")
        for project in list_all_projects(api_base, login_token)
        if isinstance(project, dict)
    }
    duplicates = [name for name, _ in PROJECTS if name in existing_names]
    if duplicates:
        raise RuntimeError(
            "Demo project(s) already exist; no data was changed: "
            + ", ".join(duplicates)
            + ". Delete them in the admin UI before retrying."
        )

    created_projects = []
    api_token_id = None
    try:
        for name, description in PROJECTS:
            project = request_json(
                api_base,
                "/projects",
                token=login_token,
                method="POST",
                payload={"name": name, "description": description},
            )
            if not isinstance(project, dict) or not project.get("id"):
                raise RuntimeError("POST /projects returned no project ID")
            created_projects.append((project["id"], name))

        token = request_json(
            api_base,
            "/tokens",
            token=login_token,
            method="POST",
            payload={"name": "Temporary demo data importer"},
        )
        if not isinstance(token, dict) or not token.get("id") or not token.get("token"):
            raise RuntimeError("POST /tokens returned no API token")
        api_token_id = token["id"]

        for project_id, name in created_projects:
            for key in FILE_KEYS:
                for version in range(1, VERSIONS_PER_FILE + 1):
                    upload_version(api_base, token["token"], project_id, key, version)
            print(f"Created {name}: 4 files, {VERSIONS_PER_FILE} versions per file")
    finally:
        if api_token_id:
            delete_api_token(api_base, login_token, api_token_id)

    print("Demo import complete: 2 projects, 8 files, 32 randomized versions.")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--base-url",
        default=os.environ.get("REWIND_BASE_URL", "http://localhost:8080"),
        help="Rewind server origin (default: REWIND_BASE_URL or http://localhost:8080)",
    )
    args = parser.parse_args()
    password = os.environ.get("REWIND_ADMIN_PASSWORD")
    if not password:
        parser.error("set REWIND_ADMIN_PASSWORD in the environment")
    api_base = args.base_url.rstrip("/") + "/api/v1"
    try:
        seed(api_base, password)
    except (RuntimeError, urllib.error.URLError) as error:
        print(f"Demo import failed: {error}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
