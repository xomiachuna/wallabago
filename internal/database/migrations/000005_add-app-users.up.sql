CREATE TABLE app_users (
	user_id TEXT PRIMARY KEY,
	is_admin INTEGER NOT NULL CHECK (is_admin IN (0, 1)),
	username TEXT UNIQUE NOT NULL
) STRICT;
