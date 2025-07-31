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