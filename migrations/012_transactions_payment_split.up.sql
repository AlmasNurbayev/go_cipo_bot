ALTER TABLE transactions
  ADD COLUMN IF NOT EXISTS sum_cash NUMERIC,
  ADD COLUMN IF NOT EXISTS sum_card NUMERIC;

COMMENT ON COLUMN transactions.sum_cash IS 'Сумма оплаты наличными; для смешанной оплаты - из разбора текста чека, иначе равна sum_operation';
COMMENT ON COLUMN transactions.sum_card IS 'Сумма оплаты картой; для смешанной оплаты - из разбора текста чека, иначе равна sum_operation';
