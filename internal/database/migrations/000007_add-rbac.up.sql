CREATE TABLE app_roles (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	description TEXT
) STRICT;

CREATE TABLE app_user_roles (
	user_id TEXT NOT NULL REFERENCES app_users(user_id) ON DELETE CASCADE,
	role_id TEXT NOT NULL REFERENCES app_roles(id) ON DELETE CASCADE,
	PRIMARY KEY (user_id, role_id)
) STRICT;

CREATE TABLE app_role_permissions (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	role_id TEXT NOT NULL REFERENCES app_roles(id) ON DELETE CASCADE,
	resource_type TEXT NOT NULL,
	operation TEXT NOT NULL,
	scope TEXT NOT NULL,
	CONSTRAINT unique_permission UNIQUE (role_id, resource_type, operation)
) STRICT;

-- Seed default roles
INSERT INTO app_roles (id, name, description) VALUES
	('admin', 'Administrator', 'Full system access'),
	('user', 'Standard User', 'Regular user access');

-- Admin permissions (global access to entries)
INSERT INTO app_role_permissions (role_id, resource_type, operation, scope) VALUES
	('admin', 'entries', 'create', 'own'),
	('admin', 'entries', 'read', 'global'),
	('admin', 'entries', 'update', 'global'),
	('admin', 'entries', 'delete', 'global'),
	('admin', 'entries', 'export', 'global');

-- Standard user permissions (own access only)
INSERT INTO app_role_permissions (role_id, resource_type, operation, scope) VALUES
	('user', 'entries', 'create', 'own'),
	('user', 'entries', 'read', 'own'),
	('user', 'entries', 'update', 'own'),
	('user', 'entries', 'delete', 'own'),
	('user', 'entries', 'export', 'own');
