CREATE TABLE IF NOT EXISTS app_bootstrap (
	condition_name TEXT NOT NULL PRIMARY KEY,
	satisfied INTEGER NOT NULL DEFAULT 0 CHECK (satisfied IN (0, 1))
) STRICT;
