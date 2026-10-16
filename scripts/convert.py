#!/usr/bin/env python3
"""
Converts SQLite queries to PostgreSQL in Go source files.

Handles:
  - ? placeholders → $1, $2, $3...
  - datetime('now') → NOW() AT TIME ZONE 'UTC'
  - strftime('%Y-%m-%dT%H:%M:%SZ', 'now') → to_char(NOW() AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
  - IFNULL() → COALESCE()
  - || string concat → kept as-is (valid in both)
  - AUTOINCREMENT → removed (PostgreSQL uses SERIAL/GENERATED)
"""

import os
import re
import sys


# ── SQLite → PostgreSQL function/expression mappings ──────────────────────────
# Order matters: more specific patterns must come before general ones.
SQLITE_TO_PG_PATTERNS = [
    # strftime with ISO8601 format and 'now'
    (
        re.compile(
            r"strftime\s*\(\s*'%Y-%m-%dT%H:%M:%SZ'\s*,\s*'now'\s*\)",
            re.IGNORECASE,
        ),
        "to_char(NOW() AT TIME ZONE 'UTC', 'YYYY-MM-DD\"T\"HH24:MI:SS\"Z\"')",
    ),
    # Generic strftime → to_char (best-effort; covers common formats)
    (
        re.compile(r"strftime\s*\(", re.IGNORECASE),
        "to_char(",
    ),
    # datetime('now') / datetime("now")
    (
        re.compile(r"datetime\s*\(\s*['\"]now['\"]\s*\)", re.IGNORECASE),
        "NOW() AT TIME ZONE 'UTC'",
    ),
    # date('now')
    (
        re.compile(r"date\s*\(\s*['\"]now['\"]\s*\)", re.IGNORECASE),
        "CURRENT_DATE",
    ),
    # IFNULL(a, b) → COALESCE(a, b)
    (
        re.compile(r"\bIFNULL\s*\(", re.IGNORECASE),
        "COALESCE(",
    ),
    # INTEGER PRIMARY KEY AUTOINCREMENT → BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY
    (
        re.compile(
            r"\bINTEGER\s+PRIMARY\s+KEY\s+AUTOINCREMENT\b", re.IGNORECASE
        ),
        "BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY",
    ),
    # Standalone AUTOINCREMENT (e.g., in non-PK columns)
    (
        re.compile(r"\bAUTOINCREMENT\b", re.IGNORECASE),
        "",
    ),
]


def convert_sqlite_to_pg(sql_string: str) -> str:
    """Apply all SQLite → PostgreSQL transformations to a SQL string."""
    result = sql_string
    for pattern, replacement in SQLITE_TO_PG_PATTERNS:
        result = pattern.sub(replacement, result)
    return result


def convert_placeholders(sql_string: str) -> str:
    """Replace ? with incrementing $1, $2, $3... in a SQL string."""
    counter = 0

    def replacer(match):
        nonlocal counter
        counter += 1
        return f"${counter}"

    return re.sub(r"\?", replacer, sql_string)


def process_go_file(filepath: str) -> bool:
    """
    Process a single Go file. Returns True if the file was modified.

    Finds backtick and double-quoted strings that look like SQL,
    then applies SQLite→PG conversions and ? → $N placeholder fixes.
    """
    with open(filepath, "r", encoding="utf-8") as f:
        content = f.read()

    original = content

    sql_keywords = re.compile(
        r"\b(SELECT|INSERT|UPDATE|DELETE|CREATE|ALTER|DROP|FROM|WHERE|INTO|VALUES|SET|JOIN)\b",
        re.IGNORECASE,
    )

    # Match backtick-quoted (raw) and double-quoted (interpreted) Go strings
    string_pattern = re.compile(r'(`[^`]*`|"(?:[^"\\]|\\.)*")')

    def replace_in_string(match):
        s = match.group(0)
        if not sql_keywords.search(s):
            return s

        quote_char = s[0]
        inner = s[1:-1]

        # Apply SQLite → PG function/expression conversions
        converted = convert_sqlite_to_pg(inner)

        # Then convert ? placeholders to $N
        converted = convert_placeholders(converted)

        if converted == inner:
            return s

        return f"{quote_char}{converted}{quote_char}"

    content = string_pattern.sub(replace_in_string, content)

    if content != original:
        with open(filepath, "w", encoding="utf-8") as f:
            f.write(content)
        return True

    return False


def scan_directory(root_dir: str) -> None:
    """Recursively scan for .go files and convert SQLite queries to PostgreSQL."""
    modified_count = 0
    scanned_count = 0

    for dirpath, _, filenames in os.walk(root_dir):
        # Skip vendor and hidden directories
        parts = dirpath.replace("\\", "/").split("/")
        if "vendor" in parts or ".git" in parts:
            continue

        for filename in filenames:
            if not filename.endswith(".go"):
                continue

            filepath = os.path.join(dirpath, filename)
            scanned_count += 1

            try:
                if process_go_file(filepath):
                    modified_count += 1
                    print(f"  ✏️  Modified: {filepath}")
                else:
                    print(f"  ✅ No change: {filepath}")
            except Exception as e:
                print(f"  ❌ Error processing {filepath}: {e}", file=sys.stderr)

    print(f"\nDone. Scanned {scanned_count} files, modified {modified_count}.")


if __name__ == "__main__":
    target = sys.argv[1] if len(sys.argv) > 1 else "."

    if not os.path.isdir(target):
        print(f"Error: '{target}' is not a valid directory.", file=sys.stderr)
        sys.exit(1)

    print(f"Scanning Go files in: {os.path.abspath(target)}\n")
    scan_directory(target)