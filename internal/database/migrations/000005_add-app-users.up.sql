CREATE TABLE wallabago.users (
	user_id TEXT PRIMARY KEY,
	is_admin BOOL NOT NULL,
	username TEXT UNIQUE NOT NULL
)
;