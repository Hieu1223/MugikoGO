#!/usr/bin/env python3
"""Drop all MugikoGO tables in safe FK-dependency order."""

import os
import sys

try:
    import psycopg2
except ImportError:
    print("Error: pip install psycopg2-binary", file=sys.stderr)
    sys.exit(1)

TABLES = [
    "subtitle_user_states",
    "subtitle_segments",
    "subtitles",
    "subtitle_provider_services",
    "subtitle_providers",
    "webnovel_paragraphs",
    "webnovel_chapters",
    "webnovel_novels",
    "manga_user_settings",
    "manga_history",
    "manga_subscriptions",
    "manga_providers",
    "user_profiles",
    "user_auths",
    "users",
]

db_url = os.environ.get("DB_PATH")
if not db_url:
    print("Error: DATABASE_URL environment variable is not set.", file=sys.stderr)
    sys.exit(1)

conn = psycopg2.connect(db_url)
conn.autocommit = True
cur = conn.cursor()

for table in TABLES:
    print(f"Dropping {table}...")
    cur.execute(f'DROP TABLE IF EXISTS "{table}" CASCADE')

cur.close()
conn.close()
print("Done. All tables dropped.")