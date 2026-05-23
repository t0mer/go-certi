-- Add issuer full DN and revocation status to certificates
ALTER TABLE certificates ADD COLUMN issuer_name TEXT NOT NULL DEFAULT '';
ALTER TABLE certificates ADD COLUMN revoked BOOLEAN NOT NULL DEFAULT 0 CHECK (revoked IN (0,1));
