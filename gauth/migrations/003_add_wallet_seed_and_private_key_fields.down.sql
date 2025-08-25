-- Revert the unique constraint change
ALTER TABLE private_keys DROP CONSTRAINT IF EXISTS private_keys_organization_id_wallet_id_name_key;
ALTER TABLE private_keys ADD CONSTRAINT private_keys_organization_id_name_key UNIQUE(organization_id, name);

-- Drop the index
DROP INDEX IF EXISTS idx_private_keys_wallet_id;

-- Remove the new columns from private_keys
ALTER TABLE private_keys DROP COLUMN IF EXISTS path;
ALTER TABLE private_keys DROP COLUMN IF EXISTS wallet_id;

-- Remove the seed_phrase column from wallets
ALTER TABLE wallets DROP COLUMN IF EXISTS seed_phrase;
