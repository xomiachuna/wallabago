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

-- name: AddUser :one
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

-- name: GetUserByUsername :one
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

-- name: DeleteUserByID :exec
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