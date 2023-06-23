BEGIN;

DROP INDEX IF EXISTS video_id_idx;

DROP TABLE IF EXISTS clips;
DROP TABLE IF EXISTS vods;
DROP TABLE IF EXISTS tracked_channels;

DROP TYPE IF EXISTS broadcasterType;

COMMIT;