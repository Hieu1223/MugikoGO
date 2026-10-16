#!/usr/bin/env python3
"""
Setup script for MugikoGO database.
Creates all tables and seeds initial data.
"""

import os
import sqlite3
import uuid
from datetime import datetime

def get_db_path():
    """Get database path from environment or default."""
    return os.getenv("DB_PATH", "./mugiko.db")

def execute_sql(conn, statements):
    """Execute a list of SQL statements."""
    cursor = conn.cursor()
    for statement in statements:
        if statement.strip():
            try:
                cursor.execute(statement)
            except sqlite3.Error as e:
                print(f"Error executing SQL: {e}")
                print(f"Statement: {statement[:100]}...")
                raise
    conn.commit()

def create_tables(conn):
    """Create all tables."""
    statements = [
        # Core tables from persistent/db.go migrations
        """
        CREATE TABLE IF NOT EXISTS user_auths (
            user_id TEXT PRIMARY KEY,
            role TEXT NOT NULL DEFAULT 'anonymous',
            joined_at DATETIME NOT NULL DEFAULT (datetime('now')),
            auth_type TEXT NOT NULL,
            auth_id TEXT NOT NULL,
            is_active BOOLEAN NOT NULL DEFAULT 1,
            UNIQUE(auth_type, auth_id)
        );
        """,
        """
        CREATE INDEX IF NOT EXISTS ix_user_auths_auth_type_auth_id ON user_auths(auth_type, auth_id);
        """,
        """
        CREATE TABLE IF NOT EXISTS user_profiles (
            user_id TEXT PRIMARY KEY,
            name TEXT NOT NULL,
            bio TEXT,
            email TEXT,
            FOREIGN KEY (user_id) REFERENCES user_auths(user_id) ON DELETE CASCADE
        );
        """,
        # Manga module tables
        """
        CREATE TABLE IF NOT EXISTS manga_providers (
            id TEXT PRIMARY KEY,
            owner_id TEXT NOT NULL,
            name TEXT NOT NULL,
            base_url TEXT NOT NULL,
            visibility TEXT NOT NULL DEFAULT 'private',
            is_builtin INTEGER NOT NULL DEFAULT 0,
            created_at TEXT NOT NULL DEFAULT (datetime('now')),
            updated_at TEXT NOT NULL DEFAULT (datetime('now')),
            FOREIGN KEY (owner_id) REFERENCES user_auths(user_id) ON DELETE CASCADE
        );
        """,
        """
        CREATE UNIQUE INDEX IF NOT EXISTS ix_manga_providers_owner_base_url ON manga_providers(owner_id, base_url);
        """,
        """
        CREATE INDEX IF NOT EXISTS ix_manga_providers_visibility ON manga_providers(visibility);
        """,
        """
        CREATE INDEX IF NOT EXISTS ix_manga_providers_owner_id ON manga_providers(owner_id);
        """,
        """
        CREATE TABLE IF NOT EXISTS manga_subscriptions (
            user_id TEXT NOT NULL,
            provider_id TEXT NOT NULL,
            created_at TEXT NOT NULL DEFAULT (datetime('now')),
            PRIMARY KEY (user_id, provider_id),
            FOREIGN KEY (user_id) REFERENCES user_auths(user_id) ON DELETE CASCADE,
            FOREIGN KEY (provider_id) REFERENCES manga_providers(id) ON DELETE CASCADE
        );
        """,
        """
        CREATE TABLE IF NOT EXISTS manga_history (
            user_id TEXT NOT NULL,
            provider_id TEXT NOT NULL,
            manga_id TEXT NOT NULL,
            chapter_id TEXT NOT NULL,
            page_index INTEGER NOT NULL DEFAULT 0,
            title TEXT NOT NULL,
            cover_url TEXT,
            last_read_at TEXT NOT NULL DEFAULT (datetime('now')),
            PRIMARY KEY (user_id, provider_id, manga_id),
            FOREIGN KEY (user_id) REFERENCES user_auths(user_id) ON DELETE CASCADE,
            FOREIGN KEY (provider_id) REFERENCES manga_providers(id) ON DELETE CASCADE
        );
        """,
        """
        CREATE INDEX IF NOT EXISTS ix_manga_history_user_last_read ON manga_history(user_id, last_read_at);
        """,
        """
        CREATE TABLE IF NOT EXISTS manga_user_settings (
            user_id TEXT PRIMARY KEY,
            schema_version INTEGER NOT NULL DEFAULT 1,
            data TEXT NOT NULL DEFAULT '{}',
            updated_at TEXT NOT NULL DEFAULT (datetime('now')),
            FOREIGN KEY (user_id) REFERENCES user_auths(user_id) ON DELETE CASCADE
        );
        """,
        # Deck module tables
        """
        CREATE TABLE IF NOT EXISTS decks (
            id TEXT PRIMARY KEY,
            owner_id TEXT NOT NULL,
            name TEXT NOT NULL,
            created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (owner_id) REFERENCES user_auths(user_id)
        );
        """,
        """
        CREATE TABLE IF NOT EXISTS deck_words (
            id TEXT PRIMARY KEY,
            deck_id TEXT NOT NULL,
            word TEXT NOT NULL,
            meaning TEXT NOT NULL,
            created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (deck_id) REFERENCES decks(id) ON DELETE CASCADE
        );
        """,
        """
        CREATE INDEX IF NOT EXISTS idx_decks_owner_id ON decks(owner_id);
        """,
        """
        CREATE INDEX IF NOT EXISTS idx_deck_words_deck_id ON deck_words(deck_id);
        """,
        # Subtitle module tables
        """
        CREATE TABLE IF NOT EXISTS subtitle_providers (
            id TEXT PRIMARY KEY,
            name TEXT NOT NULL,
            kind TEXT NOT NULL,
            base_url TEXT,
            is_active BOOLEAN DEFAULT TRUE NOT NULL,
            last_synced_at TIMESTAMP,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
        );
        """,
        """
        CREATE TABLE IF NOT EXISTS subtitle_provider_services (
            id TEXT PRIMARY KEY,
            provider_id TEXT NOT NULL REFERENCES subtitle_providers(id) ON DELETE CASCADE,
            service_key TEXT NOT NULL,
            name TEXT NOT NULL,
            options TEXT,
            is_active BOOLEAN DEFAULT TRUE NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
            UNIQUE(provider_id, service_key)
        );
        """,
        """
        CREATE TABLE IF NOT EXISTS subtitles (
            id TEXT PRIMARY KEY,
            video_id TEXT NOT NULL,
            video_url TEXT NOT NULL,
            provider_id TEXT NOT NULL REFERENCES subtitle_providers(id),
            service_id TEXT NOT NULL REFERENCES subtitle_provider_services(id),
            created_by TEXT NOT NULL,
            status TEXT NOT NULL DEFAULT 'queued' CHECK (status IN ('queued', 'error', 'completed')),
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
        );
        """,
        """
        CREATE INDEX IF NOT EXISTS ix_subtitles_video_id ON subtitles(video_id);
        """,
        """
        CREATE TABLE IF NOT EXISTS subtitle_segments (
            id TEXT PRIMARY KEY,
            subtitle_id TEXT NOT NULL REFERENCES subtitles(id) ON DELETE CASCADE,
            idx INTEGER NOT NULL,
            start_ms INTEGER NOT NULL,
            end_ms INTEGER NOT NULL,
            words TEXT,
            UNIQUE(subtitle_id, idx)
        );
        """,
        """
        CREATE TABLE IF NOT EXISTS subtitle_access (
            user_id TEXT NOT NULL,
            subtitle_id TEXT NOT NULL REFERENCES subtitles(id) ON DELETE CASCADE,
            last_accessed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
            PRIMARY KEY (user_id, subtitle_id)
        );
        """,
        """
        CREATE TABLE IF NOT EXISTS user_settings (
            user_id TEXT PRIMARY KEY,
            schema_version INTEGER DEFAULT 1 NOT NULL,
            data TEXT NOT NULL DEFAULT '{}',
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
        );
        """,
        """
        CREATE TABLE IF NOT EXISTS subtitle_user_state (
            user_id TEXT NOT NULL,
            subtitle_id TEXT NOT NULL REFERENCES subtitles(id) ON DELETE CASCADE,
            schema_version INTEGER DEFAULT 1 NOT NULL,
            data TEXT NOT NULL DEFAULT '{}',
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
            PRIMARY KEY (user_id, subtitle_id)
        );
        """,
        # Triggers for subtitle module
        """
        CREATE TRIGGER IF NOT EXISTS trg_subtitle_providers_updated_at
        AFTER UPDATE ON subtitle_providers
        BEGIN
            UPDATE subtitle_providers SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
        END;
        """,
        """
        CREATE TRIGGER IF NOT EXISTS trg_subtitle_provider_services_updated_at
        AFTER UPDATE ON subtitle_provider_services
        BEGIN
            UPDATE subtitle_provider_services SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
        END;
        """,
        """
        CREATE TRIGGER IF NOT EXISTS trg_user_settings_updated_at
        AFTER UPDATE ON user_settings
        BEGIN
            UPDATE user_settings SET updated_at = CURRENT_TIMESTAMP WHERE user_id = NEW.user_id;
        END;
        """,
        """
        CREATE TRIGGER IF NOT EXISTS trg_subtitle_user_state_updated_at
        AFTER UPDATE ON subtitle_user_state
        BEGIN
            UPDATE subtitle_user_state SET updated_at = CURRENT_TIMESTAMP WHERE user_id = NEW.user_id AND subtitle_id = NEW.subtitle_id;
        END;
        """,
        # Migration: add status column if missing (for existing tables without it)
        # SQLite doesn't support IF NOT EXISTS for ADD COLUMN, so we use a try-except pattern
        # We'll try to add the column and ignore duplicate column error
    ]
    
    # Execute core statements
    execute_sql(conn, statements)
    
    # Try to add status column to subtitles table (ignore if already exists)
    cursor = conn.cursor()
    try:
        cursor.execute("ALTER TABLE subtitles ADD COLUMN status TEXT NOT NULL DEFAULT 'queued' CHECK (status IN ('queued', 'error', 'completed'));")
        conn.commit()
    except sqlite3.OperationalError as e:
        if "duplicate column name" in str(e):
            # Column already exists, ignore
            pass
        else:
            raise

def seed_initial_data(conn):
    """Seed initial data (subtitle providers)."""
    cursor = conn.cursor()
    
    # Check if providers already exist
    cursor.execute("SELECT COUNT(*) FROM subtitle_providers")
    count = cursor.fetchone()[0]
    if count > 0:
        print("Subtitle providers already exist, skipping seeding.")
        return
    
    now = datetime.utcnow().isoformat()
    
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
            VALUES (?, ?, ?, ?, ?, ?, ?)
            """,
            (provider_id, provider["name"], provider["kind"], provider["base_url"], 1, now, now)
        )
        
        # Insert services
        for service in provider["services"]:
            service_id = str(uuid.uuid4())
            cursor.execute(
                """
                INSERT INTO subtitle_provider_services (id, provider_id, service_key, name, options, is_active, created_at, updated_at)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?)
                """,
                (service_id, provider_id, service["service_key"], service["name"], None, 1, now, now)
            )
    
    conn.commit()
    print(f"Seeded {len(providers)} subtitle providers.")

def main():
    db_path = get_db_path()
    print(f"Setting up database at {db_path}")
    
    # Ensure directory exists
    os.makedirs(os.path.dirname(os.path.abspath(db_path)), exist_ok=True)
    
    conn = sqlite3.connect(db_path)
    # Enable foreign keys
    conn.execute("PRAGMA foreign_keys = ON")
    
    try:
        create_tables(conn)
        seed_initial_data(conn)
        print("Database setup completed successfully!")
    finally:
        conn.close()

if __name__ == "__main__":
    main()