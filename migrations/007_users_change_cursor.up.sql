ALTER TABLE users
ALTER COLUMN transaction_cursor TYPE BIGINT
USING transaction_cursor::BIGINT;