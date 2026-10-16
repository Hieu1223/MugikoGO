#!/usr/bin/env python3
"""
Setup script for MugikoGO database.
Creates all tables and seeds initial data.
"""

import os
import psycopg2
import uuid
from datetime import datetime, timezone


def get_db_path():
    """Get database path from environment or default."""
    return os.getenv("DB_PATH", "postgres://postgres:postgres@localhost:5432/mugiko?sslmode=disable")


def split_sql_statements(sql):
    """Split SQL string into individual statements, respecting quoted strings and dollar quotes."""
    statements = []
    current = []
    i = 0
    n = len(sql)
    in_single_quote = False
    in_dollar_quote = False  # only for $$ quotes

    while i < n:
        c = sql[i]
        if not in_single_quote and not in_dollar_quote:
            if c == "'":
                in_single_quote = True
                current.append(c)
            elif c == '$':
                # Check if next char is also $ to open/close $$ quote
                if i + 1 < n and sql[i + 1] == '$':
                    current.append(c)
                    current.append(sql[i + 1])
                    in_dollar_quote = not in_dollar_quote  # toggle
                    i += 1  # skip the next $
                else:
                    current.append(c)
            elif c == ';':
                statements.append(''.join(current))
                current = []
            else:
                current.append(c)
        elif in_single_quote:
            if c == "'":
                # Check for escaped single quote: two single quotes in a row
                if i + 1 < n and sql[i + 1] == "'":
                    current.append(c)
                    current.append(sql[i + 1])
                    i += 1  # skip the next quote
                else:
                    in_single_quote = False
                current.append(c)
            else:
                current.append(c)
        else:  # in_dollar_quote
            if c == '$':
                # Check if next char is also $ to close $$ quote
                if i + 1 < n and sql[i + 1] == '$':
                    current.append(c)
                    current.append(sql[i + 1])
                    in_dollar_quote = False
                    i += 1  # skip the next $
                else:
                    current.append(c)
            else:
                current.append(c)
        i += 1

    if current:
        statements.append(''.join(current))
    return [s.strip() for s in statements if s.strip()]


def execute_sql(conn, statements):
    """Execute a list of SQL statements."""
    cursor = conn.cursor()
    for statement in statements:
        if statement.strip():
            try:
                cursor.execute(statement)
            except Exception as e:
                print(f"Error executing SQL: {e}")
                print(f"Statement: {statement[:100]}...")
                raise
    conn.commit()


def load_schema_from_directory(base_dir):
    """Load all schema.sql files from module directories."""
    schema_files = []
    for root, dirs, files in os.walk(base_dir):
        if 'schema.sql' in files:
            schema_files.append(os.path.join(root, 'schema.sql'))
    
    all_statements = []
    for schema_file in schema_files:
        print(f"Loading schema from {schema_file}")
        with open(schema_file, 'r', encoding='utf-8') as f:
            content = f.read()
        # Remove comment lines (starting with --) and keep line endings for SQL
        lines = content.splitlines()
        sql_lines = []
        for line in lines:
            stripped = line.lstrip()
            if stripped.startswith('--'):
                continue
            sql_lines.append(line)
        sql_content = '\n'.join(sql_lines)
        statements = split_sql_statements(sql_content)
        all_statements.extend(statements)
    
    return all_statements


def create_tables(conn):
    """Create all tables by reading schema files from module directories."""
    base_dir = os.path.join(os.path.dirname(__file__), '..', '..', 'internal', 'modules')
    statements = load_schema_from_directory(base_dir)
    execute_sql(conn, statements)


def seed_initial_data(conn):
    """Seed initial data (subtitle providers)."""
    cursor = conn.cursor()

    # Check if providers already exist
    cursor.execute("SELECT COUNT(*) FROM subtitle_providers")
    count = cursor.fetchone()[0]
    if count > 0:
        print("Subtitle providers already exist, skipping seeding.")
        return

    now = datetime.now(timezone.utc).isoformat()

    # Default providers
    providers = [
        {
            "name": "OpenSubtitles",
            "kind": "remote",
            "base_url": "https://api.opensubtitles.com",
            "services": [
                {"service_key": "search", "name": "Search Subtitles"},
                {"service_key": "download", "name": "Download Subtitles"}
            ]
        },
        {
            "name": "Subscene",
            "kind": "remote",
            "base_url": "https://subscene.com",
            "services": [
                {"service_key": "search", "name": "Search Subtitles"},
                {"service_key": "download", "name": "Download Subtitles"}
            ]
        },
        {
            "name": "YIFY Subtitles",
            "kind": "remote",
            "base_url": "https://yifysubtitles.ch",
            "services": [
                {"service_key": "search", "name": "Search Subtitles"}
            ]
        }
    ]

    for provider in providers:
        # Insert provider
        provider_id = str(uuid.uuid4())
        cursor.execute(
            """
            INSERT INTO subtitle_providers (id, name, kind, base_url, is_active, created_at, updated_at)
            VALUES (%s, %s, %s, %s, %s, %s, %s)
            """,
            (provider_id, provider["name"], provider["kind"], provider["base_url"], True, now, now)
        )

        # Insert services
        for service in provider["services"]:
            service_id = str(uuid.uuid4())
            cursor.execute(
                """
                INSERT INTO subtitle_provider_services (id, provider_id, service_key, name, options, is_active, created_at, updated_at)
                VALUES (%s, %s, %s, %s, %s, %s, %s, %s)
                """,
                (service_id, provider_id, service["service_key"], service["name"], None, True, now, now)
            )

    conn.commit()
    print(f"Seeded {len(providers)} subtitle providers.")


def main():
    db_path = get_db_path()
    print(f"Setting up database at {db_path}")
    conn = psycopg2.connect(db_path)
    try:
        create_tables(conn)
        seed_initial_data(conn)
        print("Database setup completed successfully!")
    finally:
        conn.close()


if __name__ == "__main__":
    main()