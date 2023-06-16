BEGIN;

DO $$ BEGIN
  CREATE TYPE langISO31661 AS ENUM (
	'aa','hy','or','ab','hz','os','af','id','pa','ak','ig','pl',
	'am','ii','ps','an','ik','pt','ar','io','qu','as','is','rm',
	'av','it','rn','ay','iu','ro','az','ja','ru','ba','jv','rw',
	'be','ka','sa','bg','kg','sc','bh','ki','sd','bi','kj','se'	,
	'bm','kk','sg','bn','kl','si','bo','km','sk','br','kn','sl',
	'bs','ko','sm','ca','kr','sn','ce','ks','so','ch','ku','sq',
	'co','kv','sr','cr','kw','ss','cs','ky','st','cv','lb','su',
	'cy','lg','sv','da','li','sw','de','ln','ta','dv','lo','te',
	'dz','lt','tg','ee','lu','th','el','lv','ti','en','mg','tk',
	'es','mh','tl','et','mi','tn','eu','mk','to','fa','ml','tr',
	'ff','mn','ts','fi','mr','tt','fj','ms','tw','fo','mt','ty',
	'fr','my','ug','fy','na','uk','ga','nb','ur','gd','nd','uz',
	'gl','ne','ve','gn','ng','vi','gu','nl','wa','gv','nn','wo',
	'ha','no','xh','he','nr','yi','hi','nv','yo','ho','ny','za',
	'hr','oc','zh','ht','oj','zu','hu','om','other'
	);
  CREATE TYPE broadcasterType AS ENUM ('partner', 'affiliate', 'none');
	-- https://dev.twitch.tv/docs/eventsub/eventsub-reference#events
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

CREATE TABLE IF NOT EXISTS vods (
  video_id varchar PRIMARY KEY,
  stream_id varchar NOT NULL,
  bc_id varchar NOT NULL REFERENCES tracked_channels(bc_id),
  created_at timestamp NOT NULL,
  published_at timestamp NOT NULL,
  duration_seconds int NOT NULL,
  lang langISO31661 NOT NULL,
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
  lang langISO31661 NOT NULL,
  thumbnail_url text NOT NULL,
  duration_seconds decimal(5, 2) NOT NULL,
  view_count int NOT NULL,
  vod_offset int
);

CREATE INDEX video_id_idx ON clips USING btree (video_id);

COMMIT;