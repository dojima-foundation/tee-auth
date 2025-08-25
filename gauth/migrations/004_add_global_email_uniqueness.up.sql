-- Add global email uniqueness constraint
-- This prevents the same email from being used across different organizations
ALTER TABLE users ADD CONSTRAINT users_email_unique UNIQUE(email);
