-- Make public_key fields optional to support flexible authentication patterns
-- This aligns with the updated API that allows creating users and organizations without immediate public keys

-- Make user public_key optional
ALTER TABLE users ALTER COLUMN public_key DROP NOT NULL;

-- Make wallet public_key optional (it gets generated)  
ALTER TABLE wallets ALTER COLUMN public_key DROP NOT NULL;
