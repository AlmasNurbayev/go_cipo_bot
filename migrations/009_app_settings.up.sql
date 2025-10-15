CREATE TABLE  IF NOT EXISTS app_settings (
    key TEXT PRIMARY KEY UNIQUE,
    value JSON NOT NULL,
    caption TEXT,
    updated_at TIMESTAMPTZ DEFAULT now()
);

INSERT INTO app_settings (key, caption, value)
VALUES
  ('FINANCE_OPIU_REVENUE_ITEMS', 'массив статей для выручки', '["Выручка"]'),
  ('FINANCE_OPIU_COST_ITEMS', 'массив статей для расходов',  '["Аренда помещений", "Банковские комиссии",
  "Маркетинг", "Налог за самого ИП", "Обслуживание ОС и НМА, вспомог услуги", 
  "Общие налоги", "Оплата труда, сопутст. налоги, мотивация, рекрутинг",
  "Ремонт товаров/доукомплектование/транспорт",
  "Себестоимость товаров",
  "Таргет"
    ]'),
  ('FINANCE_OPIU_SPECIAL_ITEMS', 'массив статей для отображения отдельно от ОПиУ', '["Закуп товаров", "Личные нужды"]'),
  ('FINANCE_GHEETS_SOURCES', 'массив ссылок на файлы Gsheets в виде объектов: book - id книги, sheet - имя листа, range - диапазон',
   '[{"book": "1","sheet": "2024","range": "A1:X?"}]'),
  ('FINANCE_USD_RATE', 'массив со среднегодовым курсом доллара для пересчета показателей в виде объектов - год(string): число int ' ,
   '[{"2024": 470, "2025": 520}]'),
  ('FINANCE_PLANNING_MARGIN',
   'соотношение выручки к себестоимости товара для расчета прогнозной себестоимости из текущей выручки. Массив с одним float числом',
   '[1.8]')
ON CONFLICT (key) DO NOTHING;