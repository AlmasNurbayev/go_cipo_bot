ALTER TABLE transactions
  ADD COLUMN IF NOT EXISTS cheque_json JSON;

ALTER TABLE transactions
  DROP COLUMN IF EXISTS names;
  
COMMENT ON COLUMN transactions.cheque_json IS 'Данные из чека структурированные в JSON';