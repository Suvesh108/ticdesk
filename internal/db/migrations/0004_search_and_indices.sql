-- Up migration: PostgreSQL trigram search extension and GIN indices

CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS idx_tickets_trgm_title ON tickets USING gin (title gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_tickets_trgm_description ON tickets USING gin (description gin_trgm_ops);
