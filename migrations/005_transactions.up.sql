CREATE TABLE IF NOT EXISTS transactions (
  id BIGINT GENERATED BY DEFAULT AS IDENTITY  PRIMARY KEY,
  ofd_id TEXT NOT NULL UNIQUE,
  ofd_name TEXT,
  onlinefiscalnumber BIGINT,
  offlinefiscalnumber BIGINT,
  systemdate timestamptz,
  operationdate timestamptz,
  type_operation int NOT NULL,
  subtype int,
  sum_operation NUMERIC,
  availablesum NUMERIC,
  paymenttypes int[],
  shift int,
  created_at timestamptz DEFAULT NOW(),
  organization_id BIGINT NOT NULL REFERENCES organizations(id),
  kassa_id BIGINT NOT NULL REFERENCES kassa(id),
  knumber TEXT,
  cheque TEXT,
  images TEXT[],
  names TEXT[]
);

CREATE INDEX IF NOT EXISTS transactions_idx ON transactions (knumber, operationdate);