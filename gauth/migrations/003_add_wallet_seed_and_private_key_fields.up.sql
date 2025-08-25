-- Add seed_phrase field to wallets table
ALTER TABLE wallets ADD COLUMN seed_phrase TEXT NOT NULL DEFAULT '';

-- Add wallet_id and path fields to private_keys table
ALTER TABLE private_keys ADD COLUMN wallet_id UUID REFERENCES wallets(id) ON DELETE CASCADE;
ALTER TABLE private_keys ADD COLUMN path VARCHAR(255); -- BIP32 derivation path

-- Create index for wallet_id in private_keys
CREATE INDEX idx_private_keys_wallet_id ON private_keys(wallet_id);

-- Update the unique constraint on private_keys to include wallet_id
ALTER TABLE private_keys DROP CONSTRAINT IF EXISTS private_keys_organization_id_name_key;
ALTER TABLE private_keys ADD CONSTRAINT private_keys_organization_id_wallet_id_name_key UNIQUE(organization_id, wallet_id, name);
