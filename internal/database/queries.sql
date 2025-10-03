-- name: GetBoostrapConditions :many
SELECT
	condition_name,
	satisfied
FROM
	app_bootstrap
;

-- name: MarkBootstrapConditionSatisfied :one
INSERT INTO
	app_bootstrap (condition_name, satisfied)
VALUES
	(?, 1)
ON CONFLICT (condition_name) DO UPDATE
SET
	satisfied = 1
RETURNING
	condition_name,
	satisfied
;

-- name: AddClient :one
INSERT INTO
	idp_clients (client_id, client_secret)
VALUES
	(?, ?)
RETURNING
	client_id,
	client_secret
;

-- name: GetClientByID :one
SELECT
	client_id,
	client_secret
FROM
	idp_clients
WHERE
	client_id = ?
LIMIT
	1
;

-- name: DeleteClientByID :exec
DELETE FROM idp_clients
WHERE
	client_id = ?
;

-- name: AddIdentityUser :one
INSERT INTO
	idp_users (user_id, username, email, password_hash)
VALUES
	(?, ?, ?, ?)
RETURNING
	user_id,
	username,
	email,
	password_hash
;

-- name: GetIdentityUserByUsername :one
SELECT
	user_id,
	username,
	email,
	password_hash
FROM
	idp_users
WHERE
	username = ?
LIMIT
	1
;

-- name: DeleteIdentityUserByID :exec
DELETE FROM idp_users
WHERE
	user_id = ?
;

-- name: AddRefreshToken :one
INSERT INTO
	idp_refresh_tokens (token_id, client_id, jwt, revoked)
VALUES
	(?, ?, ?, ?)
RETURNING
	token_id,
	client_id,
	jwt,
	revoked
;

-- name: GetRefreshTokenByJWT :one
SELECT
	token_id,
	client_id,
	jwt,
	revoked
FROM
	idp_refresh_tokens
WHERE
	jwt = ?
LIMIT
	1
;

-- name: RevokeRefreshTokenByID :one
UPDATE idp_refresh_tokens
SET
	revoked = 1
WHERE
	token_id = ?
RETURNING
	token_id,
	client_id,
	jwt,
	revoked
;

-- name: DeleteRefreshTokenByID :exec
DELETE FROM idp_refresh_tokens
WHERE
	token_id = ?
;

-- name: AddAccessToken :one
INSERT INTO
	idp_access_tokens (
		token_id,
		refresh_token_id,
		client_id,
		user_id,
		jwt,
		revoked,
		expires_in_seconds,
		issued_at_unix,
		scope,
		type
	)
VALUES
	(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING
	token_id,
	refresh_token_id,
	client_id,
	user_id,
	jwt,
	revoked,
	expires_in_seconds,
	issued_at_unix,
	scope,
	type
;

-- name: GetAccessTokenByJWT :one
SELECT
	token_id,
	refresh_token_id,
	client_id,
	user_id,
	jwt,
	revoked,
	expires_in_seconds,
	issued_at_unix,
	scope,
	type
FROM
	idp_access_tokens
WHERE
	jwt = ?
LIMIT
	1
;

-- name: RevokeAccessTokenByID :one
UPDATE idp_access_tokens
SET
	revoked = 1
WHERE
	token_id = ?
RETURNING
	token_id,
	refresh_token_id,
	client_id,
	user_id,
	jwt,
	revoked,
	expires_in_seconds,
	issued_at_unix,
	scope,
	type
;

-- name: DeleteAccessTokenByID :exec
DELETE FROM idp_access_tokens
WHERE
	token_id = ?
;

-- name: AddAppUser :one
INSERT INTO
	app_users (user_id, is_admin, username)
VALUES
	(?, ?, ?)
RETURNING
	user_id,
	is_admin,
	username
;

-- name: AddEntry :one
INSERT INTO
	app_entries (url, title, "content", owner_id, sha1)
VALUES
	(?, ?, ?, ?, ?)
RETURNING
	entry_id,
	owner_id,
	title,
	url,
	created_at_unix,
	retrieved_at_unix,
	sha1,
	content
;

-- name: GetEntryBySHA1 :one
SELECT
	entry_id,
	owner_id,
	title,
	url,
	created_at_unix,
	retrieved_at_unix,
	sha1,
	content
FROM
	app_entries
WHERE
	sha1 = ?
	AND owner_id = ?
LIMIT
	1
;

-- name: GetEntryByID :one
SELECT
	entry_id,
	owner_id,
	title,
	url,
	created_at_unix,
	retrieved_at_unix,
	sha1,
	content
FROM
	app_entries
WHERE
	entry_id = ?
LIMIT
	1
;

-- name: GetUserMaxScope :one
-- Returns the highest scope level for a user's permission on a resource type and operation.
-- COALESCE returns 0 if MAX is NULL (when user has no roles or no matching permissions).
-- Scope levels: 0 = none/no permission, 1 = none, 2 = own, 3 = global
SELECT
	COALESCE(
		MAX(
			CASE rp.scope
				WHEN 'global' THEN 3
				WHEN 'own' THEN 2
				WHEN 'none' THEN 1
				ELSE 0
			END
		),
		0
	) AS scope_level
FROM
	app_user_roles ur
	JOIN app_role_permissions rp ON ur.role_id = rp.role_id
WHERE
	ur.user_id = ?
	AND rp.resource_type = ?
	AND rp.operation = ?
;

-- name: GetEntryOwnerID :one
SELECT
	owner_id
FROM
	app_entries
WHERE
	entry_id = ?
LIMIT
	1
;

-- name: AssignUserRole :exec
INSERT INTO
	app_user_roles (user_id, role_id)
VALUES
	(?, ?)
ON CONFLICT DO NOTHING
;
