-- Remove private_key field from wallet_accounts table
ALTER TABLE wallet_accounts DROP COLUMN private_key;

-- Remove entropy field from wallets table
ALTER TABLE wallets DROP COLUMN entropy;
