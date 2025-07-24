ALTER TABLE transactions
  DROP COLUMN IF EXISTS cheque_json;

ALTER TABLE transactions
  ADD COLUMN IF NOT EXISTS names TEXT[];
  
COMMENT ON COLUMN transactions.cheque_json IS NULL;