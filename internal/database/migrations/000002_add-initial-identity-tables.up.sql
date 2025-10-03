CREATE TABLE IF NOT EXISTS idp_users (
	user_id TEXT PRIMARY KEY,
	username TEXT NOT NULL UNIQUE,
	email TEXT NOT NULL UNIQUE,
	password_hash BLOB NOT NULL
) STRICT;

CREATE TABLE IF NOT EXISTS idp_clients (
	client_id TEXT PRIMARY KEY,
	client_secret TEXT NOT NULL
) STRICT;

CREATE TABLE IF NOT EXISTS idp_refresh_tokens (
	token_id TEXT PRIMARY KEY,
	client_id TEXT NOT NULL REFERENCES idp_clients (client_id),
	jwt TEXT NOT NULL,
	revoked INTEGER NOT NULL CHECK (revoked IN (0, 1))
) STRICT;

CREATE TABLE IF NOT EXISTS idp_access_tokens (
	token_id TEXT PRIMARY KEY,
	refresh_token_id TEXT REFERENCES idp_refresh_tokens (token_id),
	jwt TEXT NOT NULL,
	user_id TEXT NOT NULL REFERENCES idp_users (user_id),
	client_id TEXT NOT NULL REFERENCES idp_clients (client_id),
	revoked INTEGER NOT NULL CHECK (revoked IN (0, 1))
) STRICT;
