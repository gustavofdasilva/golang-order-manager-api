ALTER TABLE users
DROP CONSTRAINT users_email_key;

CREATE UNIQUE INDEX users_email_unique
ON users(email)
WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX users_username_unique
ON users(username)
WHERE deleted_at IS NULL;