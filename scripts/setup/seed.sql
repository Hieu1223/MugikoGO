-- Seed script for subtitle providers and services
-- Run this after init_schema.sql to populate test data

BEGIN;

-- ============================================
-- Subtitle Providers
-- ============================================

INSERT INTO subtitle_providers (id, name, kind, base_url, is_active, last_synced_at, created_at, updated_at) VALUES
    -- OpenSubtitles provider
    (gen_random_uuid(), 'OpenSubtitles', 'opensubtitles', 'https://api.opensubtitles.com/api/v1', TRUE, NOW() - INTERVAL '1 hour', NOW(), NOW()),
    
    -- SubDL provider
    (gen_random_uuid(), 'SubDL', 'subdl', 'https://api.subdl.com', TRUE, NOW() - INTERVAL '2 hours', NOW(), NOW()),
    
    -- Addic7ed provider
    (gen_random_uuid(), 'Addic7ed', 'addic7ed', 'https://www.addic7ed.com', TRUE, NOW() - INTERVAL '30 minutes', NOW(), NOW()),
    
    -- Local file provider (for testing)
    (gen_random_uuid(), 'Local Files', 'local', NULL, TRUE, NULL, NOW(), NOW());

-- ============================================
-- Subtitle Provider Services
-- ============================================

-- Get the provider IDs we just created
DO $$
DECLARE
    opensubtitles_id UUID;
    subdl_id UUID;
    addic7ed_id UUID;
    local_id UUID;
BEGIN
    -- Get provider IDs
    SELECT id INTO opensubtitles_id FROM subtitle_providers WHERE name = 'OpenSubtitles';
    SELECT id INTO subdl_id FROM subtitle_providers WHERE name = 'SubDL';
    SELECT id INTO addic7ed_id FROM subtitle_providers WHERE name = 'Addic7ed';
    SELECT id INTO local_id FROM subtitle_providers WHERE name = 'Local Files';

    -- OpenSubtitles services
    INSERT INTO subtitle_provider_services (id, provider_id, service_key, name, options, is_active, created_at, updated_at) VALUES
        (gen_random_uuid(), opensubtitles_id, 'opensubtitles_en', 'English', '{"language": "en", "format": "srt"}'::jsonb, TRUE, NOW(), NOW()),
        (gen_random_uuid(), opensubtitles_id, 'opensubtitles_ja', 'Japanese', '{"language": "ja", "format": "srt"}'::jsonb, TRUE, NOW(), NOW()),
        (gen_random_uuid(), opensubtitles_id, 'opensubtitles_zh', 'Chinese', '{"language": "zh", "format": "srt"}'::jsonb, TRUE, NOW(), NOW()),
        (gen_random_uuid(), opensubtitles_id, 'opensubtitles_es', 'Spanish', '{"language": "es", "format": "srt"}'::jsonb, FALSE, NOW(), NOW());

    -- SubDL services
    INSERT INTO subtitle_provider_services (id, provider_id, service_key, name, options, is_active, created_at, updated_at) VALUES
        (gen_random_uuid(), subdl_id, 'subdl_multi', 'Multi-Language', '{"languages": ["en", "ja", "ko"], "format": "ass"}'::jsonb, TRUE, NOW(), NOW()),
        (gen_random_uuid(), subdl_id, 'subdl_anime', 'Anime Specialized', '{"type": "anime", "format": "ass", "styling": true}'::jsonb, TRUE, NOW(), NOW());

    -- Addic7ed services
    INSERT INTO subtitle_provider_services (id, provider_id, service_key, name, options, is_active, created_at, updated_at) VALUES
        (gen_random_uuid(), addic7ed_id, 'addic7ed_tv', 'TV Shows', '{"category": "tv"}'::jsonb, TRUE, NOW(), NOW()),
        (gen_random_uuid(), addic7ed_id, 'addic7ed_movies', 'Movies', '{"category": "movies"}'::jsonb, TRUE, NOW(), NOW());

    -- Local Files service
    INSERT INTO subtitle_provider_services (id, provider_id, service_key, name, options, is_active, created_at, updated_at) VALUES
        (gen_random_uuid(), local_id, 'local_auto', 'Auto-Detect', '{"directory": "/subtitles", "auto_scan": true}'::jsonb, TRUE, NOW(), NOW());

END $$;

COMMIT;