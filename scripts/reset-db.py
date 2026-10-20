#!/usr/bin/env python3
"""Drop all MugikoGO tables in safe FK-dependency order.

Also drops the Sequin CDC artifacts (publication + replication slot) so that
`make setup` can recreate the whole database from scratch, matching
scripts/setup/schema.sql.
"""

import os
import sys

# Keep stdout line-buffered so log lines interleave correctly with stderr.
if hasattr(sys.stdout, "reconfigure"):
    sys.stdout.reconfigure(line_buffering=True)

try:
    import psycopg2
except ImportError:
    print("Error: pip install psycopg2-binary", file=sys.stderr)
    sys.exit(1)

# Leaf-most tables first; DROP ... CASCADE keeps this safe either way.
TABLES = [
    "deck_words",
    "decks",
    "subtitle_access",
    "subtitle_user_states",
    "subtitle_segments",
    "subtitles",
    "subtitle_provider_services",
    "subtitle_providers",
    "ocr_results",
    "paragraphs",
    "chapters",
    "novels",
    "manga_user_settings",
    "manga_history",
    "manga_subscriptions",
    "user_settings",
    "manga_providers",
    "user_profiles",
    "users",
    # Legacy table names from before the schema match with the persistent layer.
    "webnovel_novels",
    "webnovel_chapters",
    "webnovel_paragraphs",
]

db_url = os.environ.get("DB_PATH")
if not db_url:
    print("Error: DB_PATH environment variable is not set.", file=sys.stderr)
    sys.exit(1)

conn = psycopg2.connect(db_url)
conn.autocommit = True
cur = conn.cursor()

for table in TABLES:
    print(f"Dropping {table}...")
    cur.execute(f'DROP TABLE IF EXISTS "{table}" CASCADE')

print("Dropping publication sequin_pub...")
cur.execute("DROP PUBLICATION IF EXISTS sequin_pub")

cur.execute("SELECT 1 FROM pg_replication_slots WHERE slot_name = 'sequin_slot'")
if cur.fetchone():
    print("Dropping replication slot sequin_slot...")
    try:
        cur.execute("SELECT pg_drop_replication_slot('sequin_slot')")
    except psycopg2.Error as e:
        print(
            "Warning: could not drop replication slot 'sequin_slot' "
            f"(it may be active, e.g. Sequin is still connected): {e}",
            file=sys.stderr,
        )

cur.close()
conn.close()
print("Done. All tables dropped.")
