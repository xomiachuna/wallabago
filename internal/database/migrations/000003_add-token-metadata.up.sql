-- Add issued_at (stored as Unix timestamp in seconds)
ALTER TABLE idp_access_tokens
ADD COLUMN issued_at_unix INTEGER NOT NULL DEFAULT (unixepoch());

-- Add expires_in
ALTER TABLE idp_access_tokens
ADD COLUMN expires_in_seconds INTEGER CHECK (expires_in_seconds > 0) NOT NULL DEFAULT 3600;

-- Add scope
ALTER TABLE idp_access_tokens
ADD COLUMN scope TEXT NOT NULL DEFAULT '';

-- Add token_type
ALTER TABLE idp_access_tokens
ADD COLUMN type TEXT NOT NULL DEFAULT 'bearer';
