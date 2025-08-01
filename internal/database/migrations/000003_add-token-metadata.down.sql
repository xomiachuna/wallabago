ALTER TABLE identity.access_tokens
DROP COLUMN IF EXISTS type
;

ALTER TABLE identity.access_tokens
DROP COLUMN IF EXISTS issued_at
;

ALTER TABLE identity.access_tokens
DROP COLUMN IF EXISTS expires_in_seconds
;

ALTER TABLE identity.access_tokens
DROP COLUMN IF EXISTS scope
;