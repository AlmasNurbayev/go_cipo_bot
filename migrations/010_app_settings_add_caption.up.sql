ALTER TABLE app_settings
  ADD COLUMN IF NOT EXISTS caption TEXT;

INSERT INTO app_settings (key, caption, value)
VALUES
  ('FINANCE_PLANNING_MARGIN',
   'соотношение выручки к себестоимости товара для расчета прогнозной себестоимости из текущей выручки. Массив с одним float числом',
   '[1.8]')
ON CONFLICT (key) DO NOTHING;