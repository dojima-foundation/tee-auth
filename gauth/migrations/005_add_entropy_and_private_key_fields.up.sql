-- Add entropy field to wallets table
ALTER TABLE wallets ADD COLUMN entropy TEXT;

-- Add private_key field to wallet_accounts table  
ALTER TABLE wallet_accounts ADD COLUMN private_key TEXT;
