-- Remove global email uniqueness constraint
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_unique;
