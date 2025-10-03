CREATE TABLE app_entries (
	entry_id INTEGER PRIMARY KEY AUTOINCREMENT,
	owner_id TEXT NOT NULL REFERENCES app_users (user_id),
	title TEXT NOT NULL,
	url TEXT NOT NULL,
	created_at_unix INTEGER NOT NULL DEFAULT (unixepoch()),
	retrieved_at_unix INTEGER DEFAULT (unixepoch()),
	content TEXT NOT NULL,
	sha1 BLOB NOT NULL
) STRICT;
