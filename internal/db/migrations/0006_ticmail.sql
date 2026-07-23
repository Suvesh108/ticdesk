-- Up migration: ticmail_logs table for built-in ticMail email inspector

CREATE TABLE ticmail_logs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recipient   TEXT NOT NULL,
    subject     TEXT NOT NULL,
    body_html   TEXT NOT NULL,
    status      TEXT NOT NULL DEFAULT 'DELIVERED',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_ticmail_recipient ON ticmail_logs(recipient);
CREATE INDEX idx_ticmail_created_at ON ticmail_logs(created_at DESC);
