-- Revert making public_key fields optional
-- WARNING: This will fail if there are users or wallets with NULL public_key values

-- Revert user public_key to required (this may fail if there are NULL values)
ALTER TABLE users ALTER COLUMN public_key SET NOT NULL;

-- Revert wallet public_key to required (this may fail if there are NULL values)
ALTER TABLE wallets ALTER COLUMN public_key SET NOT NULL;
