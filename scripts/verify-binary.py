#!/usr/bin/env python3
"""Exercise a built Readlane executable using disposable PostgreSQL databases.

Requires READLANE_TEST_DATABASE_URL pointing to a database ending in _test and a
role allowed to create databases and roles. No live database is reset.
"""
import http.server
import json
import os
import pathlib
import subprocess
import sys
import tempfile
import threading
import urllib.parse
import uuid


def execute(args, env=None, input_text=None, check=True, input_file=None, output_file=None):
    result = subprocess.run(
        args, env=env, input=input_text, stdin=input_file,
        stdout=output_file or subprocess.PIPE, stderr=subprocess.PIPE,
        text=True, timeout=45
    )
    if check and result.returncode:
        raise RuntimeError(args[0] + " failed: " + result.stderr)
    return result


def require(condition, message):
    if not condition:
        raise RuntimeError(message)


def client_environment(raw):
    parsed = urllib.parse.urlsplit(raw)
    query = urllib.parse.parse_qs(parsed.query)
    env = os.environ.copy()
    for key in ("PGHOST", "PGPORT", "PGDATABASE", "PGUSER", "PGPASSWORD", "PGSERVICE", "PGSERVICEFILE", "PGOPTIONS"):
        env.pop(key, None)
    env["PGDATABASE"] = urllib.parse.unquote(parsed.path.lstrip("/"))
    if parsed.username:
        env["PGUSER"] = urllib.parse.unquote(parsed.username)
    if parsed.password:
        env["PGPASSWORD"] = urllib.parse.unquote(parsed.password)
    if parsed.hostname:
        env["PGHOST"] = parsed.hostname
    if parsed.port:
        env["PGPORT"] = str(parsed.port)
    for key in ("host", "port", "sslmode", "sslrootcert", "sslcert", "sslkey", "options"):
        if key in query:
            env["PG" + key.upper()] = query[key][0]
    return env


def sql(query, env):
    return execute(["psql", "-X", "-v", "ON_ERROR_STOP=1", "-Atc", query], env).stdout.strip()


class Fixture(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.send_header("Content-Type", "application/rss+xml")
        self.end_headers()
        self.wfile.write(
            b"<rss><channel><item><title>Binary verification post</title>"
            b"<link>/post</link><description>&lt;p&gt;Stored and restored&lt;/p&gt;</description>"
            b"<pubDate>Wed, 14 Oct 2026 08:00:00 GMT</pubDate></item></channel></rss>"
        )

    def log_message(self, *args):
        pass


def main():
    original = os.environ.get("READLANE_TEST_DATABASE_URL")
    if not original:
        raise SystemExit("Set READLANE_TEST_DATABASE_URL to a disposable database ending in _test")
    binary = str(pathlib.Path(sys.argv[1] if len(sys.argv) > 1 else "readlane").resolve())
    base = urllib.parse.urlsplit(original)
    admin = client_environment(original)
    if not sql("SELECT current_database()", admin).endswith("_test"):
        raise SystemExit("Refusing verification against a database whose name does not end in _test")
    suffix = uuid.uuid4().hex[:10]
    deployment = "readlane_binary_" + suffix + "_test"
    restore = "readlane_restore_" + suffix + "_test"
    role = "readlane_worker_" + suffix
    created = []
    role_created = False
    server = http.server.ThreadingHTTPServer(("127.0.0.1", 0), Fixture)
    threading.Thread(target=server.serve_forever, daemon=True).start()

    def dsn(name):
        return urllib.parse.urlunsplit(base._replace(path="/" + name))

    counts = (
        "SELECT (SELECT count(*) FROM users),(SELECT count(*) FROM feeds),"
        "(SELECT count(*) FROM feed_follows),(SELECT count(*) FROM posts)"
    )
    try:
        with tempfile.TemporaryDirectory(prefix="readlane-binary-") as directory:
            for name in (deployment, restore):
                sql('CREATE DATABASE "' + name + '"', admin)
                created.append(name)
            config = pathlib.Path(directory) / "config.json"
            config.write_text(json.dumps({"db_url": dsn(deployment), "current_user_name": ""}))
            config.chmod(0o600)
            env = os.environ.copy()
            env["READLANE_CONFIG"] = str(config)
            env.pop("READLANE_DATABASE_URL", None)
            env.pop("GATOR_DATABASE_URL", None)
            env.pop("GATOR_CONFIG", None)
            feed_url = "http://127.0.0.1:" + str(server.server_port) + "/feed"
            commands = (
                ["migrate"], ["migrate"], ["doctor"], ["register", "verification"],
                ["addfeed", "Fixture", feed_url], ["collect", "5s", "--once"],
            )
            for args in commands:
                execute([binary, *args], env)
            browsed = execute([binary, "read", "2"], env).stdout
            require("Binary verification post" in browsed and "Stored and restored" in browsed, "browse did not show the collected fixture")
            database_env = client_environment(dsn(deployment))
            before = sql(counts, database_env)
            require(before == "1|1|1|1", "unexpected fixture counts: " + before)

            backup = pathlib.Path(directory) / "backup.dump"
            # Stream archives so matching clients may run in the CI service
            # container without needing access to the host's temporary paths.
            with backup.open("wb") as archive:
                execute(["pg_dump", "--format=custom"], database_env, output_file=archive)
            with backup.open("rb") as archive:
                restore_sql = execute(
                    ["pg_restore", "--no-owner", "--no-acl", "--file=-"],
                    input_file=archive,
                ).stdout
            execute(
                ["psql", "-X", "--set", "ON_ERROR_STOP=on", "--single-transaction"],
                client_environment(dsn(restore)), input_text=restore_sql,
            )
            require(before == sql(counts, client_environment(dsn(restore))), "restored counts differ")
            check_env = env.copy()
            check_env["READLANE_DATABASE_URL"] = dsn(restore)
            execute([binary, "doctor"], check_env)

            sql('CREATE ROLE "' + role + '" NOLOGIN', admin)
            role_created = True
            sql(
                'GRANT CONNECT ON DATABASE "' + deployment + '" TO "' + role + '"; '
                'GRANT USAGE ON SCHEMA public TO "' + role + '"; '
                'GRANT SELECT, UPDATE ON feeds TO "' + role + '"; '
                'GRANT INSERT, SELECT (url) ON posts TO "' + role + '";',
                database_env,
            )
            sql("UPDATE feeds SET next_fetch_at = NOW()", database_env)
            worker_url = urllib.parse.urlsplit(dsn(deployment))
            query = urllib.parse.parse_qs(worker_url.query)
            query["options"] = ["-c role=" + role]
            worker_env = env.copy()
            worker_env["READLANE_DATABASE_URL"] = urllib.parse.urlunsplit(
                worker_url._replace(query=urllib.parse.urlencode(query, doseq=True))
            )
            execute([binary, "agg", "5s", "--once"], worker_env)
            require(execute([binary, "reset"], env, check=False).returncode != 0, "unguarded reset succeeded")
            require(before == sql(counts, database_env), "unguarded reset changed data")
            execute([binary, "reset", "--confirm"], env)
            require(sql(counts, database_env) == "0|0|0|0", "confirmed reset left dependent data")
            print("PASS: fresh binary deployment, migrations, collection, browse, backup/restore, restricted worker role, guarded reset")
    finally:
        server.shutdown()
        server.server_close()
        for name in created:
            sql('DROP DATABASE IF EXISTS "' + name + '" WITH (FORCE)', admin)
        if role_created:
            sql('DROP ROLE IF EXISTS "' + role + '"', admin)


if __name__ == "__main__":
    main()
