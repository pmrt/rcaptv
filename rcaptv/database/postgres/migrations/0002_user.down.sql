BEGIN;

DROP INDEX IF EXISTS access_token_idx;
DROP INDEX IF EXISTS refresh_token_user_id_token_pair_idx;
DROP INDEX IF EXISTS twitch_user_id_users_idx;
DROP INDEX IF EXISTS username_users_idx;

DROP TABLE IF EXISTS token_pair;
DROP TABLE IF EXISTS users;

COMMIT;