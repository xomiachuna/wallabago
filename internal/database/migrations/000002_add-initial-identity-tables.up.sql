CREATE TABLE IF NOT EXISTS identity.users (
	user_id TEXT PRIMARY KEY,
	username TEXT NOT NULL UNIQUE,
	email TEXT NOT NULL UNIQUE,
	password_hash bytea NOT NULL
)
;

CREATE TABLE IF NOT EXISTS identity.clients (
	client_id TEXT PRIMARY KEY,
	client_secret TEXT NOT NULL
)
;

CREATE TABLE IF NOT EXISTS identity.refresh_tokens (
	token_id TEXT PRIMARY KEY,
	client_id TEXT NOT NULL REFERENCES identity.clients (client_id),
	jwt TEXT NOT NULL,
	revoked BOOL NOT NULL
)
;

CREATE TABLE IF NOT EXISTS identity.access_tokens (
	token_id TEXT PRIMARY KEY,
	refresh_token_id TEXT REFERENCES identity.refresh_tokens (token_id),
	jwt TEXT NOT NULL,
	user_id TEXT NOT NULL REFERENCES identity.users (user_id),
	client_id TEXT NOT NULL REFERENCES identity.clients (client_id),
	revoked BOOL NOT NULL
)
;