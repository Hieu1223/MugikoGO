# Database Design Document

## Overview
This document describes the database schema for the MugikoGO application, implemented using SQLite for local development. The schema follows a modular approach with separate tables for each feature module.

## Table of Contents
1. [Core Tables](#core-tables)
2. [Manga Module](#manga-module)
3. [Web Novel Module](#web-novel-module)
4. [Subtitle Module](#subtitle-module)
5. [Relationships](#relationships)
6. [Indexes](#indexes)

---

## Core Tables

### `users`
Stores user accounts and authentication information.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | TEXT (UUID) | PRIMARY KEY | Unique user identifier |
| `name` | TEXT | NOT NULL | Display name |
| `bio` | TEXT | NULLABLE | User biography |
| `email` | TEXT | UNIQUE, NULLABLE | Email address |
| `role` | TEXT | NOT NULL DEFAULT 'user' | Role: anonymous/user/admin |
| `firebase_uid` | TEXT | UNIQUE, NULLABLE | Firebase UID for auth |
| `created_at` | TEXT (ISO8601) | NOT NULL | Account creation timestamp |
| `updated_at` | TEXT (ISO8601) | NOT NULL | Last update timestamp |

---

## Manga Module

### `manga_providers`
Stores manga provider configurations (user-created sources).

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | TEXT (UUID) | PRIMARY KEY | Unique provider identifier |
| `owner_id` | TEXT (UUID) | NOT NULL, FK → users.id | Provider owner |
| `name` | TEXT | NOT NULL | Provider display name |
| `base_url` | TEXT | NOT NULL | Base URL for the provider API |
| `visibility` | TEXT | NOT NULL DEFAULT 'private' | 'public' or 'private' |
| `is_builtin` | INTEGER (BOOLEAN) | NOT NULL DEFAULT 0 | Whether built-in provider |
| `created_at` | TEXT (ISO8601) | NOT NULL | Creation timestamp |
| `updated_at` | TEXT (ISO8601) | NOT NULL | Last update timestamp |

**Indexes:**
- `idx_manga_providers_owner_id` ON `owner_id`
- `idx_manga_providers_visibility` ON `visibility`

### `manga_subscriptions`
Tracks user subscriptions to manga providers.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `user_id` | TEXT (UUID) | NOT NULL, FK → users.id | Subscribing user |
| `provider_id` | TEXT (UUID) | NOT NULL, FK → manga_providers.id | Subscribed provider |
| `created_at` | TEXT (ISO8601) | NOT NULL | Subscription timestamp |
| **PRIMARY KEY** | | | (`user_id`, `provider_id`) |

**Indexes:**
- `idx_manga_subscriptions_user_id` ON `user_id`
- `idx_manga_subscriptions_provider_id` ON `provider_id`

### `manga_history`
Stores user reading history across manga providers.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | TEXT (UUID) | PRIMARY KEY | Unique history entry |
| `user_id` | TEXT (UUID) | NOT NULL, FK → users.id | Reading user |
| `provider_id` | TEXT (UUID) | NOT NULL, FK → manga_providers.id | Source provider |
| `manga_id` | TEXT | NOT NULL | External manga identifier |
| `chapter_id` | TEXT | NOT NULL | External chapter identifier |
| `page_index` | INTEGER | NOT NULL DEFAULT 0 | Current page index |
| `title` | TEXT | NOT NULL | Manga title |
| `cover_url` | TEXT | NULLABLE | Cover image URL |
| `last_read_at` | TEXT (ISO8601) | NOT NULL | Last read timestamp |

**Indexes:**
- `idx_manga_history_user_id` ON `user_id`
- `idx_manga_history_provider_id` ON `provider_id`
- `idx_manga_history_last_read_at` ON `last_read_at` DESC
- `idx_manga_history_user_provider` ON (`user_id`, `provider_id`)
- `idx_manga_history_user_manga` ON (`user_id`, `manga_id`)

### `manga_user_settings`
Stores per-user manga module preferences.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `user_id` | TEXT (UUID) | PRIMARY KEY, FK → users.id | User identifier |
| `schema_version` | INTEGER | NOT NULL DEFAULT 1 | Settings schema version |
| `data` | TEXT (JSON) | NOT NULL DEFAULT '{}' | Flexible settings JSON |
| `updated_at` | TEXT (ISO8601) | NOT NULL | Last update timestamp |

---

## Web Novel Module

### `webnovel_novels`
Stores web novel metadata.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | TEXT (UUID) | PRIMARY KEY | Unique novel identifier |
| `source` | TEXT | NOT NULL | Source enum: 'syosetsu' |
| `source_id` | TEXT | NOT NULL | External source ID |
| `url` | TEXT | NOT NULL | Original novel URL |
| `title` | TEXT | NOT NULL | Novel title |
| `author` | TEXT | NULLABLE | Author name |
| `synopsis` | TEXT | NULLABLE | Novel synopsis |
| `added_by` | TEXT (UUID) | NOT NULL, FK → users.id | User who added novel |
| `created_at` | TEXT (ISO8601) | NOT NULL | Creation timestamp |
| **UNIQUE** | | | (`source`, `source_id`) |

**Indexes:**
- `idx_webnovel_novels_added_by` ON `added_by`
- `idx_webnovel_novels_source_source_id` ON (`source`, `source_id`)

### `webnovel_chapters`
Stores chapter metadata for web novels.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | TEXT (UUID) | PRIMARY KEY | Unique chapter identifier |
| `novel_id` | TEXT (UUID) | NOT NULL, FK → webnovel_novels.id | Parent novel |
| `idx` | INTEGER | NOT NULL | Chapter index/order |
| `title` | TEXT | NULLABLE | Chapter title |
| `created_at` | TEXT (ISO8601) | NOT NULL | Creation timestamp |

**Indexes:**
- `idx_webnovel_chapters_novel_id` ON `novel_id`
- `idx_webnovel_chapters_novel_idx` ON (`novel_id`, `idx`)

### `webnovel_paragraphs`
Stores paragraph content for chapters.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | TEXT (UUID) | PRIMARY KEY | Unique paragraph identifier |
| `chapter_id` | TEXT (UUID) | NOT NULL, FK → webnovel_chapters.id | Parent chapter |
| `idx` | INTEGER | NOT NULL | Paragraph index/order |
| `text` | TEXT | NOT NULL | Paragraph text content |
| `words` | TEXT (JSON) | NOT NULL DEFAULT '[]' | Tokenized words array |

**Indexes:**
- `idx_webnovel_paragraphs_chapter_id` ON `chapter_id`
- `idx_webnovel_paragraphs_chapter_idx` ON (`chapter_id`, `idx`)

---

## Subtitle Module

### `subtitle_providers`
Stores subtitle provider configurations.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | TEXT (UUID) | PRIMARY KEY | Unique provider identifier |
| `name` | TEXT | NOT NULL | Provider name |
| `kind` | TEXT | NOT NULL | 'remote' or 'youtube' |
| `base_url` | TEXT | NULLABLE | Base URL |
| `is_active` | INTEGER (BOOLEAN) | NOT NULL DEFAULT 1 | Active status |
| `last_synced_at` | TEXT (ISO8601) | NULLABLE | Last sync timestamp |
| `created_at` | TEXT (ISO8601) | NOT NULL | Creation timestamp |
| `updated_at` | TEXT (ISO8601) | NOT NULL | Last update timestamp |

### `subtitle_provider_services`
Stores services for subtitle providers.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | TEXT (UUID) | PRIMARY KEY | Unique service identifier |
| `provider_id` | TEXT (UUID) | NOT NULL, FK → subtitle_providers.id | Parent provider |
| `service_key` | TEXT | NOT NULL | Service identifier |
| `name` | TEXT | NOT NULL | Service display name |
| `options` | TEXT (JSON) | NULLABLE | Service options JSON |
| `is_active` | INTEGER (BOOLEAN) | NOT NULL DEFAULT 1 | Active status |
| `created_at` | TEXT (ISO8601) | NOT NULL | Creation timestamp |
| `updated_at` | TEXT (ISO8601) | NOT NULL | Last update timestamp |

### `subtitles`
Stores subtitle records.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | TEXT (UUID) | PRIMARY KEY | Unique subtitle identifier |
| `video_id` | TEXT | NOT NULL | Video identifier |
| `video_url` | TEXT | NOT NULL | Video URL |
| `provider_id` | TEXT (UUID) | NOT NULL, FK → subtitle_providers.id | Source provider |
| `service_id` | TEXT (UUID) | NOT NULL, FK → subtitle_provider_services.id | Source service |
| `created_by` | TEXT (UUID) | NOT NULL, FK → users.id | Creator user |
| `status` | TEXT | NOT NULL DEFAULT 'queued' | Status: 'queued', 'error', 'completed' |
| `created_at` | TEXT (ISO8601) | NOT NULL | Creation timestamp |
| `updated_at` | TEXT (ISO8601) | NOT NULL | Last update timestamp |

**Indexes:**
- `idx_subtitles_provider_id` ON `provider_id`
- `idx_subtitles_service_id` ON `service_id`
- `idx_subtitles_created_by` ON `created_by`
- `idx_subtitles_video_id` ON `video_id`
- `idx_subtitles_status` ON `status`

### `subtitle_segments`
Stores individual subtitle segments.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | TEXT (UUID) | PRIMARY KEY | Unique segment identifier |
| `subtitle_id` | TEXT (UUID) | NOT NULL, FK → subtitles.id | Parent subtitle |
| `idx` | INTEGER | NOT NULL | Segment order |
| `start_ms` | INTEGER | NOT NULL | Start time in milliseconds |
| `end_ms` | INTEGER | NOT NULL | End time in milliseconds |
| `words` | TEXT (JSON) | NULLABLE | Word timing data |
| `created_at` | TEXT (ISO8601) | NOT NULL | Creation timestamp |

**Indexes:**
- `idx_subtitle_segments_subtitle_id` ON `subtitle_id`
- `idx_subtitle_segments_subtitle_idx` ON (`subtitle_id`, `idx`)

### `subtitle_user_states`
Tracks per-user subtitle playback state.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `user_id` | TEXT (UUID) | NOT NULL, FK → users.id | User identifier |
| `subtitle_id` | TEXT (UUID) | NOT NULL, FK → subtitles.id | Subtitle identifier |
| `schema_version` | INTEGER | NOT NULL DEFAULT 1 | State schema version |
| `data` | TEXT (JSON) | NOT NULL DEFAULT '{}' | Playback state JSON |
| `updated_at` | TEXT (ISO8601) | NOT NULL | Last update timestamp |
| **PRIMARY KEY** | | | (`user_id`, `subtitle_id`) |

**Indexes:**
- `idx_subtitle_user_states_user_id` ON `user_id`
- `idx_subtitle_user_states_subtitle_id` ON `subtitle_id`

---

## Relationships

```
users
  │
  ├── manga_providers (owner_id) ──→ manga_subscriptions (provider_id)
  │       │                               │
  │       └───────────────────────────────┘
  │
  ├── manga_history (user_id, provider_id)
  │
  ├── manga_user_settings (user_id)
  │
  ├── webnovel_novels (added_by)
  │
  ├── subtitles (created_by)
  │
  └── subtitle_user_states (user_id)
```

### Manga Module Relationships
- **User → MangaProvider**: One-to-Many (owner)
- **User → MangaSubscription**: Many-to-Many (via manga_subscriptions)
- **User → MangaHistory**: One-to-Many
- **User → MangaUserSettings**: One-to-One
- **MangaProvider → MangaSubscription**: One-to-Many
- **MangaProvider → MangaHistory**: One-to-Many

### Web Novel Module Relationships
- **User → Novel**: One-to-Many (added_by)
- **Novel → Chapter**: One-to-Many
- **Chapter → Paragraph**: One-to-Many

### Subtitle Module Relationships
- **User → Subtitle**: One-to-Many (created_by)
- **User → SubtitleUserState**: One-to-Many
- **SubtitleProvider → SubtitleProviderService**: One-to-Many
- **SubtitleProvider → Subtitle**: One-to-Many
- **SubtitleProviderService → Subtitle**: One-to-Many
- **Subtitle → SubtitleSegment**: One-to-Many
- **Subtitle → SubtitleUserState**: One-to-Many

---

## Indexes Summary

### Manga Module
```sql
CREATE INDEX idx_manga_providers_owner_id ON manga_providers(owner_id);
CREATE INDEX idx_manga_providers_visibility ON manga_providers(visibility);

CREATE INDEX idx_manga_subscriptions_user_id ON manga_subscriptions(user_id);
CREATE INDEX idx_manga_subscriptions_provider_id ON manga_subscriptions(provider_id);

CREATE INDEX idx_manga_history_user_id ON manga_history(user_id);
CREATE INDEX idx_manga_history_provider_id ON manga_history(provider_id);
CREATE INDEX idx_manga_history_last_read_at ON manga_history(last_read_at DESC);
CREATE INDEX idx_manga_history_user_provider ON manga_history(user_id, provider_id);
CREATE INDEX idx_manga_history_user_manga ON manga_history(user_id, manga_id);
```

### Web Novel Module
```sql
CREATE INDEX idx_webnovel_novels_added_by ON webnovel_novels(added_by);
CREATE INDEX idx_webnovel_novels_source_source_id ON webnovel_novels(source, source_id);

CREATE INDEX idx_webnovel_chapters_novel_id ON webnovel_chapters(novel_id);
CREATE INDEX idx_webnovel_chapters_novel_idx ON webnovel_chapters(novel_id, idx);

CREATE INDEX idx_webnovel_paragraphs_chapter_id ON webnovel_paragraphs(chapter_id);
CREATE INDEX idx_webnovel_paragraphs_chapter_idx ON webnovel_paragraphs(chapter_id, idx);
```

### Subtitle Module
```sql
CREATE INDEX idx_subtitles_provider_id ON subtitles(provider_id);
CREATE INDEX idx_subtitles_service_id ON subtitles(service_id);
CREATE INDEX idx_subtitles_created_by ON subtitles(created_by);
CREATE INDEX idx_subtitles_video_id ON subtitles(video_id);
CREATE INDEX idx_subtitles_status ON subtitles(status);

CREATE INDEX idx_subtitle_segments_subtitle_id ON subtitle_segments(subtitle_id);
CREATE INDEX idx_subtitle_segments_subtitle_idx ON subtitle_segments(subtitle_id, idx);

CREATE INDEX idx_subtitle_user_states_user_id ON subtitle_user_states(user_id);
CREATE INDEX idx_subtitle_user_states_subtitle_id ON subtitle_user_states(subtitle_id);
```

---

## Data Types Notes

- **UUID**: Stored as TEXT in standard UUID format (xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx)
- **Timestamps**: Stored as TEXT in ISO8601 format (YYYY-MM-DDTHH:MM:SSZ)
- **Booleans**: Stored as INTEGER (0 or 1)
- **JSON**: Stored as TEXT containing valid JSON

---

## Migration Notes

The schema is initialized via the `InitializeSchema` function in each module's persistent package. Tables are created with `CREATE TABLE IF NOT EXISTS` and indexes are created separately.

For production PostgreSQL deployment:
1. Change UUID storage to native `UUID` type
2. Change timestamps to `TIMESTAMP WITH TIME ZONE`
3. Use native `BOOLEAN` type
4. Use native `JSONB` for JSON columns
5. Add foreign key constraints with `ON DELETE CASCADE` where appropriate

---

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2026-10-15 | Initial schema with users, manga, webnovel, subtitle modules |