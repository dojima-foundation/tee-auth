-- Add public_key field back to wallets table
ALTER TABLE wallets ADD COLUMN public_key TEXT;
