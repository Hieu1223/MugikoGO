-- MugikoGO subtitle module seed data
-- Requires init_schema.sql to have been applied first.
-- Safe to re-run: rows are inserted only when their stable IDs are absent.
-- Replace example URLs/IDs with real provider and video values as needed.

BEGIN;

-- Provider records
INSERT INTO subtitle_providers (
    id, name, kind, base_url, is_active,
    last_synced_at, created_at, updated_at
)
VALUES
    (
        '10000000-0000-4000-8000-000000000001',
        'YouTube',
        'youtube',
        'https://www.youtube.com',
        TRUE,
        NULL,
        TIMESTAMPTZ '2026-01-01 00:00:00+00',
        TIMESTAMPTZ '2026-01-01 00:00:00+00'
    ),
    (
        '10000000-0000-4000-8000-000000000002',
        'Example Subtitle API',
        'remote',
        'https://example.com/api/subtitles',
        FALSE,
        NULL,
        TIMESTAMPTZ '2026-01-01 00:00:00+00',
        TIMESTAMPTZ '2026-01-01 00:00:00+00'
    )
ON CONFLICT (id) DO NOTHING;

-- Provider services
INSERT INTO subtitle_provider_services (
    id, provider_id, service_key, name, options, is_active,
    created_at, updated_at
)
VALUES
    (
        '20000000-0000-4000-8000-000000000001',
        '10000000-0000-4000-8000-000000000001',
        'youtube-transcript',
        'YouTube Transcript',
        '{"languages":["ja","en"],"prefer_manual":true}'::jsonb,
        TRUE,
        TIMESTAMPTZ '2026-01-01 00:00:00+00',
        TIMESTAMPTZ '2026-01-01 00:00:00+00'
    ),
    (
        '20000000-0000-4000-8000-000000000002',
        '10000000-0000-4000-8000-000000000002',
        'remote-subtitle-search',
        'Remote Subtitle Search',
        '{"languages":["ja","en","vi"]}'::jsonb,
        FALSE,
        TIMESTAMPTZ '2026-01-01 00:00:00+00',
        TIMESTAMPTZ '2026-01-01 00:00:00+00'
    )
ON CONFLICT (id) DO NOTHING;

-- Subtitle jobs.
-- The creator must already exist in users; replace this UUID with a real user ID.
-- These are illustrative records and do not imply that the jobs have actually run.
INSERT INTO subtitles (
    id, video_id, video_url, provider_id, service_id, created_by,
    status, created_at, updated_at
)
SELECT
    seed.id,
    seed.video_id,
    seed.video_url,
    seed.provider_id,
    seed.service_id,
    seed.created_by,
    seed.status,
    seed.created_at,
    seed.updated_at
FROM (
    VALUES
    (
        '30000000-0000-4000-8000-000000000001'::uuid,
        'dQw4w9WgXcQ',
        'https://www.youtube.com/watch?v=dQw4w9WgXcQ',
        '10000000-0000-4000-8000-000000000001'::uuid,
        '20000000-0000-4000-8000-000000000001'::uuid,
        NULL::uuid,
        'completed',
        TIMESTAMPTZ '2026-01-01 00:00:00+00',
        TIMESTAMPTZ '2026-01-01 00:00:00+00'
    ),
    (
        '30000000-0000-4000-8000-000000000002'::uuid,
        'M7lc1UVf-VE',
        'https://www.youtube.com/watch?v=M7lc1UVf-VE',
        '10000000-0000-4000-8000-000000000001'::uuid,
        '20000000-0000-4000-8000-000000000001'::uuid,
        NULL::uuid,
        'queued',
        TIMESTAMPTZ '2026-01-01 00:00:00+00',
        TIMESTAMPTZ '2026-01-01 00:00:00+00'
    )
) AS seed(
    id, video_id, video_url, provider_id, service_id, created_by,
    status, created_at, updated_at
)
WHERE seed.created_by IS NOT NULL
  AND EXISTS (SELECT 1 FROM users WHERE users.id = seed.created_by)
ON CONFLICT (id) DO NOTHING;

-- Example subtitle segments are inserted only for the completed example job.
-- These are demonstration Japanese-learning lines, not fetched video captions.
INSERT INTO subtitle_segments (
    id, subtitle_id, idx, start_ms, end_ms, words, created_at
)
SELECT
    seed.id,
    seed.subtitle_id,
    seed.idx,
    seed.start_ms,
    seed.end_ms,
    seed.words,
    seed.created_at
FROM (
    VALUES
    (
        '40000000-0000-4000-8000-000000000001'::uuid,
        '30000000-0000-4000-8000-000000000001'::uuid,
        0,
        0,
        2200,
        '[{"text":"こんにちは","start_ms":0,"end_ms":900},{"text":"。","start_ms":900,"end_ms":1000}]'::jsonb,
        TIMESTAMPTZ '2026-01-01 00:00:00+00'
    ),
    (
        '40000000-0000-4000-8000-000000000002'::uuid,
        '30000000-0000-4000-8000-000000000001'::uuid,
        1,
        2200,
        4800,
        '[{"text":"今日は","start_ms":2200,"end_ms":3000},{"text":"いい天気ですね","start_ms":3000,"end_ms":4600},{"text":"。","start_ms":4600,"end_ms":4800}]'::jsonb,
        TIMESTAMPTZ '2026-01-01 00:00:00+00'
    )
) AS seed(id, subtitle_id, idx, start_ms, end_ms, words, created_at)
WHERE EXISTS (
    SELECT 1 FROM subtitles
    WHERE subtitles.id = seed.subtitle_id
      AND subtitles.status = 'completed'
)
ON CONFLICT (id) DO NOTHING;

-- Per-user playback state is intentionally omitted because it requires a real user ID.
-- Add subtitle_user_states rows after replacing the user/subtitle IDs with valid IDs.

COMMIT;
