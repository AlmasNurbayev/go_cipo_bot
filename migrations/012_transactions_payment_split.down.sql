ALTER TABLE transactions
  DROP COLUMN IF EXISTS sum_cash,
  DROP COLUMN IF EXISTS sum_card;
