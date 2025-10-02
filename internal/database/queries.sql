-- name: GetBoostrapConditions :many
SELECT
	condition_name,
	satisfied
FROM
	wallabago.bootstrap
;

-- name: MarkBootstrapConditionSatisfied :one
INSERT INTO
	wallabago.bootstrap (condition_name, satisfied)
VALUES
	($1, TRUE)
ON CONFLICT ON CONSTRAINT bootstrap_pkey DO UPDATE
SET
	satisfied = TRUE
RETURNING
	condition_name,
	satisfied
;

-- name: AddClient :one
INSERT INTO
	identity.clients (client_id, client_secret)
VALUES
	($1, $2)
RETURNING
	client_id,
	client_secret
;

-- name: GetClientByID :one
SELECT
	client_id,
	client_secret
FROM
	identity.clients
WHERE
	client_id = $1
LIMIT
	1
;

-- name: DeleteClientByID :exec
DELETE FROM identity.clients
WHERE
	client_id = $1
;

-- name: AddIdentityUser :one
INSERT INTO
	identity.users (user_id, username, email, password_hash)
VALUES
	($1, $2, $3, $4)
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
	identity.users
WHERE
	username = $1
LIMIT
	1
;

-- name: DeleteIdentityUserByID :exec
DELETE FROM identity.users
WHERE
	user_id = $1
;

-- name: AddRefreshToken :one
INSERT INTO
	identity.refresh_tokens (token_id, client_id, jwt, revoked)
VALUES
	($1, $2, $3, $4)
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
	identity.refresh_tokens
WHERE
	jwt = $1
LIMIT
	1
;

-- name: RevokeRefreshTokenByID :one
UPDATE identity.refresh_tokens
SET
	revoked = TRUE
WHERE
	token_id = $1
RETURNING
	token_id,
	client_id,
	jwt,
	revoked
;

-- name: DeleteRefreshTokenByID :exec
DELETE FROM identity.refresh_tokens
WHERE
	token_id = $1
;

-- name: AddAccessToken :one
INSERT INTO
	identity.access_tokens (
		token_id,
		refresh_token_id,
		client_id,
		user_id,
		jwt,
		revoked,
		expires_in_seconds,
		issued_at,
		scope,
		type
	)
VALUES
	($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING
	token_id,
	refresh_token_id,
	client_id,
	user_id,
	jwt,
	revoked,
	expires_in_seconds,
	issued_at,
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
	issued_at,
	scope,
	type
FROM
	identity.access_tokens
WHERE
	jwt = $1
LIMIT
	1
;

-- name: RevokeAccessTokenByID :one
UPDATE identity.access_tokens
SET
	revoked = TRUE
WHERE
	token_id = $1
RETURNING
	token_id,
	refresh_token_id,
	client_id,
	user_id,
	jwt,
	revoked,
	expires_in_seconds,
	issued_at,
	scope,
	type
;

-- name: DeleteAccessTokenByID :exec
DELETE FROM identity.access_tokens
WHERE
	token_id = $1
;

-- name: AddAppUser :one
INSERT INTO
	wallabago.users (user_id, is_admin, username)
VALUES
	($1, $2, $3)
RETURNING
	user_id,
	is_admin,
	username
;

-- name: AddEntry :one
INSERT INTO
	wallabago.entries (url, title, "content", owner_id, sha1)
VALUES
	($1, $2, $3, $4, $5)
RETURNING
	entry_id,
	owner_id,
	title,
	url,
	created_at,
	retrieved_at,
	sha1,
	content
;

-- name: GetEntryBySHA1 :one
SELECT
	entry_id,
	owner_id,
	title,
	url,
	created_at,
	retrieved_at,
	sha1,
	content
FROM
	wallabago.entries
WHERE
	sha1 = $1
	AND owner_id = $2
LIMIT
	1
;

-- name: GetEntryByID :one
SELECT
	entry_id,
	owner_id,
	title,
	url,
	created_at,
	retrieved_at,
	sha1,
	content
FROM
	wallabago.entries
WHERE
	entry_id = $1
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
	)::INT AS scope_level
FROM
	wallabago.user_roles ur
	JOIN wallabago.role_permissions rp ON ur.role_id = rp.role_id
WHERE
	ur.user_id = $1
	AND rp.resource_type = $2
	AND rp.operation = $3
;

-- name: GetEntryOwnerID :one
SELECT
	owner_id
FROM
	wallabago.entries
WHERE
	entry_id = $1
LIMIT
	1
;

-- name: AssignUserRole :exec
INSERT INTO
	wallabago.user_roles (user_id, role_id)
VALUES
	($1, $2)
ON CONFLICT DO NOTHING
;