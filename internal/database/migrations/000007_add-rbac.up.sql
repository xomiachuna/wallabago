CREATE TABLE wallabago.roles (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	description TEXT
);

CREATE TABLE wallabago.user_roles (
	user_id TEXT NOT NULL REFERENCES wallabago.users(user_id) ON DELETE CASCADE,
	role_id TEXT NOT NULL REFERENCES wallabago.roles(id) ON DELETE CASCADE,
	PRIMARY KEY (user_id, role_id)
);

CREATE TABLE wallabago.role_permissions (
	id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	role_id TEXT NOT NULL REFERENCES wallabago.roles(id) ON DELETE CASCADE,
	resource_type TEXT NOT NULL,
	operation TEXT NOT NULL,
	scope TEXT NOT NULL,
	CONSTRAINT unique_permission UNIQUE (role_id, resource_type, operation)
);

-- Seed default roles
INSERT INTO wallabago.roles (id, name, description) VALUES
	('admin', 'Administrator', 'Full system access'),
	('user', 'Standard User', 'Regular user access');

-- Admin permissions (global access to entries)
INSERT INTO wallabago.role_permissions (role_id, resource_type, operation, scope) VALUES
	('admin', 'entries', 'create', 'own'),
	('admin', 'entries', 'read', 'global'),
	('admin', 'entries', 'update', 'global'),
	('admin', 'entries', 'delete', 'global'),
	('admin', 'entries', 'export', 'global');

-- Standard user permissions (own access only)
INSERT INTO wallabago.role_permissions (role_id, resource_type, operation, scope) VALUES
	('user', 'entries', 'create', 'own'),
	('user', 'entries', 'read', 'own'),
	('user', 'entries', 'update', 'own'),
	('user', 'entries', 'delete', 'own'),
	('user', 'entries', 'export', 'own');
