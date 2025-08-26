-- Add issued_at
ALTER TABLE identity.access_tokens
ADD COLUMN IF NOT EXISTS issued_at TIMESTAMP WITH TIME ZONE NOT NULL
;

-- Add expires_in
ALTER TABLE identity.access_tokens
ADD COLUMN IF NOT EXISTS expires_in_seconds BIGINT CHECK (expires_in_seconds > 0) NOT NULL
;

-- Add scope
ALTER TABLE identity.access_tokens
ADD COLUMN IF NOT EXISTS scope TEXT NOT NULL
;

-- Add token_type
ALTER TABLE identity.access_tokens
ADD COLUMN IF NOT EXISTS type TEXT NOT NULL DEFAULT 'bearer'
;