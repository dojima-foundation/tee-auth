-- Drop all tables and functions in reverse order

DROP TRIGGER IF EXISTS update_activities_updated_at ON activities;
DROP TRIGGER IF EXISTS update_wallet_accounts_updated_at ON wallet_accounts;
DROP TRIGGER IF EXISTS update_wallets_updated_at ON wallets;
DROP TRIGGER IF EXISTS update_private_keys_updated_at ON private_keys;
DROP TRIGGER IF EXISTS update_policies_updated_at ON policies;
DROP TRIGGER IF EXISTS update_auth_methods_updated_at ON auth_methods;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TRIGGER IF EXISTS update_organizations_updated_at ON organizations;

DROP FUNCTION IF EXISTS update_updated_at_column();

DROP INDEX IF EXISTS idx_sessions_expires_at;
DROP INDEX IF EXISTS idx_sessions_token;
DROP INDEX IF EXISTS idx_proofs_activity_id;
DROP INDEX IF EXISTS idx_activities_created_by;
DROP INDEX IF EXISTS idx_activities_status;
DROP INDEX IF EXISTS idx_activities_type;
DROP INDEX IF EXISTS idx_activities_org_id;
DROP INDEX IF EXISTS idx_wallet_accounts_wallet_id;
DROP INDEX IF EXISTS idx_wallets_org_id;
DROP INDEX IF EXISTS idx_private_keys_org_id;
DROP INDEX IF EXISTS idx_tags_org_id;
DROP INDEX IF EXISTS idx_policies_org_id;
DROP INDEX IF EXISTS idx_invitations_expires_at;
DROP INDEX IF EXISTS idx_invitations_token;
DROP INDEX IF EXISTS idx_invitations_org_id;
DROP INDEX IF EXISTS idx_auth_methods_type;
DROP INDEX IF EXISTS idx_auth_methods_user_id;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_org_id;

DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS quorum_members;
DROP TABLE IF EXISTS wallet_tags;
DROP TABLE IF EXISTS private_key_tags;
DROP TABLE IF EXISTS user_tags;
DROP TABLE IF EXISTS proofs;
DROP TABLE IF EXISTS activities;
DROP TABLE IF EXISTS wallet_accounts;
DROP TABLE IF EXISTS wallets;
DROP TABLE IF EXISTS private_keys;
DROP TABLE IF EXISTS tags;
DROP TABLE IF EXISTS policies;
DROP TABLE IF EXISTS invitations;
DROP TABLE IF EXISTS auth_methods;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS organizations;

DROP EXTENSION IF EXISTS "uuid-ossp";
