BEGIN;

DO $$ BEGIN
  CREATE TYPE broadcasterType AS ENUM ('partner', 'affiliate', 'none');
EXCEPTION
  WHEN duplicate_object THEN null;
END $$;

CREATE TABLE IF NOT EXISTS tracked_channels (
  bc_id varchar PRIMARY KEY,
  bc_display_name varchar NOT NULL,
  bc_username varchar NOT NULL,
  bc_type broadcasterType NOT NULL,
  pp_url text,
  offline_pp_url text,
  tracked_since timestamp DEFAULT now(),
  seen_inactive_count int DEFAULT 0,
  enabled_status boolean DEFAULT true,
  last_modified_status timestamp,
  priority_lvl smallint DEFAULT 0
);

CREATE TABLE IF NOT EXISTS vods (
  video_id varchar PRIMARY KEY,
  stream_id varchar NOT NULL,
  bc_id varchar NOT NULL REFERENCES tracked_channels(bc_id),
  created_at timestamp NOT NULL,
  published_at timestamp NOT NULL,
  duration_seconds int NOT NULL,
  lang varchar NOT NULL,
  thumbnail_url text NOT NULL,
  title varchar NOT NULL,
  view_count int NOT NULL
);

CREATE TABLE IF NOT EXISTS clips (
  clip_id varchar PRIMARY KEY,
  bc_id varchar NOT NULL REFERENCES tracked_channels(bc_id),
  -- Keep in mind some video_id's are not just vods
  video_id varchar,
  created_at timestamp NOT NULL,
  creator_id varchar NOT NULL,
  creator_name varchar NOT NULL,
  title varchar NOT NULL,
  game_id varchar,
  lang varchar NOT NULL,
  thumbnail_url text NOT NULL,
  duration_seconds decimal(5, 2) NOT NULL,
  view_count int NOT NULL,
  vod_offset int
);

CREATE INDEX video_id_idx ON clips USING btree (video_id);
CREATE INDEX username_idx ON tracked_channels USING btree (bc_username);

COMMIT;