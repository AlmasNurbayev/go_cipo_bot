ALTER TABLE users
ALTER COLUMN transaction_cursor TYPE TEXT
USING transaction_cursor::TEXT;