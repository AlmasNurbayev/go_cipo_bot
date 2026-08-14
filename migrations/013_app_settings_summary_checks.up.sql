INSERT INTO app_settings (key, caption, value)
VALUES
  ('SUMMARY_CHECKS_MAX_DAYS',
   'за сколько дней максимум отдаются списки чеков (кнопки "Все чеки" и "Все чеки #"). Массив с одним целым числом',
   '[7]')
ON CONFLICT (key) DO NOTHING;
