BEGIN;

DROP INDEX IF EXISTS video_id_idx;
DROP INDEX IF EXISTS username_idx;
DROP INDEX IF EXISTS vods_created_at_idx;

DROP TABLE IF EXISTS clips;
DROP TABLE IF EXISTS vods;
DROP TABLE IF EXISTS tracked_channels;

DROP TYPE IF EXISTS broadcasterType;

COMMIT;