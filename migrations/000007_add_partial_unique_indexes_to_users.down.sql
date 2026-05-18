DROP INDEX IF EXISTS users_email_unique;

DROP INDEX IF EXISTS users_username_unique;

ALTER TABLE users
ADD CONSTRAINT users_email_key UNIQUE(email);