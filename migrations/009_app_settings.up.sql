CREATE TABLE  IF NOT EXISTS app_settings (
    key TEXT PRIMARY KEY UNIQUE,
    value JSON NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now()
);

INSERT INTO app_settings (key, value)
VALUES
  ('FINANCE_OPIU_REVENUE_ITEMS', '["Выручка"]'),
  ('FINANCE_OPIU_COST_ITEMS',   '["Аренда помещений", "Банковские комиссии",
  "Маркетинг", "Налог за самого ИП", "Обслуживание ОС и НМА", 
  "вспомог услуги", "Общие налоги",
  "Оплата труда, сопутст. налоги, мотивация, рекрутинг",
  "Ремонт товаров/доукомплектование/транспорт",
  "Себестоимость товаров",
  "Таргет"
    ]'),
  ('FINANCE_OPIU_SPECIAL_ITEMS', '["Закуп товаров", "Личные нужды"]'),
  ('FINANCE_GHEETS_SOURCES', '[{"book": "1...","sheet": "2024","range": "A1:X?"}]')
ON CONFLICT (key) DO NOTHING;