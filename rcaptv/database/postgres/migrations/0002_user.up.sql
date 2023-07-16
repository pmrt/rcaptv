BEGIN;

CREATE TABLE IF NOT EXISTS users (
  user_id SERIAL PRIMARY KEY,
  twitch_user_id varchar NOT NULL,
  username varchar NOT NULL,
  display_username varchar NOT NULL,
  email text NOT NULL,
  pp_url text,
  is_paid_user boolean DEFAULT false,
  is_vip boolean DEFAULT false,
  last_payment_at timestamp,
  bc_type broadcasterType NOT NULL,
  last_login_at timestamp DEFAULT now(),
  created_at timestamp DEFAULT now(),
  twitch_created_at timestamp NOT NULL
);

CREATE TABLE IF NOT EXISTS token_pairs (
  token_pair_id SERIAL PRIMARY KEY,
  user_id integer NOT NULL REFERENCES users(user_id),
  access_token text NOT NULL,
  refresh_token text NOT NULL,
  expires_at timestamp NOT NULL,
  last_modified_at timestamp DEFAULT now()
);

CREATE INDEX IF NOT EXISTS username_users_idx ON users USING btree (username);
CREATE UNIQUE INDEX IF NOT EXISTS twitch_user_id_users_idx ON users USING btree (twitch_user_id);
CREATE UNIQUE INDEX IF NOT EXISTS refresh_token_user_id_token_pair_idx ON token_pairs USING btree (user_id, refresh_token);
CREATE INDEX IF NOT EXISTS access_token_idx ON token_pairs USING btree (access_token);

COMMIT;