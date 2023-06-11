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
  enabled_status boolean DEFAULT true,
  last_modified_status timestamp,
  priority_lvl smallint DEFAULT 0
);

COMMIT;