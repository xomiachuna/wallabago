CREATE TABLE IF NOT EXISTS wallabago.bootstrap (
	condition_name TEXT NOT NULL PRIMARY KEY,
	satisfied BOOL NOT NULL DEFAULT FALSE
)
;