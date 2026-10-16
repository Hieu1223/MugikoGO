#!/usr/bin/env python3
"""
Seed script for MugikoGO database.
Seeds initial data (subtitle providers) assuming tables already exist.
"""

import os
import psycopg2
import uuid
from datetime import datetime, timezone


def get_db_path():
    """Get database path from environment or default."""
    return os.getenv("DB_PATH", "postgres://postgres:postgres@localhost:5432/mugiko?sslmode=disable")


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
    print(f"Seeding database at {db_path}")
    conn = psycopg2.connect(db_path)
    try:
        seed_initial_data(conn)
        print("Seeding completed successfully!")
    finally:
        conn.close()


if __name__ == "__main__":
    main()