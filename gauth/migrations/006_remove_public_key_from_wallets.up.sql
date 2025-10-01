-- Remove public_key field from wallets table
ALTER TABLE wallets DROP COLUMN public_key;
