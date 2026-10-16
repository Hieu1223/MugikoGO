-- MugikoGO PostgreSQL schema initialization
-- Based on the provided Database Design Document (version 1.0, 2026-10-15).
-- Run this file against the target database, for example:
--   psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f init_schema.sql

BEGIN;

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    bio TEXT,
    email TEXT UNIQUE,
    role TEXT NOT NULL DEFAULT 'user',
    firebase_uid TEXT UNIQUE,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

-- Manga module

CREATE TABLE IF NOT EXISTS manga_providers (
    id UUID PRIMARY KEY,
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    base_url TEXT NOT NULL,
    visibility TEXT NOT NULL DEFAULT 'private',
    is_builtin BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_manga_providers_owner_id
    ON manga_providers(owner_id);
CREATE INDEX IF NOT EXISTS idx_manga_providers_visibility
    ON manga_providers(visibility);

CREATE TABLE IF NOT EXISTS manga_subscriptions (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider_id UUID NOT NULL REFERENCES manga_providers(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (user_id, provider_id)
);

CREATE INDEX IF NOT EXISTS idx_manga_subscriptions_user_id
    ON manga_subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_manga_subscriptions_provider_id
    ON manga_subscriptions(provider_id);

CREATE TABLE IF NOT EXISTS manga_history (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider_id UUID NOT NULL REFERENCES manga_providers(id) ON DELETE CASCADE,
    manga_id TEXT NOT NULL,
    chapter_id TEXT NOT NULL,
    page_index INTEGER NOT NULL DEFAULT 0,
    title TEXT NOT NULL,
    cover_url TEXT,
    last_read_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_manga_history_user_id
    ON manga_history(user_id);
CREATE INDEX IF NOT EXISTS idx_manga_history_provider_id
    ON manga_history(provider_id);
CREATE INDEX IF NOT EXISTS idx_manga_history_last_read_at
    ON manga_history(last_read_at DESC);
CREATE INDEX IF NOT EXISTS idx_manga_history_user_provider
    ON manga_history(user_id, provider_id);
CREATE INDEX IF NOT EXISTS idx_manga_history_user_manga
    ON manga_history(user_id, manga_id);

CREATE TABLE IF NOT EXISTS manga_user_settings (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    schema_version INTEGER NOT NULL DEFAULT 1,
    data JSONB NOT NULL DEFAULT '{}'::jsonb,
    updated_at TIMESTAMPTZ NOT NULL
);

-- Web novel module

CREATE TABLE IF NOT EXISTS webnovel_novels (
    id UUID PRIMARY KEY,
    source TEXT NOT NULL,
    source_id TEXT NOT NULL,
    url TEXT NOT NULL,
    title TEXT NOT NULL,
    author TEXT,
    synopsis TEXT,
    added_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL,
    UNIQUE (source, source_id)
);

CREATE INDEX IF NOT EXISTS idx_webnovel_novels_added_by
    ON webnovel_novels(added_by);
CREATE INDEX IF NOT EXISTS idx_webnovel_novels_source_source_id
    ON webnovel_novels(source, source_id);

CREATE TABLE IF NOT EXISTS webnovel_chapters (
    id UUID PRIMARY KEY,
    novel_id UUID NOT NULL REFERENCES webnovel_novels(id) ON DELETE CASCADE,
    idx INTEGER NOT NULL,
    title TEXT,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_webnovel_chapters_novel_id
    ON webnovel_chapters(novel_id);
CREATE INDEX IF NOT EXISTS idx_webnovel_chapters_novel_idx
    ON webnovel_chapters(novel_id, idx);

CREATE TABLE IF NOT EXISTS webnovel_paragraphs (
    id UUID PRIMARY KEY,
    chapter_id UUID NOT NULL REFERENCES webnovel_chapters(id) ON DELETE CASCADE,
    idx INTEGER NOT NULL,
    text TEXT NOT NULL,
    words JSONB NOT NULL DEFAULT '[]'::jsonb
);

CREATE INDEX IF NOT EXISTS idx_webnovel_paragraphs_chapter_id
    ON webnovel_paragraphs(chapter_id);
CREATE INDEX IF NOT EXISTS idx_webnovel_paragraphs_chapter_idx
    ON webnovel_paragraphs(chapter_id, idx);

-- Subtitle module

CREATE TABLE IF NOT EXISTS subtitle_providers (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    kind TEXT NOT NULL,
    base_url TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_synced_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS subtitle_provider_services (
    id UUID PRIMARY KEY,
    provider_id UUID NOT NULL REFERENCES subtitle_providers(id) ON DELETE CASCADE,
    service_key TEXT NOT NULL,
    name TEXT NOT NULL,
    options JSONB,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS subtitles (
    id UUID PRIMARY KEY,
    video_id TEXT NOT NULL,
    video_url TEXT NOT NULL,
    provider_id UUID NOT NULL REFERENCES subtitle_providers(id) ON DELETE CASCADE,
    service_id UUID NOT NULL REFERENCES subtitle_provider_services(id) ON DELETE CASCADE,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'queued',
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_subtitles_provider_id
    ON subtitles(provider_id);
CREATE INDEX IF NOT EXISTS idx_subtitles_service_id
    ON subtitles(service_id);
CREATE INDEX IF NOT EXISTS idx_subtitles_created_by
    ON subtitles(created_by);
CREATE INDEX IF NOT EXISTS idx_subtitles_video_id
    ON subtitles(video_id);
CREATE INDEX IF NOT EXISTS idx_subtitles_status
    ON subtitles(status);

CREATE TABLE IF NOT EXISTS subtitle_segments (
    id UUID PRIMARY KEY,
    subtitle_id UUID NOT NULL REFERENCES subtitles(id) ON DELETE CASCADE,
    idx INTEGER NOT NULL,
    start_ms INTEGER NOT NULL,
    end_ms INTEGER NOT NULL,
    words JSONB,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_subtitle_segments_subtitle_id
    ON subtitle_segments(subtitle_id);
CREATE INDEX IF NOT EXISTS idx_subtitle_segments_subtitle_idx
    ON subtitle_segments(subtitle_id, idx);

CREATE TABLE IF NOT EXISTS subtitle_user_states (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subtitle_id UUID NOT NULL REFERENCES subtitles(id) ON DELETE CASCADE,
    schema_version INTEGER NOT NULL DEFAULT 1,
    data JSONB NOT NULL DEFAULT '{}'::jsonb,
    updated_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (user_id, subtitle_id)
);

CREATE INDEX IF NOT EXISTS idx_subtitle_user_states_user_id
    ON subtitle_user_states(user_id);
CREATE INDEX IF NOT EXISTS idx_subtitle_user_states_subtitle_id
    ON subtitle_user_states(subtitle_id);

COMMIT;
