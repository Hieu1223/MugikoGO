BEGIN;
SET session_replication_role = 'replica';

INSERT INTO "subtitle_provider_services" (id, provider_id, service_key, name, options, is_active, created_at, updated_at) VALUES ('20000000-0000-4000-8000-000000000001', '10000000-0000-4000-8000-000000000001', 'youtube-transcript', 'YouTube Transcript', {'languages': ['ja', 'en'], 'prefer_manual': True}, True, 2026-01-01 00:00:00+00:00, 2026-01-01 00:00:00+00:00);
INSERT INTO "subtitle_provider_services" (id, provider_id, service_key, name, options, is_active, created_at, updated_at) VALUES ('20000000-0000-4000-8000-000000000002', '10000000-0000-4000-8000-000000000002', 'remote-subtitle-search', 'Remote Subtitle Search', {'languages': ['ja', 'en', 'vi']}, False, 2026-01-01 00:00:00+00:00, 2026-01-01 00:00:00+00:00);
INSERT INTO "subtitle_providers" (id, name, kind, base_url, is_active, last_synced_at, created_at, updated_at) VALUES ('10000000-0000-4000-8000-000000000001', 'YouTube', 'youtube', 'https://www.youtube.com', True, NULL, 2026-01-01 00:00:00+00:00, 2026-01-01 00:00:00+00:00);
INSERT INTO "subtitle_providers" (id, name, kind, base_url, is_active, last_synced_at, created_at, updated_at) VALUES ('10000000-0000-4000-8000-000000000002', 'Example Subtitle API', 'remote', 'https://example.com/api/subtitles', False, NULL, 2026-01-01 00:00:00+00:00, 2026-01-01 00:00:00+00:00);

SET session_replication_role = 'origin';
COMMIT;
