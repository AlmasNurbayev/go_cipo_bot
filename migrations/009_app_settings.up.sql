CREATE TABLE  IF NOT EXISTS app_settings (
    key TEXT PRIMARY KEY UNIQUE,
    value JSON NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now()
);

INSERT INTO app_settings (key, value)
VALUES
  ('FINANCE_OPIU_REVENUE_ITEMS', '{"value": ["Выручка"]}'),
  ('FINANCE_OPIU_COST_ITEMS',   '{"value": ["Аренда помещений", "Банковские комиссии",
  "Маркетинг", "Налог за самого ИП", "Обслуживание ОС и НМА", 
  "вспомог услуги", "Общие налоги",
  "Оплата труда, сопутст. налоги, мотивация, рекрутинг",
  "Ремонт товаров/доукомплектование/транспорт",
  "Себестоимость товаров",
  "Таргет"
    ]}'),
  ('FINANCE_OPIU_SPECIAL_ITEMS', '{"value": ["Закуп товаров", "Личные нужды"]}')
ON CONFLICT (key) DO NOTHING;