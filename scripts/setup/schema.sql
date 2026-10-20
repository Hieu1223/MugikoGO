-- MugikoGO PostgreSQL schema initialization
-- Mirrors the persistent layer in internal/modules/**/persistent exactly.
--
-- Apply:
--   make setup
-- or:
--   <PYTHON> scripts/common/run_sql.py -f scripts/setup/schema.sql
--
-- Reset the database first with:
--   <PYTHON> scripts/reset-db.py
--
-- NOTE: the Sequin CDC section (publication + replication slot) runs outside
-- the main transaction on purpose: pg_create_logical_replication_slot is not
-- allowed in a transaction that has performed writes, so a separate
-- read-only transaction is used there.

BEGIN;

-- Authentication & core user identity
CREATE TABLE IF NOT EXISTS users (
    user_id    TEXT PRIMARY KEY,
    role       TEXT NOT NULL DEFAULT 'user',
    auth_type  TEXT NOT NULL,
    auth_id    TEXT NOT NULL,
    is_active  BOOLEAN NOT NULL DEFAULT TRUE,
    joined_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (auth_type, auth_id)
);

CREATE INDEX IF NOT EXISTS idx_users_auth
    ON users(auth_type, auth_id);
CREATE INDEX IF NOT EXISTS idx_users_role
    ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_joined_at
    ON users(joined_at DESC);

-- User profile / display data
CREATE TABLE IF NOT EXISTS user_profiles (
    user_id TEXT PRIMARY KEY REFERENCES users(user_id) ON DELETE CASCADE,
    name    TEXT NOT NULL,
    bio     TEXT,
    email   TEXT UNIQUE
);

CREATE INDEX IF NOT EXISTS idx_user_profiles_email
    ON user_profiles(email);

-- Manga module

CREATE TABLE IF NOT EXISTS manga_providers (
    id         UUID PRIMARY KEY,
    owner_id   TEXT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    base_url   TEXT NOT NULL,
    visibility TEXT NOT NULL DEFAULT 'private',
    is_builtin BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_manga_providers_owner_id
    ON manga_providers(owner_id);
CREATE INDEX IF NOT EXISTS idx_manga_providers_visibility
    ON manga_providers(visibility);

CREATE TABLE IF NOT EXISTS manga_subscriptions (
    user_id     TEXT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    provider_id UUID NOT NULL REFERENCES manga_providers(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, provider_id)
);

CREATE INDEX IF NOT EXISTS idx_manga_subscriptions_user_id
    ON manga_subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_manga_subscriptions_provider_id
    ON manga_subscriptions(provider_id);

CREATE TABLE IF NOT EXISTS manga_history (
    user_id      TEXT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    provider_id  UUID NOT NULL REFERENCES manga_providers(id) ON DELETE CASCADE,
    manga_id     TEXT NOT NULL,
    chapter_id   TEXT NOT NULL,
    page_index   INTEGER NOT NULL DEFAULT 0,
    title        TEXT NOT NULL,
    cover_url    TEXT,
    last_read_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, provider_id, manga_id)
);

CREATE INDEX IF NOT EXISTS idx_manga_history_last_read_at
    ON manga_history(last_read_at DESC);
CREATE INDEX IF NOT EXISTS idx_manga_history_user_last_read
    ON manga_history(user_id, last_read_at DESC);

CREATE TABLE IF NOT EXISTS manga_user_settings (
    user_id        TEXT PRIMARY KEY REFERENCES users(user_id) ON DELETE CASCADE,
    schema_version INTEGER NOT NULL DEFAULT 1,
    data           JSONB NOT NULL DEFAULT '{}'::jsonb,
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- User settings (general, used by subtitle module)
CREATE TABLE IF NOT EXISTS user_settings (
    user_id        TEXT PRIMARY KEY REFERENCES users(user_id) ON DELETE CASCADE,
    schema_version INTEGER NOT NULL DEFAULT 1,
    data           JSONB NOT NULL DEFAULT '{}'::jsonb,
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_settings_user_id
    ON user_settings(user_id);


-- Web novel module

CREATE TABLE IF NOT EXISTS novels (
    id         UUID PRIMARY KEY,
    source     TEXT NOT NULL,
    source_id  TEXT NOT NULL,
    url        TEXT NOT NULL,
    title      TEXT NOT NULL,
    author     TEXT,
    synopsis   TEXT,
    added_by   TEXT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (source, source_id)
);

CREATE INDEX IF NOT EXISTS idx_novels_added_by
    ON novels(added_by);
CREATE INDEX IF NOT EXISTS idx_novels_source_source_id
    ON novels(source, source_id);

CREATE TABLE IF NOT EXISTS chapters (
    id         UUID PRIMARY KEY,
    novel_id   UUID NOT NULL REFERENCES novels(id) ON DELETE CASCADE,
    idx        INTEGER NOT NULL,
    title      TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_chapters_novel_id
    ON chapters(novel_id);
CREATE INDEX IF NOT EXISTS idx_chapters_novel_idx
    ON chapters(novel_id, idx);

CREATE TABLE IF NOT EXISTS paragraphs (
    id         UUID PRIMARY KEY,
    chapter_id UUID NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
    idx        INTEGER NOT NULL,
    text       TEXT NOT NULL,
    words      JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_paragraphs_chapter_id
    ON paragraphs(chapter_id);
CREATE INDEX IF NOT EXISTS idx_paragraphs_chapter_idx
    ON paragraphs(chapter_id, idx);

-- Subtitle module

CREATE TABLE IF NOT EXISTS subtitle_providers (
    id             UUID PRIMARY KEY,
    name           TEXT NOT NULL,
    kind           TEXT NOT NULL,
    base_url       TEXT,
    is_active      BOOLEAN NOT NULL DEFAULT TRUE,
    last_synced_at TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS subtitle_provider_services (
    id          UUID PRIMARY KEY,
    provider_id UUID NOT NULL REFERENCES subtitle_providers(id) ON DELETE CASCADE,
    service_key TEXT NOT NULL,
    name        TEXT NOT NULL,
    options     JSONB,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS subtitles (
    id          UUID PRIMARY KEY,
    video_id    TEXT NOT NULL,
    video_url   TEXT NOT NULL,
    provider_id UUID NOT NULL REFERENCES subtitle_providers(id) ON DELETE CASCADE,
    service_id  UUID NOT NULL REFERENCES subtitle_provider_services(id) ON DELETE CASCADE,
    created_by  TEXT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    status      TEXT NOT NULL DEFAULT 'queued',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
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
    id          UUID PRIMARY KEY,
    subtitle_id UUID NOT NULL REFERENCES subtitles(id) ON DELETE CASCADE,
    idx         INTEGER NOT NULL,
    start_ms    INTEGER NOT NULL,
    end_ms      INTEGER NOT NULL,
    words       JSONB,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_subtitle_segments_subtitle_id
    ON subtitle_segments(subtitle_id);
CREATE INDEX IF NOT EXISTS idx_subtitle_segments_subtitle_idx
    ON subtitle_segments(subtitle_id, idx);

CREATE TABLE IF NOT EXISTS subtitle_access (
    user_id          TEXT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    subtitle_id      UUID NOT NULL REFERENCES subtitles(id) ON DELETE CASCADE,
    last_accessed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, subtitle_id)
);

CREATE INDEX IF NOT EXISTS idx_subtitle_access_subtitle_id
    ON subtitle_access(subtitle_id);
CREATE INDEX IF NOT EXISTS idx_subtitle_access_last_accessed_at
    ON subtitle_access(last_accessed_at DESC);

CREATE TABLE IF NOT EXISTS subtitle_user_states (
    user_id        TEXT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    subtitle_id    UUID NOT NULL REFERENCES subtitles(id) ON DELETE CASCADE,
    schema_version INTEGER NOT NULL DEFAULT 1,
    data           JSONB NOT NULL DEFAULT '{}'::jsonb,
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, subtitle_id)
);

CREATE INDEX IF NOT EXISTS idx_subtitle_user_states_subtitle_id
    ON subtitle_user_states(subtitle_id);


-- OCR module

CREATE TABLE IF NOT EXISTS ocr_results (
    id         UUID PRIMARY KEY,
    image_url  TEXT NOT NULL,
    status     TEXT NOT NULL DEFAULT 'queued',
    -- PaddleOCR output: {"dt_polys": [...], "rec_texts": [...], "rec_polys": [...]}
    ocr_data   JSONB,
    error      TEXT,
    created_by TEXT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ocr_results_status
    ON ocr_results(status);
CREATE INDEX IF NOT EXISTS idx_ocr_results_created_by
    ON ocr_results(created_by);
CREATE INDEX IF NOT EXISTS idx_ocr_results_created_at
    ON ocr_results(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_ocr_results_image_url
    ON ocr_results(image_url);


-- Deck / Flashcard module

CREATE TABLE IF NOT EXISTS decks (
    id         TEXT PRIMARY KEY,
    owner_id   TEXT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_decks_owner_id
    ON decks(owner_id);
CREATE INDEX IF NOT EXISTS idx_decks_created_at
    ON decks(created_at DESC);

CREATE TABLE IF NOT EXISTS deck_words (
    id         TEXT PRIMARY KEY,
    deck_id    TEXT NOT NULL REFERENCES decks(id) ON DELETE CASCADE,
    word       TEXT NOT NULL,
    meaning    TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_deck_words_deck_id
    ON deck_words(deck_id);
CREATE INDEX IF NOT EXISTS idx_deck_words_created_at
    ON deck_words(created_at DESC);


COMMIT;

-- Sequin CDC: keep the publication in sync with the persistent-layer tables it
-- must stream. Recreated on every run so that the table set always matches,
-- even after tables have been dropped and re-created above.
DO $$
BEGIN
    DROP PUBLICATION IF EXISTS sequin_pub;
    CREATE PUBLICATION sequin_pub FOR TABLE novels, subtitles, ocr_results WITH (publish_via_partition_root = true);
END
$$;

COMMIT;

-- Sequin CDC: slot creation is not transactional and is not allowed in a
-- transaction that has performed writes, so it runs in its own read-only
-- transaction (the publication DO block above was committed separately).
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_replication_slots WHERE slot_name = 'sequin_slot') THEN
        PERFORM pg_create_logical_replication_slot('sequin_slot', 'pgoutput');
    END IF;
END
$$;
