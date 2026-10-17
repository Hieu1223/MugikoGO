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
    "users",
    "user_profiles",
    "manga_providers",
    "manga_subscriptions",
    "manga_history",
    "manga_user_settings",
    "user_settings",
    "webnovel_novels",
    "webnovel_chapters",
    "webnovel_paragraphs",
    "subtitle_providers",
    "subtitle_provider_services",
    "subtitles",
    "subtitle_segments",
    "subtitle_user_states",
    "decks",
    "deck_words",
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